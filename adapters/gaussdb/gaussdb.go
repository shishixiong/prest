package gaussdb

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/prest/prest/v2/adapters"
	"github.com/prest/prest/v2/adapters/gaussdb/statements"
	"github.com/prest/prest/v2/adapters/postgres"
	"github.com/prest/prest/v2/config"
)

// GaussDB adapter for Huawei GaussDB database
// Since GaussDB 100 is highly compatible with PostgreSQL,
// we embed the PostgreSQL adapter and override only the methods
// that need changes for GaussDB compatibility.
type GaussDB struct {
	postgres.Postgres
}

// Load initializes the GaussDB adapter
// Similar to postgres.Load(), but for GaussDB
func Load() {
	config.PrestConf.Adapter = &GaussDB{}

	// For now, we'll use PostgreSQL connection logic
	// TODO: Replace with GaussDB-specific connection initialization
	// when driver is available
	slog.Warn("GaussDB adapter using PostgreSQL connection logic (driver not available)")

	// Initialize database name in connection context
	if GetDatabase() == "" {
		SetDatabase(config.PrestConf.PGDatabase)
	}

	// Test the connection
	db, err := Get()
	if err != nil {
		slog.Error("GaussDB connection get error", "err", err)
		// Don't exit here, let the server decide what to do
		// In production, this should exit(1) like PostgreSQL adapter
	} else {
		// Close the test connection
		db.Close()
	}
}

// DatabaseClause returns a SELECT from URL query params for GaussDB
// Override to handle GaussDB-specific system table queries
func (adapter *GaussDB) DatabaseClause(req *http.Request) (query string, hasCount bool) {
	queries := req.URL.Query()
	countQuery := queries.Get("_count")

	// Use GaussDB-specific statements
	query = fmt.Sprintf(statements.DatabasesSelect, statements.FieldDatabaseName)
	if countQuery != "" {
		hasCount = true
		query = fmt.Sprintf(statements.DatabasesSelect, statements.FieldCountDatabaseName)
	}
	return
}

// SchemaClause returns a SELECT from URL query params for GaussDB
// Override to handle GaussDB-specific system table queries
func (adapter *GaussDB) SchemaClause(req *http.Request) (query string, hasCount bool) {
	queries := req.URL.Query()
	countQuery := queries.Get("_count")

	// Use GaussDB-specific statements
	query = fmt.Sprintf(statements.SchemasSelect, statements.FieldSchemaName)
	if countQuery != "" {
		hasCount = true
		query = fmt.Sprintf(statements.SchemasSelect, statements.FieldCountSchemaName)
	}
	return
}

// TableClause returns a SELECT from URL query params for GaussDB
// Override to handle GaussDB-specific system table queries
func (adapter *GaussDB) TableClause() (query string) {
	// Use GaussDB-specific statements
	query = statements.TablesSelect
	return
}

// ShowTable shows table structure for GaussDB
// Override to handle GaussDB-specific information schema queries
func (adapter *GaussDB) ShowTable(schema, table string) adapters.Scanner {
	// Use GaussDB-specific query
	return adapter.Query(statements.ShowTableQuery, table, schema)
}

// ShowTableCtx shows table structure for GaussDB with context
func (adapter *GaussDB) ShowTableCtx(ctx context.Context, schema, table string) adapters.Scanner {
	// Use GaussDB-specific query
	return adapter.QueryCtx(ctx, statements.ShowTableQuery, table, schema)
}

// SchemaTablesClause returns a SELECT from URL query params for GaussDB
// Override to handle GaussDB-specific system table queries
func (adapter *GaussDB) SchemaTablesClause() (query string) {
	// Use GaussDB-specific statements
	query = statements.SchemaTablesSelect
	return
}

// Note: Other methods will automatically use the embedded Postgres adapter's implementation
// We only need to override methods that are GaussDB-specific