import api from './index';
import type { 
  GenerateLessonRequest, 
  KnowledgePoint,
  KnowledgeGraphData,
  ApiResponse,
} from '@/types';
import { getApiKeyHeaders } from '@/utils/apiKeys';

// 后端返回的生成响应结构
interface BackendGenerationResponse {
  id: string;
  status: string;
  title?: string;
  objectives?: string;
  key_points?: string;
  difficult_points?: string;
  teaching_methods?: string;
  content?: string;
  activities?: string;
  assessment?: string;
  resources?: string;
  token_count: number;
  duration_ms: number;
  error_message?: string;
}

/**
 * 生成教案 - 通过后端 API 调用
 */
export async function generateLesson(
  request: GenerateLessonRequest
): Promise<BackendGenerationResponse> {
  const response = await api.post<ApiResponse<BackendGenerationResponse>>('/generate', request);
  return response.data.data;
}

/**
 * 流式生成教案
 */
export function streamGenerateLesson(
  request: GenerateLessonRequest,
  onProgress: (event: { node: string; state: Record<string, unknown> }) => void,
  onComplete: () => void,
  onError: (error: Error) => void
): () => void {
  const controller = new AbortController();
  
  // 获取 token (pinia-plugin-persistedstate 存储格式)
  let token = '';
  try {
    const authData = localStorage.getItem('auth');
    if (authData) {
      const parsed = JSON.parse(authData);
      token = parsed.token || '';
    }
  } catch {
    token = '';
  }
  
  fetch('/api/v1/generate/stream', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
      ...getApiKeyHeaders(),
    },
    body: JSON.stringify(request),
    signal: controller.signal,
  })
    .then(async (response) => {
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const reader = response.body?.getReader();
      if (!reader) {
        throw new Error('No response body');
      }

      const decoder = new TextDecoder();
      let buffer = '';

      while (true) {
        const { done, value } = await reader.read();
        
        if (done) {
          onComplete();
          break;
        }

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() || '';

        for (const line of lines) {
          if (line.startsWith('data: ')) {
            const data = line.slice(6);
            
            if (data === '[DONE]') {
              onComplete();
              return;
            }

            try {
              const event = JSON.parse(data);
              onProgress(event);
            } catch {
              // 忽略解析错误
            }
          }
        }
      }
    })
    .catch((error) => {
      if (error.name !== 'AbortError') {
        onError(error);
      }
    });

  // 返回取消函数
  return () => controller.abort();
}

/**
 * 重新生成某个环节
 */
export async function regenerateSection(
  lessonId: string,
  section: string,
  context: {
    subject: string;
    grade: string;
    topic: string;
    duration: number;
    current: Record<string, unknown>;
  }
): Promise<{ section: string; content: Record<string, unknown> }> {
  const response = await api.post<ApiResponse<{ section: string; content: Record<string, unknown> }>>('/generate/regenerate-section', {
    lessonId,
    section,
    context,
  });
  return response.data.data;
}

/**
 * 查询知识点
 */
export async function getKnowledgePoints(
  subject: string,
  grade: string,
  topic?: string
): Promise<KnowledgePoint[]> {
  const response = await api.get<ApiResponse<KnowledgePoint[]>>(
    '/knowledge/search',
    { params: { subject, grade, topic } }
  );
  return response.data.data;
}

/**
 * 获取知识图谱子图
 */
export async function getKnowledgeSubgraph(
  id: string,
  depth: number = 2
): Promise<KnowledgeGraphData> {
  const response = await api.get<ApiResponse<KnowledgeGraphData>>(
    '/knowledge/graph',
    { params: { id, depth } }
  );
  return response.data.data;
}

// 统计数据结构（与后端GenerationStats对应）
export interface DashboardStats {
  total_count: number;
  completed_count: number;
  failed_count: number;
  total_tokens: number;
  avg_duration_ms: number;
  this_month_generations: number;
  total_lessons: number;
}

/**
 * 获取生成统计数据
 */
export async function getGenerationStats(): Promise<DashboardStats> {
  const response = await api.get<ApiResponse<DashboardStats>>('/generate/stats');
  return response.data.data;
}


export interface GenerationHistoryItem {
  id: string;
  status: string;
  prompt: string;
  token_count: number;
  duration_ms: number;
  error_msg?: string;
  created_at: string;
  completed_at?: string;
}

export interface GenerationHistoryResponse {
  items: GenerationHistoryItem[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

/**
 * 获取生成历史（含 token 使用量）
 */
export async function getGenerationHistory(page: number = 1, pageSize: number = 10): Promise<GenerationHistoryResponse> {
  const response = await api.get<ApiResponse<Record<string, unknown>>>('/generate/history', {
    params: {
      page,
      page_size: pageSize,
    },
  });

  const data = response.data.data as {
    items?: GenerationHistoryItem[];
    total?: number;
    page?: number;
    page_size?: number;
    pageSize?: number;
    total_pages?: number;
    totalPages?: number;
  };

  return {
    items: data.items || [],
    total: data.total || 0,
    page: data.page || page,
    pageSize: data.page_size || data.pageSize || pageSize,
    totalPages: data.total_pages || data.totalPages || 1,
  };
}
