package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/prest/prest/v2/adapters"
	"github.com/prest/prest/v2/config"

	"github.com/gorilla/mux"
)

// GraphQueryRequest 图查询请求
type GraphQueryRequest struct {
	QueryType   string                 `json:"query_type"`             // traverse/path/neighbors/match
	Query       string                 `json:"query,omitempty"`        // GraphQL-like查询字符串
	StartVertex interface{}            `json:"start_vertex,omitempty"`
	EndVertex   interface{}            `json:"end_vertex,omitempty"`
	MaxDepth    int                    `json:"max_depth,omitempty"`
	Direction   string                 `json:"direction,omitempty"`    // out/in/both
	Filters     map[string]interface{} `json:"filters,omitempty"`
	Limit       int                    `json:"limit,omitempty"`
	Offset      int                    `json:"offset,omitempty"`
}

// GraphQueryResponse 图查询响应
type GraphQueryResponse struct {
	Vertices []VertexResponse `json:"vertices,omitempty"`
	Edges    []EdgeResponse   `json:"edges,omitempty"`
	Paths    []PathResponse   `json:"paths,omitempty"`
	Total    int              `json:"total,omitempty"`
}

// VertexResponse 顶点响应
type VertexResponse struct {
	ID         interface{}            `json:"id"`
	Label      string                 `json:"label,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// EdgeResponse 边响应
type EdgeResponse struct {
	ID         interface{}            `json:"id"`
	From       interface{}            `json:"from"`
	To         interface{}            `json:"to"`
	Label      string                 `json:"label,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// PathResponse 路径响应
type PathResponse struct {
	Vertices []VertexResponse `json:"vertices"`
	Edges    []EdgeResponse   `json:"edges"`
	Length   int              `json:"length"`
}

// CreateGraphRequest 创建图请求
type CreateGraphRequest struct {
	GraphName string `json:"graph_name"`
	GraphType string `json:"graph_type,omitempty"` // property/directed/undirected
}

// AddVerticesRequest 添加顶点请求
type AddVerticesRequest struct {
	Vertices []VertexRequest `json:"vertices"`
}

// VertexRequest 顶点请求
type VertexRequest struct {
	ID         interface{}            `json:"id,omitempty"`
	Label      string                 `json:"label,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// AddEdgesRequest 添加边请求
type AddEdgesRequest struct {
	Edges []EdgeRequest `json:"edges"`
}

// EdgeRequest 边请求
type EdgeRequest struct {
	ID         interface{}            `json:"id,omitempty"`
	From       interface{}            `json:"from"`
	To         interface{}            `json:"to"`
	Label      string                 `json:"label,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// GraphQuery 处理图查询
func GraphQuery(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	database := vars["database"]
	schema := vars["schema"]
	table := vars["table"]

	// 解析请求体
	var req GraphQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// 设置默认值
	if req.Limit <= 0 {
		req.Limit = 100
	}
	if req.Limit > 1000 {
		req.Limit = 1000
	}
	if req.MaxDepth <= 0 {
		req.MaxDepth = 10
	}
	if req.MaxDepth > config.PrestConf.Graph.MaxTraversalDepth {
		req.MaxDepth = config.PrestConf.Graph.MaxTraversalDepth
	}
	if req.Direction == "" {
		req.Direction = "out"
	}

	// 获取适配器
	adapter := config.PrestConf.Adapter

	// 确定查询类型和参数
	var finalQueryType string
	var queryParams map[string]interface{}

	if req.Query != "" {
		// 解析GraphQL-like查询字符串
		parsedType, parsedParams, err := parseGraphQLQuery(req.Query)
		if err != nil {
			jsonError(w, fmt.Sprintf("Failed to parse GraphQL query: %v", err), http.StatusBadRequest)
			return
		}
		finalQueryType = parsedType
		queryParams = parsedParams

		// 合并请求中的参数（GraphQL解析的参数优先）
		if req.StartVertex != nil && queryParams["start_vertex"] == nil {
			queryParams["start_vertex"] = req.StartVertex
		}
		if req.EndVertex != nil && queryParams["end_vertex"] == nil {
			queryParams["end_vertex"] = req.EndVertex
		}
		if req.MaxDepth > 0 && queryParams["max_depth"] == nil {
			queryParams["max_depth"] = req.MaxDepth
		}
		if req.Direction != "" && queryParams["direction"] == nil {
			queryParams["direction"] = req.Direction
		}
		if req.Filters != nil && queryParams["filters"] == nil {
			queryParams["filters"] = req.Filters
		}
		if req.Limit > 0 && queryParams["limit"] == nil {
			queryParams["limit"] = req.Limit
		}
		if req.Offset > 0 && queryParams["offset"] == nil {
			queryParams["offset"] = req.Offset
		}
	} else {
		// 使用结构化查询
		if req.QueryType == "" {
			jsonError(w, "query_type is required when query is not provided", http.StatusBadRequest)
			return
		}
		finalQueryType = req.QueryType
		queryParams = map[string]interface{}{
			"query_type":   req.QueryType,
			"start_vertex": req.StartVertex,
			"end_vertex":   req.EndVertex,
			"max_depth":    req.MaxDepth,
			"direction":    req.Direction,
			"filters":      req.Filters,
			"limit":        req.Limit,
			"offset":       req.Offset,
		}
	}

	// 确保必要的参数存在
	if finalQueryType == "traverse" || finalQueryType == "path" || finalQueryType == "neighbors" {
		if queryParams["start_vertex"] == nil {
			jsonError(w, "start_vertex is required for this query type", http.StatusBadRequest)
			return
		}
	}
	if finalQueryType == "path" {
		if queryParams["end_vertex"] == nil {
			jsonError(w, "end_vertex is required for path query", http.StatusBadRequest)
			return
		}
	}

	// 执行图查询（支持vertex和edge单独查询）
	var sc adapters.Scanner
	var results []map[string]interface{}
	var singleResult map[string]interface{}

	switch finalQueryType {
	case "vertex":
		vertexID := queryParams["vertex_id"]
		if vertexID == nil {
			jsonError(w, "vertex_id is required for vertex query", http.StatusBadRequest)
			return
		}
		sc = adapter.GetVertex(database, schema, table, vertexID)
		if err := sc.Err(); err != nil {
			jsonError(w, fmt.Sprintf("Vertex query failed: %v", err), http.StatusBadRequest)
			return
		}
		// 解析单个结果
		if err := json.Unmarshal(sc.Bytes(), &singleResult); err != nil {
			jsonError(w, fmt.Sprintf("Failed to parse vertex result: %v", err), http.StatusBadRequest)
			return
		}
		// 转换为results数组格式以兼容buildGraphResponse
		results = []map[string]interface{}{singleResult}

	case "edge":
		edgeID := queryParams["edge_id"]
		if edgeID == nil {
			jsonError(w, "edge_id is required for edge query", http.StatusBadRequest)
			return
		}
		sc = adapter.GetEdge(database, schema, table, edgeID)
		if err := sc.Err(); err != nil {
			jsonError(w, fmt.Sprintf("Edge query failed: %v", err), http.StatusBadRequest)
			return
		}
		// 解析单个结果
		if err := json.Unmarshal(sc.Bytes(), &singleResult); err != nil {
			jsonError(w, fmt.Sprintf("Failed to parse edge result: %v", err), http.StatusBadRequest)
			return
		}
		results = []map[string]interface{}{singleResult}

	default:
		// 使用GraphQuery进行标准图查询
		sc = adapter.GraphQuery(database, schema, table, finalQueryType, queryParams)
		if err := sc.Err(); err != nil {
			jsonError(w, fmt.Sprintf("Graph query failed: %v", err), http.StatusBadRequest)
			return
		}
		// 解析结果数组
		if err := json.Unmarshal(sc.Bytes(), &results); err != nil {
			jsonError(w, fmt.Sprintf("Failed to parse results: %v", err), http.StatusBadRequest)
			return
		}
	}

	// 构建响应
	response := buildGraphResponse(finalQueryType, results)

	// 返回JSON响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateGraph 创建图结构
func CreateGraph(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	database := vars["database"]
	schema := vars["schema"]
	table := vars["table"]

	// 解析请求体
	var req CreateGraphRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// 设置默认值
	if req.GraphType == "" {
		req.GraphType = "property"
	}

	// 确定图名：如果请求中提供了graph_name则使用它，否则使用table参数
	graphName := table
	if req.GraphName != "" {
		graphName = req.GraphName
	}

	// 获取适配器
	adapter := config.PrestConf.Adapter

	// 执行创建图
	sc := adapter.CreateGraph(database, schema, graphName, req.GraphType)
	if err := sc.Err(); err != nil {
		jsonError(w, fmt.Sprintf("Failed to create graph: %v", err), http.StatusBadRequest)
		return
	}

	// 返回成功响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":    fmt.Sprintf("Graph %s created successfully", graphName),
		"graph_name": graphName,
		"graph_type": req.GraphType,
	})
}

// AddVertices 添加顶点到图
func AddVertices(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	database := vars["database"]
	schema := vars["schema"]
	table := vars["table"]

	// 解析请求体
	var req AddVerticesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// 验证必填字段
	if len(req.Vertices) == 0 {
		jsonError(w, "vertices array is required", http.StatusBadRequest)
		return
	}

	// 转换顶点为map格式以便适配器处理
	vertices := make([]map[string]interface{}, len(req.Vertices))
	for i, vertex := range req.Vertices {
		vertices[i] = map[string]interface{}{
			"id":         vertex.ID,
			"label":      vertex.Label,
			"properties": vertex.Properties,
		}
	}

	// 获取适配器
	adapter := config.PrestConf.Adapter

	// 执行添加顶点
	sc := adapter.AddVertices(database, schema, table, vertices)
	if err := sc.Err(); err != nil {
		jsonError(w, fmt.Sprintf("Failed to add vertices: %v", err), http.StatusBadRequest)
		return
	}

	// 返回成功响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  fmt.Sprintf("Added %d vertices successfully", len(req.Vertices)),
		"count":    len(req.Vertices),
		"vertices": req.Vertices,
	})
}

