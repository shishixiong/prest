package config

// VectorConfig configuration for vector search functionality
type VectorConfig struct {
	Enabled         bool   `toml:"enabled" mapstructure:"enabled"`
	DefaultDistance string `toml:"default_distance" mapstructure:"default_distance"` // cosine/l2/inner_product
	DefaultDimensions int  `toml:"default_dimensions" mapstructure:"default_dimensions"`
	IndexType       string `toml:"index_type" mapstructure:"index_type"`        // hnsw/ivfflat
	MaxConnections  int    `toml:"max_connections" mapstructure:"max_connections"`   // HNSW parameter
	EFConstruction  int    `toml:"ef_construction" mapstructure:"ef_construction"`   // HNSW parameter
	DefaultVectorField string `toml:"default_vector_field" mapstructure:"default_vector_field"` // Default vector field name
}

// DefaultVectorConfig returns default vector configuration
func DefaultVectorConfig() VectorConfig {
	return VectorConfig{
		Enabled:         false, // Disabled by default
		DefaultDistance: "cosine",
		DefaultDimensions: 768,
		IndexType:       "hnsw",
		MaxConnections:  16,
		EFConstruction:  64,
		DefaultVectorField: "vector",
	}
}

// IsValidDistanceMetric checks if distance metric is valid
func (vc *VectorConfig) IsValidDistanceMetric(metric string) bool {
	switch metric {
	case "cosine", "l2", "euclidean", "inner", "inner_product", "ip":
		return true
	default:
		return false
	}
}

// GetDistanceOperator returns PostgreSQL operator for distance metric
func (vc *VectorConfig) GetDistanceOperator(metric string) string {
	switch metric {
	case "l2", "euclidean":
		return "<->"
	case "cosine":
		return "<=>"
	case "inner", "inner_product", "ip":
		return "<#>"
	default:
		return vc.GetDistanceOperator(vc.DefaultDistance)
	}
}