package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
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

	projectID, err := optionalProjectID(params)
	if err != nil {
		return "", err
	}
	client = client.WithProjectID(projectID)

	switch name {
	case "list_datasets":
		return handleListDatasets(client)
	case "get_dataset_info":
		return handleGetDatasetInfo(client, params)
	case "query_features":
		return handleQueryFeatures(client, params)
	case "get_feature":
		return handleGetFeature(client, params)
	case "create_feature":
		return handleCreateFeature(client, params)
	case "update_feature":
		return handleUpdateFeature(client, params)
	case "delete_feature":
		return handleDeleteFeature(client, params)
	case "upload_dataset":
		return handleUploadDataset(client, params)
	case "import_source":
		return handleImportSource(client, params)
	case "get_scene_manifest":
		return handleGetSceneManifest(client)
	case "list_bodies":
		return handleListBodies(client)
	case "get_body":
		return handleGetBody(client, params)
	case "get_body_recipes":
		return handleGetBodyRecipes(client, params)
	case "execute_body_recipe":
		return handleExecuteBodyRecipe(client, params)
	case "list_operations":
		return handleListOperations(client, params)
	case "preflight_operation":
		return handlePreflightOperation(client, params)
	case "run_operation":
		return handleRunOperation(client, params)
	case "submit_operation_job":
		return handleSubmitOperationJob(client, params)
	case "submit_operation_batch":
		return handleSubmitOperationBatch(client, params)
	case "list_operation_jobs":
		return handleListOperationJobs(client, params)
	case "get_operation_job":
		return handleGetOperationJob(client, params)
	case "cancel_operation_job":
		return handleCancelOperationJob(client, params)
	case "rerun_operation_job":
		return handleRerunOperationJob(client, params)
	case "list_pipeline_operations":
		return handleListPipelineOperations(client)
	case "run_pipeline":
		return handleRunPipeline(client, params)
	case "list_pipelines":
		return handleListPipelines(client)
	case "get_pipeline":
		return handleGetPipeline(client, params)
	case "create_pipeline":
		return handleCreatePipeline(client, params)
	case "update_pipeline":
		return handleUpdatePipeline(client, params)
	case "delete_pipeline":
		return handleDeletePipeline(client, params)
	case "duplicate_pipeline":
		return handleDuplicatePipeline(client, params)
	case "execute_saved_pipeline":
		return handleExecuteSavedPipeline(client, params)
	case "list_pipeline_runs":
		return handleListPipelineRuns(client, params)
	case "get_pipeline_run":
		return handleGetPipelineRun(client, params)
	case "list_query_engines":
		return handleListQueryEngines(client)
	case "get_query_engine_info":
		return handleGetQueryEngineInfo(client, params)
	case "list_query_datasets":
		return handleListQueryDatasets(client, params)
	case "execute_sql":
		return handleExecuteSQL(client, params)
	case "save_sql_result":
		return handleSaveSQLResult(client, params)
	case "list_projects":
		return handleListProjects(client)
	case "get_project":
		return handleGetProject(client, params)
	case "create_project":
		return handleCreateProject(client, params)
	case "update_project":
		return handleUpdateProject(client, params)
	case "delete_project":
		return handleDeleteProject(client, params)
	case "get_project_workspace":
		return handleGetProjectWorkspace(client, params)
	case "set_project_workspace":
		return handleSetProjectWorkspace(client, params)
	case "publish_map":
		return handlePublishMap(client, params)
	case "list_published_maps":
		return handleListPublishedMaps(client)
	case "delete_published_map":
		return handleDeletePublishedMap(client, params)
	case "get_published_map_stats":
		return handleGetPublishedMapStats(client, params)
	case "update_map_embed_config":
		return handleUpdateMapEmbedConfig(client, params)
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
	name, err := requireStringLike(params, "name")
	if err != nil {
		return "", err
	}
	collection, err := client.GetCollection(name)
	if err != nil {
		return "", err
	}
	schema, err := client.GetDatasetSchema(name)
	if err != nil {
		return "", err
	}
	profile, err := client.GetDatasetProfile(name)
	if err != nil {
		return "", err
	}
	combined := map[string]interface{}{
		"collection": mustJSONObject(collection),
		"schema":     mustJSONObject(schema),
		"profile":    mustJSONObject(profile),
	}
	return formatJSON(combined), nil
}

