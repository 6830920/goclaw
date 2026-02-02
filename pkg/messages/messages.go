// Package messages handles message processing for Goclaw
package messages

import (
	"time"
)

// Message represents a message in the OpenClaw system
type Message struct {
	ID        string    `json:"id"`
	SessionID string    `json:"sessionId"`
	Role      string    `json:"role"`      // "user", "assistant", "system"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Metadata  Metadata  `json:"metadata,omitempty"`
}

// Metadata holds additional information about a message
type Metadata struct {
	Channel   string            `json:"channel,omitempty"`
	Author    string            `json:"author,omitempty"`
	ThreadID  string            `json:"threadId,omitempty"`
	Files     []string          `json:"files,omitempty"`
	Tags      []string          `json:"tags,omitempty"`
	Custom    map[string]string `json:"custom,omitempty"`
}

// Session represents a conversation session
type Session struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Messages  []Message `json:"messages"`
	Model     string    `json:"model"`
	Active    bool      `json:"active"`
	Title     string    `json:"title,omitempty"`
}

// Manager handles message and session operations
type Manager struct {
	sessions map[string]*Session
}

// NewManager creates a new message manager
func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]*Session),
	}
}

// CreateSession creates a new session
func (m *Manager) CreateSession(id, model string) *Session {
	session := &Session{
		ID:        id,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Messages:  []Message{},
		Model:     model,
		Active:    true,
	}
	m.sessions[id] = session
	return session
}

// GetSession retrieves a session by ID
func (m *Manager) GetSession(id string) (*Session, bool) {
	session, exists := m.sessions[id]
	return session, exists
}

// AddMessage adds a message to a session
func (m *Manager) AddMessage(sessionID string, role, content string) error {
	session, exists := m.GetSession(sessionID)
	if !exists {
		return ErrSessionNotFound
	}
	
	message := Message{
		ID:        generateID(), // In a real implementation, this would use a proper ID generator
		SessionID: sessionID,
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	}
	
	session.Messages = append(session.Messages, message)
	session.UpdatedAt = time.Now()
	
	return nil
}

// GetMessage retrieves a specific message by ID from a session
func (m *Manager) GetMessage(sessionID, messageID string) (*Message, error) {
	session, exists := m.GetSession(sessionID)
	if !exists {
		return nil, ErrSessionNotFound
	}
	
	for _, msg := range session.Messages {
		if msg.ID == messageID {
			return &msg, nil
		}
	}
	
	return nil, ErrMessageNotFound
}

// ListMessages returns all messages in a session
func (m *Manager) ListMessages(sessionID string) ([]Message, error) {
	session, exists := m.GetSession(sessionID)
	if !exists {
		return nil, ErrSessionNotFound
	}
	
	return session.Messages, nil
}

// generateID generates a unique ID (placeholder implementation)
func generateID() string {
	// In a real implementation, this would use a proper UUID generator
	return "msg_" + time.Now().String()
}

// Errors
var (
	ErrSessionNotFound = &MessageError{"session not found"}
	ErrMessageNotFound = &MessageError{"message not found"}
)

// MessageError represents an error in message operations
type MessageError struct {
	msg string
}

func (e *MessageError) Error() string {
	return e.msg
}