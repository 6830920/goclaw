// Package chat provides advanced session management for Goclaw
package chat

import (
	"fmt"
	"sync"
	"time"
)

// SessionState represents the current state of a session
type SessionState string

const (
	SessionStateActive     SessionState = "active"
	SessionStateInactive   SessionState = "inactive"
	SessionStateSuspended  SessionState = "suspended"
	SessionStateArchived   SessionState = "archived"
)

// SessionConfig contains session configuration
type SessionConfig struct {
	ActivationMode    string          // "always", "mention", "auto"
	QueueMode        string          // "queue", "immediate"
	ReplyPolicy      string          // "skip", "announce", "both"
	ThinkingLevel   string          // "off", "minimal", "low", "medium", "high"
	MaxMessages     int             // Maximum messages to keep
	AutoCleanup     bool            // Enable auto-cleanup of old sessions
	GroupRules      map[string]bool // Group-specific rules
}

// EnhancedChatSession provides advanced session capabilities
type EnhancedChatSession struct {
	ID              string
	Messages        []Message
	SystemPrompt    string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Metadata        map[string]interface{}
	State           SessionState
	Config          SessionConfig
	LastActiveTime  time.Time
	MessageCount    int
	TokenUsage      int64
	IsMainSession   bool
	IsGroupSession  bool
	GroupID         string
	UserID          string
	ChannelType     string // "web", "telegram", "whatsapp", etc.
}

// EnhancedChatManager provides advanced session management
type EnhancedChatManager struct {
	mu             sync.RWMutex
	sessions       map[string]*EnhancedChatSession
	mainSessionID  string
	config         SessionConfig
	maxMemory      int
	queue          []Message // For queue mode
}

// NewEnhancedChatManager creates a new enhanced chat manager
func NewEnhancedChatManager(maxMemory int) *EnhancedChatManager {
	if maxMemory <= 0 {
		maxMemory = 100
	}

	return &EnhancedChatManager{
		sessions:      make(map[string]*EnhancedChatSession),
		maxMemory:     maxMemory,
		queue:         make([]Message, 0),
		config: SessionConfig{
			ActivationMode: "always",
			QueueMode:       "immediate",
			ReplyPolicy:     "both",
			ThinkingLevel:  "low",
			MaxMessages:    maxMemory,
			AutoCleanup:    true,
			GroupRules:     make(map[string]bool),
		},
	}
}

// CreateEnhancedSession creates a new enhanced session
func (ecm *EnhancedChatManager) CreateEnhancedSession(id, systemPrompt string, isMain bool) *EnhancedChatSession {
	ecm.mu.Lock()
	defer ecm.mu.Unlock()

	session := &EnhancedChatSession{
		ID:             id,
		SystemPrompt:    systemPrompt,
		Messages:       make([]Message, 0),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Metadata:       make(map[string]interface{}),
		State:          SessionStateActive,
		Config:         ecm.config,
		LastActiveTime: time.Now(),
		MessageCount:   0,
		TokenUsage:      0,
		IsMainSession:  isMain,
		IsGroupSession: false,
		GroupID:        "",
		UserID:         "",
		ChannelType:    "web",
	}

	ecm.sessions[id] = session

	if isMain {
		ecm.mainSessionID = id
	}

	return session
}

// SetSessionState updates session state
func (ecm *EnhancedChatManager) SetSessionState(id string, state SessionState) error {
	ecm.mu.Lock()
	defer ecm.mu.Unlock()

	session, exists := ecm.sessions[id]
	if !exists {
		return fmt.Errorf("session not found: %s", id)
	}

	session.State = state
	session.UpdatedAt = time.Now()

	if state == SessionStateActive {
		session.LastActiveTime = time.Now()
	}

	return nil
}

// GetSessionState returns current session state
func (ecm *EnhancedChatManager) GetSessionState(id string) (SessionState, error) {
	ecm.mu.RLock()
	defer ecm.mu.RUnlock()

	session, exists := ecm.sessions[id]
	if !exists {
		return SessionStateInactive, fmt.Errorf("session not found: %s", id)
	}

	return session.State, nil
}

// SetMainSession marks a session as main session
func (ecm *EnhancedChatManager) SetMainSession(id string) error {
	ecm.mu.Lock()
	defer ecm.mu.Unlock()

	session, exists := ecm.sessions[id]
	if !exists {
		return fmt.Errorf("session not found: %s", id)
	}

	// Update old main session
	if ecm.mainSessionID != "" {
		if oldMain, exists := ecm.sessions[ecm.mainSessionID]; exists {
			oldMain.IsMainSession = false
		}
	}

	// Set new main session
	session.IsMainSession = true
	ecm.mainSessionID = id

	return nil
}

