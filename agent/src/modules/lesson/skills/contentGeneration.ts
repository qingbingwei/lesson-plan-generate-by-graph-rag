import { getDeepSeekClient } from '../../../infrastructure/clients/deepseek';
import logger from '../../../shared/utils/logger';
import type { Skill } from './index';
import type {
  LessonSection,
  LessonObjectives,
  KnowledgeContext,
  GenerateLessonRequest,
  TokenUsage,
} from '../../../shared/types';

/**
 * 教学内容生成 Skill 定义
 */
export const contentGenerationSkill: Skill = {
  name: 'content_generation',
  description: '生成完整的教学环节设计，包括教师活动、学生活动、教学内容和设计意图',
  content: `
# 教学内容生成技能

## 功能说明
根据教学目标和知识点，生成完整的教学过程设计。

## 教学环节设计原则
1. **导入环节**: 激发兴趣，建立联系，明确目标
2. **新授环节**: 循序渐进，突出重点，突破难点
3. **练习环节**: 及时反馈，巩固知识，形成技能
4. **总结环节**: 梳理知识，归纳方法，拓展延伸

## 每个环节包含
- title: 环节名称
- duration: 时长（分钟）
- teacherActivity: 教师活动描述
- studentActivity: 学生活动描述
- content: 教学内容
- designIntent: 设计意图

## 注意事项
- 各环节时间分配要合理
- 活动设计要可操作性强
- 注重学生的主体地位
- 体现师生互动

## 附加功能
- 生成重点难点
- 推荐教学方法
- 设计教学资源
- 设计课后作业
`,
};

/**
 * 构建系统提示词
 */
function buildSystemPrompt(): string {
  return `你是一位经验丰富的教学设计专家，擅长设计科学合理的教学过程。

教学环节设计原则：
1. 导入环节：激发兴趣，建立联系，明确目标
2. 新授环节：循序渐进，突出重点，突破难点
3. 练习环节：及时反馈，巩固知识，形成技能
4. 总结环节：梳理知识，归纳方法，拓展延伸

每个环节都要包含：
- 明确的教师活动
- 对应的学生活动
- 具体的教学内容
- 清晰的设计意图

注意事项：
- 各环节时间分配要合理
- 活动设计要可操作性强
- 注重学生的主体地位
- 体现师生互动`;
}

/**
 * 构建教学环节提示词
 */
function buildSectionsPrompt(
  request: GenerateLessonRequest,
  objectives: LessonObjectives,
  keyPoints: string[],
  difficultPoints: string[],
  knowledgeContext: KnowledgeContext[]
): string {
  const contextText = knowledgeContext
    .map((c) => `【${c.name}】\n${c.content?.slice(0, 300) || ''}`)
    .join('\n\n');

  return `请为以下课程设计完整的教学过程：

基本信息：
- 学科：${request.subject}
- 年级：${request.grade}
- 课题：${request.topic}
- 课时：${request.duration}分钟
${request.style ? `- 教学风格：${request.style}` : ''}

教学目标：
- 知识与技能：${objectives.knowledge}
- 过程与方法：${objectives.process}
- 情感态度价值观：${objectives.emotion}

教学重点：
${keyPoints.map((p) => `- ${p}`).join('\n')}

教学难点：
${difficultPoints.map((p) => `- ${p}`).join('\n')}

相关知识：
${contextText}

${request.requirements ? `特殊要求：${request.requirements}` : ''}

请设计5-7个教学环节，包括：导入、新授、练习、总结等。
每个环节需要详细描述教师活动、学生活动、教学内容和设计意图。
确保各环节时间之和等于${request.duration}分钟。`;
}

/**
 * 调整环节时间分配
 */
