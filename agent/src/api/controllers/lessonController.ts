import { Request, Response } from 'express';
import { Client as LangSmithClient } from 'langsmith';
import type { Run as LangSmithRun } from 'langsmith/schemas';
import config from '../../config';
import logger from '../../shared/utils/logger';
import { runLessonAgent, streamLessonAgent } from '../../modules/lesson/agent/lessonAgent';
import { runBuildGraphWorkflow, BuildGraphRequest } from '../../modules/knowledge/workflows/buildGraphWorkflow';
import { getDeepSeekClient } from '../../infrastructure/clients/deepseek';
import {
  withRequestApiKeys,
  GENERATION_API_KEY_HEADER,
  EMBEDDING_API_KEY_HEADER,
  type RequestApiKeyOverrides,
} from '../../shared/context/requestApiKeys';
import type { DeepSeekMessage, GenerateLessonRequest, RegenerateSectionRequest } from '../../shared/types';

/**
 * 健康检查
 */
export function healthCheck(_req: Request, res: Response) {
  res.json({
    status: 'ok',
    timestamp: new Date().toISOString(),
    version: '1.0.0',
  });
}

function resolveApiKeyOverrides(req: Request): RequestApiKeyOverrides {
  const generationApiKey = (req.header(GENERATION_API_KEY_HEADER) || '').trim();
  const embeddingApiKey = (req.header(EMBEDDING_API_KEY_HEADER) || '').trim();

  return {
    generationApiKey: generationApiKey || undefined,
    embeddingApiKey: embeddingApiKey || undefined,
  };
}

type LangSmithHistoryItem = {
  id: string;
  status: string;
  prompt: string;
  token_count: number;
  prompt_tokens: number;
  completion_tokens: number;
  duration_ms: number;
  error_msg?: string;
  created_at: string;
  completed_at?: string;
};

type LangSmithUsageStats = {
  total_count: number;
  completed_count: number;
  failed_count: number;
  total_tokens: number;
  avg_duration_ms: number;
  this_month_generations: number;
  total_lessons: number;
};

type LangSmithUsageResponse = {
  success: true;
  source: 'langsmith';
  project: string;
  stats: LangSmithUsageStats;
  history: {
    items: LangSmithHistoryItem[];
    total: number;
    page: number;
    pageSize: number;
    totalPages: number;
  };
};


type AssistantHistoryMessage = {
  role: 'user' | 'assistant';
  content: string;
};

type AssistantChatRequest = {
  question?: string;
  history?: AssistantHistoryMessage[];
  userId?: string;
};

const ASSISTANT_SYSTEM_PROMPT = `你是“智能教案生成系统”的 AI 助手，必须围绕本项目提供帮助。

你的能力边界：
1. 指导用户完成项目使用流程（登录、知识库上传、教案生成、编辑发布）。
2. 提供可直接复用的教案模板（支持按学科、年级、课时组织）。
3. 解答系统功能问题（知识图谱、历史版本、Token 与 API Key 配置）。
4. 提供常见故障排查建议（服务启动失败、接口报错、配置缺失等）。

回答要求：
- 使用中文，结构清晰，步骤化表达。
- 若用户请求“模板”，优先输出 Markdown 模板，包含：教学目标、重点难点、教学过程、作业评价等。
- 若信息不足，先提出 1-3 个澄清问题再给建议。
- 禁止编造未确认的项目能力；不确定时明确说明并给出可执行替代方案。`;

const MAX_ASSISTANT_HISTORY = 12;
const DEFAULT_ASSISTANT_SUGGESTIONS = [
  '如何从零开始使用这个系统？',
  '给我一个 45 分钟数学教案模板',
  '知识图谱页面怎么配合教案生成？',
  '启动失败时应该优先看哪些日志？',
];

const DEFAULT_LANGSMITH_PAGE = 1;
const DEFAULT_LANGSMITH_PAGE_SIZE = 10;
const MAX_LANGSMITH_PAGE_SIZE = 100;
const MAX_LANGSMITH_FETCH_LIMIT = 1000;
const LANGSMITH_USAGE_CACHE_TTL_MS = 60 * 1000;
const MAX_LANGSMITH_CACHE_ENTRIES = 200;


