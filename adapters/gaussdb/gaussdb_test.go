package gaussdb

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/prest/prest/v2/adapters"
	"github.com/prest/prest/v2/adapters/gaussdb/statements"
	"github.com/prest/prest/v2/config"
	"github.com/stretchr/testify/assert"
)

func init() {
	// Load config for tests
	config.Load()
}

func TestGaussDB_Load(t *testing.T) {
	// Test that Load initializes the adapter correctly
	// Since we're testing without actual database connection,
	// we'll just verify the adapter is set
	originalAdapter := config.PrestConf.Adapter

	Load()

	assert.NotNil(t, config.PrestConf.Adapter)
	assert.IsType(t, &GaussDB{}, config.PrestConf.Adapter)

	// Restore original adapter to avoid affecting other tests
	config.PrestConf.Adapter = originalAdapter
}

func TestGaussDB_Load_MultipleCalls(t *testing.T) {
	// Test that multiple calls to Load() work correctly
	originalAdapter := config.PrestConf.Adapter

	// First call
	Load()
	assert.NotNil(t, config.PrestConf.Adapter)
	assert.IsType(t, &GaussDB{}, config.PrestConf.Adapter)
	_ = config.PrestConf.Adapter // adapter1

	// Second call
	Load()
	assert.NotNil(t, config.PrestConf.Adapter)
	assert.IsType(t, &GaussDB{}, config.PrestConf.Adapter)
	_ = config.PrestConf.Adapter // adapter2

	// Each call creates a new instance
	// They might be different instances or the same, depends on implementation
	// Just verify it doesn't panic

	// Restore original adapter
	config.PrestConf.Adapter = originalAdapter
}

func TestGaussDB_DatabaseClause(t *testing.T) {
	adapter := &GaussDB{}

	// Test without _count parameter
	req1 := httptest.NewRequest("GET", "/databases", nil)
	query1, hasCount1 := adapter.DatabaseClause(req1)

	assert.Contains(t, query1, "pg_database")
	assert.Contains(t, query1, statements.FieldDatabaseName)
	assert.False(t, hasCount1)

	// Test with _count parameter
	req2 := httptest.NewRequest("GET", "/databases?_count=*", nil)
	query2, hasCount2 := adapter.DatabaseClause(req2)

	assert.Contains(t, query2, "pg_database")
	assert.Contains(t, query2, statements.FieldCountDatabaseName)
	assert.True(t, hasCount2)

	// Test with _count parameter as empty string (should not trigger count)
	req3 := httptest.NewRequest("GET", "/databases?_count=", nil)
	query3, hasCount3 := adapter.DatabaseClause(req3)

	assert.Contains(t, query3, "pg_database")
	assert.Contains(t, query3, statements.FieldDatabaseName)
	assert.False(t, hasCount3)

	// Test with additional query parameters (should be ignored)
	req4 := httptest.NewRequest("GET", "/databases?other=param&_count=*", nil)
	query4, hasCount4 := adapter.DatabaseClause(req4)

	assert.Contains(t, query4, "pg_database")
	assert.Contains(t, query4, statements.FieldCountDatabaseName)
	assert.True(t, hasCount4)
}

func TestGaussDB_DatabaseClause_NilRequest(t *testing.T) {
	adapter := &GaussDB{}

	// Test with nil request - should panic
	assert.Panics(t, func() {
		adapter.DatabaseClause(nil)
	})
}

func TestGaussDB_SchemaClause(t *testing.T) {
	adapter := &GaussDB{}

	// Test without _count parameter
	req1 := httptest.NewRequest("GET", "/schemas", nil)
	query1, hasCount1 := adapter.SchemaClause(req1)

	assert.Contains(t, query1, "information_schema.schemata")
	assert.Contains(t, query1, statements.FieldSchemaName)
	assert.False(t, hasCount1)

	// Test with _count parameter
	req2 := httptest.NewRequest("GET", "/schemas?_count=*", nil)
	query2, hasCount2 := adapter.SchemaClause(req2)

	assert.Contains(t, query2, "information_schema.schemata")
	assert.Contains(t, query2, statements.FieldCountSchemaName)
	assert.True(t, hasCount2)

	// Test with _count parameter as empty string (should not trigger count)
	req3 := httptest.NewRequest("GET", "/schemas?_count=", nil)
	query3, hasCount3 := adapter.SchemaClause(req3)

	assert.Contains(t, query3, "information_schema.schemata")
	assert.Contains(t, query3, statements.FieldSchemaName)
	assert.False(t, hasCount3)

	// Test with multiple query parameters
	req4 := httptest.NewRequest("GET", "/schemas?limit=10&_count=*&offset=0", nil)
	query4, hasCount4 := adapter.SchemaClause(req4)

	assert.Contains(t, query4, "information_schema.schemata")
	assert.Contains(t, query4, statements.FieldCountSchemaName)
	assert.True(t, hasCount4)
}

