import logger from '../utils/logger';
import { normalizeGrade } from '../utils/gradeNormalizer';
import type { WorkflowState, GenerateLessonRequest } from '../types';

/**
 * 输入分析节点
 * 解析和验证用户输入，初始化工作流状态
 */
export async function inputAnalysisNode(state: WorkflowState): Promise<Partial<WorkflowState>> {
  const startTime = Date.now();
  
  // 添加保护性检查
  if (!state || !state.input) {
    logger.error('InputAnalysisNode: Invalid state', { state });
    return {
      error: 'Invalid workflow state: missing input',
    };
  }
  
  logger.info('InputAnalysisNode executing', { topic: state.input.topic });

  try {
    // 验证输入
    validateInput(state.input);

    // 规范化输入
    const normalizedInput = normalizeInput(state.input);

    logger.info('InputAnalysisNode completed', {
      duration: Date.now() - startTime,
      subject: normalizedInput.subject,
      grade: normalizedInput.grade,
      topic: normalizedInput.topic,
    });

    return {
      input: normalizedInput,
      startTime: Date.now(),
    };
  } catch (error) {
    logger.error('InputAnalysisNode failed', { error });
    return {
      error: error instanceof Error ? error.message : 'Input analysis failed',
    };
  }
}

/**
 * 验证输入
 */
function validateInput(input: GenerateLessonRequest): void {
  // 必填字段验证
  if (!input.subject || input.subject.trim() === '') {
    throw new Error('学科不能为空');
  }

  if (!input.grade || input.grade.trim() === '') {
    throw new Error('年级不能为空');
  }

  if (!input.topic || input.topic.trim() === '') {
    throw new Error('课题不能为空');
  }

  if (!input.duration || input.duration <= 0) {
    throw new Error('课时时长必须大于0');
  }

  // 合理性验证
  if (input.duration < 20 || input.duration > 180) {
    throw new Error('课时时长应在20-180分钟之间');
  }

  // 学科验证
  const validSubjects = [
    '语文', '数学', '英语', '物理', '化学', '生物',
    '历史', '地理', '政治', '思想品德', '道德与法治',
    '科学', '信息技术', '音乐', '美术', '体育',
    'Chinese', 'Mathematics', 'English', 'Physics', 'Chemistry', 'Biology',
    'History', 'Geography', 'Politics', 'Science', 'IT', 'Music', 'Art', 'PE',
  ];

  const subjectLower = input.subject.toLowerCase();
  const isValidSubject = validSubjects.some(
    s => s.toLowerCase() === subjectLower || input.subject.includes(s) || s.includes(input.subject)
  );

  if (!isValidSubject) {
    logger.warn('Unknown subject', { subject: input.subject });
    // 不抛出错误，只是警告
  }
}

/**
 * 规范化输入
 */
function normalizeInput(input: GenerateLessonRequest): GenerateLessonRequest {
  return {
    subject: input.subject.trim(),
    grade: normalizeGrade(input.grade.trim()),
    topic: input.topic.trim(),
    duration: input.duration,
    style: input.style?.trim(),
    requirements: input.requirements?.trim(),
    context: input.context,
  };
}

export default inputAnalysisNode;
