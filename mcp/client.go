// Package mcp implements an MCP (Model Context Protocol) server that exposes
// Cairn's spatial data platform to AI agents.
package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Client is a thin HTTP wrapper for the Cairn REST API.
type Client struct {
	BaseURL    string
	APIKey     string
	SessionCookie string
	HTTPClient *http.Client
}

// NewClient creates a Client with sensible defaults.
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (c *Client) do(req *http.Request) ([]byte, int, error) {
	if c.APIKey != "" {
		req.Header.Set("X-API-Key", c.APIKey)
	}
	if c.SessionCookie != "" {
		req.Header.Set("Cookie", c.SessionCookie)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

func (c *Client) get(path string, query url.Values) ([]byte, int, error) {
	u := c.BaseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, 0, err
	}
	return c.do(req)
}

func (c *Client) postJSON(path string, payload interface{}) ([]byte, int, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}
	req, err := http.NewRequest("POST", c.BaseURL+path, bytes.NewReader(data))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req)
}

// ListDatasets calls GET /datasets.
func (c *Client) ListDatasets() (json.RawMessage, error) {
	body, code, err := c.get("/datasets", nil)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("GET /datasets returned %d: %s", code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// GetCollection calls GET /collections/{id}.
func (c *Client) GetCollection(id string) (json.RawMessage, error) {
	body, code, err := c.get("/collections/"+url.PathEscape(id), nil)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("GET /collections/%s returned %d: %s", id, code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// QueryFeatures calls GET /collections/{id}/items with optional query params.
func (c *Client) QueryFeatures(id string, params map[string]string) (json.RawMessage, error) {
	q := url.Values{}
	for k, v := range params {
		q.Set(k, v)
	}
	body, code, err := c.get("/collections/"+url.PathEscape(id)+"/items", q)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("GET /collections/%s/items returned %d: %s", id, code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// GetFeature calls GET /collections/{id}/items/{fid}.
func (c *Client) GetFeature(collectionID, featureID string) (json.RawMessage, error) {
	path := fmt.Sprintf("/collections/%s/items/%s", url.PathEscape(collectionID), url.PathEscape(featureID))
	body, code, err := c.get(path, nil)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("GET %s returned %d: %s", path, code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// UploadFile calls POST /upload with a multipart file upload.
func (c *Client) UploadFile(filePath string) (json.RawMessage, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, f); err != nil {
		return nil, err
	}
	w.Close()

	req, err := http.NewRequest("POST", c.BaseURL+"/upload", &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	body, code, err := c.do(req)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK && code != http.StatusCreated {
		return nil, fmt.Errorf("POST /upload returned %d: %s", code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// RunProcess calls POST /api/process.
func (c *Client) RunProcess(payload interface{}) (json.RawMessage, error) {
	body, code, err := c.postJSON("/api/process", payload)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("POST /api/process returned %d: %s", code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// RunPipeline calls POST /api/pipeline.
func (c *Client) RunPipeline(payload interface{}) (json.RawMessage, error) {
	body, code, err := c.postJSON("/api/pipeline", payload)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("POST /api/pipeline returned %d: %s", code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// ConvertFormat calls POST /api/convert.
func (c *Client) ConvertFormat(payload interface{}) (json.RawMessage, error) {
	body, code, err := c.postJSON("/api/convert", payload)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("POST /api/convert returned %d: %s", code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// DiffDatasets calls POST /api/diff.
func (c *Client) DiffDatasets(payload interface{}) (json.RawMessage, error) {
	body, code, err := c.postJSON("/api/diff", payload)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("POST /api/diff returned %d: %s", code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// ExecuteSQL calls POST /api/sql/query.
func (c *Client) ExecuteSQL(query string) (json.RawMessage, error) {
	body, code, err := c.postJSON("/api/sql/query", map[string]string{"query": query})
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("POST /api/sql/query returned %d: %s", code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// ListSpatialTables calls GET /api/sql/tables.
func (c *Client) ListSpatialTables() (json.RawMessage, error) {
	body, code, err := c.get("/api/sql/tables", nil)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("GET /api/sql/tables returned %d: %s", code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// Geocode calls GET /api/geocode.
func (c *Client) Geocode(address string) (json.RawMessage, error) {
	body, code, err := c.get("/api/geocode", url.Values{"q": {address}})
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("GET /api/geocode returned %d: %s", code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// ReverseGeocode calls GET /api/geocode/reverse.
func (c *Client) ReverseGeocode(lat, lon string) (json.RawMessage, error) {
	body, code, err := c.get("/api/geocode/reverse", url.Values{"lat": {lat}, "lon": {lon}})
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("GET /api/geocode/reverse returned %d: %s", code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// ComputeRoute calls POST /api/route.
func (c *Client) ComputeRoute(payload interface{}) (json.RawMessage, error) {
	body, code, err := c.postJSON("/api/route", payload)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("POST /api/route returned %d: %s", code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// ListOperations calls GET /api/operations.
func (c *Client) ListOperations() (json.RawMessage, error) {
	body, code, err := c.get("/api/operations", nil)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("GET /api/operations returned %d: %s", code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// GetDatasetSchema calls GET /api/datasets/{name}/schema.
func (c *Client) GetDatasetSchema(name string) (json.RawMessage, error) {
	body, code, err := c.get("/api/datasets/"+url.PathEscape(name)+"/schema", nil)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("GET /api/datasets/%s/schema returned %d: %s", name, code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// GetDatasetProfile calls GET /api/datasets/{name}/profile.
func (c *Client) GetDatasetProfile(name string) (json.RawMessage, error) {
	body, code, err := c.get("/api/datasets/"+url.PathEscape(name)+"/profile", nil)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("GET /api/datasets/%s/profile returned %d: %s", name, code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// BrowseCatalog calls GET /api/catalog.
func (c *Client) BrowseCatalog(params map[string]string) (json.RawMessage, error) {
	q := url.Values{}
	for k, v := range params {
		if v != "" {
			q.Set(k, v)
		}
	}
	body, code, err := c.get("/api/catalog", q)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("GET /api/catalog returned %d: %s", code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// ImportFromCatalog calls POST /api/catalog/import.
func (c *Client) ImportFromCatalog(catalogID string) (json.RawMessage, error) {
	body, code, err := c.postJSON("/api/catalog/import", map[string]string{"catalog_id": catalogID})
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK && code != http.StatusCreated {
		return nil, fmt.Errorf("POST /api/catalog/import returned %d: %s", code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// BrowseSTACCatalog calls GET /api/stac/remote.
func (c *Client) BrowseSTACCatalog(catalogURL string) (json.RawMessage, error) {
	body, code, err := c.get("/api/stac/remote", url.Values{"url": {catalogURL}})
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("GET /api/stac/remote returned %d: %s", code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// BrowseSTACCollections calls GET /api/stac/remote/collections.
func (c *Client) BrowseSTACCollections(catalogURL string) (json.RawMessage, error) {
	body, code, err := c.get("/api/stac/remote/collections", url.Values{"url": {catalogURL}})
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("GET /api/stac/remote/collections returned %d: %s", code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// BrowseSTACItems calls GET /api/stac/remote/items.
func (c *Client) BrowseSTACItems(collectionURL string, params map[string]string) (json.RawMessage, error) {
	q := url.Values{"url": {collectionURL}}
	for k, v := range params {
		if v != "" {
			q.Set(k, v)
		}
	}
	body, code, err := c.get("/api/stac/remote/items", q)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("GET /api/stac/remote/items returned %d: %s", code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// ImportSTACAsset calls POST /api/stac/import.
func (c *Client) ImportSTACAsset(assetURL, name, format string) (json.RawMessage, error) {
	payload := map[string]string{
		"asset_url": assetURL,
		"name":      name,
	}
	if format != "" {
		payload["format"] = format
	}
	body, code, err := c.postJSON("/api/stac/import", payload)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK && code != http.StatusCreated {
		return nil, fmt.Errorf("POST /api/stac/import returned %d: %s", code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// SearchSTAC calls GET /stac/search.
func (c *Client) SearchSTAC(params map[string]string) (json.RawMessage, error) {
	q := url.Values{}
	for k, v := range params {
		if v != "" {
			q.Set(k, v)
		}
	}
	body, code, err := c.get("/stac/search", q)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("GET /stac/search returned %d: %s", code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

func truncate(b []byte, n int) string {
	s := string(b)
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}
