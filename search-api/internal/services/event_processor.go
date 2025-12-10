package services

import (
	"context"
	"errors"
	"fmt"
	"log"

	"search-api/internal/clients"
	"search-api/internal/models"
	"search-api/internal/rabbitmq"
)

// ProductLookup abstracts products-api so EventProcessor can fetch the latest data.
type ProductLookup interface {
	GetProductByID(ctx context.Context, id string) (*clients.ProductDTO, error)
}

// EventProcessor implements rabbitmq.EventHandler and delegates to SearchService.
type EventProcessor struct {
	service  *SearchService
	products ProductLookup
}

// NewEventProcessor builds an EventProcessor.
func NewEventProcessor(service *SearchService, products ProductLookup) *EventProcessor {
	return &EventProcessor{service: service, products: products}
}

// HandleProductEvent processes product events emitted by products-api.
func (p *EventProcessor) HandleProductEvent(ctx context.Context, event rabbitmq.ProductEvent) error {
	switch event.Type {
	case rabbitmq.EventProductCreated, rabbitmq.EventProductUpdated:
		doc, err := p.fetchProductDocument(ctx, event.ProductID)
		if err != nil {
			if errors.Is(err, clients.ErrProductNotFound) {
				log.Printf("product %s not found on %s event, skipping index", event.ProductID, event.Type)
				return nil
			}
			return err
		}
		if err := p.service.IndexProduct(ctx, *doc); err != nil {
			return fmt.Errorf("index product: %w", err)
		}
	case rabbitmq.EventProductDeleted:
		if err := p.service.DeleteProduct(ctx, event.ProductID); err != nil {
			return fmt.Errorf("delete product: %w", err)
		}
	default:
		return fmt.Errorf("unsupported event type %s", event.Type)
	}
	return nil
}

func (p *EventProcessor) fetchProductDocument(ctx context.Context, id string) (*models.ProductDocument, error) {
	if p.products == nil {
		return nil, errors.New("products client is not configured")
	}
	product, err := p.products.GetProductByID(ctx, id)
	if err != nil {
		return nil, err
	}
	doc := models.ProductDocument{
		ID:          product.ID,
		Name:        product.Name,
		Descripcion: product.Descripcion,
		Precio:      product.Precio,
		Stock:       product.Stock,
		Tipo:        product.Tipo,
		Estacion:    product.Estacion,
		Ocasion:     product.Ocasion,
		Notas:       product.Notas,
		Genero:      product.Genero,
		Marca:       product.Marca,
		Imagen:      product.Imagen,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}
	return &doc, nil
}
