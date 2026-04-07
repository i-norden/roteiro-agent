# roteiro-agent

MCP server for Cairn's current public API.

The agent now exposes a smaller, explicit tool surface built around the stable workflows:

- datasets and collection queries
- uploads and remote dataset intake
- celestial body metadata and recipe execution
- unified vector operations and async jobs
- ad hoc and saved pipelines
- SQL query control plane
- projects and workspace state
- published map management

Legacy tools for `/api/process`, raster processing, catalog browsing, STAC import, routing, geocoding, and the old `map_api` catch-all have been removed.

## Install

```bash
go install github.com/i-norden/roteiro-agent@latest
```

## Usage

```bash
roteiro-agent --server-url http://localhost:8080 --api-key Roteiro_abc123 --project-id 42
```

The server speaks JSON-RPC 2.0 over stdio and follows the MCP protocol.

## Environment Variables

| Variable | Flag | Description |
|----------|------|-------------|
| `ROTEIRO_SERVER_URL` | `--server-url` | Cairn server base URL |
| `ROTEIRO_API_KEY` | `--api-key` | API key |
| `ROTEIRO_SESSION_COOKIE` | `--session-cookie` | Session cookie alternative |
| `ROTEIRO_PROJECT_ID` | `--project-id` | Default project scope sent as `X-Project-ID` |

## Tool Groups

- data: `list_datasets`, `get_dataset_info`, `query_features`, `get_feature`, `create_feature`, `update_feature`, `delete_feature`, `upload_dataset`, `import_source`
- celestial: `get_scene_manifest`, `list_bodies`, `get_body`, `get_body_recipes`, `execute_body_recipe`
- operations: `list_operations`, `preflight_operation`, `run_operation`, `submit_operation_job`, `submit_operation_batch`, `list_operation_jobs`, `get_operation_job`, `cancel_operation_job`, `rerun_operation_job`
- pipelines: `list_pipeline_operations`, `run_pipeline`, `list_pipelines`, `get_pipeline`, `create_pipeline`, `update_pipeline`, `delete_pipeline`, `duplicate_pipeline`, `execute_saved_pipeline`, `list_pipeline_runs`, `get_pipeline_run`
- SQL: `list_query_engines`, `get_query_engine_info`, `list_query_datasets`, `execute_sql`, `save_sql_result`
- projects: `list_projects`, `get_project`, `create_project`, `update_project`, `delete_project`, `get_project_workspace`, `set_project_workspace`
- publishing: `publish_map`, `list_published_maps`, `delete_published_map`, `get_published_map_stats`, `update_map_embed_config`

## Notes

- Most tools accept project scoping through the agent's global `--project-id` or a per-call `project_id` override.
- `upload_dataset` and `import_source` both support `body_id` so tenant-defined celestial bodies flow through the intake path.
- SQL tools operate against Cairn's engine-aware control plane, so `engine` is required where applicable.
