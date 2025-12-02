package services

import (
	"errors"
	"fmt"
	"strings"

	"users-api/internal/models"
	"users-api/internal/repositories"
	"users-api/internal/security"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrValidation         = errors.New("validation error")
)

// UserService contains business logic for registration and authentication.
type UserService struct {
	repo          repositories.UserRepository
	jwtSecret     string
	jwtExpiration int
}

func NewUserService(repo repositories.UserRepository, jwtSecret string, jwtExpiration int) *UserService {
	return &UserService{
		repo:          repo,
		jwtSecret:     jwtSecret,
		jwtExpiration: jwtExpiration,
	}
}

func (s *UserService) Register(name, email, password string) (*models.User, error) {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(strings.ToLower(email))
	password = strings.TrimSpace(password)

	if name == "" || email == "" || password == "" {
		return nil, ErrValidation
	}

	existing, err := s.repo.FindByEmail(email)
	if err == nil && existing != nil {
		return nil, ErrEmailAlreadyExists
	}
	if err != nil && !errors.Is(err, repositories.ErrUserNotFound) {
		return nil, err
	}

	hashed, err := security.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &models.User{
		Name:         name,
		Email:        email,
		PasswordHash: hashed,
		Role:         models.RoleNormal,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (s *UserService) Login(email, password string) (string, *models.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	password = strings.TrimSpace(password)

	if email == "" || password == "" {
		return "", nil, ErrValidation
	}

	user, err := s.repo.FindByEmail(email)
	if errors.Is(err, repositories.ErrUserNotFound) {
		return "", nil, ErrInvalidCredentials
	}
	if err != nil {
		return "", nil, err
	}

	if !security.CheckPassword(password, user.PasswordHash) {
		return "", nil, ErrInvalidCredentials
	}

	token, err := security.GenerateToken(user, s.jwtSecret, s.jwtExpiration)
	if err != nil {
		return "", nil, fmt.Errorf("generate token: %w", err)
	}

	return token, user, nil
}

func (s *UserService) GetByID(id uint) (*models.User, error) {
	return s.repo.FindByID(id)
}

// EnsureAdminUser seeds a default admin if not already present.
func (s *UserService) EnsureAdminUser(name, email, password string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || password == "" {
		return ErrValidation
	}

	_, err := s.repo.FindByEmail(email)
	if err == nil {
		return nil // already exists
	}
	if !errors.Is(err, repositories.ErrUserNotFound) {
		return err
	}

	hashed, err := security.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}

	admin := &models.User{
		Name:         name,
		Email:        email,
		PasswordHash: hashed,
		Role:         models.RoleAdmin,
	}

	if err := s.repo.Create(admin); err != nil {
		return fmt.Errorf("create admin: %w", err)
	}
	return nil
}
