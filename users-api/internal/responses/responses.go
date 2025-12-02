package responses

import (
	"encoding/json"
	"net/http"
)

// Error represents the error envelope used in HTTP responses.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Envelope standardizes the shape of every JSON response.
type Envelope struct {
	Data  interface{} `json:"data"`
	Error *Error      `json:"error"`
}

// WriteJSON writes a successful JSON response with the standard envelope.
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{Data: data, Error: nil})
}

// WriteError writes an error response using the standard envelope.
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
