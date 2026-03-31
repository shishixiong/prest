#!/bin/bash
#set -euo pipefail

# ============================================
# GaussDB pREST REST API 全面测试脚本
# ============================================
#
# 这个脚本测试pREST服务的所有主要REST API功能
# 针对GaussDB适配器进行验证
#
# 使用方法:
#   ./test_gaussdb_api.sh [BASE_URL]
#
# 参数:
#   BASE_URL - pREST服务地址 (默认: http://localhost:3000)
#
# 依赖:
#   - curl
#   - jq (可选，用于美化JSON输出)
# ============================================

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 统计变量
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

# 配置
BASE_URL="${1:-http://localhost:3000}"
CONTENT_TYPE="Content-Type: application/json"
TIMEOUT=10
VERBOSE="${VERBOSE:-false}"
SKIP_CLEANUP="${SKIP_CLEANUP:-false}"

# 测试表名
TEST_TABLE="test_users"
TEST_JSON_TABLE="test_json_data"
TEST_SCHEMA="public"
TEST_DATABASE="app_db"

# 向量测试表名
TEST_VECTOR_TABLE="test_vector_data"
VECTOR_FIELD="embedding"
VECTOR_DIMENSION=3

# Graph测试配置
TEST_GRAPH_NAME="test_graph_api"

# 临时文件
RESPONSE_FILE=$(mktemp)
HEADERS_FILE=$(mktemp)

# 清理函数
cleanup() {
    if [[ "$SKIP_CLEANUP" == "true" ]]; then
        echo -e "${YELLOW}[INFO] 跳过清理临时文件${NC}"
        return
    fi
    rm -f "$RESPONSE_FILE" "$HEADERS_FILE" 2>/dev/null || true
}

# 错误处理
trap cleanup EXIT
trap 'echo -e "\n${RED}[ERROR] 脚本被中断${NC}"; cleanup; exit 1' INT TERM

# 工具函数
print_header() {
    echo -e "\n${BLUE}=== $1 ===${NC}"
}

print_success() {
    echo -e "${GREEN}[✓] $1${NC}"
    ((PASSED_TESTS++)) || true
    ((TOTAL_TESTS++)) || true
}

print_failure() {
    echo -e "${RED}[✗] $1${NC}"
    ((FAILED_TESTS++)) || true
    ((TOTAL_TESTS++)) || true
}

print_skip() {
    echo -e "${YELLOW}[~] $1${NC}"
    ((SKIPPED_TESTS++)) || true
    ((TOTAL_TESTS++)) || true
}

print_info() {
    echo -e "${YELLOW}[INFO] $1${NC}"
}

# 检查命令是否存在
check_command() {
    if ! command -v "$1" &> /dev/null; then
        echo -e "${RED}[ERROR] 需要命令: $1${NC}"
        exit 1
    fi
}

# 发送HTTP请求
send_request() {
    local method="$1"
    local url="$2"
    local data="$3"
    local expected_status="${4:-200}"
    local description="$5"

    # 构建curl命令数组
    local curl_cmd=(curl -s -X "$method" "$url" -H "$CONTENT_TYPE")

    if [[ -n "$data" && "$data" != "null" ]]; then
        curl_cmd+=(-d "$data")
    fi

    curl_cmd+=(--connect-timeout "$TIMEOUT" -w "%{http_code}" -o "$RESPONSE_FILE")

    if [[ "$VERBOSE" == "true" ]]; then
        echo -e "${YELLOW}[DEBUG] 执行: ${curl_cmd[*]}${NC}"
    fi

    local status_code
    # 执行命令并捕获状态码
    status_code=$("${curl_cmd[@]}" 2>/dev/null || echo "000")

    if [[ "$status_code" == "$expected_status" ]]; then
        print_success "$description (状态码: $status_code)"
        return 0
    else
        print_failure "$description (期望: $expected_status, 实际: $status_code)"
        if [[ "$VERBOSE" == "true" && -f "$RESPONSE_FILE" ]]; then
            echo -e "${YELLOW}[DEBUG] 响应内容:${NC}"
            cat "$RESPONSE_FILE" | jq . 2>/dev/null || cat "$RESPONSE_FILE"
        fi
        return 1
    fi
}

