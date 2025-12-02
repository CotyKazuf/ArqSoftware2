package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

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
		responses.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	var req registerRequest
	if err := decodeJSON(r, &req); err != nil {
		responses.WriteError(w, http.StatusBadRequest, "bad_request", "Invalid request body")
		return
	}

	user, err := h.service.Register(req.Name, req.Email, req.Password)
	if errors.Is(err, services.ErrValidation) {
		responses.WriteError(w, http.StatusBadRequest, "bad_request", "Name, email and password are required")
		return
	}
	if errors.Is(err, services.ErrEmailAlreadyExists) {
		responses.WriteError(w, http.StatusConflict, "email_exists", "Email is already registered")
		return
	}
	if err != nil {
		responses.WriteError(w, http.StatusInternalServerError, "internal_error", "Could not create user")
		return
	}

	responses.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
		"role":  user.Role,
	})
}

// Login handles POST /users/login.
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responses.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		responses.WriteError(w, http.StatusBadRequest, "bad_request", "Invalid request body")
		return
	}

	token, user, err := h.service.Login(req.Email, req.Password)
	if errors.Is(err, services.ErrValidation) {
		responses.WriteError(w, http.StatusBadRequest, "bad_request", "Email and password are required")
		return
	}
	if errors.Is(err, services.ErrInvalidCredentials) {
		responses.WriteError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password")
		return
	}
	if err != nil {
		responses.WriteError(w, http.StatusInternalServerError, "internal_error", "Could not login user")
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
		responses.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		responses.WriteError(w, http.StatusUnauthorized, "auth_missing", "Missing authentication context")
		return
	}

	user, err := h.service.GetByID(userID)
	if errors.Is(err, repositories.ErrUserNotFound) {
		responses.WriteError(w, http.StatusNotFound, "user_not_found", "User not found")
		return
	}
	if err != nil {
		responses.WriteError(w, http.StatusInternalServerError, "internal_error", "Could not fetch user profile")
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
