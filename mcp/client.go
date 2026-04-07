// Package mcp implements an MCP (Model Context Protocol) server that exposes
// Cairn's stable public API to AI agents.
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
	"strconv"
	"strings"
	"time"
)

// Client is a thin HTTP wrapper for the Cairn REST API.
type Client struct {
	BaseURL       string
	APIKey        string
	SessionCookie string
	ProjectID     string
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
	if c.ProjectID != "" {
		req.Header.Set("X-Project-ID", c.ProjectID)
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

// WithProjectID returns a shallow clone of the client with an overridden
// project scope. An empty value preserves the existing scope.
func (c *Client) WithProjectID(projectID string) *Client {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" || projectID == c.ProjectID {
		return c
	}
	clone := *c
	clone.ProjectID = projectID
	return &clone
}

func (c *Client) request(method, path string, payload interface{}, query map[string]string, headers map[string]string, expected ...int) (json.RawMessage, error) {
	if !strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("path must start with '/': %s", path)
	}

	u := c.BaseURL + path
	if len(query) > 0 {
		q := url.Values{}
		for key, value := range query {
			if key != "" && value != "" {
				q.Set(key, value)
			}
		}
		if encoded := q.Encode(); encoded != "" {
			u += "?" + encoded
		}
	}

	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, u, body)
	if err != nil {
		return nil, err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		if strings.TrimSpace(key) != "" && strings.TrimSpace(value) != "" {
			req.Header.Set(key, value)
		}
	}

	respBody, code, err := c.do(req)
	if err != nil {
		return nil, err
	}
	if len(expected) == 0 {
		if code < 200 || code >= 300 {
			return nil, fmt.Errorf("%s %s returned %d: %s", method, path, code, truncate(respBody, 500))
		}
	} else {
		matched := false
		for _, want := range expected {
			if code == want {
				matched = true
				break
			}
		}
		if !matched {
			return nil, fmt.Errorf("%s %s returned %d: %s", method, path, code, truncate(respBody, 500))
		}
	}
	if len(respBody) == 0 {
		return json.RawMessage(`{"status":"ok"}`), nil
	}
	return json.RawMessage(respBody), nil
}

func (c *Client) UploadFile(filePath, name, projectID, bodyID string) (json.RawMessage, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, err
	}
	if strings.TrimSpace(name) != "" {
		if err := writer.WriteField("name", name); err != nil {
			return nil, err
		}
	}
	if strings.TrimSpace(projectID) != "" {
		if err := writer.WriteField("project_id", projectID); err != nil {
			return nil, err
		}
	}
	if strings.TrimSpace(bodyID) != "" {
		if err := writer.WriteField("body_id", bodyID); err != nil {
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/upload", &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	body, code, err := c.do(req)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK && code != http.StatusCreated && code != http.StatusAccepted {
		return nil, fmt.Errorf("POST /upload returned %d: %s", code, truncate(body, 500))
	}
	return json.RawMessage(body), nil
}

func (c *Client) ListDatasets() (json.RawMessage, error) {
	return c.request("GET", "/datasets", nil, nil, nil, http.StatusOK)
}

func (c *Client) GetCollection(id string) (json.RawMessage, error) {
	return c.request("GET", "/collections/"+url.PathEscape(id), nil, nil, nil, http.StatusOK)
}

func (c *Client) QueryFeatures(id string, params map[string]string) (json.RawMessage, error) {
	return c.request("GET", "/collections/"+url.PathEscape(id)+"/items", nil, params, nil, http.StatusOK)
}

func (c *Client) GetFeature(collectionID, featureID string) (json.RawMessage, error) {
	path := fmt.Sprintf("/collections/%s/items/%s", url.PathEscape(collectionID), url.PathEscape(featureID))
	return c.request("GET", path, nil, nil, nil, http.StatusOK)
}

func (c *Client) CreateFeature(collectionID string, feature interface{}) (json.RawMessage, error) {
	path := fmt.Sprintf("/collections/%s/items", url.PathEscape(collectionID))
	return c.request("POST", path, feature, nil, map[string]string{"Content-Type": "application/geo+json"}, http.StatusOK, http.StatusCreated)
}

func (c *Client) UpdateFeature(collectionID, featureID string, feature interface{}) (json.RawMessage, error) {
	path := fmt.Sprintf("/collections/%s/items/%s", url.PathEscape(collectionID), url.PathEscape(featureID))
	return c.request("PUT", path, feature, nil, map[string]string{"Content-Type": "application/geo+json"}, http.StatusOK)
}

func (c *Client) DeleteFeature(collectionID, featureID string) (json.RawMessage, error) {
	path := fmt.Sprintf("/collections/%s/items/%s", url.PathEscape(collectionID), url.PathEscape(featureID))
	return c.request("DELETE", path, nil, nil, nil, http.StatusNoContent)
}

func (c *Client) GetDatasetSchema(name string) (json.RawMessage, error) {
	return c.request("GET", "/api/v1/datasets/"+url.PathEscape(name)+"/schema", nil, nil, nil, http.StatusOK)
}