function normalizeAssistantHistory(history: unknown): DeepSeekMessage[] {
  if (!Array.isArray(history)) {
    return [];
  }

  return history
    .map((item): DeepSeekMessage | null => {
      if (!item || typeof item !== 'object') {
        return null;
      }

      const roleRaw = (item as { role?: unknown }).role;
      const contentRaw = (item as { content?: unknown }).content;

      if (typeof contentRaw !== 'string') {
        return null;
      }

      const content = contentRaw.trim();
      if (!content) {
        return null;
      }

      if (roleRaw !== 'user' && roleRaw !== 'assistant') {
        return null;
      }

      return {
        role: roleRaw,
        content,
      };
    })
    .filter((message): message is DeepSeekMessage => message !== null);
}

function buildAssistantSuggestions(question: string): string[] {
  const normalized = question.toLowerCase();

  if (normalized.includes('模板') || normalized.includes('教案')) {
    return [
      '生成一份高中数学 45 分钟教案模板',
      '生成一份初中英语阅读课教案模板',
      '给模板增加分层作业和课堂评价',
    ];
  }

  if (normalized.includes('启动') || normalized.includes('报错') || normalized.includes('日志')) {
    return [
      '后端启动失败时怎么定位问题？',
      '如何检查 Agent 与后端连通性？',
      '服务启动后如何做快速自检？',
    ];
  }

  return DEFAULT_ASSISTANT_SUGGESTIONS.slice(0, 3);
}

type LangSmithUsageCacheEntry = {
  expiresAt: number;
  payload: LangSmithUsageResponse;
};

const langSmithUsageCache = new Map<string, LangSmithUsageCacheEntry>();

function buildLangSmithCacheKey(userId: string, page: number, pageSize: number): string {
  return `${config.langsmith.project}:${userId}:${page}:${pageSize}`;
}

function getLangSmithUsageFromCache(cacheKey: string): LangSmithUsageResponse | null {
  const entry = langSmithUsageCache.get(cacheKey);
  if (!entry) {
    return null;
  }

  if (entry.expiresAt <= Date.now()) {
    langSmithUsageCache.delete(cacheKey);
    return null;
  }

  return entry.payload;
}

function setLangSmithUsageCache(cacheKey: string, payload: LangSmithUsageResponse): void {
  const now = Date.now();

  for (const [entryKey, entry] of langSmithUsageCache.entries()) {
    if (entry.expiresAt <= now) {
      langSmithUsageCache.delete(entryKey);
    }
  }

  while (langSmithUsageCache.size >= MAX_LANGSMITH_CACHE_ENTRIES) {
    const oldestKey = langSmithUsageCache.keys().next().value;
    if (!oldestKey) {
      break;
    }

    langSmithUsageCache.delete(oldestKey);
  }

  langSmithUsageCache.set(cacheKey, {
    expiresAt: now + LANGSMITH_USAGE_CACHE_TTL_MS,
    payload,
  });
}

function parsePositiveInt(value: unknown, fallback: number, maxValue?: number): number {
  const parsed = Number.parseInt(String(value ?? ''), 10);
  if (!Number.isFinite(parsed) || parsed <= 0) {
    return fallback;
  }

  if (maxValue && parsed > maxValue) {
    return maxValue;
  }

  return parsed;
}

function asObject(value: unknown): Record<string, unknown> | null {
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    return null;
  }

  return value as Record<string, unknown>;
}

function getErrorStatus(error: unknown): number | null {
  const topLevel = asObject(error);

  const directStatus = topLevel?.status;
  if (typeof directStatus === 'number' && Number.isFinite(directStatus)) {
    return directStatus;
  }

  if (typeof directStatus === 'string') {
    const parsed = Number.parseInt(directStatus, 10);
    if (Number.isFinite(parsed)) {
      return parsed;
    }
  }

  const response = asObject(topLevel?.response);
  const responseStatus = response?.status;
  if (typeof responseStatus === 'number' && Number.isFinite(responseStatus)) {
    return responseStatus;
  }

  if (typeof responseStatus === 'string') {
    const parsed = Number.parseInt(responseStatus, 10);
    if (Number.isFinite(parsed)) {
      return parsed;
    }
  }

  return null;
}

function toTimestamp(value: string | number | undefined): number | null {
  if (value == null) {
    return null;
  }

  if (typeof value === 'number') {
    if (!Number.isFinite(value) || value <= 0) {
      return null;
    }

    if (value < 1e11) {
      return Math.round(value * 1000);
    }

    return Math.round(value);
  }

  const trimmed = value.trim();
  if (!trimmed) {
    return null;
  }

  const numeric = Number(trimmed);
  if (Number.isFinite(numeric) && numeric > 0) {
    return numeric < 1e11 ? Math.round(numeric * 1000) : Math.round(numeric);
  }

  const parsed = Date.parse(trimmed);
  if (!Number.isFinite(parsed) || parsed <= 0) {
    return null;
  }

  return parsed;
}

