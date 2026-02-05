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

export interface DocumentListResponse {
  documents: KnowledgeDocument[];
  total: number;
}

export const knowledgeApi = {
  // 上传文档
  uploadDocument: (formData: FormData) => {
    return api.post('/knowledge/documents', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
  },

  // 获取文档列表
  listDocuments: (page = 1, pageSize = 20) => {
    return api.get<{ data: KnowledgeDocument[] }>('/knowledge/documents', {
      params: { page, page_size: pageSize },
    });
  },

  // 获取文档详情
  getDocument: (id: string) => {
    return api.get<{ data: KnowledgeDocument }>(`/knowledge/documents/${id}`);
  },

  // 删除文档
  deleteDocument: (id: string) => {
    return api.delete(`/knowledge/documents/${id}`);
  },

  // 获取文档处理状态
  getDocumentStatus: (id: string) => {
    return api.get<{ data: { status: string; entityCount: number; relationCount: number } }>(
      `/knowledge/documents/${id}/status`
    );
  },
};
