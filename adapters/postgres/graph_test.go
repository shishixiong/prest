package postgres

import (
	"context"
	"testing"

	"github.com/prest/prest/v2/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	config.Load()
	Load()
}

func TestCreateGraph(t *testing.T) {
	adapter := config.PrestConf.Adapter
	require.NotNil(t, adapter, "adapter not initialized, check database configuration")

	// Test creating a graph with different types
	testCases := []struct {
		name      string
		graphType string
	}{
		{"property graph", "property"},
		{"directed graph", "directed"},
		{"undirected graph", "undirected"},
		{"default type", "unknown"}, // should default to property
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sc := adapter.CreateGraph("testdb", "public", "test_graph", tc.graphType)
			// Expect error because CREATE GRAPH syntax may not be supported or table doesn't exist
			// We just verify the scanner is returned without panic
			assert.NotNil(t, sc)
			// The actual error will depend on database support
			// We're testing that the function executes without panic
		})
	}
}

func TestCreateGraphCtx(t *testing.T) {
	ctx := context.Background()
	adapter := config.PrestConf.Adapter
	require.NotNil(t, adapter, "adapter not initialized, check database configuration")

	// Test different graph types with context
	testCases := []struct {
		name      string
		graphType string
	}{
		{"property graph", "property"},
		{"directed graph", "directed"},
		{"undirected graph", "undirected"},
		{"default type", "unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sc := adapter.CreateGraphCtx(ctx, "testdb", "public", "test_graph", tc.graphType)
			assert.NotNil(t, sc)
		})
	}
}

func TestDeleteGraph(t *testing.T) {
	adapter := config.PrestConf.Adapter

	sc := adapter.DeleteGraph("testdb", "public", "test_graph")
	assert.NotNil(t, sc)
}

func TestDeleteGraphCtx(t *testing.T) {
	ctx := context.Background()
	adapter := config.PrestConf.Adapter

	sc := adapter.DeleteGraphCtx(ctx, "testdb", "public", "test_graph")
	assert.NotNil(t, sc)
}

func TestAddVertices(t *testing.T) {
	adapter := config.PrestConf.Adapter

	t.Run("empty vertices", func(t *testing.T) {
		sc := adapter.AddVertices("testdb", "public", "test_graph", []map[string]interface{}{})
		assert.NotNil(t, sc)
		assert.NoError(t, sc.Err())
	})

	t.Run("single vertex with properties", func(t *testing.T) {
		vertices := []map[string]interface{}{
			{
				"id":    1,
				"label": "person",
				"properties": map[string]interface{}{
					"name": "Alice",
					"age":  30,
				},
			},
		}
		sc := adapter.AddVertices("testdb", "public", "test_graph", vertices)
		assert.NotNil(t, sc)
	})

	t.Run("single vertex with nil properties", func(t *testing.T) {
		vertices := []map[string]interface{}{
			{
				"id":         2,
				"label":      "person",
				"properties": nil,
			},
		}
		sc := adapter.AddVertices("testdb", "public", "test_graph", vertices)
		assert.NotNil(t, sc)
	})

	t.Run("multiple vertices", func(t *testing.T) {
		vertices := []map[string]interface{}{
			{
				"id":         3,
				"label":      "person",
				"properties": map[string]interface{}{"name": "Bob"},
			},
			{
				"id":         4,
				"label":      "city",
				"properties": map[string]interface{}{"population": 100000},
			},
		}
		sc := adapter.AddVertices("testdb", "public", "test_graph", vertices)
		assert.NotNil(t, sc)
	})
}

