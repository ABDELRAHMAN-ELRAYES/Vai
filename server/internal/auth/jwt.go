package auth

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type JWTAuthenticator struct {
	secret string
	iss    string
	aud    string
}

func NewJWTuthenticator(secret, iss, aud string) *JWTAuthenticator {
	return &JWTAuthenticator{
		secret: secret,
		iss:    iss,
		aud:    aud,
	}
}
func (authenticator *JWTAuthenticator) GenerateToken(claims jwt.Claims) (string, error) {
	// Generate a JWT with configured claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Convert the token to the string format
	strToken, err := token.SignedString([]byte(authenticator.secret))
	if err != nil {
		return "", err
	}
	return strToken, nil
}

func (authenticator *JWTAuthenticator) ValidateToken(strToken string, claims jwt.Claims) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(strToken, claims, func(t *jwt.Token) (any, error) {
		// check if the used signing method had changed or replaced
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		return []byte(authenticator.secret), nil
	},
		jwt.WithExpirationRequired(),
		jwt.WithAudience(authenticator.aud),
		jwt.WithIssuer(authenticator.iss),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)
	if err != nil {
		return nil, err
	}
	return token, nil
}
