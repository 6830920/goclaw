package security

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestAPIKeyAuthMiddleware tests the API key authentication middleware
func TestAPIKeyAuthMiddleware(t *testing.T) {
	sm := NewSecurityManager("test-secret")
	scopes := []string{"read", "write"}
	key, _ := sm.GenerateAPIKey("test-key", scopes, 24*time.Hour)

	// Create a test handler that returns the API key name from context
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := GetAPIKeyFromContext(r)
		if apiKey != nil {
			w.Write([]byte(apiKey.Name))
		} else {
			w.Write([]byte("no-api-key"))
		}
	})

	middleware := sm.APIKeyAuthMiddleware("")

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid API key with Bearer prefix",
			authHeader:     "Bearer " + key,
			expectedStatus: http.StatusOK,
			expectedBody:   "test-key",
		},
		{
			name:           "Valid API key without Bearer prefix",
			authHeader:     key,
			expectedStatus: http.StatusOK,
			expectedBody:   "test-key",
		},
		{
			name:           "Missing API key",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Missing API key",
		},
		{
			name:           "Invalid API key",
			authHeader:     "Bearer invalid-key",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid API key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			handler := middleware(testHandler)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			body := strings.TrimSpace(rr.Body.String())
			if !strings.Contains(body, tt.expectedBody) {
				t.Errorf("Expected body to contain %q, got %q", tt.expectedBody, body)
			}
		})
	}
}

// TestAPIKeyAuthMiddlewareWithScope tests the API key authentication middleware with scope checking
func TestAPIKeyAuthMiddlewareWithScope(t *testing.T) {
	sm := NewSecurityManager("test-secret")
	scopes := []string{"read"}
	key, _ := sm.GenerateAPIKey("read-only-key", scopes, 24*time.Hour)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("success"))
	})

	middleware := sm.APIKeyAuthMiddleware("write")

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+key)

	rr := httptest.NewRecorder()
	handler := middleware(testHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", rr.Code)
	}
}

// TestAPIKeyAuthMiddlewareWithXAPIKey tests authentication using X-API-Key header
func TestAPIKeyAuthMiddlewareWithXAPIKey(t *testing.T) {
	sm := NewSecurityManager("test-secret")
	scopes := []string{"read"}
	key, _ := sm.GenerateAPIKey("test-key", scopes, 24*time.Hour)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := GetAPIKeyFromContext(r)
		if apiKey != nil {
			w.Write([]byte(apiKey.Name))
		}
	})

	middleware := sm.APIKeyAuthMiddleware("")

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", key)

	rr := httptest.NewRecorder()
	handler := middleware(testHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	body := strings.TrimSpace(rr.Body.String())
	if body != "test-key" {
		t.Errorf("Expected body 'test-key', got %q", body)
	}
}

// TestSessionAuthMiddleware tests the session authentication middleware
func TestSessionAuthMiddleware(t *testing.T) {
	sm := NewSecurityManager("test-secret")
	session, _ := sm.CreateSession("user-123", 1*time.Hour)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess := GetSessionFromContext(r)
		if sess != nil {
			w.Write([]byte(sess.UserID))
		} else {
			w.Write([]byte("no-session"))
		}
	})

	middleware := sm.SessionAuthMiddleware()

	tests := []struct {
		name           string
		sessionCookie  *http.Cookie
		sessionHeader  string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid session via cookie",
			sessionCookie:  &http.Cookie{Name: "session_id", Value: session.ID},
			expectedStatus: http.StatusOK,
			expectedBody:   "user-123",
		},
		{
			name:           "Valid session via header",
			sessionHeader:  session.ID,
			expectedStatus: http.StatusOK,
			expectedBody:   "user-123",
		},
		{
			name:           "Missing session",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Missing session",
		},
		{
			name:           "Invalid session",
			sessionHeader:  "invalid-session-id",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid or expired session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.sessionCookie != nil {
				req.AddCookie(tt.sessionCookie)
			}
			if tt.sessionHeader != "" {
				req.Header.Set("X-Session-ID", tt.sessionHeader)
			}

			rr := httptest.NewRecorder()
			handler := middleware(testHandler)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			body := strings.TrimSpace(rr.Body.String())
			if !strings.Contains(body, tt.expectedBody) {
				t.Errorf("Expected body to contain %q, got %q", tt.expectedBody, body)
			}
		})
	}
}