# 检查JSON响应是否包含特定字段
check_json_field() {
    local field="$1"
    local expected_value="$2"

    if [[ ! -f "$RESPONSE_FILE" ]]; then
        return 1
    fi

    local actual_value
    actual_value=$(jq -r "$field" "$RESPONSE_FILE" 2>/dev/null)

    if [[ "$actual_value" == "$expected_value" ]]; then
        return 0
    else
        if [[ "$VERBOSE" == "true" ]]; then
            echo -e "${YELLOW}[DEBUG] 字段检查失败: $field (期望: $expected_value, 实际: $actual_value)${NC}"
        fi
        return 1
    fi
}

# 检查响应是否为有效JSON
check_valid_json() {
    if [[ ! -f "$RESPONSE_FILE" ]]; then
        return 1
    fi

    if jq empty "$RESPONSE_FILE" 2>/dev/null; then
        return 0
    else
        return 1
    fi
}

# 检查向量表是否存在和vector扩展支持
check_vector_support() {
    print_header "检查向量功能支持"

    # 检查向量表是否存在
    if curl -s -f "$BASE_URL/show/$TEST_DATABASE/$TEST_SCHEMA/$TEST_VECTOR_TABLE" --connect-timeout 5 &> /dev/null; then
        print_success "向量测试表 $TEST_VECTOR_TABLE 存在"
        return 0
    else
        print_skip "向量测试表 $TEST_VECTOR_TABLE 不存在，跳过向量功能测试"
        return 1
    fi
}

# 等待服务就绪
wait_for_service() {
    print_header "等待pREST服务就绪"

    local max_attempts=30
    local attempt=1

    while [[ $attempt -le $max_attempts ]]; do
        if curl -s -f "$BASE_URL/_health" --connect-timeout 2 &> /dev/null; then
            print_success "服务已就绪 ($BASE_URL)"
            return 0
        fi

        echo -e "${YELLOW}[$attempt/$max_attempts] 等待服务启动...${NC}"
        sleep 2
        ((attempt++)) || true
    done

    print_failure "服务在 $max_attempts 次尝试后仍未就绪"
    return 1
}

# 检查测试表是否存在
check_tables() {
    print_header "检查测试表"

    # 检查 test_users 表
    if curl -s -f "$BASE_URL/show/$TEST_DATABASE/$TEST_SCHEMA/$TEST_TABLE" --connect-timeout 5 &> /dev/null; then
        print_success "测试表 $TEST_TABLE 存在"
    else
        print_failure "测试表 $TEST_TABLE 不存在或不可访问"
        echo -e "${YELLOW}"
        echo "============================================"
        echo "  需要先创建测试表"
        echo "============================================"
        echo "请执行以下命令创建测试表:"
        echo ""
        echo "  gsql -h localhost -p 18888 -U gaussdb -d prest \\"
        echo "       -f setup_gaussdb_test_tables.sql"
        echo ""
        echo "或者使用其他数据库客户端执行 SQL 脚本。"
        echo "SQL 脚本位置: $(pwd)/setup_gaussdb_test_tables.sql"
        echo -e "${NC}"
        return 1
    fi

    # 检查 test_json_data 表
    if curl -s -f "$BASE_URL/show/$TEST_DATABASE/$TEST_SCHEMA/$TEST_JSON_TABLE" --connect-timeout 5 &> /dev/null; then
        print_success "测试表 $TEST_JSON_TABLE 存在"
    else
        print_info "测试表 $TEST_JSON_TABLE 不存在，部分JSON测试将跳过"
        # 不视为错误，部分测试可以跳过
    fi

    return 0
}

# ============================================
# 测试函数
# ============================================

