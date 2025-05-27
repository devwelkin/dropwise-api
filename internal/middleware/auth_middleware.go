package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/twomotive/dropwise/internal/auth"
	"github.com/twomotive/dropwise/internal/server/httputils"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// UserIDKey is the key used to store the user ID in the request context
const UserIDKey contextKey = "userID"

// AuthMiddleware validates JWT tokens from the Authorization header
// and adds the user ID to the request context
func AuthMiddleware(jwtSecret string) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Get the Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				httputils.RespondWithError(w, http.StatusUnauthorized, "Authorization header required")
				return
			}

			// Check if the header format is correct
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				httputils.RespondWithError(w, http.StatusUnauthorized, "Invalid authorization format, expected 'Bearer TOKEN'")
				return
			}

			// Extract the token
			tokenString := parts[1]

			// Validate the token
			claims, err := auth.ValidateJWT(tokenString, jwtSecret)
			if err != nil {
				httputils.RespondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Invalid or expired token: %v", err))
				return
			}

			// Store user ID in context
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)

			// Call the next handler with the enhanced context
			next(w, r.WithContext(ctx))
		}
	}
}

// GetUserIDFromContext retrieves the user ID from the request context
// Returns the user ID and a boolean indicating if it was found
func GetUserIDFromContext(r *http.Request) (uuid.UUID, bool) {
	userID, ok := r.Context().Value(UserIDKey).(uuid.UUID)
	return userID, ok
}
