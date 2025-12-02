package security

import (
	"errors"
	"time"

	"users-api/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

// Claims wraps JWT registered claims with user-specific information.
type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken creates a signed JWT for the given user.
func GenerateToken(user *models.User, secret string, expirationMinutes int) (string, error) {
	if secret == "" {
		return "", errors.New("missing jwt secret")
	}

	expiresAt := time.Now().Add(time.Duration(expirationMinutes) * time.Minute)
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken validates a token string and extracts its claims.
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
