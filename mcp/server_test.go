package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// testServer creates a test Cairn API server and returns a connected MCP server.
func testServer(t *testing.T, handler http.Handler) *Server {
	t.Helper()
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	client := NewClient(ts.URL, "test-key")
	return NewServer(client)
}

func sendRequest(t *testing.T, server *Server, method string, id int, params interface{}) jsonRPCResponse {
	t.Helper()
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
	}
	if params != nil {
		req["params"] = params
	}
	line, _ := json.Marshal(req)

	var out bytes.Buffer
	if err := server.RunWithIO(bytes.NewReader(append(line, '\n')), &out); err != nil {
		t.Fatalf("RunWithIO error: %v", err)
	}

	var resp jsonRPCResponse
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response error: %v\nraw: %s", err, out.String())
	}
	return resp
}

func TestInitialize(t *testing.T) {
	srv := testServer(t, http.NotFoundHandler())
	resp := sendRequest(t, srv, "initialize", 1, map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"clientInfo":      map[string]string{"name": "test", "version": "1.0"},
	})

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("result is not a map")
	}
	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("protocolVersion = %v, want 2024-11-05", result["protocolVersion"])
	}
	info, _ := result["serverInfo"].(map[string]interface{})
	if info["name"] != "roteiro-agent" {
		t.Errorf("serverInfo.name = %v, want roteiro-agent", info["name"])
	}
}

func TestToolsList(t *testing.T) {
	srv := testServer(t, http.NotFoundHandler())
	resp := sendRequest(t, srv, "tools/list", 2, nil)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("result is not a map")
	}
	tools, ok := result["tools"].([]interface{})
	if !ok {
		t.Fatal("tools is not an array")
	}
	if len(tools) == 0 {
		t.Error("tools list should not be empty")
	}

	// Verify expected tools are present.
	toolNames := make(map[string]bool)
	for _, t := range tools {
		m, _ := t.(map[string]interface{})
		name, _ := m["name"].(string)
		toolNames[name] = true
	}
	for _, want := range []string{
		"list_datasets", "get_dataset_info", "query_features", "get_feature",
		"upload_dataset", "run_process", "run_pipeline", "convert_format",
		"diff_datasets", "execute_sql", "list_spatial_tables", "geocode",
		"reverse_geocode", "compute_route", "list_operations",
		"browse_catalog", "import_from_catalog", "browse_stac_catalog",
		"browse_stac_collections", "browse_stac_items", "import_stac_asset",
		"search_stac",
	} {
		if !toolNames[want] {
			t.Errorf("missing tool: %s", want)
		}
	}
}

func TestToolsCall_ListDatasets(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /datasets", func(w http.ResponseWriter, r *http.Request) {
		// Verify API key is passed.
		if r.Header.Get("X-API-Key") != "test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[{"name":"parks","format":"GeoJSON","feature_count":42}]`)
	})
	srv := testServer(t, mux)

	resp := sendRequest(t, srv, "tools/call", 3, map[string]interface{}{
		"name":      "list_datasets",
		"arguments": map[string]interface{}{},
	})

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result, _ := resp.Result.(map[string]interface{})
	content, _ := result["content"].([]interface{})
	if len(content) == 0 {
		t.Fatal("expected content")
	}
	first, _ := content[0].(map[string]interface{})
	text, _ := first["text"].(string)
	if !strings.Contains(text, "parks") {
		t.Errorf("response should contain 'parks', got: %s", text)
	}
}

