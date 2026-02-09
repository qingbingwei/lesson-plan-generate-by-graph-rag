import { StateGraph, END, START, Annotation } from '@langchain/langgraph';
import { getNeo4jTool } from '../tools/neo4j';
import { getDeepSeekClient } from '../clients/deepseek';
import logger from '../utils/logger';

/**
 * 知识图谱构建请求
 */
export interface BuildGraphRequest {
  documentId: string;
  userId: string;
  title: string;
  content: string;
  fileType: string;
  subject?: string;
  grade?: string;
}

/**
 * 提取的实体
 */
interface ExtractedEntity {
  id: string;
  name: string;
  type: string; // KnowledgePoint, Concept, Principle, etc.
  description: string;
  properties: Record<string, unknown>;
}

/**
 * 提取的关系
 */
interface ExtractedRelation {
  sourceId: string;
  targetId: string;
  type: string; // DEPENDS_ON, RELATES_TO, PART_OF, etc.
  properties?: Record<string, unknown>;
}

/**
 * 知识图谱构建状态类型
 */
interface BuildGraphState {
  request: BuildGraphRequest;
  chunks: string[];
  entities: ExtractedEntity[];
  relations: ExtractedRelation[];
  insertedEntities: number;
  insertedRelations: number;
  error?: string;
}

// 工具实例
const neo4jTool = getNeo4jTool();
const deepseekClient = getDeepSeekClient();

// 有效的节点类型及其别名映射
const TYPE_ALIASES: Record<string, string> = {
  // 标准类型
  subject: 'Subject', chapter: 'Chapter', knowledgepoint: 'KnowledgePoint',
  skill: 'Skill', concept: 'Concept', principle: 'Principle',
  formula: 'Formula', example: 'Example',
  // 中文别名
  '学科': 'Subject', '章节': 'Chapter', '知识点': 'KnowledgePoint',
  '技能': 'Skill', '概念': 'Concept', '原理': 'Principle',
  '公式': 'Formula', '示例': 'Example', '例题': 'Example',
  '定理': 'Principle', '法则': 'Principle', '规律': 'Principle',
  '方法': 'Skill', '技巧': 'Skill',
};
const VALID_TYPES = new Set(Object.values(TYPE_ALIASES));

/** 将 LLM 返回的 type 标准化为有效枚举值 */
function normalizeType(raw: string | undefined): string {
  if (!raw) return 'KnowledgePoint';
  // 优先精确匹配
  if (VALID_TYPES.has(raw)) return raw;
  // 忽略大小写匹配
  const alias = TYPE_ALIASES[raw.toLowerCase()];
  if (alias) return alias;
  // 默认
  return 'KnowledgePoint';
}

/**
 * 文档分块节点
 */
async function chunkDocumentNode(state: BuildGraphState): Promise<Partial<BuildGraphState>> {
  logger.info('ChunkDocumentNode: Starting', { title: state.request.title });
  
  const content = state.request.content;
  const chunks: string[] = [];
  
  // 按段落分块，每块不超过 2000 字符
  const paragraphs = content.split(/\n\n+/);
  let currentChunk = '';
  
  for (const para of paragraphs) {
    if (currentChunk.length + para.length > 2000) {
      if (currentChunk) {
        chunks.push(currentChunk.trim());
      }
      currentChunk = para;
    } else {
      currentChunk += '\n\n' + para;
    }
  }
  
  if (currentChunk.trim()) {
    chunks.push(currentChunk.trim());
  }
  
  logger.info('ChunkDocumentNode: Completed', { chunkCount: chunks.length });
  
  return { chunks };
}

/**
 * 实体提取节点
 */
