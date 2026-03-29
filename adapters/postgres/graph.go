package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/prest/prest/v2/adapters"
	"github.com/prest/prest/v2/adapters/scanner"
)

// CreateGraph creates a graph structure in Postgres
// PostgreSQL doesn't have native CREATE GRAPH syntax, so we create tables to simulate a graph
func (adapter *Postgres) CreateGraph(database, schema, graphName string, graphType string) adapters.Scanner {
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
	// Vertex table: id, label, properties (JSONB)
	// Edge table: id, from_vertex, to_vertex, label, properties (JSONB)
	vertexTable := fmt.Sprintf("%s_vertices", graphName)
	edgeTable := fmt.Sprintf("%s_edges", graphName)

	// Create vertex table
	vertexSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.%s.%s (
			id BIGINT PRIMARY KEY,
			label TEXT,
			properties JSONB DEFAULT '{}'::jsonb
		)`, database, schema, vertexTable)

	// Create edge table
	edgeSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.%s.%s (
			id BIGINT PRIMARY KEY,
			from_vertex BIGINT NOT NULL REFERENCES %s.%s.%s(id),
			to_vertex BIGINT NOT NULL REFERENCES %s.%s.%s(id),
			label TEXT,
			properties JSONB DEFAULT '{}'::jsonb,
			CHECK (from_vertex != to_vertex)
		)`, database, schema, edgeTable, database, schema, vertexTable, database, schema, vertexTable)

	// Create indexes for better performance
	vertexIdxSQL := fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_label ON %s.%s.%s(label)`, graphName, database, schema, vertexTable)
	edgeFromIdxSQL := fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_from ON %s.%s.%s(from_vertex)`, graphName, database, schema, edgeTable)
	edgeToIdxSQL := fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_to ON %s.%s.%s(to_vertex)`, graphName, database, schema, edgeTable)

	// Store graph metadata (optional)
	metadataSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.%s.prest_graph_metadata (
			graph_name TEXT PRIMARY KEY,
			graph_type TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`, database, schema)

	insertMetadataSQL := fmt.Sprintf(`
		INSERT INTO %s.%s.prest_graph_metadata (graph_name, graph_type)
		VALUES ($1, $2)
		ON CONFLICT (graph_name) DO UPDATE SET graph_type = EXCLUDED.graph_type
	`, database, schema)

	// Execute all SQL statements in a transaction
	tx, err := adapter.GetTransaction()
	if err != nil {
		return &scanner.PrestScanner{Error: err}
	}

	// Execute each SQL statement separately
	sqlStatements := []string{
		vertexSQL,
		edgeSQL,
		vertexIdxSQL,
		edgeFromIdxSQL,
		edgeToIdxSQL,
		metadataSQL,
	}

	for _, sql := range sqlStatements {
		if _, err := tx.Exec(sql); err != nil {
			tx.Rollback()
			return &scanner.PrestScanner{Error: err}
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return &scanner.PrestScanner{Error: err}
	}

	// Insert/update metadata
	return adapter.ExecuteScripts("POST", insertMetadataSQL, []interface{}{graphName, graphTypeSQL})
}

