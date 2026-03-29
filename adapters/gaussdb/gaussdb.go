package gaussdb

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/prest/prest/v2/adapters"
	"github.com/prest/prest/v2/adapters/gaussdb/statements"
	"github.com/prest/prest/v2/adapters/postgres"
	"github.com/prest/prest/v2/adapters/scanner"
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
	adapter := &GaussDB{}
	config.PrestConf.Adapter = adapter

	// Initialize database name in connection context
	// Use the adapter's SetDatabase method to ensure consistency
	adapter.SetDatabase(config.PrestConf.PGDatabase)

	// Test the connection using the adapter's GetDatabase method
	// We need to ensure the connection uses the correct driver
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

// buildCreateGraphSQL builds all SQL statements needed for creating a graph.
// Returns the SQL statements to execute, the insert metadata SQL with placeholders,
// and the resolved graph type SQL value.
func buildCreateGraphSQL(database, schema, graphName, graphType string) ([]string, string, string) {
	// Determine graph type SQL
	var graphTypeSQL string
	switch graphType {
	case "property":
		graphTypeSQL = "PROPERTY"
	case "directed":
		graphTypeSQL = "DIRECTED"
	case "undirected":
		graphTypeSQL = "UNDIRECTED"
	default:
		graphTypeSQL = "PROPERTY"
	}

	// Create vertex and edge tables to simulate a graph
	vertexTable := fmt.Sprintf("%s_vertices", graphName)
	edgeTable := fmt.Sprintf("%s_edges", graphName)

	// Create vertex table
	vertexSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.%s.%s (
			id BIGINT PRIMARY KEY,
			label TEXT,
			properties JSONB DEFAULT '{}'::jsonb
		)`, database, schema, vertexTable)

	// Create edge table with foreign key references
	edgeSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.%s.%s (
			id BIGINT PRIMARY KEY,
			from_vertex BIGINT NOT NULL REFERENCES %s.%s.%s(id),
			to_vertex BIGINT NOT NULL REFERENCES %s.%s.%s(id),
			label TEXT,
			properties JSONB DEFAULT '{}'::jsonb
		)`, database, schema, edgeTable, database, schema, vertexTable, database, schema, vertexTable)

	// Create indexes for better performance
	vertexIdxSQL := fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_label ON %s.%s.%s(label)`, graphName, database, schema, vertexTable)
	edgeFromIdxSQL := fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_from ON %s.%s.%s(from_vertex)`, graphName, database, schema, edgeTable)
	edgeToIdxSQL := fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_to ON %s.%s.%s(to_vertex)`, graphName, database, schema, edgeTable)

	// Store graph metadata
	metadataSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.%s.prest_graph_metadata (
			graph_name TEXT PRIMARY KEY,
			graph_type TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`, database, schema)

	// GaussDB uses ON DUPLICATE KEY UPDATE instead of ON CONFLICT
	insertMetadataSQL := fmt.Sprintf(`
		INSERT INTO %s.%s.prest_graph_metadata (graph_name, graph_type)
		VALUES ($1, $2)
		ON DUPLICATE KEY UPDATE graph_type = VALUES(graph_type)
	`, database, schema)

	sqlStatements := []string{
		vertexSQL,
		edgeSQL,
		vertexIdxSQL,
		edgeFromIdxSQL,
		edgeToIdxSQL,
		metadataSQL,
	}

	return sqlStatements, insertMetadataSQL, graphTypeSQL
}

// CreateGraph creates a graph structure in GaussDB
// GaussDB uses MySQL-compatible ON DUPLICATE KEY UPDATE syntax instead of PostgreSQL's ON CONFLICT
func (adapter *GaussDB) CreateGraph(database, schema, graphName string, graphType string) adapters.Scanner {
	sqlStatements, insertMetadataSQL, graphTypeSQL := buildCreateGraphSQL(database, schema, graphName, graphType)

	// Execute all SQL statements in a transaction
	tx, err := adapter.GetTransaction()
	if err != nil {
		return &scanner.PrestScanner{Error: err}
	}

	for _, sql := range sqlStatements {
		if _, err := tx.Exec(sql); err != nil {
			tx.Rollback()
			return &scanner.PrestScanner{Error: err}
		}
	}

	if err := tx.Commit(); err != nil {
		return &scanner.PrestScanner{Error: err}
	}

	return adapter.ExecuteScripts("POST", insertMetadataSQL, []interface{}{graphName, graphTypeSQL})
}

// CreateGraphCtx creates a graph structure in GaussDB with context
func (adapter *GaussDB) CreateGraphCtx(ctx context.Context, database, schema, graphName string, graphType string) adapters.Scanner {
	sqlStatements, insertMetadataSQL, graphTypeSQL := buildCreateGraphSQL(database, schema, graphName, graphType)

	// Execute all SQL statements in a transaction with context
	tx, err := adapter.GetTransactionCtx(ctx)
	if err != nil {
		return &scanner.PrestScanner{Error: err}
	}

	for _, sql := range sqlStatements {
		if _, err := tx.Exec(sql); err != nil {
			tx.Rollback()
			return &scanner.PrestScanner{Error: err}
		}
	}

	if err := tx.Commit(); err != nil {
		return &scanner.PrestScanner{Error: err}
	}

	return adapter.ExecuteScriptsCtx(ctx, "POST", insertMetadataSQL, []interface{}{graphName, graphTypeSQL})
}
