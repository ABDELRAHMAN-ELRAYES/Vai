package auth_test

import (
	"testing"
	"time"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

type TestClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func TestJWTAuthenticator(t *testing.T) {
	secret := "test-secret"
	iss := "test-issuer"
	aud := "test-audience"
	authenticator := auth.NewJWTuthenticator(secret, iss, aud)

	t.Run("generate and validate token", func(t *testing.T) {
		userID := "user-123"
		claims := TestClaims{
			UserID: userID,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    iss,
				Audience:  []string{aud},
			},
		}

		tokenStr, err := authenticator.GenerateToken(claims)
		assert.NoError(t, err)
		assert.NotEmpty(t, tokenStr)

		validatedClaims := &TestClaims{}
		token, err := authenticator.ValidateToken(tokenStr, validatedClaims)
		assert.NoError(t, err)
		assert.True(t, token.Valid)
		assert.Equal(t, userID, validatedClaims.UserID)
		assert.Equal(t, iss, validatedClaims.RegisteredClaims.Issuer)
		assert.Equal(t, aud, validatedClaims.RegisteredClaims.Audience[0])
	})

	t.Run("invalid token - wrong secret", func(t *testing.T) {
		otherAuthenticator := auth.NewJWTuthenticator("wrong-secret", iss, aud)
		claims := TestClaims{
			UserID: "user-123",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				Issuer:    iss,
				Audience:  []string{aud},
			},
		}

		tokenStr, _ := otherAuthenticator.GenerateToken(claims)
		
		validatedClaims := &TestClaims{}
		_, err := authenticator.ValidateToken(tokenStr, validatedClaims)
		assert.Error(t, err)
	})

	t.Run("invalid token - expired", func(t *testing.T) {
		claims := TestClaims{
			UserID: "user-123",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
				Issuer:    iss,
				Audience:  []string{aud},
			},
		}

		tokenStr, _ := authenticator.GenerateToken(claims)
		
		validatedClaims := &TestClaims{}
		_, err := authenticator.ValidateToken(tokenStr, validatedClaims)
		assert.Error(t, err)
	})
}