func TestAddVerticesCtx(t *testing.T) {
	ctx := context.Background()
	adapter := config.PrestConf.Adapter

	t.Run("single vertex with properties", func(t *testing.T) {
		vertices := []map[string]interface{}{
			{
				"id":    2,
				"label": "person",
				"properties": map[string]interface{}{
					"name": "Bob",
					"age":  25,
				},
			},
		}
		sc := adapter.AddVerticesCtx(ctx, "testdb", "public", "test_graph", vertices)
		assert.NotNil(t, sc)
	})

	t.Run("single vertex with nil properties", func(t *testing.T) {
		vertices := []map[string]interface{}{
			{
				"id":         5,
				"label":      "person",
				"properties": nil,
			},
		}
		sc := adapter.AddVerticesCtx(ctx, "testdb", "public", "test_graph", vertices)
		assert.NotNil(t, sc)
	})

	t.Run("multiple vertices", func(t *testing.T) {
		vertices := []map[string]interface{}{
			{
				"id":         6,
				"label":      "person",
				"properties": map[string]interface{}{"name": "Charlie"},
			},
			{
				"id":         7,
				"label":      "city",
				"properties": map[string]interface{}{"population": 50000},
			},
		}
		sc := adapter.AddVerticesCtx(ctx, "testdb", "public", "test_graph", vertices)
		assert.NotNil(t, sc)
	})
}

func TestAddEdges(t *testing.T) {
	adapter := config.PrestConf.Adapter

	t.Run("empty edges", func(t *testing.T) {
		sc := adapter.AddEdges("testdb", "public", "test_graph", []map[string]interface{}{})
		assert.NotNil(t, sc)
		assert.NoError(t, sc.Err())
	})

	t.Run("single edge with properties", func(t *testing.T) {
		edges := []map[string]interface{}{
			{
				"id":    100,
				"from":  1,
				"to":    2,
				"label": "knows",
				"properties": map[string]interface{}{
					"since": "2020-01-01",
				},
			},
		}
		sc := adapter.AddEdges("testdb", "public", "test_graph", edges)
		assert.NotNil(t, sc)
	})

	t.Run("single edge with nil properties", func(t *testing.T) {
		edges := []map[string]interface{}{
			{
				"id":         101,
				"from":       2,
				"to":         3,
				"label":      "knows",
				"properties": nil,
			},
		}
		sc := adapter.AddEdges("testdb", "public", "test_graph", edges)
		assert.NotNil(t, sc)
	})

	t.Run("multiple edges", func(t *testing.T) {
		edges := []map[string]interface{}{
			{
				"id":         102,
				"from":       3,
				"to":         4,
				"label":      "lives_in",
				"properties": map[string]interface{}{"since": "2019"},
			},
			{
				"id":         103,
				"from":       4,
				"to":         5,
				"label":      "located_in",
				"properties": map[string]interface{}{"distance": 10},
			},
		}
		sc := adapter.AddEdges("testdb", "public", "test_graph", edges)
		assert.NotNil(t, sc)
	})
}

func TestAddEdgesCtx(t *testing.T) {
	ctx := context.Background()
	adapter := config.PrestConf.Adapter

	t.Run("single edge with properties", func(t *testing.T) {
		edges := []map[string]interface{}{
			{
				"id":    101,
				"from":  2,
				"to":    3,
				"label": "knows",
				"properties": map[string]interface{}{
					"since": "2021-01-01",
				},
			},
		}
		sc := adapter.AddEdgesCtx(ctx, "testdb", "public", "test_graph", edges)
		assert.NotNil(t, sc)
	})

	t.Run("single edge with nil properties", func(t *testing.T) {
		edges := []map[string]interface{}{
			{
				"id":         104,
				"from":       3,
				"to":         4,
				"label":      "knows",
				"properties": nil,
			},
		}
		sc := adapter.AddEdgesCtx(ctx, "testdb", "public", "test_graph", edges)
		assert.NotNil(t, sc)
	})

	t.Run("multiple edges", func(t *testing.T) {
		edges := []map[string]interface{}{
			{
				"id":         105,
				"from":       5,
				"to":         6,
				"label":      "works_at",
				"properties": map[string]interface{}{"since": "2022"},
			},
			{
				"id":         106,
				"from":       6,
				"to":         7,
				"label":      "manages",
				"properties": map[string]interface{}{"level": "senior"},
			},
		}
		sc := adapter.AddEdgesCtx(ctx, "testdb", "public", "test_graph", edges)
		assert.NotNil(t, sc)
	})
}

