import { retrieveKnowledge } from '../skills';
import logger from '../utils/logger';
import type { WorkflowState } from '../types';

/**
 * 知识查询节点
 * 从知识图谱中检索相关知识点
 */
export async function knowledgeQueryNode(state: WorkflowState): Promise<Partial<WorkflowState>> {
  const startTime = Date.now();
  logger.info('KnowledgeQueryNode executing', {
    subject: state.input.subject,
    grade: state.input.grade,
    topic: state.input.topic,
  });

  // 如果已有错误，直接返回
  if (state.error) {
    return {};
  }

  try {
    const contexts = await retrieveKnowledge(state.input);

    logger.info('KnowledgeQueryNode completed', {
      duration: Date.now() - startTime,
      contextCount: contexts.length,
    });

    return {
      knowledgeContext: contexts,
    };
  } catch (error) {
    logger.error('KnowledgeQueryNode failed', { error });
    
    // 知识检索失败不是致命错误，继续使用空上下文
    logger.warn('Continuing with empty knowledge context');
    return {
      knowledgeContext: [],
    };
  }
}

export default knowledgeQueryNode;
