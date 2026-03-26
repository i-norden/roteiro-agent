package mcp

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListSpatialTablesPrefersCurrentDuckDBEndpoint(t *testing.T) {
	tablesCalled := false
	datasetsCalled := false

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/query/sql/datasets":
			datasetsCalled = true
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`[{"name":"roads"}]`))
		case "/api/sql/tables":
			tablesCalled = true
			http.Error(w, `{"error":"legacy route should not be called first"}`, http.StatusGone)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "")
	client.HTTPClient = srv.Client()

	body, err := client.ListSpatialTables()
	if err != nil {
		t.Fatalf("ListSpatialTables error: %v", err)
	}
	if !datasetsCalled {
		t.Fatal("expected current /api/query/sql/datasets route to be used")
	}
	if tablesCalled {
		t.Fatal("legacy /api/sql/tables route should not be called when the current endpoint succeeds")
	}
	if string(body) != `[{"name":"roads"}]` {
		t.Fatalf("unexpected body: %s", body)
	}
}
