package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/prest/prest/v2/adapters/postgres"
	"github.com/prest/prest/v2/config"
	"github.com/prest/prest/v2/testutils"
)

func initGraphRoutes() *mux.Router {
	r := mux.NewRouter()
	// Graph endpoints
	r.HandleFunc("/{database}/{schema}/{table}/_graph/create", CreateGraph).Methods("POST")
	r.HandleFunc("/{database}/{schema}/{table}/_graph", DeleteGraph).Methods("DELETE")
	r.HandleFunc("/{database}/{schema}/{table}/_graph/query", GraphQuery).Methods("POST")
	r.HandleFunc("/{database}/{schema}/{table}/_graph/vertices", AddVertices).Methods("POST")
	r.HandleFunc("/{database}/{schema}/{table}/_graph/vertices/{id}", GetVertex).Methods("GET")
	r.HandleFunc("/{database}/{schema}/{table}/_graph/vertices/{id}", UpdateVertex).Methods("PUT", "PATCH")
	r.HandleFunc("/{database}/{schema}/{table}/_graph/vertices/{id}", DeleteVertex).Methods("DELETE")
	r.HandleFunc("/{database}/{schema}/{table}/_graph/edges", AddEdges).Methods("POST")
	r.HandleFunc("/{database}/{schema}/{table}/_graph/edges/{id}", GetEdge).Methods("GET")
	r.HandleFunc("/{database}/{schema}/{table}/_graph/edges/{id}", UpdateEdge).Methods("PUT", "PATCH")
	r.HandleFunc("/{database}/{schema}/{table}/_graph/edges/{id}", DeleteEdge).Methods("DELETE")
	return r
}

func setupGraphTest() {
	config.Load()
	postgres.Load()
}

