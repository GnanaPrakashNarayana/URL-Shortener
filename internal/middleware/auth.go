package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/GnanaPrakashNarayana/url-shortener/internal/models"
	"github.com/GnanaPrakashNarayana/url-shortener/internal/services"
	"github.com/gorilla/sessions"
)

// key type for context
type contextKey string

// UserContextKey is the key for the user in the context
const UserContextKey contextKey = "user"

// AuthMiddleware handles authentication
type AuthMiddleware struct {
	authService  *services.AuthService
	sessionStore *sessions.CookieStore
	sessionName  string
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(authService *services.AuthService, sessionStore *sessions.CookieStore, sessionName string) *AuthMiddleware {
	return &AuthMiddleware{
		authService:  authService,
		sessionStore: sessionStore,
		sessionName:  sessionName,
	}
}

// RequireAuth requires authentication for a handler
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user from context (if already authenticated)
		if r.Context().Value(UserContextKey) != nil {
			next.ServeHTTP(w, r)
			return
		}

		// Redirect to login page
		http.Redirect(w, r, "/auth/login?redirect="+r.URL.Path, http.StatusSeeOther)
	})
}

// RequireRole requires a specific role for a handler
func (m *AuthMiddleware) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user from context
			user, ok := r.Context().Value(UserContextKey).(*models.User)
			if !ok || !user.HasRole(role) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin requires admin role for a handler
func (m *AuthMiddleware) RequireAdmin(next http.Handler) http.Handler {
	return m.RequireRole(models.RoleAdmin)(next)
}

// Auth authenticates a request and puts the user in the context if authenticated
func (m *AuthMiddleware) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var user *models.User
		var err error

		// Try session first
		session, _ := m.sessionStore.Get(r, m.sessionName)
		if tokenInterface, ok := session.Values["token"]; ok {
			if token, ok := tokenInterface.(string); ok {
				user, err = m.authService.ValidateToken(token)
				if err == nil {
					// Token is valid, put user in context
					ctx := context.WithValue(r.Context(), UserContextKey, user)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
				// Token is invalid, remove it from session
				delete(session.Values, "token")
				session.Save(r, w)
			}
		}

		// Try Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			// Extract the token from the Authorization header
			// Format: "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				token := parts[1]
				user, err = m.authService.ValidateToken(token)
				if err == nil {
					// Token is valid, put user in context
					ctx := context.WithValue(r.Context(), UserContextKey, user)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
		}

		// No valid authentication, continue as anonymous
		next.ServeHTTP(w, r)
	})
}

// GetUserFromContext gets the user from the context
func GetUserFromContext(ctx context.Context) *models.User {
	user, ok := ctx.Value(UserContextKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}