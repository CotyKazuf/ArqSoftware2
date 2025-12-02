package handlers

import (
	"net/http"

	"products-api/internal/responses"
)

// MethodHandler dispatches requests based on HTTP method.
type MethodHandler struct {
	Get    http.Handler
	Post   http.Handler
	Put    http.Handler
	Delete http.Handler
}

// ServeHTTP routes to the handler tied to the request method.
func (mh MethodHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if mh.Get != nil {
			mh.Get.ServeHTTP(w, r)
			return
		}
	case http.MethodPost:
		if mh.Post != nil {
			mh.Post.ServeHTTP(w, r)
			return
		}
	case http.MethodPut:
		if mh.Put != nil {
			mh.Put.ServeHTTP(w, r)
			return
		}
	case http.MethodDelete:
		if mh.Delete != nil {
			mh.Delete.ServeHTTP(w, r)
			return
		}
	}

	responses.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
}