async function extractEntitiesNode(state: BuildGraphState): Promise<Partial<BuildGraphState>> {
  logger.info('ExtractEntitiesNode: Starting', { chunkCount: state.chunks.length });
  
  const allEntities: ExtractedEntity[] = [];
  const entityMap = new Map<string, ExtractedEntity>();
  
  for (let i = 0; i < state.chunks.length; i++) {
    const chunk = state.chunks[i];
    logger.debug(`Processing chunk ${i + 1}/${state.chunks.length}`);
    
    try {
      const prompt = `请从以下教育文本中提取知识点实体。每个实体需要包含：
- name: 实体名称
- type: 实体类型，必须是以下之一：Subject（学科）、Chapter（章节）、KnowledgePoint（知识点）、Skill（技能）、Concept（概念）、Principle（原理）、Formula（公式）、Example（示例）
- description: 简短描述
- difficulty: 难度等级，必须是 easy/medium/hard 之一，根据内容的复杂程度判断
- importance: 重要程度，0-1之间的小数，核心知识点接近1，辅助性知识点接近0.3

文本内容：
${chunk}

${state.request.subject ? `学科：${state.request.subject}` : ''}
${state.request.grade ? `年级：${state.request.grade}` : ''}

请以JSON数组格式返回实体列表：
[{"name": "...", "type": "...", "description": "...", "difficulty": "...", "importance": 0.8}]

只返回JSON，不要其他内容。`;

      const { content } = await deepseekClient.chat([
        { role: 'system', content: '你是一个教育知识图谱构建专家，擅长从教育文本中提取知识点实体。' },
        { role: 'user', content: prompt },
      ], { temperature: 0.3 });
      
      // 解析实体
      const jsonMatch = content.match(/\[[\s\S]*\]/);
      if (jsonMatch) {
        const entities = JSON.parse(jsonMatch[0]) as Array<{name: string; type: string; description: string; difficulty?: string; importance?: number}>;
        
        // 有效的节点类型及常见别名映射
        const validTypes = new Set(['Subject', 'Chapter', 'KnowledgePoint', 'Skill', 'Concept', 'Principle', 'Formula', 'Example']);
        const typeAliasMap: Record<string, string> = {
          '学科': 'Subject', 'subject': 'Subject',
          '章节': 'Chapter', 'chapter': 'Chapter', '单元': 'Chapter', '模块': 'Chapter',
          '知识点': 'KnowledgePoint', 'knowledgepoint': 'KnowledgePoint', 'knowledge_point': 'KnowledgePoint', 'knowledge': 'KnowledgePoint',
          '技能': 'Skill', 'skill': 'Skill', '能力': 'Skill',
          '概念': 'Concept', 'concept': 'Concept', '定义': 'Concept',
          '原理': 'Principle', 'principle': 'Principle', '定理': 'Principle', '定律': 'Principle', '法则': 'Principle',
          '公式': 'Formula', 'formula': 'Formula', '方程': 'Formula',
          '示例': 'Example', 'example': 'Example', '例题': 'Example', '案例': 'Example',
        };
        function normalizeType(raw: string | undefined): string {
          if (!raw) return 'KnowledgePoint';
          if (validTypes.has(raw)) return raw;
          const mapped = typeAliasMap[raw.toLowerCase()] || typeAliasMap[raw];
          return mapped || 'KnowledgePoint';
        }

        for (const entity of entities) {
          const id = `${state.request.documentId}-${entity.name.replace(/\s+/g, '-')}`;
          
          if (!entityMap.has(entity.name)) {
            // 校验 difficulty
            const validDifficulties = ['easy', 'medium', 'hard'];
            const difficulty = validDifficulties.includes(entity.difficulty || '') ? entity.difficulty! : 'medium';
            // 校验 importance
            const importance = typeof entity.importance === 'number' && entity.importance >= 0 && entity.importance <= 1
              ? entity.importance : 0.5;

            const extractedEntity: ExtractedEntity = {
              id,
              name: entity.name,
              type: normalizeType(entity.type),
              description: entity.description || '',
              properties: {
                documentId: state.request.documentId,
                userId: state.request.userId,
                subject: state.request.subject,
                grade: state.request.grade,
                difficulty,
                importance,
              },
            };
            entityMap.set(entity.name, extractedEntity);
            allEntities.push(extractedEntity);
          }
        }
      }
    } catch (error) {
      logger.warn(`Failed to extract entities from chunk ${i + 1}`, { error });
    }
  }
  
  logger.info('ExtractEntitiesNode: Completed', { entityCount: allEntities.length });
  
  return { entities: allEntities };
}

