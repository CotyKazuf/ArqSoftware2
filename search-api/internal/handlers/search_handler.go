package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"search-api/internal/responses"
	"search-api/internal/services"
)

// SearchHandler wires HTTP endpoints with the service layer.
type SearchHandler struct {
	service *services.SearchService
}

// NewSearchHandler creates a SearchHandler.
func NewSearchHandler(service *services.SearchService) *SearchHandler {
	return &SearchHandler{service: service}
}

// SearchProducts handles GET /search/products.
func (h *SearchHandler) SearchProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responses.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	query := r.URL.Query()
	filters := services.SearchFilters{
		Query:    strings.TrimSpace(query.Get("q")),
		Tipo:     strings.TrimSpace(query.Get("tipo")),
		Estacion: strings.TrimSpace(query.Get("estacion")),
		Ocasion:  strings.TrimSpace(query.Get("ocasion")),
		Genero:   strings.TrimSpace(query.Get("genero")),
		Marca:    strings.TrimSpace(query.Get("marca")),
		Page:     parseInt(query.Get("page"), 1),
		Size:     parseInt(query.Get("size"), 10),
	}

	result, err := h.service.SearchProducts(r.Context(), filters)
	if err != nil {
		responses.WriteError(w, http.StatusInternalServerError, "internal_error", "Could not execute search")
		return
	}

	responses.WriteJSON(w, http.StatusOK, result)
}

// FlushCache handles POST /search/cache/flush to invalidate cached search responses.
func (h *SearchHandler) FlushCache(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responses.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	if err := h.service.FlushCaches(r.Context()); err != nil {
		responses.WriteError(w, http.StatusInternalServerError, "internal_error", "Could not flush caches")
		return
	}

	responses.WriteJSON(w, http.StatusOK, map[string]string{"message": "caches flushed"})
}

func parseInt(value string, def int) int {
	if value == "" {
		return def
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return def
	}
	return parsed
}
