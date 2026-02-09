import { generateObjectives, validateObjectives, generateKeyDifficultPoints, generateTeachingMethods } from '../skills';
import logger from '../utils/logger';
import { mergeUsage } from '../utils/tokenUsage';
import type { WorkflowState } from '../types';

/**
 * 目标设计节点
 * 生成三维教学目标和重点难点
 */
export async function objectiveDesignNode(state: WorkflowState): Promise<Partial<WorkflowState>> {
  const startTime = Date.now();
  logger.info('ObjectiveDesignNode executing', { topic: state.input.topic });

  // 如果已有错误，直接返回
  if (state.error) {
    return {};
  }

  try {
    // 生成教学目标
    const { objectives, usage: objectiveUsage } = await generateObjectives(
      state.input,
      state.knowledgeContext || []
    );

    // 验证目标
    if (!validateObjectives(objectives)) {
      throw new Error('生成的教学目标不完整');
    }

    // 生成重点难点
    const { keyPoints, difficultPoints, usage: pointsUsage } = 
      await generateKeyDifficultPoints(
        state.input,
        objectives,
        state.knowledgeContext || []
      );

    // 生成教学方法
    const { methods, usage: methodsUsage } = await generateTeachingMethods(
      state.input,
      objectives
    );

    // 合并 token 使用
    const totalUsage = mergeUsage(objectiveUsage, pointsUsage, methodsUsage);

    logger.info('ObjectiveDesignNode completed', {
      duration: Date.now() - startTime,
      usage: totalUsage,
    });

    return {
      lessonObjectives: objectives,
      keyPoints,
      difficultPoints,
      teachingMethods: methods,
      usage: mergeUsage(state.usage || { promptTokens: 0, completionTokens: 0, totalTokens: 0 }, totalUsage),
    };
  } catch (error) {
    logger.error('ObjectiveDesignNode failed', { error });
    return {
      error: error instanceof Error ? error.message : 'Objective design failed',
    };
  }
}

export default objectiveDesignNode;
