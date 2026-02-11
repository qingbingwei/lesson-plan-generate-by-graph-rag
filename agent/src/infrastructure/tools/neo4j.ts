import neo4j, { Driver, Session, Record as Neo4jRecord } from 'neo4j-driver';
import config from '../../config';
import logger from '../../shared/utils/logger';
import { gradeToNumber } from '../../shared/utils/gradeNormalizer';
import type { KnowledgeNode, KnowledgeLink, KnowledgePoint, SearchResult } from '../../shared/types';

/**
 * Neo4j 知识图谱工具类
 */
class Neo4jTool {
  private driver: Driver;

  constructor() {
    this.driver = neo4j.driver(
      config.neo4j.uri,
      neo4j.auth.basic(config.neo4j.username, config.neo4j.password)
    );
  }

  /**
   * 获取会话
   */
  private getSession(): Session {
    return this.driver.session();
  }

  /**
   * 关闭连接
   */
  async close(): Promise<void> {
    await this.driver.close();
  }

  /**
   * 根据学科和年级查询知识点
   * @param subject 学科
   * @param grade 年级
   * @param topic 主题（可选）
   * @param userId 用户ID（必须），只查询用户自己创建的知识点
   */
  async getKnowledgePoints(
    subject: string,
    grade: string,
    topic?: string,
    userId?: string
  ): Promise<KnowledgePoint[]> {
    const session = this.getSession();
    const normalizedGrade = gradeToNumber(grade);
    
    try {
      // 只查询用户自己创建的知识点（通过文档上传生成）
      let query = `
        MATCH (k:KnowledgePoint)
        WHERE k.userId = $userId
      `;
      const params: Record<string, unknown> = { 
        subject, 
        grade: normalizedGrade, 
        topic,
        userId: userId || ''
      };
      
      // 可选：按学科过滤
      if (subject) {
        query += ` AND (k.subject = $subject OR k.subject IS NULL)`;
      }
      
      if (topic) {
        query += ` AND (k.name CONTAINS $topic OR k.description CONTAINS $topic OR k.content CONTAINS $topic)`;
      }
      
      query += ` RETURN k ORDER BY k.importance DESC LIMIT 20`;
      
      logger.debug('Executing knowledge query', { query, subject, grade: normalizedGrade, topic, userId });
      
      const result = await session.run(query, params);
      
      logger.debug('Knowledge query result', { count: result.records.length });
      
      return result.records.map((record) => {
        const node = record.get('k');
        return {
          id: node.properties.id || node.elementId,
          name: node.properties.name,
          description: node.properties.description || '',
          difficulty: node.properties.difficulty || 'medium',
          grade: node.properties.grade,
          importance: node.properties.importance || 1,
          content: node.properties.content || '',
          examples: node.properties.examples || [],
        };
      });
    } catch (error) {
      logger.error('Failed to get knowledge points', { error, subject, grade, topic });
      throw error;
    } finally {
      await session.close();
    }
  }

  /**
   * 根据ID获取知识点及其相关节点
   */
  async getKnowledgePointWithRelations(
    knowledgePointId: string,
    depth: number = 2
  ): Promise<{ node: KnowledgePoint; relations: KnowledgeLink[] }> {
    const session = this.getSession();
    
    try {
      const query = `
        MATCH (k:KnowledgePoint {id: $id})
        OPTIONAL MATCH path = (k)-[r*1..${depth}]-(related)
        RETURN k, collect(DISTINCT {
          source: startNode(r[0]).id,
          target: endNode(r[0]).id,
          type: type(r[0])
        }) as relations
      `;
      
      const result = await session.run(query, { id: knowledgePointId });
      
      if (result.records.length === 0) {
        throw new Error(`Knowledge point not found: ${knowledgePointId}`);
      }
      
      const record = result.records[0];
      if (!record) {
        throw new Error(`Knowledge point not found: ${knowledgePointId}`);
      }
      const nodeData = record.get('k');
      const relationsData = record.get('relations');
      
      const node: KnowledgePoint = {
        id: nodeData.properties.id || nodeData.elementId,
        name: nodeData.properties.name,
        description: nodeData.properties.description || '',
        difficulty: nodeData.properties.difficulty || 'medium',
        grade: nodeData.properties.grade,
        importance: nodeData.properties.importance || 1,
        content: nodeData.properties.content || '',
        examples: nodeData.properties.examples || [],
      };
      
      const relations: KnowledgeLink[] = relationsData
        .filter((r: { source: string; target: string; type: string }) => r.source && r.target)
        .map((r: { source: string; target: string; type: string }) => ({
          source: r.source,
          target: r.target,
          type: r.type,
        }));
      
      return { node, relations };
    } catch (error) {
      logger.error('Failed to get knowledge point with relations', { error, knowledgePointId });
      throw error;
    } finally {
      await session.close();
    }
  }