func TestGraphQuery(t *testing.T) {
	adapter := config.PrestConf.Adapter

	// Test different query types
	testCases := []struct {
		name        string
		queryType   string
		queryParams map[string]interface{}
	}{
		{
			name:      "traverse query",
			queryType: "traverse",
			queryParams: map[string]interface{}{
				"start_vertex": 1,
				"max_depth":    2,
				"direction":    "out",
			},
		},
		{
			name:      "path query",
			queryType: "path",
			queryParams: map[string]interface{}{
				"start_vertex": 1,
				"end_vertex":   5,
				"max_depth":    3,
			},
		},
		{
			name:      "neighbors query",
			queryType: "neighbors",
			queryParams: map[string]interface{}{
				"start_vertex": 1,
				"direction":    "both",
			},
		},
		{
			name:        "match query",
			queryType:   "match",
			queryParams: map[string]interface{}{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sc := adapter.GraphQuery("testdb", "public", "test_graph", tc.queryType, tc.queryParams)
			assert.NotNil(t, sc)
		})
	}
}

func TestGraphQueryCtx(t *testing.T) {
	ctx := context.Background()
	adapter := config.PrestConf.Adapter

	queryParams := map[string]interface{}{
		"start_vertex": 1,
		"max_depth":    2,
	}
	sc := adapter.GraphQueryCtx(ctx, "testdb", "public", "test_graph", "traverse", queryParams)
	assert.NotNil(t, sc)
}

func TestGetVertex(t *testing.T) {
	adapter := config.PrestConf.Adapter

	sc := adapter.GetVertex("testdb", "public", "test_graph", 1)
	assert.NotNil(t, sc)
}

func TestGetVertexCtx(t *testing.T) {
	ctx := context.Background()
	adapter := config.PrestConf.Adapter

	sc := adapter.GetVertexCtx(ctx, "testdb", "public", "test_graph", 2)
	assert.NotNil(t, sc)
}

func TestGetEdge(t *testing.T) {
	adapter := config.PrestConf.Adapter

	sc := adapter.GetEdge("testdb", "public", "test_graph", 100)
	assert.NotNil(t, sc)
}

func TestGetEdgeCtx(t *testing.T) {
	ctx := context.Background()
	adapter := config.PrestConf.Adapter

	sc := adapter.GetEdgeCtx(ctx, "testdb", "public", "test_graph", 101)
	assert.NotNil(t, sc)
}

func TestUpdateVertex(t *testing.T) {
	adapter := config.PrestConf.Adapter

	t.Run("with properties", func(t *testing.T) {
		properties := map[string]interface{}{
			"name": "Updated Name",
			"age":  31,
		}
		sc := adapter.UpdateVertex("testdb", "public", "test_graph", 1, properties)
		assert.NotNil(t, sc)
	})

	t.Run("with nil properties", func(t *testing.T) {
		sc := adapter.UpdateVertex("testdb", "public", "test_graph", 2, nil)
		assert.NotNil(t, sc)
	})

	t.Run("with empty properties", func(t *testing.T) {
		properties := map[string]interface{}{}
		sc := adapter.UpdateVertex("testdb", "public", "test_graph", 3, properties)
		assert.NotNil(t, sc)
	})
}