func TestToolsCall_QueryFeatures(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /collections/{id}/items", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id != "buildings" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		limit := r.URL.Query().Get("limit")
		if limit == "" {
			limit = "10"
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"type":"FeatureCollection","features":[],"numberMatched":0,"limit":%s}`, limit)
	})
	srv := testServer(t, mux)

	resp := sendRequest(t, srv, "tools/call", 4, map[string]interface{}{
		"name": "query_features",
		"arguments": map[string]interface{}{
			"collection_id": "buildings",
			"limit":         "5",
		},
	})

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result, _ := resp.Result.(map[string]interface{})
	content, _ := result["content"].([]interface{})
	first, _ := content[0].(map[string]interface{})
	text, _ := first["text"].(string)
	if !strings.Contains(text, "FeatureCollection") {
		t.Errorf("response should contain 'FeatureCollection', got: %s", text)
	}
}

func TestToolsCall_ExecuteSQL(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/sql/query", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Query string `json:"query"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"columns":["count"],"rows":[[42]],"query":"%s"}`, body.Query)
	})
	srv := testServer(t, mux)

	resp := sendRequest(t, srv, "tools/call", 5, map[string]interface{}{
		"name": "execute_sql",
		"arguments": map[string]interface{}{
			"query": "SELECT count(*) FROM parks",
		},
	})

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result, _ := resp.Result.(map[string]interface{})
	isErr, _ := result["isError"].(bool)
	if isErr {
		t.Error("should not be an error")
	}
	content, _ := result["content"].([]interface{})
	first, _ := content[0].(map[string]interface{})
	text, _ := first["text"].(string)
	if !strings.Contains(text, "42") {
		t.Errorf("response should contain '42', got: %s", text)
	}
}

func TestToolsCall_UnknownTool(t *testing.T) {
	srv := testServer(t, http.NotFoundHandler())

	resp := sendRequest(t, srv, "tools/call", 6, map[string]interface{}{
		"name":      "nonexistent_tool",
		"arguments": map[string]interface{}{},
	})

	if resp.Error != nil {
		t.Fatalf("unexpected JSON-RPC error: %v", resp.Error)
	}
	result, _ := resp.Result.(map[string]interface{})
	isErr, _ := result["isError"].(bool)
	if !isErr {
		t.Error("should be an error result")
	}
}

func TestPing(t *testing.T) {
	srv := testServer(t, http.NotFoundHandler())
	resp := sendRequest(t, srv, "ping", 7, nil)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
}

func TestUnknownMethod(t *testing.T) {
	srv := testServer(t, http.NotFoundHandler())
	resp := sendRequest(t, srv, "nonexistent/method", 8, nil)

	if resp.Error == nil {
		t.Fatal("expected an error for unknown method")
	}
	if resp.Error.Code != -32601 {
		t.Errorf("error code = %d, want -32601", resp.Error.Code)
	}
}

func TestToolsCall_MissingRequiredParam(t *testing.T) {
	srv := testServer(t, http.NotFoundHandler())

	resp := sendRequest(t, srv, "tools/call", 9, map[string]interface{}{
		"name":      "get_dataset_info",
		"arguments": map[string]interface{}{},
	})

	result, _ := resp.Result.(map[string]interface{})
	isErr, _ := result["isError"].(bool)
	if !isErr {
		t.Error("should be an error when required param is missing")
	}
	content, _ := result["content"].([]interface{})
	first, _ := content[0].(map[string]interface{})
	text, _ := first["text"].(string)
	if !strings.Contains(text, "collection_id") {
		t.Errorf("error should mention missing param, got: %s", text)
	}
}

func TestToolsCall_BrowseCatalog(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/catalog", func(w http.ResponseWriter, r *http.Request) {
		search := r.URL.Query().Get("search")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `[{"id":"us-census","name":"US Census","category":"boundaries","search":"%s"}]`, search)
	})
	srv := testServer(t, mux)

	resp := sendRequest(t, srv, "tools/call", 11, map[string]interface{}{
		"name": "browse_catalog",
		"arguments": map[string]interface{}{
			"search": "census",
		},
	})

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result, _ := resp.Result.(map[string]interface{})
	content, _ := result["content"].([]interface{})
	first, _ := content[0].(map[string]interface{})
	text, _ := first["text"].(string)
	if !strings.Contains(text, "us-census") {
		t.Errorf("response should contain 'us-census', got: %s", text)
	}
}

func TestToolsCall_ImportSTACAsset(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/stac/import", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			AssetURL string `json:"asset_url"`
			Name     string `json:"name"`
			Format   string `json:"format"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{"name":"%s","path":"data/%s.geojson","format":"geojson"}`, body.Name, body.Name)
	})
	srv := testServer(t, mux)

	resp := sendRequest(t, srv, "tools/call", 12, map[string]interface{}{
		"name": "import_stac_asset",
		"arguments": map[string]interface{}{
			"asset_url": "https://example.com/buildings.geojson",
			"name":      "buildings",
		},
	})

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result, _ := resp.Result.(map[string]interface{})
	content, _ := result["content"].([]interface{})
	first, _ := content[0].(map[string]interface{})
	text, _ := first["text"].(string)
	if !strings.Contains(text, "buildings") {
		t.Errorf("response should contain 'buildings', got: %s", text)
	}
}

func TestToolsCall_RunProcess(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/process", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"output":"buffered_%s","feature_count":10}`, body["input"])
	})
	srv := testServer(t, mux)

	resp := sendRequest(t, srv, "tools/call", 10, map[string]interface{}{
		"name": "run_process",
		"arguments": map[string]interface{}{
			"operation": "buffer",
			"input":     "parks",
			"params":    map[string]interface{}{"distance": 500},
		},
	})

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result, _ := resp.Result.(map[string]interface{})
	content, _ := result["content"].([]interface{})
	first, _ := content[0].(map[string]interface{})
	text, _ := first["text"].(string)
	if !strings.Contains(text, "buffered_parks") {
		t.Errorf("response should contain output name, got: %s", text)
	}
}
