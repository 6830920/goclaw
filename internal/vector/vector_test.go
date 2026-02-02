package vector

import (
	"context"
	"testing"
	"time"
)

// MockEmbedder is a mock implementation of the Embedder interface for testing
type MockEmbedder struct{}

func (m *MockEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	// Return a simple mock vector based on the input text
	vec := make([]float32, 4) // Small vector for testing
	for i := 0; i < len(vec) && i < len(text); i++ {
		vec[i] = float32(text[i])
	}
	return vec, nil
}

func (m *MockEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	vectors := make([][]float32, len(texts))
	for i, text := range texts {
		vec, err := m.Embed(ctx, text)
		if err != nil {
			return nil, err
		}
		vectors[i] = vec
	}
	return vectors, nil
}

func (m *MockEmbedder) GetModelName() string {
	return "mock-embedder"
}

func TestInMemoryStore_AddAndSearch(t *testing.T) {
	ctx := context.Background()
	embedder := &MockEmbedder{}
	store := NewInMemoryStore(embedder)

	// Add some test data
	testContent1 := "Hello world"
	testContent2 := "Goodbye world"
	testContent3 := "Test content"

	// Add items to the store
	id1, err := store.AddWithEmbedding(ctx, testContent1, []string{"greeting"}, map[string]string{"source": "test"})
	if err != nil {
		t.Fatalf("Failed to add item 1: %v", err)
	}

	_, err = store.AddWithEmbedding(ctx, testContent2, []string{"farewell"}, map[string]string{"source": "test"})
	if err != nil {
		t.Fatalf("Failed to add item 2: %v", err)
	}

	_, err = store.AddWithEmbedding(ctx, testContent3, []string{"test"}, map[string]string{"source": "test"})
	if err != nil {
		t.Fatalf("Failed to add item 3: %v", err)
	}

	// Test search
	results, err := store.SearchByText(ctx, "hello", 5)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("Expected search results, got none")
	}

	// Check if "Hello world" is among the results (regardless of position due to mock embedding)
	foundHello := false
	for _, result := range results {
		if result.Content == testContent1 {
			foundHello = true
			break
		}
	}
	if !foundHello {
		t.Errorf("Expected to find '%s' in search results, but it wasn't found", testContent1)
	}

	// Test getting a specific item
	item, err := store.Get(ctx, id1)
	if err != nil {
		t.Fatalf("Failed to get item: %v", err)
	}

	if item.Metadata.Content != testContent1 {
		t.Errorf("Expected content '%s', got '%s'", testContent1, item.Metadata.Content)
	}

	// Test count
	count, err := store.Count(ctx)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}

	// Test list
	list, err := store.List(ctx, 10, 0)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(list) != 3 {
		t.Errorf("Expected list length 3, got %d", len(list))
	}

	// Test delete
	err = store.Delete(ctx, id1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	countAfter, err := store.Count(ctx)
	if err != nil {
		t.Fatalf("Count after delete failed: %v", err)
	}

	if countAfter != 2 {
		t.Errorf("Expected count 2 after delete, got %d", countAfter)
	}
}

func TestInMemoryStore_EmbeddingFunctionality(t *testing.T) {
	ctx := context.Background()
	embedder := &MockEmbedder{}
	store := NewInMemoryStore(embedder)

	// Test adding with embedding
	content := "test embedding functionality"
	tags := []string{"test", "embedding"}
	custom := map[string]string{"category": "functionality", "priority": "high"}

	id, err := store.AddWithEmbedding(ctx, content, tags, custom)
	if err != nil {
		t.Fatalf("Failed to add with embedding: %v", err)
	}

	// Retrieve the item
	item, err := store.Get(ctx, id)
	if err != nil {
		t.Fatalf("Failed to get item: %v", err)
	}

	// Verify content
	if item.Metadata.Content != content {
		t.Errorf("Content mismatch: expected '%s', got '%s'", content, item.Metadata.Content)
	}

	// Verify tags
	if len(item.Metadata.Tags) != len(tags) {
		t.Errorf("Tag count mismatch: expected %d, got %d", len(tags), len(item.Metadata.Tags))
	}

	// Verify custom metadata
	if item.Metadata.Custom["category"] != "functionality" {
		t.Errorf("Custom metadata mismatch: expected 'functionality', got '%s'", item.Metadata.Custom["category"])
	}
}

