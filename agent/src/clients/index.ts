import { getDeepSeekClient } from './deepseek';

/**
 * 单例 DeepSeek 客户端
 * 统一版本 - 替代 objectiveGeneration/contentGeneration/evaluationDesign 中的重复实现
 */
let clientInstance: ReturnType<typeof getDeepSeekClient> | null = null;

export function getClient() {
  if (!clientInstance) {
    clientInstance = getDeepSeekClient();
  }
  return clientInstance;
}