function adjustSectionDurations(
  sections: LessonSection[],
  totalDuration: number
): LessonSection[] {
  const currentTotal = sections.reduce((sum, s) => sum + (s.duration || 0), 0);

  if (currentTotal === totalDuration) {
    return sections;
  }

  // 按比例调整
  const ratio = totalDuration / currentTotal;

  return sections.map((section, index) => {
    if (index === sections.length - 1) {
      // 最后一个环节用剩余时间
      const usedTime = sections
        .slice(0, -1)
        .reduce((sum, s) => sum + Math.round((s.duration || 0) * ratio), 0);
      return {
        ...section,
        duration: totalDuration - usedTime,
      };
    }

    return {
      ...section,
      duration: Math.round((section.duration || 0) * ratio),
    };
  });
}

/**
 * 生成教学环节
 */
export async function generateSections(
  request: GenerateLessonRequest,
  objectives: LessonObjectives,
  keyPoints: string[],
  difficultPoints: string[],
  knowledgeContext: KnowledgeContext[]
): Promise<{ sections: LessonSection[]; usage: TokenUsage }> {
  const startTime = Date.now();
  logger.info('ContentGeneration generating sections', {
    subject: request.subject,
    topic: request.topic,
    duration: request.duration,
  });

  try {
    const client = getDeepSeekClient();
    const systemPrompt = buildSystemPrompt();
    const userPrompt = buildSectionsPrompt(
      request,
      objectives,
      keyPoints,
      difficultPoints,
      knowledgeContext
    );

    const schema = `{
      "sections": [
        {
          "title": "环节名称",
          "duration": 5,
          "teacherActivity": "教师活动描述",
          "studentActivity": "学生活动描述",
          "content": "教学内容",
          "designIntent": "设计意图"
        }
      ]
    }`;

    const { data, usage } = await client.structuredChat<{ sections: LessonSection[] }>(
      [
        { role: 'system', content: systemPrompt },
        { role: 'user', content: userPrompt },
      ],
      schema,
      { temperature: 0.7, maxTokens: 4096 }
    );

    // 验证和调整时间分配
    const adjustedSections = adjustSectionDurations(data.sections, request.duration);

    logger.info('ContentGeneration sections completed', {
      duration: Date.now() - startTime,
      sectionCount: adjustedSections.length,
      usage,
    });

    return { sections: adjustedSections, usage };
  } catch (error) {
    logger.error('ContentGeneration failed to generate sections', { error });
    throw error;
  }
}

/**
 * 生成重点难点
 */
export async function generateKeyDifficultPoints(
  request: GenerateLessonRequest,
  objectives: LessonObjectives,
  knowledgeContext: KnowledgeContext[]
): Promise<{ keyPoints: string[]; difficultPoints: string[]; usage: TokenUsage }> {
  logger.info('ContentGeneration generating key/difficult points');

  try {
    const client = getDeepSeekClient();

    const prompt = `根据以下信息，确定本节课的教学重点和难点：

学科：${request.subject}
年级：${request.grade}
课题：${request.topic}

教学目标：
- 知识与技能：${objectives.knowledge}
- 过程与方法：${objectives.process}
- 情感态度价值观：${objectives.emotion}

相关知识点：
${knowledgeContext.map((c) => `- ${c.name}: ${c.content?.slice(0, 200) || ''}`).join('\n')}

请分析并确定：
1. 教学重点：学生必须掌握的核心内容
2. 教学难点：学生较难理解或容易出错的内容`;

    const schema = `{
      "keyPoints": ["重点1", "重点2"],
      "difficultPoints": ["难点1", "难点2"]
    }`;

    const { data, usage } = await client.structuredChat<{
      keyPoints: string[];
      difficultPoints: string[];
    }>(
      [
        {
          role: 'system',
          content: '你是一位经验丰富的教学设计专家，擅长分析教学内容的重点和难点。',
        },
        { role: 'user', content: prompt },
      ],
      schema,
      { temperature: 0.5 }
    );

    return { keyPoints: data.keyPoints, difficultPoints: data.difficultPoints, usage };
  } catch (error) {
    logger.error('Failed to generate key/difficult points', { error });
    throw error;
  }
}

