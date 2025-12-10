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
		Sort:     parseSortParam(query.Get("sort")),
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

func parseSortParam(raw string) []services.SortField {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []services.SortField{{Field: "updated_at", Order: "desc"}}
	}
	parts := strings.Split(raw, ",")
	sortFields := make([]services.SortField, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		field := part
		order := "asc"
		if idx := strings.Index(part, ":"); idx >= 0 {
			field = strings.TrimSpace(part[:idx])
			order = strings.TrimSpace(part[idx+1:])
		}
		if field == "" {
			continue
		}
		orderLower := strings.ToLower(order)
		if orderLower != "desc" {
			orderLower = "asc"
		}
		sortFields = append(sortFields, services.SortField{
			Field: field,
			Order: orderLower,
		})
	}
	if len(sortFields) == 0 {
		return []services.SortField{{Field: "updated_at", Order: "desc"}}
	}
	return sortFields
}
