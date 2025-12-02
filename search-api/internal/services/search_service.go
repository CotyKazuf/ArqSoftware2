package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"search-api/internal/cache"
	"search-api/internal/models"
)

// ErrValidation indicates an invalid input value.
var ErrValidation = errors.New("validation error")

// SearchFilters encapsulates query parameters for product search.
type SearchFilters struct {
	Query    string
	Tipo     string
	Estacion string
	Ocasion  string
	Genero   string
	Marca    string
	Page     int
	Size     int
}

// SearchResult represents a paginated Solr response.
type SearchResult struct {
	Items []models.ProductDocument `json:"items"`
	Page  int                      `json:"page"`
	Size  int                      `json:"size"`
	Total int64                    `json:"total"`
}

// IndexRepository abstracts the search/index backend (Solr).
type IndexRepository interface {
	Search(ctx context.Context, filters SearchFilters) (*SearchResult, error)
	IndexProduct(ctx context.Context, product models.ProductDocument) error
	DeleteProduct(ctx context.Context, id string) error
}

// SearchService coordinates search, cache and index updates.
type SearchService struct {
	indexRepo IndexRepository
	cache     cache.Cache
	cacheTTL  time.Duration
}

// NewSearchService wires the service with its dependencies.
func NewSearchService(indexRepo IndexRepository, cache cache.Cache, cacheTTL time.Duration) *SearchService {
	return &SearchService{
		indexRepo: indexRepo,
		cache:     cache,
		cacheTTL:  cacheTTL,
	}
}

// SearchProducts runs a search over Solr with layered caching.
func (s *SearchService) SearchProducts(ctx context.Context, filters SearchFilters) (*SearchResult, error) {
	filters = normalizeFilters(filters)
	filters.Page, filters.Size = sanitizePagination(filters.Page, filters.Size)

	key := buildCacheKey(filters)
	if s.cache != nil {
		if data, ok := s.cache.Get(key); ok {
			var cached SearchResult
			if err := json.Unmarshal(data, &cached); err == nil {
				return &cached, nil
			}
		}
	}

	result, err := s.indexRepo.Search(ctx, filters)
	if err != nil {
		return nil, err
	}

	if s.cache != nil && result != nil {
		if encoded, err := json.Marshal(result); err == nil {
			s.cache.Set(key, encoded, s.cacheTTL)
		}
	}
	return result, nil
}

// IndexProduct indexes or updates a product document.
func (s *SearchService) IndexProduct(ctx context.Context, product models.ProductDocument) error {
	if strings.TrimSpace(product.ID) == "" {
		return ErrValidation
	}
	if err := s.indexRepo.IndexProduct(ctx, product); err != nil {
		return err
	}
	s.invalidateCaches()
	return nil
}

// DeleteProduct removes a product from the index.
func (s *SearchService) DeleteProduct(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrValidation
	}
	if err := s.indexRepo.DeleteProduct(ctx, id); err != nil {
		return err
	}
	s.invalidateCaches()
	return nil
}

// FlushCaches clears all caches explicitly (used by admin endpoint).
func (s *SearchService) FlushCaches(ctx context.Context) error {
	if s.cache == nil {
		return nil
	}
	return s.cache.Flush()
}

func (s *SearchService) invalidateCaches() {
	if s.cache == nil {
		return
	}
	_ = s.cache.Flush()
}

func normalizeFilters(f SearchFilters) SearchFilters {
	f.Query = strings.TrimSpace(f.Query)
	f.Tipo = strings.ToLower(strings.TrimSpace(f.Tipo))
	f.Estacion = strings.ToLower(strings.TrimSpace(f.Estacion))
	f.Ocasion = strings.ToLower(strings.TrimSpace(f.Ocasion))
	f.Genero = strings.ToLower(strings.TrimSpace(f.Genero))
	f.Marca = strings.ToLower(strings.TrimSpace(f.Marca))
	return f
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

func buildCacheKey(f SearchFilters) string {
	return fmt.Sprintf("q=%s|tipo=%s|estacion=%s|ocasion=%s|genero=%s|marca=%s|page=%d|size=%d",
		f.Query, f.Tipo, f.Estacion, f.Ocasion, f.Genero, f.Marca, f.Page, f.Size)
}
