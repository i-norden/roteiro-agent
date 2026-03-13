package mcp

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
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
	case "run_raster_process":
		return handleRunRasterProcess(client, params)
	case "preflight_process":
		return handlePreflightProcess(client, params)
	case "submit_process_job":
		return handleSubmitProcessJob(client, params)
	case "submit_process_batch":
		return handleSubmitProcessBatch(client, params)
	case "list_process_jobs":
		return handleListProcessJobs(client, params)
	case "get_process_job":
		return handleGetProcessJob(client, params)
	case "cancel_process_job":
		return handleCancelProcessJob(client, params)
	case "rerun_process_job":
		return handleRerunProcessJob(client, params)
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
	case "get_duckdb_info":
		return handleGetDuckDBInfo(client)
	case "list_duckdb_datasets":
		return handleListDuckDBDatasets(client)
	case "geocode":
		return handleGeocode(client, params)
	case "reverse_geocode":
		return handleReverseGeocode(client, params)
	case "compute_route":
		return handleComputeRoute(client, params)
	case "compute_isochrone":
		return handleComputeIsochrone(client, params)
	case "compute_route_matrix":
		return handleComputeRouteMatrix(client, params)
	case "compute_service_area":
		return handleComputeServiceArea(client, params)
	case "list_operations":
		return handleListOperations(client)
	case "list_analysis_operations":
		return handleListAnalysisOperations(client)
	case "browse_catalog":
		return handleBrowseCatalog(client, params)
	case "browse_catalog_enhanced":
		return handleBrowseEnhancedCatalog(client, params)
	case "get_catalog_entry":
		return handleGetCatalogEntry(client, params)
	case "list_catalog_categories":
		return handleListCatalogCategories(client)
	case "list_catalog_tags":
		return handleListCatalogTags(client, params)
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
	case "indoor_api":
		return handleIndoorAPI(client, params)
	case "map_api":
		return handleMapAPI(client, params)
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

type apiOperation struct {
	Method   string
	Path     string
	Mutating bool
}