test_system_info() {
    print_header "1. 系统信息查询测试"

    # 1.1 健康检查
    send_request "GET" "$BASE_URL/_health" null 200 "健康检查"

    # 1.2 数据库列表
    send_request "GET" "$BASE_URL/databases" null 200 "数据库列表"

    # 1.3 模式列表
    send_request "GET" "$BASE_URL/schemas" null 200 "模式列表"

    # 1.4 表列表
    send_request "GET" "$BASE_URL/tables" null 200 "所有表列表"

    # 1.5 特定数据库的模式列表
    send_request "GET" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA" null 200 "数据库模式列表"

    # 1.6 查看表结构
    send_request "GET" "$BASE_URL/show/$TEST_DATABASE/$TEST_SCHEMA/$TEST_TABLE" null 200 "查看表结构: $TEST_TABLE"
}

test_crud_operations() {
    print_header "2. CRUD操作测试"

    local test_data='{"name":"测试用户","email":"test_crud@example.com","age":28,"salary":6000.00}'
    local update_data='{"name":"更新用户","age":29,"salary":6500.00}'
    local patch_data='{"salary":7000.00}'

    # 2.1 插入数据
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_TABLE" "$test_data" 201 "插入测试数据"

    # 2.2 查询插入的数据
    send_request "GET" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_TABLE?email=\$eq.test_crud@example.com" null 200 "查询插入的数据"

    # 2.3 更新数据 (PUT)
    send_request "PUT" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_TABLE?email=\$eq.test_crud@example.com" "$update_data" 200 "全量更新数据"

    # 2.4 部分更新 (PATCH)
    send_request "PATCH" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_TABLE?email=\$eq.test_crud@example.com" "$patch_data" 200 "部分更新数据"

    # 2.5 删除数据
    send_request "DELETE" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_TABLE?email=\$eq.test_crud@example.com" null 200 "删除测试数据"

    # 2.6 验证删除
    send_request "GET" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_TABLE?email=\$eq.test_crud@example.com" null 200 "验证数据已删除"

    # 2.7 批量插入
    local batch_data='[{"name":"批量用户1","email":"batch1@example.com","age":25,"salary":5000},{"name":"批量用户2","email":"batch2@example.com","age":30,"salary":8000}]'
    send_request "POST" "$BASE_URL/batch/$TEST_DATABASE/$TEST_SCHEMA/$TEST_TABLE" "$batch_data" 201 "批量插入数据"

    # 2.8 清理批量数据
    send_request "DELETE" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_TABLE?email=\$in.batch1@example.com,batch2@example.com" null 200 "清理批量数据"
}

test_query_features() {
    print_header "3. 查询功能测试"

    # 先插入一些测试数据
    local sample_data='[
        {"name":"张三_test","email":"zhangsan_test@example.com","age":25,"salary":5000},
        {"name":"李四_test","email":"lisi_test@example.com","age":30,"salary":7500},
        {"name":"王五_test","email":"wangwu_test@example.com","age":28,"salary":6200},
        {"name":"赵六_test","email":"zhaoliu_test@example.com","age":35,"salary":8500},
        {"name":"孙七_test","email":"sunqi_test@example.com","age":22,"salary":4500}
    ]'

    send_request "POST" "$BASE_URL/batch/$TEST_DATABASE/$TEST_SCHEMA/$TEST_TABLE" "$sample_data" 201 "插入查询测试数据"

    # 3.1 基本查询
    send_request "GET" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_TABLE" null 200 "基本查询"

    # 3.2 过滤条件 (大于)
    send_request "GET" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_TABLE?age=\$gt.25" null 200 "过滤: 年龄大于25"

    # 3.3 过滤条件 (小于)
    send_request "GET" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_TABLE?salary=\$lt.7000" null 200 "过滤: 工资小于7000"

    # 3.4 多条件过滤
    send_request "GET" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_TABLE?age=\$gt.25&salary=\$lt.8000" null 200 "多条件过滤"

    # 3.5 排序
    send_request "GET" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_TABLE?_order=-age" null 200 "按年龄降序排序"

    # 3.6 分页
    send_request "GET" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_TABLE?_limit=2&_offset=1" null 200 "分页查询 (limit=2, offset=1)"

    # 3.7 字段选择
    send_request "GET" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_TABLE?_select=id,name,age" null 200 "选择特定字段"

    # 3.8 计数查询
    send_request "GET" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_TABLE?_count=*" null 200 "计数查询"

    # 3.9 清理测试数据
    send_request "DELETE" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_TABLE?email=\$in.zhangsan_test@example.com,lisi_test@example.com,wangwu_test@example.com,zhaoliu_test@example.com,sunqi_test@example.com" null 200 "清理查询测试数据"
}

