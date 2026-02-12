package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/microservices-go/shared/config"
)

// contextKey for auth context
type authContextKey string

const UserContextKey authContextKey = "user"

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

// Authenticate validates JWT and adds user to context
func (a *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"Missing authorization header"}}`, http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"Invalid authorization header format"}}`, http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]
		token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(a.jwtConfig.Secret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"Invalid or expired token"}}`, http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*UserClaims)
		if !ok {
			http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"Invalid token claims"}}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserFromContext extracts user claims from context
func GetUserFromContext(ctx context.Context) (*UserClaims, bool) {
	claims, ok := ctx.Value(UserContextKey).(*UserClaims)
	return claims, ok
}

// RequireRole checks if user has required role
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := GetUserFromContext(r.Context())
			if !ok {
				http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"User not authenticated"}}`, http.StatusUnauthorized)
				return
			}

			for _, role := range roles {
				if claims.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, `{"error":{"code":"FORBIDDEN","message":"Insufficient permissions"}}`, http.StatusForbidden)
		})
	}
}

// GenerateToken creates a new JWT token
func GenerateToken(userID, email, role string, jwtConfig *config.JWTConfig) (string, error) {
	claims := UserClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: jwtConfig.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtConfig.Secret))
}
