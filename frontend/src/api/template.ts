import api from './index';
import type {
  ApiResponse,
  LessonTemplate,
  CreateLessonTemplateRequest,
  AppliedLessonTemplate,
} from '@/types';

export async function listLessonTemplates(): Promise<LessonTemplate[]> {
  const response = await api.get<ApiResponse<LessonTemplate[]>>('/templates');
  return response.data.data;
}

export async function createLessonTemplate(payload: CreateLessonTemplateRequest): Promise<LessonTemplate> {
  const response = await api.post<ApiResponse<LessonTemplate>>('/templates', payload);
  return response.data.data;
}

export async function deleteLessonTemplate(id: string): Promise<void> {
  await api.delete(`/templates/${id}`);
}

export async function applyLessonTemplate(id: string): Promise<AppliedLessonTemplate> {
  const response = await api.post<ApiResponse<AppliedLessonTemplate>>(`/templates/${id}/apply`);
  return response.data.data;
}
