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

// testServer creates a test Roteiro API server and returns a connected MCP server.
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
		"upload_dataset", "run_process", "run_raster_process", "preflight_process", "submit_process_job",
		"submit_process_batch", "list_process_jobs", "get_process_job",
		"cancel_process_job", "rerun_process_job", "run_pipeline", "convert_format",
		"diff_datasets", "execute_sql", "list_spatial_tables", "get_duckdb_info",
		"list_duckdb_datasets", "geocode", "reverse_geocode", "compute_route",
		"compute_isochrone", "compute_route_matrix", "compute_service_area",
		"list_operations", "list_analysis_operations", "browse_catalog", "browse_catalog_enhanced",
		"get_catalog_entry", "list_catalog_categories", "list_catalog_tags", "import_from_catalog", "browse_stac_catalog",
		"browse_stac_collections", "browse_stac_items", "import_stac_asset",
		"search_stac",
	} {
		if !toolNames[want] {
			t.Errorf("missing tool: %s", want)
		}
	}
}

func TestToolsCall_ListAnalysisOperations(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/analysis/operations", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"operations":[{"id":"topology","name":"Topology Analysis"}]}`)
	})
	srv := testServer(t, mux)

	resp := sendRequest(t, srv, "tools/call", 22, map[string]interface{}{
		"name":      "list_analysis_operations",
		"arguments": map[string]interface{}{},
	})

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result, _ := resp.Result.(map[string]interface{})
	content, _ := result["content"].([]interface{})
	first, _ := content[0].(map[string]interface{})
	text, _ := first["text"].(string)
	if !strings.Contains(text, "Topology Analysis") {
		t.Errorf("response should contain operation name, got: %s", text)
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
	mux.HandleFunc("POST /api/query/sql", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			SQL string `json:"sql"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"columns":["count"],"rows":[[42]],"sql":"%s"}`, body.SQL)
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

func TestToolsCall_ConvertFormat_MapsFormatToOutputFormat(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/convert", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		if body["output_format"] != "parquet" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error":"missing output_format"}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"message":"conversion complete"}`)
	})
	srv := testServer(t, mux)

	resp := sendRequest(t, srv, "tools/call", 13, map[string]interface{}{
		"name": "convert_format",
		"arguments": map[string]interface{}{
			"input":  "parks",
			"format": "parquet",
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
}

func TestToolsCall_DiffDatasets_MapsBaseCompare(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/diff", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		if body["left"] != "v1" || body["right"] != "v2" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error":"missing left/right"}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"added":1,"removed":0,"modified":2}`)
	})
	srv := testServer(t, mux)

	resp := sendRequest(t, srv, "tools/call", 14, map[string]interface{}{
		"name": "diff_datasets",
		"arguments": map[string]interface{}{
			"base":    "v1",
			"compare": "v2",
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
}

func TestToolsCall_ComputeRoute_MapsOriginDestination(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/route", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Waypoints [][2]float64 `json:"waypoints"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		if len(body.Waypoints) != 2 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error":"expected 2 waypoints"}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"distance":1000,"duration":120}`)
	})
	srv := testServer(t, mux)

	resp := sendRequest(t, srv, "tools/call", 15, map[string]interface{}{
		"name": "compute_route",
		"arguments": map[string]interface{}{
			"origin":      map[string]interface{}{"lat": 39.0, "lon": -86.0},
			"destination": map[string]interface{}{"lat": 39.1, "lon": -86.1},
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
}

func TestToolsCall_ComputeRouteMatrix_MapsPoints(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/route/matrix", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Origins      [][2]float64 `json:"origins"`
			Destinations [][2]float64 `json:"destinations"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		if len(body.Origins) != 1 || len(body.Destinations) != 1 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error":"expected origins/destinations"}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"durations":[[120]],"distances":[[1000]]}`)
	})
	srv := testServer(t, mux)

	resp := sendRequest(t, srv, "tools/call", 16, map[string]interface{}{
		"name": "compute_route_matrix",
		"arguments": map[string]interface{}{
			"origins":      []interface{}{map[string]interface{}{"lat": 39.0, "lon": -86.0}},
			"destinations": []interface{}{map[string]interface{}{"lat": 39.1, "lon": -86.1}},
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
}

func TestToolsCall_ComputeIsochrone_MapsOrigin(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/route/isochrone", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		if body["lng"] == nil || body["lat"] == nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error":"missing lng/lat"}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"type":"FeatureCollection","features":[]}`)
	})
	srv := testServer(t, mux)

	resp := sendRequest(t, srv, "tools/call", 17, map[string]interface{}{
		"name": "compute_isochrone",
		"arguments": map[string]interface{}{
			"origin":  map[string]interface{}{"lat": 39.0, "lon": -86.0},
			"minutes": []interface{}{10.0, 20.0},
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
}

func TestToolsCall_GetDuckDBInfo(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/query/sql/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"available"}`)
	})
	srv := testServer(t, mux)

	resp := sendRequest(t, srv, "tools/call", 18, map[string]interface{}{
		"name":      "get_duckdb_info",
		"arguments": map[string]interface{}{},
	})

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result, _ := resp.Result.(map[string]interface{})
	content, _ := result["content"].([]interface{})
	first, _ := content[0].(map[string]interface{})
	text, _ := first["text"].(string)
	if !strings.Contains(text, "available") {
		t.Errorf("response should contain duckdb status, got: %s", text)
	}
}

