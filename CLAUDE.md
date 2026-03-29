# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

pREST (PostgreSQL REST) is a Go-based RESTful API server that provides instant, realtime, and high-performance API endpoints on top of PostgreSQL and GaussDB databases. It automatically generates CRUD endpoints, supports authentication, authorization, custom queries, plugins, middleware, and vector similarity search.

## Development Environment

- **Go version**: 1.25.0 (see `go.mod`)
- **Module name**: `github.com/prest/prest/v2`
- **Entry point**: `cmd/prestd/main.go`
- **Configuration**: Via environment variables or `prest.toml` file
- **Database**: PostgreSQL 9.5+ (tested with PostgreSQL 16 in CI), GaussDB 100+ (beta support)
- **Dev Container**: VS Code devcontainer configuration available (`.devcontainer/devcontainer.json`) with pre-configured extensions and settings

## Common Commands

### Testing
```bash
# Run all tests using Docker Compose (requires Docker)
make test

# Equivalent direct command
docker compose -f docker-compose-test.yml up --abort-on-container-exit --exit-code-from tests
docker compose -f docker-compose-test.yml down -v --remove-orphans

# Run Go tests directly (requires local PostgreSQL)
go test -v -race -failfast -covermode=atomic -coverprofile=coverage.out ./...
```

### Linting
```bash
# Run golangci-lint (requires installation)
golangci-lint run --timeout=5m

# Run go fmt
go fmt ./...

# Run go vet
go vet ./...

# Run misspell (requires installation)
misspell -error -locale US .
```

### Dependency Management
```bash
# Download dependencies
go mod download

# Tidy dependencies
go mod tidy

# Verify dependencies
go mod verify
```

### Building
```bash
# Build the binary
go build -o prestd ./cmd/prestd/main.go

# Build Docker image locally
docker build -t prest/prest:latest .

# Build with version information (release style)
docker build \
  --build-arg VERSION=v1.0.0 \
  --build-arg COMMIT=hash \
  --build-arg DATE=2026-02-11 \
  -t prest/prest:latest .
```

### Development Server
```bash
# Set required environment variables
export PREST_PG_USER=postgres
export PREST_PG_DATABASE=prest
export PREST_PG_PORT=5432
export PREST_HTTP_PORT=3010

# Run the server
go run ./cmd/prestd/main.go
```

### Database Migrations
```bash
# Run migrations up
go run ./cmd/prestd/main.go migrate up --path ./testdata/migrations

# Other migration commands (down, redo, reset, next, version)
go run ./cmd/prestd/main.go migrate [down|redo|reset|next|version]
```

### Code Generation
```bash
# Generate mocks for adapters
make mockgen
```

### Docker Compose
```bash
# Start development environment
make dc-up

# Stop and clean up
make dc-down
```

### Plugin Development
```bash
# Build a plugin (example plugin in lib/src/hello.go)
go build -o ./lib/hello.so -buildmode=plugin ./lib/src/hello.go

# Plugin endpoint will be available at /_PLUGIN/hello/<function>
# Note: Plugin system not supported on Windows
```

## Architecture Overview

### Core Components

1. **Configuration** (`config/`):
   - Central configuration via `config.PrestConf`
   - Supports environment variables, config file (`prest.toml`), and defaults
   - Manages database connections, authentication, caching, CORS, etc.

2. **Adapters** (`adapters/`):
   - Abstract database operations via `adapters.Adapter` interface
   - PostgreSQL implementation in `adapters/postgres/`
   - GaussDB implementation in `adapters/gaussdb/` (extends PostgreSQL adapter for Huawei GaussDB compatibility)
   - Handles SQL generation, query building, and permissions
   - Key responsibilities: SQL generation, query building, permissions checking

3. **Controllers** (`controllers/`):
   - HTTP handlers for REST endpoints
   - `tables.go`: CRUD operations on tables
   - `databases.go`: Database listing
   - `schemas.go`: Schema listing
   - `auth.go`: Authentication endpoints
   - `sql.go`: Custom SQL execution
   - `healthcheck.go`: Health endpoint
   - `vectors.go`: Vector similarity search operations (create/delete vector indexes, vector search)

