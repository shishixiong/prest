-- ============================================
-- GaussDB pREST 测试表创建脚本
-- ============================================
--
-- 在运行测试脚本前，请先执行此脚本创建测试表
--
-- 使用方法:
--   gsql -h localhost -p 18888 -U gaussdb -d prest -f setup_gaussdb_test_tables.sql
--
-- 注意: 确保已连接到正确的数据库 (prest)
-- ============================================

-- 删除已存在的测试表（如果存在）
-- Graph相关表会在删除图时自动清理，这里只清理手动创建的表
DROP TABLE IF EXISTS test_vector_data;
DROP TABLE IF EXISTS test_json_data;
DROP TABLE IF EXISTS test_users;

-- Graph测试用的元数据表（手动创建以便SHOW命令可以查询）
-- 注意：实际的graph vertices/edges表会在调用_create_graph API时自动创建
DROP TABLE IF EXISTS prest_graph_metadata;

CREATE TABLE IF NOT EXISTS prest_graph_metadata (
    graph_name TEXT PRIMARY KEY,
    graph_type TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建用户测试表
CREATE TABLE test_users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE,
    age INTEGER,
    salary DECIMAL(10,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT true
);

-- 创建注释（GaussDB兼容）
COMMENT ON TABLE test_users IS 'pREST API 测试用户表';
COMMENT ON COLUMN test_users.id IS '用户ID';
COMMENT ON COLUMN test_users.name IS '用户姓名';
COMMENT ON COLUMN test_users.email IS '用户邮箱';
COMMENT ON COLUMN test_users.age IS '用户年龄';
COMMENT ON COLUMN test_users.salary IS '用户薪水';
COMMENT ON COLUMN test_users.created_at IS '创建时间';
COMMENT ON COLUMN test_users.is_active IS '是否激活';

-- 创建JSON测试表
CREATE TABLE test_json_data (
    id SERIAL PRIMARY KEY,
    metadata JSONB,
    tags TEXT[],
    settings JSONB DEFAULT '{"enabled": true}'
);

-- 创建注释
COMMENT ON TABLE test_json_data IS 'pREST API JSON功能测试表';
COMMENT ON COLUMN test_json_data.id IS '记录ID';
COMMENT ON COLUMN test_json_data.metadata IS '元数据（JSON格式）';
COMMENT ON COLUMN test_json_data.tags IS '标签数组';
COMMENT ON COLUMN test_json_data.settings IS '设置（JSON格式）';

-- 创建向量测试表（兼容GaussDB）
CREATE TABLE test_vector_data (
    id SERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    content TEXT,
    embedding vector(3),  -- 3维向量，与现有测试一致
    category VARCHAR(50),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT true
);

-- 添加注释
COMMENT ON TABLE test_vector_data IS 'pREST API 向量功能测试表';
COMMENT ON COLUMN test_vector_data.id IS '记录ID';
COMMENT ON COLUMN test_vector_data.title IS '文档标题';
COMMENT ON COLUMN test_vector_data.content IS '文档内容';
COMMENT ON COLUMN test_vector_data.embedding IS '向量嵌入（3维）';
COMMENT ON COLUMN test_vector_data.category IS '文档分类';
COMMENT ON COLUMN test_vector_data.metadata IS '元数据（JSON格式）';
COMMENT ON COLUMN test_vector_data.created_at IS '创建时间';
COMMENT ON COLUMN test_vector_data.is_active IS '是否激活';

-- 插入一些示例数据（可选）
INSERT INTO test_users (name, email, age, salary) VALUES
('张三', 'zhangsan@example.com', 25, 5000.00),
('李四', 'lisi@example.com', 30, 7500.50),
('王五', 'wangwu@example.com', 28, 6200.00);

INSERT INTO test_json_data (metadata, tags) VALUES
('{"department": "IT", "role": "developer", "skills": ["Go", "PostgreSQL"]}', '{"backend", "database"}'),
('{"department": "Sales", "role": "manager", "region": "North"}', '{"sales", "management"}');

-- 插入向量测试数据（与现有向量测试一致）
INSERT INTO test_vector_data (title, content, embedding, category, metadata) VALUES
('机器学习文档', '关于机器学习的内容', '[0.1, 0.2, 0.3]', 'AI', '{"tags": ["ml", "ai"], "difficulty": "beginner"}'),
('数据库文档', '关于数据库的内容', '[0.4, 0.5, 0.6]', 'Database', '{"tags": ["sql", "db"], "difficulty": "intermediate"}'),
('编程文档', '关于编程的内容', '[0.7, 0.8, 0.9]', 'Programming', '{"tags": ["go", "python"], "difficulty": "advanced"}'),
('更多机器学习内容', '进阶机器学习内容', '[0.15, 0.25, 0.35]', 'AI', '{"tags": ["deep-learning", "neural-network"], "difficulty": "advanced"}'),
('高级数据库主题', '数据库优化和性能', '[0.45, 0.55, 0.65]', 'Database', '{"tags": ["optimization", "performance"], "difficulty": "expert"}');

-- 显示创建结果
SELECT '✅ 测试表创建完成' as message;
SELECT tablename, tableowner FROM pg_tables WHERE schemaname = 'public' AND tablename IN ('test_users', 'test_json_data', 'test_vector_data');