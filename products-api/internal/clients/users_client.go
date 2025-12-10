package clients

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// UsersClient performs HTTP requests against users-api.
type UsersClient struct {
	baseURL    string
	httpClient *http.Client
}

// UserDTO mirrors the sanitized payload returned by users-api.
type UserDTO struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// ErrUserNotFound signals that users-api returned a 404 for the requested user.
var ErrUserNotFound = errors.New("user not found")

// NewUsersClient builds a client with sensible defaults.
func NewUsersClient(baseURL string) *UsersClient {
	trimmed := strings.TrimRight(baseURL, "/")
	if trimmed == "" {
		trimmed = "http://localhost:8080"
	}
	return &UsersClient{
		baseURL: trimmed,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetUserByID fetches a user from users-api and optionally forwards an auth token.
func (c *UsersClient) GetUserByID(ctx context.Context, id uint, token string) (*UserDTO, error) {
	url := fmt.Sprintf("%s/users/%d", c.baseURL, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	if token = strings.TrimSpace(token); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("users-api request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var payload struct {
			Data  *UserDTO `json:"data"`
			Error *struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			return nil, fmt.Errorf("decode users-api response: %w", err)
		}
		if payload.Data == nil {
			return nil, errors.New("users-api returned empty data")
		}
		return payload.Data, nil
	case http.StatusNotFound:
		return nil, ErrUserNotFound
	default:
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("users-api unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
}
