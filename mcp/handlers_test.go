package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestRunWithIOParseError(t *testing.T) {
	srv := testServer(t, http.NotFoundHandler())

	var out bytes.Buffer
	if err := srv.RunWithIO(strings.NewReader("{not-json}\n"), &out); err != nil {
		t.Fatalf("RunWithIO error: %v", err)
	}

	var resp jsonRPCResponse
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response error: %v\nraw: %s", err, out.String())
	}
	if resp.Error == nil {
		t.Fatal("expected parse error response")
	}
	if resp.Error.Code != -32700 {
		t.Fatalf("error code = %d, want -32700", resp.Error.Code)
	}
}

func TestRunWithIOInitializedNotificationHasNoResponse(t *testing.T) {
	srv := testServer(t, http.NotFoundHandler())
	req := `{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}` + "\n"

	var out bytes.Buffer
	if err := srv.RunWithIO(strings.NewReader(req), &out); err != nil {
		t.Fatalf("RunWithIO error: %v", err)
	}
	if out.Len() != 0 {
		t.Fatalf("expected no response, got %q", out.String())
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

func TestToolsCallBrowseCatalogAcceptsNumericPagination(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/catalog", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("limit"); got != "25" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error":"limit=%s"}`, got)
			return
		}
		if got := r.URL.Query().Get("offset"); got != "10" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error":"offset=%s"}`, got)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[{"id":"dataset-1"}]`)
	})
	srv := testServer(t, mux)

	resp := sendRequest(t, srv, "tools/call", 102, map[string]interface{}{
		"name": "browse_catalog",
		"arguments": map[string]interface{}{
			"search": "roads",
			"limit":  25.0,
			"offset": 10.0,
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

func TestToolsCallSearchSTACAcceptsNumericLimit(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /stac/search", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("limit"); got != "3" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error":"limit=%s"}`, got)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"features":[]}`)
	})
	srv := testServer(t, mux)

	resp := sendRequest(t, srv, "tools/call", 103, map[string]interface{}{
		"name": "search_stac",
		"arguments": map[string]interface{}{
			"collections": "imagery",
			"limit":       3.0,
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

func TestToolsCallMapAPIMutationRequiresConfirm(t *testing.T) {
	srv := testServer(t, http.NotFoundHandler())

	resp := sendRequest(t, srv, "tools/call", 104, map[string]interface{}{
		"name": "map_api",
		"arguments": map[string]interface{}{
			"operation": "publish_map",
			"body":      map[string]interface{}{"map_id": "abc"},
		},
	})

	if resp.Error != nil {
		t.Fatalf("unexpected JSON-RPC error: %v", resp.Error)
	}
	result, _ := resp.Result.(map[string]interface{})
	if isErr, _ := result["isError"].(bool); !isErr {
		t.Fatalf("expected error result, got: %+v", result)
	}
	content, _ := result["content"].([]interface{})
	first, _ := content[0].(map[string]interface{})
	text, _ := first["text"].(string)
	if !strings.Contains(text, "confirm=true") {
		t.Fatalf("error text = %q, want confirm hint", text)
	}
}

func TestToolsCallMapAPIRejectsInvalidQueryObject(t *testing.T) {
	srv := testServer(t, http.NotFoundHandler())

	resp := sendRequest(t, srv, "tools/call", 105, map[string]interface{}{
		"name": "map_api",
		"arguments": map[string]interface{}{
			"operation": "get_raster_values",
			"name":      "elevation",
			"query":     "x=1",
		},
	})

	if resp.Error != nil {
		t.Fatalf("unexpected JSON-RPC error: %v", resp.Error)
	}
	result, _ := resp.Result.(map[string]interface{})
	if isErr, _ := result["isError"].(bool); !isErr {
		t.Fatalf("expected error result, got: %+v", result)
	}
	content, _ := result["content"].([]interface{})
	first, _ := content[0].(map[string]interface{})
	text, _ := first["text"].(string)
	if !strings.Contains(text, "query must be an object") {
		t.Fatalf("error text = %q, want invalid query message", text)
	}
}

func TestToolsCallMapAPIMissingPathParameter(t *testing.T) {
	srv := testServer(t, http.NotFoundHandler())

	resp := sendRequest(t, srv, "tools/call", 106, map[string]interface{}{
		"name": "map_api",
		"arguments": map[string]interface{}{
			"operation": "get_public_map",
		},
	})

	if resp.Error != nil {
		t.Fatalf("unexpected JSON-RPC error: %v", resp.Error)
	}
	result, _ := resp.Result.(map[string]interface{})
	if isErr, _ := result["isError"].(bool); !isErr {
		t.Fatalf("expected error result, got: %+v", result)
	}
	content, _ := result["content"].([]interface{})
	first, _ := content[0].(map[string]interface{})
	text, _ := first["text"].(string)
	if !strings.Contains(text, "missing required path parameter: token") {
		t.Fatalf("error text = %q, want missing path parameter", text)
	}
}

func TestParseRoutePoint(t *testing.T) {
	t.Run("object", func(t *testing.T) {
		got, err := parseRoutePoint(map[string]interface{}{"lat": 39.1, "lon": -86.2})
		if err != nil {
			t.Fatalf("parseRoutePoint returned error: %v", err)
		}
		want := [2]float64{-86.2, 39.1}
		if got != want {
			t.Fatalf("point = %#v, want %#v", got, want)
		}
	})

	t.Run("array", func(t *testing.T) {
		got, err := parseRoutePoint([]interface{}{-86.2, 39.1})
		if err != nil {
			t.Fatalf("parseRoutePoint returned error: %v", err)
		}
		want := [2]float64{-86.2, 39.1}
		if got != want {
			t.Fatalf("point = %#v, want %#v", got, want)
		}
	})

	t.Run("error", func(t *testing.T) {
		_, err := parseRoutePoint(map[string]interface{}{"lon": -86.2})
		if err == nil || !strings.Contains(err.Error(), "missing lat") {
			t.Fatalf("error = %v, want missing lat", err)
		}
	})
}

func TestParseRoutePoints(t *testing.T) {
	got, err := parseRoutePoints([]interface{}{
		map[string]interface{}{"lat": 39.0, "lon": -86.0},
		[]interface{}{-86.1, 39.1},
	})
	if err != nil {
		t.Fatalf("parseRoutePoints returned error: %v", err)
	}
	want := [][2]float64{{-86.0, 39.0}, {-86.1, 39.1}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("points = %#v, want %#v", got, want)
	}

	_, err = parseRoutePoints([]interface{}{map[string]interface{}{"lat": 39.0}})
	if err == nil || !strings.Contains(err.Error(), "index 0") {
		t.Fatalf("error = %v, want indexed parse error", err)
	}
}

func TestExtractQuery(t *testing.T) {
	got, err := extractQuery(map[string]interface{}{
		"query": map[string]interface{}{
			"limit":   5.0,
			"exact":   true,
			"dataset": "roads",
			"empty":   nil,
		},
	})
	if err != nil {
		t.Fatalf("extractQuery returned error: %v", err)
	}
	want := map[string]string{
		"limit":   "5",
		"exact":   "true",
		"dataset": "roads",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("query = %#v, want %#v", got, want)
	}

	_, err = extractQuery(map[string]interface{}{"query": "bad"})
	if err == nil || !strings.Contains(err.Error(), "query must be an object") {
		t.Fatalf("error = %v, want query object error", err)
	}
}

func TestInterpolatePath(t *testing.T) {
	got, err := interpolatePath("/collections/{collection_id}/items/{feature_id}", map[string]interface{}{
		"collection_id": "roads/2024",
		"feature_id":    "abc 123",
	})
	if err != nil {
		t.Fatalf("interpolatePath returned error: %v", err)
	}
	if got != "/collections/roads%2F2024/items/abc%20123" {
		t.Fatalf("path = %q, want %q", got, "/collections/roads%2F2024/items/abc%20123")
	}

	_, err = interpolatePath("/collections/{collection_id}", map[string]interface{}{})
	if err == nil || !strings.Contains(err.Error(), "missing required path parameter: collection_id") {
		t.Fatalf("error = %v, want missing path parameter", err)
	}
}

func TestNormalizeProcessPayload(t *testing.T) {
	params := map[string]interface{}{
		"output": "buffered_roads",
		"format": "parquet",
	}
	normalizeProcessPayload(params, "42")

	if got := params["output_name"]; got != "buffered_roads" {
		t.Fatalf("output_name = %v, want buffered_roads", got)
	}
	if got := params["output_format"]; got != "parquet" {
		t.Fatalf("output_format = %v, want parquet", got)
	}
	if _, ok := params["params"].(map[string]interface{}); !ok {
		t.Fatalf("params = %#v, want initialized params object", params["params"])
	}
	if got := params["project_id"]; got != int64(42) {
		t.Fatalf("project_id = %v, want 42", got)
	}
}

func TestOptionalProjectIDAndEnsureProjectID(t *testing.T) {
	projectID, err := optionalProjectID(map[string]interface{}{"project_id": 7.0})
	if err != nil {
		t.Fatalf("optionalProjectID returned error: %v", err)
	}
	if projectID != "7" {
		t.Fatalf("projectID = %q, want 7", projectID)
	}

	params := map[string]interface{}{}
	ensureProjectID(params, "15")
	if got := params["project_id"]; got != int64(15) {
		t.Fatalf("project_id = %v, want 15", got)
	}

	ensureProjectID(params, "19")
	if got := params["project_id"]; got != int64(15) {
		t.Fatalf("project_id should remain unchanged, got %v", got)
	}
}

func TestFormatJSON(t *testing.T) {
	pretty := formatJSON(json.RawMessage(`{"ok":true}`))
	if !strings.Contains(pretty, "\n") {
		t.Fatalf("expected pretty JSON, got %q", pretty)
	}

	raw := formatJSON(json.RawMessage(`not-json`))
	if raw != "not-json" {
		t.Fatalf("formatJSON fallback = %q, want %q", raw, "not-json")
	}
}

func TestHandleSubmitProcessBatchValidation(t *testing.T) {
	client := NewClient("https://example.test", "")

	_, err := handleSubmitProcessBatch(client, map[string]interface{}{})
	if err == nil || !strings.Contains(err.Error(), "parameter jobs must be a non-empty array") {
		t.Fatalf("error = %v, want jobs array error", err)
	}

	_, err = handleSubmitProcessBatch(client, map[string]interface{}{
		"jobs": []interface{}{"bad"},
	})
	if err == nil || !strings.Contains(err.Error(), "jobs[0] must be an object") {
		t.Fatalf("error = %v, want job object error", err)
	}

	_, err = handleSubmitProcessBatch(client, map[string]interface{}{
		"jobs": []interface{}{
			map[string]interface{}{"request": "bad"},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "jobs[0].request must be an object") {
		t.Fatalf("error = %v, want request object error", err)
	}
}
