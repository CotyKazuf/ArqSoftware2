package services

import (
	"errors"
	"testing"
	"time"

	"products-api/internal/models"
	"products-api/internal/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCreateProductSuccess(t *testing.T) {
	repo := &mockProductRepo{}
	publisher := &mockPublisher{}
	service := NewProductService(repo, publisher)

	input := CreateProductInput{
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

	product, err := service.CreateProduct(input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if product == nil {
		t.Fatalf("expected product, got nil")
	}
	if repo.createCount != 1 {
		t.Fatalf("expected create to be called once")
	}
	if !publisher.createdCalled {
		t.Fatalf("expected publish created to be called")
	}
}

func TestCreateProductValidationError(t *testing.T) {
	repo := &mockProductRepo{}
	service := NewProductService(repo, nil)

	input := CreateProductInput{
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

	_, err := service.CreateProduct(input)
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
	service := NewProductService(repo, publisher)

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

	updated, err := service.UpdateProduct(oid.Hex(), input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Name != "Nuevo" {
		t.Fatalf("expected product to be updated")
	}
	if repo.updateCount != 1 {
		t.Fatalf("expected update to be called once")
	}
	if !publisher.updatedCalled {
		t.Fatalf("expected publish updated to be called")
	}
}

func TestDeleteProduct(t *testing.T) {
	repo := &mockProductRepo{}
	publisher := &mockPublisher{}
	service := NewProductService(repo, publisher)

	err := service.DeleteProduct(primitive.NewObjectID().Hex())
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
