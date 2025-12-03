package services

import (
	"errors"
	"testing"

	"products-api/internal/models"
	"products-api/internal/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCheckoutReducesStockAndCreatesPurchase(t *testing.T) {
	productID := primitive.NewObjectID()
	productRepo := &fakePurchaseProductRepo{
		products: map[string]*models.Product{
			productID.Hex(): {
				ID:          productID,
				Name:        "Test",
				Marca:       "Marca",
				Imagen:      "https://example.com/test.jpg",
				Precio:      50,
				Stock:       5,
				Descripcion: "Fragancia",
				Tipo:        "floral",
				Estacion:    "verano",
				Ocasion:     "dia",
				Genero:      "mujer",
			},
		},
	}
	purchaseRepo := &fakePurchasesRepo{}
	publisher := &mockPublisher{}

	service := NewPurchaseService(productRepo, purchaseRepo, publisher)

	purchase, err := service.Checkout("1", []CheckoutItemInput{{ProductID: productID.Hex(), Cantidad: 2}})
	if err != nil {
		t.Fatalf("checkout returned error: %v", err)
	}
	if purchase == nil {
		t.Fatalf("expected purchase result")
	}
	if purchase.Total != 100 {
		t.Fatalf("expected total 100, got %f", purchase.Total)
	}
	if len(purchase.Items) != 1 || purchase.Items[0].Cantidad != 2 {
		t.Fatalf("expected a single purchase item")
	}
	if productRepo.products[productID.Hex()].Stock != 3 {
		t.Fatalf("expected stock to be reduced to 3")
	}
	if purchaseRepo.created == nil {
		t.Fatalf("expected purchase to be persisted")
	}
	if !publisher.updatedCalled {
		t.Fatalf("expected product.updated event to be emitted")
	}
}

func TestCheckoutFailsOnInsufficientStock(t *testing.T) {
	productID := primitive.NewObjectID()
	productRepo := &fakePurchaseProductRepo{
		products: map[string]*models.Product{
			productID.Hex(): {
				ID:          productID,
				Name:        "Test",
				Precio:      10,
				Stock:       1,
				Descripcion: "Desc",
				Tipo:        "floral",
				Estacion:    "verano",
				Ocasion:     "dia",
				Genero:      "mujer",
			},
		},
	}
	purchaseRepo := &fakePurchasesRepo{}

	service := NewPurchaseService(productRepo, purchaseRepo, nil)

	_, err := service.Checkout("1", []CheckoutItemInput{{ProductID: productID.Hex(), Cantidad: 3}})
	if err == nil {
		t.Fatalf("expected insufficient stock error")
	}
	var stockErr InsufficientStockError
	if !errors.As(err, &stockErr) {
		t.Fatalf("expected InsufficientStockError, got %T", err)
	}
	if purchaseRepo.created != nil {
		t.Fatalf("purchase should not be created on stock errors")
	}
}

// --- fakes ---

type fakePurchaseProductRepo struct {
	products map[string]*models.Product
}

func (f *fakePurchaseProductRepo) Create(p *models.Product) error {
	return nil
}

func (f *fakePurchaseProductRepo) Update(p *models.Product) error {
	f.products[p.ID.Hex()] = p
	return nil
}

func (f *fakePurchaseProductRepo) Delete(id string) error {
	delete(f.products, id)
	return nil
}

func (f *fakePurchaseProductRepo) FindByID(id string) (*models.Product, error) {
	if product, ok := f.products[id]; ok {
		copy := *product
		return &copy, nil
	}
	return nil, repositories.ErrNotFound
}

func (f *fakePurchaseProductRepo) FindAll(filter repositories.ProductFilter, pagination repositories.Pagination) ([]models.Product, int64, error) {
	return nil, 0, nil
}

type fakePurchasesRepo struct {
	created *models.Purchase
}

func (f *fakePurchasesRepo) Create(purchase *models.Purchase) error {
	cpy := *purchase
	f.created = &cpy
	return nil
}

func (f *fakePurchasesRepo) FindByUserID(userID string) ([]models.Purchase, error) {
	if f.created != nil && f.created.UserID == userID {
		return []models.Purchase{*f.created}, nil
	}
	return []models.Purchase{}, nil
}
