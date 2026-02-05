import { getNeo4jTool } from '../tools/neo4j';
import { getVectorTool } from '../tools/vector';
import config from '../config';
import logger from '../utils/logger';
import type { SearchResult, KnowledgeContext, GraphRAGConfig, KnowledgePoint } from '../types';

/**
 * Graph RAG 混合检索实现
 * 结合向量搜索和图结构搜索
 */
class GraphRAG {
  private neo4jTool = getNeo4jTool();
  private vectorTool = getVectorTool();
  private config: GraphRAGConfig;

  constructor(customConfig?: Partial<GraphRAGConfig>) {
    this.config = {
      vectorWeight: customConfig?.vectorWeight ?? config.graphRag.vectorWeight,
      graphWeight: customConfig?.graphWeight ?? config.graphRag.graphWeight,
      maxResults: customConfig?.maxResults ?? config.graphRag.maxResults,
      searchDepth: customConfig?.searchDepth ?? config.graphRag.searchDepth,
    };
  }

  /**
   * 混合检索
   * 结合向量相似度和图结构相似度
   */
  async hybridSearch(
    query: string,
    subject: string,
    grade: string,
    options?: {
      maxResults?: number;
      searchDepth?: number;
      userId?: string; // 用户ID，用于过滤个人知识库
    }
  ): Promise<KnowledgeContext[]> {
    const maxResults = options?.maxResults ?? this.config.maxResults;
    const searchDepth = options?.searchDepth ?? this.config.searchDepth;
    const userId = options?.userId;
    
    logger.info('Starting hybrid search', { query, subject, grade, maxResults, searchDepth, userId });
    const startTime = Date.now();

    try {
      // 1. 提取关键词
      const keywords = this.extractKeywords(query);
      logger.debug('Extracted keywords', { keywords });

      // 2. 并行执行向量搜索和图搜索
      const [vectorResults, graphResults] = await Promise.all([
        this.vectorSearch(query, subject, grade, maxResults * 2, userId),
        this.graphSearch(subject, grade, keywords, maxResults * 2, userId),
      ]);

      logger.debug('Search results', {
        vectorCount: vectorResults.length,
        graphCount: graphResults.length,
      });

      // 3. 融合结果
      const fusedResults = this.fuseResults(vectorResults, graphResults);

      // 4. 取前 N 个结果
      const topResults = fusedResults.slice(0, maxResults);

      // 5. 获取相关知识点的详细内容
      const contexts = await this.enrichResults(topResults, searchDepth);

      logger.info('Hybrid search completed', {
        duration: Date.now() - startTime,
        resultCount: contexts.length,
      });

      return contexts;
    } catch (error) {
      logger.error('Hybrid search failed', { error, query, subject, grade });
      throw error;
    }
  }

  /**
   * 向量搜索
   * 当 embedding 服务不可用时，回退到基于关键词的知识点匹配
   */
  private async vectorSearch(
    query: string,
    subject: string,
    grade: string,
    limit: number,
    userId?: string
  ): Promise<SearchResult[]> {
    try {
      // 获取所有候选知识点（如果有userId则只获取用户的知识点）
      const knowledgePoints = await this.neo4jTool.getKnowledgePoints(subject, grade, undefined, userId);
      
      if (knowledgePoints.length === 0) {
        logger.warn('No knowledge points found for vector search', { subject, grade });
        return [];
      }

      logger.info('Found knowledge points for search', { count: knowledgePoints.length });

      // 尝试使用 embedding 进行相似度搜索
      let queryEmbedding: number[] | null = null;
      try {
        queryEmbedding = await this.vectorTool.createEmbedding(query);
      } catch (embeddingError) {
        logger.warn('Embedding service unavailable, using fallback keyword matching', { error: embeddingError });
      }

      const results: SearchResult[] = [];
      
      for (const kp of knowledgePoints) {
        let score: number;
        
        if (queryEmbedding) {
          // 如果有查询向量，使用向量相似度
          let embedding = kp.embedding;
          if (!embedding || embedding.length === 0) {
            try {
              const textToEmbed = `${kp.name}: ${kp.description}. ${kp.content}`;
              embedding = await this.vectorTool.createEmbedding(textToEmbed);
            } catch {
              // 如果无法生成 embedding，使用关键词匹配
              score = this.calculateKeywordScore(query, kp);
              results.push(this.createSearchResult(kp, score));
              continue;
            }
          }
          score = this.vectorTool.cosineSimilarity(queryEmbedding, embedding);
        } else {
          // 回退到关键词匹配
          score = this.calculateKeywordScore(query, kp);
        }
        
        results.push(this.createSearchResult(kp, score));
      }

      // 按分数排序
      return results
        .sort((a, b) => b.score - a.score)
        .slice(0, limit);
    } catch (error) {
      logger.error('Vector search failed', { error, query });
      // 返回空数组而不是抛出错误，让图搜索继续
      return [];
    }
  }

