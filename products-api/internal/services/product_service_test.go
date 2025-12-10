package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"products-api/internal/clients"
	"products-api/internal/middleware"
	"products-api/internal/models"
	"products-api/internal/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCreateProductSuccess(t *testing.T) {
	repo := &mockProductRepo{}
	publisher := &mockPublisher{}
	usersClient := &mockUsersClient{}
	service := NewProductService(repo, publisher, usersClient)

	input := CreateProductInput{
		OwnerID:     5,
		Name:        "Luna",
		Descripcion: "Notas frescas",
		Precio:      120.5,
		Stock:       10,
		Tipo:        "fresco",
		Estacion:    "verano",
		Ocasion:     "dia",
		Notas:       []string{"bergamota", "menta"},
		Genero:      "unisex",
		Marca:       "Aromas",
		Imagen:      "https://example.com/luna.jpg",
	}

	ctx := middleware.ContextWithUser(context.Background(), 10, "admin", "")
	product, err := service.CreateProduct(ctx, "token", input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if product == nil || product.OwnerID != input.OwnerID {
		t.Fatalf("expected product with owner id %d", input.OwnerID)
	}
	if repo.createCount != 1 {
		t.Fatalf("expected create to be called once")
	}
	if !publisher.createdCalled {
		t.Fatalf("expected publish created to be called")
	}
	if usersClient.callCount != 1 {
		t.Fatalf("expected users client to be consulted")
	}
}

func TestCreateProductValidationError(t *testing.T) {
	repo := &mockProductRepo{}
	service := NewProductService(repo, nil, &mockUsersClient{})

	input := CreateProductInput{
		OwnerID:     1,
		Name:        "",
		Descripcion: "Notas",
		Precio:      -1,
		Stock:       -5,
		Tipo:        "invalid",
		Estacion:    "verano",
		Ocasion:     "dia",
		Genero:      "hombre",
		Marca:       "Marca",
		Imagen:      "ftp://invalid",
	}

	_, err := service.CreateProduct(context.Background(), "", input)
	if err == nil {
		t.Fatalf("expected validation error")
	}
	var valErr ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected validation error type")
	}
	if repo.createCount != 0 {
		t.Fatalf("repo should not be called on validation errors")
	}
}

