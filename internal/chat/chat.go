// Package chat provides conversation management for Goclaw
package chat

import (
	"fmt"
	"sync"
	"time"
)

// Message represents a chat message
type Message struct {
	Role      string                 `json:"role"` // "user", "assistant", "system"
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// ChatSession manages a single conversation session
type ChatSession struct {
	ID           string
	Messages     []Message
	SystemPrompt string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Metadata     map[string]interface{}
}

// ChatManager manages multiple chat sessions
type ChatManager struct {
	mu        sync.RWMutex
	sessions  map[string]*ChatSession
	maxMemory int
}

// NewChatManager creates a new chat manager
func NewChatManager(maxMemory int) *ChatManager {
	if maxMemory <= 0 {
		maxMemory = 100
	}

	return &ChatManager{
		sessions:  make(map[string]*ChatSession),
		maxMemory: maxMemory,
	}
}

// CreateSession creates a new chat session
func (cm *ChatManager) CreateSession(id, systemPrompt string) *ChatSession {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	session := &ChatSession{
		ID:           id,
		SystemPrompt: systemPrompt,
		Messages:     make([]Message, 0),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Metadata:     make(map[string]interface{}),
	}

	cm.sessions[id] = session
	return session
}

// GetSession retrieves a session
func (cm *ChatManager) GetSession(id string) (*ChatSession, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	session, exists := cm.sessions[id]
	return session, exists
}

// AddMessage adds a message to a session
func (cm *ChatManager) AddMessage(sessionID, role, content string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	session, exists := cm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	message := Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	}

	session.Messages = append(session.Messages, message)
	session.UpdatedAt = time.Now()

	// Prune old messages if needed
	if len(session.Messages) > cm.maxMemory {
		// Keep system prompt (if any) and last N messages
		pruned := make([]Message, 0, cm.maxMemory)

		// Add any system-like messages at the start
		for _, msg := range session.Messages {
			if msg.Role == "system" {
				pruned = append(pruned, msg)
			}
		}

		// Add last N messages
		remaining := cm.maxMemory - len(pruned)
		if remaining > 0 {
			start := len(session.Messages) - remaining
			if start < 0 {
				start = 0
			}
			pruned = append(pruned, session.Messages[start:]...)
		}

		session.Messages = pruned
	}

	return nil
}

// GetMessages returns all messages in a session
func (cm *ChatManager) GetMessages(sessionID string) ([]Message, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	session, exists := cm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return session.Messages, nil
}

// GetConversationText returns the conversation as a single text block
func (cm *ChatManager) GetConversationText(sessionID string) (string, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	session, exists := cm.sessions[sessionID]
	if !exists {
		return "", fmt.Errorf("session not found: %s", sessionID)
	}

	text := ""
	for _, msg := range session.Messages {
		role := msg.Role
		if role == "system" {
			continue // Don't include system prompt in conversation text
		}
		text += fmt.Sprintf("%s: %s\n", role, msg.Content)
	}

	return text, nil
}

// DeleteSession removes a session
func (cm *ChatManager) DeleteSession(id string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.sessions[id]; !exists {
		return fmt.Errorf("session not found: %s", id)
	}

	delete(cm.sessions, id)
	return nil
}

// ListSessions lists all session IDs
func (cm *ChatManager) ListSessions() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	ids := make([]string, 0, len(cm.sessions))
	for id := range cm.sessions {
		ids = append(ids, id)
	}

	return ids
}

// SessionCount returns the number of active sessions
func (cm *ChatManager) SessionCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.sessions)
}
