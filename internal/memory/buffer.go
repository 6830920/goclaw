package memory

import (
	"container/list"
)

// ConversationBuffer manages short-term conversation memory
type ConversationBuffer struct {
	maxSize int
	buffer  *list.List
	entries map[string]*list.Element
}

// NewConversationBuffer creates a new conversation buffer
func NewConversationBuffer(maxSize int) *ConversationBuffer {
	if maxSize <= 0 {
		maxSize = 50
	}

	return &ConversationBuffer{
		maxSize: maxSize,
		buffer:  list.New(),
		entries: make(map[string]*list.Element),
	}
}

// Add adds a new memory entry
func (cb *ConversationBuffer) Add(entry MemoryEntry) {
	// Remove oldest if at capacity
	if cb.buffer.Len() >= cb.maxSize {
		oldest := cb.buffer.Front()
		if oldest != nil {
			cb.buffer.Remove(oldest)
			delete(cb.entries, oldest.Value.(MemoryEntry).ID)
		}
	}

	elem := cb.buffer.PushBack(entry)
	cb.entries[entry.ID] = elem
}

// GetRecent returns the most recent entries
func (cb *ConversationBuffer) GetRecent(count int) []MemoryEntry {
	results := make([]MemoryEntry, 0, count)

	elem := cb.buffer.Back()
	for elem != nil && len(results) < count {
		results = append(results, elem.Value.(MemoryEntry))
		elem = elem.Prev()
	}

	return results
}

// Remove removes an entry by ID
func (cb *ConversationBuffer) Remove(id string) {
	if elem, exists := cb.entries[id]; exists {
		cb.buffer.Remove(elem)
		delete(cb.entries, id)
	}
}

// Len returns the number of entries
func (cb *ConversationBuffer) Len() int {
	return cb.buffer.Len()
}

// Clear clears all entries
func (cb *ConversationBuffer) Clear() {
	cb.buffer.Init()
	cb.entries = make(map[string]*list.Element)
}
