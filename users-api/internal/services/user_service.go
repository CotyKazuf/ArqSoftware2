package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"users-api/internal/models"
	"users-api/internal/repositories"
	"users-api/internal/security"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// ValidationError represents a client-facing validation failure.
type ValidationError struct {
	Reason string
}

func (e ValidationError) Error() string {
	return e.Reason
}

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

func (s *UserService) Register(name, username, email, password string) (*models.User, error) {
	name = strings.TrimSpace(name)
	username = strings.TrimSpace(username)
	email = normalizeEmail(email)
	password = strings.TrimSpace(password)

	if err := validateRegisterInput(name, username, email, password); err != nil {
		return nil, err
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
		Username:     username,
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
	email = normalizeEmail(email)
	password = strings.TrimSpace(password)

	if err := validateLoginInput(email, password); err != nil {
		return "", nil, err
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

func (s *UserService) GetByID(ctx context.Context, id uint) (*models.User, error) {
	return s.repo.GetByID(ctx, id)
}

// EnsureAdminUser seeds a default admin if not already present.
func (s *UserService) EnsureAdminUser(name, email, password string) error {
	email = normalizeEmail(email)
	password = strings.TrimSpace(password)
	if err := validateLoginInput(email, password); err != nil {
		return err
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
		Username:     strings.TrimSpace(name),
		Email:        email,
		PasswordHash: hashed,
		Role:         models.RoleAdmin,
	}

	if err := s.repo.Create(admin); err != nil {
		return fmt.Errorf("create admin: %w", err)
	}
	return nil
}

var emailRegex = regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}$`)

func validateRegisterInput(name, username, email, password string) error {
	if strings.TrimSpace(name) == "" {
		return ValidationError{Reason: "name is required"}
	}
	if strings.TrimSpace(username) == "" {
		return ValidationError{Reason: "username is required"}
	}
	if err := validateEmail(email); err != nil {
		return err
	}
	if password == "" {
		return ValidationError{Reason: "password is required"}
	}
	return nil
}

func validateLoginInput(email, password string) error {
	if err := validateEmail(email); err != nil {
		return err
	}
	if password == "" {
		return ValidationError{Reason: "password is required"}
	}
	return nil
}

func validateEmail(email string) error {
	if email == "" {
		return ValidationError{Reason: "email is required"}
	}
	if !emailRegex.MatchString(email) {
		return ValidationError{Reason: "email has invalid format"}
	}
	return nil
}

func normalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}
