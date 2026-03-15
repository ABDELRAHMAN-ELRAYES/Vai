package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	UserID string
	jwt.RegisteredClaims
}

type Token struct {
	UserID    string     `json:"user_id"`
	Token     []byte     `json:"token"`
	ExpiredAt time.Time  `json:"expired_at"`
}
