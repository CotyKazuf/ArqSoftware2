package solr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"search-api/internal/models"
	"search-api/internal/services"
)

// Client is a lightweight Solr HTTP client that satisfies services.IndexRepository.
type Client struct {
	baseURL    string
	core       string
	httpClient *http.Client
}

// NewClient builds a Solr client with sane defaults.
func NewClient(baseURL, core string) *Client {
	if baseURL == "" {
		baseURL = "http://localhost:8983/solr"
	}
	if core == "" {
		core = "products-core"
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		core:    core,
		httpClient: &http.Client{
			Timeout: 8 * time.Second,
		},
	}
}

// Search runs a query against Solr and returns a paginated result.
func (c *Client) Search(ctx context.Context, filters services.SearchFilters) (*services.SearchResult, error) {
	params := url.Values{}
	params.Set("wt", "json")
	params.Set("q", "*:*")

	if filters.Query != "" {
		term := escapeTerm(filters.Query)
		params.Add("fq", fmt.Sprintf("(name:*%s* OR descripcion:*%s*)", term, term))
	}
	if filters.Tipo != "" {
		params.Add("fq", fmt.Sprintf("tipo:%s", escapeTerm(filters.Tipo)))
	}
	if filters.Estacion != "" {
		params.Add("fq", fmt.Sprintf("estacion:%s", escapeTerm(filters.Estacion)))
	}
	if filters.Ocasion != "" {
		params.Add("fq", fmt.Sprintf("ocasion:%s", escapeTerm(filters.Ocasion)))
	}
	if filters.Genero != "" {
		params.Add("fq", fmt.Sprintf("genero:%s", escapeTerm(filters.Genero)))
	}
	if filters.Marca != "" {
		params.Add("fq", fmt.Sprintf("marca:%s", escapeTerm(filters.Marca)))
	}

	start := (filters.Page - 1) * filters.Size
	if start < 0 {
		start = 0
	}
	params.Set("start", strconv.Itoa(start))
	params.Set("rows", strconv.Itoa(filters.Size))
	params.Set("sort", "updated_at desc")

	endpoint := fmt.Sprintf("%s/%s/select?%s", c.baseURL, c.core, params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build solr request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("solr request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("solr returned status %d", resp.StatusCode)
	}

	var solrResp solrResponse
	if err := json.NewDecoder(resp.Body).Decode(&solrResp); err != nil {
		return nil, fmt.Errorf("decode solr response: %w", err)
	}

	return &services.SearchResult{
		Items: solrResp.Response.Docs,
		Page:  filters.Page,
		Size:  filters.Size,
		Total: solrResp.Response.NumFound,
	}, nil
}

// IndexProduct adds or updates a product document.
func (c *Client) IndexProduct(ctx context.Context, product models.ProductDocument) error {
	payload := map[string]interface{}{
		"add": map[string]interface{}{
			"doc":       product,
			"overwrite": true,
		},
	}
	return c.postUpdate(ctx, payload)
}

// DeleteProduct removes a product document by id.
func (c *Client) DeleteProduct(ctx context.Context, id string) error {
	payload := map[string]interface{}{
		"delete": map[string]string{"id": id},
	}
	return c.postUpdate(ctx, payload)
}

func (c *Client) postUpdate(ctx context.Context, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	endpoint := fmt.Sprintf("%s/%s/update?commit=true", c.baseURL, c.core)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build solr update request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("solr update request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("solr update returned status %d", resp.StatusCode)
	}
	return nil
}

func escapeTerm(term string) string {
	term = strings.ReplaceAll(term, ":", "\\:")
	term = strings.ReplaceAll(term, " ", "\\ ")
	return term
}

type solrResponse struct {
	Response struct {
		NumFound int64                    `json:"numFound"`
		Docs     []models.ProductDocument `json:"docs"`
	} `json:"response"`
}
