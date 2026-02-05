import { z } from 'zod';
import { getGraphRAG } from '../rag/graphRag';
import logger from '../utils/logger';
import type { KnowledgeContext, GenerateLessonRequest } from '../types';

// Skill 定义 Schema
const SkillSchema = z.object({
  name: z.string(),
  description: z.string(),
  content: z.string(),
});

type Skill = z.infer<typeof SkillSchema>;

/**
 * 知识检索 Skill 定义
 */
export const knowledgeRetrievalSkill: Skill = {
  name: 'knowledge_retrieval',
  description: '从用户的个人知识图谱中检索与教学主题相关的知识点，支持混合搜索（向量+图谱）',
  content: `
# 知识检索技能

## 功能说明
从用户上传的文档构建的知识图谱中检索相关知识点，用于教案生成的知识支撑。

## 检索策略
1. **向量检索**: 基于语义相似度匹配相关知识点
2. **图谱检索**: 基于知识点关系进行扩展检索
3. **混合排序**: 结合向量相似度和图谱距离进行重排序

## 检索参数
- maxResults: 最大返回结果数（默认10）
- searchDepth: 图谱搜索深度（默认2）
- userId: 用户ID，用于过滤个人知识库

## 返回结构
每个知识点包含：
- id: 知识点唯一标识
- name: 知识点名称
- content: 知识点内容
- relevanceScore: 相关性得分
- source: 来源（knowledge_graph）
`,
};

// GraphRAG 实例（单例）
let graphRAGInstance: ReturnType<typeof getGraphRAG> | null = null;

function getGraphRAGInstance() {
  if (!graphRAGInstance) {
    graphRAGInstance = getGraphRAG();
  }
  return graphRAGInstance;
}

/**
 * 构建检索查询
 */
function buildQuery(request: GenerateLessonRequest): string {
  const parts = [request.topic, request.subject, request.grade];

  if (request.requirements) {
    parts.push(request.requirements);
  }

  return parts.join(' ');
}

/**
 * 过滤和合并上下文
 */
function filterContexts(
  contexts: KnowledgeContext[],
  request: GenerateLessonRequest
): KnowledgeContext[] {
  // 如果请求中已有上下文，合并
  if (request.context && request.context.length > 0) {
    const existingIds = new Set(request.context.map((c) => c.id));
    const newContexts = contexts.filter((c) => !existingIds.has(c.id));
    return [...request.context, ...newContexts];
  }

  return contexts;
}

/**
 * 执行知识检索
 */
export async function retrieveKnowledge(
  request: GenerateLessonRequest
): Promise<KnowledgeContext[]> {
  const startTime = Date.now();
  logger.info('KnowledgeRetrieval executing', {
    subject: request.subject,
    grade: request.grade,
    topic: request.topic,
  });

  try {
    const graphRAG = getGraphRAGInstance();

    // 构建检索查询
    const query = buildQuery(request);

    // 执行混合检索
    const contexts = await graphRAG.hybridSearch(query, request.subject, request.grade, {
      maxResults: 10,
      searchDepth: 2,
      userId: request.userId, // 传递用户ID用于过滤个人知识库
    });

    // 如果有额外的上下文要求，进行过滤
    const filteredContexts = filterContexts(contexts, request);

    logger.info('KnowledgeRetrieval completed', {
      duration: Date.now() - startTime,
      contextCount: filteredContexts.length,
    });

    return filteredContexts;
  } catch (error) {
    logger.error('KnowledgeRetrieval failed', { error, request });
    throw error;
  }
}

/**
 * 获取单个知识点的详细信息
 */
export async function getKnowledgeDetail(
  knowledgeId: string
): Promise<KnowledgeContext | null> {
  try {
    const graphRAG = getGraphRAGInstance();
    const subgraph = await graphRAG.getKnowledgeSubgraph(knowledgeId, 1);

    if (!subgraph.nodes || subgraph.nodes.length === 0) {
      return null;
    }

    const node = subgraph.nodes[0] as {
      id: string;
      name: string;
      properties?: Record<string, unknown>;
    };

    return {
      id: node.id,
      name: node.name,
      content: String(node.properties?.content || node.properties?.description || ''),
      source: 'knowledge_graph',
    };
  } catch (error) {
    logger.error('Failed to get knowledge detail', { error, knowledgeId });
    return null;
  }
}
