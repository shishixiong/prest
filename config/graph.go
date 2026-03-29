package config

// GraphConfig configuration for graph data functionality
type GraphConfig struct {
	Enabled              bool   `toml:"enabled" mapstructure:"enabled"`
	DefaultGraphType     string `toml:"default_graph_type" mapstructure:"default_graph_type"`         // property/directed/undirected
	DefaultVertexLabel   string `toml:"default_vertex_label" mapstructure:"default_vertex_label"`     // Default vertex label
	DefaultEdgeLabel     string `toml:"default_edge_label" mapstructure:"default_edge_label"`         // Default edge label
	MaxTraversalDepth    int    `toml:"max_traversal_depth" mapstructure:"max_traversal_depth"`       // Maximum traversal depth for graph queries
	MaxResults           int    `toml:"max_results" mapstructure:"max_results"`                       // Maximum results per graph query
	EnableGraphIndices   bool   `toml:"enable_graph_indices" mapstructure:"enable_graph_indices"`     // Enable automatic graph indices
	DefaultVertexTable   string `toml:"default_vertex_table" mapstructure:"default_vertex_table"`     // Default vertex table name pattern
	DefaultEdgeTable     string `toml:"default_edge_table" mapstructure:"default_edge_table"`         // Default edge table name pattern
	QueryTimeout         int    `toml:"query_timeout" mapstructure:"query_timeout"`                   // Query timeout in seconds
	BatchSize            int    `toml:"batch_size" mapstructure:"batch_size"`                         // Batch size for bulk operations
}

// DefaultGraphConfig returns default graph configuration
func DefaultGraphConfig() GraphConfig {
	return GraphConfig{
		Enabled:            false, // Disabled by default
		DefaultGraphType:   "property",
		DefaultVertexLabel: "vertex",
		DefaultEdgeLabel:   "edge",
		MaxTraversalDepth:  10,
		MaxResults:         1000,
		EnableGraphIndices: true,
		DefaultVertexTable: "_graph_vertices",
		DefaultEdgeTable:   "_graph_edges",
		QueryTimeout:       30,
		BatchSize:          100,
	}
}

// IsValidGraphType checks if graph type is valid
func (gc *GraphConfig) IsValidGraphType(graphType string) bool {
	switch graphType {
	case "property", "directed", "undirected":
		return true
	default:
		return false
	}
}

// IsValidTraversalDirection checks if traversal direction is valid
func (gc *GraphConfig) IsValidTraversalDirection(direction string) bool {
	switch direction {
	case "out", "in", "both":
		return true
	default:
		return false
	}
}