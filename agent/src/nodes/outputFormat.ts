import logger from '../utils/logger';
import type { WorkflowState, GeneratedLesson } from '../types';

/**
 * 输出格式化节点
 * 将所有生成的内容整合成最终的教案格式
 */
export async function outputFormatNode(state: WorkflowState): Promise<Partial<WorkflowState>> {
  const startTime = Date.now();
  logger.info('OutputFormatNode executing', { topic: state.input.topic });

  // 如果已有错误，直接返回
  if (state.error) {
    return {
      endTime: Date.now(),
    };
  }

  try {
    // 验证必要的字段
    if (!state.lessonObjectives) {
      throw new Error('缺少教学目标');
    }

    if (!state.sections || state.sections.length === 0) {
      throw new Error('缺少教学环节');
    }

    // 生成标题
    const title = generateTitle(state);

    // 组装最终教案
    const output: GeneratedLesson = {
      title,
      objectives: state.lessonObjectives,
      keyPoints: state.keyPoints || [],
      difficultPoints: state.difficultPoints || [],
      teachingMethods: state.teachingMethods || [],
      content: {
        sections: state.sections,
        materials: state.materials || [],
        homework: state.homework || '',
      },
      evaluation: state.evaluation || '',
      reflection: generateReflectionTemplate(state),
    };

    // 验证输出
    validateOutput(output);

    logger.info('OutputFormatNode completed', {
      duration: Date.now() - startTime,
      title: output.title,
      sectionCount: output.content.sections.length,
    });

    return {
      output,
      endTime: Date.now(),
    };
  } catch (error) {
    logger.error('OutputFormatNode failed', { error });
    return {
      error: error instanceof Error ? error.message : 'Output format failed',
      endTime: Date.now(),
    };
  }
}

/**
 * 生成教案标题
 */
function generateTitle(state: WorkflowState): string {
  const { subject, grade, topic } = state.input;
  return `${subject} ${grade} 《${topic}》教学设计`;
}

/**
 * 生成教学反思模板
 */
function generateReflectionTemplate(state: WorkflowState): string {
  return `## 教学反思

### 一、目标达成情况
1. 知识与技能目标是否达成？
2. 过程与方法目标是否达成？
3. 情感态度价值观目标是否达成？

### 二、教学过程反思
1. 导入环节是否有效激发了学生兴趣？
2. 新授环节是否突出了重点、突破了难点？
3. 练习环节是否起到了巩固作用？
4. 时间分配是否合理？

### 三、学生表现
1. 学生参与度如何？
2. 学生的理解程度如何？
3. 有哪些意料之外的问题？

### 四、改进措施
1. 教学内容方面需要如何改进？
2. 教学方法方面需要如何改进？
3. 教学评价方面需要如何改进？

---
课题：${state.input.topic}
授课时间：______
授课班级：______
反思时间：______
`;
}

/**
 * 验证输出
 */
function validateOutput(output: GeneratedLesson): void {
  // 检查必填字段
  if (!output.title) {
    throw new Error('教案标题不能为空');
  }

  if (!output.objectives || !output.objectives.knowledge) {
    throw new Error('教学目标不完整');
  }

  if (!output.content.sections || output.content.sections.length === 0) {
    throw new Error('教学环节不能为空');
  }

  // 检查每个环节的完整性
  for (const section of output.content.sections) {
    if (!section.title) {
      throw new Error('教学环节标题不能为空');
    }
    if (!section.duration || section.duration <= 0) {
      throw new Error(`环节"${section.title}"的时长必须大于0`);
    }
  }

  // 检查重点难点
  if (!output.keyPoints || output.keyPoints.length === 0) {
    logger.warn('教学重点为空');
  }

  if (!output.difficultPoints || output.difficultPoints.length === 0) {
    logger.warn('教学难点为空');
  }
}

export default outputFormatNode;
