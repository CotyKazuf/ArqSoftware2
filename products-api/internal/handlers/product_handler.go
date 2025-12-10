package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"products-api/internal/middleware"
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
	OwnerID     *uint    `json:"owner_id,omitempty"`
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
	Imagen      string   `json:"imagen"`
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
		responses.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Product ID is required")
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
	if err := decodeJSON(r, &req); err != nil {
		responses.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON payload")
		return
	}

	owner, err := h.resolveOwner(r.Context(), req.OwnerID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	product, err := h.service.CreateProduct(r.Context(), extractBearerToken(r), toCreateInput(req, owner.value))
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
		responses.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Product ID is required")
		return
	}

	var req productRequest
	if err := decodeJSON(r, &req); err != nil {
		responses.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON payload")
		return
	}

	owner, err := h.resolveOwner(r.Context(), req.OwnerID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	var ownerOverride *uint
	if owner.force {
		value := owner.value
		ownerOverride = &value
	}

	product, err := h.service.UpdateProduct(r.Context(), extractBearerToken(r), id, toUpdateInput(req, ownerOverride))
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
		responses.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Product ID is required")
		return
	}

	if err := h.service.DeleteProduct(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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

func toCreateInput(req productRequest, ownerID uint) services.CreateProductInput {
	return services.CreateProductInput{
		OwnerID:     ownerID,
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
		Imagen:      req.Imagen,
	}
}

func toUpdateInput(req productRequest, ownerID *uint) services.UpdateProductInput {
	return services.UpdateProductInput{
		OwnerID:     ownerID,
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
		Imagen:      req.Imagen,
	}
}

func handleServiceError(w http.ResponseWriter, err error) {
	var valErr services.ValidationError
	if errors.As(err, &valErr) {
		code := valErr.Code
		if code == "" {
			code = "VALIDATION_ERROR"
		}
		responses.WriteError(w, http.StatusBadRequest, code, valErr.Error())
		return
	}

	if errors.Is(err, services.ErrForbidden) {
		responses.WriteError(w, http.StatusForbidden, "FORBIDDEN", "You are not allowed to perform this action")
		return
	}
	if errors.Is(err, services.ErrOwnerNotFound) {
		responses.WriteError(w, http.StatusNotFound, "OWNER_NOT_FOUND", "Owner does not exist")
		return
	}

	if errors.Is(err, repositories.ErrNotFound) {
		responses.WriteError(w, http.StatusNotFound, "PRODUCT_NOT_FOUND", "Product not found")
		return
	}

	log.Printf("product operation failed: %v", err)
	responses.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unexpected error")
}

type ownerResolution struct {
	value uint
	force bool
}

func (h *ProductHandler) resolveOwner(ctx context.Context, requestedOwnerID *uint) (ownerResolution, error) {
	userIDStr, ok := middleware.GetUserID(ctx)
	if !ok {
		return ownerResolution{}, services.ErrForbidden
	}
	parsedUserID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil || parsedUserID == 0 {
		return ownerResolution{}, services.ErrForbidden
	}

	role, _ := middleware.GetUserRole(ctx)
	if role == "admin" {
		if requestedOwnerID != nil {
			return ownerResolution{value: *requestedOwnerID, force: true}, nil
		}
		return ownerResolution{value: uint(parsedUserID), force: false}, nil
	}

	return ownerResolution{value: uint(parsedUserID), force: true}, nil
}

func decodeJSON(r *http.Request, dest interface{}) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dest)
}

func extractBearerToken(r *http.Request) string {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		return ""
	}
	if strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