func TestGraphQuery_Validation(t *testing.T) {
	setupGraphTest()

	router := initGraphRoutes()
	server := httptest.NewServer(router)
	defer server.Close()

	tests := []struct {
		name           string
		database       string
		schema         string
		table          string
		requestBody    map[string]interface{}
		expectedStatus int
	}{
		{
			name:           "missing both query_type and query",
			database:       "testdb",
			schema:         "public",
			table:          "test_graph",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "traverse query missing start_vertex",
			database: "testdb",
			schema:   "public",
			table:    "test_graph",
			requestBody: map[string]interface{}{
				"query_type": "traverse",
				"max_depth":  2,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "path query missing end_vertex",
			database: "testdb",
			schema:   "public",
			table:    "test_graph",
			requestBody: map[string]interface{}{
				"query_type":   "path",
				"start_vertex": 1,
				"max_depth":    3,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "valid traverse query with structured params",
			database: "testdb",
			schema:   "public",
			table:    "test_graph",
			requestBody: map[string]interface{}{
				"query_type":   "traverse",
				"start_vertex": 1,
				"max_depth":    2,
				"direction":    "out",
			},
			expectedStatus: http.StatusBadRequest, // Will fail due to missing DB/table
		},
		{
			name:     "valid GraphQL-like traverse query",
			database: "testdb",
			schema:   "public",
			table:    "test_graph",
			requestBody: map[string]interface{}{
				"query": "traverse(from: 1, depth: 2, direction: \"out\")",
			},
			expectedStatus: http.StatusBadRequest, // Will fail due to missing DB/table
		},
		{
			name:     "GraphQL-like vertex query",
			database: "testdb",
			schema:   "public",
			table:    "test_graph",
			requestBody: map[string]interface{}{
				"query": "vertex(id: 1)",
			},
			expectedStatus: http.StatusBadRequest, // Will fail due to missing DB/table
		},
		{
			name:     "GraphQL-like edge query",
			database: "testdb",
			schema:   "public",
			table:    "test_graph",
			requestBody: map[string]interface{}{
				"query": "edge(id: 1)",
			},
			expectedStatus: http.StatusBadRequest, // Will fail due to missing DB/table
		},
		{
			name:     "GraphQL-like path query",
			database: "testdb",
			schema:   "public",
			table:    "test_graph",
			requestBody: map[string]interface{}{
				"query": "path(from: 1, to: 5, maxDepth: 3)",
			},
			expectedStatus: http.StatusBadRequest, // Will fail due to missing DB/table
		},
		{
			name:     "GraphQL-like neighbors query",
			database: "testdb",
			schema:   "public",
			table:    "test_graph",
			requestBody: map[string]interface{}{
				"query": "neighbors(vertex: 1, direction: \"both\")",
			},
			expectedStatus: http.StatusBadRequest, // Will fail due to missing DB/table
		},
		{
			name:     "GraphQL-like match query",
			database: "testdb",
			schema:   "public",
			table:    "test_graph",
			requestBody: map[string]interface{}{
				"query": "match",
			},
			expectedStatus: http.StatusBadRequest, // Will fail due to missing DB/table
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			url := fmt.Sprintf("%s/%s/%s/%s/_graph/query", server.URL, tc.database, tc.schema, tc.table)
			testutils.DoRequest(t, url, tc.requestBody, "POST", tc.expectedStatus, "GraphQuery")
		})
	}
}

func TestParseGraphQLQuery(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectedType   string
		expectedParams map[string]interface{}
		expectError    bool
	}{
		{
			name:         "traverse query with parameters",
			query:        `traverse(from: 1, depth: 2, direction: "out")`,
			expectedType: "traverse",
			expectedParams: map[string]interface{}{
				"start_vertex": 1,
				"max_depth":    2,
				"direction":    "out",
			},
			expectError: false,
		},
		{
			name:         "path query with parameters",
			query:        `path(from: 1, to: 5, maxDepth: 3)`,
			expectedType: "path",
			expectedParams: map[string]interface{}{
				"start_vertex": 1,
				"end_vertex":   5,
				"max_depth":    3,
			},
			expectError: false,
		},
		{
			name:         "neighbors query with parameters",
			query:        `neighbors(vertex: 1, direction: "both")`,
			expectedType: "neighbors",
			expectedParams: map[string]interface{}{
				"start_vertex": 1,
				"direction":    "both",
			},
			expectError: false,
		},
		{
			name:         "vertex query",
			query:        `vertex(id: 123)`,
			expectedType: "vertex",
			expectedParams: map[string]interface{}{
				"vertex_id": 123,
			},
			expectError: false,
		},
		{
			name:         "edge query",
			query:        `edge(id: 456)`,
			expectedType: "edge",
			expectedParams: map[string]interface{}{
				"edge_id": 456,
			},
			expectError: false,
		},
		{
			name:         "match query",
			query:        `match`,
			expectedType: "match",
			expectedParams: map[string]interface{}{
				"limit": 100,
			},
			expectError: false,
		},
		{
			name:         "empty query",
			query:        "",
			expectedType: "",
			expectedParams: nil,
			expectError:    true,
		},
		{
			name:         "invalid query format",
			query:        `invalid(abc: 123)`,
			expectedType: "match", // default fallback
			expectedParams: map[string]interface{}{
				"limit": 100,
			},
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			queryType, params, err := parseGraphQLQuery(tc.query)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if queryType != tc.expectedType {
				t.Errorf("Expected query type %s, got %s", tc.expectedType, queryType)
			}

			// Compare parameters
			for key, expectedValue := range tc.expectedParams {
				actualValue, ok := params[key]
				if !ok {
					t.Errorf("Missing expected parameter %s", key)
					continue
				}

				// Compare values
				switch v := expectedValue.(type) {
				case int:
					if actualInt, ok := actualValue.(int); !ok || actualInt != v {
						t.Errorf("Parameter %s: expected %v (%T), got %v (%T)", key, v, v, actualValue, actualValue)
					}
				case string:
					if actualStr, ok := actualValue.(string); !ok || actualStr != v {
						t.Errorf("Parameter %s: expected %v (%T), got %v (%T)", key, v, v, actualValue, actualValue)
					}
				default:
					// For other types, just check existence
					if actualValue != expectedValue {
						t.Errorf("Parameter %s: expected %v (%T), got %v (%T)", key, expectedValue, expectedValue, actualValue, actualValue)
					}
				}
			}

			// Check for unexpected extra parameters
			for key := range params {
				if _, ok := tc.expectedParams[key]; !ok && key != "query_type" {
					t.Errorf("Unexpected parameter %s with value %v", key, params[key])
				}
			}
		})
	}
}

