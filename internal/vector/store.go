// Package vector provides vector storage and retrieval capabilities
package vector

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// MemoryMetadata stores metadata for a vector entry
type MemoryMetadata struct {
	ID        string            `json:"id"`
	Content   string            `json:"content"`
	Timestamp int64             `json:"timestamp"`
	Tags      []string          `json:"tags"`
	Custom    map[string]string `json:"custom,omitempty"`
}

// VectorEntry represents a stored vector with metadata
type VectorEntry struct {
	Vector   []float32      `json:"vector"`
	Metadata MemoryMetadata `json:"metadata"`
}

// VectorStore interface for storing and retrieving vectors
type VectorStore interface {
	Add(ctx context.Context, vector []float32, metadata MemoryMetadata) (string, error)
	Search(ctx context.Context, query []float32, limit int) ([]SearchResult, error)
	Get(ctx context.Context, id string) (*VectorEntry, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]VectorEntry, error)
	Count(ctx context.Context) (int, error)
	Save(ctx context.Context, path string) error
	Load(ctx context.Context, path string) error
}

// InMemoryStore is a simple in-memory vector store
type InMemoryStore struct {
	mu       sync.RWMutex
	vectors  map[string]*VectorEntry
	embedder Embedder
}

// SearchResult represents a search match
type SearchResult struct {
	ID       string         `json:"id"`
	Score    float32        `json:"score"`
	Content  string         `json:"content"`
	Metadata MemoryMetadata `json:"metadata"`
}

// NewInMemoryStore creates a new in-memory vector store
func NewInMemoryStore(embedder Embedder) *InMemoryStore {
	return &InMemoryStore{
		vectors:  make(map[string]*VectorEntry),
		embedder: embedder,
	}
}

// Add adds a new vector to the store
func (s *InMemoryStore) Add(ctx context.Context, vector []float32, metadata MemoryMetadata) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate ID if not provided
	if metadata.ID == "" {
		metadata.ID = fmt.Sprintf("vec_%d_%d", len(s.vectors), now())
	}

	entry := &VectorEntry{
		Vector:   vector,
		Metadata: metadata,
	}

	s.vectors[metadata.ID] = entry
	return metadata.ID, nil
}

// AddWithEmbedding adds text and automatically generates embedding
func (s *InMemoryStore) AddWithEmbedding(ctx context.Context, content string, tags []string, custom map[string]string) (string, error) {
	// Generate embedding if embedder is available
	var vector []float32
	if s.embedder != nil {
		emb, err := s.embedder.Embed(ctx, content)
		if err != nil {
			return "", fmt.Errorf("failed to generate embedding: %w", err)
		}
		vector = emb
	}

	metadata := MemoryMetadata{
		Content:   content,
		Timestamp: now(),
		Tags:      tags,
		Custom:    custom,
	}

	return s.Add(ctx, vector, metadata)
}

// Search finds the most similar vectors
func (s *InMemoryStore) Search(ctx context.Context, query []float32, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	type scoredEntry struct {
		id         string
		entry      *VectorEntry
		similarity float32
	}

	var results []scoredEntry
	for id, entry := range s.vectors {
		score := Similarity(query, entry.Vector)
		results = append(results, scoredEntry{
			id:         id,
			entry:      entry,
			similarity: score,
		})
	}

	// Sort by similarity (highest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].similarity > results[j].similarity
	})

	// Take top k
	if len(results) > limit {
		results = results[:limit]
	}

	// Convert to search results
	searchResults := make([]SearchResult, len(results))
	for i, r := range results {
		searchResults[i] = SearchResult{
			ID:       r.id,
			Score:    r.similarity,
			Content:  r.entry.Metadata.Content,
			Metadata: r.entry.Metadata,
		}
	}

	return searchResults, nil
}

// SearchByText searches using text query (generates embedding automatically)
func (s *InMemoryStore) SearchByText(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	if s.embedder == nil {
		return nil, fmt.Errorf("no embedder configured for text search")
	}

	embedding, err := s.embedder.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	return s.Search(ctx, embedding, limit)
}

// Get retrieves a vector by ID
func (s *InMemoryStore) Get(ctx context.Context, id string) (*VectorEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.vectors[id]
	if !exists {
		return nil, fmt.Errorf("vector not found: %s", id)
	}

	return entry, nil
}

// Delete removes a vector from the store
func (s *InMemoryStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.vectors[id]; !exists {
		return fmt.Errorf("vector not found: %s", id)
	}

	delete(s.vectors, id)
	return nil
}

// List returns all vectors with pagination
func (s *InMemoryStore) List(ctx context.Context, limit, offset int) ([]VectorEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 100
	}

	entries := make([]VectorEntry, 0, limit)
	i := 0
	for _, entry := range s.vectors {
		if i >= offset+limit {
			break
		}
		if i >= offset {
			entries = append(entries, *entry)
		}
		i++
	}

	return entries, nil
}

// Count returns the number of vectors in the store
func (s *InMemoryStore) Count(ctx context.Context) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.vectors), nil
}

// Save saves the store to a JSON file
func (s *InMemoryStore) Save(ctx context.Context, path string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Convert to serializable format
	type SerializedEntry struct {
		Vector   []float32      `json:"vector"`
		Metadata MemoryMetadata `json:"metadata"`
	}

	serialized := make(map[string]SerializedEntry)
	for id, entry := range s.vectors {
		serialized[id] = SerializedEntry{
			Vector:   entry.Vector,
			Metadata: entry.Metadata,
		}
	}

	data, err := json.MarshalIndent(serialized, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// Load loads the store from a JSON file
func (s *InMemoryStore) Load(ctx context.Context, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No file to load
		}
		return fmt.Errorf("failed to read file: %w", err)
	}

	type SerializedEntry struct {
		Vector   []float32      `json:"vector"`
		Metadata MemoryMetadata `json:"metadata"`
	}

	serialized := make(map[string]SerializedEntry)
	if err := json.Unmarshal(data, &serialized); err != nil {
		return fmt.Errorf("failed to unmarshal: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.vectors = make(map[string]*VectorEntry)
	for id, entry := range serialized {
		s.vectors[id] = &VectorEntry{
			Vector:   entry.Vector,
			Metadata: entry.Metadata,
		}
	}

	return nil
}

// now returns current Unix timestamp
func now() int64 {
	return time.Now().Unix()
}

// NewSQLiteStore creates a new SQLite-based vector store
// If SQLite build tag is not set, returns in-memory store
func NewSQLiteStore(embedder Embedder, dbPath string) (VectorStore, error) {
	// For now, return in-memory store as fallback
	// Actual SQLite implementation would require CGO
	return NewInMemoryStore(embedder), nil
}