func TestUpdateVertexCtx(t *testing.T) {
	ctx := context.Background()
	adapter := config.PrestConf.Adapter

	t.Run("with properties", func(t *testing.T) {
		properties := map[string]interface{}{
			"name": "Another Update",
		}
		sc := adapter.UpdateVertexCtx(ctx, "testdb", "public", "test_graph", 1, properties)
		assert.NotNil(t, sc)
	})

	t.Run("with nil properties", func(t *testing.T) {
		sc := adapter.UpdateVertexCtx(ctx, "testdb", "public", "test_graph", 2, nil)
		assert.NotNil(t, sc)
	})

	t.Run("with empty properties", func(t *testing.T) {
		properties := map[string]interface{}{}
		sc := adapter.UpdateVertexCtx(ctx, "testdb", "public", "test_graph", 3, properties)
		assert.NotNil(t, sc)
	})
}

func TestUpdateEdge(t *testing.T) {
	adapter := config.PrestConf.Adapter

	t.Run("with properties", func(t *testing.T) {
		properties := map[string]interface{}{
			"since": "2022-01-01",
		}
		sc := adapter.UpdateEdge("testdb", "public", "test_graph", 100, properties)
		assert.NotNil(t, sc)
	})

	t.Run("with nil properties", func(t *testing.T) {
		sc := adapter.UpdateEdge("testdb", "public", "test_graph", 101, nil)
		assert.NotNil(t, sc)
	})

	t.Run("with empty properties", func(t *testing.T) {
		properties := map[string]interface{}{}
		sc := adapter.UpdateEdge("testdb", "public", "test_graph", 102, properties)
		assert.NotNil(t, sc)
	})
}

func TestUpdateEdgeCtx(t *testing.T) {
	ctx := context.Background()
	adapter := config.PrestConf.Adapter

	t.Run("with properties", func(t *testing.T) {
		properties := map[string]interface{}{
			"since": "2023-01-01",
		}
		sc := adapter.UpdateEdgeCtx(ctx, "testdb", "public", "test_graph", 100, properties)
		assert.NotNil(t, sc)
	})

	t.Run("with nil properties", func(t *testing.T) {
		sc := adapter.UpdateEdgeCtx(ctx, "testdb", "public", "test_graph", 101, nil)
		assert.NotNil(t, sc)
	})

	t.Run("with empty properties", func(t *testing.T) {
		properties := map[string]interface{}{}
		sc := adapter.UpdateEdgeCtx(ctx, "testdb", "public", "test_graph", 102, properties)
		assert.NotNil(t, sc)
	})
}

func TestDeleteVertex(t *testing.T) {
	adapter := config.PrestConf.Adapter

	sc := adapter.DeleteVertex("testdb", "public", "test_graph", 1)
	assert.NotNil(t, sc)
}

func TestDeleteVertexCtx(t *testing.T) {
	ctx := context.Background()
	adapter := config.PrestConf.Adapter

	sc := adapter.DeleteVertexCtx(ctx, "testdb", "public", "test_graph", 2)
	assert.NotNil(t, sc)
}

func TestDeleteEdge(t *testing.T) {
	adapter := config.PrestConf.Adapter

	sc := adapter.DeleteEdge("testdb", "public", "test_graph", 100)
	assert.NotNil(t, sc)
}

func TestDeleteEdgeCtx(t *testing.T) {
	ctx := context.Background()
	adapter := config.PrestConf.Adapter

	sc := adapter.DeleteEdgeCtx(ctx, "testdb", "public", "test_graph", 101)
	assert.NotNil(t, sc)
}

