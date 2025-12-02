package middleware

import (
	"context"
	"net/http"
	"strings"

	"users-api/internal/responses"
	"users-api/internal/security"
)

type contextKey string

const (
	userIDKey   contextKey = "userID"
	userRoleKey contextKey = "userRole"
	userEmail   contextKey = "userEmail"
)

// AuthMiddleware validates a Bearer token and stores its claims in the request context.
func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				responses.WriteError(w, http.StatusUnauthorized, "AUTHENTICATION_FAILED", "Authorization header is required")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				responses.WriteError(w, http.StatusUnauthorized, "AUTHENTICATION_FAILED", "Authorization header must be Bearer <token>")
				return
			}

			tokenString := strings.TrimSpace(parts[1])
			claims, err := security.ParseToken(tokenString, jwtSecret)
			if err != nil {
				responses.WriteError(w, http.StatusUnauthorized, "AUTHENTICATION_FAILED", "Invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			ctx = context.WithValue(ctx, userRoleKey, claims.Role)
			ctx = context.WithValue(ctx, userEmail, claims.Email)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID extracts the authenticated user ID from context.
func GetUserID(ctx context.Context) (uint, bool) {
	val, ok := ctx.Value(userIDKey).(uint)
	return val, ok
}

// GetUserRole extracts the authenticated user role from context.
func GetUserRole(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(userRoleKey).(string)
	return val, ok
}

// GetUserEmail extracts the authenticated user email from context.
func GetUserEmail(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(userEmail).(string)
	return val, ok
}