  /**
   * 基于关键词计算相似度分数
   */
  private calculateKeywordScore(query: string, kp: KnowledgePoint): number {
    const queryLower = query.toLowerCase();
    const keywords = queryLower.split(/\s+/).filter(k => k.length > 0);
    
    let matchCount = 0;
    const searchText = `${kp.name} ${kp.description} ${kp.content}`.toLowerCase();
    
    for (const keyword of keywords) {
      if (searchText.includes(keyword)) {
        matchCount++;
      }
    }
    
    // 基础分数 + 关键词匹配加分 + 重要性加分
    const keywordScore = keywords.length > 0 ? matchCount / keywords.length : 0;
    const importanceScore = (kp.importance || 1) / 10;
    
    return 0.3 + keywordScore * 0.5 + importanceScore * 0.2;
  }

  /**
   * 创建搜索结果对象
   */
  private createSearchResult(kp: KnowledgePoint, score: number): SearchResult {
    return {
      node: {
        id: kp.id,
        name: kp.name,
        type: 'KnowledgePoint',
        properties: {
          description: kp.description,
          difficulty: kp.difficulty,
          grade: kp.grade,
          importance: kp.importance,
          content: kp.content,
          examples: kp.examples,
        },
      },
      score,
      vectorScore: score,
    };
  }

  /**
   * 图结构搜索
   */
  private async graphSearch(
    subject: string,
    grade: string,
    keywords: string[],
    limit: number,
    userId?: string
  ): Promise<SearchResult[]> {
    try {
      const results = await this.neo4jTool.graphSimilaritySearch(
        subject,
        grade,
        keywords,
        limit,
        userId
      );
      
      return results.map(r => ({
        ...r,
        graphScore: r.score,
      }));
    } catch (error) {
      logger.error('Graph search failed', { error, subject, grade, keywords });
      // 返回空数组而不是抛出错误，让向量搜索结果继续
      return [];
    }
  }

  /**
   * 融合向量搜索和图搜索结果
   */
  private fuseResults(
    vectorResults: SearchResult[],
    graphResults: SearchResult[]
  ): SearchResult[] {
    const resultMap = new Map<string, SearchResult>();

    // 归一化分数
    const normalizeScores = (results: SearchResult[]): SearchResult[] => {
      if (results.length === 0) return results;
      
      const maxScore = Math.max(...results.map(r => r.score));
      const minScore = Math.min(...results.map(r => r.score));
      const range = maxScore - minScore || 1;
      
      return results.map(r => ({
        ...r,
        score: (r.score - minScore) / range,
      }));
    };

    const normalizedVectorResults = normalizeScores(vectorResults);
    const normalizedGraphResults = normalizeScores(graphResults);

    // 添加向量搜索结果
    for (const result of normalizedVectorResults) {
      resultMap.set(result.node.id, {
        ...result,
        vectorScore: result.score,
        graphScore: 0,
        score: result.score * this.config.vectorWeight,
      });
    }

    // 合并图搜索结果
    for (const result of normalizedGraphResults) {
      const existing = resultMap.get(result.node.id);
      
      if (existing) {
        // 如果已存在，合并分数
        existing.graphScore = result.score;
        existing.score = 
          (existing.vectorScore || 0) * this.config.vectorWeight +
          result.score * this.config.graphWeight;
      } else {
        // 如果不存在，添加新结果
        resultMap.set(result.node.id, {
          ...result,
          vectorScore: 0,
          graphScore: result.score,
          score: result.score * this.config.graphWeight,
        });
      }
    }

    // 转换为数组并排序
    return Array.from(resultMap.values())
      .sort((a, b) => b.score - a.score);
  }

