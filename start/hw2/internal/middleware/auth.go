package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
)

// ContextKey is a custom type to avoid context key collisions
type ContextKey string

const (
	// UserContextKey is the key used to store the username in the request context
	UserContextKey ContextKey = "username"
)

func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Extract the Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error": "missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			// 2. Extract the token from the header
			// Expected format: "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, `{"error": "invalid authorization header format. Use: Bearer <token>"}`, http.StatusUnauthorized)
				return
			}
			tokenString := parts[1]

			// 3. Parse and validate the token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

				// Verify that the signing method is HMAC
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})

			if err != nil {
				http.Error(w, `{"error": "invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			// 4. Extract the username from the token claims
			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				username, ok := claims["sub"].(string)
				if !ok || username == "" {
					http.Error(w, `{"error": "invalid token claims"}`, http.StatusUnauthorized)
					return
				}

				// 5. Store the username in the request context
				ctx := context.WithValue(r.Context(), UserContextKey, username)
				r = r.WithContext(ctx)

				//6. Call the next handler (actual endpoint)
				next.ServeHTTP(w, r)

			} else {
				http.Error(w, `{"error": "invalid token claims"}`, http.StatusUnauthorized)
				return
			}
		})
	}
}

func GetUsernameFromContext(r *http.Request) (string, bool) {
	username, ok := r.Context().Value(UserContextKey).(string)
	return username, ok
}
