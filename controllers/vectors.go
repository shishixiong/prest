package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/prest/prest/v2/adapters/postgres/formatters"
	"github.com/prest/prest/v2/config"

	"github.com/gorilla/mux"
)

// VectorSearchRequest 向量搜索请求
type VectorSearchRequest struct {
	QueryVector     []float64              `json:"query_vector"`
	VectorField     string                 `json:"vector_field,omitempty"`
	DistanceMetric  string                 `json:"distance_metric,omitempty"` // cosine/l2/inner_product
	Limit           int                    `json:"limit,omitempty"`
	Where           map[string]interface{} `json:"where,omitempty"`
	IncludeDistance bool                   `json:"include_distance,omitempty"`
	Offset          int                    `json:"offset,omitempty"`
}

// VectorSearchResponse 向量搜索响应项
type VectorSearchResponseItem struct {
	ID       interface{}            `json:"id"`
	Distance *float64               `json:"distance,omitempty"`
	Data     map[string]interface{} `json:"data"`
}

// VectorSearch 处理向量相似性搜索
func VectorSearch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	database := vars["database"]
	schema := vars["schema"]
	table := vars["table"]

	// 解析请求体
	var req VectorSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// 验证必填字段
	if len(req.QueryVector) == 0 {
		jsonError(w, "query_vector is required", http.StatusBadRequest)
		return
	}
	if len(req.VectorField) == 0 {
		jsonError(w, "vector_field is required", http.StatusBadRequest)
		return
	}

	// 设置默认值
	if req.DistanceMetric == "" {
		req.DistanceMetric = "cosine" // 默认使用余弦距离
	}
	if req.Limit <= 0 {
		req.Limit = 10 // 默认返回10条结果
	}
	if req.Limit > 100 {
		req.Limit = 100 // 限制最大返回数量
	}

	// 获取适配器
	adapter := config.PrestConf.Adapter

	// 构建向量搜索SQL
	// 格式: SELECT *, vector_field <-> $1 AS distance FROM table WHERE ... ORDER BY distance LIMIT n
	distanceOp := getDistanceOperator(req.DistanceMetric)
	if distanceOp == "" {
		jsonError(w, fmt.Sprintf("Unsupported distance metric: %s. Supported: cosine, l2, inner_product", req.DistanceMetric), http.StatusBadRequest)
		return
	}

	// 使用格式化器格式化向量值
	vectorStr := formatters.FormatVector(req.QueryVector)
	if vectorStr == "" {
		jsonError(w, "Failed to format query vector", http.StatusBadRequest)
		return
	}

	// 构建基础查询
	sql := fmt.Sprintf(`SELECT *, %s %s $1 AS _vector_distance FROM %s.%s.%s`,
		req.VectorField, distanceOp, database, schema, table)

	// 处理WHERE条件
	whereValues := []interface{}{vectorStr} // $1 是向量值
	if req.Where != nil && len(req.Where) > 0 {
		// 将map转换为URL查询参数，以便使用适配器的WhereByRequest
		// 创建虚拟请求
		// 注意：这里简化处理，实际应该使用适配器方法
		whereParts := []string{}
		i := 2 // 参数索引从2开始（$1是查询向量）
		for field, value := range req.Where {
			// 简单的相等条件
			whereParts = append(whereParts, fmt.Sprintf(`"%s" = $%d`, field, i))
			whereValues = append(whereValues, value)
			i++
		}
		if len(whereParts) > 0 {
			sql += " WHERE " + strings.Join(whereParts, " AND ")
		}
	}

	// 添加排序和限制
	sql += fmt.Sprintf(" ORDER BY _vector_distance LIMIT %d", req.Limit)
	if req.Offset > 0 {
		sql += fmt.Sprintf(" OFFSET %d", req.Offset)
	}

	// 执行查询
	sc := adapter.Query(sql, whereValues...)
	if err := sc.Err(); err != nil {
		jsonError(w, fmt.Sprintf("Vector search failed: %v", err), http.StatusBadRequest)
		return
	}

	// 解析结果
	var results []map[string]interface{}
	if err := json.Unmarshal(sc.Bytes(), &results); err != nil {
		jsonError(w, fmt.Sprintf("Failed to parse results: %v", err), http.StatusBadRequest)
		return
	}

	// 构建响应
	responseItems := make([]VectorSearchResponseItem, len(results))
	for i, row := range results {
		item := VectorSearchResponseItem{
			Data: row,
		}

		// 提取ID（假设有id字段）
		if id, ok := row["id"]; ok {
			item.ID = id
		}

		// 提取距离
		if distance, ok := row["_vector_distance"].(float64); ok && req.IncludeDistance {
			item.Distance = &distance
			// 从数据中移除距离字段
			delete(row, "_vector_distance")
		}

		responseItems[i] = item
	}

	// 返回JSON响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseItems)
}

