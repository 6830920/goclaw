package security

import (
	"strings"
	"testing"
	"time"
)

func TestNewSecurityManager(t *testing.T) {
	sm := NewSecurityManager("test-secret")
	if sm == nil {
		t.Fatal("Failed to create SecurityManager")
	}

	if len(sm.tokenSecret) == 0 {
		t.Error("Token secret should not be empty")
	}
}

func TestGenerateAPIKey(t *testing.T) {
	sm := NewSecurityManager("test-secret")

	scopes := []string{"read", "write"}
	key, err := sm.GenerateAPIKey("test-key", scopes, 24*time.Hour)

	if err != nil {
		t.Fatalf("Failed to generate API key: %v", err)
	}

	if key == "" {
		t.Error("Generated key should not be empty")
	}

	if !strings.HasPrefix(key, "goclaw_") {
		t.Error("Key should start with 'goclaw_' prefix")
	}
}

func TestValidateAPIKey(t *testing.T) {
	sm := NewSecurityManager("test-secret")

	scopes := []string{"read", "write"}
	key, err := sm.GenerateAPIKey("test-key", scopes, 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate API key: %v", err)
	}

	// 验证有效的密钥
	apiKey, err := sm.ValidateAPIKey(key)
	if err != nil {
		t.Fatalf("Failed to validate valid API key: %v", err)
	}

	if apiKey.Name != "test-key" {
		t.Errorf("Expected name 'test-key', got '%s'", apiKey.Name)
	}

	// 验证无效的密钥
	_, err = sm.ValidateAPIKey("invalid-key")
	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken for invalid key, got %v", err)
	}
}

func TestCheckScope(t *testing.T) {
	sm := NewSecurityManager("test-secret")

	scopes := []string{"read", "write"}
	key, err := sm.GenerateAPIKey("test-key", scopes, 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate API key: %v", err)
	}

	// 测试有效的scope
	if !sm.CheckScope(key, "read") {
		t.Error("Expected scope 'read' to be valid")
	}

	// 测试无效的scope
	if sm.CheckScope(key, "delete") {
		t.Error("Expected scope 'delete' to be invalid")
	}

	// 测试通配符
	wildcardScopes := []string{"*"}
	wildcardKey, err := sm.GenerateAPIKey("wildcard-key", wildcardScopes, 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate wildcard API key: %v", err)
	}

	if !sm.CheckScope(wildcardKey, "any-scope") {
		t.Error("Wildcard scope should allow any scope")
	}
}

func TestCreateSession(t *testing.T) {
	sm := NewSecurityManager("test-secret")

	session, err := sm.CreateSession("user-123", 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	if session.ID == "" {
		t.Error("Session ID should not be empty")
	}

	if session.UserID != "user-123" {
		t.Errorf("Expected user ID 'user-123', got '%s'", session.UserID)
	}

	if time.Since(session.CreatedAt) > time.Second {
		t.Error("Session creation time should be recent")
	}
}

func TestValidateSession(t *testing.T) {
	sm := NewSecurityManager("test-secret")

	session, err := sm.CreateSession("user-123", 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// 验证有效的会话
	validatedSession, err := sm.ValidateSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to validate valid session: %v", err)
	}

	if validatedSession.UserID != "user-123" {
		t.Errorf("Expected user ID 'user-123', got '%s'", validatedSession.UserID)
	}

	// 验证无效的会话
	_, err = sm.ValidateSession("invalid-session-id")
	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken for invalid session, got %v", err)
	}
}

func TestRefreshSession(t *testing.T) {
	sm := NewSecurityManager("test-secret")

	session, err := sm.CreateSession("user-123", 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	originalExpiry := session.ExpiresAt
	time.Sleep(10 * time.Millisecond)

	refreshedSession, err := sm.RefreshSession(session.ID, 2*time.Hour)
	if err != nil {
		t.Fatalf("Failed to refresh session: %v", err)
	}

	if refreshedSession.ExpiresAt.Before(originalExpiry) {
		t.Error("Refreshed session should have later expiry")
	}
}

func TestRevokeSession(t *testing.T) {
	sm := NewSecurityManager("test-secret")

	session, err := sm.CreateSession("user-123", 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	err = sm.RevokeSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to revoke session: %v", err)
	}

	// 验证会话已被撤销
	_, err = sm.ValidateSession(session.ID)
	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken for revoked session, got %v", err)
	}
}

