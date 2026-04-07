package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

func TestToolsListCurrentSurface(t *testing.T) {
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
		t.Fatal("tools list should not be empty")
	}

	toolNames := make(map[string]bool, len(tools))
	for _, item := range tools {
		m, _ := item.(map[string]interface{})
		name, _ := m["name"].(string)
		toolNames[name] = true
	}

	for _, want := range []string{
		"list_datasets",
		"import_source",
		"get_scene_manifest",
		"list_bodies",
		"execute_body_recipe",
		"list_operations",
		"submit_operation_job",
		"run_pipeline",
		"list_query_engines",
		"execute_sql",
		"save_sql_result",
		"list_projects",
		"set_project_workspace",
		"publish_map",
		"update_map_embed_config",
	} {
		if !toolNames[want] {
			t.Errorf("missing tool: %s", want)
		}
	}

	for _, removed := range []string{
		"run_process",
		"run_raster_process",
		"convert_format",
		"diff_datasets",
		"browse_catalog",
		"search_stac",
		"map_api",
	} {
		if toolNames[removed] {
			t.Errorf("legacy tool should have been removed: %s", removed)
		}
	}
}

func TestToolsCallQueryFeaturesAcceptsNumericPagination(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /collections/{id}/items", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("limit"); got != "5" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error":"limit=%s"}`, got)
			return
		}
		if got := r.URL.Query().Get("offset"); got != "2" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error":"offset=%s"}`, got)
			return
		}
		if got := r.Header.Get("X-Project-ID"); got != "42" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error":"project=%s"}`, got)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"type":"FeatureCollection","features":[]}`)
	})
	srv := testServer(t, mux)

	resp := sendRequest(t, srv, "tools/call", 101, map[string]interface{}{
		"name": "query_features",
		"arguments": map[string]interface{}{
			"collection_id": "buildings",
			"limit":         5.0,
			"offset":        2.0,
			"project_id":    42.0,
		},
	})

	if resp.Error != nil {
		t.Fatalf("unexpected JSON-RPC error: %v", resp.Error)
	}
	result, _ := resp.Result.(map[string]interface{})
	if isErr, _ := result["isError"].(bool); isErr {
		t.Fatalf("expected success result, got error: %+v", result)
	}
}

func TestToolsCallUploadDatasetIncludesBodyID(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "roads.geojson")
	if err := os.WriteFile(filePath, []byte(`{"type":"FeatureCollection","features":[]}`), 0o600); err != nil {
		t.Fatalf("write temp upload file: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /upload", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Project-ID"); got != "42" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error":"project=%s"}`, got)
			return
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error":"parse=%v"}`, err)
			return
		}
		if got := r.FormValue("body_id"); got != "mars" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error":"body_id=%s"}`, got)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"name":"roads","body_id":"mars"}`)
	})
	srv := testServer(t, mux)

	resp := sendRequest(t, srv, "tools/call", 102, map[string]interface{}{
		"name": "upload_dataset",
		"arguments": map[string]interface{}{
			"file_path":  filePath,
			"name":       "roads",
			"body_id":    "mars",
			"project_id": 42.0,
		},
	})

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result, _ := resp.Result.(map[string]interface{})
	content, _ := result["content"].([]interface{})
	first, _ := content[0].(map[string]interface{})
	text, _ := first["text"].(string)
	if !strings.Contains(text, "mars") {
		t.Fatalf("response should contain body_id, got: %s", text)
	}
}

func TestToolsCallExecuteSQLUsesEngineAwareRoute(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/query/sql", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("engine"); got != "duckdb" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error":"engine=%s"}`, got)
			return
		}
		var body struct {
			SQL string `json:"sql"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"columns":["count"],"rows":[[42]],"sql":%q}`, body.SQL)
	})
	srv := testServer(t, mux)

	resp := sendRequest(t, srv, "tools/call", 103, map[string]interface{}{
		"name": "execute_sql",
		"arguments": map[string]interface{}{
			"engine": "duckdb",
			"query":  "SELECT count(*) FROM parks",
		},
	})

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result, _ := resp.Result.(map[string]interface{})
	isErr, _ := result["isError"].(bool)
	if isErr {
		t.Fatalf("expected success result, got error: %+v", result)
	}
	content, _ := result["content"].([]interface{})
	first, _ := content[0].(map[string]interface{})
	text, _ := first["text"].(string)
	if !strings.Contains(text, "SELECT count(*) FROM parks") {
		t.Fatalf("response should contain SQL text, got: %s", text)
	}
}

func TestToolsCallListBodies(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/bodies", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[{"slug":"earth"},{"slug":"mars"}]`)
	})
	srv := testServer(t, mux)

	resp := sendRequest(t, srv, "tools/call", 104, map[string]interface{}{
		"name":      "list_bodies",
		"arguments": map[string]interface{}{},
	})

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result, _ := resp.Result.(map[string]interface{})
	content, _ := result["content"].([]interface{})
	first, _ := content[0].(map[string]interface{})
	text, _ := first["text"].(string)
	if !strings.Contains(text, "mars") {
		t.Fatalf("response should contain mars, got: %s", text)
	}
}

func TestToolsCallUnknownTool(t *testing.T) {
	srv := testServer(t, http.NotFoundHandler())

	resp := sendRequest(t, srv, "tools/call", 105, map[string]interface{}{
		"name":      "nonexistent_tool",
		"arguments": map[string]interface{}{},
	})

	if resp.Error != nil {
		t.Fatalf("unexpected JSON-RPC error: %v", resp.Error)
	}
	result, _ := resp.Result.(map[string]interface{})
	isErr, _ := result["isError"].(bool)
	if !isErr {
		t.Fatal("should be an error result")
	}
}

func TestPing(t *testing.T) {
	srv := testServer(t, http.NotFoundHandler())
	resp := sendRequest(t, srv, "ping", 106, nil)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
}

func TestUnknownMethod(t *testing.T) {
	srv := testServer(t, http.NotFoundHandler())
	resp := sendRequest(t, srv, "nonexistent/method", 107, nil)
	if resp.Error == nil {
		t.Fatal("expected an error for unknown method")
	}
	if resp.Error.Code != -32601 {
		t.Fatalf("error code = %d, want -32601", resp.Error.Code)
	}
}

func TestToolsCallMissingRequiredParam(t *testing.T) {
	srv := testServer(t, http.NotFoundHandler())

	resp := sendRequest(t, srv, "tools/call", 108, map[string]interface{}{
		"name":      "get_dataset_info",
		"arguments": map[string]interface{}{},
	})

	result, _ := resp.Result.(map[string]interface{})
	isErr, _ := result["isError"].(bool)
	if !isErr {
		t.Fatal("should be an error when required param is missing")
	}
	content, _ := result["content"].([]interface{})
	first, _ := content[0].(map[string]interface{})
	text, _ := first["text"].(string)
	if !strings.Contains(text, "name") {
		t.Fatalf("error should mention missing param, got: %s", text)
	}
}
