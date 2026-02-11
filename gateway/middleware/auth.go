package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/microservices-go/shared/config"
)

// UserClaims represents JWT claims
type UserClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// AuthMiddleware validates JWT token
type AuthMiddleware struct {
	jwtConfig *config.JWTConfig
}

// NewAuthMiddleware creates new auth middleware
func NewAuthMiddleware(jwtConfig *config.JWTConfig) *AuthMiddleware {
	return &AuthMiddleware{jwtConfig: jwtConfig}
}

// Middleware returns the auth middleware function
func (a *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// Allow unauthenticated requests (some queries might be public)
			next.ServeHTTP(w, r)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			next.ServeHTTP(w, r)
			return
		}

		tokenString := parts[1]
		token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(a.jwtConfig.Secret), nil
		})

		if err != nil || !token.Valid {
			next.ServeHTTP(w, r)
			return
		}

		// Add auth header to context for forwarding to services
		ctx := context.WithValue(r.Context(), "Authorization", authHeader)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}