// AddEdges 添加边到图
func AddEdges(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	database := vars["database"]
	schema := vars["schema"]
	table := vars["table"]

	// 解析请求体
	var req AddEdgesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// 验证必填字段
	if len(req.Edges) == 0 {
		jsonError(w, "edges array is required", http.StatusBadRequest)
		return
	}

	// 转换边为map格式以便适配器处理
	edges := make([]map[string]interface{}, len(req.Edges))
	for i, edge := range req.Edges {
		edges[i] = map[string]interface{}{
			"id":         edge.ID,
			"from":       edge.From,
			"to":         edge.To,
			"label":      edge.Label,
			"properties": edge.Properties,
		}
	}

	// 获取适配器
	adapter := config.PrestConf.Adapter

	// 执行添加边
	sc := adapter.AddEdges(database, schema, table, edges)
	if err := sc.Err(); err != nil {
		jsonError(w, fmt.Sprintf("Failed to add edges: %v", err), http.StatusBadRequest)
		return
	}

	// 返回成功响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": fmt.Sprintf("Added %d edges successfully", len(req.Edges)),
		"count":   len(req.Edges),
		"edges":   req.Edges,
	})
}

// buildGraphResponse 根据查询类型构建图响应
func buildGraphResponse(queryType string, results []map[string]interface{}) GraphQueryResponse {
	response := GraphQueryResponse{}

	switch queryType {
	case "traverse", "neighbors":
		// 返回顶点和边
		for _, row := range results {
			if vertex, ok := row["vertex"].(map[string]interface{}); ok {
				response.Vertices = append(response.Vertices, VertexResponse{
					ID:         vertex["id"],
					Label:      getString(vertex, "label"),
					Properties: getMap(vertex, "properties"),
				})
			}
			if edge, ok := row["edge"].(map[string]interface{}); ok {
				response.Edges = append(response.Edges, EdgeResponse{
					ID:         edge["id"],
					From:       edge["from"],
					To:         edge["to"],
					Label:      getString(edge, "label"),
					Properties: getMap(edge, "properties"),
				})
			}
		}
	case "path":
		// 返回路径
		for _, row := range results {
			if path, ok := row["path"].([]interface{}); ok {
				pathResponse := PathResponse{}
				for _, item := range path {
					if itemMap, ok := item.(map[string]interface{}); ok {
						if itemType, ok := itemMap["type"].(string); ok {
							if itemType == "vertex" {
								pathResponse.Vertices = append(pathResponse.Vertices, VertexResponse{
									ID:         itemMap["id"],
									Label:      getString(itemMap, "label"),
									Properties: getMap(itemMap, "properties"),
								})
							} else if itemType == "edge" {
								pathResponse.Edges = append(pathResponse.Edges, EdgeResponse{
									ID:         itemMap["id"],
									From:       itemMap["from"],
									To:         itemMap["to"],
									Label:      getString(itemMap, "label"),
									Properties: getMap(itemMap, "properties"),
								})
							}
						}
					}
				}
				pathResponse.Length = len(pathResponse.Vertices) + len(pathResponse.Edges)
				response.Paths = append(response.Paths, pathResponse)
			}
		}
	case "match":
		// 返回匹配结果
		response.Vertices = make([]VertexResponse, 0)
		response.Edges = make([]EdgeResponse, 0)
		for _, row := range results {
			// 处理匹配查询的结果格式
			for key, value := range row {
				if strings.HasSuffix(key, "_vertex") {
					if vertex, ok := value.(map[string]interface{}); ok {
						response.Vertices = append(response.Vertices, VertexResponse{
							ID:         vertex["id"],
							Label:      getString(vertex, "label"),
							Properties: getMap(vertex, "properties"),
						})
					}
				} else if strings.HasSuffix(key, "_edge") {
					if edge, ok := value.(map[string]interface{}); ok {
						response.Edges = append(response.Edges, EdgeResponse{
							ID:         edge["id"],
							From:       edge["from"],
							To:         edge["to"],
							Label:      getString(edge, "label"),
							Properties: getMap(edge, "properties"),
						})
					}
				}
			}
		}
	}

	response.Total = len(response.Vertices) + len(response.Edges) + len(response.Paths)
	return response
}

