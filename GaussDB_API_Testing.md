# GaussDB pREST REST API 测试指南

本文档提供使用测试脚本验证GaussDB pREST REST API功能的完整指南。

## 前置条件

### 1. 环境要求
- 运行中的pREST服务（端口3000）
- GaussDB实例（端口18888）
- 已创建测试数据库 `prest`
- curl 命令行工具
- jq（可选，用于美化JSON输出）

### 2. 测试表准备

在运行测试前，需要先创建测试表。执行以下命令：

```bash
# 连接到GaussDB并创建测试表
docker exec gaussdb bash -c 'su - omm -c "gsql -d postgres -U gaussdb -W XXXX -f  setup_gaussdb_test_tables.sql"
```

或者手动执行SQL语句：

```sql
-- 创建测试表
CREATE TABLE test_users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE,
    age INTEGER,
    salary DECIMAL(10,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT true
);

CREATE TABLE test_json_data (
    id SERIAL PRIMARY KEY,
    metadata JSONB,
    tags TEXT[],
    settings JSONB DEFAULT '{"enabled": true}'
);
```

## 测试脚本使用

### 1. 基本用法

```bash
# 授予执行权限
chmod +x test_gaussdb_api.sh

# 运行测试（默认地址：http://localhost:3000）
./test_gaussdb_api.sh

# 指定pREST服务地址
./test_gaussdb_api.sh http://localhost:3000
```

### 2. 环境变量配置

```bash
# 启用详细输出
export VERBOSE=true

# 跳过清理临时文件
export SKIP_CLEANUP=true

# 运行测试
./test_gaussdb_api.sh
```

## 测试覆盖范围

### 1. 系统信息查询 ✅
- 健康检查 (`GET /_health`)
- 数据库列表 (`GET /databases`)
- 模式列表 (`GET /schemas`)
- 表列表 (`GET /tables`)
- 表结构查询 (`GET /show/{database}/{schema}/{table}`)

### 2. CRUD操作测试 ✅
- 数据插入 (`POST /{database}/{schema}/{table}`)
- 数据查询 (`GET /{database}/{schema}/{table}`)
- 数据更新 (`PUT /{database}/{schema}/{table}`)
- 数据部分更新 (`PATCH /{database}/{schema}/{table}`)
- 数据删除 (`DELETE /{database}/{schema}/{table}`)
- 批量插入 (`POST /batch/{database}/{schema}/{table}`)

### 3. 查询功能测试 ✅
- 过滤条件 (`?field=operator.value`)
- 多条件过滤 (`?field1=op1.val1&field2=op2.val2`)
- 排序 (`?_order=field.direction`)
- 分页 (`?_limit=N&_offset=M`)
- 字段选择 (`?_select=field1,field2`)
- 计数查询 (`?_count=*`)

### 4. JSON功能测试 ✅
- JSON字段插入和查询
- JSON字段过滤 (`?jsonfield->>key=eq.value`)
- 数组字段查询 (`?arrayfield=cs.{value}`)

### 5. 错误处理测试 ✅
- 不存在的表 (404)
- 不存在的字段 (400)
- 无效查询语法 (400)
- 主键冲突 (409)

### 6. 特殊情况测试 ✅
- 空表查询
- 特殊字符处理
- 大数字处理

## 测试执行流程

```
1. 检查依赖 (curl, jq)
2. 等待服务就绪
3. 检查测试表是否存在
4. 运行所有测试类别
5. 生成测试报告
```

## 测试输出示例

```
============================================
  GaussDB pREST REST API 测试脚本
============================================

服务地址: http://localhost:3000
测试数据库: prest
测试模式: public
测试表: test_users, test_json_data

[✓] 找到 jq 命令 (JSON处理)

=== 等待pREST服务就绪 ===
[✓] 服务已就绪 (http://localhost:3000)

=== 检查测试表 ===
[✓] 测试表 test_users 存在
[✓] 测试表 test_json_data 存在

=== 1. 系统信息查询测试 ===
[✓] 健康检查 (状态码: 200)
[✓] 数据库列表 (状态码: 200)
...

=== 测试总结 ===
总计测试: 45
通过: 43
失败: 2
跳过: 0

⚠️  有 2 个测试失败
```

## 故障排除

### 1. 服务连接失败
```bash
# 检查pREST服务是否运行
curl http://localhost:3000/_health

# 检查端口占用
netstat -an | grep 3000

# 查看服务日志
journalctl -u prest.service  # systemd服务
# 或查看控制台输出
```

### 2. 数据库连接失败
```bash
# 检查GaussDB连接
gsql -h localhost -p 18888 -U gaussdb -d prest -c "SELECT 1"

# 检查数据库配置
cat prest-dev.toml | grep -A5 -B5 "pg"
```

### 3. 测试表不存在
```bash
# 手动创建表
gsql -h localhost -p 18888 -U gaussdb -d prest -f setup_gaussdb_test_tables.sql

# 验证表是否存在
gsql -h localhost -p 18888 -U gaussdb -d prest -c "\dt public.*"
```

### 4. 权限问题
```bash
# 确保gaussdb用户有足够权限
gsql -h localhost -p 18888 -U gaussdb -d prest -c "GRANT ALL ON ALL TABLES IN SCHEMA public TO gaussdb;"
```

## 自动化测试

### 1. 集成到CI/CD
```yaml
# GitHub Actions 示例
name: GaussDB API Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run GaussDB API Tests
        run: |
          chmod +x test_gaussdb_api.sh
          ./test_gaussdb_api.sh http://localhost:3000
```

### 2. 定时测试
```bash
# 使用cron定时执行
0 2 * * * cd /path/to/prest && ./test_gaussdb_api.sh > /var/log/prest-test.log 2>&1
```

## 扩展测试

### 1. 添加自定义测试
编辑 `test_gaussdb_api.sh`，在相应测试函数中添加新测试用例：

```bash
# 示例：添加新查询测试
send_request "GET" "$BASE_URL/$TEST_DATABASE/$TEST_SCHEMA/$TEST_TABLE?age=gte.18&age=lte.65" null 200 "年龄范围查询"
```

### 2. 性能测试
```bash
# 使用ab进行并发测试
ab -n 1000 -c 10 http://localhost:3000/prest/public/test_users

# 使用wrk进行压力测试
wrk -t4 -c100 -d30s http://localhost:3000/prest/public/test_users
```

## 相关文件

- `test_gaussdb_api.sh` - 主测试脚本
- `setup_gaussdb_test_tables.sql` - 测试表创建脚本
- `prest-dev.toml` - 开发环境配置
- `adapters/gaussdb/` - GaussDB适配器代码

## 支持与反馈

如遇问题或需要新功能，请：
1. 查看pREST服务日志
2. 启用详细模式 (`VERBOSE=true`)
3. 检查测试表状态
4. 提交GitHub Issue

---

**注意**: 测试脚本会修改测试表数据，建议在测试环境中运行。
