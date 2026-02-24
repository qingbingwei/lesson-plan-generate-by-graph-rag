import api from './index';
import type {
  Lesson,
  LessonVersion,
  LessonQualityReview,
  LessonVersionDiff,
  ExportLayout,
  LessonComment,
  LessonFavorite,
  ApiResponse,
  PaginatedResponse,
  PaginationParams,
} from '@/types';

type RawLesson = Lesson & {
  user_id?: string;
  created_at?: string;
  updated_at?: string;
  key_points?: string[];
  difficult_points?: string[];
  teaching_methods?: string[];
};

type RawPaginatedLessonResponse = {
  items: RawLesson[];
  total: number;
  page: number;
  page_size?: number;
  pageSize?: number;
  total_pages?: number;
  totalPages?: number;
};

function normalizeLesson(raw: RawLesson): Lesson {
  return {
    ...raw,
    userId: raw.userId || raw.user_id || '',
    keyPoints: raw.keyPoints || raw.key_points || [],
    difficultPoints: raw.difficultPoints || raw.difficult_points || [],
    teachingMethods: raw.teachingMethods || raw.teaching_methods || [],
    createdAt: raw.createdAt || raw.created_at || '',
    updatedAt: raw.updatedAt || raw.updated_at || '',
  } as Lesson;
}

function normalizeLessonPage(raw: RawPaginatedLessonResponse): PaginatedResponse<Lesson> {
  return {
    items: (raw.items || []).map(normalizeLesson),
    total: raw.total || 0,
    page: raw.page || 1,
    pageSize: raw.pageSize || raw.page_size || 10,
    totalPages: raw.totalPages || raw.total_pages || 1,
  };
}

/**
 * 获取教案列表
 */
export async function getLessons(
  params?: PaginationParams & { 
    subject?: string; 
    grade?: string; 
    status?: string;
    keyword?: string;
  }
): Promise<PaginatedResponse<Lesson>> {
  const response = await api.get<ApiResponse<RawPaginatedLessonResponse>>('/lessons', { params });
  return normalizeLessonPage(response.data.data);
}

/**
 * 获取教案详情
 */
export async function getLesson(id: string): Promise<Lesson> {
  const response = await api.get<ApiResponse<RawLesson>>(`/lessons/${id}`);
  return normalizeLesson(response.data.data);
}

/**
 * 创建教案
 */
export async function createLesson(data: Partial<Lesson>): Promise<Lesson> {
  const response = await api.post<ApiResponse<RawLesson>>('/lessons', data);
  return normalizeLesson(response.data.data);
}

/**
 * 更新教案
 */
export async function updateLesson(id: string, data: Partial<Lesson>): Promise<Lesson> {
  const response = await api.put<ApiResponse<RawLesson>>(`/lessons/${id}`, data);
  return normalizeLesson(response.data.data);
}

/**
 * 删除教案
 */
export async function deleteLesson(id: string): Promise<void> {
  await api.delete(`/lessons/${id}`);
}

/**
 * 发布教案
 */
export async function publishLesson(id: string): Promise<void> {
  await api.post(`/lessons/${id}/publish`);
}

/**
 * 获取教案版本列表
 */
export async function getLessonVersions(lessonId: string): Promise<LessonVersion[]> {
  const response = await api.get<ApiResponse<LessonVersion[]>>(`/lessons/${lessonId}/versions`);
  return response.data.data;
}

/**
 * 获取教案指定版本
 */
export async function getLessonVersion(lessonId: string, version: number): Promise<LessonVersion> {
  const response = await api.get<ApiResponse<LessonVersion>>(`/lessons/${lessonId}/versions/${version}`);
  return response.data.data;
}

/**
 * 回滚到指定版本
 */
export async function rollbackToVersion(lessonId: string, version: number): Promise<Lesson> {
  const response = await api.post<ApiResponse<RawLesson>>(`/lessons/${lessonId}/versions/${version}/rollback`);
  return normalizeLesson(response.data.data);
}

/**
 * 获取教案质量审查结果
 */
export async function getLessonQualityReview(lessonId: string): Promise<LessonQualityReview> {
  const response = await api.get<ApiResponse<LessonQualityReview>>(`/lessons/${lessonId}/quality-review`);
  return response.data.data;
}

/**
 * 获取教案版本差异
 */
export async function getLessonVersionDiff(
  lessonId: string,
  fromVersion: number | string,
  toVersion: number | string = 'current'
): Promise<LessonVersionDiff> {
  const response = await api.get<ApiResponse<LessonVersionDiff>>(`/lessons/${lessonId}/versions/diff`, {
    params: {
      from: String(fromVersion),
      to: String(toVersion),
    },
  });
  return response.data.data;
}

/**
 * 获取导出模板列表
 */
export async function getExportLayouts(): Promise<ExportLayout[]> {
  const response = await api.get<ApiResponse<ExportLayout[]>>('/lessons/export/layouts');
  return response.data.data;
}

/**
 * 获取教案评论
 */
export async function getLessonComments(lessonId: number): Promise<LessonComment[]> {
  const response = await api.get<ApiResponse<LessonComment[]>>(`/lessons/${lessonId}/comments`);
  return response.data.data;
}

/**
 * 添加评论
 */
export async function addComment(
  lessonId: number, 
  data: { content: string; parentId?: number }
): Promise<LessonComment> {
  const response = await api.post<ApiResponse<LessonComment>>(
    `/lessons/${lessonId}/comments`, 
    data
  );
  return response.data.data;
}

/**
 * 删除评论
 */
export async function deleteComment(lessonId: number, commentId: number): Promise<void> {
  await api.delete(`/lessons/${lessonId}/comments/${commentId}`);
}

/**
 * 收藏教案
 */
export async function favoriteLesson(lessonId: number): Promise<LessonFavorite> {
  const response = await api.post<ApiResponse<LessonFavorite>>(`/lessons/${lessonId}/favorite`);
  return response.data.data;
}

/**
 * 取消收藏
 */
export async function unfavoriteLesson(lessonId: number): Promise<void> {
  await api.delete(`/lessons/${lessonId}/favorite`);
}

/**
 * 获取我的收藏
 */
export async function getMyFavorites(params?: PaginationParams): Promise<PaginatedResponse<Lesson>> {
  const response = await api.get<ApiResponse<PaginatedResponse<Lesson>>>('/user/favorites', { params });
  return response.data.data;
}

/**
 * 导出教案为 Word
 */
export async function exportLessonAsWord(lessonId: number): Promise<Blob> {
  const response = await api.get(`/lessons/${lessonId}/export/word`, {
    responseType: 'blob',
  });
  return response.data;
}

/**
 * 导出教案为 PDF
 */
export async function exportLessonAsPdf(lessonId: number): Promise<Blob> {
  const response = await api.get(`/lessons/${lessonId}/export/pdf`, {
    responseType: 'blob',
  });
  return response.data;
}
