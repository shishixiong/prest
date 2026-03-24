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
- ⚠️ Uses PostgreSQL driver as fallback (no GaussDB driver integrated yet)
- ⚠️ Not fully tested (no test environment available)

## Configuration

To use GaussDB adapter, set the database type in your configuration:

### Configuration File (`prest.toml`)

```toml
[database]
type = "gaussdb"  # Options: "postgres" (default), "gaussdb"

[pg]
host = "localhost"
port = 5432
user = "gaussdb"
pass = "your_password"
database = "your_database"
ssl.mode = "disable"
```

### Environment Variables

```bash
export PREST_DATABASE_TYPE=gaussdb
export PREST_PG_HOST=localhost
export PREST_PG_PORT=5432
export PREST_PG_USER=gaussdb
export PREST_PG_PASS=your_password
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

Currently, the adapter uses the PostgreSQL driver as a fallback since the actual GaussDB Go driver is not yet integrated. This works because GaussDB 100 is highly compatible with PostgreSQL.

### Future Integration

When the GaussDB driver is available:

1. Import the GaussDB Go driver: `import "github.com/huaweicloud/gaussdb-go-driver"`
2. Update `connection.go` to use driver name `"gaussdb"` instead of `"postgres"`
3. Verify connection string format for GaussDB

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

1. **Driver Dependency**: Uses PostgreSQL driver instead of native GaussDB driver
2. **System Table Differences**: Assumes PostgreSQL-compatible system tables; may need adjustment
3. **Untested Features**: Core CRUD operations inherit from PostgreSQL but untested on GaussDB
4. **Distributed Features**: GaussDB distributed database features not yet supported
5. **Performance Optimizations**: No GaussDB-specific optimizations implemented

## Future Enhancements

1. **Native Driver Integration**: Integrate official GaussDB Go driver
2. **System Table Research**: Verify and update system table queries based on GaussDB documentation
3. **Distributed Database Support**: Add support for GaussDB distributed features
4. **Performance Optimizations**: Implement GaussDB-specific query optimizations
5. **Comprehensive Testing**: Add unit and integration tests with actual GaussDB instance
6. **Monitoring Integration**: Add GaussDB-specific monitoring and metrics

## Contributing

To contribute to the GaussDB adapter:

1. Review Huawei GaussDB documentation for system table differences
2. Test with actual GaussDB instance if available
3. Update SQL statements in `adapters/gaussdb/statements/queries.go`
4. Add tests in `adapters/gaussdb/gaussdb_test.go`
5. Update documentation with findings

## References

1. [Huawei GaussDB Documentation](https://support.huaweicloud.com/gaussdb/index.html)
2. [GaussDB 100 Compatibility Guide](https://support.huaweicloud.com/intl/en-us/productdesc-gaussdb/gaussdb_01_0010.html)
3. [PostgreSQL pREST Adapter](../adapters/postgres/) - Reference implementation

## License

Same as pREST project.