package services

import (
	"context"
	"testing"
	"time"

	"search-api/internal/models"
	"search-api/internal/rabbitmq"
)

func TestSearchServiceCachesResults(t *testing.T) {
	repo := &mockIndexRepo{
		result: &SearchResult{
			Items: []models.ProductDocument{{ID: "1", Name: "Floral"}},
			Page:  1,
			Size:  10,
			Total: 1,
		},
	}
	cache := newMapCache()
	service := NewSearchService(repo, cache, time.Minute)

	_, err := service.SearchProducts(context.Background(), SearchFilters{Query: "floral"})
	if err != nil {
		t.Fatalf("search returned error: %v", err)
	}
	if repo.searchCount != 1 {
		t.Fatalf("expected search to hit backend once, got %d", repo.searchCount)
	}

	_, err = service.SearchProducts(context.Background(), SearchFilters{Query: "floral"})
	if err != nil {
		t.Fatalf("cached search returned error: %v", err)
	}
	if repo.searchCount != 1 {
		t.Fatalf("expected cached search to bypass backend, got %d", repo.searchCount)
	}
}

func TestEventProcessorRoutesEvents(t *testing.T) {
	repo := &mockIndexRepo{}
	cache := newMapCache()
	service := NewSearchService(repo, cache, time.Minute)
	processor := NewEventProcessor(service, "")

	createEvent := rabbitmq.ProductEvent{
		Type: rabbitmq.EventProductCreated,
		Product: models.ProductDocument{
			ID:   "abc",
			Name: "Test",
		},
	}

	if err := processor.HandleProductEvent(context.Background(), createEvent); err != nil {
		t.Fatalf("handle create event: %v", err)
	}
	if repo.indexCount != 1 {
		t.Fatalf("expected index to be called on create event")
	}
	if cache.flushes != 1 {
		t.Fatalf("expected caches to be flushed after create")
	}

	deleteEvent := rabbitmq.ProductEvent{
		Type: rabbitmq.EventProductDeleted,
		Product: models.ProductDocument{
			ID: "abc",
		},
	}
	if err := processor.HandleProductEvent(context.Background(), deleteEvent); err != nil {
		t.Fatalf("handle delete event: %v", err)
	}
	if repo.deleteCount != 1 {
		t.Fatalf("expected delete to be called on delete event")
	}
	if cache.flushes != 2 {
		t.Fatalf("expected caches flushed twice")
	}
}

// --- test doubles ---

type mockIndexRepo struct {
	searchCount int
	indexCount  int
	deleteCount int
	result      *SearchResult
	searchErr   error
}

func (m *mockIndexRepo) Search(ctx context.Context, filters SearchFilters) (*SearchResult, error) {
	m.searchCount++
	if m.searchErr != nil {
		return nil, m.searchErr
	}
	if m.result != nil {
		return m.result, nil
	}
	return &SearchResult{Items: []models.ProductDocument{}, Page: filters.Page, Size: filters.Size, Total: 0}, nil
}

func (m *mockIndexRepo) IndexProduct(ctx context.Context, product models.ProductDocument) error {
	m.indexCount++
	return nil
}

func (m *mockIndexRepo) DeleteProduct(ctx context.Context, id string) error {
	m.deleteCount++
	return nil
}

type mapCache struct {
	store   map[string][]byte
	flushes int
}

func newMapCache() *mapCache {
	return &mapCache{store: map[string][]byte{}}
}

func (m *mapCache) Get(key string) ([]byte, bool) {
	val, ok := m.store[key]
	return val, ok
}

func (m *mapCache) Set(key string, value []byte, ttl time.Duration) {
	m.store[key] = value
}

func (m *mapCache) Delete(key string) {
	delete(m.store, key)
}

func (m *mapCache) Flush() error {
	m.store = map[string][]byte{}
	m.flushes++
	return nil
}
