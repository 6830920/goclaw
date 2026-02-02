// Package memory provides memory management for OpenClaw-Go
// Implements short-term, long-term, and working memory
package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"openclaw-go/internal/vector"
)

// MemoryType defines the type of memory
type MemoryType string

const (
	MemoryTypeShort MemoryType = "short"
	MemoryTypeLong  MemoryType = "long"
	MemoryTypeWork  MemoryType = "working"
)

// MemoryEntry represents a single memory entry
type MemoryEntry struct {
	ID        string                 `json:"id"`
	Type      MemoryType             `json:"type"`
	Content   string                 `json:"content"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Embedding []float32              `json:"embedding,omitempty"`
}

// MemoryStore manages all types of memory
type MemoryStore struct {
	mu          sync.RWMutex
	shortTerm   *ConversationBuffer
	longTerm    *VectorMemory
	workingSet  *WorkingMemory
	config      MemoryConfig
}

// MemoryConfig holds memory configuration
type MemoryConfig struct {
	ShortTermMax   int     // Maximum short-term memories
	WorkingMax     int     // Maximum working memory items
	SimilarityCut  float32 // Similarity threshold for long-term memory
}

// MemorySearchResult represents a memory search result
type MemorySearchResult struct {
	Entry   MemoryEntry `json:"entry"`
	Score   float32     `json:"score"`
	Reasons []string    `json:"reasons,omitempty"`
}

// NewMemoryStore creates a new memory store
func NewMemoryStore(config MemoryConfig) *MemoryStore {
	return &MemoryStore{
		config:     config,
		shortTerm:  NewConversationBuffer(config.ShortTermMax),
		longTerm:   NewVectorMemory(),
		workingSet: NewWorkingMemory(config.WorkingMax),
	}
}

// DefaultConfig returns default memory configuration
func DefaultConfig() MemoryConfig {
	return MemoryConfig{
		ShortTermMax:   50,    // Keep last 50 messages
		WorkingMax:     10,    // Keep 10 working items
		SimilarityCut:  0.7,   // 70% similarity threshold
	}
}

// AddShortTerm adds a short-term memory (conversation)
func (m *MemoryStore) AddShortTerm(content string, metadata map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry := MemoryEntry{
		ID:        fmt.Sprintf("st_%d", time.Now().UnixNano()),
		Type:      MemoryTypeShort,
		Content:   content,
		Timestamp: time.Now(),
		Metadata:  metadata,
	}

	m.shortTerm.Add(entry)
}

// AddLongTerm adds a long-term memory with embedding
func (m *MemoryStore) AddLongTerm(content string, embedding []float32, metadata map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry := MemoryEntry{
		ID:        fmt.Sprintf("lt_%d", time.Now().UnixNano()),
		Type:      MemoryTypeLong,
		Content:   content,
		Timestamp: time.Now(),
		Metadata:  metadata,
	}

	return m.longTerm.Add(entry, embedding)
}

// AddWorking adds to working memory
func (m *MemoryStore) AddWorking(content string, priority int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry := MemoryEntry{
		ID:        fmt.Sprintf("wm_%d", time.Now().UnixNano()),
		Type:      MemoryTypeWork,
		Content:   content,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"priority": priority,
		},
	}

	m.workingSet.Add(entry)
}

// Search searches long-term memory
func (m *MemoryStore) Search(ctx context.Context, query string, embedding []float32, limit int) ([]MemorySearchResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	results, err := m.longTerm.Search(ctx, embedding, limit)
	if err != nil {
		return nil, err
	}

	memoryResults := make([]MemorySearchResult, len(results))
	for i, r := range results {
		memoryResults[i] = MemorySearchResult{
			Entry: MemoryEntry{
				ID:        r.ID,
				Content:   r.Content,
				Timestamp: time.Unix(r.Metadata.Timestamp, 0),
			},
			Score: r.Score,
		}
	}

	return memoryResults, nil
}

// GetContext retrieves all relevant context for a conversation
func (m *MemoryStore) GetContext(ctx context.Context, query string, embedding []float32, maxTokens int) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var contextParts []string

	// 1. Get working memory
	for _, entry := range m.workingSet.GetAll() {
		if len(contextParts) >= maxTokens/3 {
			break
		}
		contextParts = append(contextParts, fmt.Sprintf("[WORKING]: %s", entry.Content))
	}

	// 2. Get relevant long-term memories
	longTerm, err := m.longTerm.Search(ctx, embedding, 5)
	if err == nil {
		for _, r := range longTerm {
			if len(contextParts) >= maxTokens*2/3 {
				break
			}
			if r.Score >= m.config.SimilarityCut {
				contextParts = append(contextParts, 
					fmt.Sprintf("[MEMORY (%.2f)]: %s", r.Score, r.Content))
			}
		}
	}

	// 3. Get recent short-term memories
	recent := m.shortTerm.GetRecent(10)
	for _, entry := range recent {
		contextParts = append(contextParts, 
			fmt.Sprintf("[RECENT]: %s", entry.Content))
	}

	// Combine context
	context := ""
	for i, part := range contextParts {
		if i > 0 {
			context += "\n"
		}
		context += part
	}

	return context, nil
}

// Consolidate moves important short-term memories to long-term
func (m *MemoryStore) Consolidate(embedder *vector.OllamaEmbedder) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	recent := m.shortTerm.GetRecent(20)
	for _, entry := range recent {
		// Check if this memory should be consolidated
		// For now, consolidate all memories older than 1 hour
		if time.Since(entry.Timestamp) > time.Hour {
			// Generate embedding
			var embedding []float32
			if embedder != nil {
				emb, err := embedder.Embed(context.Background(), entry.Content)
				if err != nil {
					continue
				}
				embedding = emb
			}

			// Add to long-term
			m.longTerm.Add(entry, embedding)
			
			// Remove from short-term
			m.shortTerm.Remove(entry.ID)
		}
	}

	return nil
}

// Clear clears all memories
func (m *MemoryStore) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.shortTerm.Clear()
	m.longTerm.Clear()
	m.workingSet.Clear()
}

// Stats returns memory statistics
func (m *MemoryStore) Stats() MemoryStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return MemoryStats{
		ShortTermCount: m.shortTerm.Len(),
		LongTermCount:  m.longTerm.Len(),
		WorkingCount:   m.workingSet.Len(),
	}
}

// MemoryStats holds statistics about memory usage
type MemoryStats struct {
	ShortTermCount int `json:"shortTermCount"`
	LongTermCount  int `json:"longTermCount"`
	WorkingCount   int `json:"workingCount"`
}
