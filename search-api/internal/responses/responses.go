package responses

import (
	"encoding/json"
	"net/http"
)

// Error is the error envelope used for responses.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Envelope standardizes JSON responses.
type Envelope struct {
	Data  interface{} `json:"data"`
	Error *Error      `json:"error"`
}

// WriteJSON sends a success response using the standard envelope.
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{Data: data, Error: nil})
}

// WriteError sends an error response using the standard envelope.
func WriteError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{
		Data: nil,
		Error: &Error{
			Code:    code,
			Message: message,
		},
	})
}