test_json_features() {
    print_header "4. JSON功能测试"

    # 检查JSON测试表是否存在
    if ! curl -s -f "$BASE_URL/show/$TEST_DATABASE/$TEST_SCHEMA/$TEST_JSON_TABLE" --connect-timeout 5 &> /dev/null; then
        print_skip "JSON测试表 $TEST_JSON_TABLE 不存在，跳过JSON功能测试"
        return 0
    fi

    # 4.1 插入JSON数据
    local json_data='{"metadata":"{\"department\":\"IT\",\"role\":\"developer\",\"skills\":[\"Go\",\"PostgreSQL\"]}","tags":["backend","database"]}'
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_JSON_TABLE" "$json_data" 201 "插入JSON数据"

    # 4.2 查询JSON字段
    send_request "GET" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_JSON_TABLE?metadata->>department:jsonb=IT" null 200 "查询JSON字段"

    # # 4.3 数组字段查询
    # send_request "GET" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_JSON_TABLE?tags[0]=backend" null 200 "数组字段查询"

    # 4.4 清理JSON数据
    send_request "DELETE" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_JSON_TABLE?metadata->>department:jsonb=IT" null 200 "清理JSON测试数据"
}

test_error_handling() {
    print_header "5. 错误处理测试"

    # 5.1 不存在的表
    send_request "GET" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/nonexistent_table" null 400 "查询不存在的表"

    # 5.2 不存在的字段
    send_request "GET" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_TABLE?nonexistent_field=\$eq.test" null 400 "查询不存在的字段"

    # 5.3 无效的查询语法
    send_request "GET" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_TABLE?age=\$invalid.25" null 400 "无效的查询语法"

    # 5.4 主键冲突 (需要先插入数据)
    local conflict_data='{"name":"冲突测试","email":"conflict@example.com","age":30,"salary":5000}'
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_TABLE" "$conflict_data" 201 "插入冲突测试数据 (第一次)"
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_TABLE" "$conflict_data" 400 "插入冲突测试数据 (第二次，期望冲突)"

    # 5.5 清理冲突测试数据
    send_request "DELETE" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_TABLE?email=\$eq.conflict@example.com" null 200 "清理冲突测试数据"
}

test_special_cases() {
    print_header "6. 特殊情况测试"

    # 6.1 空表查询
    send_request "GET" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_TABLE" null 200 "查询空表"

    # 6.2 包含特殊字符的数据
    local special_data='{"name":"测试&特殊<字符>","email":"special@example.com","age":99,"salary":9999}'
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_TABLE" "$special_data" 201 "插入包含特殊字符的数据"

    # 6.3 清理特殊字符数据
    send_request "DELETE" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_TABLE?email=\$eq.special@example.com" null 200 "清理特殊字符数据"

    # 6.4 大数字测试
    local big_number_data='{"name":"大数字测试","email":"bignumber@example.com","age":150,"salary":999999.99}'
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_TABLE" "$big_number_data" 201 "插入大数字数据"

    # 6.5 清理大数字数据
    send_request "DELETE" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_TABLE?email=\$eq.bignumber@example.com" null 200 "清理大数字数据"
}

# 向量CRUD操作测试
test_vector_crud_operations() {
    print_header "7. 向量CRUD操作测试"

    local test_vector_data='{
        "title": "向量测试文档",
        "content": "这是一个向量测试文档",
        "embedding": [0.3, 0.4, 0.5],
        "category": "Test",
        "metadata": "{\"tags\": [\"test\", \"vector\"]}",
        "is_active": true
    }'

    local update_data='{
        "embedding": [0.6, 0.7, 0.8],
        "title": "更新后的向量文档"
    }'

    # 7.1 插入向量数据
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_VECTOR_TABLE" "$test_vector_data" 201 "插入向量测试数据"

    # 7.2 查询向量数据
    send_request "GET" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_VECTOR_TABLE?title=\$eq.向量测试文档" null 200 "查询插入的向量数据"

    # 7.3 更新向量数据
    send_request "PATCH" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_VECTOR_TABLE?title=\$eq.向量测试文档" "$update_data" 200 "更新向量数据"

    # 7.4 删除向量数据
    send_request "DELETE" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_VECTOR_TABLE?title=\$eq.更新后的向量文档" null 200 "删除向量测试数据"
}

