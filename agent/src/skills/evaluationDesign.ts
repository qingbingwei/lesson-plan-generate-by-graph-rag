import { getClient } from '../clients';
import logger from '../utils/logger';
import type { Skill } from './index';
import type { LessonObjectives, LessonSection, GenerateLessonRequest, TokenUsage } from '../types';

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
    const client = getClient();
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

/**
 * 生成评价量规
 */
export async function generateRubric(
  request: GenerateLessonRequest,
  objectives: LessonObjectives
): Promise<{ rubric: string; usage: TokenUsage }> {
  logger.info('EvaluationDesign generating rubric');

  try {
    const client = getClient();

    const prompt = `请为以下课程设计评价量规：

学科：${request.subject}
年级：${request.grade}
课题：${request.topic}

教学目标：
- 知识与技能：${objectives.knowledge}
- 过程与方法：${objectives.process}
- 情感态度价值观：${objectives.emotion}

请设计一个包含以下维度的评价量规：
1. 评价维度（3-4个维度）
2. 每个维度的评价标准（优秀/良好/合格/待提高）
3. 每个等级的具体描述

要求：
- 评价标准要具体、可操作
- 与教学目标紧密对应
- 语言简洁明了`;

    const { content, usage } = await client.chat(
      [
        { role: 'system', content: '你是一位教学评价专家，擅长设计科学的评价量规。' },
        { role: 'user', content: prompt },
      ],
      { temperature: 0.5 }
    );

    return { rubric: content, usage };
  } catch (error) {
    logger.error('Failed to generate rubric', { error });
    throw error;
  }
}

/**
 * 生成形成性评价方案
 */
export async function generateFormativeAssessment(
  sections: LessonSection[]
): Promise<{ assessment: string; usage: TokenUsage }> {
  logger.info('EvaluationDesign generating formative assessment');

  try {
    const client = getClient();

    const sectionsSummary = sections
      .map((s) => `${s.title}（${s.duration}分钟）：${s.content?.slice(0, 100) || ''}`)
      .join('\n');

    const prompt = `请为以下教学过程设计形成性评价方案：

教学环节：
${sectionsSummary}

请为每个教学环节设计：
1. 评价时机（何时进行评价）
2. 评价方式（如：提问、观察、练习、互评等）
3. 评价内容（评价什么）
4. 反馈方式（如何给予学生反馈）

要求：
- 评价要贯穿教学全过程
- 评价方式要多样化
- 注重过程性评价
- 强调评价的诊断和改进功能`;

    const { content, usage } = await client.chat(
      [
        { role: 'system', content: '你是一位教学评价专家，擅长设计形成性评价方案。' },
        { role: 'user', content: prompt },
      ],
      { temperature: 0.6 }
    );

    return { assessment: content, usage };
  } catch (error) {
    logger.error('Failed to generate formative assessment', { error });
    throw error;
  }
}

/**
 * 生成教学反思框架
 */
export async function generateReflectionFramework(
  request: GenerateLessonRequest,
  objectives: LessonObjectives
): Promise<{ reflection: string; usage: TokenUsage }> {
  logger.info('EvaluationDesign generating reflection framework');

  try {
    const client = getClient();

    const prompt = `请为以下课程设计教学反思框架：

学科：${request.subject}
课题：${request.topic}

教学目标：
${objectives.knowledge}
${objectives.process}
${objectives.emotion}

请提供教学反思的框架，包括：
1. 目标达成反思（教学目标是否达成）
2. 教学过程反思（教学活动是否有效）
3. 学生表现反思（学生的学习效果）
4. 改进措施（下次教学如何改进）

每个方面提供2-3个引导性问题。`;

    const { content, usage } = await client.chat(
      [
        { role: 'system', content: '你是一位教学专家，擅长引导教师进行教学反思。' },
        { role: 'user', content: prompt },
      ],
      { temperature: 0.5 }
    );

    return { reflection: content, usage };
  } catch (error) {
    logger.error('Failed to generate reflection framework', { error });
    throw error;
  }
}
