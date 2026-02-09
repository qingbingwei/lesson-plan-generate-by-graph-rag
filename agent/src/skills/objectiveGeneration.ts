import { getClient } from '../clients';
import logger from '../utils/logger';
import type { Skill } from './index';
import type { LessonObjectives, KnowledgeContext, GenerateLessonRequest, TokenUsage } from '../types';

/**
 * 教学目标生成 Skill 定义
 */
export const objectiveGenerationSkill: Skill = {
  name: 'objective_generation',
  description: '根据课程信息和知识点生成三维教学目标（知识与技能、过程与方法、情感态度价值观）',
  content: `
# 教学目标生成技能

## 功能说明
根据课程标准和教学内容设计科学合理的三维教学目标。

## 三维目标体系
1. **知识与技能目标**: 明确、具体、可测量，描述学生应掌握的核心知识和基本技能
2. **过程与方法目标**: 关注学习过程，培养学生的学习能力和思维方法
3. **情感态度价值观目标**: 关注学生的情感体验、学习态度和价值观培养

## 设计原则
- 目标要符合学生年龄特点和认知水平
- 使用行为动词（如：理解、掌握、运用、分析、评价等）
- 目标要具体可操作，便于教学评价
- 三个维度要相互关联、有机统一

## 输入参数
- subject: 学科
- grade: 年级
- topic: 课题
- duration: 课时（分钟）
- style: 教学风格（可选）
- requirements: 特殊要求（可选）
- knowledgeContext: 相关知识点列表

## 输出结构
{
  "knowledge": "知识与技能目标",
  "process": "过程与方法目标",
  "emotion": "情感态度价值观目标"
}
`,
};

/**
 * 构建系统提示词
 */
function buildSystemPrompt(): string {
  return `你是一位经验丰富的教学设计专家，擅长根据课程标准和教学内容设计科学合理的三维教学目标。

教学目标设计原则：
1. 知识与技能目标：明确、具体、可测量，描述学生应掌握的核心知识和基本技能
2. 过程与方法目标：关注学习过程，培养学生的学习能力和思维方法
3. 情感态度价值观目标：关注学生的情感体验、学习态度和价值观培养

要求：
- 目标要符合学生年龄特点和认知水平
- 使用行为动词（如：理解、掌握、运用、分析、评价等）
- 目标要具体可操作，便于教学评价
- 三个维度要相互关联、有机统一`;
}

/**
 * 构建用户提示词
 */
function buildUserPrompt(
  request: GenerateLessonRequest,
  knowledgeContext: KnowledgeContext[]
): string {
  const contextText = knowledgeContext
    .map((c) => `【${c.name}】\n${c.content}`)
    .join('\n\n');

  return `请为以下课程设计三维教学目标：

学科：${request.subject}
年级：${request.grade}
课题：${request.topic}
课时：${request.duration}分钟
${request.style ? `教学风格：${request.style}` : ''}
${request.requirements ? `特殊要求：${request.requirements}` : ''}

相关知识点：
${contextText}

请根据以上信息，设计符合课程标准的三维教学目标。`;
}

/**
 * 生成教学目标
 */
export async function generateObjectives(
  request: GenerateLessonRequest,
  knowledgeContext: KnowledgeContext[]
): Promise<{ objectives: LessonObjectives; usage: TokenUsage }> {
  const startTime = Date.now();
  logger.info('ObjectiveGeneration executing', {
    subject: request.subject,
    grade: request.grade,
    topic: request.topic,
  });

  try {
    const client = getClient();
    const systemPrompt = buildSystemPrompt();
    const userPrompt = buildUserPrompt(request, knowledgeContext);

    const schema = `{
      "knowledge": "知识与技能目标（描述学生需要掌握的知识和技能）",
      "process": "过程与方法目标（描述学习过程和方法）",
      "emotion": "情感态度价值观目标（描述情感态度和价值观培养）"
    }`;

    const { data, usage } = await client.structuredChat<LessonObjectives>(
      [
        { role: 'system', content: systemPrompt },
        { role: 'user', content: userPrompt },
      ],
      schema,
      { temperature: 0.5 }
    );

    logger.info('ObjectiveGeneration completed', {
      duration: Date.now() - startTime,
      usage,
    });

    return { objectives: data, usage };
  } catch (error) {
    logger.error('ObjectiveGeneration failed', { error, request });
    throw error;
  }
}

/**
 * 验证教学目标
 */
export function validateObjectives(objectives: LessonObjectives): boolean {
  // 检查必填字段
  if (!objectives.knowledge || !objectives.process || !objectives.emotion) {
    return false;
  }

  // 检查内容长度
  if (
    objectives.knowledge.length < 10 ||
    objectives.process.length < 10 ||
    objectives.emotion.length < 10
  ) {
    return false;
  }

  return true;
}
