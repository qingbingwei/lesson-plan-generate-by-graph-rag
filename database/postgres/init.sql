-- PostgreSQL 数据库初始化脚本

-- 创建扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ==================== 用户表 ====================
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100),
    avatar_url VARCHAR(500),
    role VARCHAR(20) DEFAULT 'teacher' CHECK (role IN ('admin', 'teacher', 'student')),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'banned')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP
);

-- 兼容旧库：补齐软删除列（GORM gorm.DeletedAt 依赖）
ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

-- 用户表索引
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_status ON users(status);

-- ==================== 教案表 ====================
CREATE TABLE IF NOT EXISTS lessons (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    subject VARCHAR(50) NOT NULL,
    grade VARCHAR(20) NOT NULL,
    topic VARCHAR(200) NOT NULL,
    duration INTEGER NOT NULL, -- 课时时长（分钟）
    
    -- 教案内容（JSONB格式）
    content JSONB NOT NULL DEFAULT '{}',
    
    -- 教学目标
    objectives JSONB DEFAULT '{}',
    
    -- 教学重难点
    key_points JSONB DEFAULT '[]',
    difficult_points JSONB DEFAULT '[]',
    
    -- 教学方法
    teaching_methods JSONB DEFAULT '[]',
    
    -- 状态管理
    status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'review', 'published', 'archived')),
    
    -- 统计信息
    view_count INTEGER DEFAULT 0,
    like_count INTEGER DEFAULT 0,
    fork_count INTEGER DEFAULT 0,
    
    -- 元数据
    tags JSONB DEFAULT '[]',
    metadata JSONB DEFAULT '{}',
    
    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMP
);

-- 兼容旧库：补齐软删除列
ALTER TABLE lessons ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
CREATE INDEX IF NOT EXISTS idx_lessons_deleted_at ON lessons(deleted_at);

-- 教案表索引
CREATE INDEX idx_lessons_user_id ON lessons(user_id);
CREATE INDEX idx_lessons_subject ON lessons(subject);
CREATE INDEX idx_lessons_grade ON lessons(grade);
CREATE INDEX idx_lessons_status ON lessons(status);
CREATE INDEX idx_lessons_created_at ON lessons(created_at DESC);
CREATE INDEX idx_lessons_published_at ON lessons(published_at DESC);
CREATE INDEX idx_lessons_content_gin ON lessons USING gin(content);
CREATE INDEX idx_lessons_tags_gin ON lessons USING gin(tags);

-- ==================== 教案版本表 ====================
CREATE TABLE IF NOT EXISTS lesson_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    lesson_id UUID NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    version_number INTEGER NOT NULL,
    content JSONB NOT NULL,
    change_summary TEXT,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(lesson_id, version_number)
);

-- 兼容旧库：补齐软删除列
ALTER TABLE lesson_versions ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
CREATE INDEX IF NOT EXISTS idx_lesson_versions_deleted_at ON lesson_versions(deleted_at);

-- 版本表索引
CREATE INDEX idx_lesson_versions_lesson_id ON lesson_versions(lesson_id);
CREATE INDEX idx_lesson_versions_created_at ON lesson_versions(created_at DESC);

-- ==================== 教案评论表 ====================
CREATE TABLE IF NOT EXISTS lesson_comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    lesson_id UUID NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES lesson_comments(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    rating INTEGER CHECK (rating >= 1 AND rating <= 5),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 兼容旧库：补齐软删除列
ALTER TABLE lesson_comments ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
CREATE INDEX IF NOT EXISTS idx_lesson_comments_deleted_at ON lesson_comments(deleted_at);

-- 评论表索引
CREATE INDEX idx_lesson_comments_lesson_id ON lesson_comments(lesson_id);
CREATE INDEX idx_lesson_comments_user_id ON lesson_comments(user_id);
CREATE INDEX idx_lesson_comments_parent_id ON lesson_comments(parent_id);

-- ==================== 教案收藏表 ====================
CREATE TABLE IF NOT EXISTS lesson_favorites (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    lesson_id UUID NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(lesson_id, user_id)
);

-- 收藏表索引
CREATE INDEX idx_lesson_favorites_user_id ON lesson_favorites(user_id);
CREATE INDEX idx_lesson_favorites_lesson_id ON lesson_favorites(lesson_id);

-- ==================== AI生成记录表 ====================
CREATE TABLE IF NOT EXISTS generation_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    lesson_id UUID REFERENCES lessons(id) ON DELETE SET NULL,
    
    -- 输入参数
    input_params JSONB NOT NULL,
    
    -- 生成状态
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    
    -- 生成结果
    result JSONB,
    error_message TEXT,
    
    -- 性能指标
    duration_ms INTEGER,
    token_count INTEGER,
    
    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);

-- 生成记录索引
CREATE INDEX idx_generation_logs_user_id ON generation_logs(user_id);
CREATE INDEX idx_generation_logs_status ON generation_logs(status);
CREATE INDEX idx_generation_logs_created_at ON generation_logs(created_at DESC);

-- ==================== 生成记录表（GORM模型使用） ====================
CREATE TABLE IF NOT EXISTS generations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    lesson_id UUID REFERENCES lessons(id) ON DELETE SET NULL,
    prompt TEXT NOT NULL,
    parameters JSONB,
    result TEXT,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    token_count INTEGER DEFAULT 0,
    duration_ms BIGINT DEFAULT 0,
    error_msg TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);

