package clients

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ErrProductNotFound indicates that the product does not exist in products-api.
var ErrProductNotFound = errors.New("product not found")

// ProductDTO mirrors the JSON returned by products-api for a product entity.
type ProductDTO struct {
	ID          string    `json:"id"`
	OwnerID     uint      `json:"owner_id"`
	Name        string    `json:"name"`
	Descripcion string    `json:"descripcion"`
	Precio      float64   `json:"precio"`
	Stock       int       `json:"stock"`
	Tipo        string    `json:"tipo"`
	Estacion    string    `json:"estacion"`
	Ocasion     string    `json:"ocasion"`
	Notas       []string  `json:"notas"`
	Genero      string    `json:"genero"`
	Marca       string    `json:"marca"`
	Imagen      string    `json:"imagen"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProductsClient performs HTTP requests against products-api.
type ProductsClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewProductsClient creates a client with sane defaults.
func NewProductsClient(baseURL string) *ProductsClient {
	if baseURL == "" {
		baseURL = "http://localhost:8081"
	}
	return &ProductsClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetProductByID fetches a product using products-api.
func (c *ProductsClient) GetProductByID(ctx context.Context, id string) (*ProductDTO, error) {
	trimmed := strings.TrimSpace(id)
	if trimmed == "" {
		return nil, errors.New("product id is required")
	}

	endpoint := fmt.Sprintf("%s/products/%s", c.baseURL, trimmed)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("products-api request failed: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var product ProductDTO
		if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
			return nil, fmt.Errorf("decode product: %w", err)
		}
		return &product, nil
	case http.StatusNotFound:
		return nil, ErrProductNotFound
	default:
		return nil, fmt.Errorf("products-api unexpected status %d", resp.StatusCode)
	}
}
