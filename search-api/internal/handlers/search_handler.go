package handlers

import (
	"errors"
	"log"
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
		responses.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
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
	if sortParam := strings.TrimSpace(query.Get("sort")); sortParam != "" {
		filters.Sorts = parseSorts(sortParam)
	}

	result, err := h.service.SearchProducts(r.Context(), filters)
	if err != nil {
		var valErr services.ValidationError
		if errors.As(err, &valErr) {
			responses.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", valErr.Error())
			return
		}
		var backendErr services.BackendError
		if errors.As(err, &backendErr) {
			log.Printf("search backend error: %v", backendErr)
			responses.WriteError(w, http.StatusInternalServerError, "SEARCH_BACKEND_ERROR", "No se pudo consultar el índice de búsqueda.")
			return
		}
		log.Printf("search products: %v", err)
		responses.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not execute search")
		return
	}

	responses.WriteJSON(w, http.StatusOK, result)
}

// FlushCache handles POST /search/cache/flush to invalidate cached search responses.
func (h *SearchHandler) FlushCache(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responses.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	if err := h.service.FlushCaches(r.Context()); err != nil {
		log.Printf("flush caches: %v", err)
		responses.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not flush caches")
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

func parseSorts(raw string) []services.SortOption {
	parts := strings.Split(raw, ",")
	var result []services.SortOption
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		desc := false
		field := part
		if strings.HasPrefix(part, "-") {
			desc = true
			field = strings.TrimPrefix(part, "-")
		} else if strings.HasPrefix(part, "+") {
			field = strings.TrimPrefix(part, "+")
		}
		result = append(result, services.SortOption{Field: strings.ToLower(field), Desc: desc})
	}
	return result
}