# 向量相似性搜索测试
test_vector_search() {
    print_header "8. 向量相似性搜索测试"

    # 8.1 余弦距离搜索
    local cosine_search_data='{
        "query_vector": [0.1, 0.2, 0.3],
        "distance_metric": "cosine",
        "limit": 3,
        "vector_field": "embedding",
        "include_distance": true
    }'
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_VECTOR_TABLE/_vector/search" "$cosine_search_data" 200 "余弦距离向量搜索"

    # 8.2 欧氏距离搜索
    local l2_search_data='{
        "query_vector": [0.1, 0.2, 0.3],
        "distance_metric": "l2",
        "vector_field": "embedding",
        "limit": 2
    }'
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_VECTOR_TABLE/_vector/search" "$l2_search_data" 200 "欧氏距离向量搜索"

    # 8.3 hannming距离搜索
#    local inner_search_data='{
#        "query_vector": [0.1, 0.2, 0.3],
#        "distance_metric": "inner",
#        "vector_field": "embedding",
#        "limit": 2
#    }'
#    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_VECTOR_TABLE/_vector/search" "$inner_search_data" 200 "内积距离向量搜索"
}

# 混合查询测试（向量搜索+传统过滤条件）
test_hybrid_search() {
    print_header "9. 混合查询测试（向量+过滤条件）"

    # 9.1 向量搜索 + 类别过滤
    local category_filter_data='{
        "query_vector": [0.1, 0.2, 0.3],
        "distance_metric": "cosine",
        "vector_field": "embedding",
        "limit": 5,
        "where": {
            "category": "AI"
        }
    }'
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_VECTOR_TABLE/_vector/search" "$category_filter_data" 200 "向量搜索+类别过滤"

    # 9.2 向量搜索 + 多条件过滤
    local multi_filter_data='{
        "query_vector": [0.1, 0.2, 0.3],
        "distance_metric": "cosine",
        "vector_field": "embedding",
        "limit": 5,
        "where": {
            "category": "AI",
            "is_active": true
        }
    }'
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_VECTOR_TABLE/_vector/search" "$multi_filter_data" 200 "向量搜索+多条件过滤"

    # 9.3 向量搜索 + JSON字段过滤
    local json_filter_data='{
        "query_vector": [0.1, 0.2, 0.3],
        "distance_metric": "cosine",
        "vector_field": "embedding",
        "limit": 5,
        "where": {
            "metadata->>difficulty:jsonb": "advanced"
        }
    }'
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_VECTOR_TABLE/_vector/search" "$json_filter_data" 200 "向量搜索+JSON字段过滤"
}

# 向量索引操作测试
test_vector_index_operations() {
    print_header "10. 向量索引操作测试"

    # 10.1 创建GSDISKANN索引
    local hnsw_index_data='{
        "index_type": "gsdiskann",
        "vector_field": "embedding",
        "distance_metric": "cosine",
        "pq_nseg" : 3,
        "pq_nclus" : 16,
        "queue_size" : 100,
        "num_parallels" : 30,
        "enable_pq" : true,
        "using_clustering_for_parallel" : false
    }'
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_VECTOR_TABLE/_vector/index" "$hnsw_index_data" 200 "创建GSDISKANN向量索引"

    # 10.2 删除向量索引
    local delete_index_data='{
        "vector_field": "embedding"
    }'
    send_request "DELETE" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_VECTOR_TABLE/_vector/index" "$delete_index_data" 200 "删除向量索引"

    # 10.3 创建GSIVFFLAT索引
    local ivfflat_index_data='{
        "index_type": "gsivfflat",
        "vector_field": "embedding",
        "distance_metric": "cosine",
        "lists": 100
    }'
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_VECTOR_TABLE/_vector/index" "$ivfflat_index_data" 200 "创建GSIVFFLAT向量索引"

    # 10.4 清理：删除IVFFlat索引
    send_request "DELETE" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_VECTOR_TABLE/_vector/index" "$delete_index_data" 200 "清理GSIVFFLAT索引"
}