func TestToolsCall_BrowseEnhancedCatalog(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/catalog/enhanced", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("live_only") != "true" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error":"missing live_only"}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[{"id":"us-census","name":"US Census"}]`)
	})
	srv := testServer(t, mux)

	resp := sendRequest(t, srv, "tools/call", 19, map[string]interface{}{
		"name": "browse_catalog_enhanced",
		"arguments": map[string]interface{}{
			"search":    "census",
			"live_only": true,
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
		w.WriteHeader(http.StatusAccepted)
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
		if body["output_name"] != "parks_buffered" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error":"missing output_name"}`)
			return
		}
		if body["output_format"] != "parquet" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error":"missing output_format"}`)
			return
		}
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
			"output":    "parks_buffered",
			"format":    "parquet",
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

func TestToolsCall_PreflightProcess(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/process/preflight", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		if body["output_name"] != "parks_buffered" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error":"missing output_name"}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"valid":true,"resolved_params":{"distance":500}}`)
	})
	srv := testServer(t, mux)

	resp := sendRequest(t, srv, "tools/call", 23, map[string]interface{}{
		"name": "preflight_process",
		"arguments": map[string]interface{}{
			"operation": "buffer",
			"input":     "parks",
			"params":    map[string]interface{}{"distance": 500},
			"output":    "parks_buffered",
		},
	})

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result, _ := resp.Result.(map[string]interface{})
	content, _ := result["content"].([]interface{})
	first, _ := content[0].(map[string]interface{})
	text, _ := first["text"].(string)
	if !strings.Contains(text, `"valid": true`) {
		t.Errorf("response should contain valid preflight, got: %s", text)
	}
}

func TestToolsCall_RunRasterProcess(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/raster/process", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		if body["operation"] != "slope" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error":"missing operation"}`)
			return
		}
		if _, ok := body["params"]; !ok {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error":"missing params"}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"width":2,"height":2,"data":[1,2,3,4]}`)
	})
	srv := testServer(t, mux)

	resp := sendRequest(t, srv, "tools/call", 231, map[string]interface{}{
		"name": "run_raster_process",
		"arguments": map[string]interface{}{
			"operation":  "slope",
			"input_path": "/data/dem.tif",
		},
	})

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result, _ := resp.Result.(map[string]interface{})
	content, _ := result["content"].([]interface{})
	first, _ := content[0].(map[string]interface{})
	text, _ := first["text"].(string)
	if !strings.Contains(text, `"width": 2`) {
		t.Errorf("response should contain raster result, got: %s", text)
	}
}