function toISOTime(value: string | number | undefined): string | undefined {
  const timestamp = toTimestamp(value);
  if (!timestamp) {
    return undefined;
  }

  return new Date(timestamp).toISOString();
}

function extractRunUserId(run: LangSmithRun): string {
  const extra = asObject(run.extra);
  const invocationParams = asObject(extra?.invocation_params);

  const metadataCandidates: Array<Record<string, unknown> | null> = [
    asObject(extra?.metadata),
    asObject(invocationParams?.metadata),
  ];

  for (const metadata of metadataCandidates) {
    const userId = metadata?.userId;
    if (typeof userId === 'string' && userId.trim()) {
      return userId.trim();
    }
  }

  const inputs = asObject(run.inputs);
  const directUserId = inputs?.userId;
  if (typeof directUserId === 'string' && directUserId.trim()) {
    return directUserId.trim();
  }

  const nestedInput = asObject(inputs?.input);
  const nestedUserId = nestedInput?.userId;
  if (typeof nestedUserId === 'string' && nestedUserId.trim()) {
    return nestedUserId.trim();
  }

  const messages = inputs?.messages;
  if (Array.isArray(messages)) {
    for (const message of messages) {
      if (typeof message === 'string') {
        const matched = message.match(/"userId"\s*:\s*"([^"]+)"/);
        if (matched?.[1]) {
          return matched[1].trim();
        }
      }

      const messageObj = asObject(message);
      const content = messageObj?.content;
      if (typeof content === 'string') {
        const matched = content.match(/"userId"\s*:\s*"([^"]+)"/);
        if (matched?.[1]) {
          return matched[1].trim();
        }
      }
    }
  }

  return '';
}

function extractTokenUsage(run: LangSmithRun): { promptTokens: number; completionTokens: number; totalTokens: number } {
  const promptTokens = Number(run.prompt_tokens || 0);
  const completionTokens = Number(run.completion_tokens || 0);
  const totalTokensRaw = Number(run.total_tokens || 0);
  const totalTokens = totalTokensRaw > 0 ? totalTokensRaw : promptTokens + completionTokens;

  return {
    promptTokens: Math.max(0, Math.round(promptTokens)),
    completionTokens: Math.max(0, Math.round(completionTokens)),
    totalTokens: Math.max(0, Math.round(totalTokens)),
  };
}

function buildHistoryItem(run: LangSmithRun): LangSmithHistoryItem {
  const startTime = toTimestamp(run.start_time);
  const endTime = toTimestamp(run.end_time);
  const durationMs = startTime && endTime && endTime >= startTime ? endTime - startTime : 0;
  const usage = extractTokenUsage(run);

  let status = 'running';
  if (run.error) {
    status = 'failed';
  } else if (endTime) {
    status = 'completed';
  }

  return {
    id: run.id,
    status,
    prompt: run.name || run.run_type || 'lesson_generation',
    token_count: usage.totalTokens,
    prompt_tokens: usage.promptTokens,
    completion_tokens: usage.completionTokens,
    duration_ms: durationMs,
    error_msg: run.error || undefined,
    created_at: toISOTime(run.start_time) || new Date(0).toISOString(),
    completed_at: toISOTime(run.end_time),
  };
}

function buildUsageStats(runs: LangSmithRun[]): LangSmithUsageStats {
  const monthStart = new Date();
  monthStart.setDate(1);
  monthStart.setHours(0, 0, 0, 0);

  let totalTokens = 0;
  let totalDurationMs = 0;
  let completedCount = 0;
  let failedCount = 0;
  let thisMonthCount = 0;

  for (const run of runs) {
    const usage = extractTokenUsage(run);
    totalTokens += usage.totalTokens;

    const startTime = toTimestamp(run.start_time);
    const endTime = toTimestamp(run.end_time);

    if (run.error) {
      failedCount += 1;
    } else if (endTime) {
      completedCount += 1;
    }

    if (startTime && startTime >= monthStart.getTime()) {
      thisMonthCount += 1;
    }

    if (startTime && endTime && endTime >= startTime) {
      totalDurationMs += endTime - startTime;
    }
  }

  return {
    total_count: runs.length,
    completed_count: completedCount,
    failed_count: failedCount,
    total_tokens: totalTokens,
    avg_duration_ms: runs.length > 0 ? totalDurationMs / runs.length : 0,
    this_month_generations: thisMonthCount,
    total_lessons: 0,
  };
}

