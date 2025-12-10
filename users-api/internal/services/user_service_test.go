package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	"users-api/internal/models"
	"users-api/internal/repositories"
	"users-api/internal/security"
)

type mockUserRepo struct {
	byEmail map[string]*models.User
	byID    map[uint]*models.User
	nextID  uint
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		byEmail: map[string]*models.User{},
		byID:    map[uint]*models.User{},
		nextID:  1,
	}
}

func (m *mockUserRepo) Create(user *models.User) error {
	if user.ID == 0 {
		user.ID = m.nextID
		m.nextID++
	}
	key := strings.ToLower(user.Email)
	m.byEmail[key] = user
	m.byID[user.ID] = user
	return nil
}

func (m *mockUserRepo) FindByEmail(email string) (*models.User, error) {
	key := strings.ToLower(email)
	if user, ok := m.byEmail[key]; ok {
		return user, nil
	}
	return nil, repositories.ErrUserNotFound
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uint) (*models.User, error) {
	if user, ok := m.byID[id]; ok {
		return user, nil
	}
	return nil, repositories.ErrUserNotFound
}

func TestUserServiceLoginSuccess(t *testing.T) {
	repo := newMockUserRepo()
	hash, _ := security.HashPassword("secret")
	_ = repo.Create(&models.User{
		Name:         "Test",
		Email:        "test@example.com",
		PasswordHash: hash,
		Role:         models.RoleNormal,
	})

	service := NewUserService(repo, "jwtsecret", 15)
	token, user, err := service.Login("test@example.com", "secret")
	if err != nil {
		t.Fatalf("expected login success, got error %v", err)
	}
	if token == "" {
		t.Fatalf("expected token to be returned")
	}
	if user.Email != "test@example.com" {
		t.Fatalf("unexpected user returned")
	}
}

func TestUserServiceLoginInvalidPassword(t *testing.T) {
	repo := newMockUserRepo()
	hash, _ := security.HashPassword("secret")
	_ = repo.Create(&models.User{
		Name:         "Test",
		Email:        "test@example.com",
		PasswordHash: hash,
		Role:         models.RoleNormal,
	})

	service := NewUserService(repo, "jwtsecret", 15)
	_, _, err := service.Login("test@example.com", "wrong")
	if err == nil || err != ErrInvalidCredentials {
		t.Fatalf("expected invalid credentials error, got %v", err)
	}
}

func TestUserServiceLoginUserNotFound(t *testing.T) {
	repo := newMockUserRepo()
	service := NewUserService(repo, "jwtsecret", 15)

	_, _, err := service.Login("missing@example.com", "secret")
	if err == nil || err != ErrInvalidCredentials {
		t.Fatalf("expected invalid credentials error for missing user, got %v", err)
	}
}

func TestUserServiceGetByIDSuccess(t *testing.T) {
	repo := newMockUserRepo()
	user := &models.User{
		Name:         "Jane",
		Username:     "jane",
		Email:        "jane@example.com",
		PasswordHash: "hash",
		Role:         models.RoleNormal,
	}
	if err := repo.Create(user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	service := NewUserService(repo, "jwtsecret", 15)
	found, err := service.GetByID(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if found == nil || found.ID != user.ID {
		t.Fatalf("expected to find user with id %d", user.ID)
	}
}

func TestUserServiceGetByIDNotFound(t *testing.T) {
	repo := newMockUserRepo()
	service := NewUserService(repo, "jwtsecret", 15)

	_, err := service.GetByID(context.Background(), 99)
	if err == nil || !errors.Is(err, repositories.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}
