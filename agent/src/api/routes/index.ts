import { Router } from 'express';
import {
  healthCheck,
  generateLesson,
  streamGenerateLesson,
  buildGraph,
  deleteDocumentNodes,
  regenerateSection,
  queryKnowledge,
  getKnowledgeSubgraph,
  createEmbedding,
  getLangSmithTokenUsage,
  chatAssistant,
} from '../controllers/lessonController';
import { snapshotMetrics } from '../../shared/observability/metrics';

const router = Router();

// 健康检查
router.get('/health', healthCheck);
router.get('/metrics', (_req, res) => {
  res.json({
    success: true,
    code: 0,
    message: 'success',
    data: snapshotMetrics(),
    trace_id: res.locals.traceId,
  });
});

// 教案生成
router.post('/api/generate', generateLesson);
router.post('/api/generate/stream', streamGenerateLesson);
router.post('/api/assistant/chat', chatAssistant);
router.post('/api/regenerate-section', regenerateSection);
router.post('/api/embedding', createEmbedding);

// 知识图谱
router.post('/api/build-graph', buildGraph);
router.post('/api/delete-document-nodes', deleteDocumentNodes);
router.get('/api/knowledge', queryKnowledge);
router.get('/api/knowledge/:id/subgraph', getKnowledgeSubgraph);
router.get('/api/langsmith/token-usage', getLangSmithTokenUsage);

export default router;
