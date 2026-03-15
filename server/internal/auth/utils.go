package auth

import (
	"crypto/sha256"

	"github.com/google/uuid"
)

func GenerateRandomToken() string {
	return uuid.New().String()
}

func HashToken(token string) []byte {
	hashed := sha256.Sum256([]byte(token))
	return hashed[:]
}
