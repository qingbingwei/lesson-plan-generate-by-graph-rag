import logger from '../../shared/utils/logger';
import config from '../../config';
import { getRequestApiKeys } from '../../shared/context/requestApiKeys';
import { getTraceIdFromContext } from '../../shared/context/traceContext';
import { recordDownstream } from '../../shared/observability/metrics';

/**
 * 千问 API 客户端
 * 用于 Embedding 生成
 */
class QwenClient {
  private apiKey: string;
  private embeddingUrl: string;
  private embeddingModel: string;

  constructor() {
    this.apiKey = config.qwen.apiKey;
    this.embeddingUrl = config.qwen.embeddingUrl;
    this.embeddingModel = config.qwen.embeddingModel;

    if (!this.apiKey) {
      logger.warn('QWEN_API_KEY is not set, embedding will not work');
    }
  }

  /**
   * 生成 Embedding
   * 使用千问 text-embedding-v4 模型
   */
  async createEmbedding(text: string, overrideApiKey?: string): Promise<number[]> {
    const startTime = Date.now();
    const traceId = getTraceIdFromContext();
    let statusCode = 0;

    try {
      logger.debug('Creating embedding with Qwen', { trace_id: traceId, textLength: text.length });

      const runtimeApiKey = (overrideApiKey || getRequestApiKeys().embeddingApiKey || this.apiKey || '').trim();
      if (!runtimeApiKey) {
        throw new Error('QWEN_API_KEY is not set');
      }

      const response = await fetch(this.embeddingUrl, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${runtimeApiKey}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          model: this.embeddingModel,
          input: text,
        }),
      });
      statusCode = response.status;

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(`Qwen Embedding API error: ${response.status} - ${JSON.stringify(errorData)}`);
      }

      const data = await response.json() as { data?: Array<{ embedding: number[] }> };
      const embedding = data.data?.[0]?.embedding || [];
      const duration = Date.now() - startTime;
      recordDownstream('qwen', 'embedding', statusCode, duration);
      
      logger.debug('Embedding created with Qwen', {
        trace_id: traceId,
        duration,
        dimension: embedding.length,
      });
      
      return embedding;
    } catch (error) {
      const duration = Date.now() - startTime;
      recordDownstream('qwen', 'embedding', statusCode, duration);
      logger.error('Qwen embedding creation error', {
        trace_id: traceId,
        status: statusCode,
        duration,
        error,
      });
      throw error;
    }
  }

  /**
   * 批量生成 Embedding
   * 使用千问 text-embedding-v4 模型
   */
  async createEmbeddings(texts: string[], overrideApiKey?: string): Promise<number[][]> {
    const startTime = Date.now();
    const traceId = getTraceIdFromContext();
    let statusCode = 0;

    try {
      logger.debug('Creating batch embeddings with Qwen', { trace_id: traceId, count: texts.length });

      const runtimeApiKey = (overrideApiKey || getRequestApiKeys().embeddingApiKey || this.apiKey || '').trim();
      if (!runtimeApiKey) {
        throw new Error('QWEN_API_KEY is not set');
      }

      const response = await fetch(this.embeddingUrl, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${runtimeApiKey}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          model: this.embeddingModel,
          input: texts,
        }),
      });
      statusCode = response.status;

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(`Qwen Embedding API error: ${response.status} - ${JSON.stringify(errorData)}`);
      }

      const data = await response.json() as { data?: Array<{ embedding: number[] }> };
      const embeddings = data.data?.map((d) => d.embedding) || [];
      const duration = Date.now() - startTime;
      recordDownstream('qwen', 'embedding_batch', statusCode, duration);
      
      logger.debug('Batch embeddings created with Qwen', {
        trace_id: traceId,
        duration,
        count: embeddings.length,
        dimension: embeddings[0]?.length || 0,
      });
      
      return embeddings;
    } catch (error) {
      const duration = Date.now() - startTime;
      recordDownstream('qwen', 'embedding_batch', statusCode, duration);
      logger.error('Qwen batch embedding creation error', {
        trace_id: traceId,
        status: statusCode,
        duration,
        error,
      });
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