  /**
   * 查询与指定知识点相关的前置知识
   */
  async getPrerequisites(knowledgePointId: string): Promise<KnowledgePoint[]> {
    const session = this.getSession();
    
    try {
      const query = `
        MATCH (k:KnowledgePoint {id: $id})<-[:PREREQUISITE_FOR]-(pre:KnowledgePoint)
        RETURN pre
        ORDER BY pre.importance DESC
      `;
      
      const result = await session.run(query, { id: knowledgePointId });
      
      return result.records.map((record) => {
        const node = record.get('pre');
        return {
          id: node.properties.id || node.elementId,
          name: node.properties.name,
          description: node.properties.description || '',
          difficulty: node.properties.difficulty || 'medium',
          grade: node.properties.grade,
          importance: node.properties.importance || 1,
          content: node.properties.content || '',
          examples: node.properties.examples || [],
        };
      });
    } catch (error) {
      logger.error('Failed to get prerequisites', { error, knowledgePointId });
      throw error;
    } finally {
      await session.close();
    }
  }

  /**
   * 基于图结构的相似度搜索
   * @param userId 用户ID（必须），只搜索用户自己创建的知识点
   */
  async graphSimilaritySearch(
    subject: string,
    grade: string,
    keywords: string[],
    limit: number = 10,
    userId?: string
  ): Promise<SearchResult[]> {
    const session = this.getSession();
    
    try {
      // 构建关键词匹配查询
      const keywordConditions = keywords
        .map((_, i) => `k.name CONTAINS $keyword${i} OR k.description CONTAINS $keyword${i}`)
        .join(' OR ');
      
      const params: Record<string, string | number | ReturnType<typeof neo4j.int>> = {
        subject,
        grade,
        limit: neo4j.int(limit),
        userId: userId || '',
      };
      
      keywords.forEach((keyword, i) => {
        params[`keyword${i}`] = keyword;
      });
      
      // 只查询用户自己创建的知识点（通过文档上传生成）
      const query = `
        MATCH (k:KnowledgePoint)
        WHERE k.userId = $userId AND (${keywordConditions || 'true'})
        ${subject ? 'AND (k.subject = $subject OR k.subject IS NULL)' : ''}
        WITH k, COALESCE(k.importance, 1) as importance
        RETURN k, importance as score
        ORDER BY score DESC
        LIMIT $limit
      `;
      
      const result = await session.run(query, params);
      
      return result.records.map((record) => {
        const node = record.get('k');
        const score = record.get('score');
        
        return {
          node: {
            id: node.properties.id || node.elementId,
            name: node.properties.name,
            type: 'KnowledgePoint',
            properties: node.properties,
          },
          score: typeof score === 'number' ? score : (score?.toNumber ? score.toNumber() : 0),
          graphScore: typeof score === 'number' ? score : (score?.toNumber ? score.toNumber() : 0),
        };
      });
    } catch (error) {
      logger.error('Failed to perform graph similarity search', { error, subject, grade, keywords });
      throw error;
    } finally {
      await session.close();
    }
  }

  /**
   * 获取知识图谱子图（用于可视化）
   */
  async getSubgraph(
    centerNodeId: string,
    depth: number = 2
  ): Promise<{ nodes: KnowledgeNode[]; links: KnowledgeLink[] }> {
    const session = this.getSession();
    
    try {
      const query = `
        MATCH (center {id: $id})
        CALL apoc.path.subgraphAll(center, {maxLevel: $depth})
        YIELD nodes, relationships
        RETURN nodes, relationships
      `;
      
      // 如果没有APOC插件，使用备用查询
      const fallbackQuery = `
        MATCH path = (center {id: $id})-[*0..${depth}]-(related)
        WITH collect(DISTINCT center) + collect(DISTINCT related) as allNodes,
             collect(DISTINCT relationships(path)) as allRels
        UNWIND allNodes as n
        WITH collect(DISTINCT n) as nodes, allRels
        UNWIND allRels as rels
        UNWIND rels as r
        WITH nodes, collect(DISTINCT r) as relationships
        RETURN nodes, relationships
      `;
      
      let result;
      try {
        result = await session.run(query, { id: centerNodeId, depth: neo4j.int(depth) });
      } catch {
        // APOC不可用，使用备用查询
        result = await session.run(fallbackQuery, { id: centerNodeId });
      }
      
      if (result.records.length === 0) {
        return { nodes: [], links: [] };
      }
      
      const record = result.records[0];
      if (!record) {
        return { nodes: [], links: [] };
      }
      const nodesData = record.get('nodes');
      const relationshipsData = record.get('relationships');
      
      const nodes: KnowledgeNode[] = nodesData.map((n: { properties: Record<string, unknown>; labels: string[]; elementId: string }) => ({
        id: n.properties.id || n.elementId,
        name: String(n.properties.name || ''),
        type: n.labels?.[0] || 'Unknown',
        properties: n.properties,
      }));
      
      const links: KnowledgeLink[] = relationshipsData.map((r: { startNodeElementId: string; endNodeElementId: string; type: string; properties: Record<string, unknown> }) => ({
        source: r.startNodeElementId,
        target: r.endNodeElementId,
        type: r.type,
        properties: r.properties,
      }));
      
      return { nodes, links };
    } catch (error) {
      logger.error('Failed to get subgraph', { error, centerNodeId, depth });
      throw error;
    } finally {
      await session.close();
    }
  }

