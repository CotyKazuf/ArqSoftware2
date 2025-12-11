package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"search-api/internal/models"
	"search-api/internal/rabbitmq"
)

// EventProcessor implements rabbitmq.EventHandler and delegates to SearchService.
type EventProcessor struct {
	service        *SearchService
	productsAPIURL string
	httpClient     *http.Client
}

// NewEventProcessor builds an EventProcessor.
func NewEventProcessor(service *SearchService, productsAPIURL string) *EventProcessor {
	return &EventProcessor{
		service:        service,
		productsAPIURL: strings.TrimRight(productsAPIURL, "/"),
		httpClient:     &http.Client{},
	}
}

// HandleProductEvent processes product events emitted by products-api.
func (p *EventProcessor) HandleProductEvent(ctx context.Context, event rabbitmq.ProductEvent) error {
	switch event.Type {
	case rabbitmq.EventProductCreated, rabbitmq.EventProductUpdated:
		doc := event.Product
		if fetched, err := p.fetchProduct(ctx, event.Product.ID); err == nil && fetched.ID != "" {
			doc = fetched
		}
		if err := p.service.IndexProduct(ctx, doc); err != nil {
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

func (p *EventProcessor) fetchProduct(ctx context.Context, id string) (models.ProductDocument, error) {
	if p.productsAPIURL == "" || strings.TrimSpace(id) == "" {
		return models.ProductDocument{}, fmt.Errorf("missing products api url or id")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/products/%s", p.productsAPIURL, id), nil)
	if err != nil {
		return models.ProductDocument{}, err
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return models.ProductDocument{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return models.ProductDocument{}, fmt.Errorf("products-api returned %d", resp.StatusCode)
	}

	var envelope struct {
		Data  models.ProductDocument `json:"data"`
		Error interface{}            `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return models.ProductDocument{}, err
	}

	return envelope.Data, nil
}
