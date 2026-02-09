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
} from '../controllers/lessonController';

const router = Router();

// 健康检查
router.get('/health', healthCheck);

// 教案生成
router.post('/api/generate', generateLesson);
router.post('/api/generate/stream', streamGenerateLesson);
router.post('/api/regenerate-section', regenerateSection);

// 知识图谱
router.post('/api/build-graph', buildGraph);
router.post('/api/delete-document-nodes', deleteDocumentNodes);
router.get('/api/knowledge', queryKnowledge);
router.get('/api/knowledge/:id/subgraph', getKnowledgeSubgraph);

export default router;