var indoorOperations = map[string]apiOperation{
	"list_campuses":            {Method: "GET", Path: "/api/indoor/campuses"},
	"create_campus":            {Method: "POST", Path: "/api/indoor/campuses", Mutating: true},
	"get_campus":               {Method: "GET", Path: "/api/indoor/campuses/{campus_id}"},
	"update_campus":            {Method: "PUT", Path: "/api/indoor/campuses/{campus_id}", Mutating: true},
	"delete_campus":            {Method: "DELETE", Path: "/api/indoor/campuses/{campus_id}", Mutating: true},
	"list_campus_buildings":    {Method: "GET", Path: "/api/indoor/campuses/{campus_id}/buildings"},
	"list_buildings":           {Method: "GET", Path: "/api/indoor/buildings"},
	"create_building":          {Method: "POST", Path: "/api/indoor/buildings", Mutating: true},
	"get_building":             {Method: "GET", Path: "/api/indoor/buildings/{building_id}"},
	"update_building":          {Method: "PUT", Path: "/api/indoor/buildings/{building_id}", Mutating: true},
	"delete_building":          {Method: "DELETE", Path: "/api/indoor/buildings/{building_id}", Mutating: true},
	"list_floors":              {Method: "GET", Path: "/api/indoor/buildings/{building_id}/floors"},
	"create_floor":             {Method: "POST", Path: "/api/indoor/buildings/{building_id}/floors", Mutating: true},
	"upload_floor_plan":        {Method: "POST", Path: "/api/indoor/buildings/{building_id}/floors/{floor_id}/plan", Mutating: true},
	"list_floor_spaces":        {Method: "GET", Path: "/api/indoor/buildings/{building_id}/floors/{level}/spaces"},
	"create_space":             {Method: "POST", Path: "/api/indoor/buildings/{building_id}/spaces", Mutating: true},
	"get_space":                {Method: "GET", Path: "/api/indoor/buildings/{building_id}/spaces/{space_id}"},
	"update_space":             {Method: "PUT", Path: "/api/indoor/buildings/{building_id}/spaces/{space_id}", Mutating: true},
	"delete_space":             {Method: "DELETE", Path: "/api/indoor/buildings/{building_id}/spaces/{space_id}", Mutating: true},
	"list_transitions":         {Method: "GET", Path: "/api/indoor/buildings/{building_id}/transitions"},
	"list_boundaries":          {Method: "GET", Path: "/api/indoor/buildings/{building_id}/boundaries"},
	"create_boundary":          {Method: "POST", Path: "/api/indoor/buildings/{building_id}/boundaries", Mutating: true},
	"get_boundary":             {Method: "GET", Path: "/api/indoor/buildings/{building_id}/boundaries/{boundary_id}"},
	"update_boundary":          {Method: "PUT", Path: "/api/indoor/buildings/{building_id}/boundaries/{boundary_id}", Mutating: true},
	"delete_boundary":          {Method: "DELETE", Path: "/api/indoor/buildings/{building_id}/boundaries/{boundary_id}", Mutating: true},
	"list_assets":              {Method: "GET", Path: "/api/indoor/buildings/{building_id}/assets"},
	"create_asset":             {Method: "POST", Path: "/api/indoor/buildings/{building_id}/assets", Mutating: true},
	"update_asset_position":    {Method: "PUT", Path: "/api/indoor/buildings/{building_id}/assets/{asset_id}", Mutating: true},
	"navigate":                 {Method: "POST", Path: "/api/indoor/navigate"},
	"find_nearest":             {Method: "POST", Path: "/api/indoor/navigate/nearest"},
	"navigate_outdoor":         {Method: "POST", Path: "/api/indoor/navigate/outdoor"},
	"import_building":          {Method: "POST", Path: "/api/indoor/import", Mutating: true},
	"export_imdf":              {Method: "GET", Path: "/api/indoor/buildings/{building_id}/export/imdf"},
	"export_indoorgml":         {Method: "GET", Path: "/api/indoor/buildings/{building_id}/export/indoorgml"},
	"get_occupancy":            {Method: "GET", Path: "/api/indoor/buildings/{building_id}/occupancy"},
	"get_stats":                {Method: "GET", Path: "/api/indoor/buildings/{building_id}/stats"},
	"validate_building":        {Method: "GET", Path: "/api/indoor/buildings/{building_id}/validate"},
	"get_accessibility":        {Method: "GET", Path: "/api/indoor/buildings/{building_id}/analysis/accessibility"},
	"get_historical_analytics": {Method: "GET", Path: "/api/indoor/buildings/{building_id}/analytics"},
	"ingest_sensor_data":       {Method: "POST", Path: "/api/indoor/buildings/{building_id}/sensors", Mutating: true},
	"get_sensor_data":          {Method: "GET", Path: "/api/indoor/buildings/{building_id}/sensors"},
	"get_sensor_heatmap":       {Method: "GET", Path: "/api/indoor/buildings/{building_id}/sensors/heatmap"},
	"get_sensor_timeseries":    {Method: "GET", Path: "/api/indoor/buildings/{building_id}/sensors/timeseries"},
	"ingest_positions":         {Method: "POST", Path: "/api/indoor/buildings/{building_id}/positions", Mutating: true},
	"get_latest_positions":     {Method: "GET", Path: "/api/indoor/buildings/{building_id}/positions/latest"},
	"get_position_history":     {Method: "GET", Path: "/api/indoor/buildings/{building_id}/positions/{device_id}/history"},
	"create_booking":           {Method: "POST", Path: "/api/indoor/buildings/{building_id}/bookings", Mutating: true},
	"list_bookings":            {Method: "GET", Path: "/api/indoor/buildings/{building_id}/bookings"},
	"cancel_booking":           {Method: "DELETE", Path: "/api/indoor/buildings/{building_id}/bookings/{booking_id}", Mutating: true},
	"checkin_booking":          {Method: "POST", Path: "/api/indoor/buildings/{building_id}/bookings/{booking_id}/checkin", Mutating: true},
	"list_geofences":           {Method: "GET", Path: "/api/indoor/buildings/{building_id}/geofences"},
	"create_geofence":          {Method: "POST", Path: "/api/indoor/buildings/{building_id}/geofences", Mutating: true},
	"get_geofence":             {Method: "GET", Path: "/api/indoor/buildings/{building_id}/geofences/{geofence_id}"},
	"update_geofence":          {Method: "PUT", Path: "/api/indoor/buildings/{building_id}/geofences/{geofence_id}", Mutating: true},
	"delete_geofence":          {Method: "DELETE", Path: "/api/indoor/buildings/{building_id}/geofences/{geofence_id}", Mutating: true},
	"get_evacuation_routes":    {Method: "POST", Path: "/api/indoor/buildings/{building_id}/evacuation"},
	"trigger_evacuation_alert": {Method: "POST", Path: "/api/indoor/buildings/{building_id}/evacuation/alert", Mutating: true},
	"topology_resilience":      {Method: "POST", Path: "/api/indoor/buildings/{building_id}/topology/resilience"},
	"topology_track":           {Method: "POST", Path: "/api/indoor/buildings/{building_id}/topology/track"},
	"topology_compare":         {Method: "POST", Path: "/api/indoor/buildings/{building_id}/topology/compare"},
	"spectral_analysis":        {Method: "POST", Path: "/api/indoor/buildings/{building_id}/spectral"},
	"evacuation_optimal":       {Method: "POST", Path: "/api/indoor/buildings/{building_id}/evacuation/optimal"},
	"simulate_fire_evacuation": {Method: "POST", Path: "/api/indoor/buildings/{building_id}/simulate/fire-evacuation"},
	"simulate_airflow":         {Method: "POST", Path: "/api/indoor/floors/{floor_id}/simulate/airflow"},
	"simulate_acoustics":       {Method: "POST", Path: "/api/indoor/floors/{floor_id}/simulate/acoustics"},
	"simulate_thermal":         {Method: "POST", Path: "/api/indoor/floors/{floor_id}/simulate/thermal"},
	"simulate_rf":              {Method: "POST", Path: "/api/indoor/floors/{floor_id}/simulate/rf"},
	"navigate_manifold":        {Method: "POST", Path: "/api/indoor/floors/{floor_id}/navigate/manifold"},
	"forecast":                 {Method: "POST", Path: "/api/indoor/floors/{floor_id}/forecast"},
	"simulate_infection_risk":  {Method: "POST", Path: "/api/indoor/floors/{floor_id}/simulate/infection-risk"},
}

