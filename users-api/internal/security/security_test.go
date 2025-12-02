package security

import (
	"testing"

	"users-api/internal/models"
)

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("mypassword")
	if err != nil {
		t.Fatalf("expected no error hashing password, got %v", err)
	}

	if !CheckPassword("mypassword", hash) {
		t.Fatalf("expected password check to succeed")
	}

	if CheckPassword("wrong", hash) {
		t.Fatalf("expected password check to fail with wrong password")
	}
}

func TestGenerateAndParseToken(t *testing.T) {
	user := &models.User{
		ID:    1,
		Email: "test@example.com",
		Role:  models.RoleNormal,
	}

	secret := "testsecret"
	token, err := GenerateToken(user, secret, 10)
	if err != nil {
		t.Fatalf("expected no error generating token, got %v", err)
	}

	claims, err := ParseToken(token, secret)
	if err != nil {
		t.Fatalf("expected no error parsing token, got %v", err)
	}

	if claims.UserID != user.ID || claims.Email != user.Email || claims.Role != user.Role {
		t.Fatalf("claims do not match original user data")
	}
}