/**
 * 关系提取节点
 */
async function extractRelationsNode(state: BuildGraphState): Promise<Partial<BuildGraphState>> {
  logger.info('ExtractRelationsNode: Starting', { entityCount: state.entities.length });
  
  if (state.entities.length < 2) {
    logger.info('ExtractRelationsNode: Not enough entities for relations');
    return { relations: [] };
  }
  
  const entityNames = state.entities.map(e => e.name);
  const allRelations: ExtractedRelation[] = [];
  
  // 对每个文本块提取关系
  for (let i = 0; i < state.chunks.length; i++) {
    const chunk = state.chunks[i];
    
    try {
      const prompt = `请分析以下教育文本中知识点之间的关系。

已知的知识点实体：
${entityNames.join(', ')}

文本内容：
${chunk}

请识别知识点之间的关系，关系类型包括：
- DEPENDS_ON: 前置依赖关系（学习A需要先学习B）
- RELATES_TO: 相关关系（A和B有关联）
- PART_OF: 包含关系（A是B的一部分）
- SIMILAR_TO: 相似关系

请以JSON数组格式返回关系列表：
[{"source": "知识点A名称", "target": "知识点B名称", "type": "关系类型"}]

只返回JSON，不要其他内容。如果没有发现明确的关系，返回空数组 []。`;

      const { content } = await deepseekClient.chat([
        { role: 'system', content: '你是一个教育知识图谱构建专家，擅长识别知识点之间的关系。' },
        { role: 'user', content: prompt },
      ], { temperature: 0.3 });
      
      const jsonMatch = content.match(/\[[\s\S]*\]/);
      if (jsonMatch) {
        const relations = JSON.parse(jsonMatch[0]) as Array<{source: string; target: string; type: string}>;
        
        for (const rel of relations) {
          const sourceEntity = state.entities.find(e => e.name === rel.source);
          const targetEntity = state.entities.find(e => e.name === rel.target);
          
          if (sourceEntity && targetEntity && sourceEntity.id !== targetEntity.id) {
            allRelations.push({
              sourceId: sourceEntity.id,
              targetId: targetEntity.id,
              type: rel.type || 'RELATES_TO',
            });
          }
        }
      }
    } catch (error) {
      logger.warn(`Failed to extract relations from chunk ${i + 1}`, { error });
    }
  }
  
  // 去重
  const uniqueRelations = allRelations.filter((rel, index, self) =>
    index === self.findIndex(r => 
      r.sourceId === rel.sourceId && r.targetId === rel.targetId && r.type === rel.type
    )
  );
  
  logger.info('ExtractRelationsNode: Completed', { relationCount: uniqueRelations.length });
  
  return { relations: uniqueRelations };
}

/**
 * 插入Neo4j节点
 */
async function insertToNeo4jNode(state: BuildGraphState): Promise<Partial<BuildGraphState>> {
  logger.info('InsertToNeo4jNode: Starting', { 
    entityCount: state.entities.length,
    relationCount: state.relations.length,
  });
  
  let insertedEntities = 0;
  let insertedRelations = 0;
  
  try {
    // 插入实体
    for (const entity of state.entities) {
      try {
        await neo4jTool.createKnowledgePoint({
          id: entity.id,
          name: entity.name,
          type: entity.type || 'KnowledgePoint',
          description: entity.description,
          difficulty: (entity.properties.difficulty as string) || 'medium',
          grade: state.request.grade || '',
          importance: (entity.properties.importance as number) ?? 0.5,
          content: entity.description,
          examples: [],
          ...entity.properties,
        });
        insertedEntities++;
      } catch (error) {
        logger.warn('Failed to insert entity', { entity: entity.name, error });
      }
    }
    
    // 插入关系
    for (const relation of state.relations) {
      try {
        await neo4jTool.createRelation(
          relation.sourceId,
          relation.targetId,
          relation.type
        );
        insertedRelations++;
      } catch (error) {
        logger.warn('Failed to insert relation', { relation, error });
      }
    }
    
    logger.info('InsertToNeo4jNode: Completed', { insertedEntities, insertedRelations });
    
    return { insertedEntities, insertedRelations };
  } catch (error) {
    const errorMsg = error instanceof Error ? error.message : 'Unknown error';
    logger.error('InsertToNeo4jNode: Failed', { error });
    return { error: errorMsg, insertedEntities, insertedRelations };
  }
}

