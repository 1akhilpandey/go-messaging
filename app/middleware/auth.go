package middleware

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

// UserContextKey is the key for user information in the request context.
const UserContextKey = "username"

// AuthMiddleware validates the JWT provided in the Authorization header
// and attaches the token claims to the request context.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}
		tokenString := parts[1]

		secretKey := []byte("mysecret")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return secretKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			// Handle the error: Claims are not of expected type (jwt.MapClaims)
			fmt.Println("Unexpected claims type:", reflect.TypeOf(token.Claims))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		username, ok := claims["username"].(string)
		if !ok {
			// Handle the error: Username claim is not a string or not present
			fmt.Println("Username claim not found or not string")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), UserContextKey, username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
