package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rizky/smart-grant/pkg/response"
)

const (
	AuthUserIDKey contextKey = "auth_user_id"
	AuthEmailKey  contextKey = "auth_email"
	AuthRoleKey   contextKey = "auth_role"
)

func Authenticate(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Error(w, http.StatusUnauthorized, "missing_token", "authorization header is required")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				response.Error(w, http.StatusUnauthorized, "invalid_token", "invalid authorization header format")
				return
			}

			tokenStr := parts[1]
			claims := &AuthClaims{}

			token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})
			if err != nil || !token.Valid {
				response.Error(w, http.StatusUnauthorized, "invalid_token", "invalid or expired token")
				return
			}

			if claims.TokenType != "access" {
				response.Error(w, http.StatusUnauthorized, "invalid_token", "invalid token type")
				return
			}

			ctx := context.WithValue(r.Context(), AuthUserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, AuthEmailKey, claims.Email)
			ctx = context.WithValue(ctx, AuthRoleKey, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(AuthRoleKey).(string)
			if !ok {
				response.Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
				return
			}

			for _, allowed := range roles {
				if role == allowed {
					next.ServeHTTP(w, r)
					return
				}
			}

			response.Error(w, http.StatusForbidden, "forbidden", "insufficient permissions")
		})
	}
}

type AuthClaims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}