# 向量错误场景测试
test_vector_error_scenarios() {
    print_header "11. 向量错误场景测试"

    # 11.1 无效向量格式
    local invalid_vector_data='{
        "query_vector": [0.1, "not-a-number", 0.3],
        "distance_metric": "cosine"
    }'
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_VECTOR_TABLE/_vector/search" "$invalid_vector_data" 400 "无效向量格式"

    # 11.2 无效距离度量
    local invalid_metric_data='{
        "query_vector": [0.1, 0.2, 0.3],
        "distance_metric": "invalid_metric"
    }'
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_VECTOR_TABLE/_vector/search" "$invalid_metric_data" 400 "无效距离度量"

    # 11.3 缺失查询向量
    local missing_vector_data='{
        "distance_metric": "cosine"
    }'
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_VECTOR_TABLE/_vector/search" "$missing_vector_data" 400 "缺失查询向量"

    # 11.4 不存在的向量字段
    local invalid_field_data='{
        "query_vector": [0.1, 0.2, 0.3],
        "distance_metric": "cosine",
        "vector_field": "nonexistent_field"
    }'
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_VECTOR_TABLE/_vector/search" "$invalid_field_data" 400 "不存在的向量字段"
}

# 向量功能测试主函数
test_vector_features() {
    if ! check_vector_support; then
        return 0
    fi

    test_vector_crud_operations
    test_vector_search
    test_hybrid_search
    test_vector_index_operations
    test_vector_error_scenarios
}

# ============================================
# Graph功能测试
# ============================================

