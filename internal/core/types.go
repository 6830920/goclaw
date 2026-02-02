// Package core contains the core functionality of Goclaw
package core

import (
	"context"
	"time"
)

// Session represents an agent session
type Session struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	Messages  []Message
	Model     string
	// Additional session properties...
}

// Message represents a message in a session
type Message struct {
	ID        string
	SessionID string
	Role      string // "user", "assistant", "system"
	Content   string
	Timestamp time.Time
}

// Agent manages AI interactions
type Agent struct {
	ID      string
	Name    string
	Model   string
	Context context.Context
}

// Channel represents a communication channel (WhatsApp, Telegram, etc.)
type Channel struct {
	ID          string
	Name        string
	Type        string // "whatsapp", "telegram", "discord", etc.
	Config      map[string]interface{}
	Connected   bool
	LastActive  time.Time
}

// Gateway is the main control plane
type Gateway struct {
	Sessions map[string]*Session
	Channels map[string]*Channel
	Agents   map[string]*Agent
}