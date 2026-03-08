# roteiro-agent

MCP (Model Context Protocol) server for [Cairn](https://github.com/i-norden/cairn) — a spatial data platform. Enables AI agents (Claude Desktop, VS Code, Cursor) to work with geospatial datasets, run geoprocessing operations, execute PostGIS queries, and more.

## Installation

```bash
go install github.com/i-norden/roteiro-agent@latest
```

Or build from source:

```bash
git clone https://github.com/i-norden/roteiro-agent
cd roteiro-agent
go build -o roteiro-agent .
```

## Usage

```bash
roteiro-agent --server-url http://localhost:8080 --api-key cairn_abc123
```

The server communicates via JSON-RPC 2.0 over stdio (stdin/stdout), following the MCP specification.

### Environment variables

| Variable | Flag | Description |
|----------|------|-------------|
| `CAIRN_SERVER_URL` | `--server-url` | Cairn server base URL |
| `CAIRN_API_KEY` | `--api-key` | Cairn API key |
| `CAIRN_SESSION_COOKIE` | `--session-cookie` | Session cookie (alternative to API key) |

## MCP Client Configuration

### Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "cairn": {
      "command": "roteiro-agent",
      "args": ["--server-url", "https://your-cairn-instance.com", "--api-key", "cairn_abc123"]
    }
  }
}
```

### VS Code (Copilot)

Add to `.vscode/mcp.json`:

```json
{
  "servers": {
    "cairn": {
      "command": "roteiro-agent",
      "args": ["--server-url", "http://localhost:8080", "--api-key", "cairn_abc123"]
    }
  }
}
```

### Claude Code

Add to `.mcp.json`:

```json
{
  "mcpServers": {
    "cairn": {
      "command": "roteiro-agent",
      "args": ["--server-url", "http://localhost:8080", "--api-key", "cairn_abc123"]
    }
  }
}
```

## Available Tools

| Tool | Description |
|------|-------------|
| `list_datasets` | List all registered datasets |
| `get_dataset_info` | Dataset schema, CRS, bounds, feature count |
| `get_dataset_schema` | Field names and types |
| `get_dataset_profile` | Statistical profile of a dataset |
| `query_features` | Query with bbox, CQL2 filter, limit, properties |
| `get_feature` | Single feature by ID |
| `upload_dataset` | Upload a spatial data file |
| `run_process` | Single geoprocessing operation (buffer, clip, simplify, etc.) |
| `run_pipeline` | Multi-step geoprocessing pipeline |
| `convert_format` | Convert between formats (GeoJSON, Shapefile, etc.) |
| `diff_datasets` | Compare two dataset versions |
| `execute_sql` | Run PostGIS SQL query |
| `list_spatial_tables` | List spatial tables in the database |
| `get_duckdb_info` | DuckDB SQL engine status/capabilities |
| `list_duckdb_datasets` | Datasets available to DuckDB SQL |
| `geocode` | Address to coordinates |
| `reverse_geocode` | Coordinates to address |
| `compute_route` | Driving/walking route computation |
| `compute_isochrone` | Travel-time isochrone polygons |
| `compute_route_matrix` | Origin-destination time/distance matrix |
| `compute_service_area` | Distance-based service area polygons |
| `list_operations` | Available geoprocessing operations |
| `browse_catalog` | Browse the built-in data catalog |
| `browse_catalog_enhanced` | Browse enhanced catalog with filters |
| `get_catalog_entry` | Get enhanced catalog entry by ID |
| `list_catalog_categories` | List catalog categories |
| `list_catalog_tags` | List catalog tags |
| `import_from_catalog` | Import a dataset from the data catalog |
| `browse_stac_catalog` | Browse a remote STAC catalog |
| `browse_stac_collections` | List collections in a remote STAC catalog |
| `browse_stac_items` | List items in a remote STAC collection |
| `import_stac_asset` | Import a STAC asset as a local dataset |
| `search_stac` | Search local STAC with bbox/datetime/CQL2 filters |
| `indoor_api` | Allowlisted Indoor GIS endpoint access (buildings/floors/spaces/navigation/analytics/sensors/booking/geofences/simulations). Mutations require `confirm=true`. |
| `map_api` | Allowlisted map endpoint access (publish/unpublish/stats/embed config, raster metadata/values, OGC feature edit ops). Mutations require `confirm=true`. |

## Example Workflows

**"Show me all parks larger than 10 acres near downtown"**
1. Agent calls `list_datasets` to find the parks dataset
2. Agent calls `query_features` with a CQL2 filter: `area_acres > 10` and bbox around downtown
3. Returns matching parks as GeoJSON features

**"Buffer all schools by 1km and find which residential zones intersect"**
1. Agent calls `run_pipeline` with two steps:
   - Buffer "schools" by 1000m
   - Spatial join the buffer result with "residential_zones"
2. Returns the intersection result

**"What's the average building height per neighborhood?"**
1. Agent calls `execute_sql` with PostGIS SQL:
   ```sql
   SELECT n.name, AVG(b.height) as avg_height
   FROM neighborhoods n
   JOIN buildings b ON ST_Intersects(n.geom, b.geom)
   GROUP BY n.name ORDER BY avg_height DESC
   ```

**"Import building footprints from a STAC catalog and calculate total area"**
1. Agent calls `browse_stac_collections` to discover available collections
2. Agent calls `browse_stac_items` to preview the buildings collection
3. Agent calls `import_stac_asset` to download and register the data
4. Agent calls `execute_sql` to calculate total building area with PostGIS

**"Find open data about transportation in our catalog"**
1. Agent calls `browse_catalog` with search="transportation"
2. Agent calls `import_from_catalog` to import the desired dataset
3. Agent calls `get_dataset_info` to inspect the imported data

## License

MIT
