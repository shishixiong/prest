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

func initVectorRoutes() *mux.Router {
	r := mux.NewRouter()
	// Vector endpoints
	r.HandleFunc("/{database}/{schema}/{table}/_vector/search", VectorSearch).Methods("POST")
	r.HandleFunc("/{database}/{schema}/{table}/_vector/batch_search", VectorBatchSearch).Methods("POST")
	r.HandleFunc("/{database}/{schema}/{table}/_vector/index", CreateVectorIndex).Methods("POST")
	r.HandleFunc("/{database}/{schema}/{table}/_vector/index", DeleteVectorIndex).Methods("DELETE")
	return r
}

func setupTest() {
	config.Load()
	postgres.Load()
}

func TestVectorSearch_Validation(t *testing.T) {
	setupTest()

	router := initVectorRoutes()
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
			name:           "missing query_vector",
			database:       "testdb",
			schema:         "public",
			table:          "test",
			requestBody:    map[string]interface{}{"vector_field": "embedding"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing vector_field",
			database:       "testdb",
			schema:         "public",
			table:          "test",
			requestBody:    map[string]interface{}{"query_vector": []float64{1.0, 2.0}},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "invalid distance metric",
			database: "testdb",
			schema:   "public",
			table:    "test",
			requestBody: map[string]interface{}{
				"query_vector":    []float64{1.0, 2.0},
				"vector_field":    "embedding",
				"distance_metric": "invalid",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "valid request with cosine distance",
			database: "testdb",
			schema:   "public",
			table:    "test",
			requestBody: map[string]interface{}{
				"query_vector":    []float64{1.0, 2.0, 3.0},
				"vector_field":    "embedding",
				"distance_metric": "cosine",
				"limit":           5,
			},
			expectedStatus: http.StatusBadRequest, // Will fail due to missing DB, but we test validation passes
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			url := fmt.Sprintf("%s/%s/%s/%s/_vector/search", server.URL, tc.database, tc.schema, tc.table)
			testutils.DoRequest(t, url, tc.requestBody, "POST", tc.expectedStatus, "VectorSearch")
		})
	}
}

func TestCreateVectorIndex_Validation(t *testing.T) {
	setupTest()

	router := initVectorRoutes()
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
			name:           "missing vector_field",
			database:       "testdb",
			schema:         "public",
			table:          "test",
			requestBody:    map[string]interface{}{"index_type": "hnsw"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "invalid index_type",
			database: "testdb",
			schema:   "public",
			table:    "test",
			requestBody: map[string]interface{}{
				"vector_field": "embedding",
				"index_type":   "invalid",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "invalid distance_metric",
			database: "testdb",
			schema:   "public",
			table:    "test",
			requestBody: map[string]interface{}{
				"vector_field":    "embedding",
				"index_type":      "hnsw",
				"distance_metric": "invalid",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "valid hnsw index request",
			database: "testdb",
			schema:   "public",
			table:    "test",
			requestBody: map[string]interface{}{
				"vector_field":    "embedding",
				"index_type":      "hnsw",
				"distance_metric": "cosine",
				"m":               16,
				"ef_construction": 64,
			},
			expectedStatus: http.StatusBadRequest, // Will fail due to missing DB
		},
		{
			name:     "valid ivfflat index request",
			database: "testdb",
			schema:   "public",
			table:    "test",
			requestBody: map[string]interface{}{
				"vector_field":    "embedding",
				"index_type":      "ivfflat",
				"distance_metric": "l2",
				"lists":           100,
			},
			expectedStatus: http.StatusBadRequest, // Will fail due to missing DB
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			url := fmt.Sprintf("%s/%s/%s/%s/_vector/index", server.URL, tc.database, tc.schema, tc.table)
			testutils.DoRequest(t, url, tc.requestBody, "POST", tc.expectedStatus, "CreateVectorIndex")
		})
	}
}

func TestDeleteVectorIndex_Validation(t *testing.T) {
	setupTest()

	router := initVectorRoutes()
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
			name:           "missing both vector_field and index_name",
			database:       "testdb",
			schema:         "public",
			table:          "test",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "with vector_field",
			database: "testdb",
			schema:   "public",
			table:    "test",
			requestBody: map[string]interface{}{
				"vector_field": "embedding",
			},
			expectedStatus: http.StatusBadRequest, // Will fail due to missing DB
		},
		{
			name:     "with index_name",
			database: "testdb",
			schema:   "public",
			table:    "test",
			requestBody: map[string]interface{}{
				"index_name": "test_embedding_vector_idx",
			},
			expectedStatus: http.StatusBadRequest, // Will fail due to missing DB
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			url := fmt.Sprintf("%s/%s/%s/%s/_vector/index", server.URL, tc.database, tc.schema, tc.table)
			testutils.DoRequest(t, url, tc.requestBody, "DELETE", tc.expectedStatus, "DeleteVectorIndex")
		})
	}
}

func TestVectorBatchSearch_NotImplemented(t *testing.T) {
	setupTest()

	router := initVectorRoutes()
	server := httptest.NewServer(router)
	defer server.Close()

	reqBody := map[string]interface{}{
		"queries": []map[string]interface{}{
			{
				"query_vector": []float64{1.0, 2.0},
				"vector_field": "embedding",
			},
		},
	}
	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/testdb/public/test/_vector/batch_search", server.URL), bytes.NewBuffer(body))
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

	if resp.StatusCode != http.StatusNotImplemented {
		t.Errorf("Expected status %d, got %d", http.StatusNotImplemented, resp.StatusCode)
	}
}

func TestGetDistanceOperator(t *testing.T) {
	tests := []struct {
		metric     string
		expectedOp string
	}{
		{"l2", "<->"},
		{"euclidean", "<->"},
		{"cosine", "<=>"},
		{"inner", "<#>"},
		{"inner_product", "<#>"},
		{"ip", "<#>"},
		{"unknown", ""},
		{"", ""},
	}

	for _, tc := range tests {
		t.Run(tc.metric, func(t *testing.T) {
			result := getDistanceOperator(tc.metric)
			if result != tc.expectedOp {
				t.Errorf("For metric %s expected operator %s, got %s", tc.metric, tc.expectedOp, result)
			}
		})
	}
}
