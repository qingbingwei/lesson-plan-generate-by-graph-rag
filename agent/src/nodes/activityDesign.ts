import { generateMaterials, generateHomework, generateEvaluation } from '../skills';
import logger from '../utils/logger';
import type { WorkflowState, TokenUsage } from '../types';

/**
 * 活动设计节点
 * 生成教学资源、作业和评价方案
 */
export async function activityDesignNode(state: WorkflowState): Promise<Partial<WorkflowState>> {
  const startTime = Date.now();
  logger.info('ActivityDesignNode executing', { topic: state.input.topic });

  // 如果已有错误，直接返回
  if (state.error) {
    return {};
  }

  // 检查前置条件
  if (!state.sections || state.sections.length === 0) {
    return {
      error: '缺少教学环节，无法生成活动设计',
    };
  }

  if (!state.lessonObjectives) {
    return {
      error: '缺少教学目标，无法生成活动设计',
    };
  }

  try {
    // 并行生成教学资源、作业和评价
    const [
      { materials, usage: materialsUsage },
      { homework, usage: homeworkUsage },
      { evaluation, usage: evaluationUsage },
    ] = await Promise.all([
      generateMaterials(state.input, state.sections),
      generateHomework(
        state.input,
        state.lessonObjectives,
        state.keyPoints || []
      ),
      generateEvaluation(
        state.input,
        state.lessonObjectives,
        state.keyPoints || [],
        state.sections
      ),
    ]);

    // 合并 token 使用
    const totalUsage = mergeUsage(materialsUsage, homeworkUsage, evaluationUsage);

    logger.info('ActivityDesignNode completed', {
      duration: Date.now() - startTime,
      materialCount: materials.length,
      usage: totalUsage,
    });

    return {
      materials,
      homework,
      evaluation,
      usage: mergeUsage(state.usage || { promptTokens: 0, completionTokens: 0, totalTokens: 0 }, totalUsage),
    };
  } catch (error) {
    logger.error('ActivityDesignNode failed', { error });
    return {
      error: error instanceof Error ? error.message : 'Activity design failed',
    };
  }
}

/**
 * 合并 token 使用统计
 */
function mergeUsage(...usages: (TokenUsage | undefined)[]): TokenUsage {
  return usages.reduce<TokenUsage>(
    (acc, usage) => ({
      promptTokens: acc.promptTokens + (usage?.promptTokens || 0),
      completionTokens: acc.completionTokens + (usage?.completionTokens || 0),
      totalTokens: acc.totalTokens + (usage?.totalTokens || 0),
    }),
    { promptTokens: 0, completionTokens: 0, totalTokens: 0 }
  );
}

export default activityDesignNode;