func TestCreateGraph_Validation(t *testing.T) {
	setupGraphTest()

	router := initGraphRoutes()
	server := httptest.NewServer(router)
	defer server.Close()

	tests := []struct {
		name           string
		database       string
		schema         string
		table          string
		requestBody    map[string]interface{}
		expectedStatus int
	}{
		{
			name:           "valid create graph request",
			database:       "testdb",
			schema:         "public",
			table:          "test_graph",
			requestBody:    map[string]interface{}{"graph_name": "mygraph", "graph_type": "property"},
			expectedStatus: http.StatusBadRequest, // Will fail due to missing DB
		},
		{
			name:           "create graph with default type",
			database:       "testdb",
			schema:         "public",
			table:          "test_graph",
			requestBody:    map[string]interface{}{"graph_name": "mygraph"},
			expectedStatus: http.StatusBadRequest, // Will fail due to missing DB
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			url := fmt.Sprintf("%s/%s/%s/%s/_graph/create", server.URL, tc.database, tc.schema, tc.table)
			testutils.DoRequest(t, url, tc.requestBody, "POST", tc.expectedStatus, "CreateGraph")
		})
	}
}

func TestAddVertices_Validation(t *testing.T) {
	setupGraphTest()

	router := initGraphRoutes()
	server := httptest.NewServer(router)
	defer server.Close()

	tests := []struct {
		name           string
		database       string
		schema         string
		table          string
		requestBody    map[string]interface{}
		expectedStatus int
	}{
		{
			name:           "empty vertices array",
			database:       "testdb",
			schema:         "public",
			table:          "test_graph",
			requestBody:    map[string]interface{}{"vertices": []interface{}{}},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing vertices array",
			database:       "testdb",
			schema:         "public",
			table:          "test_graph",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "valid vertices request",
			database: "testdb",
			schema:   "public",
			table:    "test_graph",
			requestBody: map[string]interface{}{
				"vertices": []interface{}{
					map[string]interface{}{
						"id":    1,
						"label": "person",
						"properties": map[string]interface{}{
							"name": "Alice",
							"age":  30,
						},
					},
				},
			},
			expectedStatus: http.StatusBadRequest, // Will fail due to missing DB
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			url := fmt.Sprintf("%s/%s/%s/%s/_graph/vertices", server.URL, tc.database, tc.schema, tc.table)
			testutils.DoRequest(t, url, tc.requestBody, "POST", tc.expectedStatus, "AddVertices")
		})
	}
}

func TestAddEdges_Validation(t *testing.T) {
	setupGraphTest()

	router := initGraphRoutes()
	server := httptest.NewServer(router)
	defer server.Close()

	tests := []struct {
		name           string
		database       string
		schema         string
		table          string
		requestBody    map[string]interface{}
		expectedStatus int
	}{
		{
			name:           "empty edges array",
			database:       "testdb",
			schema:         "public",
			table:          "test_graph",
			requestBody:    map[string]interface{}{"edges": []interface{}{}},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing edges array",
			database:       "testdb",
			schema:         "public",
			table:          "test_graph",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "valid edges request",
			database: "testdb",
			schema:   "public",
			table:    "test_graph",
			requestBody: map[string]interface{}{
				"edges": []interface{}{
					map[string]interface{}{
						"id":    100,
						"from":  1,
						"to":    2,
						"label": "knows",
						"properties": map[string]interface{}{
							"since": "2020-01-01",
						},
					},
				},
			},
			expectedStatus: http.StatusBadRequest, // Will fail due to missing DB
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			url := fmt.Sprintf("%s/%s/%s/%s/_graph/edges", server.URL, tc.database, tc.schema, tc.table)
			testutils.DoRequest(t, url, tc.requestBody, "POST", tc.expectedStatus, "AddEdges")
		})
	}
}

func TestGetVertex_Validation(t *testing.T) {
	setupGraphTest()

	router := initGraphRoutes()
	server := httptest.NewServer(router)
	defer server.Close()

	// Test GET endpoint - no request body needed
	url := fmt.Sprintf("%s/testdb/public/test_graph/_graph/vertices/1", server.URL)
	body, _ := json.Marshal(map[string]interface{}{})
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestGetEdge_Validation(t *testing.T) {
	setupGraphTest()

	router := initGraphRoutes()
	server := httptest.NewServer(router)
	defer server.Close()

	// Test GET endpoint - no request body needed
	url := fmt.Sprintf("%s/testdb/public/test_graph/_graph/edges/100", server.URL)
	body, _ := json.Marshal(map[string]interface{}{})
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}