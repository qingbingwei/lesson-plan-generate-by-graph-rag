import api from './index';

export interface KnowledgeDocument {
  id: string;
  title: string;
  fileName: string;
  fileType: string;
  fileSize: number;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  entityCount: number;
  relationCount: number;
  subject?: string;
  grade?: string;
  errorMsg?: string;
  createdAt: string;
  updatedAt: string;
}

/** 上传文档 */
export function uploadDocument(formData: FormData, onProgress?: (percent: number) => void) {
  return api.post('/knowledge/documents', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
    onUploadProgress: (event) => {
      if (onProgress && event.total) {
        const percent = Math.round((event.loaded * 100) / event.total);
        onProgress(percent);
      }
    },
  });
}

/** 获取文档列表 */
export function listDocuments(page = 1, pageSize = 20) {
  return api.get<{ data: KnowledgeDocument[] }>('/knowledge/documents', {
    params: { page, page_size: pageSize },
  });
}

/** 删除文档 */
export function deleteDocument(id: string) {
  return api.delete(`/knowledge/documents/${id}`);
}

// 保持向后兼容的命名空间导出
export const knowledgeApi = {
  uploadDocument,
  listDocuments,
  deleteDocument,
};
