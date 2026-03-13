// Package mcp implements an MCP (Model Context Protocol) server that exposes
// Roteiro's spatial data platform to AI agents.
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

// Client is a thin HTTP wrapper for the Roteiro REST API.
type Client struct {
	BaseURL       string
	APIKey        string
	SessionCookie string
	HTTPClient    *http.Client
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

func (c *Client) callJSON(method, path string, payload interface{}, query map[string]string) ([]byte, int, error) {
	if !strings.HasPrefix(path, "/") {
		return nil, 0, fmt.Errorf("path must start with '/': %s", path)
	}
	u := c.BaseURL + path
	if len(query) > 0 {
		q := url.Values{}
		for k, v := range query {
			if k != "" && v != "" {
				q.Set(k, v)
			}
		}
		if enc := q.Encode(); enc != "" {
			u += "?" + enc
		}
	}

	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, 0, err
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, u, body)
	if err != nil {
		return nil, 0, err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.do(req)
}

// APIRequest sends a constrained API request used by allowlisted MCP tools.
func (c *Client) APIRequest(method, path string, payload interface{}, query map[string]string) (json.RawMessage, error) {
	body, code, err := c.callJSON(method, path, payload, query)
	if err != nil {
		return nil, err
	}
	if code < 200 || code >= 300 {
		return nil, fmt.Errorf("%s %s returned %d: %s", method, path, code, truncate(body, 500))
	}
	if len(body) == 0 {
		return json.RawMessage(`{"status":"ok"}`), nil
	}
	return json.RawMessage(body), nil
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

// RunRasterProcess calls POST /api/raster/process.
func (c *Client) RunRasterProcess(payload interface{}) (json.RawMessage, error) {
	body, code, err := c.postJSON("/api/raster/process", payload)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("POST /api/raster/process returned %d: %s", code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// PreflightProcess calls POST /api/process/preflight.
func (c *Client) PreflightProcess(payload interface{}) (json.RawMessage, error) {
	body, code, err := c.postJSON("/api/process/preflight", payload)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("POST /api/process/preflight returned %d: %s", code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// SubmitProcessJob calls POST /api/process/jobs.
func (c *Client) SubmitProcessJob(payload interface{}) (json.RawMessage, error) {
	body, code, err := c.postJSON("/api/process/jobs", payload)
	if err != nil {
		return nil, err
	}
	if code != http.StatusAccepted {
		return nil, fmt.Errorf("POST /api/process/jobs returned %d: %s", code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// SubmitProcessBatch calls POST /api/process/jobs/batch.
func (c *Client) SubmitProcessBatch(payload interface{}) (json.RawMessage, error) {
	body, code, err := c.postJSON("/api/process/jobs/batch", payload)
	if err != nil {
		return nil, err
	}
	if code != http.StatusAccepted {
		return nil, fmt.Errorf("POST /api/process/jobs/batch returned %d: %s", code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// ListProcessJobs calls GET /api/process/jobs.
func (c *Client) ListProcessJobs(params map[string]string) (json.RawMessage, error) {
	q := url.Values{}
	for k, v := range params {
		if v != "" {
			q.Set(k, v)
		}
	}
	body, code, err := c.get("/api/process/jobs", q)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("GET /api/process/jobs returned %d: %s", code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// GetProcessJob calls GET /api/process/jobs/{id}.
func (c *Client) GetProcessJob(id string) (json.RawMessage, error) {
	body, code, err := c.get("/api/process/jobs/"+url.PathEscape(id), nil)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("GET /api/process/jobs/%s returned %d: %s", id, code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// CancelProcessJob calls DELETE /api/process/jobs/{id}.
func (c *Client) CancelProcessJob(id string) (json.RawMessage, error) {
	body, code, err := c.callJSON("DELETE", "/api/process/jobs/"+url.PathEscape(id), nil, nil)
	if err != nil {
		return nil, err
	}
	if code != http.StatusNoContent {
		return nil, fmt.Errorf("DELETE /api/process/jobs/%s returned %d: %s", id, code, truncate(body, 500))
	}
	return json.RawMessage(`{"status":"cancelled"}`), nil
}

// RerunProcessJob calls POST /api/process/jobs/{id}/rerun.
func (c *Client) RerunProcessJob(id string) (json.RawMessage, error) {
	body, code, err := c.postJSON("/api/process/jobs/"+url.PathEscape(id)+"/rerun", map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	if code != http.StatusAccepted {
		return nil, fmt.Errorf("POST /api/process/jobs/%s/rerun returned %d: %s", id, code, truncate(body, 500))
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
	if code != http.StatusOK && code != http.StatusCreated {
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

// ExecuteSQL calls POST /api/query/sql.
func (c *Client) ExecuteSQL(query string) (json.RawMessage, error) {
	body, code, err := c.postJSON("/api/query/sql", map[string]string{"sql": query})
	if err != nil {
		return nil, err
	}
	if code == http.StatusNotFound || code == http.StatusMethodNotAllowed {
		// Backward compatibility for older Roteiro deployments.
		body, code, err = c.postJSON("/api/sql/query", map[string]string{"sql": query})
		if err != nil {
			return nil, err
		}
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("POST SQL query endpoint returned %d: %s", code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// ListSpatialTables calls GET /api/sql/tables (or fallback /api/query/sql/datasets).
func (c *Client) ListSpatialTables() (json.RawMessage, error) {
	body, code, err := c.get("/api/sql/tables", nil)
	if err != nil {
		return nil, err
	}
	if code == http.StatusNotFound || code == http.StatusMethodNotAllowed {
		body, code, err = c.get("/api/query/sql/datasets", nil)
		if err != nil {
			return nil, err
		}
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("GET SQL tables endpoint returned %d: %s", code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// GetDuckDBInfo calls GET /api/query/sql/info.
func (c *Client) GetDuckDBInfo() (json.RawMessage, error) {
	body, code, err := c.get("/api/query/sql/info", nil)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("GET /api/query/sql/info returned %d: %s", code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// ListDuckDBDatasets calls GET /api/query/sql/datasets.
func (c *Client) ListDuckDBDatasets() (json.RawMessage, error) {
	body, code, err := c.get("/api/query/sql/datasets", nil)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("GET /api/query/sql/datasets returned %d: %s", code, truncate(body, 500))
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

// ComputeIsochrone calls POST /api/route/isochrone.
func (c *Client) ComputeIsochrone(payload interface{}) (json.RawMessage, error) {
	body, code, err := c.postJSON("/api/route/isochrone", payload)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("POST /api/route/isochrone returned %d: %s", code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// ComputeRouteMatrix calls POST /api/route/matrix.
func (c *Client) ComputeRouteMatrix(payload interface{}) (json.RawMessage, error) {
	body, code, err := c.postJSON("/api/route/matrix", payload)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("POST /api/route/matrix returned %d: %s", code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// ComputeServiceArea calls POST /api/route/service-area.
func (c *Client) ComputeServiceArea(payload interface{}) (json.RawMessage, error) {
	body, code, err := c.postJSON("/api/route/service-area", payload)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("POST /api/route/service-area returned %d: %s", code, truncate(body, 500))
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

// ListAnalysisOperations calls GET /api/analysis/operations.
func (c *Client) ListAnalysisOperations() (json.RawMessage, error) {
	body, code, err := c.get("/api/analysis/operations", nil)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("GET /api/analysis/operations returned %d: %s", code, truncate(body, 500))
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

// BrowseEnhancedCatalog calls GET /api/catalog/enhanced.
func (c *Client) BrowseEnhancedCatalog(params map[string]string) (json.RawMessage, error) {
	q := url.Values{}
	for k, v := range params {
		if v != "" {
			q.Set(k, v)
		}
	}
	body, code, err := c.get("/api/catalog/enhanced", q)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("GET /api/catalog/enhanced returned %d: %s", code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// GetCatalogEntry calls GET /api/catalog/enhanced/{id}.
func (c *Client) GetCatalogEntry(id string) (json.RawMessage, error) {
	body, code, err := c.get("/api/catalog/enhanced/"+url.PathEscape(id), nil)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("GET /api/catalog/enhanced/%s returned %d: %s", id, code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// ListCatalogCategories calls GET /api/catalog/categories.
func (c *Client) ListCatalogCategories() (json.RawMessage, error) {
	body, code, err := c.get("/api/catalog/categories", nil)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("GET /api/catalog/categories returned %d: %s", code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// ListCatalogTags calls GET /api/catalog/tags.
func (c *Client) ListCatalogTags(limit string) (json.RawMessage, error) {
	q := url.Values{}
	if limit != "" {
		q.Set("limit", limit)
	}
	body, code, err := c.get("/api/catalog/tags", q)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("GET /api/catalog/tags returned %d: %s", code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

// ImportFromCatalog calls POST /api/catalog/import.
func (c *Client) ImportFromCatalog(catalogID string) (json.RawMessage, error) {
	body, code, err := c.postJSON("/api/catalog/import", map[string]string{"catalog_id": catalogID})
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK && code != http.StatusCreated && code != http.StatusAccepted {
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
	if code != http.StatusOK && code != http.StatusCreated && code != http.StatusAccepted {
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
