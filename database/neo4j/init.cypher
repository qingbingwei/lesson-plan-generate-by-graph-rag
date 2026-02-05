// Neo4j 图数据库初始化脚本
// 知识图谱由用户上传文档动态生成，不再预置数据

// ==================== 创建约束和索引 ====================

// 创建唯一性约束
CREATE CONSTRAINT knowledge_point_id IF NOT EXISTS FOR (k:KnowledgePoint) REQUIRE k.id IS UNIQUE;

// 创建索引 - 用于知识点查询优化
CREATE INDEX knowledge_point_name IF NOT EXISTS FOR (k:KnowledgePoint) ON (k.name);
CREATE INDEX knowledge_point_userId IF NOT EXISTS FOR (k:KnowledgePoint) ON (k.userId);
CREATE INDEX knowledge_point_documentId IF NOT EXISTS FOR (k:KnowledgePoint) ON (k.documentId);
CREATE INDEX knowledge_point_subject IF NOT EXISTS FOR (k:KnowledgePoint) ON (k.subject);
CREATE INDEX knowledge_point_grade IF NOT EXISTS FOR (k:KnowledgePoint) ON (k.grade);

// 验证完成
RETURN '知识图谱数据库初始化完成！知识点将由用户上传文档动态生成。' AS Status;
