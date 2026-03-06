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
			Description: "List all datasets registered in Cairn with their names, formats, feature counts, and geometry types.",
			InputSchema: InputSchema{Type: "object"},
		},
		{
			Name:        "get_dataset_info",
			Description: "Get detailed information about a dataset including its schema (field names and types), CRS, bounds, feature count, and geometry type.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"collection_id": {Type: "string", Description: "The dataset/collection identifier."},
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
					"filter":        {Type: "string", Description: "CQL2 filter expression (e.g. \"population > 10000\")."},
					"limit":         {Type: "string", Description: "Maximum number of features to return (default 10)."},
					"offset":        {Type: "string", Description: "Number of features to skip for pagination."},
					"properties":    {Type: "string", Description: "Comma-separated list of properties to include in the response."},
					"sortby":        {Type: "string", Description: "Property to sort by, prefix with '-' for descending."},
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
					"file_path": {Type: "string", Description: "Local file path to the spatial data file to upload."},
				},
				Required: []string{"file_path"},
			},
		},
		{
			Name:        "run_process",
			Description: "Run a single geoprocessing operation on a dataset. Operations include: buffer, clip, simplify, reproject, centroid, convex_hull, intersection, union, difference, sjoin, dissolve, voronoi, spatial_stats, morans_i, hotspot, kernel_density, and more.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"operation": {Type: "string", Description: "The geoprocessing operation to run (e.g. 'buffer', 'clip', 'simplify')."},
					"input":     {Type: "string", Description: "Input dataset name."},
					"params": {
						Type:        "object",
						Description: "Operation-specific parameters (e.g. {\"distance\": 500} for buffer, {\"tolerance\": 0.001} for simplify).",
					},
					"output": {Type: "string", Description: "Output dataset name (optional, auto-generated if not provided)."},
				},
				Required: []string{"operation", "input"},
			},
		},
		{
			Name:        "run_pipeline",
			Description: "Run a multi-step geoprocessing pipeline. Each step's output feeds into the next step. Useful for chaining operations like buffer → clip → simplify.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
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
				},
				Required: []string{"steps"},
			},
		},
		{
			Name:        "convert_format",
			Description: "Convert a dataset between spatial data formats. Supported formats: geojson, shapefile, geopackage, kml, csv, flatgeobuf, parquet.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"input":  {Type: "string", Description: "Input dataset name."},
					"format": {Type: "string", Description: "Target format (e.g. 'geojson', 'shapefile', 'geopackage', 'kml', 'csv', 'flatgeobuf', 'parquet')."},
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
					"base":    {Type: "string", Description: "Base dataset name (the 'before' version)."},
					"compare": {Type: "string", Description: "Compare dataset name (the 'after' version)."},
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
					"lat": {Type: "string", Description: "Latitude."},
					"lon": {Type: "string", Description: "Longitude."},
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
						Description: "Optional intermediate waypoints.",
						Items:       &PropertySchema{Type: "object"},
					},
					"profile": {Type: "string", Description: "Routing profile: 'driving', 'walking', 'cycling' (default: 'driving')."},
				},
				Required: []string{"origin", "destination"},
			},
		},
		{
			Name:        "list_operations",
			Description: "List all available geoprocessing operations with their parameter schemas. Useful for discovering what operations are supported and what parameters they accept.",
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
			Name:        "import_from_catalog",
			Description: "Import a dataset from the built-in data catalog by its catalog ID. Downloads and registers the dataset for querying and processing.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"catalog_id": {Type: "string", Description: "The catalog entry ID to import."},
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
					"asset_url": {Type: "string", Description: "Direct URL of the STAC asset to download (e.g. a GeoJSON or GeoParquet file URL)."},
					"name":      {Type: "string", Description: "Name for the imported dataset."},
					"format":    {Type: "string", Description: "Data format hint: 'geojson', 'parquet', 'gpkg', or 'csv'. Auto-detected if omitted."},
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
	}
}
