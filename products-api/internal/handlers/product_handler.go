package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"products-api/internal/repositories"
	"products-api/internal/responses"
	"products-api/internal/services"
)

// ProductHandler wires HTTP endpoints with the service layer.
type ProductHandler struct {
	service *services.ProductService
}

// NewProductHandler creates a ProductHandler.
func NewProductHandler(service *services.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

type productRequest struct {
	Name        string   `json:"name"`
	Descripcion string   `json:"descripcion"`
	Precio      float64  `json:"precio"`
	Stock       int      `json:"stock"`
	Tipo        string   `json:"tipo"`
	Estacion    string   `json:"estacion"`
	Ocasion     string   `json:"ocasion"`
	Notas       []string `json:"notas"`
	Genero      string   `json:"genero"`
	Marca       string   `json:"marca"`
}

// ListProducts handles GET /products.
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	filter := repositories.ProductFilter{
		Tipo:     query.Get("tipo"),
		Estacion: query.Get("estacion"),
		Ocasion:  query.Get("ocasion"),
		Genero:   query.Get("genero"),
		Marca:    query.Get("marca"),
		Texto:    query.Get("q"),
	}

	page := parseInt(query.Get("page"), 1)
	size := parseInt(query.Get("size"), 10)

	items, pagination, total, err := h.service.ListProducts(filter, repositories.Pagination{
		Page:     page,
		PageSize: size,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	responses.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
		"page":  pagination.Page,
		"size":  pagination.PageSize,
		"total": total,
	})
}

// GetProduct handles GET /products/:id.
func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id, err := extractID(r.URL.Path)
	if err != nil {
		responses.WriteError(w, http.StatusBadRequest, "invalid_id", "Product ID is required")
		return
	}

	product, err := h.service.GetProductByID(id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	responses.WriteJSON(w, http.StatusOK, product)
}

// CreateProduct handles POST /products.
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req productRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responses.WriteError(w, http.StatusBadRequest, "invalid_body", "Invalid JSON payload")
		return
	}

	product, err := h.service.CreateProduct(toInput(req))
	if err != nil {
		handleServiceError(w, err)
		return
	}

	responses.WriteJSON(w, http.StatusCreated, product)
}

// UpdateProduct handles PUT /products/:id.
func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id, err := extractID(r.URL.Path)
	if err != nil {
		responses.WriteError(w, http.StatusBadRequest, "invalid_id", "Product ID is required")
		return
	}

	var req productRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responses.WriteError(w, http.StatusBadRequest, "invalid_body", "Invalid JSON payload")
		return
	}

	product, err := h.service.UpdateProduct(id, toInput(req))
	if err != nil {
		handleServiceError(w, err)
		return
	}

	responses.WriteJSON(w, http.StatusOK, product)
}

// DeleteProduct handles DELETE /products/:id.
func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, err := extractID(r.URL.Path)
	if err != nil {
		responses.WriteError(w, http.StatusBadRequest, "invalid_id", "Product ID is required")
		return
	}

	if err := h.service.DeleteProduct(id); err != nil {
		handleServiceError(w, err)
		return
	}

	responses.WriteJSON(w, http.StatusOK, map[string]string{"message": "product deleted"})
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

func extractID(path string) (string, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return "", http.ErrNoLocation
	}
	id := strings.TrimSpace(parts[1])
	if id == "" {
		return "", http.ErrNoLocation
	}
	return id, nil
}

func toInput(req productRequest) services.CreateProductInput {
	return services.CreateProductInput{
		Name:        req.Name,
		Descripcion: req.Descripcion,
		Precio:      req.Precio,
		Stock:       req.Stock,
		Tipo:        req.Tipo,
		Estacion:    req.Estacion,
		Ocasion:     req.Ocasion,
		Notas:       req.Notas,
		Genero:      req.Genero,
		Marca:       req.Marca,
	}
}

func handleServiceError(w http.ResponseWriter, err error) {
	var valErr services.ValidationError
	if errors.As(err, &valErr) {
		responses.WriteError(w, http.StatusBadRequest, "invalid_input", valErr.Error())
		return
	}

	if errors.Is(err, repositories.ErrNotFound) {
		responses.WriteError(w, http.StatusNotFound, "product_not_found", "Product not found")
		return
	}

	responses.WriteError(w, http.StatusInternalServerError, "internal_error", "Unexpected error")
}