/**
 * 生成教案
 */
export async function generateLesson(req: Request, res: Response) {
  try {
    const request = req.body as GenerateLessonRequest;

    if (!request.subject || !request.grade || !request.topic || !request.duration) {
      res.status(400).json({
        success: false,
        error: '缺少必要参数：subject, grade, topic, duration',
      });
      return;
    }

    logger.info('Generate lesson request', {
      subject: request.subject,
      grade: request.grade,
      topic: request.topic,
      duration: request.duration,
    });

    const apiKeyOverrides = resolveApiKeyOverrides(req);
    const result = await withRequestApiKeys(apiKeyOverrides, async () => runLessonAgent(request));
    res.json(result);
  } catch (error) {
    logger.error('Generate lesson error', { error });
    res.status(500).json({
      success: false,
      error: error instanceof Error ? error.message : 'Internal server error',
    });
  }
}

/**
 * 流式生成教案
 */
export async function streamGenerateLesson(req: Request, res: Response) {
  try {
    const request = req.body as GenerateLessonRequest;

    if (!request.subject || !request.grade || !request.topic || !request.duration) {
      res.status(400).json({
        success: false,
        error: '缺少必要参数：subject, grade, topic, duration',
      });
      return;
    }

    // 设置 SSE 响应头
    res.setHeader('Content-Type', 'text/event-stream');
    res.setHeader('Cache-Control', 'no-cache');
    res.setHeader('Connection', 'keep-alive');

    logger.info('Stream generate lesson request', {
      subject: request.subject,
      topic: request.topic,
    });

    const apiKeyOverrides = resolveApiKeyOverrides(req);

    await withRequestApiKeys(apiKeyOverrides, async () => {
      for await (const event of streamLessonAgent(request)) {
        res.write(`data: ${JSON.stringify(event)}\n\n`);
      }
    });

    res.write('data: [DONE]\n\n');
    res.end();
  } catch (error) {
    logger.error('Stream generate lesson error', { error });

    if (!res.headersSent) {
      res.status(500).json({
        success: false,
        error: error instanceof Error ? error.message : 'Internal server error',
      });
    } else {
      res.write(`data: ${JSON.stringify({ error: error instanceof Error ? error.message : 'Stream error' })}\n\n`);
      res.end();
    }
  }
}

/**
 * 构建知识图谱
 */
export async function buildGraph(req: Request, res: Response) {
  try {
    const request = req.body as BuildGraphRequest;

    if (!request.documentId || !request.content) {
      res.status(400).json({
        success: false,
        error: '缺少必要参数：documentId, content',
      });
      return;
    }

    logger.info('Build graph request', {
      documentId: request.documentId,
      title: request.title,
      contentLength: request.content.length,
    });

    const apiKeyOverrides = resolveApiKeyOverrides(req);
    const result = await withRequestApiKeys(apiKeyOverrides, async () => runBuildGraphWorkflow(request));
    res.json(result);
  } catch (error) {
    logger.error('Build graph error', { error });
    res.status(500).json({
      success: false,
      entityCount: 0,
      relationCount: 0,
      error: error instanceof Error ? error.message : 'Internal server error',
    });
  }
}

/**
 * 删除文档知识图谱节点
 */
export async function deleteDocumentNodes(req: Request, res: Response) {
  try {
    const { documentId } = req.body;

    if (!documentId) {
      res.status(400).json({
        success: false,
        error: '缺少必要参数：documentId',
      });
      return;
    }

    logger.info('Delete document nodes request', { documentId });

    const { deleteDocumentNodes: deleteNodes } = await import('../../infrastructure/tools/neo4j');
    await deleteNodes(documentId);

    res.json({
      success: true,
      message: '文档节点删除成功',
    });
  } catch (error) {
    logger.error('Delete document nodes error', { error });
    res.status(500).json({
      success: false,
      error: error instanceof Error ? error.message : 'Internal server error',
    });
  }
}

/**
 * 重新生成某个环节
 */
