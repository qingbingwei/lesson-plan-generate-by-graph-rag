import OpenAI from 'openai';
import config from '../../config';
import logger from '../../shared/utils/logger';
import { getQwenClient } from './qwen';
import { getRequestApiKeys } from '../../shared/context/requestApiKeys';
import type { DeepSeekMessage, TokenUsage } from '../../shared/types';

/**
 * DeepSeek API 客户端
 * 使用 OpenAI SDK 兼容接口
 */
class DeepSeekClient {
  private client: OpenAI;
  private model: string;
  private temperature: number;
  private maxTokens: number;
  private qwenClient = getQwenClient();

  constructor() {
    this.client = new OpenAI({
      apiKey: config.deepseek.apiKey,
      baseURL: config.deepseek.baseUrl,
    });
    this.model = config.deepseek.model;
    this.temperature = config.deepseek.temperature;
    this.maxTokens = config.deepseek.maxTokens;
  }

  /**
   * 发送聊天请求
   */
  async chat(
    messages: DeepSeekMessage[],
    options?: {
      temperature?: number;
      maxTokens?: number;
      stream?: boolean;
    }
  ): Promise<{ content: string; usage: TokenUsage }> {
    const startTime = Date.now();
    
    try {
      logger.debug('DeepSeek chat request', {
        model: this.model,
        messageCount: messages.length,
        temperature: options?.temperature ?? this.temperature,
      });

      const { generationApiKey } = getRequestApiKeys();
      const runtimeClient = generationApiKey
        ? new OpenAI({
            apiKey: generationApiKey,
            baseURL: config.deepseek.baseUrl,
          })
        : this.client;

      const response = await runtimeClient.chat.completions.create({
        model: this.model,
        messages: messages.map(m => ({
          role: m.role,
          content: m.content,
        })),
        temperature: options?.temperature ?? this.temperature,
        max_tokens: options?.maxTokens ?? this.maxTokens,
        stream: false,
      });

      const content = response.choices[0]?.message?.content || '';
      const usage: TokenUsage = {
        promptTokens: response.usage?.prompt_tokens || 0,
        completionTokens: response.usage?.completion_tokens || 0,
        totalTokens: response.usage?.total_tokens || 0,
      };

      logger.info('DeepSeek chat completed', {
        duration: Date.now() - startTime,
        usage,
      });

      return { content, usage };
    } catch (error) {
      logger.error('DeepSeek chat error', { error });
      throw error;
    }
  }

  /**
   * 生成 Embedding
   * 委托给千问客户端
   */
  async createEmbedding(text: string): Promise<number[]> {
    const { embeddingApiKey } = getRequestApiKeys();
    return this.qwenClient.createEmbedding(text, embeddingApiKey);
  }

  /**
   * 批量生成 Embedding
   * 委托给千问客户端
   */
  async createEmbeddings(texts: string[]): Promise<number[][]> {
    const { embeddingApiKey } = getRequestApiKeys();
    return this.qwenClient.createEmbeddings(texts, embeddingApiKey);
  }

  /**
   * 结构化输出请求
   * 使用 JSON 模式获取结构化响应
   */
  async structuredChat<T>(
    messages: DeepSeekMessage[],
    schema: string,
    options?: {
      temperature?: number;
      maxTokens?: number;
    }
  ): Promise<{ data: T; usage: TokenUsage }> {
    // 添加 JSON 输出提示
    const systemPrompt: DeepSeekMessage = {
      role: 'system',
      content: `You must respond with valid JSON that matches this schema:\n${schema}\n\nDo not include any text outside the JSON object.`,
    };

    const allMessages = [systemPrompt, ...messages];
    
    const { content, usage } = await this.chat(allMessages, {
      ...options,
      temperature: options?.temperature ?? 0.3, // 降低温度以获得更稳定的JSON输出
    });

    try {
      // 提取 JSON
      const jsonMatch = content.match(/\{[\s\S]*\}/);
      if (!jsonMatch) {
        throw new Error('No JSON found in response');
      }
      
      const data = JSON.parse(jsonMatch[0]) as T;
      return { data, usage };
    } catch (parseError) {
      logger.error('Failed to parse structured response', {
        content,
        error: parseError,
      });
      throw new Error(`Failed to parse structured response: ${parseError}`);
    }
  }
}

// 单例模式
let instance: DeepSeekClient | null = null;

export function getDeepSeekClient(): DeepSeekClient {
  if (!instance) {
    instance = new DeepSeekClient();
  }
  return instance;
}

export default DeepSeekClient;
