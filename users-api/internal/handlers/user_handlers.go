package handlers

import (
	"encoding/json"
	"errors"
	"log"
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