// TestOptionalAuthMiddleware tests the optional authentication middleware
func TestOptionalAuthMiddleware(t *testing.T) {
	sm := NewSecurityManager("test-secret")
	scopes := []string{"read"}
	key, _ := sm.GenerateAPIKey("test-key", scopes, 24*time.Hour)
	session, _ := sm.CreateSession("user-123", 1*time.Hour)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := GetAPIKeyFromContext(r)
		session := GetSessionFromContext(r)

		result := map[string]interface{}{}
		if apiKey != nil {
			result["api_key"] = apiKey.Name
		}
		if session != nil {
			result["session_user"] = session.UserID
		}
		if apiKey == nil && session == nil {
			result["auth"] = "none"
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	middleware := sm.OptionalAuthMiddleware()

	tests := []struct {
		name           string
		authHeader     string
		sessionHeader  string
		expectedHasKey bool
		expectedHasSess bool
	}{
		{
			name:           "With API key",
			authHeader:     "Bearer " + key,
			expectedHasKey: true,
		},
		{
			name:           "With session",
			sessionHeader:  session.ID,
			expectedHasSess: true,
		},
		{
			name:           "Without auth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			if tt.sessionHeader != "" {
				req.Header.Set("X-Session-ID", tt.sessionHeader)
			}

			rr := httptest.NewRecorder()
			handler := middleware(testHandler)
			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", rr.Code)
			}

			var result map[string]interface{}
			json.Unmarshal(rr.Body.Bytes(), &result)

			if tt.expectedHasKey && result["api_key"] == nil {
				t.Error("Expected API key in result")
			}
			if tt.expectedHasSess && result["session_user"] == nil {
				t.Error("Expected session in result")
			}
			if !tt.expectedHasKey && !tt.expectedHasSess && result["auth"] != "none" {
				t.Error("Expected no auth in result")
			}
		})
	}
}

// TestCORSMiddleware tests the CORS middleware
func TestCORSMiddleware(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	middleware := CORSMiddleware([]string{"https://example.com", "http://localhost:*"})

	tests := []struct {
		name               string
		origin             string
		method             string
		expectAllowOrigin  bool
		expectedStatus     int
	}{
		{
			name:              "Allowed origin",
			origin:            "https://example.com",
			method:            "GET",
			expectAllowOrigin: true,
			expectedStatus:    http.StatusOK,
		},
		{
			name:              "Preflight request",
			origin:            "https://example.com",
			method:            "OPTIONS",
			expectAllowOrigin: true,
			expectedStatus:    http.StatusOK,
		},
		{
			name:              "Disallowed origin",
			origin:            "https://evil.com",
			method:            "GET",
			expectAllowOrigin: false,
			expectedStatus:    http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/test", nil)
			req.Header.Set("Origin", tt.origin)

			rr := httptest.NewRecorder()
			handler := middleware(testHandler)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			allowOrigin := rr.Header().Get("Access-Control-Allow-Origin")
			hasAllowOrigin := allowOrigin != ""

			if tt.expectAllowOrigin && !hasAllowOrigin {
				t.Error("Expected Access-Control-Allow-Origin header")
			}
			if !tt.expectAllowOrigin && hasAllowOrigin {
				t.Error("Did not expect Access-Control-Allow-Origin header")
			}
		})
	}
}

// TestRecoveryMiddleware tests the recovery middleware
func TestRecoveryMiddleware(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	middleware := RecoveryMiddleware()

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	handler := middleware(testHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rr.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	if response["status"] != "error" {
		t.Error("Expected error status in response")
	}
}

// TestGetAPIKeyFromContext tests retrieving API key from context
func TestGetAPIKeyFromContext(t *testing.T) {
	sm := NewSecurityManager("test-secret")
	scopes := []string{"read"}
	key, _ := sm.GenerateAPIKey("test-key", scopes, 24*time.Hour)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := GetAPIKeyFromContext(r)
		if apiKey == nil {
			w.Write([]byte("nil"))
		} else {
			w.Write([]byte(apiKey.Key))
		}
	})

	middleware := sm.APIKeyAuthMiddleware("")

	// Test with valid key
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+key)

	rr := httptest.NewRecorder()
	handler := middleware(testHandler)
	handler.ServeHTTP(rr, req)

	body := strings.TrimSpace(rr.Body.String())
	if body != key {
		t.Errorf("Expected key %q, got %q", key, body)
	}

	// Test without key
	req2 := httptest.NewRequest("GET", "/test", nil)
	rr2 := httptest.NewRecorder()
	handler2 := middleware(testHandler)
	handler2.ServeHTTP(rr2, req2)

	body2 := strings.TrimSpace(rr2.Body.String())
	if body2 != "Missing API key" {
		t.Errorf("Expected error message, got %q", body2)
	}
}

// TestGetSessionFromContext tests retrieving session from context
func TestGetSessionFromContext(t *testing.T) {
	sm := NewSecurityManager("test-secret")
	session, _ := sm.CreateSession("user-123", 1*time.Hour)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess := GetSessionFromContext(r)
		if sess == nil {
			w.Write([]byte("nil"))
		} else {
			w.Write([]byte(sess.UserID))
		}
	})

	middleware := sm.SessionAuthMiddleware()

	// Test with valid session
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Session-ID", session.ID)

	rr := httptest.NewRecorder()
	handler := middleware(testHandler)
	handler.ServeHTTP(rr, req)

	body := strings.TrimSpace(rr.Body.String())
	if body != "user-123" {
		t.Errorf("Expected user ID 'user-123', got %q", body)
	}
}
