// Package security provides HTTP middleware for authentication and authorization
package security

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// Context keys for storing security information in request context
type contextKey string

const (
	// APIKeyContextKey is the context key for storing validated API key
	APIKeyContextKey contextKey = "api_key"
	// SessionContextKey is the context key for storing validated session
	SessionContextKey contextKey = "session"
)

// APIKeyAuthMiddleware creates a middleware that validates API keys
func (sm *SecurityManager) APIKeyAuthMiddleware(requiredScope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract API key from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// Also check for X-API-Key header as an alternative
				authHeader = r.Header.Get("X-API-Key")
			}

			if authHeader == "" {
				respondUnauthorized(w, "Missing API key")
				return
			}

			// Remove "Bearer " prefix if present
			apiKey := strings.TrimPrefix(authHeader, "Bearer ")
			apiKey = strings.TrimSpace(apiKey)

			// Validate API key
			validatedKey, err := sm.ValidateAPIKey(apiKey)
			if err != nil {
				log.Printf("API key validation failed: %v", err)
				respondUnauthorized(w, "Invalid API key")
				return
			}

			// Check scope if required
			if requiredScope != "" && !sm.CheckScope(apiKey, requiredScope) {
				respondForbidden(w, "Insufficient permissions")
				return
			}

			// Store validated key in context
			ctx := context.WithValue(r.Context(), APIKeyContextKey, validatedKey)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// SessionAuthMiddleware creates a middleware that validates user sessions
func (sm *SecurityManager) SessionAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract session ID from cookie or header
			sessionID := extractSessionID(r)

			if sessionID == "" {
				respondUnauthorized(w, "Missing session")
				return
			}

			// Validate session
			session, err := sm.ValidateSession(sessionID)
			if err != nil {
				log.Printf("Session validation failed: %v", err)
				respondUnauthorized(w, "Invalid or expired session")
				return
			}

			// Store session in context
			ctx := context.WithValue(r.Context(), SessionContextKey, session)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuthMiddleware creates a middleware that optionally validates authentication
// If authentication is provided, it stores the info in context, but doesn't require it
func (sm *SecurityManager) OptionalAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Try API key authentication first
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				authHeader = r.Header.Get("X-API-Key")
			}

			if authHeader != "" {
				apiKey := strings.TrimPrefix(authHeader, "Bearer ")
				apiKey = strings.TrimSpace(apiKey)

				if validatedKey, err := sm.ValidateAPIKey(apiKey); err == nil {
					ctx = context.WithValue(ctx, APIKeyContextKey, validatedKey)
				}
			}

			// If no API key, try session authentication
			if ctx.Value(APIKeyContextKey) == nil {
				if sessionID := extractSessionID(r); sessionID != "" {
					if session, err := sm.ValidateSession(sessionID); err == nil {
						ctx = context.WithValue(ctx, SessionContextKey, session)
					}
				}
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// CORSMiddleware creates a middleware that handles CORS headers
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			// Set other CORS headers
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400")

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// LoggingMiddleware creates a middleware that logs HTTP requests
func LoggingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("%s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
			next.ServeHTTP(w, r)
		})
	}
}

// RecoveryMiddleware creates a middleware that recovers from panics
func RecoveryMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("Panic recovered: %v", err)
					respondError(w, http.StatusInternalServerError, "Internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// GetAPIKeyFromContext retrieves the validated API key from the request context
func GetAPIKeyFromContext(r *http.Request) *APIKey {
	if apiKey, ok := r.Context().Value(APIKeyContextKey).(*APIKey); ok {
		return apiKey
	}
	return nil
}

// GetSessionFromContext retrieves the validated session from the request context
func GetSessionFromContext(r *http.Request) *Session {
	if session, ok := r.Context().Value(SessionContextKey).(*Session); ok {
		return session
	}
	return nil
}

// extractSessionID extracts session ID from cookie or header
func extractSessionID(r *http.Request) string {
	// Try cookie first
	if cookie, err := r.Cookie("session_id"); err == nil {
		return cookie.Value
	}

	// Try header
	if sessionID := r.Header.Get("X-Session-ID"); sessionID != "" {
		return sessionID
	}

	return ""
}

// Response helpers
func respondUnauthorized(w http.ResponseWriter, message string) {
	respondError(w, http.StatusUnauthorized, message)
}

func respondForbidden(w http.ResponseWriter, message string) {
	respondError(w, http.StatusForbidden, message)
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "error",
		"message": message,
	})
}