func (c *Client) GetDatasetProfile(name string) (json.RawMessage, error) {
	return c.request("GET", "/api/v1/datasets/"+url.PathEscape(name)+"/profile", nil, nil, nil, http.StatusOK)
}

func (c *Client) ImportSource(payload interface{}) (json.RawMessage, error) {
	return c.request("POST", "/api/v1/datasets/import-source", payload, nil, nil, http.StatusOK, http.StatusCreated, http.StatusAccepted)
}

func (c *Client) GetSceneManifest() (json.RawMessage, error) {
	return c.request("GET", "/api/v1/scene-manifest", nil, nil, nil, http.StatusOK)
}

func (c *Client) ListBodies() (json.RawMessage, error) {
	return c.request("GET", "/api/v1/bodies", nil, nil, nil, http.StatusOK)
}

func (c *Client) GetBody(slug string) (json.RawMessage, error) {
	return c.request("GET", "/api/v1/bodies/"+url.PathEscape(slug), nil, nil, nil, http.StatusOK)
}

func (c *Client) GetBodyRecipes(slug string) (json.RawMessage, error) {
	return c.request("GET", "/api/v1/bodies/"+url.PathEscape(slug)+"/recipes", nil, nil, nil, http.StatusOK)
}

func (c *Client) ExecuteBodyRecipe(slug, sourceID string) (json.RawMessage, error) {
	path := fmt.Sprintf("/api/v1/bodies/%s/recipes/%s/execute", url.PathEscape(slug), url.PathEscape(sourceID))
	return c.request("POST", path, map[string]interface{}{}, nil, nil, http.StatusOK)
}

func (c *Client) ListOperations(domain string) (json.RawMessage, error) {
	query := map[string]string{}
	if strings.TrimSpace(domain) != "" {
		query["domain"] = domain
	}
	return c.request("GET", "/api/v1/ops", nil, query, nil, http.StatusOK)
}

func (c *Client) PreflightOperation(payload interface{}) (json.RawMessage, error) {
	return c.request("POST", "/api/v1/ops/preflight", payload, nil, nil, http.StatusOK)
}

func (c *Client) RunOperation(operation string, payload interface{}) (json.RawMessage, error) {
	return c.request("POST", "/api/v1/ops/"+url.PathEscape(operation), payload, nil, nil, http.StatusOK)
}

func (c *Client) SubmitOperationJob(operation string, payload interface{}) (json.RawMessage, error) {
	path := "/api/v1/ops/by-operation/" + url.PathEscape(operation) + "/jobs"
	return c.request("POST", path, payload, nil, nil, http.StatusOK, http.StatusAccepted)
}

func (c *Client) SubmitOperationBatch(payload interface{}) (json.RawMessage, error) {
	return c.request("POST", "/api/v1/ops/jobs/batch", payload, nil, nil, http.StatusOK, http.StatusAccepted)
}

func (c *Client) ListOperationJobs(params map[string]string) (json.RawMessage, error) {
	return c.request("GET", "/api/v1/ops/jobs", nil, params, nil, http.StatusOK)
}

func (c *Client) GetOperationJob(id string) (json.RawMessage, error) {
	return c.request("GET", "/api/v1/ops/jobs/"+url.PathEscape(id), nil, nil, nil, http.StatusOK)
}

func (c *Client) CancelOperationJob(id string) (json.RawMessage, error) {
	return c.request("DELETE", "/api/v1/ops/jobs/"+url.PathEscape(id), nil, nil, nil, http.StatusNoContent)
}

func (c *Client) RerunOperationJob(id string) (json.RawMessage, error) {
	return c.request("POST", "/api/v1/ops/jobs/"+url.PathEscape(id)+"/rerun", map[string]interface{}{}, nil, nil, http.StatusOK, http.StatusAccepted)
}

func (c *Client) ListPipelineOperations() (json.RawMessage, error) {
	return c.request("GET", "/api/v1/pipeline/operations", nil, nil, nil, http.StatusOK)
}

func (c *Client) RunPipeline(payload interface{}) (json.RawMessage, error) {
	return c.request("POST", "/api/v1/pipeline", payload, nil, nil, http.StatusOK)
}

func (c *Client) ListPipelines() (json.RawMessage, error) {
	return c.request("GET", "/api/v1/pipelines", nil, nil, nil, http.StatusOK)
}

func (c *Client) GetPipeline(id string) (json.RawMessage, error) {
	return c.request("GET", "/api/v1/pipelines/"+url.PathEscape(id), nil, nil, nil, http.StatusOK)
}

func (c *Client) CreatePipeline(payload interface{}) (json.RawMessage, error) {
	return c.request("POST", "/api/v1/pipelines", payload, nil, nil, http.StatusOK, http.StatusCreated)
}

func (c *Client) UpdatePipeline(id string, payload interface{}) (json.RawMessage, error) {
	return c.request("PUT", "/api/v1/pipelines/"+url.PathEscape(id), payload, nil, nil, http.StatusOK)
}