// CreateGraphCtx creates a graph structure with context
func (adapter *Postgres) CreateGraphCtx(ctx context.Context, database, schema, graphName string, graphType string) adapters.Scanner {
	// Same implementation but with context
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

	// Create edge table
	edgeSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.%s.%s (
			id BIGINT PRIMARY KEY,
			from_vertex BIGINT NOT NULL REFERENCES %s.%s.%s(id),
			to_vertex BIGINT NOT NULL REFERENCES %s.%s.%s(id),
			label TEXT,
			properties JSONB DEFAULT '{}'::jsonb,
			CHECK (from_vertex != to_vertex)
		)`, database, schema, edgeTable, database, schema, vertexTable, database, schema, vertexTable)

	// Create indexes
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

	insertMetadataSQL := fmt.Sprintf(`
		INSERT INTO %s.%s.prest_graph_metadata (graph_name, graph_type)
		VALUES ($1, $2)
		ON CONFLICT (graph_name) DO UPDATE SET graph_type = EXCLUDED.graph_type
	`, database, schema)

	// Execute all SQL statements in a transaction with context
	tx, err := adapter.GetTransactionCtx(ctx)
	if err != nil {
		return &scanner.PrestScanner{Error: err}
	}

	// Execute each SQL statement separately
	sqlStatements := []string{
		vertexSQL,
		edgeSQL,
		vertexIdxSQL,
		edgeFromIdxSQL,
		edgeToIdxSQL,
		metadataSQL,
	}

	for _, sql := range sqlStatements {
		if _, err := tx.Exec(sql); err != nil {
			tx.Rollback()
			return &scanner.PrestScanner{Error: err}
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return &scanner.PrestScanner{Error: err}
	}

	// Insert/update metadata
	return adapter.ExecuteScriptsCtx(ctx, "POST", insertMetadataSQL, []interface{}{graphName, graphTypeSQL})
}

// DeleteGraph deletes a graph structure from Postgres
func (adapter *Postgres) DeleteGraph(database, schema, graphName string) adapters.Scanner {
	// Delete graph tables and metadata
	vertexTable := fmt.Sprintf("%s_vertices", graphName)
	edgeTable := fmt.Sprintf("%s_edges", graphName)

	// Drop tables in correct order (edges first due to foreign key constraints)
	dropEdgeSQL := fmt.Sprintf(`DROP TABLE IF EXISTS %s.%s.%s CASCADE`, database, schema, edgeTable)
	dropVertexSQL := fmt.Sprintf(`DROP TABLE IF EXISTS %s.%s.%s CASCADE`, database, schema, vertexTable)
	deleteMetadataSQL := fmt.Sprintf(`DELETE FROM %s.%s.prest_graph_metadata WHERE graph_name = $1`, database, schema)

	// Execute in transaction
	tx, err := adapter.GetTransaction()
	if err != nil {
		return &scanner.PrestScanner{Error: err}
	}

	// Execute drop statements
	if _, err := tx.Exec(dropEdgeSQL); err != nil {
		tx.Rollback()
		return &scanner.PrestScanner{Error: err}
	}

	if _, err := tx.Exec(dropVertexSQL); err != nil {
		tx.Rollback()
		return &scanner.PrestScanner{Error: err}
	}

	// Delete metadata with parameter
	if _, err := tx.Exec(deleteMetadataSQL, graphName); err != nil {
		tx.Rollback()
		return &scanner.PrestScanner{Error: err}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return &scanner.PrestScanner{Error: err}
	}

	// Return empty success scanner
	return &scanner.PrestScanner{}
}

// DeleteGraphCtx deletes a graph structure with context
func (adapter *Postgres) DeleteGraphCtx(ctx context.Context, database, schema, graphName string) adapters.Scanner {
	// Delete graph tables and metadata
	vertexTable := fmt.Sprintf("%s_vertices", graphName)
	edgeTable := fmt.Sprintf("%s_edges", graphName)

	// Drop tables in correct order
	dropEdgeSQL := fmt.Sprintf(`DROP TABLE IF EXISTS %s.%s.%s CASCADE`, database, schema, edgeTable)
	dropVertexSQL := fmt.Sprintf(`DROP TABLE IF EXISTS %s.%s.%s CASCADE`, database, schema, vertexTable)
	deleteMetadataSQL := fmt.Sprintf(`DELETE FROM %s.%s.prest_graph_metadata WHERE graph_name = $1`, database, schema)

	// Execute in transaction with context
	tx, err := adapter.GetTransactionCtx(ctx)
	if err != nil {
		return &scanner.PrestScanner{Error: err}
	}

	// Execute drop statements
	if _, err := tx.Exec(dropEdgeSQL); err != nil {
		tx.Rollback()
		return &scanner.PrestScanner{Error: err}
	}

	if _, err := tx.Exec(dropVertexSQL); err != nil {
		tx.Rollback()
		return &scanner.PrestScanner{Error: err}
	}

	// Delete metadata with parameter
	if _, err := tx.Exec(deleteMetadataSQL, graphName); err != nil {
		tx.Rollback()
		return &scanner.PrestScanner{Error: err}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return &scanner.PrestScanner{Error: err}
	}

	// Return empty success scanner
	return &scanner.PrestScanner{}
}

// AddVertices adds vertices to a graph in Postgres
// Postgres uses INSERT INTO vertex_table syntax for graph vertices
func (adapter *Postgres) AddVertices(database, schema, graphName string, vertices []map[string]interface{}) adapters.Scanner {
	if len(vertices) == 0 {
		return adapter.Query("SELECT 1")
	}

	// Build batch insert SQL
	// Assuming vertex table name follows pattern: graphname_vertices
	tableName := fmt.Sprintf("%s_vertices", graphName)
	columns := []string{"id", "label", "properties"}
	placeholders := []string{}
	values := []interface{}{}

	for i, vertex := range vertices {
		id := vertex["id"]
		label := vertex["label"]
		properties := vertex["properties"]

		// Convert properties to JSON string
		propsJSON := "{}"
		if properties != nil {
			if propsBytes, err := json.Marshal(properties); err == nil {
				propsJSON = string(propsBytes)
			}
		}

		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d::jsonb)", i*3+1, i*3+2, i*3+3))
		values = append(values, id, label, propsJSON)
	}

	sql := fmt.Sprintf(`INSERT INTO %s.%s.%s (%s) VALUES %s`,
		database, schema, tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	return adapter.Insert(sql, values...)
}

// AddVerticesCtx adds vertices with context
func (adapter *Postgres) AddVerticesCtx(ctx context.Context, database, schema, graphName string, vertices []map[string]interface{}) adapters.Scanner {
	if len(vertices) == 0 {
		return adapter.QueryCtx(ctx, "SELECT 1")
	}

	tableName := fmt.Sprintf("%s_vertices", graphName)
	columns := []string{"id", "label", "properties"}
	placeholders := []string{}
	values := []interface{}{}

	for i, vertex := range vertices {
		id := vertex["id"]
		label := vertex["label"]
		properties := vertex["properties"]

		propsJSON := "{}"
		if properties != nil {
			if propsBytes, err := json.Marshal(properties); err == nil {
				propsJSON = string(propsBytes)
			}
		}

		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d::jsonb)", i*3+1, i*3+2, i*3+3))
		values = append(values, id, label, propsJSON)
	}

	sql := fmt.Sprintf(`INSERT INTO %s.%s.%s (%s) VALUES %s`,
		database, schema, tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	return adapter.InsertCtx(ctx, sql, values...)
}

// AddEdges adds edges to a graph in Postgres
func (adapter *Postgres) AddEdges(database, schema, graphName string, edges []map[string]interface{}) adapters.Scanner {
	if len(edges) == 0 {
		return adapter.Query("SELECT 1")
	}

	// Assuming edge table name follows pattern: graphname_edges
	tableName := fmt.Sprintf("%s_edges", graphName)
	columns := []string{"id", "from_vertex", "to_vertex", "label", "properties"}
	placeholders := []string{}
	values := []interface{}{}

	for i, edge := range edges {
		id := edge["id"]
		from := edge["from"]
		to := edge["to"]
		label := edge["label"]
		properties := edge["properties"]

		propsJSON := "{}"
		if properties != nil {
			if propsBytes, err := json.Marshal(properties); err == nil {
				propsJSON = string(propsBytes)
			}
		}

		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d::jsonb)", i*5+1, i*5+2, i*5+3, i*5+4, i*5+5))
		values = append(values, id, from, to, label, propsJSON)
	}

	sql := fmt.Sprintf(`INSERT INTO %s.%s.%s (%s) VALUES %s`,
		database, schema, tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	return adapter.Insert(sql, values...)
}

// AddEdgesCtx adds edges with context
func (adapter *Postgres) AddEdgesCtx(ctx context.Context, database, schema, graphName string, edges []map[string]interface{}) adapters.Scanner {
	if len(edges) == 0 {
		return adapter.QueryCtx(ctx, "SELECT 1")
	}

	tableName := fmt.Sprintf("%s_edges", graphName)
	columns := []string{"id", "from_vertex", "to_vertex", "label", "properties"}
	placeholders := []string{}
	values := []interface{}{}

	for i, edge := range edges {
		id := edge["id"]
		from := edge["from"]
		to := edge["to"]
		label := edge["label"]
		properties := edge["properties"]

		propsJSON := "{}"
		if properties != nil {
			if propsBytes, err := json.Marshal(properties); err == nil {
				propsJSON = string(propsBytes)
			}
		}

		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d::jsonb)", i*5+1, i*5+2, i*5+3, i*5+4, i*5+5))
		values = append(values, id, from, to, label, propsJSON)
	}

	sql := fmt.Sprintf(`INSERT INTO %s.%s.%s (%s) VALUES %s`,
		database, schema, tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	return adapter.InsertCtx(ctx, sql, values...)
}

// GraphQuery executes a graph query in Postgres
// Supports different query types: traverse, path, neighbors, match
func (adapter *Postgres) GraphQuery(database, schema, graphName string, queryType string, queryParams map[string]interface{}) adapters.Scanner {
	// Build graph query based on query type
	var sql string
	switch queryType {
	case "traverse":
		sql = buildTraverseQuery(database, schema, graphName, queryParams)
	case "path":
		sql = buildPathQuery(database, schema, graphName, queryParams)
	case "neighbors":
		sql = buildNeighborsQuery(database, schema, graphName, queryParams)
	case "match":
		sql = buildMatchQuery(database, schema, graphName, queryParams)
	default:
		// Default to simple vertex query
		sql = fmt.Sprintf(`SELECT * FROM %s.%s.%s_vertices LIMIT 100`, database, schema, graphName)
	}

	// Extract parameters for prepared statement
	values := extractQueryParams(queryType, queryParams)
	return adapter.Query(sql, values...)
}

// GraphQueryCtx executes a graph query with context
func (adapter *Postgres) GraphQueryCtx(ctx context.Context, database, schema, graphName string, queryType string, queryParams map[string]interface{}) adapters.Scanner {
	var sql string
	switch queryType {
	case "traverse":
		sql = buildTraverseQuery(database, schema, graphName, queryParams)
	case "path":
		sql = buildPathQuery(database, schema, graphName, queryParams)
	case "neighbors":
		sql = buildNeighborsQuery(database, schema, graphName, queryParams)
	case "match":
		sql = buildMatchQuery(database, schema, graphName, queryParams)
	default:
		sql = fmt.Sprintf(`SELECT * FROM %s.%s.%s_vertices LIMIT 100`, database, schema, graphName)
	}

	values := extractQueryParams(queryType, queryParams)
	return adapter.QueryCtx(ctx, sql, values...)
}

// GetVertex retrieves a vertex by ID
func (adapter *Postgres) GetVertex(database, schema, graphName string, vertexID interface{}) adapters.Scanner {
	tableName := fmt.Sprintf("%s_vertices", graphName)
	sql := fmt.Sprintf(`SELECT * FROM %s.%s.%s WHERE id = $1`, database, schema, tableName)
	return adapter.Query(sql, vertexID)
}

// GetVertexCtx retrieves a vertex by ID with context
func (adapter *Postgres) GetVertexCtx(ctx context.Context, database, schema, graphName string, vertexID interface{}) adapters.Scanner {
	tableName := fmt.Sprintf("%s_vertices", graphName)
	sql := fmt.Sprintf(`SELECT * FROM %s.%s.%s WHERE id = $1`, database, schema, tableName)
	return adapter.QueryCtx(ctx, sql, vertexID)
}

// GetEdge retrieves an edge by ID
func (adapter *Postgres) GetEdge(database, schema, graphName string, edgeID interface{}) adapters.Scanner {
	tableName := fmt.Sprintf("%s_edges", graphName)
	sql := fmt.Sprintf(`SELECT * FROM %s.%s.%s WHERE id = $1`, database, schema, tableName)
	return adapter.Query(sql, edgeID)
}

// GetEdgeCtx retrieves an edge by ID with context
func (adapter *Postgres) GetEdgeCtx(ctx context.Context, database, schema, graphName string, edgeID interface{}) adapters.Scanner {
	tableName := fmt.Sprintf("%s_edges", graphName)
	sql := fmt.Sprintf(`SELECT * FROM %s.%s.%s WHERE id = $1`, database, schema, tableName)
	return adapter.QueryCtx(ctx, sql, edgeID)
}

// UpdateVertex updates a vertex properties
func (adapter *Postgres) UpdateVertex(database, schema, graphName string, vertexID interface{}, properties map[string]interface{}) adapters.Scanner {
	tableName := fmt.Sprintf("%s_vertices", graphName)
	// Convert properties to JSON
	propsJSON := "{}"
	if properties != nil {
		if propsBytes, err := json.Marshal(properties); err == nil {
			propsJSON = string(propsBytes)
		}
	}
	sql := fmt.Sprintf(`UPDATE %s.%s.%s SET properties = $1 WHERE id = $2`, database, schema, tableName)
	return adapter.Update(sql, propsJSON, vertexID)
}

// UpdateVertexCtx updates a vertex properties with context
func (adapter *Postgres) UpdateVertexCtx(ctx context.Context, database, schema, graphName string, vertexID interface{}, properties map[string]interface{}) adapters.Scanner {
	tableName := fmt.Sprintf("%s_vertices", graphName)
	propsJSON := "{}"
	if properties != nil {
		if propsBytes, err := json.Marshal(properties); err == nil {
			propsJSON = string(propsBytes)
		}
	}
	sql := fmt.Sprintf(`UPDATE %s.%s.%s SET properties = $1 WHERE id = $2`, database, schema, tableName)
	return adapter.UpdateCtx(ctx, sql, propsJSON, vertexID)
}

// UpdateEdge updates an edge properties
func (adapter *Postgres) UpdateEdge(database, schema, graphName string, edgeID interface{}, properties map[string]interface{}) adapters.Scanner {
	tableName := fmt.Sprintf("%s_edges", graphName)
	propsJSON := "{}"
	if properties != nil {
		if propsBytes, err := json.Marshal(properties); err == nil {
			propsJSON = string(propsBytes)
		}
	}
	sql := fmt.Sprintf(`UPDATE %s.%s.%s SET properties = $1 WHERE id = $2`, database, schema, tableName)
	return adapter.Update(sql, propsJSON, edgeID)
}

// UpdateEdgeCtx updates an edge properties with context
func (adapter *Postgres) UpdateEdgeCtx(ctx context.Context, database, schema, graphName string, edgeID interface{}, properties map[string]interface{}) adapters.Scanner {
	tableName := fmt.Sprintf("%s_edges", graphName)
	propsJSON := "{}"
	if properties != nil {
		if propsBytes, err := json.Marshal(properties); err == nil {
			propsJSON = string(propsBytes)
		}
	}
	sql := fmt.Sprintf(`UPDATE %s.%s.%s SET properties = $1 WHERE id = $2`, database, schema, tableName)
	return adapter.UpdateCtx(ctx, sql, propsJSON, edgeID)
}

// DeleteVertex deletes a vertex
func (adapter *Postgres) DeleteVertex(database, schema, graphName string, vertexID interface{}) adapters.Scanner {
	tableName := fmt.Sprintf("%s_vertices", graphName)
	sql := fmt.Sprintf(`DELETE FROM %s.%s.%s WHERE id = $1`, database, schema, tableName)
	return adapter.Delete(sql, vertexID)
}

// DeleteVertexCtx deletes a vertex with context
func (adapter *Postgres) DeleteVertexCtx(ctx context.Context, database, schema, graphName string, vertexID interface{}) adapters.Scanner {
	tableName := fmt.Sprintf("%s_vertices", graphName)
	sql := fmt.Sprintf(`DELETE FROM %s.%s.%s WHERE id = $1`, database, schema, tableName)
	return adapter.DeleteCtx(ctx, sql, vertexID)
}

// DeleteEdge deletes an edge
func (adapter *Postgres) DeleteEdge(database, schema, graphName string, edgeID interface{}) adapters.Scanner {
	tableName := fmt.Sprintf("%s_edges", graphName)
	sql := fmt.Sprintf(`DELETE FROM %s.%s.%s WHERE id = $1`, database, schema, tableName)
	return adapter.Delete(sql, edgeID)
}

// DeleteEdgeCtx deletes an edge with context
func (adapter *Postgres) DeleteEdgeCtx(ctx context.Context, database, schema, graphName string, edgeID interface{}) adapters.Scanner {
	tableName := fmt.Sprintf("%s_edges", graphName)
	sql := fmt.Sprintf(`DELETE FROM %s.%s.%s WHERE id = $1`, database, schema, tableName)
	return adapter.DeleteCtx(ctx, sql, edgeID)
}

// Helper functions for building graph queries

func buildTraverseQuery(database, schema, graphName string, params map[string]interface{}) string {
	// Set default values for max_depth and direction if not provided
	if params["max_depth"] == nil {
		params["max_depth"] = 10
	}
	if params["direction"] == nil {
		params["direction"] = "out"
	}
	direction := params["direction"].(string)

	// Simple traversal query using recursive CTE
	// This is a simplified implementation assuming standard table structure
	return fmt.Sprintf(`
		WITH RECURSIVE traversal AS (
			SELECT v.*, 0 as depth
			FROM %s.%s.%s_vertices v
			WHERE v.id = $1
			UNION ALL
			SELECT v.*, t.depth + 1
			FROM traversal t
			JOIN %s.%s.%s_edges e ON %s
			JOIN %s.%s.%s_vertices v ON v.id = %s
			WHERE t.depth < $2
		)
		SELECT row_to_json(t) as vertex FROM (SELECT id, label, properties, depth FROM traversal) t
	`, database, schema, graphName,
		database, schema, graphName,
		getEdgeJoinCondition(direction, "t.id"),
		database, schema, graphName,
		getVertexJoinCondition(direction, "e"))
}

func buildPathQuery(database, schema, graphName string, params map[string]interface{}) string {
	// Set default value for max_depth if not provided
	if params["max_depth"] == nil {
		params["max_depth"] = 10
	}

	return fmt.Sprintf(`
		WITH RECURSIVE path_find AS (
			SELECT
				ARRAY[v.id] as path,
				v.id as current,
				0 as depth
			FROM %s.%s.%s_vertices v
			WHERE v.id = $1
			UNION ALL
			SELECT
				p.path || v.id,
				v.id,
				p.depth + 1
			FROM path_find p
			JOIN %s.%s.%s_edges e ON e.from_vertex = p.current
			JOIN %s.%s.%s_vertices v ON v.id = e.to_vertex
			WHERE p.depth < $2
				AND v.id != ALL(p.path)  -- Avoid cycles
				AND v.id != $3  -- Stop when reaching end vertex
		)
		SELECT path FROM path_find WHERE current = $3
		LIMIT 1
	`, database, schema, graphName,
		database, schema, graphName,
		database, schema, graphName)
}

func buildNeighborsQuery(database, schema, graphName string, params map[string]interface{}) string {
	direction := params["direction"]
	if direction == nil {
		direction = "out"
	}

	var joinCondition string
	switch direction.(string) {
	case "out":
		joinCondition = "e.from_vertex = v.id"
	case "in":
		joinCondition = "e.to_vertex = v.id"
	case "both":
		joinCondition = "(e.from_vertex = v.id OR e.to_vertex = v.id)"
	}

	return fmt.Sprintf(`
		SELECT row_to_json(n) as vertex
		FROM (
			SELECT DISTINCT n.id, n.label, n.properties
			FROM %s.%s.%s_vertices v
			JOIN %s.%s.%s_edges e ON v.id = $1 AND %s
			JOIN %s.%s.%s_vertices n ON (n.id = e.to_vertex OR n.id = e.from_vertex) AND n.id != v.id
		) n
	`, database, schema, graphName,
		database, schema, graphName, joinCondition,
		database, schema, graphName)
}

func buildMatchQuery(database, schema, graphName string, params map[string]interface{}) string {
	// Simple pattern matching query
	// For now, return all vertices and edges (simplified)
	return fmt.Sprintf(`
		SELECT
			v.* as vertex,
			e.* as edge
		FROM %s.%s.%s_vertices v
		LEFT JOIN %s.%s.%s_edges e ON e.from_vertex = v.id OR e.to_vertex = v.id
		LIMIT 100
	`, database, schema, graphName,
		database, schema, graphName)
}

func getEdgeJoinCondition(direction, vertexAlias string) string {
	switch direction {
	case "out":
		return fmt.Sprintf("%s = e.from_vertex", vertexAlias)
	case "in":
		return fmt.Sprintf("%s = e.to_vertex", vertexAlias)
	case "both":
		return fmt.Sprintf("(%s = e.from_vertex OR %s = e.to_vertex)", vertexAlias, vertexAlias)
	default:
		return fmt.Sprintf("%s = e.from_vertex", vertexAlias)
	}
}

func getVertexJoinCondition(direction, edgeAlias string) string {
	switch direction {
	case "out":
		return fmt.Sprintf("%s.to_vertex", edgeAlias)
	case "in":
		return fmt.Sprintf("%s.from_vertex", edgeAlias)
	case "both":
		return fmt.Sprintf("CASE WHEN %s.from_vertex = t.id THEN %s.to_vertex ELSE %s.from_vertex END", edgeAlias, edgeAlias, edgeAlias)
	default:
		return fmt.Sprintf("%s.to_vertex", edgeAlias)
	}
}

func extractQueryParams(queryType string, params map[string]interface{}) []interface{} {
	values := []interface{}{}

	switch queryType {
	case "traverse":
		// Traverse needs start_vertex and max_depth
		if startVertex, ok := params["start_vertex"]; ok {
			values = append(values, startVertex)
		}
		if maxDepth, ok := params["max_depth"]; ok {
			values = append(values, maxDepth)
		}
	case "path":
		// Path needs start_vertex, max_depth, and end_vertex
		if startVertex, ok := params["start_vertex"]; ok {
			values = append(values, startVertex)
		}
		if maxDepth, ok := params["max_depth"]; ok {
			values = append(values, maxDepth)
		}
		if endVertex, ok := params["end_vertex"]; ok {
			values = append(values, endVertex)
		}
	case "neighbors":
		// Neighbors needs start_vertex only
		if startVertex, ok := params["start_vertex"]; ok {
			values = append(values, startVertex)
		}
	case "match":
		// Match needs no parameters
	default:
		// Default: extract all known parameters
		if startVertex, ok := params["start_vertex"]; ok {
			values = append(values, startVertex)
		}
		if maxDepth, ok := params["max_depth"]; ok {
			values = append(values, maxDepth)
		}
		if endVertex, ok := params["end_vertex"]; ok {
			values = append(values, endVertex)
		}
	}

	return values
}