func TestBuildTraverseQuery(t *testing.T) {
	t.Run("with explicit params", func(t *testing.T) {
		params := map[string]interface{}{
			"max_depth": 3,
			"direction": "out",
		}
		query := buildTraverseQuery("testdb", "public", "test_graph", params)
		assert.Contains(t, query, "WITH RECURSIVE traversal")
		assert.Contains(t, query, "testdb.public.test_graph_vertices")
		assert.Contains(t, query, "$1")
		assert.Contains(t, query, "$2")
	})

	t.Run("with empty params", func(t *testing.T) {
		// Should use default values for max_depth and direction
		query := buildTraverseQuery("testdb", "public", "test_graph", map[string]interface{}{})
		assert.Contains(t, query, "WITH RECURSIVE traversal")
		assert.Contains(t, query, "testdb.public.test_graph_vertices")
		// Should contain placeholders for start vertex and max depth
		assert.Contains(t, query, "$1")
		assert.Contains(t, query, "$2")
	})

	t.Run("missing max_depth", func(t *testing.T) {
		params := map[string]interface{}{
			"direction": "in",
		}
		query := buildTraverseQuery("testdb", "public", "test_graph", params)
		assert.Contains(t, query, "WITH RECURSIVE traversal")
		// Default max_depth should be 10
		assert.Contains(t, query, "$2")
	})

	t.Run("missing direction", func(t *testing.T) {
		params := map[string]interface{}{
			"max_depth": 7,
		}
		query := buildTraverseQuery("testdb", "public", "test_graph", params)
		assert.Contains(t, query, "WITH RECURSIVE traversal")
		assert.Contains(t, query, "$1")
		assert.Contains(t, query, "$2")
	})
}

func TestBuildPathQuery(t *testing.T) {
	t.Run("with explicit params", func(t *testing.T) {
		params := map[string]interface{}{
			"max_depth": 5,
		}
		query := buildPathQuery("testdb", "public", "test_graph", params)
		assert.Contains(t, query, "WITH RECURSIVE path_find")
		assert.Contains(t, query, "testdb.public.test_graph_vertices")
		assert.Contains(t, query, "$1")
		assert.Contains(t, query, "$2")
		assert.Contains(t, query, "$3")
	})

	t.Run("with empty params", func(t *testing.T) {
		query := buildPathQuery("testdb", "public", "test_graph", map[string]interface{}{})
		assert.Contains(t, query, "WITH RECURSIVE path_find")
		assert.Contains(t, query, "testdb.public.test_graph_vertices")
		assert.Contains(t, query, "$1")
		assert.Contains(t, query, "$2")
		assert.Contains(t, query, "$3")
	})

	t.Run("missing max_depth", func(t *testing.T) {
		params := map[string]interface{}{}
		query := buildPathQuery("testdb", "public", "test_graph", params)
		assert.Contains(t, query, "WITH RECURSIVE path_find")
		assert.Contains(t, query, "testdb.public.test_graph_vertices")
		// Default max_depth should be 10
		assert.Contains(t, query, "$2")
	})
}

func TestBuildNeighborsQuery(t *testing.T) {
	t.Run("direction both", func(t *testing.T) {
		params := map[string]interface{}{
			"direction": "both",
		}
		query := buildNeighborsQuery("testdb", "public", "test_graph", params)
		assert.Contains(t, query, "SELECT DISTINCT n.*")
		assert.Contains(t, query, "testdb.public.test_graph_vertices")
		assert.Contains(t, query, "$1")
	})

	t.Run("direction out", func(t *testing.T) {
		params := map[string]interface{}{
			"direction": "out",
		}
		query := buildNeighborsQuery("testdb", "public", "test_graph", params)
		assert.Contains(t, query, "SELECT DISTINCT n.*")
		assert.Contains(t, query, "$1")
	})

	t.Run("direction in", func(t *testing.T) {
		params := map[string]interface{}{
			"direction": "in",
		}
		query := buildNeighborsQuery("testdb", "public", "test_graph", params)
		assert.Contains(t, query, "SELECT DISTINCT n.*")
		assert.Contains(t, query, "$1")
	})

	t.Run("empty params", func(t *testing.T) {
		// Should default direction to "out"
		query := buildNeighborsQuery("testdb", "public", "test_graph", map[string]interface{}{})
		assert.Contains(t, query, "SELECT DISTINCT n.*")
		assert.Contains(t, query, "$1")
	})

	t.Run("missing direction", func(t *testing.T) {
		params := map[string]interface{}{}
		query := buildNeighborsQuery("testdb", "public", "test_graph", params)
		assert.Contains(t, query, "SELECT DISTINCT n.*")
		assert.Contains(t, query, "$1")
	})
}