  /**
   * 存储 Embedding 向量
   */
  async storeEmbedding(
    nodeId: string,
    embedding: number[]
  ): Promise<void> {
    const session = this.getSession();
    
    try {
      const query = `
        MATCH (n {id: $id})
        SET n.embedding = $embedding
        RETURN n
      `;
      
      await session.run(query, { id: nodeId, embedding });
      
      logger.debug('Embedding stored', { nodeId, dimension: embedding.length });
    } catch (error) {
      logger.error('Failed to store embedding', { error, nodeId });
      throw error;
    } finally {
      await session.close();
    }
  }

  /**
   * 执行自定义 Cypher 查询
   */
  async runQuery(
    query: string,
    params: Record<string, unknown> = {}
  ): Promise<Neo4jRecord[]> {
    const session = this.getSession();
    
    try {
      const result = await session.run(query, params);
      return result.records;
    } catch (error) {
      logger.error('Failed to run query', { error, query });
      throw error;
    } finally {
      await session.close();
    }
  }

  /**
   * 创建知识点节点
   */
  async createKnowledgePoint(point: {
    id: string;
    name: string;
    type?: string;
    description: string;
    difficulty: string;
    grade: string;
    importance: number;
    content: string;
    examples: string[];
    documentId?: string;
    userId?: string;
    subject?: string;
  }): Promise<void> {
    const session = this.getSession();
    
    try {
      const query = `
        MERGE (k:KnowledgePoint {id: $id})
        SET k.name = $name,
            k.type = $type,
            k.description = $description,
            k.difficulty = $difficulty,
            k.grade = $grade,
            k.importance = $importance,
            k.content = $content,
            k.examples = $examples,
            k.documentId = $documentId,
            k.userId = $userId,
            k.subject = $subject,
            k.createdAt = datetime()
        RETURN k
      `;
      
      await session.run(query, {
        id: point.id,
        name: point.name,
        type: point.type || 'KnowledgePoint',
        description: point.description,
        difficulty: point.difficulty,
        grade: point.grade,
        importance: point.importance,
        content: point.content,
        examples: point.examples,
        documentId: point.documentId || null,
        userId: point.userId || null,
        subject: point.subject || null,
      });
      
      logger.debug('Created knowledge point', { id: point.id, name: point.name });
    } catch (error) {
      logger.error('Failed to create knowledge point', { error, point });
      throw error;
    } finally {
      await session.close();
    }
  }

  /**
   * 创建实体间关系
   */
  async createRelation(
    sourceId: string,
    targetId: string,
    relationType: string,
    properties?: Record<string, unknown>
  ): Promise<void> {
    const session = this.getSession();
    
    try {
      // 动态创建关系类型
      const query = `
        MATCH (source:KnowledgePoint {id: $sourceId})
        MATCH (target:KnowledgePoint {id: $targetId})
        MERGE (source)-[r:${relationType}]->(target)
        SET r.createdAt = datetime()
        ${properties ? ', r += $properties' : ''}
        RETURN r
      `;
      
      await session.run(query, {
        sourceId,
        targetId,
        properties: properties || {},
      });
      
      logger.debug('Created relation', { sourceId, targetId, type: relationType });
    } catch (error) {
      logger.error('Failed to create relation', { error, sourceId, targetId, relationType });
      throw error;
    } finally {
      await session.close();
    }
  }

  /**
   * 删除用户文档相关的所有节点和关系
   */
  async deleteDocumentNodes(documentId: string): Promise<void> {
    const session = this.getSession();
    
    try {
      const query = `
        MATCH (k:KnowledgePoint {documentId: $documentId})
        DETACH DELETE k
      `;
      
      await session.run(query, { documentId });
      
      logger.info('Deleted document nodes', { documentId });
    } catch (error) {
      logger.error('Failed to delete document nodes', { error, documentId });
      throw error;
    } finally {
      await session.close();
    }
  }
}

// 单例模式
let instance: Neo4jTool | null = null;

export function getNeo4jTool(): Neo4jTool {
  if (!instance) {
    instance = new Neo4jTool();
  }
  return instance;
}

// 导出删除文档节点的便捷函数
export async function deleteDocumentNodes(documentId: string): Promise<void> {
  const tool = getNeo4jTool();
  return tool.deleteDocumentNodes(documentId);
}

export default Neo4jTool;
