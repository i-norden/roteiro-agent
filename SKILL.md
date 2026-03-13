# Roteiro Spatial Platform — Agent Guide

## What is Roteiro?

Roteiro is a full-featured spatial data platform. It stores, processes, and serves geospatial datasets. Think of it as a self-hosted GIS server with a REST API.

## Authentication

All requests require either an API key (`X-API-Key` header) or a session cookie. The roteiro-agent MCP server handles this automatically — you just need to provide credentials when starting it.

## Key Concepts

- **Dataset**: A named collection of spatial features (points, lines, polygons). Can be GeoJSON, Shapefile, GeoPackage, etc.
- **Collection**: OGC API term for a dataset. Used interchangeably.
- **Feature**: A single geographic entity with geometry and properties (attributes).
- **CQL2**: Common Query Language v2 — a standard for filtering features by attributes and spatial relationships.
- **Pipeline**: A chain of geoprocessing operations where each step's output feeds the next.

## Working with Data

### Discovering datasets

Start with `list_datasets` to see what's available. Use `get_dataset_info` to drill into a specific dataset's schema, CRS, extent, and feature count. Use `get_dataset_schema` for just the field types, or `get_dataset_profile` for statistical summaries.

### Querying features

Use `query_features` with:
- `bbox`: spatial bounding box filter (`west,south,east,north`)
- `filter`: CQL2 expression (e.g. `population > 10000 AND status = 'active'`)
- `limit`: max features (default 10, use higher values carefully)
- `properties`: comma-separated list of properties to include (reduces response size)
- `sortby`: property to sort by (prefix with `-` for descending)

### SQL queries

Use `execute_sql` for complex spatial queries. Roteiro exposes PostGIS, so all spatial functions are available:
- `ST_Area`, `ST_Length`, `ST_Distance` — measurements
- `ST_Buffer`, `ST_Intersection`, `ST_Union` — geometry operations
- `ST_Intersects`, `ST_Contains`, `ST_Within` — spatial predicates
- `ST_Transform` — coordinate system transformation

Queries must be SELECT-only (read-only).

## Geoprocessing Operations

Use `preflight_process` to validate and normalize a request first. Use `run_process` for synchronous execution, `submit_process_job` or `submit_process_batch` for async execution, and `run_pipeline` for chains.

Always call `list_operations` first to fetch the live server operation catalog and parameter names. The server now returns rich metadata including category, UI availability, projected-CRS requirements, and typed parameter definitions.

Important parameter names for common ops:
- `geodesic_buffer` uses metric `distance` in meters
- `clip` uses `mask`
- `sjoin` uses `right` and `predicate`
- `reproject` uses `from_crs` and `to_crs`
- `dissolve` uses `group_by`

Async process jobs expose queue state, phase, progress, and failure metadata via the `/api/process/jobs*` endpoints.

Use `list_analysis_operations` for advanced analysis catalog endpoints under `/api/analysis/operations`.

The async process workflow is available through `submit_process_job`, `submit_process_batch`, `list_process_jobs`, `get_process_job`, `cancel_process_job`, and `rerun_process_job`.

For raster analysis, use `run_raster_process` with file paths when you need the generic `/api/raster/process` endpoint. Typical operation families include terrain, hydrology, distance/cost, spectral/change, classification, and raster-vector conversion.

For registered raster datasets and JSON-returning raster endpoints, use `map_api` with operations such as `get_raster_info`, `get_raster_stats`, `get_raster_histogram`, `get_raster_dimensions`, `get_raster_values`, `raster_zonal_stats`, `export_raster_band`, `raster_contour`, `raster_viewshed`, `raster_profile`, and `raster_kde`.

## Data Catalog & STAC

Roteiro includes a built-in data catalog and supports importing from remote STAC (SpatioTemporal Asset Catalog) servers.

### Built-in catalog

Use `browse_catalog` to discover datasets available for import. Filter by `search` (text) or `category`. Use `import_from_catalog` with a `catalog_id` to download and register a dataset.

### Remote STAC catalogs

For external data sources:
1. `browse_stac_catalog` — inspect a remote STAC catalog by URL
2. `browse_stac_collections` — list available collections
3. `browse_stac_items` — preview items with optional `bbox` and `datetime` filters
4. `import_stac_asset` — download an asset URL and register it as a local dataset

### Local STAC search

Use `search_stac` to search Roteiro's own STAC endpoint with spatial (`bbox`), temporal (`datetime`), collection, and CQL2 (`filter`) criteria.

## Tips for Effective Use

1. **Start with discovery**: Always `list_datasets` first to understand what's available.
2. **Use small limits**: Default to `limit=10` when exploring. Increase only when needed.
3. **Prefer SQL for analytics**: For aggregations, joins, and complex spatial queries, `execute_sql` is more efficient than fetching features and computing client-side.
4. **Chain operations with pipelines**: Instead of running operations one by one, use `run_pipeline` to chain them in a single request.
5. **Check schemas before querying**: Use `get_dataset_schema` to see available fields before writing CQL2 filters or SQL.

## Common Patterns

### Find features near a point
```sql
SELECT * FROM parks
WHERE ST_DWithin(geom, ST_SetSRID(ST_MakePoint(-73.97, 40.77), 4326), 0.01)
LIMIT 20
```

### Aggregate by region
```sql
SELECT r.name, COUNT(p.*) as count, SUM(ST_Area(p.geom::geography)) as total_area_m2
FROM regions r JOIN parcels p ON ST_Intersects(r.geom, p.geom)
GROUP BY r.name ORDER BY count DESC
```

### Buffer and intersect
```json
{
  "steps": [
    {"operation": "buffer", "input": "schools", "params": {"distance": 1000}},
    {"operation": "intersect", "params": {"mask": "residential_zones"}}
  ]
}
```