4. **Router** (`router/`):
   - Defines all API routes in `router.go`
   - Main endpoints:
     - `GET /databases`, `GET /schemas`, `GET /tables`
     - `GET /{database}/{schema}/{table}` (CRUD operations)
     - `POST /auth` (when auth enabled)
     - `GET /_QUERIES/{queriesLocation}/{script}` (custom queries)
     - `GET /_PLUGIN/{file}/{func}` (plugin execution, non-Windows)
     - `GET /_health` (health check)
     - `POST /{database}/{schema}/{table}/_vector/search` (vector similarity search)
     - `POST /{database}/{schema}/{table}/_vector/index` (create vector index)
     - `DELETE /{database}/{schema}/{table}/_vector/index` (delete vector index)

5. **Middlewares** (`middlewares/`):
   - `AuthMiddleware`: JWT authentication
   - `AccessControl`: Table/field-level permissions
   - `ExposureMiddleware`: Controls database/schema/table exposure
   - `CacheMiddleware`: Response caching
   - Applied to all CRUD routes via Negroni

6. **Plugins** (`plugins/`):
   - Go plugin system for custom middleware and handlers
   - Windows not supported due to Go plugin limitations

7. **Cache** (`cache/`):
   - BuntDB-based caching system
   - Configurable cache duration and storage paths

8. **Transactions** (`transactions/`):
   - Transaction management utilities

9. **Helpers** (`helpers/`):
   - Utility functions for HTTP responses, URL parsing, etc.

### Key Design Patterns

- **Adapter Pattern**: Database abstraction allowing support for multiple databases (PostgreSQL and GaussDB)
- **Middleware Chain**: All CRUD requests pass through authentication, access control, exposure, and cache middlewares
- **Configuration Centralization**: Single `Prest` struct holds all runtime configuration
- **Plugin System**: Extensible via Go plugins for custom middleware and handlers

## Configuration

Configuration is loaded via `config.Load()` in this order:
1. Environment variables (prefixed with `PREST_`, dots replaced with underscores)
2. Configuration file (`prest.toml` or path from `PREST_CONF` env var)
3. Default values (see `config/config.go` for defaults)

Key configuration sections:
- `auth`: Authentication settings (enabled, table, schema, etc.)
- `http`: Server host, port, timeout
- `pg`: PostgreSQL connection parameters
- `jwt`: JWT authentication settings
- `cors`: CORS configuration
- `cache`: Caching settings
- `access`: Table and field-level permissions
- `expose`: Control database/schema/table listing exposure
- `vector`: Vector search configuration (enabled, default distance metric, dimensions, index type)

## Testing Strategy

- Tests require PostgreSQL running (handled automatically via Docker Compose in CI)
- Test databases: `prest-test` and `secondary-db`
- Test setup: Creates databases, loads schema (`testdata/schema.sql`), runs migrations
- Plugin tests: Builds plugin `.so` file from `lib/src/hello.go`
- Use `testdata/runtest.sh` as the main test script
- GaussDB tests: Separate test scripts available (`test_gaussdb_api.sh`), requires manual GaussDB instance setup

## Release Process

- Uses GoReleaser for multi-platform builds (see `.goreleaser.yml`)
- Docker images built with version arguments
- Release builds inject version, commit, and date via ldflags

## Important Notes

- Windows: Plugin endpoints are disabled on Windows due to Go plugin system limitations
- Authentication: When enabled, requires JWT token for CRUD operations (except `/auth`)
- Public Mode: Running without access restriction shows warning
- Debug Mode: Enabled via `debug` config or `PREST_DEBUG` env var
- Single vs Multi Database: Configured via `pg.single` (defaults to single database mode)
- GaussDB Support: Beta support for Huawei GaussDB database (configure via `database.type = "gaussdb"` or `PREST_DATABASE_TYPE=gaussdb`)
- Vector Search: Vector similarity search support for PostgreSQL (configure via `vector.enabled = true` or `PREST_VECTOR_ENABLED=true`)