import api from './index';
import type {
  Lesson,
  LessonVersion,
  LessonComment,
  LessonFavorite,
  ApiResponse,
  PaginatedResponse,
  PaginationParams,
} from '@/types';

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
  const response = await api.get<ApiResponse<PaginatedResponse<Lesson>>>('/lessons', { params });
  return response.data.data;
}

/**
 * 获取教案详情
 */
export async function getLesson(id: string): Promise<Lesson> {
  const response = await api.get<ApiResponse<Lesson>>(`/lessons/${id}`);
  return response.data.data;
}

/**
 * 创建教案
 */
export async function createLesson(data: Partial<Lesson>): Promise<Lesson> {
  const response = await api.post<ApiResponse<Lesson>>('/lessons', data);
  return response.data.data;
}

/**
 * 更新教案
 */
export async function updateLesson(id: string, data: Partial<Lesson>): Promise<Lesson> {
  const response = await api.put<ApiResponse<Lesson>>(`/lessons/${id}`, data);
  return response.data.data;
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
  const response = await api.post<ApiResponse<Lesson>>(`/lessons/${lessonId}/versions/${version}/rollback`);
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
