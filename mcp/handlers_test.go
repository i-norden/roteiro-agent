package mcp

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNormalizeImportSourcePayload(t *testing.T) {
	payload, err := normalizeImportSourcePayload(map[string]interface{}{
		"name":        "mola-dem",
		"source":      "https://example.test/mola.tif",
		"format":      "tif",
		"body_id":     "mars",
		"source_type": "remote_url",
	}, "42")
	if err != nil {
		t.Fatalf("normalizeImportSourcePayload error: %v", err)
	}

	if got := payload["name"]; got != "mola-dem" {
		t.Fatalf("name = %v, want mola-dem", got)
	}
	if got := payload["source"]; got != "https://example.test/mola.tif" {
		t.Fatalf("source = %v, want remote URL", got)
	}
	if got := payload["body_id"]; got != "mars" {
		t.Fatalf("body_id = %v, want mars", got)
	}
	if got := payload["project_id"]; got != int64(42) {
		t.Fatalf("project_id = %v, want 42", got)
	}
}

func TestNormalizeOperationPayloadMapsCompatibilityAliases(t *testing.T) {
	payload, operation, err := normalizeOperationPayload(map[string]interface{}{
		"operation": "buffer",
		"input":     "roads",
		"output":    "roads_buffered",
		"format":    "parquet",
	}, "7")
	if err != nil {
		t.Fatalf("normalizeOperationPayload error: %v", err)
	}
	if operation != "buffer" {
		t.Fatalf("operation = %q, want buffer", operation)
	}
	if got := payload["output_name"]; got != "roads_buffered" {
		t.Fatalf("output_name = %v, want roads_buffered", got)
	}
	if got := payload["output_format"]; got != "parquet" {
		t.Fatalf("output_format = %v, want parquet", got)
	}
	if _, ok := payload["params"].(map[string]interface{}); !ok {
		t.Fatalf("params = %#v, want initialized params object", payload["params"])
	}
	if got := payload["project_id"]; got != int64(7) {
		t.Fatalf("project_id = %v, want 7", got)
	}
}

func TestNormalizeSQLPayloads(t *testing.T) {
	execPayload, err := normalizeSQLExecutePayload(map[string]interface{}{
		"query":       "SELECT 1",
		"limit":       5.0,
		"timeout_sec": 10.0,
	})
	if err != nil {
		t.Fatalf("normalizeSQLExecutePayload error: %v", err)
	}
	if got := execPayload["sql"]; got != "SELECT 1" {
		t.Fatalf("sql = %v, want SELECT 1", got)
	}

	savePayload, err := normalizeSQLSavePayload(map[string]interface{}{
		"sql":             "SELECT * FROM roads",
		"output_name":     "roads_export",
		"geometry_column": "geom",
	})
	if err != nil {
		t.Fatalf("normalizeSQLSavePayload error: %v", err)
	}
	if got := savePayload["output_name"]; got != "roads_export" {
		t.Fatalf("output_name = %v, want roads_export", got)
	}
	if got := savePayload["geometry_column"]; got != "geom" {
		t.Fatalf("geometry_column = %v, want geom", got)
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

func TestHandleSubmitOperationBatchValidation(t *testing.T) {
	client := NewClient("https://example.test", "")

	_, err := handleSubmitOperationBatch(client, map[string]interface{}{})
	if err == nil || !strings.Contains(err.Error(), "missing required parameter: jobs") {
		t.Fatalf("error = %v, want missing jobs error", err)
	}

	_, err = handleSubmitOperationBatch(client, map[string]interface{}{
		"jobs": []interface{}{"bad"},
	})
	if err == nil || !strings.Contains(err.Error(), "jobs[0] must be an object") {
		t.Fatalf("error = %v, want job object error", err)
	}

	_, err = handleSubmitOperationBatch(client, map[string]interface{}{
		"jobs": []interface{}{
			map[string]interface{}{"request": "bad"},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "jobs[0]: parameter request must be an object") {
		t.Fatalf("error = %v, want request object error", err)
	}
}
