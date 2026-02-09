import { generateSections } from '../skills';
import logger from '../utils/logger';
import { mergeUsage } from '../utils/tokenUsage';
import type { WorkflowState } from '../types';

/**
 * 内容设计节点
 * 生成教学环节和内容
 */
export async function contentDesignNode(state: WorkflowState): Promise<Partial<WorkflowState>> {
  const startTime = Date.now();
  logger.info('ContentDesignNode executing', { topic: state.input.topic });

  // 如果已有错误，直接返回
  if (state.error) {
    return {};
  }

  // 检查前置条件
  if (!state.lessonObjectives) {
    return {
      error: '缺少教学目标，无法生成教学内容',
    };
  }

  try {
    // 生成教学环节
    const { sections, usage } = await generateSections(
      state.input,
      state.lessonObjectives,
      state.keyPoints || [],
      state.difficultPoints || [],
      state.knowledgeContext || []
    );

    // 验证环节
    if (!sections || sections.length === 0) {
      throw new Error('未能生成教学环节');
    }

    // 验证时间分配
    const totalDuration = sections.reduce((sum, s) => sum + (s.duration || 0), 0);
    if (Math.abs(totalDuration - state.input.duration) > 5) {
      logger.warn('Section durations do not match total duration', {
        totalDuration,
        expectedDuration: state.input.duration,
      });
    }

    logger.info('ContentDesignNode completed', {
      duration: Date.now() - startTime,
      sectionCount: sections.length,
      usage,
    });

    return {
      sections,
      usage: mergeUsage(state.usage, usage),
    };
  } catch (error) {
    logger.error('ContentDesignNode failed', { error });
    return {
      error: error instanceof Error ? error.message : 'Content design failed',
    };
  }
}

export default contentDesignNode;