// GetMainSession returns the main session
func (ecm *EnhancedChatManager) GetMainSession() (*EnhancedChatSession, error) {
	ecm.mu.RLock()
	defer ecm.mu.RUnlock()

	if ecm.mainSessionID == "" {
		return nil, fmt.Errorf("no main session set")
	}

	session, exists := ecm.sessions[ecm.mainSessionID]
	if !exists {
		return nil, fmt.Errorf("main session not found: %s", ecm.mainSessionID)
	}

	return session, nil
}

// AddEnhancedMessage adds a message to a session with tracking
func (ecm *EnhancedChatManager) AddEnhancedMessage(sessionID, role, content string) error {
	ecm.mu.Lock()
	defer ecm.mu.Unlock()

	session, exists := ecm.sessions[sessionID]
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
	session.MessageCount++
	session.LastActiveTime = time.Now()

	// Estimate token usage (rough estimate: 4 chars per token)
	session.TokenUsage += int64(len(content) / 4)

	// Prune old messages based on config
	maxMessages := ecm.config.MaxMessages
	if maxMessages <= 0 {
		maxMessages = ecm.maxMemory
	}

	if len(session.Messages) > maxMessages {
		// Keep system messages and last N messages
		pruned := make([]Message, 0)
		for _, msg := range session.Messages {
			if msg.Role == "system" {
				pruned = append(pruned, msg)
			}
		}

		remaining := maxMessages - len(pruned)
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

// GetSessionMetadata returns session metadata
func (ecm *EnhancedChatManager) GetSessionMetadata(id string) (map[string]interface{}, error) {
	ecm.mu.RLock()
	defer ecm.mu.RUnlock()

	session, exists := ecm.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", id)
	}

	metadata := map[string]interface{}{
		"id":              session.ID,
		"state":           session.State,
		"messageCount":    session.MessageCount,
		"tokenUsage":      session.TokenUsage,
		"isMain":         session.IsMainSession,
		"isGroup":        session.IsGroupSession,
		"channel":        session.ChannelType,
		"createdAt":       session.CreatedAt,
		"updatedAt":       session.UpdatedAt,
		"lastActiveTime": session.LastActiveTime,
		"config":         session.Config,
	}

	return metadata, nil
}

// SetSessionConfig updates session configuration
func (ecm *EnhancedChatManager) SetSessionConfig(id string, config SessionConfig) error {
	ecm.mu.Lock()
	defer ecm.mu.Unlock()

	session, exists := ecm.sessions[id]
	if !exists {
		return fmt.Errorf("session not found: %s", id)
	}

	session.Config = config
	session.UpdatedAt = time.Now()

	return nil
}

// CleanupInactiveSessions removes inactive sessions
func (ecm *EnhancedChatManager) CleanupInactiveSessions(maxInactiveTime time.Duration) int {
	ecm.mu.Lock()
	defer ecm.mu.Unlock()

	cleaned := 0
	now := time.Now()

	for id, session := range ecm.sessions {
		if session.State == SessionStateInactive {
			if now.Sub(session.LastActiveTime) > maxInactiveTime {
				delete(ecm.sessions, id)
				cleaned++
			}
		}
	}

	return cleaned
}

// GetActiveSessions returns all active sessions
func (ecm *EnhancedChatManager) GetActiveSessions() []*EnhancedChatSession {
	ecm.mu.RLock()
	defer ecm.mu.RUnlock()

	active := make([]*EnhancedChatSession, 0)
	for _, session := range ecm.sessions {
		if session.State == SessionStateActive {
			active = append(active, session)
		}
	}

	return active
}

// SuspendSession suspends a session temporarily
func (ecm *EnhancedChatManager) SuspendSession(id string) error {
	return ecm.SetSessionState(id, SessionStateSuspended)
}

// ResumeSession resumes a suspended session
func (ecm *EnhancedChatManager) ResumeSession(id string) error {
	return ecm.SetSessionState(id, SessionStateActive)
}

// ArchiveSession archives a session
func (ecm *EnhancedChatManager) ArchiveSession(id string) error {
	return ecm.SetSessionState(id, SessionStateArchived)
}

// GetSessionStatistics returns overall session statistics
func (ecm *EnhancedChatManager) GetSessionStatistics() map[string]interface{} {
	ecm.mu.RLock()
	defer ecm.mu.RUnlock()

	totalMessages := 0
	totalTokens := int64(0)
	activeCount := 0
	groupSessions := 0

	for _, session := range ecm.sessions {
		totalMessages += session.MessageCount
		totalTokens += session.TokenUsage
		if session.State == SessionStateActive {
			activeCount++
		}
		if session.IsGroupSession {
			groupSessions++
		}
	}

	return map[string]interface{}{
		"totalSessions":    len(ecm.sessions),
		"activeSessions":   activeCount,
		"groupSessions":    groupSessions,
		"totalMessages":    totalMessages,
		"totalTokens":      totalTokens,
		"mainSessionID":   ecm.mainSessionID,
		"hasMainSession":   ecm.mainSessionID != "",
	}
}