# Graph CRUD操作测试
test_graph_crud_operations() {
    print_header "12. Graph CRUD操作测试"

    # 12.1 创建Graph (property类型)
    local create_graph_data="{
        \"graph_name\": \"${TEST_GRAPH_NAME}_property\",
        \"graph_type\": \"property\"
    }"
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_GRAPH_NAME/_graph/create" "$create_graph_data" 200 "创建Property Graph"

    # 12.2 创建Graph (directed类型)
    local create_directed_graph_data="{
        \"graph_name\": \"${TEST_GRAPH_NAME}_directed\",
        \"graph_type\": \"directed\"
    }"
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_GRAPH_NAME/_graph/create" "$create_directed_graph_data" 200 "创建Directed Graph"

    # 12.3 添加顶点到Property Graph
    local add_vertices_data="{
        \"vertices\": [
            {\"id\": 1, \"label\": \"user\", \"properties\": {\"name\": \"Alice\", \"age\": 30}},
            {\"id\": 2, \"label\": \"user\", \"properties\": {\"name\": \"Bob\", \"age\": 25}},
            {\"id\": 3, \"label\": \"user\", \"properties\": {\"name\": \"Charlie\", \"age\": 35}},
            {\"id\": 4, \"label\": \"product\", \"properties\": {\"name\": \"Laptop\", \"price\": 999.99}},
            {\"id\": 5, \"label\": \"product\", \"properties\": {\"name\": \"Mouse\", \"price\": 29.99}}
        ]
    }"
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/${TEST_GRAPH_NAME}_property/_graph/vertices" "$add_vertices_data" 200 "添加顶点到Property Graph"

    # 12.4 添加边到Property Graph
    local add_edges_data="{
        \"edges\": [
            {\"id\": 1, \"from\": 1, \"to\": 2, \"label\": \"knows\", \"properties\": {\"since\": \"2020\"}},
            {\"id\": 2, \"from\": 1, \"to\": 3, \"label\": \"knows\", \"properties\": {\"since\": \"2019\"}},
            {\"id\": 3, \"from\": 2, \"to\": 4, \"label\": \"bought\", \"properties\": {\"date\": \"2024-01-15\"}},
            {\"id\": 4, \"from\": 3, \"to\": 5, \"label\": \"bought\", \"properties\": {\"date\": \"2024-02-20\"}},
            {\"id\": 5, \"from\": 1, \"to\": 4, \"label\": \"viewed\", \"properties\": {\"date\": \"2024-03-01\"}}
        ]
    }"
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/${TEST_GRAPH_NAME}_property/_graph/edges" "$add_edges_data" 200 "添加边到Property Graph"

    # 12.5 添加顶点到Directed Graph
    local add_vertices_directed_data="{
        \"vertices\": [
            {\"id\": 1, \"label\": \"nodeA\", \"properties\": {\"value\": 100}},
            {\"id\": 2, \"label\": \"nodeB\", \"properties\": {\"value\": 200}},
            {\"id\": 3, \"label\": \"nodeC\", \"properties\": {\"value\": 300}}
        ]
    }"
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/${TEST_GRAPH_NAME}_directed/_graph/vertices" "$add_vertices_directed_data" 200 "添加顶点到Directed Graph"

    # 12.6 添加边到Directed Graph
    local add_edges_directed_data="{
        \"edges\": [
            {\"id\": 1, \"from\": 1, \"to\": 2, \"label\": \"depends_on\", \"properties\": {\"weight\": 5}},
            {\"id\": 2, \"from\": 2, \"to\": 3, \"label\": \"depends_on\", \"properties\": {\"weight\": 10}}
        ]
    }"
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/${TEST_GRAPH_NAME}_directed/_graph/edges" "$add_edges_directed_data" 200 "添加边到Directed Graph"
}

# Graph查询测试
test_graph_queries() {
    print_header "13. Graph查询测试"

    # 13.1 查询所有顶点 (match query)
    local match_query_data='{
        "query_type": "match"
    }'
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/${TEST_GRAPH_NAME}_property/_graph/query" "$match_query_data" 200 "Match查询-获取所有顶点"

    # 13.2 遍历查询 (traverse from vertex 1)
    local traverse_query_data='{
        "query_type": "traverse",
        "start_vertex": 1,
        "max_depth": 2,
        "direction": "out"
    }'
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/${TEST_GRAPH_NAME}_property/_graph/query" "$traverse_query_data" 200 "Traverse查询-从顶点1出发"

    # 13.3 邻居查询 (neighbors of vertex 1)
    local neighbors_query_data='{
        "query_type": "neighbors",
        "start_vertex": 1,
        "direction": "out"
    }'
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/${TEST_GRAPH_NAME}_property/_graph/query" "$neighbors_query_data" 200 "Neighbors查询-获取顶点1的邻居"

    # 13.4 路径查询 (path from vertex 1 to vertex 3)
    local path_query_data='{
        "query_type": "path",
        "start_vertex": 1,
        "end_vertex": 3,
        "max_depth": 5
    }'
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/${TEST_GRAPH_NAME}_property/_graph/query" "$path_query_data" 200 "Path查询-从顶点1到顶点3"

    # 13.5 GraphQL-like查询语法
    local graphql_query_data='{
        "query": "traverse(from: 1, depth: 2, direction: \"out\")"
    }'
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/${TEST_GRAPH_NAME}_property/_graph/query" "$graphql_query_data" 200 "GraphQL语法查询"

    # 13.6 带过滤条件的查询
    local filtered_query_data='{
        "query_type": "neighbors",
        "start_vertex": 1,
        "direction": "out",
        "filters": {"label": "user"}
    }'
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/${TEST_GRAPH_NAME}_property/_graph/query" "$filtered_query_data" 200 "带过滤条件的邻居查询"
}