// getString 从map中获取字符串值
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// getMap 从map中获取map值
func getMap(m map[string]interface{}, key string) map[string]interface{} {
	if val, ok := m[key]; ok {
		if mp, ok := val.(map[string]interface{}); ok {
			return mp
		}
	}
	return nil
}

// DeleteGraph 删除图结构
func DeleteGraph(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	database := vars["database"]
	schema := vars["schema"]
	table := vars["table"]

	// 使用table参数作为图名
	graphName := table

	// 获取适配器
	adapter := config.PrestConf.Adapter

	// 执行删除图
	sc := adapter.DeleteGraph(database, schema, graphName)
	if err := sc.Err(); err != nil {
		jsonError(w, fmt.Sprintf("Failed to delete graph: %v", err), http.StatusBadRequest)
		return
	}

	// 返回成功响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":    fmt.Sprintf("Graph %s deleted successfully", graphName),
		"graph_name": graphName,
	})
}

// GetVertex 获取顶点
func GetVertex(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	database := vars["database"]
	schema := vars["schema"]
	table := vars["table"]
	vertexID := vars["id"]

	// 获取适配器
	adapter := config.PrestConf.Adapter

	// 执行获取顶点
	sc := adapter.GetVertex(database, schema, table, vertexID)
	if err := sc.Err(); err != nil {
		jsonError(w, fmt.Sprintf("Failed to get vertex: %v", err), http.StatusBadRequest)
		return
	}

	// 解析结果（适配器返回数组格式）
	var results []map[string]interface{}
	if err := json.Unmarshal(sc.Bytes(), &results); err != nil {
		jsonError(w, fmt.Sprintf("Failed to parse result: %v", err), http.StatusBadRequest)
		return
	}

	// 检查是否找到顶点
	if len(results) == 0 {
		jsonError(w, fmt.Sprintf("Vertex with ID %v not found", vertexID), http.StatusNotFound)
		return
	}

	// 获取第一个结果（根据ID查询应该只有一行）
	result := results[0]

	// 构建顶点响应
	vertex := VertexResponse{
		ID:         result["id"],
		Label:      getString(result, "label"),
		Properties: getMap(result, "properties"),
	}

	// 返回JSON响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vertex)
}

