package services

import (
	"context"
	"fmt"

	"search-api/internal/rabbitmq"
)

// EventProcessor implements rabbitmq.EventHandler and delegates to SearchService.
type EventProcessor struct {
	service *SearchService
}

// NewEventProcessor builds an EventProcessor.
func NewEventProcessor(service *SearchService) *EventProcessor {
	return &EventProcessor{service: service}
}

// HandleProductEvent processes product events emitted by products-api.
func (p *EventProcessor) HandleProductEvent(ctx context.Context, event rabbitmq.ProductEvent) error {
	switch event.Type {
	case rabbitmq.EventProductCreated, rabbitmq.EventProductUpdated:
		if err := p.service.IndexProduct(ctx, event.Product); err != nil {
			return fmt.Errorf("index product: %w", err)
		}
	case rabbitmq.EventProductDeleted:
		if err := p.service.DeleteProduct(ctx, event.Product.ID); err != nil {
			return fmt.Errorf("delete product: %w", err)
		}
	default:
		return fmt.Errorf("unsupported event type %s", event.Type)
	}
	return nil
}
