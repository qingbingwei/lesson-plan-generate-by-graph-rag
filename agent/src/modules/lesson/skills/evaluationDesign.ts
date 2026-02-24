import { getDeepSeekClient } from '../../../infrastructure/clients/deepseek';
import logger from '../../../shared/utils/logger';
import type { Skill } from './index';
import type { LessonObjectives, LessonSection, GenerateLessonRequest, TokenUsage } from '../../../shared/types';

/**
 * 评价设计 Skill 定义
 */
export const evaluationDesignSkill: Skill = {
  name: 'evaluation_design',
  description: '设计科学合理的教学评价方案，包括过程性评价、终结性评价和评价量规',
  content: `
# 评价设计技能

## 功能说明
设计科学合理的教学评价方案，帮助教师评估学生学习效果。

## 评价设计原则
1. 评价目标与教学目标一致
2. 评价方式多样化（观察、提问、练习、作品、测试等）
3. 过程性评价与终结性评价相结合
4. 自评、互评、师评相结合
5. 评价要具有诊断、反馈、激励功能

## 评价设计内容
1. **评价目标**: 明确评价要达成的目的
2. **评价内容**: 确定评价的具体方面
3. **评价标准**: 制定清晰的评价标准
4. **评价方式**: 选择适当的评价方法
5. **评价工具**: 设计必要的评价工具

## 附加功能
- 生成评价量规
- 设计形成性评价方案
- 生成教学反思框架
`,
};

/**
 * 构建系统提示词
 */
function buildSystemPrompt(): string {
  return `你是一位教学评价专家，擅长设计科学合理的教学评价方案。

教学评价设计原则：
1. 评价目标与教学目标一致
2. 评价方式多样化（观察、提问、练习、作品、测试等）
3. 过程性评价与终结性评价相结合
4. 自评、互评、师评相结合
5. 评价要具有诊断、反馈、激励功能

评价设计内容：
1. 评价目标：明确评价要达成的目的
2. 评价内容：确定评价的具体方面
3. 评价标准：制定清晰的评价标准
4. 评价方式：选择适当的评价方法
5. 评价工具：设计必要的评价工具`;
}

/**
 * 构建用户提示词
 */
function buildUserPrompt(
  request: GenerateLessonRequest,
  objectives: LessonObjectives,
  keyPoints: string[],
  sections: LessonSection[]
): string {
  const sectionsSummary = sections.map((s) => `${s.title}（${s.duration}分钟）`).join('、');

  return `请为以下课程设计教学评价方案：

基本信息：
- 学科：${request.subject}
- 年级：${request.grade}
- 课题：${request.topic}

教学目标：
- 知识与技能：${objectives.knowledge}
- 过程与方法：${objectives.process}
- 情感态度价值观：${objectives.emotion}

教学重点：
${keyPoints.map((p) => `- ${p}`).join('\n')}

教学环节：${sectionsSummary}

请设计包含以下内容的评价方案：
1. 课堂评价（过程性评价）
   - 每个环节的评价要点
   - 评价方式和工具
2. 学习效果评价（终结性评价）
   - 评价标准
   - 评价方式
3. 评价建议
   - 对教师的评价建议
   - 对学生的评价建议`;
}

/**
 * 生成教学评价设计
 */
export async function generateEvaluation(
  request: GenerateLessonRequest,
  objectives: LessonObjectives,
  keyPoints: string[],
  sections: LessonSection[]
): Promise<{ evaluation: string; usage: TokenUsage }> {
  const startTime = Date.now();
  logger.info('EvaluationDesign executing', {
    subject: request.subject,
    topic: request.topic,
  });

  try {
    const client = getDeepSeekClient();
    const systemPrompt = buildSystemPrompt();
    const userPrompt = buildUserPrompt(request, objectives, keyPoints, sections);

    const { content, usage } = await client.chat(
      [
        { role: 'system', content: systemPrompt },
        { role: 'user', content: userPrompt },
      ],
      { temperature: 0.6 }
    );

    logger.info('EvaluationDesign completed', {
      duration: Date.now() - startTime,
      usage,
    });

    return { evaluation: content, usage };
  } catch (error) {
    logger.error('EvaluationDesign failed', { error, request });
    throw error;
  }
}
