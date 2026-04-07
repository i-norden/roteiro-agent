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

func tool(name, description string, properties map[string]PropertySchema, required ...string) Tool {
	return Tool{
		Name:        name,
		Description: description,
		InputSchema: InputSchema{
			Type:       "object",
			Properties: properties,
			Required:   required,
		},
	}
}

func stringProp(description string) PropertySchema {
	return PropertySchema{Type: "string", Description: description}
}

func boolProp(description string) PropertySchema {
	return PropertySchema{Type: "boolean", Description: description}
}

func numberProp(description string) PropertySchema {
	return PropertySchema{Type: "number", Description: description}
}

func objectProp(description string) PropertySchema {
	return PropertySchema{Type: "object", Description: description}
}

func arrayProp(description string, itemType string) PropertySchema {
	return PropertySchema{Type: "array", Description: description, Items: &PropertySchema{Type: itemType}}
}

// AllTools returns the complete list of MCP tools that this server exposes.
func AllTools() []Tool {
	return []Tool{
		tool("list_datasets", "List datasets registered in Cairn.", map[string]PropertySchema{
			"project_id": stringProp("Optional project scope override. Also configurable globally via --project-id."),
		}),
		tool("get_dataset_info", "Get combined collection, schema, and profile information for a dataset.", map[string]PropertySchema{
			"name":       stringProp("Dataset or collection identifier."),
			"project_id": stringProp("Optional project scope override. Also configurable globally via --project-id."),
		}, "name"),
		tool("query_features", "Query features from a collection with optional spatial and attribute filters.", map[string]PropertySchema{
			"collection_id": stringProp("Collection identifier."),
			"bbox":          stringProp("Bounding box filter as 'west,south,east,north'."),
			"bbox_crs":      stringProp("Optional CRS for the bbox coordinates, forwarded as bbox-crs."),
			"crs":           stringProp("Optional CRS identifier for returned geometries."),
			"filter":        stringProp("CQL2 filter expression."),
			"datetime":      stringProp("Temporal filter as RFC3339 instant or interval."),
			"limit":         stringProp("Maximum features to return. Defaults to 10."),
			"offset":        stringProp("Pagination offset."),
			"cursor":        stringProp("Opaque pagination cursor when supported by the backend."),
			"project_id":    stringProp("Optional project scope override. Also configurable globally via --project-id."),
		}, "collection_id"),
		tool("get_feature", "Fetch a single feature from a collection.", map[string]PropertySchema{
			"collection_id": stringProp("Collection identifier."),
			"feature_id":    stringProp("Feature identifier."),
			"project_id":    stringProp("Optional project scope override. Also configurable globally via --project-id."),
		}, "collection_id", "feature_id"),
		tool("create_feature", "Create a feature in a collection.", map[string]PropertySchema{
			"collection_id": stringProp("Collection identifier."),
			"feature":       objectProp("GeoJSON Feature payload."),
			"project_id":    stringProp("Optional project scope override. Also configurable globally via --project-id."),
		}, "collection_id", "feature"),
		tool("update_feature", "Replace an existing feature in a collection.", map[string]PropertySchema{
			"collection_id": stringProp("Collection identifier."),
			"feature_id":    stringProp("Feature identifier."),
			"feature":       objectProp("GeoJSON Feature payload."),
			"project_id":    stringProp("Optional project scope override. Also configurable globally via --project-id."),
		}, "collection_id", "feature_id", "feature"),
		tool("delete_feature", "Delete a feature from a collection.", map[string]PropertySchema{
			"collection_id": stringProp("Collection identifier."),
			"feature_id":    stringProp("Feature identifier."),
			"project_id":    stringProp("Optional project scope override. Also configurable globally via --project-id."),
		}, "collection_id", "feature_id"),
		tool("upload_dataset", "Upload a local geospatial file and register it as a dataset.", map[string]PropertySchema{
			"file_path":  stringProp("Local file path to upload."),
			"name":       stringProp("Optional dataset name. Defaults to the file stem if omitted."),
			"body_id":    stringProp("Optional celestial body identifier for the dataset."),
			"project_id": stringProp("Optional project override. Also configurable globally via --project-id."),
		}, "file_path"),
		tool("import_source", "Import a remote or catalog-backed source through the current dataset intake endpoint.", map[string]PropertySchema{
			"name":        stringProp("Dataset name to register."),
			"source":      stringProp("Source URL or source reference."),
			"format":      stringProp("Optional format hint such as geojson, parquet, gpkg, tif, or csv."),
			"crs":         stringProp("Optional CRS hint."),
			"body_id":     stringProp("Optional celestial body identifier."),
			"project_id":  stringProp("Optional project override. Also configurable globally via --project-id."),
			"source_type": stringProp("Optional source type such as remote_url."),
			"catalog_url": stringProp("Optional source catalog URL for provenance."),
			"collection":  stringProp("Optional upstream collection identifier."),
		}, "name", "source"),
		tool("get_scene_manifest", "Fetch the current body-aware scene manifest.", map[string]PropertySchema{}),
		tool("list_bodies", "List available celestial bodies for the current tenant.", map[string]PropertySchema{}),
		tool("get_body", "Fetch a single celestial body definition by slug.", map[string]PropertySchema{
			"slug": stringProp("Body slug."),
		}, "slug"),
		tool("get_body_recipes", "List recipe sources configured for a celestial body.", map[string]PropertySchema{
			"slug": stringProp("Body slug."),
		}, "slug"),
		tool("execute_body_recipe", "Execute a configured recipe source for a celestial body.", map[string]PropertySchema{
			"slug":      stringProp("Body slug."),
			"source_id": stringProp("Recipe source identifier."),
		}, "slug", "source_id"),
		tool("list_operations", "List the current unified operation catalog.", map[string]PropertySchema{
			"domain": stringProp("Optional operation domain filter."),
		}),
		tool("preflight_operation", "Validate and normalize an operation request before execution.", map[string]PropertySchema{
			"operation":     stringProp("Operation identifier."),
			"input":         stringProp("Input dataset name."),
			"input_geojson": objectProp("Inline GeoJSON input."),
			"params":        objectProp("Operation-specific parameters."),
			"output":        stringProp("Compatibility alias for output_name."),
			"output_name":   stringProp("Optional output dataset name."),
			"format":        stringProp("Compatibility alias for output_format."),
			"output_format": stringProp("Requested output format."),
			"register":      boolProp("Whether to register the result as a dataset."),
			"project_id":    stringProp("Optional project override. Also configurable globally via --project-id."),
		}, "operation"),
		tool("run_operation", "Run a synchronous unified operation.", map[string]PropertySchema{
			"operation":     stringProp("Operation identifier."),
			"input":         stringProp("Input dataset name."),
			"input_geojson": objectProp("Inline GeoJSON input."),
			"params":        objectProp("Operation-specific parameters."),
			"output":        stringProp("Compatibility alias for output_name."),
			"output_name":   stringProp("Optional output dataset name."),
			"format":        stringProp("Compatibility alias for output_format."),
			"output_format": stringProp("Requested output format."),
			"register":      boolProp("Whether to register the result as a dataset."),
			"project_id":    stringProp("Optional project override. Also configurable globally via --project-id."),
		}, "operation"),
		tool("submit_operation_job", "Queue an asynchronous unified operation job.", map[string]PropertySchema{
			"operation":     stringProp("Operation identifier."),
			"input":         stringProp("Input dataset name."),
			"input_geojson": objectProp("Inline GeoJSON input."),
			"params":        objectProp("Operation-specific parameters."),
			"output":        stringProp("Compatibility alias for output_name."),
			"output_name":   stringProp("Optional output dataset name."),
			"format":        stringProp("Compatibility alias for output_format."),
			"output_format": stringProp("Requested output format."),
			"register":      boolProp("Whether to register the result as a dataset."),
			"project_id":    stringProp("Optional project override. Also configurable globally via --project-id."),
		}, "operation"),
		tool("submit_operation_batch", "Submit a dependent batch of operation jobs.", map[string]PropertySchema{
			"jobs": {
				Type:        "array",
				Description: "Array of batch jobs. Each item supports client_id, depends_on, and request.",
				Items: &PropertySchema{
					Type: "object",
					Properties: map[string]PropertySchema{
						"client_id":  stringProp("Optional client-side identifier."),
						"depends_on": arrayProp("Optional dependency references.", "string"),
						"request":    objectProp("Operation request payload matching submit_operation_job."),
					},
				},
			},
		}, "jobs"),
		tool("list_operation_jobs", "List asynchronous operation jobs.", map[string]PropertySchema{
			"status": stringProp("Optional status filter."),
			"search": stringProp("Optional free-text search."),
			"limit":  stringProp("Maximum jobs to return."),
			"offset": stringProp("Pagination offset."),
		}),
		tool("get_operation_job", "Fetch a queued operation job by ID.", map[string]PropertySchema{
			"job_id": stringProp("Operation job identifier."),
		}, "job_id"),
		tool("cancel_operation_job", "Cancel an operation job by ID.", map[string]PropertySchema{
			"job_id": stringProp("Operation job identifier."),
		}, "job_id"),
		tool("rerun_operation_job", "Re-submit a previous operation job.", map[string]PropertySchema{
			"job_id": stringProp("Operation job identifier."),
		}, "job_id"),
		tool("list_pipeline_operations", "List the current ad hoc pipeline operation catalog.", map[string]PropertySchema{}),
		tool("run_pipeline", "Run an ad hoc multi-step pipeline.", map[string]PropertySchema{
			"input":         stringProp("Input dataset name."),
			"input_geojson": objectProp("Inline GeoJSON input."),
			"steps": {
				Type:        "array",
				Description: "Ordered pipeline steps.",
				Items: &PropertySchema{
					Type: "object",
					Properties: map[string]PropertySchema{
						"operation": stringProp("Operation identifier."),
						"params":    objectProp("Operation-specific parameters."),
					},
				},
			},
			"output":      stringProp("Compatibility alias for output_name."),
			"output_name": stringProp("Optional output dataset name."),
			"register":    boolProp("Whether to register the result as a dataset."),
		}, "steps"),
		tool("list_pipelines", "List persisted pipelines.", map[string]PropertySchema{}),
		tool("get_pipeline", "Fetch a persisted pipeline by ID.", map[string]PropertySchema{
			"pipeline_id": stringProp("Pipeline identifier."),
		}, "pipeline_id"),
		tool("create_pipeline", "Create a persisted pipeline definition.", map[string]PropertySchema{
			"name":        stringProp("Pipeline name."),
			"description": stringProp("Optional pipeline description."),
			"graph":       objectProp("Pipeline graph payload."),
			"canvas":      objectProp("Pipeline canvas payload."),
		}, "name"),
		tool("update_pipeline", "Update a persisted pipeline definition.", map[string]PropertySchema{
			"pipeline_id": stringProp("Pipeline identifier."),
			"name":        stringProp("Pipeline name."),
			"description": stringProp("Optional pipeline description."),
			"graph":       objectProp("Pipeline graph payload."),
			"canvas":      objectProp("Pipeline canvas payload."),
			"version":     numberProp("Current pipeline version."),
		}, "pipeline_id", "name", "version"),
		tool("delete_pipeline", "Delete a persisted pipeline by ID.", map[string]PropertySchema{
			"pipeline_id": stringProp("Pipeline identifier."),
		}, "pipeline_id"),
		tool("duplicate_pipeline", "Duplicate a persisted pipeline by ID.", map[string]PropertySchema{
			"pipeline_id": stringProp("Pipeline identifier."),
		}, "pipeline_id"),
		tool("execute_saved_pipeline", "Execute a persisted pipeline by ID.", map[string]PropertySchema{
			"pipeline_id": stringProp("Pipeline identifier."),
		}, "pipeline_id"),
		tool("list_pipeline_runs", "List execution runs for a persisted pipeline.", map[string]PropertySchema{
			"pipeline_id": stringProp("Pipeline identifier."),
		}, "pipeline_id"),
		tool("get_pipeline_run", "Fetch a pipeline run by ID.", map[string]PropertySchema{
			"run_id": stringProp("Pipeline run identifier."),
		}, "run_id"),
		tool("list_query_engines", "List available SQL query engines.", map[string]PropertySchema{}),
		tool("get_query_engine_info", "Fetch SQL engine info for a specific engine.", map[string]PropertySchema{
			"engine": stringProp("Query engine identifier such as duckdb or postgis."),
		}, "engine"),
		tool("list_query_datasets", "List datasets visible to a specific SQL engine.", map[string]PropertySchema{
			"engine": stringProp("Query engine identifier such as duckdb or postgis."),
		}, "engine"),
		tool("execute_sql", "Execute a SQL query through Cairn's engine-aware query control plane.", map[string]PropertySchema{
			"engine":        stringProp("Query engine identifier such as duckdb or postgis."),
			"query":         stringProp("Compatibility alias for sql."),
			"sql":           stringProp("SQL text to execute."),
			"limit":         numberProp("Optional result limit."),
			"timeout_sec":   numberProp("Optional execution timeout in seconds."),
			"format":        stringProp("Result format, typically json or arrow."),
			"query_options": objectProp("Optional engine-specific options."),
		}, "engine"),
		tool("save_sql_result", "Execute SQL and save the result as a dataset.", map[string]PropertySchema{
			"engine":          stringProp("Query engine identifier such as duckdb or postgis."),
			"query":           stringProp("Compatibility alias for sql."),
			"sql":             stringProp("SQL text to execute."),
			"output_name":     stringProp("Dataset name for the saved result."),
			"geometry_column": stringProp("Optional geometry column override."),
			"limit":           numberProp("Optional result limit."),
			"timeout_sec":     numberProp("Optional execution timeout in seconds."),
			"query_options":   objectProp("Optional engine-specific options."),
			"project_id":      stringProp("Optional project override. Also configurable globally via --project-id."),
		}, "engine", "output_name"),
		tool("list_projects", "List accessible projects.", map[string]PropertySchema{}),
		tool("get_project", "Fetch a project by ID.", map[string]PropertySchema{
			"project_id": stringProp("Project identifier."),
		}, "project_id"),
		tool("create_project", "Create a new project.", map[string]PropertySchema{
			"name":        stringProp("Project name."),
			"description": stringProp("Optional description."),
		}, "name"),
		tool("update_project", "Update an existing project.", map[string]PropertySchema{
			"project_id":  stringProp("Project identifier."),
			"name":        stringProp("Project name."),
			"description": stringProp("Optional description. Use an empty string to clear when supported by the server."),
		}, "project_id"),
		tool("delete_project", "Delete a project by ID.", map[string]PropertySchema{
			"project_id": stringProp("Project identifier."),
		}, "project_id"),
		tool("get_project_workspace", "Fetch workspace state for a project.", map[string]PropertySchema{
			"project_id": stringProp("Project identifier."),
		}, "project_id"),
		tool("set_project_workspace", "Replace workspace state for a project.", map[string]PropertySchema{
			"project_id":   stringProp("Project identifier."),
			"map_state":    objectProp("Workspace map state."),
			"layer_styles": objectProp("Optional layer style map."),
		}, "project_id", "map_state"),
		tool("publish_map", "Publish a map state and receive a public token.", map[string]PropertySchema{
			"title":         stringProp("Optional published title."),
			"description":   stringProp("Optional published description."),
			"map_state":     objectProp("Serializable map state payload."),
			"expires_hours": numberProp("Optional expiration window in hours."),
			"embed_config":  objectProp("Optional embed configuration."),
		}, "map_state"),
		tool("list_published_maps", "List published map links.", map[string]PropertySchema{}),
		tool("delete_published_map", "Delete a published map by token.", map[string]PropertySchema{
			"token": stringProp("Published map token."),
		}, "token"),
		tool("get_published_map_stats", "Fetch statistics for a published map.", map[string]PropertySchema{
			"token": stringProp("Published map token."),
		}, "token"),
		tool("update_map_embed_config", "Update embed configuration for a published map.", map[string]PropertySchema{
			"token":        stringProp("Published map token."),
			"embed_config": objectProp("Embed configuration payload."),
		}, "token", "embed_config"),
	}
}