export async function regenerateSection(req: Request, res: Response) {
  try {
    const request = req.body as RegenerateSectionRequest;

    if (!request.lessonId || !request.section || !request.context) {
      res.status(400).json({
        success: false,
        error: '缺少必要参数：lessonId, section, context',
      });
      return;
    }

    logger.info('Regenerate section request', {
      lessonId: request.lessonId,
      section: request.section,
    });

    const apiKeyOverrides = resolveApiKeyOverrides(req);

    const { content, usage } = await withRequestApiKeys(apiKeyOverrides, async () => {
      const deepseek = getDeepSeekClient();
      const prompt = buildRegenerationPrompt(request);
      return deepseek.chat(
        [
          { role: 'system', content: '你是一位经验丰富的教学设计专家。请根据用户的要求重新生成教案的指定部分。' },
          { role: 'user', content: prompt },
        ],
        { temperature: 0.7 }
      );
    });

    res.json({
      success: true,
      section: request.section,
      content,
      usage,
    });
  } catch (error) {
    logger.error('Regenerate section error', { error });
    res.status(500).json({
      success: false,
      error: error instanceof Error ? error.message : 'Internal server error',
    });
  }
}



export async function chatAssistant(req: Request, res: Response) {
  try {
    const request = req.body as AssistantChatRequest;
    const question = typeof request.question === 'string' ? request.question.trim() : '';

    if (!question) {
      res.status(400).json({
        success: false,
        error: '缺少必要参数：question',
      });
      return;
    }

    const history = normalizeAssistantHistory(request.history).slice(-MAX_ASSISTANT_HISTORY);
    const messages: DeepSeekMessage[] = [
      { role: 'system', content: ASSISTANT_SYSTEM_PROMPT },
      ...history,
      { role: 'user', content: question },
    ];

    const apiKeyOverrides = resolveApiKeyOverrides(req);
    const { content, usage } = await withRequestApiKeys(apiKeyOverrides, async () => {
      const deepseek = getDeepSeekClient();
      return deepseek.chat(messages, {
        temperature: 0.3,
        maxTokens: 1800,
      });
    });

    res.json({
      success: true,
      data: {
        answer: content.trim(),
        suggestions: buildAssistantSuggestions(question),
      },
      usage,
    });
  } catch (error) {
    logger.error('Assistant chat error', { error });
    res.status(500).json({
      success: false,
      error: error instanceof Error ? error.message : 'Internal server error',
    });
  }
}

/**
 * 知识图谱查询
 */

/**
 * 生成 Embedding
 */
export async function createEmbedding(req: Request, res: Response) {
  try {
    const { text } = req.body as { text?: string };
    if (!text || !text.trim()) {
      res.status(400).json({
        error: '缺少必要参数：text',
      });
      return;
    }

    const apiKeyOverrides = resolveApiKeyOverrides(req);
    const embedding = await withRequestApiKeys(apiKeyOverrides, async () => {
      const deepseek = getDeepSeekClient();
      return deepseek.createEmbedding(text);
    });

    res.json({
      embedding,
    });
  } catch (error) {
    logger.error('Create embedding error', { error });
    res.status(500).json({
      error: error instanceof Error ? error.message : 'Internal server error',
    });
  }
}

export async function queryKnowledge(req: Request, res: Response) {
  try {
    const { subject, grade, topic } = req.query;

    if (!subject || !grade) {
      res.status(400).json({
        success: false,
        error: '缺少必要参数：subject, grade',
      });
      return;
    }

    const { getNeo4jTool } = await import('../../infrastructure/tools/neo4j');
    const neo4j = getNeo4jTool();

    const knowledgePoints = await neo4j.getKnowledgePoints(
      subject as string,
      grade as string,
      topic as string | undefined
    );

    res.json({
      success: true,
      data: knowledgePoints,
    });
  } catch (error) {
    logger.error('Knowledge query error', { error });
    res.status(500).json({
      success: false,
      error: error instanceof Error ? error.message : 'Internal server error',
    });
  }
}

/**
 * 知识图谱子图
 */
export async function getKnowledgeSubgraph(req: Request, res: Response) {
  try {
    const id = req.params.id;
    if (!id) {
      res.status(400).json({ success: false, error: 'Missing knowledge id' });
      return;
    }
    const depth = parseInt(req.query.depth as string) || 2;

    const { getGraphRAG } = await import('../../modules/knowledge/rag/graphRag');
    const graphRag = getGraphRAG();

    const subgraph = await graphRag.getKnowledgeSubgraph(id, depth);

    res.json({
      success: true,
      data: subgraph,
    });
  } catch (error) {
    logger.error('Knowledge subgraph error', { error });
    res.status(500).json({
      success: false,
      error: error instanceof Error ? error.message : 'Internal server error',
    });
  }
}

