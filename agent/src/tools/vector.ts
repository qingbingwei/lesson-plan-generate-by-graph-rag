import { getDeepSeekClient } from '../clients/deepseek';
import logger from '../utils/logger';
import config from '../config';

/**
 * 向量搜索工具类
 * 用于处理 Embedding 生成和相似度计算
 */
class VectorTool {
  private deepseekClient = getDeepSeekClient();
  private embeddingDimension = config.graphRag.embeddingDimension;

  /**
   * 生成文本的 Embedding 向量
   */
  async createEmbedding(text: string): Promise<number[]> {
    try {
      const embedding = await this.deepseekClient.createEmbedding(text);
      return embedding;
    } catch (error) {
      logger.error('Failed to create embedding', { error, textLength: text.length });
      throw error;
    }
  }

  /**
   * 批量生成 Embedding 向量
   */
  async createEmbeddings(texts: string[]): Promise<number[][]> {
    try {
      const embeddings = await this.deepseekClient.createEmbeddings(texts);
      return embeddings;
    } catch (error) {
      logger.error('Failed to create batch embeddings', { error, count: texts.length });
      throw error;
    }
  }

  /**
   * 计算余弦相似度
   */
  cosineSimilarity(vecA: number[], vecB: number[]): number {
    if (vecA.length !== vecB.length) {
      throw new Error('Vectors must have the same dimension');
    }

    let dotProduct = 0;
    let normA = 0;
    let normB = 0;

    for (let i = 0; i < vecA.length; i++) {
      const a = vecA[i] ?? 0;
      const b = vecB[i] ?? 0;
      dotProduct += a * b;
      normA += a * a;
      normB += b * b;
    }

    normA = Math.sqrt(normA);
    normB = Math.sqrt(normB);

    if (normA === 0 || normB === 0) {
      return 0;
    }

    return dotProduct / (normA * normB);
  }

  /**
   * 计算欧几里得距离
   */
  euclideanDistance(vecA: number[], vecB: number[]): number {
    if (vecA.length !== vecB.length) {
      throw new Error('Vectors must have the same dimension');
    }

    let sum = 0;
    for (let i = 0; i < vecA.length; i++) {
      const diff = (vecA[i] ?? 0) - (vecB[i] ?? 0);
      sum += diff * diff;
    }

    return Math.sqrt(sum);
  }

  /**
   * 在向量集合中搜索最相似的向量
   */
  async searchSimilar(
    query: string,
    vectors: { id: string; embedding: number[] }[],
    topK: number = 10
  ): Promise<{ id: string; score: number }[]> {
    try {
      // 生成查询向量
      const queryEmbedding = await this.createEmbedding(query);

      // 计算所有相似度
      const similarities = vectors.map(v => ({
        id: v.id,
        score: this.cosineSimilarity(queryEmbedding, v.embedding),
      }));

      // 按相似度降序排序并取前 topK
      return similarities
        .sort((a, b) => b.score - a.score)
        .slice(0, topK);
    } catch (error) {
      logger.error('Failed to search similar vectors', { error, query });
      throw error;
    }
  }

  /**
   * 批量计算相似度矩阵
   */
  computeSimilarityMatrix(vectors: number[][]): number[][] {
    const n = vectors.length;
    const matrix: number[][] = Array(n).fill(null).map(() => Array(n).fill(0) as number[]);

    for (let i = 0; i < n; i++) {
      for (let j = i; j < n; j++) {
        const vecI = vectors[i];
        const vecJ = vectors[j];
        const rowI = matrix[i];
        const rowJ = matrix[j];
        if (vecI && vecJ && rowI && rowJ) {
          const similarity = this.cosineSimilarity(vecI, vecJ);
          rowI[j] = similarity;
          rowJ[i] = similarity; // 对称矩阵
        }
      }
    }

    return matrix;
  }

  /**
   * 向量归一化
   */
  normalize(vector: number[]): number[] {
    const norm = Math.sqrt(vector.reduce((sum, v) => sum + v * v, 0));
    if (norm === 0) {
      return vector;
    }
    return vector.map(v => v / norm);
  }

  /**
   * 向量加权平均
   */
  weightedAverage(
    vectors: number[][],
    weights: number[]
  ): number[] {
    if (vectors.length !== weights.length) {
      throw new Error('Vectors and weights must have the same length');
    }

    if (vectors.length === 0) {
      return [];
    }

    const firstVector = vectors[0];
    if (!firstVector) {
      return [];
    }

    const dimension = firstVector.length;
    const result = new Array<number>(dimension).fill(0);
    const totalWeight = weights.reduce((sum, w) => sum + w, 0);

    for (let i = 0; i < vectors.length; i++) {
      const vec = vectors[i];
      const weight = weights[i] ?? 0;
      if (vec) {
        const normalizedWeight = weight / totalWeight;
        for (let j = 0; j < dimension; j++) {
          result[j] = (result[j] ?? 0) + (vec[j] ?? 0) * normalizedWeight;
        }
      }
    }

    return result;
  }

  /**
   * 检查 Embedding 维度是否正确
   */
  validateEmbedding(embedding: number[]): boolean {
    return embedding.length === this.embeddingDimension;
  }

  /**
   * 获取 Embedding 维度
   */
  getEmbeddingDimension(): number {
    return this.embeddingDimension;
  }
}

// 单例模式
let instance: VectorTool | null = null;

export function getVectorTool(): VectorTool {
  if (!instance) {
    instance = new VectorTool();
  }
  return instance;
}

export default VectorTool;
