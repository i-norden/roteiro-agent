# Cairn Agent Guide

## Shape

`roteiro-agent` is a narrow MCP wrapper around Cairn's current stable workflows. Prefer the explicit tools over inventing raw REST calls.

## Core Workflows

### Data discovery and query

- Start with `list_datasets`.
- Use `get_dataset_info` before writing filters or SQL.
- Use `query_features` for bounded inspection and `get_feature` for a single record.

### Intake

- Use `upload_dataset` for local files.
- Use `import_source` for remote URLs or catalog-backed sources.
- Set `body_id` when the dataset belongs to Earth, Moon, Mars, or a tenant-defined body.

### Celestial bodies

- Use `get_scene_manifest` to inspect the current body-aware scene configuration.
- Use `list_bodies`, `get_body`, and `get_body_recipes` to discover body metadata.
- Use `execute_body_recipe` to trigger a configured recipe source for a body.

### Operations and pipelines

- Call `list_operations` first.
- Use `preflight_operation` before expensive operations.
- Use `run_operation` for synchronous work and the `*_operation_job` tools for async work.
- Use `run_pipeline` for ad hoc multi-step chains.
- Use saved pipeline tools only when the user is clearly asking about persisted workflows.

### SQL

- Use `list_query_engines` and `get_query_engine_info` to discover available engines.
- Use `execute_sql` for analysis and `save_sql_result` when the result should become a dataset.
- Always specify `engine`.

### Projects and publishing

- Use project tools for workspace state and basic lifecycle.
- Use publish tools for public map links and embed configuration.

## Guardrails

- Keep feature and query requests bounded with `limit` unless the user explicitly needs a large result.
- Prefer the explicit MCP tools here over legacy routes like `/api/process`, `/api/query/sql`, `/api/catalog`, `/api/stac`, or the old `map_api` wrapper.
- Do not assume Earth-only data. Carry `body_id` or body slug context through the workflow when the task is celestial.