export async function getLangSmithTokenUsage(req: Request, res: Response) {
  try {
    const userId = String(req.query.userId || '').trim();
    if (!userId) {
      res.status(400).json({
        success: false,
        error: '缺少必要参数：userId',
      });
      return;
    }

    if (!config.langsmith.enabled || !config.langsmith.apiKey) {
      res.status(503).json({
        success: false,
        error: 'LangSmith tracing 未启用或未配置 API Key',
      });
      return;
    }

    const page = parsePositiveInt(req.query.page, DEFAULT_LANGSMITH_PAGE);
    const pageSize = parsePositiveInt(req.query.pageSize, DEFAULT_LANGSMITH_PAGE_SIZE, MAX_LANGSMITH_PAGE_SIZE);

    const cacheKey = buildLangSmithCacheKey(userId, page, pageSize);
    const cachedPayload = getLangSmithUsageFromCache(cacheKey);
    if (cachedPayload) {
      res.json(cachedPayload);
      return;
    }

    const client = new LangSmithClient({
      apiUrl: config.langsmith.endpoint,
      apiKey: config.langsmith.apiKey,
      timeout_ms: 30000,
    });

    const userRuns: LangSmithRun[] = [];
    const listRunAttempts: Array<{ projectName: string; isRoot?: boolean; limit?: number }> = [
      { projectName: config.langsmith.project, isRoot: true, limit: MAX_LANGSMITH_FETCH_LIMIT },
      { projectName: config.langsmith.project, isRoot: true, limit: 200 },
      { projectName: config.langsmith.project, isRoot: true },
      { projectName: config.langsmith.project, limit: 200 },
      { projectName: config.langsmith.project },
    ];

    let lastListRunsError: unknown = null;
    let listRunsSucceeded = false;

    for (const attempt of listRunAttempts) {
      userRuns.length = 0;

      try {
        for await (const run of client.listRuns(attempt)) {
          if (extractRunUserId(run) === userId) {
            userRuns.push(run);

            if (userRuns.length >= MAX_LANGSMITH_FETCH_LIMIT) {
              break;
            }
          }
        }

        listRunsSucceeded = true;
        break;
      } catch (attemptError) {
        lastListRunsError = attemptError;
        const status = getErrorStatus(attemptError);

        logger.warn('LangSmith listRuns attempt failed', {
          status,
          attempt,
          error: attemptError instanceof Error ? attemptError.message : String(attemptError),
        });

        if (status !== 400) {
          throw attemptError;
        }
      }
    }

    if (!listRunsSucceeded) {
      throw lastListRunsError || new Error('LangSmith listRuns failed');
    }

    userRuns.sort((first, second) => {
      const secondStart = toTimestamp(second.start_time) || 0;
      const firstStart = toTimestamp(first.start_time) || 0;
      return secondStart - firstStart;
    });

    const total = userRuns.length;
    const totalPages = total > 0 ? Math.ceil(total / pageSize) : 0;
    const offset = (page - 1) * pageSize;
    const items = userRuns.slice(offset, offset + pageSize).map(buildHistoryItem);

    const payload: LangSmithUsageResponse = {
      success: true,
      source: 'langsmith',
      project: config.langsmith.project,
      stats: buildUsageStats(userRuns),
      history: {
        items,
        total,
        page,
        pageSize,
        totalPages,
      },
    };

    setLangSmithUsageCache(cacheKey, payload);
    res.json(payload);
  } catch (error) {
    logger.error('LangSmith token usage query failed', {
      status: getErrorStatus(error),
      message: error instanceof Error ? error.message : String(error),
      error,
    });
    res.status(500).json({
      success: false,
      error: error instanceof Error ? error.message : 'Internal server error',
    });
  }
}

/**
 * 构建重新生成提示词
 */
function buildRegenerationPrompt(request: RegenerateSectionRequest): string {
  const { section, context } = request;
  const { subject, grade, topic, duration, current } = context;

  let prompt = `请为以下课程重新生成"${section}"部分：

基本信息：
- 学科：${subject}
- 年级：${grade}
- 课题：${topic}
- 课时：${duration}分钟

`;

  if (current && Object.keys(current).length > 0) {
    prompt += `当前内容（需要改进）：
${JSON.stringify(current, null, 2)}

请生成更好的版本。`;
  }

  return prompt;
}
