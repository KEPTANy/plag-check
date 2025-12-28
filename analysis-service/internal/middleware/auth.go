package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/KEPTANy/plag-check/shared/jwt"
	"github.com/gofrs/uuid/v5"
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	UsernameKey contextKey = "username"
	RoleKey     contextKey = "role"
)

func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/health") {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error": "authorization header required"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, `{"error": "invalid authorization header format"}`, http.StatusUnauthorized)
				return
			}

			token := parts[1]

			claims, err := jwt.GetTokenClaims(token, jwtSecret)
			if err != nil {
				http.Error(w, `{"error": "invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UsernameKey, claims.Username)
			ctx = context.WithValue(ctx, RoleKey, claims.Role)

			r.Header.Set("X-User-ID", claims.UserID.String())
			r.Header.Set("X-User-Role", claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(RoleKey).(string)
			if !ok || role != requiredRole {
				http.Error(w, `{"error": "insufficient permissions"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return userID, ok
}

func GetRoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(RoleKey).(string)
	return role, ok
}

func GetUsernameFromContext(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(UsernameKey).(string)
	return username, ok
}