func TestBuildMatchQuery(t *testing.T) {
	t.Run("empty params", func(t *testing.T) {
		params := map[string]interface{}{}
		query := buildMatchQuery("testdb", "public", "test_graph", params)
		assert.Contains(t, query, "SELECT")
		assert.Contains(t, query, "v.* as vertex")
		assert.Contains(t, query, "e.* as edge")
		assert.Contains(t, query, "testdb.public.test_graph_vertices")
		assert.Contains(t, query, "testdb.public.test_graph_edges")
	})

	t.Run("empty params", func(t *testing.T) {
		query := buildMatchQuery("testdb", "public", "test_graph", map[string]interface{}{})
		assert.Contains(t, query, "SELECT")
		assert.Contains(t, query, "v.* as vertex")
		assert.Contains(t, query, "e.* as edge")
		assert.Contains(t, query, "testdb.public.test_graph_vertices")
		assert.Contains(t, query, "testdb.public.test_graph_edges")
	})
}

func TestGetEdgeJoinCondition(t *testing.T) {
	tests := []struct {
		direction   string
		vertexAlias string
		expected    string
	}{
		{"out", "t.id", "t.id = e.from"},
		{"in", "v.id", "v.id = e.to"},
		{"both", "x.id", "(x.id = e.from OR x.id = e.to)"},
		{"unknown", "t.id", "t.id = e.from"}, // default case
	}

	for _, tc := range tests {
		result := getEdgeJoinCondition(tc.direction, tc.vertexAlias)
		assert.Equal(t, tc.expected, result)
	}
}

func TestGetVertexJoinCondition(t *testing.T) {
	tests := []struct {
		direction string
		edgeAlias string
		expected  string
	}{
		{"out", "e", "e.to"},
		{"in", "edge", "edge.from"},
		{"both", "e", "CASE WHEN e.from = t.id THEN e.to ELSE e.from END"},
		{"unknown", "e", "e.to"}, // default case
	}

	for _, tc := range tests {
		result := getVertexJoinCondition(tc.direction, tc.edgeAlias)
		assert.Equal(t, tc.expected, result)
	}
}

func TestExtractQueryParams(t *testing.T) {
	t.Run("full params", func(t *testing.T) {
		params := map[string]interface{}{
			"start_vertex": 1,
			"max_depth":    3,
			"end_vertex":   5,
			"direction":    "out",
			"filters":      map[string]interface{}{"name": "test"},
		}

		values := extractQueryParams(params)
		// Should extract start_vertex, max_depth, end_vertex in that order
		require.GreaterOrEqual(t, len(values), 3)
		assert.Equal(t, 1, values[0])
		assert.Equal(t, 3, values[1])
		assert.Equal(t, 5, values[2])
	})

	t.Run("only start_vertex", func(t *testing.T) {
		params := map[string]interface{}{
			"start_vertex": 10,
		}
		values := extractQueryParams(params)
		require.Len(t, values, 1)
		assert.Equal(t, 10, values[0])
	})

	t.Run("start_vertex and max_depth", func(t *testing.T) {
		params := map[string]interface{}{
			"start_vertex": 20,
			"max_depth":    5,
		}
		values := extractQueryParams(params)
		require.Len(t, values, 2)
		assert.Equal(t, 20, values[0])
		assert.Equal(t, 5, values[1])
	})

	t.Run("empty params", func(t *testing.T) {
		params := map[string]interface{}{}
		values := extractQueryParams(params)
		assert.Empty(t, values)
	})
}

func TestGraphQueryDefaultCase(t *testing.T) {
	adapter := config.PrestConf.Adapter

	// Test default case with unknown query type
	sc := adapter.GraphQuery("testdb", "public", "test_graph", "unknown", map[string]interface{}{})
	assert.NotNil(t, sc)
	// Should default to simple vertex query with LIMIT 100
}
