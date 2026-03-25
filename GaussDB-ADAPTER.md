# GaussDB Adapter for pREST

This adapter adds support for Huawei GaussDB database to pREST.

## Overview

GaussDB is Huawei's distributed database based on PostgreSQL. GaussDB 100 is highly compatible with PostgreSQL, allowing us to reuse much of the PostgreSQL adapter implementation.

## Current Status

**Beta - Development Phase**

The adapter is currently in development with the following status:

- ✅ Basic adapter structure implemented
- ✅ Configuration-driven database type selection
- ✅ System table queries overridden for GaussDB compatibility
- ✅ openGauss driver integrated (`gitee.com/opengauss/openGauss-connector-go-pq`)
- ⚠️ Not fully tested (limited test environment available)

## Configuration

To use GaussDB adapter, set the database type in your configuration:

### Configuration File (`prest.toml`)

```toml
[database]
type = "gaussdb"  # Options: "postgres" (default), "gaussdb"

[pg]
host = "localhost"
port = 18888  # Default GaussDB port is 5432, local test instance uses 18888
user = "gaussdb"
pass = "your_password"
database = "your_database"
ssl.mode = "disable"
```

### Environment Variables

```bash
export PREST_DATABASE_TYPE=gaussdb
export PREST_PG_HOST=localhost
export PREST_PG_PORT=18888  # Default GaussDB port is 5432, local test instance uses 18888
export PREST_PG_USER=gaussdb
export PREST_PG_PASS=Ssx@1234  # Example password for local test instance
export PREST_PG_DATABASE=your_database
```

## Architecture

### Inheritance-Based Design

The GaussDB adapter embeds the PostgreSQL adapter and overrides only the methods that need changes:

```go
type GaussDB struct {
    postgres.Postgres
}
```

### Overridden Methods

The following methods are overridden to handle GaussDB-specific system table differences:

1. `DatabaseClause()` - Database listing query
2. `SchemaClause()` - Schema listing query
3. `TableClause()` - Table listing query
4. `ShowTable()` - Table structure query
5. `ShowTableCtx()` - Table structure query with context
6. `SchemaTablesClause()` - Schema tables query

### SQL Statements

GaussDB-specific SQL statements are defined in `adapters/gaussdb/statements/queries.go`. Currently, these use PostgreSQL-compatible queries with TODO comments for GaussDB-specific adjustments.

## Connection Management

### Current Implementation

The adapter uses the openGauss driver (`gitee.com/opengauss/openGauss-connector-go-pq`) which is a PostgreSQL-compatible driver for GaussDB/openGauss databases. The driver registers itself as `"postgres"` driver name and uses PostgreSQL connection string format.

### Driver Integration

1. Import the openGauss driver: `import _ "gitee.com/opengauss/openGauss-connector-go-pq"`
2. The driver name is `"postgres"` (same as PostgreSQL driver)
3. Connection string format is PostgreSQL-compatible: `host=... port=... user=... password=... dbname=... sslmode=...`

### Local Test Instance

For local testing with the provided GaussDB instance:
- Host: `localhost`
- Port: `18888`
- User: `gaussdb`
- Password: `Ssx@1234`
- SSL mode: `disable` (for local testing)

## Testing

### Testing Limitations

Currently, there is no test environment available for GaussDB. Testing is based on:

1. Documentation review of GaussDB system tables
2. Assumption of PostgreSQL compatibility for GaussDB 100
3. Compilation and basic initialization tests

### Test Strategy

When a test environment becomes available:

1. **Unit Tests**: Test individual adapter methods
2. **Integration Tests**: Test actual database connections and queries
3. **Compatibility Tests**: Verify PostgreSQL feature compatibility
4. **Performance Tests**: Compare performance with PostgreSQL adapter

## Known Issues and Limitations

1. **Driver Compatibility**: Uses openGauss driver which registers as `"postgres"` driver name, potentially conflicting with PostgreSQL driver if both are imported
2. **System Table Differences**: Assumes PostgreSQL-compatible system tables; may need adjustment based on actual GaussDB system table structure
3. **Limited Testing**: Core CRUD operations inherit from PostgreSQL but limited testing on actual GaussDB
4. **Distributed Features**: GaussDB distributed database features not yet supported
5. **Performance Optimizations**: No GaussDB-specific optimizations implemented

## Future Enhancements

1. **System Table Research**: Verify and update system table queries based on actual GaussDB system table structure
2. **Distributed Database Support**: Add support for GaussDB distributed features
3. **Performance Optimizations**: Implement GaussDB-specific query optimizations
4. **Comprehensive Testing**: Add unit and integration tests with actual GaussDB instance
5. **Monitoring Integration**: Add GaussDB-specific monitoring and metrics
6. **Driver Improvements**: Consider using native GaussDB driver if available, or contribute to openGauss driver for better GaussDB support

## Contributing

To contribute to the GaussDB adapter:

1. Test with actual GaussDB instance and verify system table differences
2. Update SQL statements in `adapters/gaussdb/statements/queries.go` based on actual GaussDB system tables
3. Add tests in `adapters/gaussdb/gaussdb_test.go`
4. Test connection with local GaussDB instance (port 18888, user gaussdb, password Ssx@1234)
5. Update documentation with findings

## References

1. [Huawei GaussDB Documentation](https://support.huaweicloud.com/gaussdb/index.html)
2. [GaussDB 100 Compatibility Guide](https://support.huaweicloud.com/intl/en-us/productdesc-gaussdb/gaussdb_01_0010.html)
3. [openGauss Go Driver](https://gitee.com/opengauss/openGauss-connector-go-pq) - PostgreSQL-compatible driver for GaussDB/openGauss
4. [PostgreSQL pREST Adapter](../adapters/postgres/) - Reference implementation

## License

Same as pREST project.