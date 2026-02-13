import { AsyncLocalStorage } from 'node:async_hooks';
import { v4 as uuidv4 } from 'uuid';
import { z } from 'zod';
import { ChatOpenAI } from '@langchain/openai';
import { MemorySaver } from '@langchain/langgraph-checkpoint';
import {
  context,
  createAgent,
  createMiddleware,
  tool,
  type ToolRuntime,
} from 'langchain';

import config from '../../../config';
import logger from '../../../shared/utils/logger';
import { getRequestApiKeys } from '../../../shared/context/requestApiKeys';
import { mergeUsage } from '../../../shared/utils/tokenUsage';
import type {
  GenerateLessonRequest,
  GenerateLessonResponse,
  GeneratedLesson,
  TokenUsage,
  WorkflowState,
} from '../../../shared/types';
import {
  contentGenerationSkill,
  evaluationDesignSkill,
  generateEvaluation,
  generateHomework,
  generateKeyDifficultPoints,
  generateMaterials,
  generateObjectives,
  generateSections,
  generateTeachingMethods,
  knowledgeRetrievalSkill,
  objectiveGenerationSkill,
  retrieveKnowledge,
  validateObjectives,
} from '../skills';

type LessonProgressEvent = {
  node: string;
  state: Partial<WorkflowState>;
};

type LessonToolPayload = {
  lesson: GeneratedLesson;
  usage: TokenUsage;
};

const progressStorage = new AsyncLocalStorage<(event: LessonProgressEvent) => void>();

const LESSON_RESULT_PREFIX = 'LESSON_RESULT::';
const LOAD_SKILL_TOOL_NAME = 'load_skill';
const GENERATE_TOOL_NAME = 'generate_lesson_with_skills';

const defaultUsage: TokenUsage = {
  promptTokens: 0,
  completionTokens: 0,
  totalTokens: 0,
};

const SkillSchema = z.object({
  name: z.string(),
  description: z.string(),
  content: z.string(),
});

type AgentSkill = z.infer<typeof SkillSchema>;

const SKILLS: AgentSkill[] = [
  knowledgeRetrievalSkill,
  objectiveGenerationSkill,
  contentGenerationSkill,
  evaluationDesignSkill,
].map(skill =>
  SkillSchema.parse({
    name: skill.name,
    description: skill.description,
    content: context`${skill.content}`,
  })
);

const GenerateLessonToolInputSchema = z.object({
  subject: z.string().min(1),
  grade: z.string().min(1),
  topic: z.string().min(1),
  duration: z.number().int().positive(),
  style: z.string().optional(),
  requirements: z.string().optional(),
  userId: z.string().optional(),
  context: z
    .array(
      z.object({
        id: z.string(),
        name: z.string(),
        content: z.string(),
        relevanceScore: z.number().optional(),
        source: z.string().optional(),
      })
    )
    .optional(),
});

function emitProgress(node: string, state: Partial<WorkflowState>) {
  const emitter = progressStorage.getStore();
  if (!emitter) {
    return;
  }
  emitter({ node, state });
}

function createRuntimeModel() {
  const { generationApiKey } = getRequestApiKeys();
  const apiKey = (generationApiKey || config.deepseek.apiKey || '').trim();

  if (!apiKey) {
    throw new Error('缺少生成模型 API Key，请先配置 DEEPSEEK_API_KEY 或请求头 x-generation-api-key');
  }

  return new ChatOpenAI({
    model: config.deepseek.model,
    temperature: 0,
    maxTokens: config.deepseek.maxTokens,
    apiKey,
    configuration: {
      baseURL: config.deepseek.baseUrl,
    },
  });
}

function buildSkillCatalogPrompt() {
  const skillList = SKILLS.map(skill => `- **${skill.name}**: ${skill.description}`).join('\n');

  return context`
    ## Available Skills

    ${skillList}

    当你需要技能细节时，请调用工具 ${LOAD_SKILL_TOOL_NAME}。
    最终必须调用工具 ${GENERATE_TOOL_NAME} 一次来生成结构化教案。
  `;
}

function encodeLessonToolPayload(payload: LessonToolPayload): string {
  const json = JSON.stringify(payload);
  const encoded = Buffer.from(json, 'utf8').toString('base64');
  return `${LESSON_RESULT_PREFIX}${encoded}`;
}

function decodeLessonToolPayload(text: string): LessonToolPayload | null {
  const match = text.match(/LESSON_RESULT::([A-Za-z0-9+/=]+)/);
  if (!match || !match[1]) {
    return null;
  }

  try {
    const decoded = Buffer.from(match[1], 'base64').toString('utf8');
    const parsed = JSON.parse(decoded) as LessonToolPayload;

    if (!parsed?.lesson || !parsed?.usage) {
      return null;
    }

    return parsed;
  } catch (error) {
    logger.warn('Failed to decode lesson tool payload', { error });
    return null;
  }
}