func TestToolsCall_MapAPIExportRasterBand(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /raster/{name}/export", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("name") != "elevation" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message":"exported to /tmp/elevation.tif"}`)
	})
	srv := testServer(t, mux)

	resp := sendRequest(t, srv, "tools/call", 232, map[string]interface{}{
		"name": "map_api",
		"arguments": map[string]interface{}{
			"operation": "export_raster_band",
			"name":      "elevation",
			"body": map[string]interface{}{
				"output_path": "elevation.tif",
				"band":        0,
			},
		},
	})

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result, _ := resp.Result.(map[string]interface{})
	content, _ := result["content"].([]interface{})
	first, _ := content[0].(map[string]interface{})
	text, _ := first["text"].(string)
	if !strings.Contains(text, "exported to /tmp/elevation.tif") {
		t.Errorf("response should contain export message, got: %s", text)
	}
}

func TestToolsCall_MapAPIContour(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /raster/{name}/contour", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("name") != "elevation" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"type":"FeatureCollection","features":[]}`)
	})
	srv := testServer(t, mux)

	resp := sendRequest(t, srv, "tools/call", 233, map[string]interface{}{
		"name": "map_api",
		"arguments": map[string]interface{}{
			"operation": "raster_contour",
			"name":      "elevation",
			"body": map[string]interface{}{
				"interval": 10,
			},
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
		t.Errorf("response should contain contour GeoJSON, got: %s", text)
	}
}

func TestToolsCall_MapAPIKDE(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/raster/kde", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"width":2,"height":2,"data":[0.1,0.2,0.3,0.4]}`)
	})
	srv := testServer(t, mux)

	resp := sendRequest(t, srv, "tools/call", 234, map[string]interface{}{
		"name": "map_api",
		"arguments": map[string]interface{}{
			"operation": "raster_kde",
			"body": map[string]interface{}{
				"dataset":   "points",
				"bandwidth": 50,
			},
		},
	})

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result, _ := resp.Result.(map[string]interface{})
	content, _ := result["content"].([]interface{})
	first, _ := content[0].(map[string]interface{})
	text, _ := first["text"].(string)
	if !strings.Contains(text, `"width": 2`) {
		t.Errorf("response should contain kde grid, got: %s", text)
	}
}

func TestToolsCall_SubmitProcessJob(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/process/jobs", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		if body["output_name"] != "parks_buffered" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error":"missing output_name"}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, `{"id":"job_123","status":"queued"}`)
	})
	srv := testServer(t, mux)

	resp := sendRequest(t, srv, "tools/call", 24, map[string]interface{}{
		"name": "submit_process_job",
		"arguments": map[string]interface{}{
			"operation": "buffer",
			"input":     "parks",
			"params":    map[string]interface{}{"distance": 500},
			"output":    "parks_buffered",
		},
	})

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result, _ := resp.Result.(map[string]interface{})
	content, _ := result["content"].([]interface{})
	first, _ := content[0].(map[string]interface{})
	text, _ := first["text"].(string)
	if !strings.Contains(text, "job_123") {
		t.Errorf("response should contain job id, got: %s", text)
	}
}

func TestToolsCall_ListProcessJobs(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/process/jobs", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("status") != "queued" || r.URL.Query().Get("limit") != "25" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error":"missing filters"}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[{"id":"job_123","status":"queued"}]`)
	})
	srv := testServer(t, mux)

	resp := sendRequest(t, srv, "tools/call", 25, map[string]interface{}{
		"name": "list_process_jobs",
		"arguments": map[string]interface{}{
			"status": "queued",
			"limit":  25.0,
		},
	})

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result, _ := resp.Result.(map[string]interface{})
	content, _ := result["content"].([]interface{})
	first, _ := content[0].(map[string]interface{})
	text, _ := first["text"].(string)
	if !strings.Contains(text, "job_123") {
		t.Errorf("response should contain job id, got: %s", text)
	}
}

func TestToolsCall_CancelProcessJob(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /api/process/jobs/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "job_123" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	srv := testServer(t, mux)

	resp := sendRequest(t, srv, "tools/call", 26, map[string]interface{}{
		"name": "cancel_process_job",
		"arguments": map[string]interface{}{
			"job_id": "job_123",
		},
	})

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result, _ := resp.Result.(map[string]interface{})
	content, _ := result["content"].([]interface{})
	first, _ := content[0].(map[string]interface{})
	text, _ := first["text"].(string)
	if !strings.Contains(text, "cancelled") {
		t.Errorf("response should contain cancellation status, got: %s", text)
	}
}
