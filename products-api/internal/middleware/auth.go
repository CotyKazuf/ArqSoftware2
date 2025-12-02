package middleware

import (
	"context"
	"net/http"
	"strings"

	"products-api/internal/responses"
	"products-api/internal/security"
)

type contextKey string

const (
	userIDKey contextKey = "userID"
	userRole  contextKey = "userRole"
	userEmail contextKey = "userEmail"
)

// AuthMiddleware validates a Bearer token and injects claims into the context.
func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				responses.WriteError(w, http.StatusUnauthorized, "auth_missing", "Authorization header is required")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				responses.WriteError(w, http.StatusUnauthorized, "auth_invalid", "Authorization header must be Bearer <token>")
				return
			}

			tokenString := strings.TrimSpace(parts[1])
			claims, err := security.ParseToken(tokenString, jwtSecret)
			if err != nil {
				responses.WriteError(w, http.StatusUnauthorized, "invalid_token", "Invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			ctx = context.WithValue(ctx, userRole, claims.Role)
			ctx = context.WithValue(ctx, userEmail, claims.Email)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAdmin ensures the authenticated user has the admin role.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := GetUserRole(r.Context())
		if !ok || role != "admin" {
			responses.WriteError(w, http.StatusForbidden, "forbidden", "Admin role required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// GetUserRole extracts the role from context.
func GetUserRole(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(userRole).(string)
	return val, ok
}
