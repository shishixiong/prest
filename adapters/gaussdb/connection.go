package gaussdb

import (
	"fmt"
	"log/slog"

	// GaussDB driver for openGauss/opengauss
	_ "gitee.com/opengauss/openGauss-connector-go-pq"
	"github.com/jmoiron/sqlx"
	"github.com/prest/prest/v2/config"
)

// GetURI generates GaussDB connection string for openGauss driver
// The openGauss driver uses PostgreSQL-compatible connection string format
func GetURI(DBName string) string {
	cfg := config.PrestConf

	// openGauss driver uses PostgreSQL-compatible connection string format
	if cfg.PGURL != "" {
		// If PGURL is provided, use it
		return cfg.PGURL
	}

	if DBName == "" {
		DBName = cfg.PGDatabase
	}

	// Build connection string from components
	sslMode := cfg.PGSSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	// PostgreSQL-compatible format for openGauss driver
	connStr := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s",
		cfg.PGHost,
		cfg.PGPort,
		cfg.PGUser,
		DBName,
		sslMode)

	// Only add password if it's not empty
	if cfg.PGPass != "" {
		connStr += " password=" + cfg.PGPass
	}

	return connStr
}

// Get establishes a connection to GaussDB database using openGauss driver
func Get() (*sqlx.DB, error) {
	// Using openGauss driver for GaussDB connection
	// The openGauss driver registers itself as "postgres" driver name
	driverName := "postgres"

	dbName := GetDatabase()
	if dbName == "" {
		dbName = config.PrestConf.PGDatabase
		SetDatabase(dbName)
	}

	uri := GetURI(dbName)
	return sqlx.Connect(driverName, uri)
}

// SetDatabase sets the current database name
func SetDatabase(name string) {
	// TODO: Implement GaussDB-specific database context setting
	// For now, reuse PostgreSQL logic through embedded adapter
	// This would require access to internal connection package
	slog.Debug("SetDatabase called for GaussDB", "database", name)
	// Implementation depends on how connection is managed
}

// GetDatabase gets the current database name
func GetDatabase() string {
	// TODO: Implement GaussDB-specific database context getting
	slog.Debug("GetDatabase called for GaussDB")
	return config.PrestConf.PGDatabase
}

// Note: In a complete implementation, we would need to manage
// GaussDB-specific connection pooling, driver registration, etc.