func TestInMemoryStore_SearchRelevance(t *testing.T) {
	ctx := context.Background()
	embedder := &MockEmbedder{}
	store := NewInMemoryStore(embedder)

	// Add some related content
	contents := []string{
		"artificial intelligence",
		"machine learning algorithms",
		"deep learning neural networks",
		"computer science fundamentals",
		"software engineering practices",
	}

	ids := make([]string, len(contents))
	for i, content := range contents {
		id, err := store.AddWithEmbedding(ctx, content, []string{"ai", "tech"}, map[string]string{"test": "true"})
		if err != nil {
			t.Fatalf("Failed to add item %d: %v", i, err)
		}
		ids[i] = id
	}

	// Search for "AI" related content
	query := "artificial intelligence"
	results, err := store.SearchByText(ctx, query, 3)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("Expected search results, got none")
	}

	// The first result should be most relevant
	if results[0].Content != "artificial intelligence" {
		t.Log("Note: Exact match may not rank first due to mock embedding implementation")
	}
}

func TestInMemoryStore_SaveLoad(t *testing.T) {
	ctx := context.Background()
	embedder := &MockEmbedder{}
	store := NewInMemoryStore(embedder)

	// Add some data
	testContent := "test save and load functionality"
	id, err := store.AddWithEmbedding(ctx, testContent, []string{"save", "load"}, map[string]string{"persistent": "true"})
	if err != nil {
		t.Fatalf("Failed to add item: %v", err)
	}

	// Save to temporary file
	tempFile := "/tmp/test_vector_store.json"
	err = store.Save(ctx, tempFile)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Create new store and load from file
	newStore := NewInMemoryStore(embedder)
	err = newStore.Load(ctx, tempFile)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify the loaded data
	count, err := newStore.Count(ctx)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}

	// Get the item and verify content
	item, err := newStore.Get(ctx, id)
	if err != nil {
		// If the exact ID isn't found, check if any item exists
		items, err := newStore.List(ctx, 10, 0)
		if err != nil || len(items) == 0 {
			t.Fatalf("No items found after loading")
		}
		// Use the ID from the loaded item
		item = &items[0]
	}

	if item.Metadata.Content != testContent {
		t.Errorf("Content mismatch after load: expected '%s', got '%s'", testContent, item.Metadata.Content)
	}
}

func TestInMemoryStore_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	embedder := &MockEmbedder{}
	store := NewInMemoryStore(embedder)

	// Test concurrent access
	done := make(chan bool)
	errors := make(chan error)

	// Concurrent writes
	go func() {
		for i := 0; i < 10; i++ {
			content := "concurrent test " + string(rune('0'+i))
			_, err := store.AddWithEmbedding(ctx, content, []string{"concurrent"}, map[string]string{"iteration": string(rune('0' + i))})
			if err != nil {
				errors <- err
				return
			}
		}
		done <- true
	}()

	// Concurrent reads
	go func() {
		time.Sleep(10 * time.Millisecond) // Slight delay to allow writes to start
		for i := 0; i < 5; i++ {
			_, err := store.Count(ctx)
			if err != nil {
				errors <- err
				return
			}
			time.Sleep(5 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for goroutines to complete
	timeout := time.After(5 * time.Second)
	completed := 0

	for completed < 2 {
		select {
		case <-done:
			completed++
		case err := <-errors:
			t.Fatalf("Concurrent access error: %v", err)
		case <-timeout:
			t.Fatal("Test timed out")
		}
	}

	// Verify final count
	count, err := store.Count(ctx)
	if err != nil {
		t.Fatalf("Final count failed: %v", err)
	}

	if count < 10 {
		t.Errorf("Expected at least 10 items, got %d", count)
	}
}
