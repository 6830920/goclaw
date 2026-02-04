package security

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ErrUnauthorized 权限不足错误
var ErrUnauthorized = errors.New("unauthorized access")

// ErrInvalidToken 无效令牌错误
var ErrInvalidToken = errors.New("invalid token")

// SecurityManager 安全管理器
type SecurityManager struct {
	mu          sync.RWMutex
	apiKeys     map[string]APIKey
	sessions    map[string]*Session
	tokenSecret []byte
}

// APIKey API密钥信息
type APIKey struct {
	Key        string    `json:"key"`
	Name       string    `json:"name"`
	Scopes     []string  `json:"scopes"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	LastUsedAt time.Time `json:"last_used_at"`
	Active     bool      `json:"active"`
}

// Session 会话信息
type Session struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	CreatedAt time.Time              `json:"created_at"`
	ExpiresAt time.Time              `json:"expires_at"`
	LastSeen  time.Time              `json:"last_seen"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// NewSecurityManager 创建安全管理器
func NewSecurityManager(secret string) *SecurityManager {
	if secret == "" {
		secret = generateSecret()
	}

	return &SecurityManager{
		apiKeys:     make(map[string]APIKey),
		sessions:    make(map[string]*Session),
		tokenSecret: []byte(secret),
	}
}

// generateSecret 生成随机密钥
func generateSecret() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based secret
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// GenerateAPIKey 生成API密钥
func (sm *SecurityManager) GenerateAPIKey(name string, scopes []string, ttl time.Duration) (string, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	key := generateKey()
	expiresAt := time.Now().Add(ttl)

	apiKey := APIKey{
		Key:        key,
		Name:       name,
		Scopes:     scopes,
		CreatedAt:  time.Now(),
		ExpiresAt:  expiresAt,
		LastUsedAt: time.Time{},
		Active:     true,
	}

	sm.apiKeys[key] = apiKey
	return key, nil
}

// generateKey 生成API密钥字符串
func generateKey() string {
	prefix := "goclaw_" + time.Now().Format("20060102")
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		randomBytes = []byte(strings.Repeat("x", 16))
	}
	return prefix + "_" + hex.EncodeToString(randomBytes)
}

// ValidateAPIKey 验证API密钥
func (sm *SecurityManager) ValidateAPIKey(key string) (*APIKey, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	apiKey, exists := sm.apiKeys[key]
	if !exists {
		return nil, ErrInvalidToken
	}

	if !apiKey.Active {
		return nil, ErrUnauthorized
	}

	if time.Now().After(apiKey.ExpiresAt) {
		return nil, ErrInvalidToken
	}

	// 更新最后使用时间
	apiKey.LastUsedAt = time.Now()
	sm.apiKeys[key] = apiKey

	return &apiKey, nil
}

// CheckScope 检查API密钥是否有指定权限
func (sm *SecurityManager) CheckScope(key string, requiredScope string) bool {
	apiKey, err := sm.ValidateAPIKey(key)
	if err != nil {
		return false
	}

	for _, scope := range apiKey.Scopes {
		if scope == requiredScope || scope == "*" {
			return true
		}
	}

	return false
}

// CreateSession 创建会话
func (sm *SecurityManager) CreateSession(userID string, ttl time.Duration) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sessionID := generateSecret()
	expiresAt := time.Now().Add(ttl)

	session := &Session{
		ID:        sessionID,
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
		LastSeen:  time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	sm.sessions[sessionID] = session
	return session, nil
}

// ValidateSession 验证会话
func (sm *SecurityManager) ValidateSession(sessionID string) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, ErrInvalidToken
	}

	if time.Now().After(session.ExpiresAt) {
		// 清理过期会话
		delete(sm.sessions, sessionID)
		return nil, ErrInvalidToken
	}

	// 更新最后访问时间
	session.LastSeen = time.Now()

	return session, nil
}

// RefreshSession 刷新会话
func (sm *SecurityManager) RefreshSession(sessionID string, ttl time.Duration) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, ErrInvalidToken
	}

	session.ExpiresAt = time.Now().Add(ttl)
	session.LastSeen = time.Now()

	return session, nil
}

// RevokeSession 撤销会话
func (sm *SecurityManager) RevokeSession(sessionID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.sessions[sessionID]; !exists {
		return ErrInvalidToken
	}

	delete(sm.sessions, sessionID)
	return nil
}

// RevokeAPIKey 撤销API密钥
func (sm *SecurityManager) RevokeAPIKey(key string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.apiKeys[key]; !exists {
		return ErrInvalidToken
	}

	apiKey := sm.apiKeys[key]
	apiKey.Active = false
	sm.apiKeys[key] = apiKey

	return nil
}

// ListAPIKeys 列出所有API密钥
func (sm *SecurityManager) ListAPIKeys() []APIKey {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	keys := make([]APIKey, 0, len(sm.apiKeys))
	for _, key := range sm.apiKeys {
		keys = append(keys, key)
	}

	return keys
}

// ListSessions 列出所有会话
func (sm *SecurityManager) ListSessions() []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessions := make([]*Session, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		sessions = append(sessions, session)
	}

	return sessions
}

// CleanupExpired 清理过期的API密钥和会话
func (sm *SecurityManager) CleanupExpired() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()

	// 清理过期会话
	for sessionID, session := range sm.sessions {
		if now.After(session.ExpiresAt) {
			delete(sm.sessions, sessionID)
		}
	}

	// 清理过期API密钥
	for key, apiKey := range sm.apiKeys {
		if now.After(apiKey.ExpiresAt) {
			delete(sm.apiKeys, key)
		}
	}
}

// GetStats 获取统计信息
func (sm *SecurityManager) GetStats() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	activeKeys := 0
	for _, key := range sm.apiKeys {
		if key.Active {
			activeKeys++
		}
	}

	return map[string]interface{}{
		"total_api_keys":    len(sm.apiKeys),
		"active_api_keys":   activeKeys,
		"total_sessions":    len(sm.sessions),
		"cleanup_needed":    len(sm.sessions) > 0 || len(sm.apiKeys) > 0,
	}
}