func TestRevokeAPIKey(t *testing.T) {
	sm := NewSecurityManager("test-secret")

	scopes := []string{"read", "write"}
	key, err := sm.GenerateAPIKey("test-key", scopes, 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate API key: %v", err)
	}

	err = sm.RevokeAPIKey(key)
	if err != nil {
		t.Fatalf("Failed to revoke API key: %v", err)
	}

	// 验证密钥已被撤销
	_, err = sm.ValidateAPIKey(key)
	if err != ErrUnauthorized {
		t.Errorf("Expected ErrUnauthorized for revoked key, got %v", err)
	}
}

func TestCleanupExpired(t *testing.T) {
	sm := NewSecurityManager("test-secret")

	// 创建一个立即过期的会话
	session, err := sm.CreateSession("user-123", 1*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// 创建一个立即过期的API密钥
	scopes := []string{"read"}
	key, err := sm.GenerateAPIKey("test-key", scopes, 1*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to generate API key: %v", err)
	}

	// 等待过期
	time.Sleep(10 * time.Millisecond)

	// 清理过期项
	sm.CleanupExpired()

	// 验证过期会话已被清理
	_, err = sm.ValidateSession(session.ID)
	if err != ErrInvalidToken {
		t.Errorf("Expected expired session to be cleaned up, got %v", err)
	}

	// 验证过期密钥已被清理
	_, err = sm.ValidateAPIKey(key)
	if err != ErrInvalidToken {
		t.Errorf("Expected expired API key to be cleaned up, got %v", err)
	}
}

func TestGetStats(t *testing.T) {
	sm := NewSecurityManager("test-secret")

	stats := sm.GetStats()
	if stats == nil {
		t.Fatal("Stats should not be nil")
	}

	if stats["total_api_keys"].(int) != 0 {
		t.Error("Expected 0 total API keys")
	}

	if stats["total_sessions"].(int) != 0 {
		t.Error("Expected 0 total sessions")
	}

	// 创建一些测试数据
	sm.CreateSession("user-1", 1*time.Hour)
	sm.CreateSession("user-2", 1*time.Hour)

	scopes := []string{"read"}
	sm.GenerateAPIKey("key-1", scopes, 24*time.Hour)

	stats = sm.GetStats()
	if stats["total_sessions"].(int) != 2 {
		t.Errorf("Expected 2 total sessions, got %d", stats["total_sessions"].(int))
	}

	if stats["total_api_keys"].(int) != 1 {
		t.Errorf("Expected 1 total API key, got %d", stats["total_api_keys"].(int))
	}

	if stats["active_api_keys"].(int) != 1 {
		t.Errorf("Expected 1 active API key, got %d", stats["active_api_keys"].(int))
	}
}

func TestListAPIKeys(t *testing.T) {
	sm := NewSecurityManager("test-secret")

	scopes := []string{"read"}
	sm.GenerateAPIKey("key-1", scopes, 24*time.Hour)
	sm.GenerateAPIKey("key-2", scopes, 24*time.Hour)

	keys := sm.ListAPIKeys()
	if len(keys) != 2 {
		t.Errorf("Expected 2 API keys, got %d", len(keys))
	}
}

func TestListSessions(t *testing.T) {
	sm := NewSecurityManager("test-secret")

	sm.CreateSession("user-1", 1*time.Hour)
	sm.CreateSession("user-2", 1*time.Hour)

	sessions := sm.ListSessions()
	if len(sessions) != 2 {
		t.Errorf("Expected 2 sessions, got %d", len(sessions))
	}
}
