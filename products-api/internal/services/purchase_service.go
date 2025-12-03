package services

import (
	"fmt"
	"log"
	"strings"
	"time"

	"products-api/internal/models"
	"products-api/internal/repositories"
)

// CheckoutItemInput represents a product included in the checkout payload.
type CheckoutItemInput struct {
	ProductID string
	Cantidad  int
}

// PurchaseService coordinates checkout operations.
type PurchaseService struct {
	productRepo  repositories.ProductRepository
	purchaseRepo repositories.PurchaseRepository
	publisher    EventPublisher
}

// NewPurchaseService builds a PurchaseService.
func NewPurchaseService(productRepo repositories.ProductRepository, purchaseRepo repositories.PurchaseRepository, publisher EventPublisher) *PurchaseService {
	return &PurchaseService{
		productRepo:  productRepo,
		purchaseRepo: purchaseRepo,
		publisher:    publisher,
	}
}

// InsufficientStockError indicates a product without stock.
type InsufficientStockError struct {
	ProductID string
	Requested int
	Available int
}

func (e InsufficientStockError) Error() string {
	return fmt.Sprintf("insufficient stock for product %s", e.ProductID)
}

// Checkout validates stock, stores a purchase and emits stock updates.
func (s *PurchaseService) Checkout(userID string, items []CheckoutItemInput) (*models.Purchase, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ValidationError{Code: "VALIDATION_ERROR", Message: "user id is required"}
	}

	aggregated, err := aggregateItems(items)
	if err != nil {
		return nil, err
	}
	if len(aggregated) == 0 {
		return nil, ValidationError{Code: "VALIDATION_ERROR", Message: "no purchase items received"}
	}

	type selection struct {
		product  *models.Product
		quantity int
	}

	selected := make([]selection, 0, len(aggregated))
	for _, entry := range aggregated {
		product, err := s.productRepo.FindByID(entry.productID)
		if err != nil {
			return nil, err
		}
		if product.Stock < entry.quantity {
			return nil, InsufficientStockError{
				ProductID: product.ID.Hex(),
				Requested: entry.quantity,
				Available: product.Stock,
			}
		}
		selected = append(selected, selection{
			product:  product,
			quantity: entry.quantity,
		})
	}

	now := time.Now().UTC()
	itemsSnapshot := make([]models.PurchaseItem, 0, len(selected))
	var total float64

	for _, entry := range selected {
		entry.product.Stock -= entry.quantity
		if entry.product.Stock < 0 {
			entry.product.Stock = 0
		}
		entry.product.UpdatedAt = now
		if err := s.productRepo.Update(entry.product); err != nil {
			return nil, err
		}
		if s.publisher != nil {
			if err := s.publisher.PublishProductUpdated(entry.product); err != nil {
				log.Printf("publish product.updated after checkout failed: %v", err)
			}
		}

		itemsSnapshot = append(itemsSnapshot, models.PurchaseItem{
			ProductID:      entry.product.ID,
			Nombre:         entry.product.Name,
			Marca:          entry.product.Marca,
			Imagen:         entry.product.Imagen,
			PrecioUnitario: entry.product.Precio,
			Cantidad:       entry.quantity,
		})
		total += entry.product.Precio * float64(entry.quantity)
	}

	purchase := &models.Purchase{
		UserID:      userID,
		FechaCompra: now,
		Total:       total,
		Items:       itemsSnapshot,
	}

	if err := s.purchaseRepo.Create(purchase); err != nil {
		return nil, err
	}

	return purchase, nil
}

// ListByUser returns the purchase history for a user.
func (s *PurchaseService) ListByUser(userID string) ([]models.Purchase, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ValidationError{Code: "VALIDATION_ERROR", Message: "user id is required"}
	}
	return s.purchaseRepo.FindByUserID(userID)
}

type aggregatedItem struct {
	productID string
	quantity  int
}

func aggregateItems(items []CheckoutItemInput) ([]aggregatedItem, error) {
	if len(items) == 0 {
		return nil, ValidationError{Code: "VALIDATION_ERROR", Message: "items are required"}
	}

	order := make([]aggregatedItem, 0, len(items))
	seen := make(map[string]int)
	for _, item := range items {
		id := strings.TrimSpace(item.ProductID)
		if id == "" {
			return nil, ValidationError{Code: "VALIDATION_ERROR", Message: "producto_id is required"}
		}
		if item.Cantidad <= 0 {
			return nil, ValidationError{Code: "VALIDATION_ERROR", Message: "cantidad must be greater than zero"}
		}
		if idx, ok := seen[id]; ok {
			order[idx].quantity += item.Cantidad
			continue
		}
		seen[id] = len(order)
		order = append(order, aggregatedItem{
			productID: id,
			quantity:  item.Cantidad,
		})
	}
	return order, nil
}
