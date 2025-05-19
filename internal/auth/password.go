package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword generates a bcrypt hash for the given password.
// The cost parameter for bcrypt.GenerateFromPassword defaults to bcrypt.DefaultCost (10),
// which is generally a good balance between security and performance.
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// CheckPasswordHash compares a plain-text password with a stored bcrypt hash.
// It returns true if the password and hash match, false otherwise.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil // err is nil on success (match), and an error on failure (mismatch or other bcrypt error)
}