var mapOperations = map[string]apiOperation{
	"publish_map":             {Method: "POST", Path: "/api/maps/publish", Mutating: true},
	"list_published_maps":     {Method: "GET", Path: "/api/maps/published"},
	"unpublish_map":           {Method: "DELETE", Path: "/api/maps/published/{token}", Mutating: true},
	"get_published_map_stats": {Method: "GET", Path: "/api/maps/published/{token}/stats"},
	"update_map_embed_config": {Method: "PUT", Path: "/api/maps/published/{token}/embed-config", Mutating: true},
	"get_public_map":          {Method: "GET", Path: "/public/maps/{token}"},
	"get_raster_info":         {Method: "GET", Path: "/raster/{name}/info"},
	"get_raster_stats":        {Method: "GET", Path: "/raster/{name}/stats"},
	"get_raster_histogram":    {Method: "GET", Path: "/raster/{name}/histogram"},
	"get_raster_dimensions":   {Method: "GET", Path: "/raster/{name}/dimensions"},
	"get_raster_values":       {Method: "GET", Path: "/raster/{name}/values"},
	"raster_zonal_stats":      {Method: "POST", Path: "/raster/{name}/zonal-stats"},
	"export_raster_band":      {Method: "POST", Path: "/raster/{name}/export"},
	"raster_contour":          {Method: "POST", Path: "/raster/{name}/contour"},
	"raster_viewshed":         {Method: "POST", Path: "/raster/{name}/viewshed"},
	"raster_profile":          {Method: "POST", Path: "/raster/{name}/profile"},
	"raster_kde":              {Method: "POST", Path: "/api/raster/kde"},
	"create_feature":          {Method: "POST", Path: "/collections/{collection_id}/items", Mutating: true},
	"update_feature":          {Method: "PUT", Path: "/collections/{collection_id}/items/{feature_id}", Mutating: true},
	"delete_feature":          {Method: "DELETE", Path: "/collections/{collection_id}/items/{feature_id}", Mutating: true},
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
	for _, key := range []string{"bbox", "filter", "datetime", "limit", "offset", "properties", "sortby"} {
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
	normalizeProcessPayload(params)
	data, err := client.RunProcess(params)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleRunRasterProcess(client *Client, params map[string]interface{}) (string, error) {
	normalizeRasterProcessPayload(params)
	data, err := client.RunRasterProcess(params)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handlePreflightProcess(client *Client, params map[string]interface{}) (string, error) {
	normalizeProcessPayload(params)
	data, err := client.PreflightProcess(params)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleSubmitProcessJob(client *Client, params map[string]interface{}) (string, error) {
	normalizeProcessPayload(params)
	data, err := client.SubmitProcessJob(params)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleSubmitProcessBatch(client *Client, params map[string]interface{}) (string, error) {
	jobs, ok := params["jobs"].([]interface{})
	if !ok || len(jobs) == 0 {
		return "", fmt.Errorf("parameter jobs must be a non-empty array")
	}
	for i, item := range jobs {
		job, ok := item.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("jobs[%d] must be an object", i)
		}
		req, ok := job["request"].(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("jobs[%d].request must be an object", i)
		}
		normalizeProcessPayload(req)
	}
	data, err := client.SubmitProcessBatch(params)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleListProcessJobs(client *Client, params map[string]interface{}) (string, error) {
	qp := map[string]string{}
	for _, key := range []string{"status", "search", "limit", "offset"} {
		if v, ok := params[key]; ok {
			s, err := stringify(v)
			if err == nil && s != "" {
				qp[key] = s
			}
		}
	}
	data, err := client.ListProcessJobs(qp)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleGetProcessJob(client *Client, params map[string]interface{}) (string, error) {
	jobID, err := requireString(params, "job_id")
	if err != nil {
		return "", err
	}
	data, err := client.GetProcessJob(jobID)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleCancelProcessJob(client *Client, params map[string]interface{}) (string, error) {
	jobID, err := requireString(params, "job_id")
	if err != nil {
		return "", err
	}
	data, err := client.CancelProcessJob(jobID)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleRerunProcessJob(client *Client, params map[string]interface{}) (string, error) {
	jobID, err := requireString(params, "job_id")
	if err != nil {
		return "", err
	}
	data, err := client.RerunProcessJob(jobID)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleRunPipeline(client *Client, params map[string]interface{}) (string, error) {
	if _, ok := params["output_name"]; !ok {
		if out, ok := params["output"].(string); ok && out != "" {
			params["output_name"] = out
		}
	}
	if _, ok := params["input"]; !ok {
		if steps, ok := params["steps"].([]interface{}); ok && len(steps) > 0 {
			if step0, ok := steps[0].(map[string]interface{}); ok {
				if in, ok := step0["input"].(string); ok && in != "" {
					params["input"] = in
				}
			}
		}
	}
	data, err := client.RunPipeline(params)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleConvertFormat(client *Client, params map[string]interface{}) (string, error) {
	if _, ok := params["output_format"]; !ok {
		if format, ok := params["format"].(string); ok && format != "" {
			params["output_format"] = format
		}
	}
	if _, ok := params["output_name"]; !ok {
		if out, ok := params["output"].(string); ok && out != "" {
			params["output_name"] = out
		}
	}
	data, err := client.ConvertFormat(params)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func normalizeProcessPayload(params map[string]interface{}) {
	if _, ok := params["output_name"]; !ok {
		if out, ok := params["output"].(string); ok && out != "" {
			params["output_name"] = out
		}
	}
	if _, ok := params["output_format"]; !ok {
		if format, ok := params["format"].(string); ok && format != "" {
			params["output_format"] = format
		}
	}
	if _, ok := params["params"]; !ok || params["params"] == nil {
		params["params"] = map[string]interface{}{}
	}
}

func normalizeRasterProcessPayload(params map[string]interface{}) {
	if _, ok := params["params"]; !ok || params["params"] == nil {
		params["params"] = map[string]interface{}{}
	}
}

func handleDiffDatasets(client *Client, params map[string]interface{}) (string, error) {
	if _, ok := params["left"]; !ok {
		if base, ok := params["base"].(string); ok && base != "" {
			params["left"] = base
		}
	}
	if _, ok := params["right"]; !ok {
		if cmp, ok := params["compare"].(string); ok && cmp != "" {
			params["right"] = cmp
		}
	}
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

func handleGetDuckDBInfo(client *Client) (string, error) {
	data, err := client.GetDuckDBInfo()
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleListDuckDBDatasets(client *Client) (string, error) {
	data, err := client.ListDuckDBDatasets()
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
	lat, err := requireStringLike(params, "lat")
	if err != nil {
		return "", err
	}
	lon, err := requireStringLike(params, "lon")
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
	origin, hasOrigin := params["origin"]
	destination, hasDestination := params["destination"]
	if hasOrigin && hasDestination {
		waypoints := make([][2]float64, 0, 2)
		pt, err := parseRoutePoint(origin)
		if err != nil {
			return "", fmt.Errorf("origin: %w", err)
		}
		waypoints = append(waypoints, pt)

		if raw, ok := params["waypoints"].([]interface{}); ok {
			for i, wp := range raw {
				pt, err := parseRoutePoint(wp)
				if err != nil {
					return "", fmt.Errorf("waypoints[%d]: %w", i, err)
				}
				waypoints = append(waypoints, pt)
			}
		}

		pt, err = parseRoutePoint(destination)
		if err != nil {
			return "", fmt.Errorf("destination: %w", err)
		}
		waypoints = append(waypoints, pt)
		params["waypoints"] = waypoints
	}
	data, err := client.ComputeRoute(params)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleComputeIsochrone(client *Client, params map[string]interface{}) (string, error) {
	if origin, ok := params["origin"]; ok {
		pt, err := parseRoutePoint(origin)
		if err != nil {
			return "", fmt.Errorf("origin: %w", err)
		}
		params["lng"] = pt[0]
		params["lat"] = pt[1]
	}
	data, err := client.ComputeIsochrone(params)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleComputeRouteMatrix(client *Client, params map[string]interface{}) (string, error) {
	if raw, ok := params["origins"].([]interface{}); ok {
		pts, err := parseRoutePoints(raw)
		if err != nil {
			return "", fmt.Errorf("origins: %w", err)
		}
		params["origins"] = pts
	}
	if raw, ok := params["destinations"].([]interface{}); ok {
		pts, err := parseRoutePoints(raw)
		if err != nil {
			return "", fmt.Errorf("destinations: %w", err)
		}
		params["destinations"] = pts
	}
	data, err := client.ComputeRouteMatrix(params)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleComputeServiceArea(client *Client, params map[string]interface{}) (string, error) {
	if origin, ok := params["origin"]; ok {
		pt, err := parseRoutePoint(origin)
		if err != nil {
			return "", fmt.Errorf("origin: %w", err)
		}
		params["lng"] = pt[0]
		params["lat"] = pt[1]
	}
	data, err := client.ComputeServiceArea(params)
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

func handleListAnalysisOperations(client *Client) (string, error) {
	data, err := client.ListAnalysisOperations()
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

// requireStringLike extracts a required parameter as a string, accepting strings and numbers.
func requireStringLike(params map[string]interface{}, key string) (string, error) {
	v, ok := params[key]
	if !ok {
		return "", fmt.Errorf("missing required parameter: %s", key)
	}
	s, err := stringify(v)
	if err != nil {
		return "", fmt.Errorf("parameter %s must be a string or number", key)
	}
	if s == "" {
		return "", fmt.Errorf("parameter %s must not be empty", key)
	}
	return s, nil
}

func parseRoutePoint(v interface{}) ([2]float64, error) {
	if arr, ok := v.([]interface{}); ok {
		if len(arr) != 2 {
			return [2]float64{}, fmt.Errorf("array form must have 2 numbers [lon,lat]")
		}
		lon, err := parseFloat64(arr[0])
		if err != nil {
			return [2]float64{}, fmt.Errorf("invalid lon")
		}
		lat, err := parseFloat64(arr[1])
		if err != nil {
			return [2]float64{}, fmt.Errorf("invalid lat")
		}
		return [2]float64{lon, lat}, nil
	}
	m, ok := v.(map[string]interface{})
	if !ok {
		return [2]float64{}, fmt.Errorf("must be an object with lat/lon")
	}
	latRaw, ok := m["lat"]
	if !ok {
		return [2]float64{}, fmt.Errorf("missing lat")
	}
	lonRaw, ok := m["lon"]
	if !ok {
		return [2]float64{}, fmt.Errorf("missing lon")
	}
	lat, err := parseFloat64(latRaw)
	if err != nil {
		return [2]float64{}, fmt.Errorf("invalid lat")
	}
	lon, err := parseFloat64(lonRaw)
	if err != nil {
		return [2]float64{}, fmt.Errorf("invalid lon")
	}
	return [2]float64{lon, lat}, nil
}

func parseRoutePoints(values []interface{}) ([][2]float64, error) {
	pts := make([][2]float64, 0, len(values))
	for i, v := range values {
		pt, err := parseRoutePoint(v)
		if err != nil {
			return nil, fmt.Errorf("index %d: %w", i, err)
		}
		pts = append(pts, pt)
	}
	return pts, nil
}

func parseFloat64(v interface{}) (float64, error) {
	s, err := stringify(v)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(s, 64)
}

func stringify(v interface{}) (string, error) {
	if s, ok := v.(string); ok {
		return s, nil
	}
	if f, ok := v.(float64); ok {
		return strconv.FormatFloat(f, 'f', -1, 64), nil
	}
	if b, ok := v.(bool); ok {
		if b {
			return "true", nil
		}
		return "false", nil
	}
	return "", fmt.Errorf("unsupported type")
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

func handleBrowseEnhancedCatalog(client *Client, params map[string]interface{}) (string, error) {
	qp := map[string]string{}
	for _, key := range []string{"search", "category", "formats", "tags", "live_only", "sort", "order", "bbox", "limit", "offset"} {
		if v, ok := params[key]; ok {
			s, err := stringify(v)
			if err == nil && s != "" {
				qp[key] = s
			}
		}
	}
	data, err := client.BrowseEnhancedCatalog(qp)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleGetCatalogEntry(client *Client, params map[string]interface{}) (string, error) {
	id, err := requireString(params, "id")
	if err != nil {
		return "", err
	}
	data, err := client.GetCatalogEntry(id)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleListCatalogCategories(client *Client) (string, error) {
	data, err := client.ListCatalogCategories()
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleListCatalogTags(client *Client, params map[string]interface{}) (string, error) {
	limit := ""
	if v, ok := params["limit"]; ok {
		if s, err := stringify(v); err == nil {
			limit = s
		}
	}
	data, err := client.ListCatalogTags(limit)
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

func handleIndoorAPI(client *Client, params map[string]interface{}) (string, error) {
	return handleScopedAPI(client, params, indoorOperations, "indoor")
}

func handleMapAPI(client *Client, params map[string]interface{}) (string, error) {
	return handleScopedAPI(client, params, mapOperations, "map")
}

func handleScopedAPI(client *Client, params map[string]interface{}, ops map[string]apiOperation, scope string) (string, error) {
	opName, err := requireString(params, "operation")
	if err != nil {
		return "", err
	}
	op, ok := ops[opName]
	if !ok {
		return "", fmt.Errorf("unknown %s operation: %s", scope, opName)
	}
	if op.Mutating && !requireConfirm(params) {
		return "", fmt.Errorf("%s operation %q mutates data; pass confirm=true to proceed", scope, opName)
	}

	path, err := interpolatePath(op.Path, params)
	if err != nil {
		return "", err
	}
	query, err := extractQuery(params)
	if err != nil {
		return "", err
	}
	body := extractBody(params)
	if op.Method == "GET" || op.Method == "DELETE" {
		body = nil
	}

	data, err := client.APIRequest(op.Method, path, body, query)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func requireConfirm(params map[string]interface{}) bool {
	v, ok := params["confirm"]
	if !ok {
		return false
	}
	b, ok := v.(bool)
	return ok && b
}

func extractBody(params map[string]interface{}) interface{} {
	if raw, ok := params["body"]; ok {
		if _, ok := raw.(map[string]interface{}); ok {
			return raw
		}
		if _, ok := raw.([]interface{}); ok {
			return raw
		}
	}
	return nil
}

func extractQuery(params map[string]interface{}) (map[string]string, error) {
	raw, ok := params["query"]
	if !ok {
		return nil, nil
	}
	m, ok := raw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("query must be an object of key/value pairs")
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		if k == "" || v == nil {
			continue
		}
		s, err := stringify(v)
		if err != nil {
			return nil, fmt.Errorf("query.%s must be string/number/bool", k)
		}
		out[k] = s
	}
	return out, nil
}

func interpolatePath(tpl string, params map[string]interface{}) (string, error) {
	out := tpl
	for {
		start := strings.IndexByte(out, '{')
		if start == -1 {
			return out, nil
		}
		end := strings.IndexByte(out[start:], '}')
		if end == -1 {
			return "", fmt.Errorf("invalid operation path template: %s", tpl)
		}
		end += start
		key := out[start+1 : end]
		val, ok := params[key]
		if !ok {
			return "", fmt.Errorf("missing required path parameter: %s", key)
		}
		s, err := stringify(val)
		if err != nil || strings.TrimSpace(s) == "" {
			return "", fmt.Errorf("invalid path parameter %s", key)
		}
		out = out[:start] + url.PathEscape(s) + out[end+1:]
	}
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
