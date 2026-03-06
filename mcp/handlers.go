package mcp

import (
	"encoding/json"
	"fmt"
)

// HandleToolCall dispatches a tool call to the appropriate handler and returns
// the result as text content for the MCP response.
func HandleToolCall(client *Client, name string, args json.RawMessage) (string, error) {
	var params map[string]interface{}
	if len(args) > 0 {
		if err := json.Unmarshal(args, &params); err != nil {
			return "", fmt.Errorf("invalid arguments: %w", err)
		}
	}
	if params == nil {
		params = map[string]interface{}{}
	}

	switch name {
	case "list_datasets":
		return handleListDatasets(client)
	case "get_dataset_info":
		return handleGetDatasetInfo(client, params)
	case "get_dataset_schema":
		return handleGetDatasetSchema(client, params)
	case "get_dataset_profile":
		return handleGetDatasetProfile(client, params)
	case "query_features":
		return handleQueryFeatures(client, params)
	case "get_feature":
		return handleGetFeature(client, params)
	case "upload_dataset":
		return handleUploadDataset(client, params)
	case "run_process":
		return handleRunProcess(client, params)
	case "run_pipeline":
		return handleRunPipeline(client, params)
	case "convert_format":
		return handleConvertFormat(client, params)
	case "diff_datasets":
		return handleDiffDatasets(client, params)
	case "execute_sql":
		return handleExecuteSQL(client, params)
	case "list_spatial_tables":
		return handleListSpatialTables(client)
	case "geocode":
		return handleGeocode(client, params)
	case "reverse_geocode":
		return handleReverseGeocode(client, params)
	case "compute_route":
		return handleComputeRoute(client, params)
	case "list_operations":
		return handleListOperations(client)
	case "browse_catalog":
		return handleBrowseCatalog(client, params)
	case "import_from_catalog":
		return handleImportFromCatalog(client, params)
	case "browse_stac_catalog":
		return handleBrowseSTACCatalog(client, params)
	case "browse_stac_collections":
		return handleBrowseSTACCollections(client, params)
	case "browse_stac_items":
		return handleBrowseSTACItems(client, params)
	case "import_stac_asset":
		return handleImportSTACAsset(client, params)
	case "search_stac":
		return handleSearchSTAC(client, params)
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

func handleListDatasets(client *Client) (string, error) {
	data, err := client.ListDatasets()
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleGetDatasetInfo(client *Client, params map[string]interface{}) (string, error) {
	id, err := requireString(params, "collection_id")
	if err != nil {
		return "", err
	}
	data, err := client.GetCollection(id)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleGetDatasetSchema(client *Client, params map[string]interface{}) (string, error) {
	name, err := requireString(params, "name")
	if err != nil {
		return "", err
	}
	data, err := client.GetDatasetSchema(name)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleGetDatasetProfile(client *Client, params map[string]interface{}) (string, error) {
	name, err := requireString(params, "name")
	if err != nil {
		return "", err
	}
	data, err := client.GetDatasetProfile(name)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleQueryFeatures(client *Client, params map[string]interface{}) (string, error) {
	id, err := requireString(params, "collection_id")
	if err != nil {
		return "", err
	}
	qp := map[string]string{}
	for _, key := range []string{"bbox", "filter", "limit", "offset", "properties", "sortby"} {
		if v, ok := params[key].(string); ok && v != "" {
			qp[key] = v
		}
	}
	// Default to a reasonable limit to avoid dumping huge responses.
	if _, ok := qp["limit"]; !ok {
		qp["limit"] = "10"
	}
	data, err := client.QueryFeatures(id, qp)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleGetFeature(client *Client, params map[string]interface{}) (string, error) {
	collID, err := requireString(params, "collection_id")
	if err != nil {
		return "", err
	}
	fid, err := requireString(params, "feature_id")
	if err != nil {
		return "", err
	}
	data, err := client.GetFeature(collID, fid)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleUploadDataset(client *Client, params map[string]interface{}) (string, error) {
	path, err := requireString(params, "file_path")
	if err != nil {
		return "", err
	}
	data, err := client.UploadFile(path)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleRunProcess(client *Client, params map[string]interface{}) (string, error) {
	data, err := client.RunProcess(params)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleRunPipeline(client *Client, params map[string]interface{}) (string, error) {
	data, err := client.RunPipeline(params)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleConvertFormat(client *Client, params map[string]interface{}) (string, error) {
	data, err := client.ConvertFormat(params)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleDiffDatasets(client *Client, params map[string]interface{}) (string, error) {
	data, err := client.DiffDatasets(params)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleExecuteSQL(client *Client, params map[string]interface{}) (string, error) {
	query, err := requireString(params, "query")
	if err != nil {
		return "", err
	}
	data, err := client.ExecuteSQL(query)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleListSpatialTables(client *Client) (string, error) {
	data, err := client.ListSpatialTables()
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleGeocode(client *Client, params map[string]interface{}) (string, error) {
	addr, err := requireString(params, "address")
	if err != nil {
		return "", err
	}
	data, err := client.Geocode(addr)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleReverseGeocode(client *Client, params map[string]interface{}) (string, error) {
	lat, err := requireString(params, "lat")
	if err != nil {
		return "", err
	}
	lon, err := requireString(params, "lon")
	if err != nil {
		return "", err
	}
	data, err := client.ReverseGeocode(lat, lon)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleComputeRoute(client *Client, params map[string]interface{}) (string, error) {
	data, err := client.ComputeRoute(params)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleListOperations(client *Client) (string, error) {
	data, err := client.ListOperations()
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

// requireString extracts a required string parameter.
func requireString(params map[string]interface{}, key string) (string, error) {
	v, ok := params[key]
	if !ok {
		return "", fmt.Errorf("missing required parameter: %s", key)
	}
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("parameter %s must be a string", key)
	}
	if s == "" {
		return "", fmt.Errorf("parameter %s must not be empty", key)
	}
	return s, nil
}

func handleBrowseCatalog(client *Client, params map[string]interface{}) (string, error) {
	qp := map[string]string{}
	for _, key := range []string{"search", "category", "limit", "offset"} {
		if v, ok := params[key].(string); ok && v != "" {
			qp[key] = v
		}
	}
	data, err := client.BrowseCatalog(qp)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleImportFromCatalog(client *Client, params map[string]interface{}) (string, error) {
	catalogID, err := requireString(params, "catalog_id")
	if err != nil {
		return "", err
	}
	data, err := client.ImportFromCatalog(catalogID)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleBrowseSTACCatalog(client *Client, params map[string]interface{}) (string, error) {
	catalogURL, err := requireString(params, "url")
	if err != nil {
		return "", err
	}
	data, err := client.BrowseSTACCatalog(catalogURL)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleBrowseSTACCollections(client *Client, params map[string]interface{}) (string, error) {
	catalogURL, err := requireString(params, "url")
	if err != nil {
		return "", err
	}
	data, err := client.BrowseSTACCollections(catalogURL)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleBrowseSTACItems(client *Client, params map[string]interface{}) (string, error) {
	collURL, err := requireString(params, "url")
	if err != nil {
		return "", err
	}
	qp := map[string]string{}
	for _, key := range []string{"bbox", "datetime"} {
		if v, ok := params[key].(string); ok && v != "" {
			qp[key] = v
		}
	}
	data, err := client.BrowseSTACItems(collURL, qp)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleImportSTACAsset(client *Client, params map[string]interface{}) (string, error) {
	assetURL, err := requireString(params, "asset_url")
	if err != nil {
		return "", err
	}
	name, err := requireString(params, "name")
	if err != nil {
		return "", err
	}
	format, _ := params["format"].(string)
	data, err := client.ImportSTACAsset(assetURL, name, format)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleSearchSTAC(client *Client, params map[string]interface{}) (string, error) {
	qp := map[string]string{}
	for _, key := range []string{"bbox", "datetime", "collections", "limit", "filter"} {
		if v, ok := params[key].(string); ok && v != "" {
			qp[key] = v
		}
	}
	data, err := client.SearchSTAC(qp)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

// formatJSON pretty-prints JSON for readability in agent responses.
func formatJSON(data json.RawMessage) string {
	var out json.RawMessage
	if err := json.Unmarshal(data, &out); err != nil {
		return string(data)
	}
	pretty, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return string(data)
	}
	return string(pretty)
}
