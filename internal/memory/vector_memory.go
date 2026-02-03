package memory

import (
	"context"
	"sync"
)

// SearchResult represents a search result
type SearchResult struct {
	ID       string
	Score    float32
	Content  string
	Metadata MemoryMetadata
}

// MemoryMetadata holds metadata for memory
type MemoryMetadata struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

// VectorMemory manages long-term vector-based memory
type VectorMemory struct {
	mu      sync.RWMutex
	entries map[string]MemoryEntry
	vectors map[string][]float32
}

// NewVectorMemory creates a new vector memory store
func NewVectorMemory() *VectorMemory {
	return &VectorMemory{
		entries: make(map[string]MemoryEntry),
		vectors: make(map[string][]float32),
	}
}

// Add adds a memory entry with its embedding
func (vm *VectorMemory) Add(entry MemoryEntry, embedding []float32) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	vm.entries[entry.ID] = entry
	vm.vectors[entry.ID] = embedding

	return nil
}

// Search searches for similar memories
func (vm *VectorMemory) Search(ctx context.Context, query []float32, limit int) ([]SearchResult, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	type scoredEntry struct {
		id         string
		similarity float32
	}

	var results []scoredEntry
	for id, vector := range vm.vectors {
		score := cosineSimilarity(query, vector)
		results = append(results, scoredEntry{
			id:         id,
			similarity: score,
		})
	}

	// Sort by similarity
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].similarity > results[i].similarity {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// Take top k
	if len(results) > limit {
		results = results[:limit]
	}

	searchResults := make([]SearchResult, len(results))
	for i, r := range results {
		entry := vm.entries[r.id]
		searchResults[i] = SearchResult{
			ID:      r.id,
			Score:   r.similarity,
			Content: entry.Content,
			Metadata: MemoryMetadata{
				Timestamp: entry.Timestamp.Unix(),
				Content:   entry.Content,
			},
		}
	}

	return searchResults, nil
}

// Get retrieves a memory entry
func (vm *VectorMemory) Get(id string) (*MemoryEntry, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	entry, exists := vm.entries[id]
	if !exists {
		return nil, nil
	}
	return &entry, nil
}

// Len returns the number of entries
func (vm *VectorMemory) Len() int {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return len(vm.entries)
}

// Clear clears all entries
func (vm *VectorMemory) Clear() {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.entries = make(map[string]MemoryEntry)
	vm.vectors = make(map[string][]float32)
}

// cosineSimilarity calculates cosine similarity between two vectors
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct, normA, normB float32
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	normA = float32(sqrt(float64(normA)))
	normB = float32(sqrt(float64(normB)))

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (normA * normB)
}

func sqrt(x float64) float64 {
	// Simple square root approximation
	if x < 0 {
		return 0
	}

	// Using Newton's method
	z := x / 2
	for i := 0; i < 20; i++ {
		z = (z + x/z) / 2
	}
	return z
}