// getDistanceOperator 获取距离操作符
func getDistanceOperator(metric string) string {
	switch strings.ToLower(metric) {
	case "l2", "euclidean":
		return "<->"
	case "cosine":
		return "<=>"
	case "inner", "inner_product", "ip":
		return "<#>"
	default:
		return ""
	}
}

// CreateVectorIndex 创建向量索引
func CreateVectorIndex(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	database := vars["database"]
	schema := vars["schema"]
	table := vars["table"]

	type IndexRequest struct {
		VectorType     string `json:"vector_type,omitempty"` //vector/bit/sparsevec
		VectorField    string `json:"vector_field,omitempty"`
		IndexType      string `json:"index_type"`                // hnsw/ivfflat
		DistanceMetric string `json:"distance_metric"`           // cosine/l2/inner_product
		M              int    `json:"m,omitempty"`               // HNSW参数
		EFConstruction int    `json:"ef_construction,omitempty"` // HNSW参数
		Lists          int    `json:"lists,omitempty"`           // IVFFlat参数
	}

	var req IndexRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if req.VectorType == "" {
		req.VectorType = "vector"
	}
	if req.IndexType == "" {
		req.IndexType = "hnsw"
	}
	if req.DistanceMetric == "" {
		req.DistanceMetric = "cosine"
	}

	// 构建创建索引的SQL
	var sql string
	indexName := fmt.Sprintf("%s_%s_vector_idx", table, req.VectorField)

	switch req.IndexType {
	case "hnsw":
		if req.M <= 0 {
			req.M = 16
		}
		if req.EFConstruction <= 0 {
			req.EFConstruction = 64
		}
		sql = fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %s ON %s.%s.%s USING hnsw (%s %s_cosine_ops) WITH (m = %d, ef_construction = %d)`,
			indexName, database, schema, table, req.VectorField, req.VectorType, req.M, req.EFConstruction)
	case "ivfflat":
		if req.Lists <= 0 {
			req.Lists = 100
		}
		sql = fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %s ON %s.%s.%s USING ivfflat (%s %s_cosine_ops) WITH (lists = %d)`,
			indexName, database, schema, table, req.VectorField, req.VectorType, req.Lists)
	default:
		jsonError(w, fmt.Sprintf("Unsupported index type: %s. Supported: hnsw, ivfflat", req.IndexType), http.StatusBadRequest)
		return
	}

	// 执行创建索引
	adapter := config.PrestConf.Adapter
	sc := adapter.Query(sql)
	if err := sc.Err(); err != nil {
		jsonError(w, fmt.Sprintf("Failed to create vector index: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":    fmt.Sprintf("Vector index %s created successfully", indexName),
		"index_name": indexName,
	})
}

// DeleteVectorIndex 删除向量索引
func DeleteVectorIndex(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	database := vars["database"]
	schema := vars["schema"]
	table := vars["table"]

	type DeleteRequest struct {
		VectorField string `json:"vector_field,omitempty"`
		IndexName   string `json:"index_name,omitempty"`
	}

	var req DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// 确定索引名
	indexName := req.IndexName
	if indexName == "" {
		if req.VectorField == "" {
			req.VectorField = "vector"
		}
		indexName = fmt.Sprintf("%s_%s_vector_idx", table, req.VectorField)
	}

	// 构建删除索引的SQL
	sql := fmt.Sprintf(`DROP INDEX IF EXISTS %s.%s.%s`,
		database, schema, indexName)

	adapter := config.PrestConf.Adapter
	sc := adapter.Query(sql)
	if err := sc.Err(); err != nil {
		jsonError(w, fmt.Sprintf("Failed to delete vector index: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Vector index %s deleted successfully", indexName),
	})
}

// VectorBatchSearch 批量向量搜索
func VectorBatchSearch(w http.ResponseWriter, r *http.Request) {
	jsonError(w, "Batch vector search not yet implemented", http.StatusNotImplemented)
}