  /**
   * 丰富搜索结果，获取相关知识点的详细内容
   */
  private async enrichResults(
    results: SearchResult[],
    depth: number
  ): Promise<KnowledgeContext[]> {
    const contexts: KnowledgeContext[] = [];

    for (const result of results) {
      try {
        // 获取知识点详情及其关系
        const { node, relations } = await this.neo4jTool.getKnowledgePointWithRelations(
          result.node.id,
          depth
        );

        // 获取前置知识
        const prerequisites = await this.neo4jTool.getPrerequisites(result.node.id);

        // 构建上下文
        let content = `${node.name}\n\n${node.description}\n\n${node.content}`;

        if (node.examples && node.examples.length > 0) {
          content += `\n\n示例：\n${node.examples.join('\n')}`;
        }

        if (prerequisites.length > 0) {
          content += `\n\n前置知识：\n${prerequisites.map(p => `- ${p.name}: ${p.description}`).join('\n')}`;
        }

        contexts.push({
          id: result.node.id,
          name: node.name,
          content,
          relevanceScore: result.score,
          source: 'knowledge_graph',
        });
      } catch (error) {
        logger.warn('Failed to enrich result', { error, nodeId: result.node.id });
        // 使用基本信息
        contexts.push({
          id: result.node.id,
          name: result.node.name,
          content: String(result.node.properties?.content || result.node.properties?.description || ''),
          relevanceScore: result.score,
          source: 'knowledge_graph',
        });
      }
    }

    return contexts;
  }

  /**
   * 提取关键词
   */
  private extractKeywords(text: string): string[] {
    // 简单的关键词提取：分词并过滤停用词
    const stopWords = new Set([
      '的', '了', '是', '在', '和', '与', '或', '等', '这', '那',
      '有', '为', '以', '及', '被', '把', '给', '对', '让', '使',
      'the', 'a', 'an', 'is', 'are', 'was', 'were', 'be', 'been',
      'being', 'have', 'has', 'had', 'do', 'does', 'did', 'will',
      'would', 'could', 'should', 'may', 'might', 'must', 'can',
    ]);

    // 按空格、标点分词
    const words = text
      .toLowerCase()
      .split(/[\s,，。！？、；：""''（）()[\]【】{}]+/)
      .filter(word => word.length > 1 && !stopWords.has(word));

    // 去重
    return [...new Set(words)];
  }

  /**
   * 获取知识图谱子图（用于可视化）
   */
  async getKnowledgeSubgraph(
    centerNodeId: string,
    depth: number = 2
  ): Promise<{ nodes: unknown[]; links: unknown[] }> {
    return this.neo4jTool.getSubgraph(centerNodeId, depth);
  }

  /**
   * 更新配置
   */
  updateConfig(newConfig: Partial<GraphRAGConfig>): void {
    this.config = { ...this.config, ...newConfig };
    logger.info('GraphRAG config updated', this.config);
  }

  /**
   * 获取当前配置
   */
  getConfig(): GraphRAGConfig {
    return { ...this.config };
  }
}

// 单例模式
let instance: GraphRAG | null = null;

export function getGraphRAG(config?: Partial<GraphRAGConfig>): GraphRAG {
  if (!instance) {
    instance = new GraphRAG(config);
  }
  return instance;
}

export default GraphRAG;