// GetEdge 获取边
func GetEdge(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	database := vars["database"]
	schema := vars["schema"]
	table := vars["table"]
	edgeID := vars["id"]

	// 获取适配器
	adapter := config.PrestConf.Adapter

	// 执行获取边
	sc := adapter.GetEdge(database, schema, table, edgeID)
	if err := sc.Err(); err != nil {
		jsonError(w, fmt.Sprintf("Failed to get edge: %v", err), http.StatusBadRequest)
		return
	}

	// 解析结果（适配器返回数组格式）
	var results []map[string]interface{}
	if err := json.Unmarshal(sc.Bytes(), &results); err != nil {
		jsonError(w, fmt.Sprintf("Failed to parse result: %v", err), http.StatusBadRequest)
		return
	}

	// 检查是否找到边
	if len(results) == 0 {
		jsonError(w, fmt.Sprintf("Edge with ID %v not found", edgeID), http.StatusNotFound)
		return
	}

	// 获取第一个结果（根据ID查询应该只有一行）
	result := results[0]

	// 构建边响应
	edge := EdgeResponse{
		ID:         result["id"],
		From:       result["from"],
		To:         result["to"],
		Label:      getString(result, "label"),
		Properties: getMap(result, "properties"),
	}

	// 返回JSON响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(edge)
}

