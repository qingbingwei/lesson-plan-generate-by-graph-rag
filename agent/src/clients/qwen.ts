import logger from '../utils/logger';

/**
 * 千问 API 客户端
 * 用于 Embedding 生成
 */
class QwenClient {
  private apiKey: string;
  private embeddingUrl: string;
  private embeddingModel: string;

  constructor() {
    this.apiKey = process.env.QWEN_API_KEY || '';
    this.embeddingUrl = process.env.QWEN_EMBEDDING_URL || 'https://dashscope.aliyuncs.com/compatible-mode/v1/embeddings';
    this.embeddingModel = process.env.QWEN_EMBEDDING_MODEL || 'text-embedding-v4';

    if (!this.apiKey) {
      logger.warn('QWEN_API_KEY is not set, embedding will not work');
    }
  }

  /**
   * 生成 Embedding
   * 使用千问 text-embedding-v4 模型
   */
  async createEmbedding(text: string): Promise<number[]> {
    try {
      logger.debug('Creating embedding with Qwen', { textLength: text.length });

      const response = await fetch(this.embeddingUrl, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${this.apiKey}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          model: this.embeddingModel,
          input: text,
        }),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(`Qwen Embedding API error: ${response.status} - ${JSON.stringify(errorData)}`);
      }

      const data = await response.json() as { data?: Array<{ embedding: number[] }> };
      const embedding = data.data?.[0]?.embedding || [];
      
      logger.debug('Embedding created with Qwen', { dimension: embedding.length });
      
      return embedding;
    } catch (error) {
      logger.error('Qwen embedding creation error', { error });
      throw error;
    }
  }

  /**
   * 批量生成 Embedding
   * 使用千问 text-embedding-v4 模型
   */
  async createEmbeddings(texts: string[]): Promise<number[][]> {
    try {
      logger.debug('Creating batch embeddings with Qwen', { count: texts.length });

      const response = await fetch(this.embeddingUrl, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${this.apiKey}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          model: this.embeddingModel,
          input: texts,
        }),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(`Qwen Embedding API error: ${response.status} - ${JSON.stringify(errorData)}`);
      }

      const data = await response.json() as { data?: Array<{ embedding: number[] }> };
      const embeddings = data.data?.map((d) => d.embedding) || [];
      
      logger.debug('Batch embeddings created with Qwen', {
        count: embeddings.length,
        dimension: embeddings[0]?.length || 0,
      });
      
      return embeddings;
    } catch (error) {
      logger.error('Qwen batch embedding creation error', { error });
      throw error;
    }
  }
}

// 单例模式
let instance: QwenClient | null = null;

export function getQwenClient(): QwenClient {
  if (!instance) {
    instance = new QwenClient();
  }
  return instance;
}

export default QwenClient;