/**
 * 生成教学方法
 */
export async function generateTeachingMethods(
  request: GenerateLessonRequest,
  objectives: LessonObjectives
): Promise<{ methods: string[]; usage: TokenUsage }> {
  logger.info('ContentGeneration generating teaching methods');

  try {
    const client = getDeepSeekClient();

    const prompt = `根据以下信息，推荐适合的教学方法：

学科：${request.subject}
年级：${request.grade}
课题：${request.topic}
课时：${request.duration}分钟
${request.style ? `教学风格：${request.style}` : ''}

教学目标：
- 知识与技能：${objectives.knowledge}
- 过程与方法：${objectives.process}
- 情感态度价值观：${objectives.emotion}

请推荐3-5种适合本节课的教学方法，并简要说明使用场景。`;

    const schema = `{
      "methods": ["教学方法1（使用场景说明）", "教学方法2（使用场景说明）"]
    }`;

    const { data, usage } = await client.structuredChat<{ methods: string[] }>(
      [
        { role: 'system', content: '你是一位教学法专家，熟悉各种教学方法及其适用场景。' },
        { role: 'user', content: prompt },
      ],
      schema,
      { temperature: 0.6 }
    );

    return { methods: data.methods, usage };
  } catch (error) {
    logger.error('Failed to generate teaching methods', { error });
    throw error;
  }
}

/**
 * 生成教学资源和材料
 */
export async function generateMaterials(
  request: GenerateLessonRequest,
  sections: LessonSection[]
): Promise<{ materials: string[]; usage: TokenUsage }> {
  logger.info('ContentGeneration generating materials');

  try {
    const client = getDeepSeekClient();

    const sectionsSummary = sections
      .map((s) => `${s.title}: ${s.content?.slice(0, 100) || ''}`)
      .join('\n');

    const prompt = `根据以下教学设计，列出需要准备的教学资源和材料：

学科：${request.subject}
年级：${request.grade}
课题：${request.topic}

教学环节：
${sectionsSummary}

请列出：
1. 教具（如：实物、模型等）
2. 学具（学生需要准备的材料）
3. 多媒体资源（如：PPT、视频、音频等）
4. 其他资源（如：导学案、练习题等）`;

    const schema = `{
      "materials": ["资源1", "资源2"]
    }`;

    const { data, usage } = await client.structuredChat<{ materials: string[] }>(
      [
        { role: 'system', content: '你是一位经验丰富的教师，擅长准备教学资源和材料。' },
        { role: 'user', content: prompt },
      ],
      schema,
      { temperature: 0.5 }
    );

    return { materials: data.materials, usage };
  } catch (error) {
    logger.error('Failed to generate materials', { error });
    throw error;
  }
}

/**
 * 生成作业设计
 */
export async function generateHomework(
  request: GenerateLessonRequest,
  objectives: LessonObjectives,
  keyPoints: string[]
): Promise<{ homework: string; usage: TokenUsage }> {
  logger.info('ContentGeneration generating homework');

  try {
    const client = getDeepSeekClient();

    const prompt = `根据以下教学内容，设计课后作业：

学科：${request.subject}
年级：${request.grade}
课题：${request.topic}

教学目标：
${objectives.knowledge}

教学重点：
${keyPoints.join('\n')}

请设计：
1. 基础作业（巩固基础知识）
2. 提高作业（能力提升，可选做）
3. 实践作业（联系生活实际，可选做）

要求：
- 作业量适中，符合"双减"政策
- 分层设计，照顾不同水平学生
- 形式多样，不局限于书面作业`;

    const { content, usage } = await client.chat(
      [
        { role: 'system', content: '你是一位经验丰富的教师，擅长设计科学合理的课后作业。' },
        { role: 'user', content: prompt },
      ],
      { temperature: 0.6 }
    );

    return { homework: content, usage };
  } catch (error) {
    logger.error('Failed to generate homework', { error });
    throw error;
  }
}
