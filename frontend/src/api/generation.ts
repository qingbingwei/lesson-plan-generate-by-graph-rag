import axios from 'axios';
import api, { agentApi } from './index';
import { useAuthStore } from '@/stores/auth';
import type { 
  GenerateLessonRequest,
  ApiResponse
} from '@/types';

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

type RawGenerationHistoryResponse = {
  items?: GenerationHistoryItem[];
  total?: number;
  page?: number;
  page_size?: number;
  pageSize?: number;
  total_pages?: number;
  totalPages?: number;
};

export interface LangSmithUsageResponse {
  source: string;
  project?: string;
  stats: DashboardStats;
  history: GenerationHistoryResponse;
}

function normalizeHistory(raw: RawGenerationHistoryResponse, fallbackPage: number, fallbackPageSize: number): GenerationHistoryResponse {
  const total = Number(raw.total || 0);
  const resolvedPage = Number(raw.page || fallbackPage || 1);
  const resolvedPageSize = Number(raw.pageSize || raw.page_size || fallbackPageSize || 10);
  const resolvedTotalPages = Number(
    raw.totalPages || raw.total_pages || (resolvedPageSize > 0 ? Math.ceil(total / resolvedPageSize) : 0)
  );

  return {
    items: Array.isArray(raw.items) ? raw.items : [],
    total,
    page: resolvedPage > 0 ? resolvedPage : 1,
    pageSize: resolvedPageSize > 0 ? resolvedPageSize : 10,
    totalPages: resolvedTotalPages >= 0 ? resolvedTotalPages : 0,
  };
}

export async function getGenerationHistory(page: number = 1, pageSize: number = 10): Promise<GenerationHistoryResponse> {
  const response = await api.get<ApiResponse<RawGenerationHistoryResponse>>('/generate/history', {
    params: {
      page,
      page_size: pageSize,
    },
  });

  return normalizeHistory(response.data.data || {}, page, pageSize);
}

export async function getLangSmithUsage(page: number = 1, pageSize: number = 10): Promise<LangSmithUsageResponse> {
  try {
    const response = await api.get<ApiResponse<LangSmithUsageResponse>>('/generate/langsmith/usage', {
      params: {
        page,
        page_size: pageSize,
      },
    });

    return {
      ...response.data.data,
      history: normalizeHistory(response.data.data?.history || {}, page, pageSize),
    };
  } catch (error) {
    const statusCode = axios.isAxiosError(error) ? error.response?.status : undefined;

    const shouldFallbackToAgent = statusCode === 404 || (typeof statusCode === 'number' && statusCode >= 500);
    if (!shouldFallbackToAgent) {
      throw error;
    }

    const authStore = useAuthStore();
    const userId = authStore.user?.id;

    if (!userId) {
      throw new Error('无法获取当前用户信息，请重新登录后重试');
    }

    const fallback = await agentApi.get<LangSmithUsageResponse & { success?: boolean; error?: string }>(
      '/api/langsmith/token-usage',
      {
        params: {
          userId: String(userId),
          page,
          pageSize,
        },
      }
    );

    if (fallback.data && fallback.data.success === false) {
      throw new Error(fallback.data.error || '获取 LangSmith 数据失败');
    }

    return {
      source: fallback.data.source || 'langsmith',
      project: fallback.data.project,
      stats: fallback.data.stats,
      history: normalizeHistory(fallback.data.history || {}, page, pageSize),
    };
  }
}

export async function getTokenUsageBundle(
  page: number = 1,
  pageSize: number = 10,
  options?: { fallbackStats?: DashboardStats }
): Promise<LangSmithUsageResponse> {
  try {
    return await getLangSmithUsage(page, pageSize);
  } catch (langSmithError) {
    try {
      const stats = options?.fallbackStats || (await getGenerationStats());
      const history = await getGenerationHistory(page, pageSize).catch(() =>
        normalizeHistory({}, page, pageSize)
      );

      return {
        source: 'generation_db',
        stats,
        history,
      };
    } catch {
      throw langSmithError;
    }
  }
}