func TestUpdateProductSuccess(t *testing.T) {
	oid := primitive.NewObjectID()
	existing := &models.Product{
		ID:          oid,
		OwnerID:     7,
		Name:        "Viejo",
		Descripcion: "Desc",
		Precio:      100,
		Stock:       5,
		Tipo:        "floral",
		Estacion:    "primavera",
		Ocasion:     "dia",
		Genero:      "mujer",
		Marca:       "Marca",
		Imagen:      "https://example.com/viejo.jpg",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	repo := &mockProductRepo{
		findProduct: existing,
	}
	publisher := &mockPublisher{}
	usersClient := &mockUsersClient{}
	service := NewProductService(repo, publisher, usersClient)

	newOwnerID := uint(7)
	input := UpdateProductInput{
		OwnerID:     &newOwnerID,
		Name:        "Nuevo",
		Descripcion: "Actualizado",
		Precio:      150,
		Stock:       7,
		Tipo:        "amaderado",
		Estacion:    "otono",
		Ocasion:     "noche",
		Notas:       []string{"sandalo"},
		Genero:      "mujer",
		Marca:       "MarcaX",
		Imagen:      "https://example.com/nuevo.jpg",
	}

	ctx := middleware.ContextWithUser(context.Background(), 7, "normal", "")
	updated, err := service.UpdateProduct(ctx, "token", oid.Hex(), input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Name != "Nuevo" || updated.OwnerID != newOwnerID {
		t.Fatalf("expected product to be updated with new data")
	}
	if repo.updateCount != 1 {
		t.Fatalf("expected update to be called once")
	}
	if !publisher.updatedCalled {
		t.Fatalf("expected publish updated to be called")
	}
	if usersClient.callCount != 1 {
		t.Fatalf("expected users client to be consulted on update")
	}
}

func TestUpdateProductKeepsOwnerWhenNotProvided(t *testing.T) {
	oid := primitive.NewObjectID()
	existing := &models.Product{
		ID:          oid,
		OwnerID:     7,
		Name:        "Viejo",
		Descripcion: "Desc",
		Precio:      100,
		Stock:       5,
		Tipo:        "floral",
		Estacion:    "primavera",
		Ocasion:     "dia",
		Genero:      "mujer",
		Marca:       "Marca",
		Imagen:      "https://example.com/viejo.jpg",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	repo := &mockProductRepo{
		findProduct: existing,
	}
	publisher := &mockPublisher{}
	usersClient := &mockUsersClient{}
	service := NewProductService(repo, publisher, usersClient)

	input := UpdateProductInput{
		Name:        "Nuevo",
		Descripcion: "Actualizado",
		Precio:      150,
		Stock:       7,
		Tipo:        "amaderado",
		Estacion:    "otono",
		Ocasion:     "noche",
		Notas:       []string{"sandalo"},
		Genero:      "mujer",
		Marca:       "MarcaX",
		Imagen:      "https://example.com/nuevo.jpg",
	}

	ctx := middleware.ContextWithUser(context.Background(), 99, "admin", "")
	updated, err := service.UpdateProduct(ctx, "token", oid.Hex(), input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.OwnerID != existing.OwnerID {
		t.Fatalf("expected owner to remain %d, got %d", existing.OwnerID, updated.OwnerID)
	}
	if repo.updateCount != 1 {
		t.Fatalf("expected update to be called once")
	}
	if !publisher.updatedCalled {
		t.Fatalf("expected publish updated to be called")
	}
	if usersClient.callCount != 1 {
		t.Fatalf("expected users client to be consulted on update")
	}
}

func TestDeleteProduct(t *testing.T) {
	oid := primitive.NewObjectID()
	existing := &models.Product{
		ID:      oid,
		OwnerID: 3,
	}
	repo := &mockProductRepo{findProduct: existing}
	publisher := &mockPublisher{}
	service := NewProductService(repo, publisher, &mockUsersClient{})

	ctx := middleware.ContextWithUser(context.Background(), 3, "normal", "")
	err := service.DeleteProduct(ctx, oid.Hex())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.deleteCount != 1 {
		t.Fatalf("expected delete count to be 1")
	}
	if !publisher.deletedCalled {
		t.Fatalf("expected delete event to be published")
	}
}

func TestCreateProductOwnerNotFound(t *testing.T) {
	repo := &mockProductRepo{}
	usersClient := &mockUsersClient{err: clients.ErrUserNotFound}
	service := NewProductService(repo, nil, usersClient)

	input := CreateProductInput{
		OwnerID:     99,
		Name:        "Nuevo",
		Descripcion: "Desc",
		Precio:      10,
		Stock:       1,
		Tipo:        "fresco",
		Estacion:    "verano",
		Ocasion:     "dia",
		Notas:       []string{"menta"},
		Genero:      "unisex",
		Marca:       "Marca",
		Imagen:      "https://example.com/img.jpg",
	}

	ctx := middleware.ContextWithUser(context.Background(), 99, "normal", "")
	_, err := service.CreateProduct(ctx, "token", input)
	if err == nil || !errors.Is(err, ErrOwnerNotFound) {
		t.Fatalf("expected ErrOwnerNotFound, got %v", err)
	}
}

func TestUpdateProductForbiddenForNonOwner(t *testing.T) {
	oid := primitive.NewObjectID()
	existing := &models.Product{
		ID:      oid,
		OwnerID: 10,
	}
	repo := &mockProductRepo{findProduct: existing}
	service := NewProductService(repo, nil, &mockUsersClient{})

	ownerID := uint(10)
	input := UpdateProductInput{
		OwnerID:     &ownerID,
		Name:        "Nuevo",
		Descripcion: "Desc",
		Precio:      50,
		Stock:       5,
		Tipo:        "fresco",
		Estacion:    "verano",
		Ocasion:     "dia",
		Notas:       []string{"menta"},
		Genero:      "unisex",
		Marca:       "Marca",
		Imagen:      "https://example.com/img.jpg",
	}

	ctx := middleware.ContextWithUser(context.Background(), 8, "normal", "")
	_, err := service.UpdateProduct(ctx, "token", oid.Hex(), input)
	if err == nil || !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

type mockProductRepo struct {
	createCount int
	updateCount int
	deleteCount int

	findProduct *models.Product
}

func (m *mockProductRepo) Create(p *models.Product) error {
	m.createCount++
	p.ID = primitive.NewObjectID()
	return nil
}

func (m *mockProductRepo) Update(p *models.Product) error {
	m.updateCount++
	return nil
}

func (m *mockProductRepo) Delete(id string) error {
	m.deleteCount++
	return nil
}

func (m *mockProductRepo) FindByID(id string) (*models.Product, error) {
	if m.findProduct != nil {
		return m.findProduct, nil
	}
	return nil, repositories.ErrNotFound
}

func (m *mockProductRepo) FindAll(filter repositories.ProductFilter, pagination repositories.Pagination) ([]models.Product, int64, error) {
	return []models.Product{}, 0, nil
}

type mockPublisher struct {
	createdCalled bool
	updatedCalled bool
	deletedCalled bool
}

func (m *mockPublisher) PublishProductCreated(product *models.Product) error {
	m.createdCalled = true
	return nil
}

func (m *mockPublisher) PublishProductUpdated(product *models.Product) error {
	m.updatedCalled = true
	return nil
}

func (m *mockPublisher) PublishProductDeleted(id string) error {
	m.deletedCalled = true
	return nil
}

type mockUsersClient struct {
	err       error
	callCount int
}

func (m *mockUsersClient) GetUserByID(ctx context.Context, id uint, token string) (*clients.UserDTO, error) {
	m.callCount++
	if m.err != nil {
		return nil, m.err
	}
	return &clients.UserDTO{ID: id}, nil
}
