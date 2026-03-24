package statements

import "fmt"

const (
	// FieldDatabaseName define database name
	FieldDatabaseName = "datname"

	// FieldSchemaName define schema name
	FieldSchemaName = "schema_name"

	// FieldCountDatabaseName count by database name
	FieldCountDatabaseName = "COUNT(datname)"

	// FieldCountSchemaName count by schema name
	FieldCountSchemaName = "COUNT(schema_name)"

	// Databases list all data bases

	// DatabasesSelect clause
	// TODO: Verify GaussDB system table for databases
	// PostgreSQL uses pg_database, GaussDB might use gaussdb_database or similar
	DatabasesSelect = `
SELECT
	%s
FROM
	pg_database`

	// DatabasesWhere clause
	DatabasesWhere = `
WHERE
	NOT datistemplate`
	// DatabasesOrderBy clause
	DatabasesOrderBy = `
ORDER BY
	%s ASC`
	// Schemas list all schema on data base

	// SchemasSelect clause
	// TODO: Verify GaussDB system table for schemas
	// PostgreSQL uses information_schema.schemata
	SchemasSelect = `
SELECT
	%s
FROM
	information_schema.schemata`

	// SchemasGroupBy clause
	SchemasGroupBy = `
GROUP BY
	%s`

	// SchemasOrderBy clause
	SchemasOrderBy = `
ORDER BY
	%s ASC`

	// Tables list all tables

	// TablesSelect clause
	// TODO: Verify GaussDB system tables for tables
	// PostgreSQL uses pg_catalog.pg_class and pg_catalog.pg_namespace
	TablesSelect = `
SELECT
	n.nspname as "schema",
	c.relname as "name",
	CASE c.relkind
		WHEN 'r' THEN 'table'
		WHEN 'v' THEN 'view'
		WHEN 'm' THEN 'materialized_view'
		WHEN 'i' THEN 'index'
		WHEN 'S' THEN 'sequence'
		WHEN 's' THEN 'special'
		WHEN 'f' THEN 'foreign_table'
	END as "type",
	pg_catalog.pg_get_userbyid(c.relowner) as "owner"
FROM
	pg_catalog.pg_class c
LEFT JOIN
	pg_catalog.pg_namespace n ON n.oid = c.relnamespace `
	// TablesWhere clause
	TablesWhere = `
WHERE
	c.relkind IN ('r','v','m','S','s','') AND
	n.nspname !~ '^pg_toast' AND
	n.nspname NOT IN ('information_schema', 'pg_catalog') AND
	has_schema_privilege(n.nspname, 'USAGE') `
	// TablesOrderBy clause
	TablesOrderBy = `
ORDER BY 1, 2`
	// Tables default query
	Tables = TablesSelect + TablesWhere + TablesOrderBy
	// list all tables in schema and database

	// SchemaTablesSelect clause
	// TODO: Verify GaussDB system tables for schema tables
	// PostgreSQL uses pg_catalog.pg_tables and information_schema.schemata
	SchemaTablesSelect = `
SELECT
	t.tablename as "name",
	t.schemaname as "schema",
	sc.catalog_name as "database"
FROM
	pg_catalog.pg_tables t
INNER JOIN
	information_schema.schemata sc ON sc.schema_name = t.schemaname`

	// SchemaTablesWhere clause
	SchemaTablesWhere = `
WHERE
	sc.catalog_name = $1 AND
	t.schemaname = $2`

	// SchemaTablesOrderBy clause
	SchemaTablesOrderBy = `
ORDER BY
	t.tablename ASC`

	// SchemaTables default query
	SchemaTables = SchemaTablesSelect + SchemaTablesWhere + SchemaTablesOrderBy

	// SelectInTable default query
	SelectInTable = `
SELECT
	*
FROM`

	// InsertQuery query
	InsertQuery = `INSERT INTO "%s"."%s"."%s"(%s) VALUES%s`

	// DeleteQuery query
	DeleteQuery = `DELETE FROM "%s"."%s"."%s"`

	// UpdateQuery query
	UpdateQuery = `UPDATE "%s"."%s"."%s" SET %s`

	// GroupBy query
	GroupBy = `GROUP BY %s`

	// Having query
	Having = `HAVING %s %s %s`

	// ShowTableQuery query for table structure
	// TODO: Verify GaussDB information_schema.columns compatibility
	ShowTableQuery = `SELECT table_schema, table_name, ordinal_position as position, column_name,data_type,
			  	CASE WHEN character_maximum_length is not null
					THEN character_maximum_length
					ELSE numeric_precision end as max_length,
			  	is_nullable,
			  	is_generated,
			  	is_updatable,
			  	column_default as default_value
			 FROM information_schema.columns
			 WHERE table_name=$1 AND table_schema=$2
			 ORDER BY table_schema, table_name, ordinal_position`
)

var (
	// Databases default query
	Databases = fmt.Sprintf(DatabasesSelect, FieldDatabaseName) + DatabasesWhere + fmt.Sprintf(DatabasesOrderBy, FieldDatabaseName)

	// Schemas default query
	Schemas = fmt.Sprintf(SchemasSelect, FieldSchemaName) + fmt.Sprintf(SchemasOrderBy, FieldSchemaName)
)