// 定义状态注解 - LangGraph 1.0+ 需要这个来正确传递和合并状态
const BuildGraphStateAnnotation = Annotation.Root({
  request: Annotation<BuildGraphRequest>({
    reducer: (_, b) => b,
    default: () => ({ documentId: '', userId: '', title: '', content: '', fileType: '' }),
  }),
  chunks: Annotation<string[]>({
    reducer: (_, b) => b,
    default: () => [],
  }),
  entities: Annotation<ExtractedEntity[]>({
    reducer: (_, b) => b,
    default: () => [],
  }),
  relations: Annotation<ExtractedRelation[]>({
    reducer: (_, b) => b,
    default: () => [],
  }),
  insertedEntities: Annotation<number>({
    reducer: (_, b) => b,
    default: () => 0,
  }),
  insertedRelations: Annotation<number>({
    reducer: (_, b) => b,
    default: () => 0,
  }),
  error: Annotation<string | undefined>({
    reducer: (_, b) => b,
    default: () => undefined,
  }),
});

/**
 * 创建知识图谱构建工作流
 */
function createBuildGraphWorkflow() {
  const workflow = new StateGraph(BuildGraphStateAnnotation)
    .addNode('chunkDocument', chunkDocumentNode)
    .addNode('extractEntities', extractEntitiesNode)
    .addNode('extractRelations', extractRelationsNode)
    .addNode('insertToNeo4j', insertToNeo4jNode)
    .addEdge(START, 'chunkDocument')
    .addEdge('chunkDocument', 'extractEntities')
    .addEdge('extractEntities', 'extractRelations')
    .addEdge('extractRelations', 'insertToNeo4j')
    .addEdge('insertToNeo4j', END);
  
  return workflow.compile();
}

// 编译工作流
const buildGraphApp = createBuildGraphWorkflow();

/**
 * 运行知识图谱构建工作流
 */
export async function runBuildGraphWorkflow(request: BuildGraphRequest): Promise<{
  success: boolean;
  entityCount: number;
  relationCount: number;
  error?: string;
}> {
  logger.info('Starting build graph workflow', {
    documentId: request.documentId,
    title: request.title,
    contentLength: request.content.length,
  });
  
  const startTime = Date.now();
  
  try {
    const result = await buildGraphApp.invoke({
      request,
      chunks: [],
      entities: [],
      relations: [],
      insertedEntities: 0,
      insertedRelations: 0,
    }) as {
      insertedEntities: number;
      insertedRelations: number;
      error?: string;
    };
    
    logger.info('Build graph workflow completed', {
      documentId: request.documentId,
      duration: Date.now() - startTime,
      entityCount: result.insertedEntities,
      relationCount: result.insertedRelations,
    });
    
    if (result.error) {
      return {
        success: false,
        entityCount: result.insertedEntities,
        relationCount: result.insertedRelations,
        error: result.error,
      };
    }
    
    return {
      success: true,
      entityCount: result.insertedEntities,
      relationCount: result.insertedRelations,
    };
  } catch (error) {
    const errorMsg = error instanceof Error ? error.message : 'Unknown error';
    logger.error('Build graph workflow failed', { error, documentId: request.documentId });
    
    return {
      success: false,
      entityCount: 0,
      relationCount: 0,
      error: errorMsg,
    };
  }
}

export default runBuildGraphWorkflow;
