package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
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
	repo         repositories.ProductRepository
	publisher    EventPublisher
	usersAPIURL  string
	httpClient   *http.Client
	requestTimeout time.Duration
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
func NewProductService(repo repositories.ProductRepository, publisher EventPublisher, usersAPIURL string) *ProductService {
	return &ProductService{
		repo:          repo,
		publisher:     publisher,
		usersAPIURL:   strings.TrimRight(usersAPIURL, "/"),
		httpClient:    &http.Client{Timeout: 4 * time.Second},
		requestTimeout: 4 * time.Second,
	}
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
	Imagen      string
	OwnerID     string
}

// UpdateProductInput mirrors the fields that can be updated.
type UpdateProductInput = CreateProductInput

// CreateProduct validates and persists a new product.
func (s *ProductService) CreateProduct(input CreateProductInput, authToken string) (*models.Product, error) {
	if err := validateProductInput(input); err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.OwnerID) == "" {
		return nil, ValidationError{Code: "VALIDATION_ERROR", Message: "owner_id is required"}
	}
	if err := s.ensureUserExists(input.OwnerID, authToken); err != nil {
		return nil, err
	}

	score := computeProductScore(input)

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
		Imagen:      strings.TrimSpace(input.Imagen),
		OwnerID:     input.OwnerID,
		Score:       score,
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
func (s *ProductService) UpdateProduct(id string, requesterID string, isAdmin bool, authToken string, input UpdateProductInput) (*models.Product, error) {
	product, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if !isAdmin && product.OwnerID != "" && product.OwnerID != requesterID {
		return nil, ValidationError{Code: "FORBIDDEN", Message: "Admin role or ownership required"}
	}

	if err := validateProductInput(input); err != nil {
		return nil, err
	}

	ownerToValidate := product.OwnerID
	if ownerToValidate == "" {
		ownerToValidate = requesterID
	}
	if err := s.ensureUserExists(ownerToValidate, authToken); err != nil {
		return nil, err
	}

	score := computeProductScore(input)

	if strings.TrimSpace(product.OwnerID) == "" {
		product.OwnerID = requesterID
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
	product.Imagen = strings.TrimSpace(input.Imagen)
	product.Score = score
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
func (s *ProductService) DeleteProduct(id string, requesterID string, isAdmin bool, authToken string) error {
	product, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if product.OwnerID == "" && !isAdmin {
		return ValidationError{Code: "FORBIDDEN", Message: "Admin role or ownership required"}
	}
	if !isAdmin && product.OwnerID != "" && product.OwnerID != requesterID {
		return ValidationError{Code: "FORBIDDEN", Message: "Admin role or ownership required"}
	}
	if product.OwnerID != "" {
		if err := s.ensureUserExists(product.OwnerID, authToken); err != nil {
			return err
		}
	}

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
	if err := validateImageURL(input.Imagen); err != nil {
		return err
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

func validateImageURL(raw string) error {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ValidationError{Code: "VALIDATION_ERROR", Message: "imagen is required"}
	}
	if !(strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://")) {
		return ValidationError{Code: "VALIDATION_ERROR", Message: "imagen must start with http:// or https://"}
	}
	return nil
}

func (s *ProductService) ensureUserExists(userID string, authToken string) error {
	if s.usersAPIURL == "" {
		return ValidationError{Code: "VALIDATION_ERROR", Message: "users API url is not configured"}
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.requestTimeout)
	defer cancel()

	url := fmt.Sprintf("%s/users/%s", s.usersAPIURL, strings.TrimSpace(userID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build users-api request: %w", err)
	}
	if strings.TrimSpace(authToken) != "" {
		req.Header.Set("Authorization", authToken)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("validate owner against users-api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return ValidationError{Code: "VALIDATION_ERROR", Message: "owner does not exist"}
		}
		return fmt.Errorf("users-api returned %d", resp.StatusCode)
	}
	return nil
}

// computeProductScore simulates a small concurrent computation that aggregates metrics.
func computeProductScore(input CreateProductInput) float64 {
	type partial struct {
		value float64
	}
	results := make(chan partial, 2)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		priceScore := 1.0
		switch {
		case input.Precio > 200:
			priceScore = 0.6
		case input.Precio > 120:
			priceScore = 0.8
		default:
			priceScore = 1.0
		}
		results <- partial{value: priceScore}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		diversityScore := 0.5 + float64(len(input.Notas))/10.0
		results <- partial{value: diversityScore}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var total float64
	for res := range results {
		total += res.value
	}
	// simple average of two contributors
	return total / 2
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
