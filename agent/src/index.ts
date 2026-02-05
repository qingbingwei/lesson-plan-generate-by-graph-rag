import express, { Request, Response, NextFunction } from 'express';
import cors from 'cors';
import helmet from 'helmet';
import compression from 'compression';
import config from './config';
import logger from './utils/logger';
import { runLessonWorkflow, streamLessonWorkflow } from './workflow/lessonWorkflow';
import { runBuildGraphWorkflow, BuildGraphRequest } from './workflow/buildGraphWorkflow';
import type { GenerateLessonRequest, RegenerateSectionRequest } from './types';
import { getDeepSeekClient } from './clients/deepseek';

const app = express();

// 中间件
app.use(helmet());
app.use(cors());
app.use(compression());
app.use(express.json({ limit: '10mb' }));

// 请求日志
app.use((req: Request, _res: Response, next: NextFunction) => {
  logger.info('Incoming request', {
    method: req.method,
    path: req.path,
    ip: req.ip,
  });
  next();
});

// 健康检查
app.get('/health', (_req: Request, res: Response) => {
  res.json({
    status: 'ok',
    timestamp: new Date().toISOString(),
    version: '1.0.0',
  });
});

// 生成教案
app.post('/api/generate', async (req: Request, res: Response) => {
  try {
    const request = req.body as GenerateLessonRequest;

    // 验证请求
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

    const result = await runLessonWorkflow(request);
    res.json(result);
  } catch (error) {
    logger.error('Generate lesson error', { error });
    res.status(500).json({
      success: false,
      error: error instanceof Error ? error.message : 'Internal server error',
    });
  }
});

// 流式生成教案
app.post('/api/generate/stream', async (req: Request, res: Response) => {
  try {
    const request = req.body as GenerateLessonRequest;

    // 验证请求
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

    // 流式生成
    for await (const event of streamLessonWorkflow(request)) {
      res.write(`data: ${JSON.stringify(event)}\n\n`);
    }

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
});

// 构建知识图谱
app.post('/api/build-graph', async (req: Request, res: Response) => {
  try {
    const request = req.body as BuildGraphRequest;

    // 验证请求
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

    const result = await runBuildGraphWorkflow(request);
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
});

// 删除文档相关的知识图谱节点
app.post('/api/delete-document-nodes', async (req: Request, res: Response) => {
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

    const { deleteDocumentNodes } = await import('./tools/neo4j');
    await deleteDocumentNodes(documentId);

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
});

// 重新生成某个环节
app.post('/api/regenerate-section', async (req: Request, res: Response) => {
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

    const deepseek = getDeepSeekClient();

    // 根据不同的 section 类型构建提示词
    const prompt = buildRegenerationPrompt(request);

    const { content, usage } = await deepseek.chat(
      [
        { role: 'system', content: '你是一位经验丰富的教学设计专家。请根据用户的要求重新生成教案的指定部分。' },
        { role: 'user', content: prompt },
      ],
      { temperature: 0.7 }
    );

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
});

// 知识图谱查询
app.get('/api/knowledge', async (req: Request, res: Response) => {
  try {
    const { subject, grade, topic } = req.query;

    if (!subject || !grade) {
      res.status(400).json({
        success: false,
        error: '缺少必要参数：subject, grade',
      });
      return;
    }

    // 动态导入避免循环依赖
    const { getNeo4jTool } = await import('./tools/neo4j');
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
});

// 知识图谱子图
app.get('/api/knowledge/:id/subgraph', async (req: Request, res: Response) => {
  try {
    const id = req.params.id;
    if (!id) {
      res.status(400).json({ success: false, error: 'Missing knowledge id' });
      return;
    }
    const depth = parseInt(req.query.depth as string) || 2;

    const { getGraphRAG } = await import('./rag/graphRag');
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
});

// 构建重新生成提示词
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

// 错误处理
app.use((err: Error, _req: Request, res: Response, _next: NextFunction) => {
  logger.error('Unhandled error', { error: err });
  res.status(500).json({
    success: false,
    error: 'Internal server error',
  });
});

// 启动服务器
const PORT = config.port;

app.listen(PORT, () => {
  logger.info(`Agent service started`, {
    port: PORT,
    env: config.env,
  });
});

export default app;
