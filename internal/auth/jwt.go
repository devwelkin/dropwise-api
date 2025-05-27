package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims defines the structure of the JWT claims.
// It includes the standard RegisteredClaims and a custom UserID claim.
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateJWT creates a new JWT string for a given user ID.
// It signs the token using HS256 algorithm with the provided secret key and sets an expiration time.
func GenerateJWT(userID uuid.UUID, secretKey string, expirationDuration time.Duration) (string, error) {
	expirationTime := time.Now().Add(expirationDuration)

	// Create the claims
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID.String(), // Standard claim for user identifier
			Issuer:    "dropwise-api",  // Optional: identifies the issuer of the JWT
		},
	}

	// Create the token using HS256 signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateJWT parses and validates a JWT string.
// It checks the signature, expiration, and other standard claims.
// It returns the custom Claims if the token is valid, otherwise an error.
func ValidateJWT(tokenString string, secretKey string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is what we expect (HS256)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Return the secret key for validation
		return []byte(secretKey), nil
	})

	if err != nil {
		// This will catch errors like expired tokens, malformed tokens, signature mismatch, etc.
		// jwt.ErrTokenExpired, jwt.ErrTokenNotValidYet, jwt.ErrTokenMalformed, etc.
		return nil, fmt.Errorf("failed to parse or validate token: %w", err)
	}

	// Check if the token is valid
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Token is valid, return the claims
	return claims, nil
}