function stringifyMessageContent(content: unknown): string {
  if (typeof content === 'string') {
    return content;
  }

  if (Array.isArray(content)) {
    return content
      .map(item => {
        if (typeof item === 'string') {
          return item;
        }
        if (item && typeof item === 'object' && 'text' in item) {
          const text = (item as { text?: unknown }).text;
          return typeof text === 'string' ? text : '';
        }
        return '';
      })
      .join('\n');
  }

  if (content == null) {
    return '';
  }

  return String(content);
}

function extractLessonToolPayloadFromMessages(messages: unknown[]): LessonToolPayload | null {
  for (let index = messages.length - 1; index >= 0; index--) {
    const message = messages[index] as { content?: unknown } | undefined;
    if (!message) {
      continue;
    }

    const contentText = stringifyMessageContent(message.content);
    const payload = decodeLessonToolPayload(contentText);
    if (payload) {
      return payload;
    }
  }

  return null;
}

function generateTitle(request: GenerateLessonRequest): string {
  return `${request.subject} ${request.grade} 《${request.topic}》教学设计`;
}

function generateReflectionTemplate(request: GenerateLessonRequest): string {
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
课题：${request.topic}
授课时间：______
授课班级：______
反思时间：______`;
}

async function generateLessonWithSkills(request: GenerateLessonRequest): Promise<LessonToolPayload> {
  const startTime = Date.now();

  emitProgress('inputAnalysis', {
    input: request,
    startTime,
  });

  let knowledgeContext = request.context || [];
  if (knowledgeContext.length === 0) {
    try {
      knowledgeContext = await retrieveKnowledge(request);
    } catch (error) {
      logger.warn('Knowledge retrieval failed, continue with empty context', { error });
      knowledgeContext = [];
    }
  }

  emitProgress('knowledgeQuery', {
    knowledgeContext,
  });

  const { objectives, usage: objectivesUsage } = await generateObjectives(request, knowledgeContext);
  if (!validateObjectives(objectives)) {
    throw new Error('生成的教学目标不完整');
  }

  const { keyPoints, difficultPoints, usage: keyDifficultUsage } = await generateKeyDifficultPoints(
    request,
    objectives,
    knowledgeContext
  );

  const { methods: teachingMethods, usage: methodsUsage } = await generateTeachingMethods(
    request,
    objectives
  );

  const objectiveStageUsage = mergeUsage(objectivesUsage, keyDifficultUsage, methodsUsage);

  emitProgress('objectiveDesign', {
    lessonObjectives: objectives,
    keyPoints,
    difficultPoints,
    teachingMethods,
    usage: objectiveStageUsage,
  });

  const { sections, usage: sectionsUsage } = await generateSections(
    request,
    objectives,
    keyPoints,
    difficultPoints,
    knowledgeContext
  );

  const contentStageUsage = mergeUsage(objectiveStageUsage, sectionsUsage);

  emitProgress('contentDesign', {
    sections,
    usage: contentStageUsage,
  });

  const [{ materials, usage: materialsUsage }, { homework, usage: homeworkUsage }, { evaluation, usage: evaluationUsage }] =
    await Promise.all([
      generateMaterials(request, sections),
      generateHomework(request, objectives, keyPoints),
      generateEvaluation(request, objectives, keyPoints, sections),
    ]);

  const activityStageUsage = mergeUsage(contentStageUsage, materialsUsage, homeworkUsage, evaluationUsage);

  emitProgress('activityDesign', {
    materials,
    homework,
    evaluation,
    usage: activityStageUsage,
  });

  const output: GeneratedLesson = {
    title: generateTitle(request),
    objectives,
    keyPoints,
    difficultPoints,
    teachingMethods,
    content: {
      sections,
      materials,
      homework,
    },
    evaluation,
    reflection: generateReflectionTemplate(request),
  };

  const finalUsage = mergeUsage(activityStageUsage);
  emitProgress('outputFormat', {
    output,
    usage: finalUsage,
    endTime: Date.now(),
  });

  logger.info('Lesson generated by skills tool', {
    duration: Date.now() - startTime,
    usage: finalUsage,
  });

  return {
    lesson: output,
    usage: finalUsage,
  };
}

const loadSkillTool = tool(
  async ({ skillName }: { skillName: string }) => {
    const target = SKILLS.find(skill => skill.name === skillName);
    if (!target) {
      return `Skill '${skillName}' not found. Available skills: ${SKILLS.map(skill => skill.name).join(', ')}`;
    }

    return `Loaded skill: ${target.name}\n\n${target.content}`;
  },
  {
    name: LOAD_SKILL_TOOL_NAME,
    description: context`
      加载指定 skill 的完整内容。
      当你需要技能详细规则、输入输出和边界条件时，调用该工具。
    `,
    schema: z.object({
      skillName: z.string().describe('要加载的 skill 名称'),
    }),
  }
);

const generateLessonBySkillsTool = tool(
  async (input: z.infer<typeof GenerateLessonToolInputSchema>, _runtime?: ToolRuntime) => {
    const payload = await generateLessonWithSkills(input);
    return encodeLessonToolPayload(payload);
  },
  {
    name: GENERATE_TOOL_NAME,
    description: context`
      基于技能链路生成完整教案，并返回可解析的结构化结果。
      当用户需要教案结果时，必须调用该工具。
    `,
    schema: GenerateLessonToolInputSchema,
  }
);

const skillMiddleware = createMiddleware({
  name: 'skillMiddleware',
  tools: [loadSkillTool],
  wrapModelCall: async (request, handler) => {
    const skillsAddendum = buildSkillCatalogPrompt();
    const baseSystemPrompt =
      typeof request.systemPrompt === 'string' ? request.systemPrompt : '';
    const mergedSystemPrompt = baseSystemPrompt.includes('## Available Skills')
      ? baseSystemPrompt
      : baseSystemPrompt
      ? `${baseSystemPrompt}\n\n${skillsAddendum}`
      : skillsAddendum;

    return handler({
      ...request,
      systemPrompt: mergedSystemPrompt,
    });
  },
});

function createLessonAgent() {
  const model = createRuntimeModel();

  return createAgent({
    model,
    tools: [generateLessonBySkillsTool],
    middleware: [skillMiddleware],
    checkpointer: new MemorySaver(),
    systemPrompt: context`
      你是一个教案生成智能体。你需要遵守以下规则：

1) 先理解用户请求。
2) 当需要技能细节时，调用 ${LOAD_SKILL_TOOL_NAME}。
3) 生成教案时，必须调用 ${GENERATE_TOOL_NAME}。
4) 不要编造教案内容，必须基于工具结果返回。

      返回最终答案前，确保包含结构化教案结果。
    `,
    name: 'lesson_plan_agent',
  });
}

function buildAgentUserPrompt(request: GenerateLessonRequest): string {
  return [
    '请生成一份完整教案。',
    `请务必调用工具 ${GENERATE_TOOL_NAME}，参数如下：`,
    JSON.stringify(request, null, 2),
  ].join('\n');
}

function buildAgentInvocationMetadata(request: GenerateLessonRequest): Record<string, unknown> {
  return {
    feature: 'lesson_generation',
    subject: request.subject,
    grade: request.grade,
    topic: request.topic,
    userId: request.userId || 'anonymous',
  };
}

async function runAgentInvocation(request: GenerateLessonRequest): Promise<LessonToolPayload | null> {
  const agent = createLessonAgent();
  const threadId = `lesson-agent-${request.userId || 'anonymous'}-${uuidv4()}`;

  const result = await agent.invoke(
    {
      messages: [
        {
          role: 'user',
          content: buildAgentUserPrompt(request),
        },
      ],
    },
    {
      configurable: {
        thread_id: threadId,
      },
      runName: 'lesson_generation_agent',
      tags: ['lesson-plan', 'agent', 'lesson-generation'],
      metadata: buildAgentInvocationMetadata(request),
    }
  );

  const messages = (result as { messages?: unknown[] }).messages || [];
  return extractLessonToolPayloadFromMessages(messages);
}

export async function runLessonAgent(request: GenerateLessonRequest): Promise<GenerateLessonResponse> {
  try {
    const payload = await runAgentInvocation(request);
    if (payload) {
      return {
        success: true,
        data: payload.lesson,
        usage: payload.usage,
      };
    }

    logger.warn('Agent did not return tool payload, fallback to direct skill execution');
    const fallbackPayload = await generateLessonWithSkills(request);
    return {
      success: true,
      data: fallbackPayload.lesson,
      usage: fallbackPayload.usage,
    };
  } catch (error) {
    logger.error('Lesson agent execution failed', { error });
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Agent 执行失败',
      usage: defaultUsage,
    };
  }
}

export async function* streamLessonAgent(
  request: GenerateLessonRequest
): AsyncGenerator<{ node: string; state: Partial<WorkflowState> }> {
  const queue: LessonProgressEvent[] = [];
  let wakeup: (() => void) | null = null;
  let completed = false;
  let streamError: Error | null = null;
  let hasOutputEvent = false;

  const emit = (event: LessonProgressEvent) => {
    if (event.node === 'outputFormat' && event.state.output) {
      hasOutputEvent = true;
    }
    queue.push(event);
    if (wakeup) {
      wakeup();
      wakeup = null;
    }
  };

  const runner = progressStorage.run(emit, async () => {
    try {
      const result = await runLessonAgent(request);

      if (!result.success) {
        emit({
          node: 'outputFormat',
          state: {
            error: result.error || '生成失败',
            endTime: Date.now(),
          },
        });
      } else if (!hasOutputEvent && result.data) {
        emit({
          node: 'outputFormat',
          state: {
            output: result.data,
            usage: result.usage,
            endTime: Date.now(),
          },
        });
      }
    } catch (error) {
      streamError = error instanceof Error ? error : new Error(String(error));
      emit({
        node: 'outputFormat',
        state: {
          error: streamError.message,
          endTime: Date.now(),
        },
      });
    } finally {
      completed = true;
      if (wakeup) {
        wakeup();
        wakeup = null;
      }
    }
  });

  while (!completed || queue.length > 0) {
    if (queue.length === 0) {
      await new Promise<void>(resolve => {
        wakeup = resolve;
      });
      continue;
    }

    const next = queue.shift();
    if (next) {
      yield next;
    }
  }

  await runner;

  if (streamError) {
    throw streamError;
  }
}
