package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"products-api/internal/clients"
	"products-api/internal/middleware"
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
	users     UsersClient
}

// ValidationError represents a user-facing validation failure.
type ValidationError struct {
	Code    string
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

// UsersClient describes the subset of users-api needed by the service.
type UsersClient interface {
	GetUserByID(ctx context.Context, id uint, token string) (*clients.UserDTO, error)
}

var (
	ErrOwnerNotFound = errors.New("owner not found")
	ErrForbidden     = errors.New("forbidden")
)

// NewProductService wires a service with its dependencies.
func NewProductService(repo repositories.ProductRepository, publisher EventPublisher, users UsersClient) *ProductService {
	return &ProductService{repo: repo, publisher: publisher, users: users}
}

// CreateProductInput groups the fields required to create a product.
type CreateProductInput struct {
	OwnerID     uint
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
}

// UpdateProductInput mirrors the fields that can be updated.
type UpdateProductInput struct {
	OwnerID     *uint
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
}

// CreateProduct validates and persists a new product.
func (s *ProductService) CreateProduct(ctx context.Context, token string, input CreateProductInput) (*models.Product, error) {
	if err := validateProductInput(input); err != nil {
		return nil, err
	}
	if err := s.authorizeOwnerAction(ctx, input.OwnerID); err != nil {
		return nil, err
	}
	if err := s.ensureOwnerExists(ctx, input.OwnerID, token); err != nil {
		return nil, err
	}

	// Derive slug and tags in parallel to keep the request latency predictable.
	type derivedResult struct {
		task string
		slug string
		tags []string
		err  error
	}

	const (
		taskSlug = "slug"
		taskTags = "tags"
	)

	results := make(chan derivedResult, 2)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		slug, err := generateSlug(input.Name)
		results <- derivedResult{task: taskSlug, slug: slug, err: err}
	}()

	go func() {
		defer wg.Done()
		tags, err := generateTags(input.Descripcion, input.Notas)
		results <- derivedResult{task: taskTags, tags: tags, err: err}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var (
		slug string
		tags []string
	)

	for res := range results {
		if res.err != nil {
			return nil, res.err
		}
		switch res.task {
		case taskSlug:
			slug = res.slug
		case taskTags:
			tags = res.tags
		}
	}

	now := time.Now().UTC()
	product := &models.Product{
		OwnerID:     input.OwnerID,
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
		Slug:        slug,
		Tags:        tags,
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
func (s *ProductService) UpdateProduct(ctx context.Context, token string, id string, input UpdateProductInput) (*models.Product, error) {
	product, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if err := s.authorizeOwnerAction(ctx, product.OwnerID); err != nil {
		return nil, err
	}

	ownerID := product.OwnerID
	if input.OwnerID != nil {
		ownerID = *input.OwnerID
	}

	payload := CreateProductInput{
		OwnerID:     ownerID,
		Name:        input.Name,
		Descripcion: input.Descripcion,
		Precio:      input.Precio,
		Stock:       input.Stock,
		Tipo:        input.Tipo,
		Estacion:    input.Estacion,
		Ocasion:     input.Ocasion,
		Notas:       input.Notas,
		Genero:      input.Genero,
		Marca:       input.Marca,
		Imagen:      input.Imagen,
	}

	if err := validateProductInput(payload); err != nil {
		return nil, err
	}
	if err := s.ensureOwnerExists(ctx, ownerID, token); err != nil {
		return nil, err
	}

	product.OwnerID = ownerID
	product.Name = strings.TrimSpace(payload.Name)
	product.Descripcion = strings.TrimSpace(payload.Descripcion)
	product.Precio = payload.Precio
	product.Stock = payload.Stock
	product.Tipo = payload.Tipo
	product.Estacion = payload.Estacion
	product.Ocasion = payload.Ocasion
	product.Notas = payload.Notas
	product.Genero = payload.Genero
	product.Marca = payload.Marca
	product.Imagen = strings.TrimSpace(payload.Imagen)
	slug, err := generateSlug(payload.Name)
	if err != nil {
		return nil, err
	}
	tags, err := generateTags(payload.Descripcion, payload.Notas)
	if err != nil {
		return nil, err
	}
	product.Slug = slug
	product.Tags = tags
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
func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {
	product, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if err := s.authorizeOwnerAction(ctx, product.OwnerID); err != nil {
		return err
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

func (s *ProductService) authorizeOwnerAction(ctx context.Context, ownerID uint) error {
	if ownerID == 0 {
		return ValidationError{Code: "VALIDATION_ERROR", Message: "owner_id is required"}
	}
	role, _ := middleware.GetUserRole(ctx)
	if role == "admin" {
		return nil
	}

	userIDStr, ok := middleware.GetUserID(ctx)
	if !ok {
		return ErrForbidden
	}
	parsed, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		return ErrForbidden
	}
	if uint(parsed) != ownerID {
		return ErrForbidden
	}
	return nil
}

func (s *ProductService) ensureOwnerExists(ctx context.Context, ownerID uint, token string) error {
	if s.users == nil {
		return errors.New("users client is required")
	}
	if _, err := s.users.GetUserByID(ctx, ownerID, token); err != nil {
		if errors.Is(err, clients.ErrUserNotFound) {
			return ErrOwnerNotFound
		}
		return fmt.Errorf("users lookup: %w", err)
	}
	return nil
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

func generateSlug(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", ValidationError{Code: "VALIDATION_ERROR", Message: "name is required"}
	}

	trimmed = strings.ToLower(trimmed)
	var b strings.Builder
	lastDash := false

	for _, r := range trimmed {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		case r == ' ' || r == '_' || r == '-':
			if !lastDash && b.Len() > 0 {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}

	slug := strings.Trim(b.String(), "-")
	if slug == "" {
		return "", errors.New("unable to generate slug")
	}
	return slug, nil
}

func generateTags(description string, notas []string) ([]string, error) {
	text := strings.ToLower(strings.TrimSpace(description))
	if text == "" {
		return nil, ValidationError{Code: "VALIDATION_ERROR", Message: "descripcion is required"}
	}
	seen := make(map[string]struct{})
	var tags []string

	addTag := func(raw string) {
		tag := strings.ToLower(strings.Trim(strings.TrimSpace(raw), ",.;:!?"))
		if tag == "" {
			return
		}
		if _, ok := seen[tag]; ok {
			return
		}
		seen[tag] = struct{}{}
		tags = append(tags, tag)
	}

	for _, nota := range notas {
		addTag(nota)
	}

	for _, word := range strings.Fields(text) {
		if len(word) < 4 {
			continue
		}
		addTag(word)
		if len(tags) >= 12 {
			break
		}
	}

	return tags, nil
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
	if input.OwnerID == 0 {
		return ValidationError{Code: "VALIDATION_ERROR", Message: "owner_id is required"}
	}
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
