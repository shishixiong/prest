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
DROP TABLE IF EXISTS test_json_data;
DROP TABLE IF EXISTS test_users;

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

-- 插入一些示例数据（可选）
INSERT INTO test_users (name, email, age, salary) VALUES
('张三', 'zhangsan@example.com', 25, 5000.00),
('李四', 'lisi@example.com', 30, 7500.50),
('王五', 'wangwu@example.com', 28, 6200.00);

INSERT INTO test_json_data (metadata, tags) VALUES
('{"department": "IT", "role": "developer", "skills": ["Go", "PostgreSQL"]}', '{"backend", "database"}'),
('{"department": "Sales", "role": "manager", "region": "North"}', '{"sales", "management"}');

-- 显示创建结果
SELECT '✅ 测试表创建完成' as message;
SELECT tablename, tableowner FROM pg_tables WHERE schemaname = 'public' AND tablename IN ('test_users', 'test_json_data');