// UpdateVertex 更新顶点属性
func UpdateVertex(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	database := vars["database"]
	schema := vars["schema"]
	table := vars["table"]
	vertexID := vars["id"]

	// 解析请求体
	type UpdateVertexRequest struct {
		Properties map[string]interface{} `json:"properties"`
	}
	var req UpdateVertexRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// 验证必填字段
	if req.Properties == nil {
		jsonError(w, "properties is required", http.StatusBadRequest)
		return
	}

	// 获取适配器
	adapter := config.PrestConf.Adapter

	// 执行更新顶点
	sc := adapter.UpdateVertex(database, schema, table, vertexID, req.Properties)
	if err := sc.Err(); err != nil {
		jsonError(w, fmt.Sprintf("Failed to update vertex: %v", err), http.StatusBadRequest)
		return
	}

	// 返回成功响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   fmt.Sprintf("Vertex %v updated successfully", vertexID),
		"vertex_id": vertexID,
		"properties": req.Properties,
	})
}

// UpdateEdge 更新边属性
func UpdateEdge(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	database := vars["database"]
	schema := vars["schema"]
	table := vars["table"]
	edgeID := vars["id"]

	// 解析请求体
	type UpdateEdgeRequest struct {
		Properties map[string]interface{} `json:"properties"`
	}
	var req UpdateEdgeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// 验证必填字段
	if req.Properties == nil {
		jsonError(w, "properties is required", http.StatusBadRequest)
		return
	}

	// 获取适配器
	adapter := config.PrestConf.Adapter

	// 执行更新边
	sc := adapter.UpdateEdge(database, schema, table, edgeID, req.Properties)
	if err := sc.Err(); err != nil {
		jsonError(w, fmt.Sprintf("Failed to update edge: %v", err), http.StatusBadRequest)
		return
	}

	// 返回成功响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": fmt.Sprintf("Edge %v updated successfully", edgeID),
		"edge_id": edgeID,
		"properties": req.Properties,
	})
}

// DeleteVertex 删除顶点
func DeleteVertex(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	database := vars["database"]
	schema := vars["schema"]
	table := vars["table"]
	vertexID := vars["id"]

	// 获取适配器
	adapter := config.PrestConf.Adapter

	// 执行删除顶点
	sc := adapter.DeleteVertex(database, schema, table, vertexID)
	if err := sc.Err(); err != nil {
		jsonError(w, fmt.Sprintf("Failed to delete vertex: %v", err), http.StatusBadRequest)
		return
	}

	// 返回成功响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   fmt.Sprintf("Vertex %v deleted successfully", vertexID),
		"vertex_id": vertexID,
	})
}

