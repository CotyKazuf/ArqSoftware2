package services

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"products-api/internal/models"
	"products-api/internal/repositories"
)

// EventPublisher abstracts the event emission layer.
type EventPublisher interface {
	PublishProductCreated(product *models.Product) error
	PublishProductUpdated(product *models.Product) error
	PublishProductDeleted(id string) error
}

// ProductService coordinates product operations.
type ProductService struct {
	repo      repositories.ProductRepository
	publisher EventPublisher
}

// ValidationError represents a user-facing validation failure.
type ValidationError struct {
	Code    string
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

// NewProductService wires a service with its dependencies.
func NewProductService(repo repositories.ProductRepository, publisher EventPublisher) *ProductService {
	return &ProductService{repo: repo, publisher: publisher}
}

// CreateProductInput groups the fields required to create a product.
type CreateProductInput struct {
	Name        string
	Descripcion string
	Precio      float64
	Stock       int
	Tipo        string
	Estacion    string
	Ocasion     string
	Notas       []string
	Genero      string
	Marca       string
}

// UpdateProductInput mirrors the fields that can be updated.
type UpdateProductInput = CreateProductInput

// CreateProduct validates and persists a new product.
func (s *ProductService) CreateProduct(input CreateProductInput) (*models.Product, error) {
	if err := validateProductInput(input); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	product := &models.Product{
		Name:        strings.TrimSpace(input.Name),
		Descripcion: strings.TrimSpace(input.Descripcion),
		Precio:      input.Precio,
		Stock:       input.Stock,
		Tipo:        input.Tipo,
		Estacion:    input.Estacion,
		Ocasion:     input.Ocasion,
		Notas:       input.Notas,
		Genero:      input.Genero,
		Marca:       input.Marca,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(product); err != nil {
		return nil, err
	}

	if s.publisher != nil {
		if err := s.publisher.PublishProductCreated(product); err != nil {
			log.Printf("publish product.created failed: %v", err)
		}
	}
	return product, nil
}

// UpdateProduct updates a product by id.
func (s *ProductService) UpdateProduct(id string, input UpdateProductInput) (*models.Product, error) {
	product, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if err := validateProductInput(input); err != nil {
		return nil, err
	}

	product.Name = strings.TrimSpace(input.Name)
	product.Descripcion = strings.TrimSpace(input.Descripcion)
	product.Precio = input.Precio
	product.Stock = input.Stock
	product.Tipo = input.Tipo
	product.Estacion = input.Estacion
	product.Ocasion = input.Ocasion
	product.Notas = input.Notas
	product.Genero = input.Genero
	product.Marca = input.Marca
	product.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(product); err != nil {
		return nil, err
	}

	if s.publisher != nil {
		if err := s.publisher.PublishProductUpdated(product); err != nil {
			log.Printf("publish product.updated failed: %v", err)
		}
	}

	return product, nil
}

// DeleteProduct removes a product.
func (s *ProductService) DeleteProduct(id string) error {
	if err := s.repo.Delete(id); err != nil {
		return err
	}

	if s.publisher != nil {
		if err := s.publisher.PublishProductDeleted(id); err != nil {
			log.Printf("publish product.deleted failed: %v", err)
		}
	}

	return nil
}

// GetProductByID fetches a product.
func (s *ProductService) GetProductByID(id string) (*models.Product, error) {
	return s.repo.FindByID(id)
}

// ListProducts obtains products and the total count.
func (s *ProductService) ListProducts(filter repositories.ProductFilter, pagination repositories.Pagination) ([]models.Product, repositories.Pagination, int64, error) {
	page, size := sanitizePagination(pagination.Page, pagination.PageSize)
	pagination.Page = page
	pagination.PageSize = size

	items, total, err := s.repo.FindAll(filter, pagination)
	return items, pagination, total, err
}

func sanitizePagination(page, size int) (int, int) {
	const (
		defaultPageSize = 10
		maxPageSize     = 50
	)
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = defaultPageSize
	}
	if size > maxPageSize {
		size = maxPageSize
	}
	return page, size
}

var (
	allowedTipos = map[string]struct{}{
		"floral":    {},
		"citrico":   {},
		"fresco":    {},
		"amaderado": {},
	}
	allowedEstaciones = map[string]struct{}{
		"verano":    {},
		"otono":     {},
		"invierno":  {},
		"primavera": {},
	}
	allowedOcasion = map[string]struct{}{
		"dia":   {},
		"noche": {},
	}
	allowedGenero = map[string]struct{}{
		"hombre": {},
		"mujer":  {},
		"unisex": {},
	}
	allowedNotas = map[string]struct{}{
		"bergamota": {}, "rosa": {}, "pera": {}, "menta": {}, "lavanda": {}, "sandalo": {},
		"vainilla": {}, "caramelo": {}, "eucalipto": {}, "coco": {}, "jazmin": {}, "mandarina": {},
		"amaderado": {}, "gengibre": {}, "pachuli": {}, "cardamomo": {},
	}
)

func validateProductInput(input CreateProductInput) error {
	if strings.TrimSpace(input.Name) == "" {
		return ValidationError{Code: "VALIDATION_ERROR", Message: "name is required"}
	}
	if strings.TrimSpace(input.Descripcion) == "" {
		return ValidationError{Code: "VALIDATION_ERROR", Message: "descripcion is required"}
	}
	if input.Precio <= 0 {
		return ValidationError{Code: "INVALID_FIELD_VALUE", Message: "precio must be greater than zero"}
	}
	if input.Stock < 0 {
		return ValidationError{Code: "INVALID_FIELD_VALUE", Message: "stock must be zero or greater"}
	}
	if err := validateField("tipo", input.Tipo, allowedTipos); err != nil {
		return err
	}
	if err := validateField("estacion", input.Estacion, allowedEstaciones); err != nil {
		return err
	}
	if err := validateField("ocasion", input.Ocasion, allowedOcasion); err != nil {
		return err
	}
	if err := validateField("genero", input.Genero, allowedGenero); err != nil {
		return err
	}
	if strings.TrimSpace(input.Marca) == "" {
		return ValidationError{Code: "VALIDATION_ERROR", Message: "marca is required"}
	}
	for _, nota := range input.Notas {
		if _, ok := allowedNotas[nota]; !ok {
			return ValidationError{Code: "INVALID_FIELD_VALUE", Message: fmt.Sprintf("nota %s is not supported", nota)}
		}
	}
	return nil
}

func validateField(name, value string, allowed map[string]struct{}) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return ValidationError{Code: "VALIDATION_ERROR", Message: fmt.Sprintf("%s is required", name)}
	}
	if _, ok := allowed[value]; !ok {
		return ValidationError{Code: "INVALID_FIELD_VALUE", Message: fmt.Sprintf("%s has an invalid value", name)}
	}
	return nil
}

// TranslateError exposes repository level errors so handlers can map them.
func TranslateError(err error) string {
	if err == nil {
		return ""
	}
	var valErr ValidationError
	switch {
	case errors.As(err, &valErr):
		return "validation_error"
	case errors.Is(err, repositories.ErrNotFound):
		return "not_found"
	default:
		return "internal_error"
	}
}
