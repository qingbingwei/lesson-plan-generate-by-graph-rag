import type { TokenUsage } from '../types';

/**
 * 合并多个 TokenUsage 统计
 * 统一版本 - 替代 objectiveDesign/contentDesign/activityDesign 中的重复实现
 */
export function mergeUsage(...usages: (TokenUsage | undefined)[]): TokenUsage {
  return usages.reduce<TokenUsage>(
    (acc, usage) => ({
      promptTokens: acc.promptTokens + (usage?.promptTokens || 0),
      completionTokens: acc.completionTokens + (usage?.completionTokens || 0),
      totalTokens: acc.totalTokens + (usage?.totalTokens || 0),
    }),
    { promptTokens: 0, completionTokens: 0, totalTokens: 0 }
  );
}
