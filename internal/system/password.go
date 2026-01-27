package system

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost = 12
)

// hashPassword hashes a plaintext password using bcrypt.
func hashPassword(plain string) (string, error) {
	if plain == "" {
		return "", errors.New("password cannot be empty")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// verifyPassword compares a bcrypt hash with a plaintext password.
func verifyPassword(storedHash, plain string) bool {
	if storedHash == "" || plain == "" {
		return false
	}

	err := bcrypt.CompareHashAndPassword(
		[]byte(storedHash),
		[]byte(plain),
	)

	return err == nil
}
