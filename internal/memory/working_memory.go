package memory

import (
	"container/heap"
	"sync"
	"time"
)

// WorkingMemory manages active working memory items
type WorkingMemory struct {
	mu    sync.RWMutex
	items WorkingHeap
}

// WorkingHeap is a priority queue for working memory items
type WorkingHeap []WorkingItem

func (wh WorkingHeap) Len() int           { return len(wh) }
func (wh WorkingHeap) Less(i, j int) bool { return wh[i].Priority > wh[j].Priority }
func (wh WorkingHeap) Swap(i, j int)      { wh[i], wh[j] = wh[j], wh[i] }

func (wh *WorkingHeap) Push(x interface{}) {
	*wh = append(*wh, x.(WorkingItem))
}

func (wh *WorkingHeap) Pop() interface{} {
	old := *wh
	n := len(old)
	item := old[n-1]
	old[n-1] = WorkingItem{}
	*wh = old[0 : n-1]
	return item
}

// WorkingItem represents a single working memory item
type WorkingItem struct {
	ID        string
	Content   string
	Priority  int
	Timestamp time.Time
}

// NewWorkingMemory creates a new working memory
func NewWorkingMemory(maxSize int) *WorkingMemory {
	if maxSize <= 0 {
		maxSize = 10
	}

	wm := &WorkingMemory{
		items: make(WorkingHeap, 0, maxSize),
	}
	heap.Init(&wm.items)

	return wm
}

// Add adds a new item to working memory
func (wm *WorkingMemory) Add(entry MemoryEntry) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	priority := 0
	if p, ok := entry.Metadata["priority"].(int); ok {
		priority = p
	}

	item := WorkingItem{
		ID:        entry.ID,
		Content:   entry.Content,
		Priority:  priority,
		Timestamp: entry.Timestamp,
	}

	heap.Push(&wm.items, item)
}

// GetAll returns all working memory items
func (wm *WorkingMemory) GetAll() []MemoryEntry {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	entries := make([]MemoryEntry, len(wm.items))
	for i, item := range wm.items {
		entries[i] = MemoryEntry{
			ID:        item.ID,
			Content:   item.Content,
			Timestamp: item.Timestamp,
		}
	}

	return entries
}

// Len returns the number of items
func (wm *WorkingMemory) Len() int {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return len(wm.items)
}

// Clear clears all items
func (wm *WorkingMemory) Clear() {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.items = make(WorkingHeap, 0)
}