func handleQueryFeatures(client *Client, params map[string]interface{}) (string, error) {
	id, err := requireStringLike(params, "collection_id")
	if err != nil {
		return "", err
	}
	query := map[string]string{}
	for _, key := range []string{"bbox", "filter", "datetime", "limit", "offset", "cursor"} {
		if value, ok := params[key]; ok {
			text, err := stringify(value)
			if err != nil {
				return "", fmt.Errorf("%s: %w", key, err)
			}
			if text != "" {
				query[key] = text
			}
		}
	}
	if value, ok := params["bbox_crs"]; ok {
		text, err := stringify(value)
		if err != nil {
			return "", fmt.Errorf("bbox_crs: %w", err)
		}
		if text != "" {
			query["bbox-crs"] = text
		}
	}
	if value, ok := params["crs"]; ok {
		text, err := stringify(value)
		if err != nil {
			return "", fmt.Errorf("crs: %w", err)
		}
		if text != "" {
			query["crs"] = text
		}
	}
	if _, ok := query["limit"]; !ok {
		query["limit"] = "10"
	}
	data, err := client.QueryFeatures(id, query)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleGetFeature(client *Client, params map[string]interface{}) (string, error) {
	collectionID, err := requireStringLike(params, "collection_id")
	if err != nil {
		return "", err
	}
	featureID, err := requireStringLike(params, "feature_id")
	if err != nil {
		return "", err
	}
	data, err := client.GetFeature(collectionID, featureID)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleCreateFeature(client *Client, params map[string]interface{}) (string, error) {
	collectionID, err := requireStringLike(params, "collection_id")
	if err != nil {
		return "", err
	}
	feature, err := requireObject(params, "feature")
	if err != nil {
		return "", err
	}
	data, err := client.CreateFeature(collectionID, feature)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleUpdateFeature(client *Client, params map[string]interface{}) (string, error) {
	collectionID, err := requireStringLike(params, "collection_id")
	if err != nil {
		return "", err
	}
	featureID, err := requireStringLike(params, "feature_id")
	if err != nil {
		return "", err
	}
	feature, err := requireObject(params, "feature")
	if err != nil {
		return "", err
	}
	data, err := client.UpdateFeature(collectionID, featureID, feature)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleDeleteFeature(client *Client, params map[string]interface{}) (string, error) {
	collectionID, err := requireStringLike(params, "collection_id")
	if err != nil {
		return "", err
	}
	featureID, err := requireStringLike(params, "feature_id")
	if err != nil {
		return "", err
	}
	data, err := client.DeleteFeature(collectionID, featureID)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleUploadDataset(client *Client, params map[string]interface{}) (string, error) {
	filePath, err := requireStringLike(params, "file_path")
	if err != nil {
		return "", err
	}
	name, _ := optionalStringLike(params, "name")
	bodyID, _ := optionalStringLike(params, "body_id")
	data, err := client.UploadFile(filePath, name, client.ProjectID, bodyID)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleImportSource(client *Client, params map[string]interface{}) (string, error) {
	payload, err := normalizeImportSourcePayload(params, client.ProjectID)
	if err != nil {
		return "", err
	}
	data, err := client.ImportSource(payload)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleGetSceneManifest(client *Client) (string, error) {
	data, err := client.GetSceneManifest()
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleListBodies(client *Client) (string, error) {
	data, err := client.ListBodies()
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleGetBody(client *Client, params map[string]interface{}) (string, error) {
	slug, err := requireStringLike(params, "slug")
	if err != nil {
		return "", err
	}
	data, err := client.GetBody(slug)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleGetBodyRecipes(client *Client, params map[string]interface{}) (string, error) {
	slug, err := requireStringLike(params, "slug")
	if err != nil {
		return "", err
	}
	data, err := client.GetBodyRecipes(slug)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleExecuteBodyRecipe(client *Client, params map[string]interface{}) (string, error) {
	slug, err := requireStringLike(params, "slug")
	if err != nil {
		return "", err
	}
	sourceID, err := requireStringLike(params, "source_id")
	if err != nil {
		return "", err
	}
	data, err := client.ExecuteBodyRecipe(slug, sourceID)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleListOperations(client *Client, params map[string]interface{}) (string, error) {
	domain, _ := optionalStringLike(params, "domain")
	data, err := client.ListOperations(domain)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handlePreflightOperation(client *Client, params map[string]interface{}) (string, error) {
	payload, operation, err := normalizeOperationPayload(params, client.ProjectID)
	if err != nil {
		return "", err
	}
	payload["operation"] = operation
	data, err := client.PreflightOperation(payload)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleRunOperation(client *Client, params map[string]interface{}) (string, error) {
	payload, operation, err := normalizeOperationPayload(params, client.ProjectID)
	if err != nil {
		return "", err
	}
	data, err := client.RunOperation(operation, payload)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleSubmitOperationJob(client *Client, params map[string]interface{}) (string, error) {
	payload, operation, err := normalizeOperationPayload(params, client.ProjectID)
	if err != nil {
		return "", err
	}
	data, err := client.SubmitOperationJob(operation, payload)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleSubmitOperationBatch(client *Client, params map[string]interface{}) (string, error) {
	rawJobs, err := requireArray(params, "jobs")
	if err != nil {
		return "", err
	}
	jobs := make([]map[string]interface{}, 0, len(rawJobs))
	for i, item := range rawJobs {
		job, ok := item.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("jobs[%d] must be an object", i)
		}
		normalized := map[string]interface{}{}
		if clientID, err := optionalStringLike(job, "client_id"); err == nil && clientID != "" {
			normalized["client_id"] = clientID
		}
		if dependsOn, ok := job["depends_on"]; ok {
			normalized["depends_on"] = dependsOn
		}
		requestObject, err := requireObject(job, "request")
		if err != nil {
			return "", fmt.Errorf("jobs[%d]: %w", i, err)
		}
		payload, operation, err := normalizeOperationPayload(requestObject, client.ProjectID)
		if err != nil {
			return "", fmt.Errorf("jobs[%d]: %w", i, err)
		}
		normalized["request"] = map[string]interface{}{
			"operation": operation,
		}
		for key, value := range payload {
			normalized["request"].(map[string]interface{})[key] = value
		}
		jobs = append(jobs, normalized)
	}
	data, err := client.SubmitOperationBatch(map[string]interface{}{"jobs": jobs})
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleListOperationJobs(client *Client, params map[string]interface{}) (string, error) {
	query := map[string]string{}
	for _, key := range []string{"status", "search", "limit", "offset"} {
		if value, ok := params[key]; ok {
			text, err := stringify(value)
			if err != nil {
				return "", fmt.Errorf("%s: %w", key, err)
			}
			if text != "" {
				query[key] = text
			}
		}
	}
	data, err := client.ListOperationJobs(query)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleGetOperationJob(client *Client, params map[string]interface{}) (string, error) {
	jobID, err := requireStringLike(params, "job_id")
	if err != nil {
		return "", err
	}
	data, err := client.GetOperationJob(jobID)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleCancelOperationJob(client *Client, params map[string]interface{}) (string, error) {
	jobID, err := requireStringLike(params, "job_id")
	if err != nil {
		return "", err
	}
	data, err := client.CancelOperationJob(jobID)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleRerunOperationJob(client *Client, params map[string]interface{}) (string, error) {
	jobID, err := requireStringLike(params, "job_id")
	if err != nil {
		return "", err
	}
	data, err := client.RerunOperationJob(jobID)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleListPipelineOperations(client *Client) (string, error) {
	data, err := client.ListPipelineOperations()
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleRunPipeline(client *Client, params map[string]interface{}) (string, error) {
	payload, err := normalizePipelinePayload(params)
	if err != nil {
		return "", err
	}
	data, err := client.RunPipeline(payload)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleListPipelines(client *Client) (string, error) {
	data, err := client.ListPipelines()
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleGetPipeline(client *Client, params map[string]interface{}) (string, error) {
	pipelineID, err := requireStringLike(params, "pipeline_id")
	if err != nil {
		return "", err
	}
	data, err := client.GetPipeline(pipelineID)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleCreatePipeline(client *Client, params map[string]interface{}) (string, error) {
	name, err := requireStringLike(params, "name")
	if err != nil {
		return "", err
	}
	payload := map[string]interface{}{"name": name}
	if description, err := optionalStringLike(params, "description"); err == nil && description != "" {
		payload["description"] = description
	}
	if graph, ok := params["graph"]; ok {
		payload["graph"] = graph
	}
	if canvas, ok := params["canvas"]; ok {
		payload["canvas"] = canvas
	}
	data, err := client.CreatePipeline(payload)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleUpdatePipeline(client *Client, params map[string]interface{}) (string, error) {
	pipelineID, err := requireStringLike(params, "pipeline_id")
	if err != nil {
		return "", err
	}
	name, err := requireStringLike(params, "name")
	if err != nil {
		return "", err
	}
	version, err := requireIntLike(params, "version")
	if err != nil {
		return "", err
	}
	payload := map[string]interface{}{"name": name, "version": version}
	if description, err := optionalStringLike(params, "description"); err == nil && description != "" {
		payload["description"] = description
	}
	if graph, ok := params["graph"]; ok {
		payload["graph"] = graph
	}
	if canvas, ok := params["canvas"]; ok {
		payload["canvas"] = canvas
	}
	data, err := client.UpdatePipeline(pipelineID, payload)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleDeletePipeline(client *Client, params map[string]interface{}) (string, error) {
	pipelineID, err := requireStringLike(params, "pipeline_id")
	if err != nil {
		return "", err
	}
	data, err := client.DeletePipeline(pipelineID)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleDuplicatePipeline(client *Client, params map[string]interface{}) (string, error) {
	pipelineID, err := requireStringLike(params, "pipeline_id")
	if err != nil {
		return "", err
	}
	data, err := client.DuplicatePipeline(pipelineID)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleExecuteSavedPipeline(client *Client, params map[string]interface{}) (string, error) {
	pipelineID, err := requireStringLike(params, "pipeline_id")
	if err != nil {
		return "", err
	}
	data, err := client.ExecutePipeline(pipelineID)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleListPipelineRuns(client *Client, params map[string]interface{}) (string, error) {
	pipelineID, err := requireStringLike(params, "pipeline_id")
	if err != nil {
		return "", err
	}
	data, err := client.ListPipelineRuns(pipelineID)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleGetPipelineRun(client *Client, params map[string]interface{}) (string, error) {
	runID, err := requireStringLike(params, "run_id")
	if err != nil {
		return "", err
	}
	data, err := client.GetPipelineRun(runID)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleListQueryEngines(client *Client) (string, error) {
	data, err := client.ListQueryEngines()
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleGetQueryEngineInfo(client *Client, params map[string]interface{}) (string, error) {
	engine, err := requireStringLike(params, "engine")
	if err != nil {
		return "", err
	}
	data, err := client.GetQueryEngineInfo(engine)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleListQueryDatasets(client *Client, params map[string]interface{}) (string, error) {
	engine, err := requireStringLike(params, "engine")
	if err != nil {
		return "", err
	}
	data, err := client.ListQueryDatasets(engine)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleExecuteSQL(client *Client, params map[string]interface{}) (string, error) {
	engine, err := requireStringLike(params, "engine")
	if err != nil {
		return "", err
	}
	payload, err := normalizeSQLExecutePayload(params)
	if err != nil {
		return "", err
	}
	data, err := client.ExecuteSQL(engine, payload)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleSaveSQLResult(client *Client, params map[string]interface{}) (string, error) {
	engine, err := requireStringLike(params, "engine")
	if err != nil {
		return "", err
	}
	payload, err := normalizeSQLSavePayload(params)
	if err != nil {
		return "", err
	}
	ensureProjectID(payload, client.ProjectID)
	data, err := client.SaveSQLResult(engine, payload)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleListProjects(client *Client) (string, error) {
	data, err := client.ListProjects()
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleGetProject(client *Client, params map[string]interface{}) (string, error) {
	projectID, err := requireStringLike(params, "project_id")
	if err != nil {
		return "", err
	}
	data, err := client.GetProject(projectID)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleCreateProject(client *Client, params map[string]interface{}) (string, error) {
	name, err := requireStringLike(params, "name")
	if err != nil {
		return "", err
	}
	payload := map[string]interface{}{"name": name}
	if description, err := optionalStringLike(params, "description"); err == nil && description != "" {
		payload["description"] = description
	}
	data, err := client.CreateProject(payload)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleUpdateProject(client *Client, params map[string]interface{}) (string, error) {
	projectID, err := requireStringLike(params, "project_id")
	if err != nil {
		return "", err
	}
	payload := map[string]interface{}{}
	if name, err := optionalStringLike(params, "name"); err == nil && name != "" {
		payload["name"] = name
	}
	if description, ok := params["description"]; ok {
		payload["description"] = description
	}
	data, err := client.UpdateProject(projectID, payload)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleDeleteProject(client *Client, params map[string]interface{}) (string, error) {
	projectID, err := requireStringLike(params, "project_id")
	if err != nil {
		return "", err
	}
	data, err := client.DeleteProject(projectID)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleGetProjectWorkspace(client *Client, params map[string]interface{}) (string, error) {
	projectID, err := requireStringLike(params, "project_id")
	if err != nil {
		return "", err
	}
	data, err := client.GetProjectWorkspace(projectID)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleSetProjectWorkspace(client *Client, params map[string]interface{}) (string, error) {
	projectID, err := requireStringLike(params, "project_id")
	if err != nil {
		return "", err
	}
	mapState, ok := params["map_state"]
	if !ok {
		return "", fmt.Errorf("missing required parameter: map_state")
	}
	payload := map[string]interface{}{"map_state": mapState}
	if layerStyles, ok := params["layer_styles"]; ok {
		payload["layer_styles"] = layerStyles
	}
	data, err := client.SetProjectWorkspace(projectID, payload)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handlePublishMap(client *Client, params map[string]interface{}) (string, error) {
	mapState, ok := params["map_state"]
	if !ok {
		return "", fmt.Errorf("missing required parameter: map_state")
	}
	payload := map[string]interface{}{"map_state": mapState}
	for _, key := range []string{"title", "description", "expires_hours", "embed_config"} {
		if value, ok := params[key]; ok {
			payload[key] = value
		}
	}
	data, err := client.PublishMap(payload)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleListPublishedMaps(client *Client) (string, error) {
	data, err := client.ListPublishedMaps()
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleDeletePublishedMap(client *Client, params map[string]interface{}) (string, error) {
	token, err := requireStringLike(params, "token")
	if err != nil {
		return "", err
	}
	data, err := client.DeletePublishedMap(token)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleGetPublishedMapStats(client *Client, params map[string]interface{}) (string, error) {
	token, err := requireStringLike(params, "token")
	if err != nil {
		return "", err
	}
	data, err := client.GetPublishedMapStats(token)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func handleUpdateMapEmbedConfig(client *Client, params map[string]interface{}) (string, error) {
	token, err := requireStringLike(params, "token")
	if err != nil {
		return "", err
	}
	embedConfig, err := requireObject(params, "embed_config")
	if err != nil {
		return "", err
	}
	data, err := client.UpdateMapEmbedConfig(token, embedConfig)
	if err != nil {
		return "", err
	}
	return formatJSON(data), nil
}

func normalizeImportSourcePayload(params map[string]interface{}, projectID string) (map[string]interface{}, error) {
	name, err := requireStringLike(params, "name")
	if err != nil {
		return nil, err
	}
	source, err := requireStringLike(params, "source")
	if err != nil {
		return nil, err
	}
	payload := map[string]interface{}{"name": name, "source": source}
	for _, key := range []string{"format", "crs", "body_id", "source_type", "catalog_url", "collection"} {
		if value, ok := params[key]; ok {
			payload[key] = value
		}
	}
	ensureProjectID(payload, projectID)
	return payload, nil
}

func normalizeOperationPayload(params map[string]interface{}, projectID string) (map[string]interface{}, string, error) {
	operation, err := requireStringLike(params, "operation")
	if err != nil {
		return nil, "", err
	}
	payload := map[string]interface{}{}
	for _, key := range []string{"input", "input_geojson", "params", "output_name", "output_format", "register"} {
		if value, ok := params[key]; ok {
			payload[key] = value
		}
	}
	if _, ok := payload["output_name"]; !ok {
		if alias, err := optionalStringLike(params, "output"); err == nil && alias != "" {
			payload["output_name"] = alias
		}
	}
	if _, ok := payload["output_format"]; !ok {
		if alias, err := optionalStringLike(params, "format"); err == nil && alias != "" {
			payload["output_format"] = alias
		}
	}
	if _, ok := payload["params"]; !ok {
		payload["params"] = map[string]interface{}{}
	}
	ensureProjectID(payload, projectID)
	return payload, operation, nil
}

func normalizePipelinePayload(params map[string]interface{}) (map[string]interface{}, error) {
	payload := map[string]interface{}{}
	if input, err := optionalStringLike(params, "input"); err == nil && input != "" {
		payload["input"] = input
	}
	if inputGeoJSON, ok := params["input_geojson"]; ok {
		payload["input_geojson"] = inputGeoJSON
	}
	steps, err := requireArray(params, "steps")
	if err != nil {
		return nil, err
	}
	payload["steps"] = steps
	if _, ok := params["output_name"]; ok {
		payload["output_name"] = params["output_name"]
	} else if alias, err := optionalStringLike(params, "output"); err == nil && alias != "" {
		payload["output_name"] = alias
	}
	if register, ok := params["register"]; ok {
		payload["register"] = register
	}
	return payload, nil
}

func normalizeSQLExecutePayload(params map[string]interface{}) (map[string]interface{}, error) {
	sqlText, _ := optionalStringLike(params, "sql")
	if sqlText == "" {
		alias, _ := optionalStringLike(params, "query")
		sqlText = alias
	}
	if strings.TrimSpace(sqlText) == "" {
		return nil, fmt.Errorf("missing required parameter: sql")
	}
	payload := map[string]interface{}{"sql": sqlText}
	for _, key := range []string{"limit", "timeout_sec", "format", "query_options"} {
		if value, ok := params[key]; ok {
			payload[key] = value
		}
	}
	return payload, nil
}

func normalizeSQLSavePayload(params map[string]interface{}) (map[string]interface{}, error) {
	payload, err := normalizeSQLExecutePayload(params)
	if err != nil {
		return nil, err
	}
	outputName, err := requireStringLike(params, "output_name")
	if err != nil {
		return nil, err
	}
	payload["output_name"] = outputName
	if geometryColumn, err := optionalStringLike(params, "geometry_column"); err == nil && geometryColumn != "" {
		payload["geometry_column"] = geometryColumn
	}
	return payload, nil
}

func formatJSON(value interface{}) string {
	var raw []byte
	switch v := value.(type) {
	case json.RawMessage:
		raw = v
	case []byte:
		raw = v
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		raw = data
	}
	if len(raw) == 0 {
		return "{}"
	}
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, raw, "", "  "); err == nil {
		return pretty.String()
	}
	return string(raw)
}

func mustJSONObject(raw json.RawMessage) interface{} {
	var value interface{}
	if err := json.Unmarshal(raw, &value); err != nil {
		return string(raw)
	}
	return value
}

func optionalProjectID(params map[string]interface{}) (string, error) {
	value, ok := params["project_id"]
	if !ok || value == nil {
		return "", nil
	}
	return stringify(value)
}

func ensureProjectID(payload map[string]interface{}, projectID string) {
	if strings.TrimSpace(projectID) == "" {
		return
	}
	if _, exists := payload["project_id"]; exists {
		return
	}
	payload["project_id"] = projectIDJSONValue(projectID)
}

func requireObject(params map[string]interface{}, key string) (map[string]interface{}, error) {
	value, ok := params[key]
	if !ok || value == nil {
		return nil, fmt.Errorf("missing required parameter: %s", key)
	}
	object, ok := value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("parameter %s must be an object", key)
	}
	return object, nil
}

func requireArray(params map[string]interface{}, key string) ([]interface{}, error) {
	value, ok := params[key]
	if !ok || value == nil {
		return nil, fmt.Errorf("missing required parameter: %s", key)
	}
	array, ok := value.([]interface{})
	if !ok || len(array) == 0 {
		return nil, fmt.Errorf("parameter %s must be a non-empty array", key)
	}
	return array, nil
}

func requireStringLike(params map[string]interface{}, key string) (string, error) {
	value, ok := params[key]
	if !ok || value == nil {
		return "", fmt.Errorf("missing required parameter: %s", key)
	}
	text, err := stringify(value)
	if err != nil {
		return "", fmt.Errorf("parameter %s: %w", key, err)
	}
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("parameter %s must not be empty", key)
	}
	return text, nil
}

func optionalStringLike(params map[string]interface{}, key string) (string, error) {
	value, ok := params[key]
	if !ok || value == nil {
		return "", nil
	}
	return stringify(value)
}

func requireIntLike(params map[string]interface{}, key string) (int64, error) {
	value, ok := params[key]
	if !ok || value == nil {
		return 0, fmt.Errorf("missing required parameter: %s", key)
	}
	switch v := value.(type) {
	case float64:
		return int64(v), nil
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case json.Number:
		return v.Int64()
	case string:
		n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parameter %s must be an integer", key)
		}
		return n, nil
	default:
		return 0, fmt.Errorf("parameter %s must be an integer", key)
	}
}

func stringify(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v), nil
	case float64:
		return strconv.FormatInt(int64(v), 10), nil
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32), nil
	case int:
		return strconv.Itoa(v), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case int32:
		return strconv.FormatInt(int64(v), 10), nil
	case json.Number:
		return v.String(), nil
	case bool:
		if v {
			return "true", nil
		}
		return "false", nil
	default:
		return "", fmt.Errorf("expected string-like value")
	}
}
