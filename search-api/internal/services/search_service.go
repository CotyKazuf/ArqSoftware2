package services

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"search-api/internal/cache"
	"search-api/internal/models"
)

// ValidationError indicates an invalid input value.
type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

// BackendError represents failures talking to the search backend (Solr).
type BackendError struct {
	Message string
	Err     error
}

func (e BackendError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap lets errors.Is/As inspect the wrapped error.
func (e BackendError) Unwrap() error {
	return e.Err
}

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
	filters = applySearchDefaults(filters)
	filters = normalizeFilters(filters)
	if err := ValidateSearchFilters(filters); err != nil {
		return nil, err
	}
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
		return ValidationError{Message: "product id is required"}
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
		return ValidationError{Message: "product id is required"}
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
		maxPageSize     = 100
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
	rawKey := fmt.Sprintf("q=%s|tipo=%s|estacion=%s|ocasion=%s|genero=%s|marca=%s|page=%d|size=%d",
		f.Query, f.Tipo, f.Estacion, f.Ocasion, f.Genero, f.Marca, f.Page, f.Size)
	sum := sha256.Sum256([]byte(rawKey))
	return fmt.Sprintf("search:%x", sum[:])
}

// ValidateSearchFilters ensures the incoming filters respect limits.
func ValidateSearchFilters(filters SearchFilters) error {
	if filters.Query != "" {
		if len([]rune(filters.Query)) < 2 {
			return ValidationError{Message: "q must have at least 2 characters"}
		}
		if len([]rune(filters.Query)) > 200 {
			return ValidationError{Message: "q must have 200 characters or fewer"}
		}
	}
	if filters.Page < 1 {
		return ValidationError{Message: "page must be greater or equal to 1"}
	}
	if filters.Size < 1 || filters.Size > 100 {
		return ValidationError{Message: "size must be between 1 and 100"}
	}
	return nil
}

func applySearchDefaults(filters SearchFilters) SearchFilters {
	if filters.Page == 0 {
		filters.Page = 1
	}
	if filters.Size == 0 {
		filters.Size = 10
	}
	return filters
}
