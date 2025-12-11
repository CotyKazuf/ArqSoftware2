package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"users-api/internal/middleware"
	"users-api/internal/models"
	"users-api/internal/repositories"
	"users-api/internal/responses"
	"users-api/internal/services"
)

type UserHandler struct {
	service *services.UserService
}

func NewUserHandler(service *services.UserService) *UserHandler {
	return &UserHandler{service: service}
}

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// Register handles POST /users/register.
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responses.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	var req registerRequest
	if err := decodeJSON(r, &req); err != nil {
		responses.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	user, err := h.service.Register(req.Name, req.Email, req.Password)
	var valErr services.ValidationError
	if errors.As(err, &valErr) {
		responses.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", valErr.Error())
		return
	}
	if errors.Is(err, services.ErrEmailAlreadyExists) {
		responses.WriteError(w, http.StatusConflict, "EMAIL_ALREADY_EXISTS", "Email is already registered")
		return
	}
	if err != nil {
		log.Printf("register user: %v", err)
		responses.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not create user")
		return
	}

	responses.WriteJSON(w, http.StatusCreated, sanitizeUser(user))
}

// Login handles POST /users/login.
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responses.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		responses.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	token, user, err := h.service.Login(req.Email, req.Password)
	var valErr services.ValidationError
	if errors.As(err, &valErr) {
		responses.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", valErr.Error())
		return
	}
	if errors.Is(err, services.ErrInvalidCredentials) {
		responses.WriteError(w, http.StatusUnauthorized, "AUTHENTICATION_FAILED", "Invalid email or password")
		return
	}
	if err != nil {
		log.Printf("login user: %v", err)
		responses.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not login user")
		return
	}

	responses.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"user":  sanitizeUser(user),
	})
}

// Me handles GET /users/me.
func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responses.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		responses.WriteError(w, http.StatusUnauthorized, "AUTHENTICATION_FAILED", "Missing authentication context")
		return
	}

	user, err := h.service.GetByID(userID)
	if errors.Is(err, repositories.ErrUserNotFound) {
		responses.WriteError(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
		return
	}
	if err != nil {
		log.Printf("get profile: %v", err)
		responses.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not fetch user profile")
		return
	}

	responses.WriteJSON(w, http.StatusOK, sanitizeUser(user))
}

// GetUserByID handles GET /users/{id} and allows only the owner or an admin to read profiles.
func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responses.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	targetID, err := extractUserID(r.URL.Path)
	if err != nil {
		responses.WriteError(w, http.StatusBadRequest, "INVALID_ID", "User ID is required")
		return
	}

	requesterID, ok := middleware.GetUserID(r.Context())
	if !ok {
		responses.WriteError(w, http.StatusUnauthorized, "AUTHENTICATION_FAILED", "Authorization header is required")
		return
	}
	role, _ := middleware.GetUserRole(r.Context())
	if role != models.RoleAdmin && requesterID != targetID {
		responses.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Admin role or ownership required")
		return
	}

	user, err := h.service.GetByID(targetID)
	if errors.Is(err, repositories.ErrUserNotFound) {
		responses.WriteError(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
		return
	}
	if err != nil {
		log.Printf("get user by id: %v", err)
		responses.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not fetch user")
		return
	}

	responses.WriteJSON(w, http.StatusOK, sanitizeUser(user))
}

func decodeJSON(r *http.Request, dest interface{}) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dest)
}

func sanitizeUser(user *models.User) userResponse {
	return userResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
	}
}

func extractUserID(path string) (uint, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return 0, http.ErrNoLocation
	}
	idStr := strings.TrimSpace(parts[len(parts)-1])
	if idStr == "" {
		return 0, http.ErrNoLocation
	}
	parsed, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(parsed), nil
}
