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
		return nil, services.BackendError{Message: "build solr request", Err: err}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, services.BackendError{Message: "solr request failed", Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, services.BackendError{
			Message: fmt.Sprintf("solr returned status %d", resp.StatusCode),
		}
	}

	var solrResp solrResponse
	if err := json.NewDecoder(resp.Body).Decode(&solrResp); err != nil {
		return nil, services.BackendError{Message: "decode solr response", Err: err}
	}

	return &services.SearchResult{
		Items: convertDocs(solrResp.Response.Docs),
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
		Docs     []map[string]interface{} `json:"docs"`
	} `json:"response"`
}

func convertDocs(docs []map[string]interface{}) []models.ProductDocument {
	results := make([]models.ProductDocument, 0, len(docs))
	for _, doc := range docs {
		results = append(results, mapSolrDoc(doc))
	}
	return results
}

func mapSolrDoc(doc map[string]interface{}) models.ProductDocument {
	return models.ProductDocument{
		ID:          stringValue(doc["id"]),
		Name:        stringValue(doc["name"]),
		Descripcion: stringValue(doc["descripcion"]),
		Precio:      floatValue(doc["precio"]),
		Stock:       intValue(doc["stock"]),
		Tipo:        stringValue(doc["tipo"]),
		Estacion:    stringValue(doc["estacion"]),
		Ocasion:     stringValue(doc["ocasion"]),
		Notas:       stringSliceValue(doc["notas"]),
		Genero:      stringValue(doc["genero"]),
		Marca:       stringValue(doc["marca"]),
		CreatedAt:   timeValue(doc["created_at"]),
		UpdatedAt:   timeValue(doc["updated_at"]),
	}
}

func firstScalar(value interface{}) interface{} {
	switch v := value.(type) {
	case nil:
		return nil
	case []interface{}:
		for _, item := range v {
			if scalar := firstScalar(item); scalar != nil {
				return scalar
			}
		}
	case []string:
		if len(v) > 0 {
			return v[0]
		}
	default:
		return value
	}
	return nil
}

func stringValue(value interface{}) string {
	scalar := firstScalar(value)
	switch v := scalar.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case bool:
		if v {
			return "true"
		}
		return "false"
	case nil:
		return ""
	default:
		if scalar != nil {
			return fmt.Sprintf("%v", scalar)
		}
		return ""
	}
}

func floatValue(value interface{}) float64 {
	scalar := firstScalar(value)
	switch v := scalar.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case uint64:
		return float64(v)
	case string:
		if parsed, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
			return parsed
		}
	}
	return 0
}

func intValue(value interface{}) int {
	scalar := firstScalar(value)
	switch v := scalar.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case uint64:
		return int(v)
	case float64:
		return int(v)
	case float32:
		return int(v)
	case string:
		if parsed, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			return parsed
		}
	}
	return 0
}

func stringSliceValue(value interface{}) []string {
	switch v := value.(type) {
	case []string:
		out := make([]string, len(v))
		copy(out, v)
		return out
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if str := stringValue(item); str != "" {
				out = append(out, str)
			}
		}
		return out
	case string:
		if strings.TrimSpace(v) == "" {
			return nil
		}
		return []string{v}
	default:
		if str := stringValue(value); str != "" {
			return []string{str}
		}
	}
	return nil
}

func timeValue(value interface{}) time.Time {
	if str := stringValue(value); str != "" {
		if parsed, err := time.Parse(time.RFC3339, str); err == nil {
			return parsed
		}
	}
	return time.Time{}
}