# Graph错误场景测试
test_graph_error_scenarios() {
    print_header "14. Graph错误场景测试"

    # 14.1 查询不存在的图
    local nonexistent_graph_data='{
        "query_type": "match"
    }'
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/nonexistent_graph/_graph/query" "$nonexistent_graph_data" 400 "查询不存在的图"

    # 14.2 traverse查询缺少start_vertex
    local missing_start_data='{
        "query_type": "traverse"
    }'
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/${TEST_GRAPH_NAME}_property/_graph/query" "$missing_start_data" 400 "Traverse缺少start_vertex"

    # 14.3 path查询缺少end_vertex
    local missing_end_data='{
        "query_type": "path",
        "start_vertex": 1
    }'
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/${TEST_GRAPH_NAME}_property/_graph/query" "$missing_end_data" 400 "Path缺少end_vertex"

    # 14.4 创建图时缺少graph_name（使用table作为默认名）
    local create_no_name_data="{
        \"graph_type\": \"property\"
    }"
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/test_table_for_graph/_graph/create" "$create_no_name_data" 200 "创建图(使用默认表名)"

    # 14.5 添加空顶点数组
    local empty_vertices_data='{
        "vertices": []
    }'
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/${TEST_GRAPH_NAME}_property/_graph/vertices" "$empty_vertices_data" 400 "添加空顶点数组"

    # 14.6 添加空边数组
    local empty_edges_data='{
        "edges": []
    }'
    send_request "POST" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/${TEST_GRAPH_NAME}_property/_graph/edges" "$empty_edges_data" 400 "添加空边数组"
}

# Graph清理测试
test_graph_cleanup() {
    print_header "15. Graph清理测试"

    # 15.1 删除Property Graph
    send_request "DELETE" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/${TEST_GRAPH_NAME}_property/_graph" null 200 "删除Property Graph"

    # 15.2 删除Directed Graph
    send_request "DELETE" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/${TEST_GRAPH_NAME}_directed/_graph" null 200 "删除Directed Graph"

    # 15.3 删除默认表名创建的图
    send_request "DELETE" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/test_table_for_graph/_graph" null 200 "删除默认表名Graph"
}

# Graph功能测试主函数
test_graph_features() {
    test_graph_crud_operations
    test_graph_queries
    test_graph_error_scenarios
    test_graph_cleanup
}

# ============================================
# 主函数
# ============================================

main() {
    echo -e "${BLUE}"
    echo "============================================"
    echo "  GaussDB pREST REST API 测试脚本"
    echo "============================================"
    echo -e "${NC}"
    echo "服务地址: $BASE_URL"
    echo "测试数据库: $TEST_DATABASE"
    echo "测试模式: $TEST_SCHEMA"
    echo "测试表: $TEST_TABLE, $TEST_JSON_TABLE, $TEST_VECTOR_TABLE"
    echo "测试图: ${TEST_GRAPH_NAME}_property, ${TEST_GRAPH_NAME}_directed"
    echo ""

    # 检查依赖
    check_command curl
    if command -v jq &> /dev/null; then
        print_success "找到 jq 命令 (JSON处理)"
    else
        print_info "未找到 jq 命令，JSON输出将不被格式化"
    fi

    # 等待服务就绪
    if ! wait_for_service; then
        echo -e "${RED}[ERROR] 无法连接到pREST服务，请检查服务是否运行${NC}"
        exit 1
    fi

    # # 检查测试表
    # if ! check_tables; then
    #     echo -e "${YELLOW}[WARN] 测试表检查失败，继续运行测试可能会失败${NC}"
    #     read -p "是否继续？ (y/N): " -n 1 -r
    #     echo
    #     if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    #         exit 1
    #     fi
    # fi

    # 运行测试
    test_system_info
    test_crud_operations
    test_query_features
    test_json_features
    test_error_handling
    test_special_cases

    # 运行向量测试（如果向量表存在且支持）
    test_vector_features

    # 运行Graph测试
#    test_graph_features

    # 打印总结
    print_header "测试总结"
    echo -e "${BLUE}总计测试: $TOTAL_TESTS${NC}"
    echo -e "${GREEN}通过: $PASSED_TESTS${NC}"
    echo -e "${RED}失败: $FAILED_TESTS${NC}"
    echo -e "${YELLOW}跳过: $SKIPPED_TESTS${NC}"

    if [[ $FAILED_TESTS -eq 0 ]]; then
        echo -e "\n${GREEN}🎉 所有测试通过！${NC}"
        return 0
    else
        echo -e "\n${RED}⚠️  有 $FAILED_TESTS 个测试失败${NC}"
        return 1
    fi
}

# 运行主函数
main "$@"