func (c *Client) DeletePipeline(id string) (json.RawMessage, error) {
	return c.request("DELETE", "/api/v1/pipelines/"+url.PathEscape(id), nil, nil, nil, http.StatusNoContent)
}

func (c *Client) DuplicatePipeline(id string) (json.RawMessage, error) {
	return c.request("POST", "/api/v1/pipelines/"+url.PathEscape(id)+"/duplicate", map[string]interface{}{}, nil, nil, http.StatusOK, http.StatusCreated)
}

func (c *Client) ExecutePipeline(id string) (json.RawMessage, error) {
	return c.request("POST", "/api/v1/pipelines/"+url.PathEscape(id)+"/execute", map[string]interface{}{}, nil, nil, http.StatusOK)
}

func (c *Client) ListPipelineRuns(id string) (json.RawMessage, error) {
	return c.request("GET", "/api/v1/pipelines/"+url.PathEscape(id)+"/runs", nil, nil, nil, http.StatusOK)
}

func (c *Client) GetPipelineRun(id string) (json.RawMessage, error) {
	return c.request("GET", "/api/v1/pipeline-runs/"+url.PathEscape(id), nil, nil, nil, http.StatusOK)
}

func (c *Client) ListQueryEngines() (json.RawMessage, error) {
	return c.request("GET", "/api/v1/query/engines", nil, nil, nil, http.StatusOK)
}

func (c *Client) GetQueryEngineInfo(engine string) (json.RawMessage, error) {
	return c.request("GET", "/api/v1/query/sql/info", nil, map[string]string{"engine": engine}, nil, http.StatusOK)
}

func (c *Client) ListQueryDatasets(engine string) (json.RawMessage, error) {
	return c.request("GET", "/api/v1/query/sql/datasets", nil, map[string]string{"engine": engine}, nil, http.StatusOK)
}

func (c *Client) ExecuteSQL(engine string, payload interface{}) (json.RawMessage, error) {
	return c.request("POST", "/api/v1/query/sql", payload, map[string]string{"engine": engine}, nil, http.StatusOK)
}

func (c *Client) SaveSQLResult(engine string, payload interface{}) (json.RawMessage, error) {
	return c.request("POST", "/api/v1/query/sql/save", payload, map[string]string{"engine": engine}, nil, http.StatusCreated)
}

func (c *Client) ListProjects() (json.RawMessage, error) {
	return c.request("GET", "/api/v1/projects", nil, nil, nil, http.StatusOK)
}

func (c *Client) GetProject(id string) (json.RawMessage, error) {
	return c.request("GET", "/api/v1/projects/"+url.PathEscape(id), nil, nil, nil, http.StatusOK)
}

func (c *Client) CreateProject(payload interface{}) (json.RawMessage, error) {
	return c.request("POST", "/api/v1/projects", payload, nil, nil, http.StatusOK, http.StatusCreated)
}

func (c *Client) UpdateProject(id string, payload interface{}) (json.RawMessage, error) {
	return c.request("PUT", "/api/v1/projects/"+url.PathEscape(id), payload, nil, nil, http.StatusOK)
}

func (c *Client) DeleteProject(id string) (json.RawMessage, error) {
	return c.request("DELETE", "/api/v1/projects/"+url.PathEscape(id), nil, nil, nil, http.StatusNoContent)
}

func (c *Client) GetProjectWorkspace(id string) (json.RawMessage, error) {
	return c.request("GET", "/api/v1/projects/"+url.PathEscape(id)+"/workspace", nil, nil, nil, http.StatusOK)
}

func (c *Client) SetProjectWorkspace(id string, payload interface{}) (json.RawMessage, error) {
	return c.request("PUT", "/api/v1/projects/"+url.PathEscape(id)+"/workspace", payload, nil, nil, http.StatusOK)
}

func (c *Client) PublishMap(payload interface{}) (json.RawMessage, error) {
	return c.request("POST", "/api/v1/maps/publish", payload, nil, nil, http.StatusOK, http.StatusCreated)
}

func (c *Client) ListPublishedMaps() (json.RawMessage, error) {
	return c.request("GET", "/api/v1/maps/published", nil, nil, nil, http.StatusOK)
}

func (c *Client) DeletePublishedMap(token string) (json.RawMessage, error) {
	return c.request("DELETE", "/api/v1/maps/published/"+url.PathEscape(token), nil, nil, nil, http.StatusNoContent)
}

func (c *Client) GetPublishedMapStats(token string) (json.RawMessage, error) {
	return c.request("GET", "/api/v1/maps/published/"+url.PathEscape(token)+"/stats", nil, nil, nil, http.StatusOK)
}

func (c *Client) UpdateMapEmbedConfig(token string, payload interface{}) (json.RawMessage, error) {
	return c.request("PUT", "/api/v1/maps/published/"+url.PathEscape(token)+"/embed-config", payload, nil, nil, http.StatusOK)
}

func projectIDJSONValue(projectID string) interface{} {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil
	}
	if n, err := strconv.ParseInt(projectID, 10, 64); err == nil {
		return n
	}
	return projectID
}

func truncate(b []byte, n int) string {
	s := string(b)
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}