// DeleteEdge 删除边
func DeleteEdge(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	database := vars["database"]
	schema := vars["schema"]
	table := vars["table"]
	edgeID := vars["id"]

	// 获取适配器
	adapter := config.PrestConf.Adapter

	// 执行删除边
	sc := adapter.DeleteEdge(database, schema, table, edgeID)
	if err := sc.Err(); err != nil {
		jsonError(w, fmt.Sprintf("Failed to delete edge: %v", err), http.StatusBadRequest)
		return
	}

	// 返回成功响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": fmt.Sprintf("Edge %v deleted successfully", edgeID),
		"edge_id": edgeID,
	})
}

// parseGraphQLQuery 解析GraphQL-like查询字符串，转换为queryType和queryParams
func parseGraphQLQuery(query string) (queryType string, params map[string]interface{}, err error) {
	params = make(map[string]interface{})

	// 简单解析：识别查询类型和参数
	// 移除空格和换行
	query = strings.TrimSpace(query)

	// 检查是否为空
	if query == "" {
		return "", nil, fmt.Errorf("empty query")
	}

	// 简单模式匹配：根据关键字识别查询类型
	switch {
	case strings.Contains(query, "traverse"):
		queryType = "traverse"
		// 提取参数：from, depth, direction
		if matches := regexp.MustCompile(`from:\s*(\d+)`).FindStringSubmatch(query); len(matches) > 1 {
			if val, err := strconv.Atoi(matches[1]); err == nil {
				params["start_vertex"] = val
			}
		}
		if matches := regexp.MustCompile(`depth:\s*(\d+)`).FindStringSubmatch(query); len(matches) > 1 {
			if val, err := strconv.Atoi(matches[1]); err == nil {
				params["max_depth"] = val
			}
		}
		if matches := regexp.MustCompile(`direction:\s*"(\w+)"`).FindStringSubmatch(query); len(matches) > 1 {
			params["direction"] = matches[1]
		}

	case strings.Contains(query, "path"):
		queryType = "path"
		// 提取参数：from, to, maxDepth
		if matches := regexp.MustCompile(`from:\s*(\d+)`).FindStringSubmatch(query); len(matches) > 1 {
			if val, err := strconv.Atoi(matches[1]); err == nil {
				params["start_vertex"] = val
			}
		}
		if matches := regexp.MustCompile(`to:\s*(\d+)`).FindStringSubmatch(query); len(matches) > 1 {
			if val, err := strconv.Atoi(matches[1]); err == nil {
				params["end_vertex"] = val
			}
		}
		if matches := regexp.MustCompile(`maxDepth:\s*(\d+)`).FindStringSubmatch(query); len(matches) > 1 {
			if val, err := strconv.Atoi(matches[1]); err == nil {
				params["max_depth"] = val
			}
		}

	case strings.Contains(query, "neighbors"):
		queryType = "neighbors"
		// 提取参数：vertex, direction
		if matches := regexp.MustCompile(`vertex:\s*(\d+)`).FindStringSubmatch(query); len(matches) > 1 {
			if val, err := strconv.Atoi(matches[1]); err == nil {
				params["start_vertex"] = val
			}
		}
		if matches := regexp.MustCompile(`direction:\s*"(\w+)"`).FindStringSubmatch(query); len(matches) > 1 {
			params["direction"] = matches[1]
		}

	case strings.Contains(query, "match"):
		queryType = "match"
		// 匹配查询，提取模式参数
		// 简单实现：返回所有顶点和边
		params["limit"] = 100

	case strings.Contains(query, "vertex"):
		queryType = "vertex"
		// 单个顶点查询
		if matches := regexp.MustCompile(`vertex\(\s*id:\s*(\d+)\s*\)`).FindStringSubmatch(query); len(matches) > 1 {
			if val, err := strconv.Atoi(matches[1]); err == nil {
				params["vertex_id"] = val
			}
		}

	case strings.Contains(query, "edge"):
		queryType = "edge"
		// 单个边查询
		if matches := regexp.MustCompile(`edge\(\s*id:\s*(\d+)\s*\)`).FindStringSubmatch(query); len(matches) > 1 {
			if val, err := strconv.Atoi(matches[1]); err == nil {
				params["edge_id"] = val
			}
		}

	default:
		// 默认匹配查询
		queryType = "match"
		params["limit"] = 100
	}

	return queryType, params, nil
}