func TestGaussDB_SchemaClause_NilRequest(t *testing.T) {
	adapter := &GaussDB{}

	// Test with nil request - should panic
	assert.Panics(t, func() {
		adapter.SchemaClause(nil)
	})
}

func TestGaussDB_TableClause(t *testing.T) {
	adapter := &GaussDB{}

	query := adapter.TableClause()

	// Should return the TablesSelect statement
	assert.Contains(t, query, "pg_catalog.pg_class")
	assert.Contains(t, query, "pg_catalog.pg_namespace")
	assert.Contains(t, query, "n.nspname as \"schema\"")
	assert.Contains(t, query, "c.relname as \"name\"")
}

func TestGaussDB_SchemaTablesClause(t *testing.T) {
	adapter := &GaussDB{}

	query := adapter.SchemaTablesClause()

	// Should return the SchemaTablesSelect statement
	assert.Contains(t, query, "pg_catalog.pg_tables")
	assert.Contains(t, query, "information_schema.schemata")
	assert.Contains(t, query, "t.tablename as \"name\"")
	assert.Contains(t, query, "t.schemaname as \"schema\"")
	assert.Contains(t, query, "sc.catalog_name as \"database\"")
}

func TestGaussDB_ShowTable(t *testing.T) {
	adapter := &GaussDB{}

	// This test verifies that ShowTable returns a scanner
	// Since we can't actually query the database in unit tests,
	// we just verify the method exists and returns something
	scanner := adapter.ShowTable("public", "test_table")

	assert.NotNil(t, scanner)
	// The scanner will be nil if there's an error, but for unit tests
	// we're just checking the method signature works
}

func TestGaussDB_ShowTableCtx(t *testing.T) {
	adapter := &GaussDB{}
	ctx := context.Background()

	// This test verifies that ShowTableCtx returns a scanner with context
	scanner := adapter.ShowTableCtx(ctx, "public", "test_table")

	assert.NotNil(t, scanner)
	// The scanner will be nil if there's an error, but for unit tests
	// we're just checking the method signature works
}

func TestGaussDB_AdapterInterface(t *testing.T) {
	// Verify that GaussDB implements the required adapter interface
	// This is a compile-time check, but we can verify at runtime
	var _ adapters.Adapter = &GaussDB{}

	adapter := &GaussDB{}
	assert.NotNil(t, adapter)

	// The adapter should have all methods from the embedded Postgres adapter
	// We can check that it can be assigned to the interface type
	var adapterInterface adapters.Adapter = adapter
	assert.NotNil(t, adapterInterface)
}

func TestGaussDB_ConnectionFunctions(t *testing.T) {
	// Test connection-related functions
	// These functions require actual database connection, so we'll
	// just test their basic functionality without connecting

	// Test GetDatabase returns default from config
	config.PrestConf.PGDatabase = "testdb"
	dbName := GetDatabase()
	assert.Equal(t, "testdb", dbName)

	// Test SetDatabase logs appropriately (we can't verify much without actual connection)
	SetDatabase("newdb")

	// Test GetURI generates correct connection string
	uri := GetURI("testdb")
	assert.Contains(t, uri, "dbname=testdb")
	assert.Contains(t, uri, "sslmode=disable") // default ssl mode
}

func TestGetURI_WithDifferentConfigurations(t *testing.T) {
	// Save original config values
	originalHost := config.PrestConf.PGHost
	originalPort := config.PrestConf.PGPort
	originalUser := config.PrestConf.PGUser
	originalPass := config.PrestConf.PGPass
	originalSSLMode := config.PrestConf.PGSSLMode
	originalURL := config.PrestConf.PGURL
	originalDatabase := config.PrestConf.PGDatabase

	// Restore original values after test
	defer func() {
		config.PrestConf.PGHost = originalHost
		config.PrestConf.PGPort = originalPort
		config.PrestConf.PGUser = originalUser
		config.PrestConf.PGPass = originalPass
		config.PrestConf.PGSSLMode = originalSSLMode
		config.PrestConf.PGURL = originalURL
		config.PrestConf.PGDatabase = originalDatabase
	}()

	// Test 1: Basic configuration
	config.PrestConf.PGHost = "localhost"
	config.PrestConf.PGPort = 18888
	config.PrestConf.PGUser = "gaussdb"
	config.PrestConf.PGPass = "Ssx@1234"
	config.PrestConf.PGSSLMode = "disable"
	config.PrestConf.PGURL = ""
	config.PrestConf.PGDatabase = "defaultdb"

	uri := GetURI("mydb")
	expected := "host=localhost port=18888 user=gaussdb dbname=mydb sslmode=disable password=Ssx@1234"
	assert.Equal(t, expected, uri)

	// Test 2: With SSL enabled
	config.PrestConf.PGSSLMode = "require"
	uri2 := GetURI("mydb")
	assert.Contains(t, uri2, "sslmode=require")

	// Test 3: With empty password
	config.PrestConf.PGPass = ""
	uri3 := GetURI("mydb")
	assert.NotContains(t, uri3, "password=")

	// Test 4: With PGURL set (should return PGURL directly)
	config.PrestConf.PGURL = "postgres://gaussdb:Ssx@1234@localhost:18888/mydb?sslmode=disable"
	uri4 := GetURI("otherdb")
	assert.Equal(t, config.PrestConf.PGURL, uri4)

	// Test 5: Reset PGURL and test with empty database name
	config.PrestConf.PGURL = ""
	uri5 := GetURI("")
	assert.Contains(t, uri5, "dbname=defaultdb")
}

