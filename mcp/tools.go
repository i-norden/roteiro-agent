package mcp

// Tool describes an MCP tool with its name, description, and JSON Schema for inputs.
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

// InputSchema is the JSON Schema for a tool's input parameters.
type InputSchema struct {
	Type       string                    `json:"type"`
	Properties map[string]PropertySchema `json:"properties,omitempty"`
	Required   []string                  `json:"required,omitempty"`
}

// PropertySchema describes a single property in the input schema.
type PropertySchema struct {
	Type        string                    `json:"type"`
	Description string                    `json:"description,omitempty"`
	Items       *PropertySchema           `json:"items,omitempty"`
	Properties  map[string]PropertySchema `json:"properties,omitempty"`
	Enum        []string                  `json:"enum,omitempty"`
	Default     interface{}               `json:"default,omitempty"`
}

// AllTools returns the complete list of MCP tools that this server exposes.
func AllTools() []Tool {
	return []Tool{
		{
			Name:        "list_datasets",
			Description: "List all datasets registered in Roteiro with their names, formats, feature counts, and geometry types. Optionally scope the listing to a project.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"project_id": {Type: "string", Description: "Optional project scope override. Also configurable globally via --project-id."},
				},
			},
		},
		{
			Name:        "get_dataset_info",
			Description: "Get detailed information about a dataset including its schema (field names and types), CRS, bounds, feature count, and geometry type.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"collection_id": {Type: "string", Description: "The dataset/collection identifier."},
					"project_id":    {Type: "string", Description: "Optional project scope override. Also configurable globally via --project-id."},
				},
				Required: []string{"collection_id"},
			},
		},
		{
			Name:        "get_dataset_schema",
			Description: "Get the field schema (column names, types) for a dataset. Useful for understanding what attributes are available before querying.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"name": {Type: "string", Description: "The dataset name."},
				},
				Required: []string{"name"},
			},
		},
		{
			Name:        "get_dataset_profile",
			Description: "Get a statistical profile of a dataset including value distributions, min/max, and null counts for each field.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"name": {Type: "string", Description: "The dataset name."},
				},
				Required: []string{"name"},
			},
		},
		{
			Name:        "query_features",
			Description: "Query features from a collection with optional spatial/attribute filters. Returns GeoJSON FeatureCollection. Use bbox for spatial filtering and filter for CQL2 attribute filtering.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"collection_id": {Type: "string", Description: "The collection identifier."},
					"bbox":          {Type: "string", Description: "Bounding box filter as 'west,south,east,north' (EPSG:4326)."},
					"bbox_crs":      {Type: "string", Description: "Optional CRS identifier for the bbox coordinates, forwarded as `bbox-crs`."},
					"crs":           {Type: "string", Description: "Optional CRS identifier for returned geometries."},
					"filter":        {Type: "string", Description: "CQL2 filter expression (e.g. \"population > 10000\")."},
					"datetime":      {Type: "string", Description: "Temporal filter as RFC3339 instant or interval 'start/end'."},
					"limit":         {Type: "string", Description: "Maximum number of features to return (default 10)."},
					"offset":        {Type: "string", Description: "Number of features to skip for pagination."},
					"properties":    {Type: "string", Description: "Comma-separated list of properties to include in the response."},
					"sortby":        {Type: "string", Description: "Property to sort by, prefix with '-' for descending."},
					"project_id":    {Type: "string", Description: "Optional project scope override. Also configurable globally via --project-id."},
				},
				Required: []string{"collection_id"},
			},
		},
		{
			Name:        "get_feature",
			Description: "Get a single feature by its ID from a collection. Returns a GeoJSON Feature with all properties and geometry.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"collection_id": {Type: "string", Description: "The collection identifier."},
					"feature_id":    {Type: "string", Description: "The feature identifier."},
				},
				Required: []string{"collection_id", "feature_id"},
			},
		},
		{
			Name:        "upload_dataset",
			Description: "Upload a spatial data file (GeoJSON, Shapefile, GeoPackage, KML, CSV, etc.) to register it as a new dataset.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"file_path":  {Type: "string", Description: "Local file path to the spatial data file to upload."},
					"name":       {Type: "string", Description: "Optional dataset name. Defaults to the file stem if omitted."},
					"project_id": {Type: "string", Description: "Optional project to attach the uploaded dataset to. Also configurable globally via --project-id."},
				},
				Required: []string{"file_path"},
			},
		},
		{
			Name:        "run_process",
			Description: "Run a single geoprocessing operation on a dataset via /api/process. Use list_operations first to discover the live operation catalog and parameter names on the connected server.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"operation": {Type: "string", Description: "The geoprocessing operation to run (e.g. 'buffer', 'clip', 'simplify')."},
					"input":     {Type: "string", Description: "Input dataset name. Provide either 'input' or 'input_geojson'."},
					"input_geojson": {
						Type:        "object",
						Description: "Inline GeoJSON input. Provide either 'input' or 'input_geojson'.",
					},
					"params": {
						Type:        "object",
						Description: "Operation-specific parameters (e.g. {\"distance\": 500} for buffer, {\"tolerance\": 0.001} for simplify).",
					},
					"output":        {Type: "string", Description: "Compatibility alias for 'output_name'."},
					"output_name":   {Type: "string", Description: "Output dataset name when registering or naming results."},
					"output_format": {Type: "string", Description: "Requested output format (for example 'geojson', 'parquet', or 'csv')."},
					"format":        {Type: "string", Description: "Compatibility alias for 'output_format'."},
					"register":      {Type: "boolean", Description: "Whether to register the result as a dataset."},
					"project_id":    {Type: "string", Description: "Optional project scope override and output attachment target."},
				},
				Required: []string{"operation"},
			},
		},
		{
			Name:        "run_raster_process",
			Description: "Run a generic raster processing operation via /api/raster/process using file paths. Current backend operations include terrain (slope, aspect, profile_curvature, plan_curvature, general_curvature), hydrology (fill, flow_direction, flow_accumulation, watershed, stream_order, snap_pour_point, basin_labels), distance/cost tools, spectral/change tools, classification, and raster-vector conversion.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"operation":   {Type: "string", Description: "Raster operation name such as 'slope', 'fill', 'cost_distance', 'spectral_index', 'kmeans', 'raster_to_polygons', or 'rasterize'."},
					"input_path":  {Type: "string", Description: "Input raster file path. Retrieve dataset paths with list_datasets or get_dataset_info when needed."},
					"output_path": {Type: "string", Description: "Optional output GeoTIFF path for file-writing operations."},
					"band":        {Type: "string", Description: "Optional raster band index."},
					"params":      {Type: "object", Description: "Operation-specific parameters, matching the backend raster analysis docs."},
				},
				Required: []string{"operation"},
			},
		},
		{
			Name:        "preflight_process",
			Description: "Validate and normalize a processing request via /api/process/preflight before executing it. Useful for checking required params, CRS constraints, and normalized dataset references.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"operation":     {Type: "string", Description: "The geoprocessing operation to validate."},
					"input":         {Type: "string", Description: "Input dataset name. Provide either 'input' or 'input_geojson'."},
					"input_geojson": {Type: "object", Description: "Inline GeoJSON input. Provide either 'input' or 'input_geojson'."},
					"params":        {Type: "object", Description: "Operation-specific parameters."},
					"output":        {Type: "string", Description: "Compatibility alias for 'output_name'."},
					"output_name":   {Type: "string", Description: "Optional output dataset name."},
					"output_format": {Type: "string", Description: "Requested output format."},
					"format":        {Type: "string", Description: "Compatibility alias for 'output_format'."},
					"register":      {Type: "boolean", Description: "Whether the eventual result should be registered as a dataset."},
					"project_id":    {Type: "string", Description: "Optional project scope override and output attachment target."},
				},
				Required: []string{"operation"},
			},
		},
		{
			Name:        "submit_process_job",
			Description: "Submit an asynchronous processing job via /api/process/jobs. Use this when the operation may take longer, or when preflight recommends async execution.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"operation":     {Type: "string", Description: "The geoprocessing operation to queue."},
					"input":         {Type: "string", Description: "Input dataset name. Provide either 'input' or 'input_geojson'."},
					"input_geojson": {Type: "object", Description: "Inline GeoJSON input. Provide either 'input' or 'input_geojson'."},
					"params":        {Type: "object", Description: "Operation-specific parameters."},
					"output":        {Type: "string", Description: "Compatibility alias for 'output_name'."},
					"output_name":   {Type: "string", Description: "Optional output dataset name."},
					"output_format": {Type: "string", Description: "Requested output format."},
					"format":        {Type: "string", Description: "Compatibility alias for 'output_format'."},
					"register":      {Type: "boolean", Description: "Whether to register the result as a dataset."},
					"project_id":    {Type: "string", Description: "Optional project scope override and output attachment target."},
				},
				Required: []string{"operation"},
			},
		},
		{
			Name:        "submit_process_batch",
			Description: "Submit a dependent batch of asynchronous processing jobs via /api/process/jobs/batch.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"jobs": {
						Type:        "array",
						Description: "Array of batch jobs. Each item supports client_id, depends_on, and request fields matching submit_process_job. Default project scope is inherited unless a request already sets project_id.",
						Items: &PropertySchema{
							Type: "object",
							Properties: map[string]PropertySchema{
								"client_id":  {Type: "string", Description: "Optional client-side identifier used for dependency references."},
								"depends_on": {Type: "array", Description: "Optional list of batch client IDs or job IDs this job depends on.", Items: &PropertySchema{Type: "string"}},
								"request":    {Type: "object", Description: "Process request payload matching submit_process_job."},
							},
						},
					},
				},
				Required: []string{"jobs"},
			},
		},
		{
			Name:        "list_process_jobs",
			Description: "List asynchronous processing jobs via /api/process/jobs with optional filtering.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"status": {Type: "string", Description: "Optional job status filter: queued, processing, completed, failed, cancelled."},
					"search": {Type: "string", Description: "Optional substring match against operation or job metadata."},
					"limit":  {Type: "string", Description: "Optional max jobs to return."},
					"offset": {Type: "string", Description: "Optional pagination offset."},
				},
			},
		},
		{
			Name:        "get_process_job",
			Description: "Fetch a single asynchronous processing job by ID.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"job_id": {Type: "string", Description: "The async processing job ID."},
				},
				Required: []string{"job_id"},
			},
		},
		{
			Name:        "cancel_process_job",
			Description: "Cancel an asynchronous processing job by ID.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"job_id": {Type: "string", Description: "The async processing job ID."},
				},
				Required: []string{"job_id"},
			},
		},
		{
			Name:        "rerun_process_job",
			Description: "Re-submit a previous asynchronous processing job by ID.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"job_id": {Type: "string", Description: "The async processing job ID."},
				},
				Required: []string{"job_id"},
			},
		},
		{
			Name:        "run_pipeline",
			Description: "Run a multi-step geoprocessing pipeline. Each step's output feeds into the next step. Useful for chaining operations like buffer → clip → simplify.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"input": {Type: "string", Description: "Input dataset name."},
					"steps": {
						Type:        "array",
						Description: "Array of pipeline steps, each with 'operation', 'input' (or uses previous output), and 'params'.",
						Items: &PropertySchema{
							Type: "object",
							Properties: map[string]PropertySchema{
								"operation": {Type: "string", Description: "The geoprocessing operation."},
								"input":     {Type: "string", Description: "Input dataset (optional if chaining from previous step)."},
								"params":    {Type: "object", Description: "Operation-specific parameters."},
								"output":    {Type: "string", Description: "Output dataset name (optional)."},
							},
						},
					},
					"output":     {Type: "string", Description: "Output dataset name when registering results (optional)."},
					"register":   {Type: "boolean", Description: "Persist the pipeline result as a dataset."},
					"project_id": {Type: "string", Description: "Optional project scope override. Also used when the pipeline registers a dataset."},
				},
				Required: []string{"input", "steps"},
			},
		},
		{
			Name:        "convert_format",
			Description: "Convert a dataset between spatial data formats. Supported formats: geojson, shapefile, geopackage, kml, csv, flatgeobuf, parquet.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"input":      {Type: "string", Description: "Input dataset name."},
					"format":     {Type: "string", Description: "Target format (e.g. 'geojson', 'shapefile', 'geopackage', 'kml', 'csv', 'flatgeobuf', 'parquet')."},
					"output":     {Type: "string", Description: "Optional output dataset name when registering results."},
					"register":   {Type: "boolean", Description: "Persist converted output as a dataset."},
					"project_id": {Type: "string", Description: "Optional project to attach the converted dataset to."},
				},
				Required: []string{"input", "format"},
			},
		},
		{
			Name:        "diff_datasets",
			Description: "Compare two dataset versions and show added, removed, and modified features.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"base":        {Type: "string", Description: "Base dataset name (the 'before' version)."},
					"compare":     {Type: "string", Description: "Compare dataset name (the 'after' version)."},
					"match_field": {Type: "string", Description: "Optional stable feature ID field for matching rows."},
				},
				Required: []string{"base", "compare"},
			},
		},
		{
			Name:        "execute_sql",
			Description: "Execute a read-only PostGIS SQL query against the spatial database. Supports all PostGIS spatial functions (ST_Area, ST_Buffer, ST_Intersects, etc.). Results are returned as JSON rows.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"query": {Type: "string", Description: "The SQL query to execute. Must be a SELECT statement (read-only)."},
				},
				Required: []string{"query"},
			},
		},
		{
			Name:        "list_spatial_tables",
			Description: "List all spatial tables in the PostGIS database with their schemas, geometry columns, SRIDs, and geometry types.",
			InputSchema: InputSchema{Type: "object"},
		},
		{
			Name:        "get_duckdb_info",
			Description: "Get DuckDB SQL engine status, capabilities, supported functions, and safety limits.",
			InputSchema: InputSchema{Type: "object"},
		},
		{
			Name:        "list_duckdb_datasets",
			Description: "List datasets available to the DuckDB SQL endpoint (/api/query/sql/datasets).",
			InputSchema: InputSchema{Type: "object"},
		},
		{
			Name:        "geocode",
			Description: "Forward geocode: convert an address or place name to geographic coordinates (latitude/longitude).",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"address": {Type: "string", Description: "The address or place name to geocode."},
				},
				Required: []string{"address"},
			},
		},
		{
			Name:        "reverse_geocode",
			Description: "Reverse geocode: convert geographic coordinates to an address or place name.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"lat": {Type: "number", Description: "Latitude."},
					"lon": {Type: "number", Description: "Longitude."},
				},
				Required: []string{"lat", "lon"},
			},
		},
		{
			Name:        "compute_route",
			Description: "Compute a driving/walking route between two or more points. Returns the route geometry, distance, and duration.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"origin":      {Type: "object", Description: "Origin point as {\"lat\": number, \"lon\": number}."},
					"destination": {Type: "object", Description: "Destination point as {\"lat\": number, \"lon\": number}."},
					"waypoints": {
						Type:        "array",
						Description: "Optional intermediate waypoints as [{\"lat\": number, \"lon\": number}, ...].",
						Items:       &PropertySchema{Type: "object"},
					},
					"profile": {Type: "string", Description: "Routing profile: 'driving', 'walking', 'cycling' (default: 'driving')."},
				},
				Required: []string{"origin", "destination"},
			},
		},
		{
			Name:        "compute_isochrone",
			Description: "Compute travel-time isochrone polygons from an origin point.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"origin": {Type: "object", Description: "Origin point as {\"lat\": number, \"lon\": number}."},
					"minutes": {
						Type:        "array",
						Description: "Travel-time thresholds in minutes (1-120).",
						Items:       &PropertySchema{Type: "number"},
					},
					"profile": {Type: "string", Description: "Routing profile: 'driving', 'walking', 'cycling' (default: 'driving')."},
				},
				Required: []string{"origin", "minutes"},
			},
		},
		{
			Name:        "compute_route_matrix",
			Description: "Compute origin-destination travel-time and distance matrices.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"origins": {
						Type:        "array",
						Description: "Origin points as [{\"lat\": number, \"lon\": number}, ...].",
						Items:       &PropertySchema{Type: "object"},
					},
					"destinations": {
						Type:        "array",
						Description: "Destination points as [{\"lat\": number, \"lon\": number}, ...].",
						Items:       &PropertySchema{Type: "object"},
					},
					"profile": {Type: "string", Description: "Routing profile: 'driving', 'walking', 'cycling' (default: 'driving')."},
				},
				Required: []string{"origins", "destinations"},
			},
		},
		{
			Name:        "compute_service_area",
			Description: "Compute distance-based service area polygons from an origin point.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"origin": {Type: "object", Description: "Origin point as {\"lat\": number, \"lon\": number}."},
					"meters": {
						Type:        "array",
						Description: "Distance thresholds in meters (1-100000).",
						Items:       &PropertySchema{Type: "number"},
					},
					"profile": {Type: "string", Description: "Routing profile: 'driving', 'walking', 'cycling' (default: 'driving')."},
				},
				Required: []string{"origin", "meters"},
			},
		},
		{
			Name:        "list_operations",
			Description: "List all available geoprocessing operations with their parameter schemas. Useful for discovering what operations are supported and what parameters they accept.",
			InputSchema: InputSchema{Type: "object"},
		},
		{
			Name:        "list_analysis_operations",
			Description: "List all available advanced analysis operations from /api/analysis/operations.",
			InputSchema: InputSchema{Type: "object"},
		},
		{
			Name:        "browse_catalog",
			Description: "Browse the built-in data catalog to discover available datasets for import. Supports text search and category filtering.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"search":   {Type: "string", Description: "Text search across dataset names, descriptions, and providers."},
					"category": {Type: "string", Description: "Filter by category (e.g. 'boundaries', 'transportation', 'environment')."},
					"limit":    {Type: "string", Description: "Maximum results to return (default 50, max 500)."},
					"offset":   {Type: "string", Description: "Pagination offset."},
				},
			},
		},
		{
			Name:        "browse_catalog_enhanced",
			Description: "Browse the enhanced catalog with advanced filters (formats, tags, live_only, bbox, sorting).",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"search":    {Type: "string", Description: "Text search query."},
					"category":  {Type: "string", Description: "Category filter."},
					"formats":   {Type: "string", Description: "Comma-separated formats (e.g. 'geojson,parquet')."},
					"tags":      {Type: "string", Description: "Comma-separated tags filter."},
					"live_only": {Type: "boolean", Description: "Only include live/updating datasets."},
					"sort":      {Type: "string", Description: "Sort key: popularity, name, recent, updated."},
					"order":     {Type: "string", Description: "Sort order: asc or desc."},
					"bbox":      {Type: "string", Description: "Bounding box filter as 'minLon,minLat,maxLon,maxLat'."},
					"limit":     {Type: "string", Description: "Maximum results (1-500)."},
					"offset":    {Type: "string", Description: "Pagination offset."},
				},
			},
		},
		{
			Name:        "get_catalog_entry",
			Description: "Get a single enhanced catalog entry by ID.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"id": {Type: "string", Description: "Enhanced catalog entry ID."},
				},
				Required: []string{"id"},
			},
		},
		{
			Name:        "list_catalog_categories",
			Description: "List available catalog categories.",
			InputSchema: InputSchema{Type: "object"},
		},
		{
			Name:        "list_catalog_tags",
			Description: "List catalog tags ordered by frequency.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"limit": {Type: "string", Description: "Maximum tags to return (1-200)."},
				},
			},
		},
		{
			Name:        "import_from_catalog",
			Description: "Import a dataset from the built-in data catalog by its catalog ID. Downloads and registers the dataset for querying and processing.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"catalog_id": {Type: "string", Description: "The catalog entry ID to import."},
					"project_id": {Type: "string", Description: "Optional project to attach the imported dataset to. Also configurable globally via --project-id."},
				},
				Required: []string{"catalog_id"},
			},
		},
		{
			Name:        "browse_stac_catalog",
			Description: "Browse a remote STAC (SpatioTemporal Asset Catalog) catalog by URL. Returns the catalog metadata and links.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"url": {Type: "string", Description: "URL of the remote STAC catalog root (e.g. 'https://planetarycomputer.microsoft.com/api/stac/v1')."},
				},
				Required: []string{"url"},
			},
		},
		{
			Name:        "browse_stac_collections",
			Description: "List collections available in a remote STAC catalog.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"url": {Type: "string", Description: "URL of the remote STAC catalog root."},
				},
				Required: []string{"url"},
			},
		},
		{
			Name:        "browse_stac_items",
			Description: "List items (features) in a remote STAC collection. Supports bbox and datetime filtering.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"url":      {Type: "string", Description: "URL of the remote STAC collection."},
					"bbox":     {Type: "string", Description: "Bounding box filter as 'west,south,east,north'."},
					"datetime": {Type: "string", Description: "Temporal filter as ISO 8601 datetime or range 'start/end'."},
				},
				Required: []string{"url"},
			},
		},
		{
			Name:        "import_stac_asset",
			Description: "Import an asset from a remote STAC catalog as a local dataset. Downloads the asset and registers it.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"asset_url":   {Type: "string", Description: "Direct URL of the STAC asset to download (e.g. a GeoJSON or GeoParquet file URL)."},
					"name":        {Type: "string", Description: "Name for the imported dataset."},
					"format":      {Type: "string", Description: "Data format hint: 'geojson', 'parquet', 'gpkg', or 'csv'. Auto-detected if omitted."},
					"namespace":   {Type: "string", Description: "Optional dataset namespace prefix."},
					"collection":  {Type: "string", Description: "Optional STAC collection identifier."},
					"catalog_url": {Type: "string", Description: "Optional source catalog URL for provenance."},
					"project_id":  {Type: "string", Description: "Optional project to attach the imported dataset to. Also configurable globally via --project-id."},
				},
				Required: []string{"asset_url", "name"},
			},
		},
		{
			Name:        "search_stac",
			Description: "Search the local STAC catalog with spatial, temporal, and attribute filters. Supports bbox, datetime ranges, collection filtering, and CQL2 expressions.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"bbox":        {Type: "string", Description: "Bounding box as 'west,south,east,north'."},
					"datetime":    {Type: "string", Description: "Temporal filter as ISO 8601 datetime or range."},
					"collections": {Type: "string", Description: "Comma-separated collection IDs to search within."},
					"limit":       {Type: "string", Description: "Maximum results (default 10)."},
					"filter":      {Type: "string", Description: "CQL2-text filter expression."},
				},
			},
		},
		{
			Name:        "map_api",
			Description: "Access map-focused operations (publishing, raster metadata, slope/aspect, geodesic area/length, raster classification, OGC feature edits) through a strict allowlist. Mutating operations require confirm=true.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"operation": {
						Type:        "string",
						Description: "Map operation name.",
						Enum: []string{
							"publish_map", "list_published_maps", "unpublish_map", "get_published_map_stats",
							"update_map_embed_config", "get_public_map",
							"get_raster_info", "get_raster_stats", "get_raster_histogram", "get_raster_dimensions", "get_raster_values",
							"raster_zonal_stats", "export_raster_band", "raster_contour", "raster_viewshed", "raster_profile", "raster_kde",
							"raster_slope", "raster_aspect",
							"geodesic_area", "geodesic_length",
							"classify_kmeans", "classify_isodata", "classify_ml", "classify_rf",
							"create_feature", "update_feature", "delete_feature",
						},
					},
					"token":         {Type: "string", Description: "Published map token path parameter."},
					"name":          {Type: "string", Description: "Raster name path parameter."},
					"collection_id": {Type: "string", Description: "Collection identifier for feature edit operations."},
					"feature_id":    {Type: "string", Description: "Feature ID for update/delete operations."},
					"query":         {Type: "object", Description: "Optional query string key/value map."},
					"body":          {Type: "object", Description: "Optional JSON request body."},
					"confirm":       {Type: "boolean", Description: "Required true for mutating operations."},
				},
				Required: []string{"operation"},
			},
		},
	}
}
