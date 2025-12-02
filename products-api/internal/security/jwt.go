package security

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the expected JWT payload issued by users-api.
type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// ParseToken validates and parses a JWT using the shared secret.
func ParseToken(tokenString, secret string) (*Claims, error) {
	if secret == "" {
		return nil, errors.New("missing jwt secret")
	}

	parsed, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
