package mcp

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListQueryDatasetsUsesEngineAwareEndpoint(t *testing.T) {
	t.Helper()
	var gotPath string
	var gotEngine string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotEngine = r.URL.Query().Get("engine")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"name":"roads"}]`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "")
	client.HTTPClient = srv.Client()

	body, err := client.ListQueryDatasets("duckdb")
	if err != nil {
		t.Fatalf("ListQueryDatasets error: %v", err)
	}
	if gotPath != "/api/v1/query/sql/datasets" {
		t.Fatalf("path = %q, want %q", gotPath, "/api/v1/query/sql/datasets")
	}
	if gotEngine != "duckdb" {
		t.Fatalf("engine = %q, want %q", gotEngine, "duckdb")
	}
	if string(body) != `[{"name":"roads"}]` {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestExecuteSQLUsesCurrentEndpointOnly(t *testing.T) {
	legacyCalled := false

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/query/sql":
			if got := r.URL.Query().Get("engine"); got != "duckdb" {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"wrong engine"}`))
				return
			}
			http.NotFound(w, r)
		case "/api/query/sql":
			legacyCalled = true
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"rows":[{"value":1}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "")
	client.HTTPClient = srv.Client()

	_, err := client.ExecuteSQL("duckdb", map[string]interface{}{"sql": "SELECT 1"})
	if err == nil {
		t.Fatal("expected ExecuteSQL to fail when the current endpoint is unavailable")
	}
	if legacyCalled {
		t.Fatal("legacy /api/query/sql route should not be called")
	}
}

func TestDeleteFeatureSynthesizesOKResponseOnNoContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/collections/roads/items/f-1" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "")
	client.HTTPClient = srv.Client()

	body, err := client.DeleteFeature("roads", "f-1")
	if err != nil {
		t.Fatalf("DeleteFeature error: %v", err)
	}
	if string(body) != `{"status":"ok"}` {
		t.Fatalf("body = %q, want ok status JSON", body)
	}
}

func TestUploadFileIncludesProjectAndBodyID(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "roads.geojson")
	if err := os.WriteFile(filePath, []byte(`{"type":"FeatureCollection","features":[]}`), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	var gotProjectHeader string
	var gotName string
	var gotProjectID string
	var gotBodyID string
	var gotFileName string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotProjectHeader = r.Header.Get("X-Project-ID")
		reader, err := r.MultipartReader()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
			data, _ := io.ReadAll(part)
			switch part.FormName() {
			case "name":
				gotName = string(data)
			case "project_id":
				gotProjectID = string(data)
			case "body_id":
				gotBodyID = string(data)
			case "file":
				gotFileName = part.FileName()
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"roads"}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "")
	client.HTTPClient = srv.Client()
	client.ProjectID = "42"

	body, err := client.UploadFile(filePath, "roads", "42", "mars")
	if err != nil {
		t.Fatalf("UploadFile error: %v", err)
	}
	if gotProjectHeader != "42" {
		t.Fatalf("X-Project-ID = %q, want 42", gotProjectHeader)
	}
	if gotName != "roads" {
		t.Fatalf("name = %q, want roads", gotName)
	}
	if gotProjectID != "42" {
		t.Fatalf("project_id = %q, want 42", gotProjectID)
	}
	if gotBodyID != "mars" {
		t.Fatalf("body_id = %q, want mars", gotBodyID)
	}
	if !strings.HasSuffix(gotFileName, "roads.geojson") {
		t.Fatalf("file name = %q, want roads.geojson", gotFileName)
	}
	if string(body) != `{"name":"roads"}` {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestUploadFileSendsMultipartContentType(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "roads.geojson")
	if err := os.WriteFile(filePath, []byte(`{"type":"FeatureCollection","features":[]}`), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	var contentType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType = r.Header.Get("Content-Type")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "")
	client.HTTPClient = srv.Client()

	if _, err := client.UploadFile(filePath, "", "", ""); err != nil {
		t.Fatalf("UploadFile error: %v", err)
	}
	if !strings.HasPrefix(contentType, "multipart/form-data;") {
		t.Fatalf("content type = %q, want multipart form-data", contentType)
	}
}