-- 生成记录表索引
CREATE INDEX idx_generations_user_id ON generations(user_id);
CREATE INDEX idx_generations_status ON generations(status);
CREATE INDEX idx_generations_created_at ON generations(created_at DESC);

-- ==================== 知识点映射表 ====================
-- 用于PostgreSQL和Neo4j之间的映射
CREATE TABLE IF NOT EXISTS knowledge_mappings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    neo4j_id VARCHAR(100) UNIQUE NOT NULL,
    node_type VARCHAR(50) NOT NULL,
    name VARCHAR(200) NOT NULL,
    properties JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 知识点映射索引
CREATE INDEX idx_knowledge_mappings_neo4j_id ON knowledge_mappings(neo4j_id);
CREATE INDEX idx_knowledge_mappings_node_type ON knowledge_mappings(node_type);
CREATE INDEX idx_knowledge_mappings_name ON knowledge_mappings(name);

-- ==================== 教案知识点关联表 ====================
CREATE TABLE IF NOT EXISTS lesson_knowledge_points (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    lesson_id UUID NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    knowledge_point_id UUID NOT NULL REFERENCES knowledge_mappings(id) ON DELETE CASCADE,
    relevance_score DECIMAL(3, 2) DEFAULT 1.0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(lesson_id, knowledge_point_id)
);

-- 关联表索引
CREATE INDEX idx_lesson_knowledge_points_lesson_id ON lesson_knowledge_points(lesson_id);
CREATE INDEX idx_lesson_knowledge_points_knowledge_id ON lesson_knowledge_points(knowledge_point_id);

-- ==================== 知识文档表 ====================
-- 用户上传的知识文档，用于构建个人知识图谱
CREATE TABLE IF NOT EXISTS knowledge_documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_type VARCHAR(20) NOT NULL CHECK (file_type IN ('txt', 'md')),
    file_size INTEGER NOT NULL,
    content TEXT NOT NULL,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    error_msg TEXT,
    entity_count INTEGER DEFAULT 0,
    relation_count INTEGER DEFAULT 0,
    subject VARCHAR(50),
    grade VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 知识文档索引
CREATE INDEX idx_knowledge_documents_user_id ON knowledge_documents(user_id);
CREATE INDEX idx_knowledge_documents_status ON knowledge_documents(status);
CREATE INDEX idx_knowledge_documents_created_at ON knowledge_documents(created_at DESC);
CREATE INDEX idx_knowledge_documents_subject ON knowledge_documents(subject);

-- 知识文档更新触发器
CREATE TRIGGER update_knowledge_documents_updated_at BEFORE UPDATE ON knowledge_documents
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ==================== 触发器函数 ====================

-- 更新 updated_at 字段的触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 为各表创建触发器
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_lessons_updated_at BEFORE UPDATE ON lessons
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_lesson_comments_updated_at BEFORE UPDATE ON lesson_comments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_knowledge_mappings_updated_at BEFORE UPDATE ON knowledge_mappings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 教案版本自动递增触发器
CREATE OR REPLACE FUNCTION increment_lesson_version()
RETURNS TRIGGER AS $$
BEGIN
    SELECT COALESCE(MAX(version_number), 0) + 1 INTO NEW.version_number
    FROM lesson_versions
    WHERE lesson_id = NEW.lesson_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER auto_increment_version BEFORE INSERT ON lesson_versions
    FOR EACH ROW EXECUTE FUNCTION increment_lesson_version();

-- ==================== 视图 ====================

-- 教案详情视图（包含用户信息）
CREATE OR REPLACE VIEW lesson_details AS
SELECT 
    l.*,
    u.username,
    u.full_name AS author_name,
    u.avatar_url AS author_avatar,
    COUNT(DISTINCT lc.id) AS comment_count,
    COUNT(DISTINCT lf.id) AS favorite_count
FROM lessons l
LEFT JOIN users u ON l.user_id = u.id
LEFT JOIN lesson_comments lc ON l.id = lc.lesson_id
LEFT JOIN lesson_favorites lf ON l.id = lf.lesson_id
GROUP BY l.id, u.username, u.full_name, u.avatar_url;

-- 用户统计视图
CREATE OR REPLACE VIEW user_statistics AS
SELECT 
    u.id,
    u.username,
    COUNT(DISTINCT l.id) AS lesson_count,
    COUNT(DISTINCT lf.id) AS favorite_count,
    COUNT(DISTINCT lc.id) AS comment_count,
    SUM(l.view_count) AS total_views,
    SUM(l.like_count) AS total_likes
FROM users u
LEFT JOIN lessons l ON u.id = l.user_id
LEFT JOIN lesson_favorites lf ON u.id = lf.user_id
LEFT JOIN lesson_comments lc ON u.id = lc.user_id
GROUP BY u.id, u.username;

-- ==================== 权限设置 ====================

-- 创建应用用户（如果不存在）
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_user WHERE usename = 'lesson_plan_app') THEN
        CREATE USER lesson_plan_app WITH PASSWORD 'app_password_change_me';
    END IF;
END
$$;

-- 授予权限
GRANT CONNECT ON DATABASE lesson_plan TO lesson_plan_app;
GRANT USAGE ON SCHEMA public TO lesson_plan_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO lesson_plan_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO lesson_plan_app;

-- 默认权限
ALTER DEFAULT PRIVILEGES IN SCHEMA public 
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO lesson_plan_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public 
    GRANT USAGE, SELECT ON SEQUENCES TO lesson_plan_app;

-- ==================== 完成 ====================
SELECT 'Database initialization completed successfully!' AS status;