func TestGetURI_EdgeCases(t *testing.T) {
	// Save original config values
	originalHost := config.PrestConf.PGHost
	originalPort := config.PrestConf.PGPort
	originalUser := config.PrestConf.PGUser
	originalPass := config.PrestConf.PGPass
	originalSSLMode := config.PrestConf.PGSSLMode
	originalURL := config.PrestConf.PGURL
	originalDatabase := config.PrestConf.PGDatabase

	// Restore original values after test
	defer func() {
		config.PrestConf.PGHost = originalHost
		config.PrestConf.PGPort = originalPort
		config.PrestConf.PGUser = originalUser
		config.PrestConf.PGPass = originalPass
		config.PrestConf.PGSSLMode = originalSSLMode
		config.PrestConf.PGURL = originalURL
		config.PrestConf.PGDatabase = originalDatabase
	}()

	// Test with empty host (default is "127.0.0.1" from config defaults)
	config.PrestConf.PGHost = ""
	config.PrestConf.PGPort = 5432
	config.PrestConf.PGUser = "testuser"
	config.PrestConf.PGPass = ""
	config.PrestConf.PGSSLMode = "disable"
	config.PrestConf.PGURL = ""
	config.PrestConf.PGDatabase = "testdb"

	uri := GetURI("mydb")
	assert.Contains(t, uri, "host=")
	// Empty host should still be included in connection string

	// Test with port 0 (unlikely but possible)
	config.PrestConf.PGPort = 0
	uri2 := GetURI("mydb")
	assert.Contains(t, uri2, "port=0")

	// Test with special characters in password
	config.PrestConf.PGPass = "pass@word#123"
	config.PrestConf.PGPort = 5432
	uri3 := GetURI("mydb")
	assert.Contains(t, uri3, "password=pass@word#123")

	// Test with SSL mode "verify-full"
	config.PrestConf.PGSSLMode = "verify-full"
	uri4 := GetURI("mydb")
	assert.Contains(t, uri4, "sslmode=verify-full")

	// Test that PGURL takes precedence even with other config
	config.PrestConf.PGURL = "postgres://user:pass@example.com:5432/db"
	config.PrestConf.PGHost = "ignored"
	config.PrestConf.PGPort = 9999
	config.PrestConf.PGUser = "ignored"
	uri5 := GetURI("ignored")
	assert.Equal(t, "postgres://user:pass@example.com:5432/db", uri5)
}

func TestGaussDB_StatementsPackage(t *testing.T) {
	// Verify that the statements package constants are accessible
	assert.NotEmpty(t, statements.FieldDatabaseName)
	assert.NotEmpty(t, statements.FieldSchemaName)
	assert.NotEmpty(t, statements.FieldCountDatabaseName)
	assert.NotEmpty(t, statements.FieldCountSchemaName)
	assert.NotEmpty(t, statements.DatabasesSelect)
	assert.NotEmpty(t, statements.SchemasSelect)
	assert.NotEmpty(t, statements.TablesSelect)
	assert.NotEmpty(t, statements.SchemaTablesSelect)
	assert.NotEmpty(t, statements.ShowTableQuery)
}

func TestSetDatabaseAndGetDatabase(t *testing.T) {
	// Save original database name
	originalDB := config.PrestConf.PGDatabase

	// Test GetDatabase returns configured database
	config.PrestConf.PGDatabase = "testdb"
	assert.Equal(t, "testdb", GetDatabase())

	// Test SetDatabase doesn't panic (just logs)
	SetDatabase("newdb")

	// Verify GetDatabase still returns configured database (not changed by SetDatabase)
	assert.Equal(t, "testdb", GetDatabase())

	// Restore original
	config.PrestConf.PGDatabase = originalDB
}
