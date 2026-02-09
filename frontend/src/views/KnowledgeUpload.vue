<template>
  <div class="knowledge-upload">
    <div class="page-header">
      <h1>知识库管理</h1>
      <p class="subtitle">上传文档，自动构建个人知识图谱</p>
    </div>

    <!-- 上传区域 -->
    <div class="upload-section">
      <div 
        class="upload-area" 
        :class="{ 'drag-over': isDragOver }"
        @dragover.prevent="isDragOver = true"
        @dragleave.prevent="isDragOver = false"
        @drop.prevent="handleDrop"
        @click="triggerFileInput"
      >
        <input 
          ref="fileInputRef"
          type="file" 
          accept=".txt,.md" 
          @change="handleFileSelect" 
          class="hidden-input"
        />
        <div class="upload-icon">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5m-13.5-9L12 3m0 0l4.5 4.5M12 3v13.5" />
          </svg>
        </div>
        <p class="upload-text">拖拽文件到此处，或 <span class="upload-link">点击上传</span></p>
        <p class="upload-hint">支持 .txt 和 .md 格式，单个文件不超过 5MB</p>
      </div>

      <!-- 上传选项 -->
      <div v-if="selectedFile" class="upload-options">
        <div class="file-info">
          <span class="file-name">{{ selectedFile.name }}</span>
          <span class="file-size">{{ formatFileSize(selectedFile.size) }}</span>
          <button class="btn-remove" @click="removeFile">✕</button>
        </div>
        
        <div class="options-form">
          <div class="form-group">
            <label>文档标题</label>
            <input v-model="uploadForm.title" type="text" placeholder="可选，默认使用文件名" />
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>学科（可选）</label>
              <select v-model="uploadForm.subject">
                <option value="">不指定</option>
                <option value="语文">语文</option>
                <option value="数学">数学</option>
                <option value="英语">英语</option>
                <option value="物理">物理</option>
                <option value="化学">化学</option>
                <option value="生物">生物</option>
                <option value="历史">历史</option>
                <option value="地理">地理</option>
                <option value="政治">政治</option>
                <option value="科学">科学</option>
                <option value="信息技术">信息技术</option>
                <option value="音乐">音乐</option>
                <option value="美术">美术</option>
                <option value="体育">体育</option>
              </select>
            </div>
            <div class="form-group">
              <label>年级（可选）</label>
              <select v-model="uploadForm.grade">
                <option value="">不指定</option>
                <option value="7">七年级</option>
                <option value="8">八年级</option>
                <option value="9">九年级</option>
                <option value="10">高一</option>
                <option value="11">高二</option>
                <option value="12">高三</option>
              </select>
            </div>
          </div>
          <button class="btn-upload" @click="uploadDocument" :disabled="uploading">
            <span v-if="uploading">上传中... {{ uploadProgress }}%</span>
            <span v-else>开始上传</span>
          </button>
          <!-- 上传进度条 -->
          <div v-if="uploading" class="upload-progress">
            <div class="progress-bar">
              <div class="progress-fill" :style="{ width: uploadProgress + '%' }"></div>
            </div>
            <span class="progress-text">{{ uploadProgress }}%</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 文档列表 -->
    <div class="documents-section">
      <h2>我的知识文档</h2>
      
      <div v-if="loading" class="loading-state">
        <div class="spinner"></div>
        <p>加载中...</p>
      </div>

      <div v-else-if="documents.length === 0" class="empty-state">
        <p>暂无上传的文档</p>
      </div>

      <div v-else class="document-list">
        <div v-for="doc in documents" :key="doc.id" class="document-card">
          <div class="doc-icon">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m2.25 0H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z" />
            </svg>
          </div>
          <div class="doc-info">
            <h3>{{ doc.title }}</h3>
            <p class="doc-meta">
              <span>{{ doc.fileName }}</span>
              <span>{{ formatFileSize(doc.fileSize) }}</span>
              <span>{{ formatDate(doc.createdAt) }}</span>
            </p>
            <div class="doc-stats" v-if="doc.status === 'completed'">
              <span class="stat">{{ doc.entityCount }} 个实体</span>
              <span class="stat">{{ doc.relationCount }} 个关系</span>
            </div>
          </div>
          <div class="doc-status">
            <span :class="['status-badge', `status-${doc.status}`]">
              <span v-if="doc.status === 'processing' || doc.status === 'pending'" class="status-spinner"></span>
              <span class="status-text">{{ statusText[doc.status] || doc.status }}</span>
            </span>
          </div>
          <div class="doc-actions">
            <button class="btn-action btn-delete" @click="deleteDocument(doc.id)" title="删除">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
              </svg>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { knowledgeApi } from '@/api';

interface KnowledgeDocument {
  id: string;
  title: string;
  fileName: string;
  fileType: string;
  fileSize: number;
  status: string;
  entityCount: number;
  relationCount: number;
  subject?: string;
  grade?: string;
  createdAt: string;
}

const documents = ref<KnowledgeDocument[]>([]);
const loading = ref(false);
const uploading = ref(false);
const uploadProgress = ref(0);
const isDragOver = ref(false);
const selectedFile = ref<File | null>(null);
const fileInputRef = ref<HTMLInputElement | null>(null);

const uploadForm = ref({
  title: '',
  subject: '',
  grade: '',
});

// 触发文件选择
const triggerFileInput = () => {
  fileInputRef.value?.click();
};

const statusText: Record<string, string> = {
  pending: '待处理',
  processing: '处理中',
  completed: '已完成',
  failed: '处理失败',
};

// 加载文档列表
const loadDocuments = async () => {
  loading.value = true;
  try {
    const response = await knowledgeApi.listDocuments();
    documents.value = response.data.data || [];
  } catch (error) {
    console.error('Failed to load documents:', error);
  } finally {
    loading.value = false;
  }
};

// 处理文件选择
const handleFileSelect = (event: Event) => {
  const input = event.target as HTMLInputElement;
  if (input.files && input.files[0]) {
    selectFile(input.files[0]);
  }
};

// 处理拖拽
const handleDrop = (event: DragEvent) => {
  isDragOver.value = false;
  const files = event.dataTransfer?.files;
  if (files && files[0]) {
    selectFile(files[0]);
  }
};

// 选择文件
const selectFile = (file: File) => {
  const ext = file.name.split('.').pop()?.toLowerCase();
  if (ext !== 'txt' && ext !== 'md') {
    alert('仅支持 .txt 和 .md 格式文件');
    return;
  }
  if (file.size > 5 * 1024 * 1024) {
    alert('文件大小不能超过 5MB');
    return;
  }
  selectedFile.value = file;
  uploadForm.value.title = file.name.replace(/\.(txt|md)$/i, '');
};

// 移除文件
const removeFile = () => {
  selectedFile.value = null;
  uploadForm.value = { title: '', subject: '', grade: '' };
};

// 上传文档
const uploadDocument = async () => {
  if (!selectedFile.value) return;
  
  uploading.value = true;
  uploadProgress.value = 0;
  try {
    const formData = new FormData();
    formData.append('file', selectedFile.value);
    if (uploadForm.value.title) {
      formData.append('title', uploadForm.value.title);
    }
    if (uploadForm.value.subject) {
      formData.append('subject', uploadForm.value.subject);
    }
    if (uploadForm.value.grade) {
      formData.append('grade', uploadForm.value.grade);
    }
    
    await knowledgeApi.uploadDocument(formData, (percent) => {
      uploadProgress.value = percent;
    });
    
    // 重置表单
    removeFile();
    
    // 刷新列表
    await loadDocuments();
    
    alert('文档上传成功，正在后台处理中');
  } catch (error) {
    console.error('Upload failed:', error);
    alert('上传失败，请重试');
  } finally {
    uploading.value = false;
  }
};

// 删除文档
const deleteDocument = async (id: string) => {
  if (!confirm('确定要删除该文档吗？相关的知识图谱数据也会被删除。')) {
    return;
  }
  
  try {
    await knowledgeApi.deleteDocument(id);
    // 重置上传状态，允许重新上传
    selectedFile.value = null;
    uploadForm.value = { title: '', subject: '', grade: '' };
    // 重置文件输入
    if (fileInputRef.value) {
      fileInputRef.value.value = '';
    }
    await loadDocuments();
  } catch (error) {
    console.error('Delete failed:', error);
    alert('删除失败');
  }
};

// 格式化文件大小
const formatFileSize = (bytes: number): string => {
  if (bytes < 1024) return bytes + ' B';
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
};

// 格式化日期
const formatDate = (dateStr: string): string => {
  const date = new Date(dateStr);
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  });
};

onMounted(() => {
  loadDocuments();
  
  // 轮询更新处理中的文档状态
  setInterval(() => {
    const hasProcessing = documents.value.some(d => d.status === 'processing' || d.status === 'pending');
    if (hasProcessing) {
      loadDocuments();
    }
  }, 5000);
});
</script>

<style scoped>
.knowledge-upload {
  max-width: 900px;
  margin: 0 auto;
  padding: 2rem;
}

.page-header {
  margin-bottom: 2rem;
}

.page-header h1 {
  font-size: 1.75rem;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 0.5rem;
}

.subtitle {
  color: #6b7280;
}

/* 上传区域 */
.upload-section {
  background: white;
  border-radius: 12px;
  padding: 2rem;
  margin-bottom: 2rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.upload-area {
  display: block;
  border: 2px dashed #d1d5db;
  border-radius: 8px;
  padding: 3rem 2rem;
  text-align: center;
  transition: all 0.2s;
  cursor: pointer;
  position: relative;
}

.upload-area:hover,
.upload-area.drag-over {
  border-color: #3b82f6;
  background: #eff6ff;
}

.hidden-input {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

.upload-icon {
  width: 48px;
  height: 48px;
  margin: 0 auto 1rem;
  color: #9ca3af;
}

.upload-icon svg {
  width: 100%;
  height: 100%;
}

.upload-text {
  color: #4b5563;
  margin-bottom: 0.5rem;
}

.upload-link {
  color: #3b82f6;
  cursor: pointer;
}

.upload-link:hover {
  text-decoration: underline;
}

.upload-hint {
  font-size: 0.875rem;
  color: #9ca3af;
}

/* 上传选项 */
.upload-options {
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 1px solid #e5e7eb;
}

.file-info {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.75rem 1rem;
  background: #f3f4f6;
  border-radius: 6px;
  margin-bottom: 1rem;
}

.file-name {
  font-weight: 500;
  color: #1f2937;
}

.file-size {
  color: #6b7280;
  font-size: 0.875rem;
}

.btn-remove {
  margin-left: auto;
  background: none;
  border: none;
  color: #9ca3af;
  cursor: pointer;
  font-size: 1.25rem;
}

.btn-remove:hover {
  color: #ef4444;
}

.options-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.form-group label {
  display: block;
  font-size: 0.875rem;
  font-weight: 500;
  color: #374151;
  margin-bottom: 0.5rem;
}

.form-group input,
.form-group select {
  width: 100%;
  padding: 0.5rem 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 0.875rem;
}

.form-group input:focus,
.form-group select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.btn-upload {
  padding: 0.75rem 1.5rem;
  background: #3b82f6;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 1rem;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.2s;
}

.btn-upload:hover:not(:disabled) {
  background: #2563eb;
}

.btn-upload:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* 文档列表 */
.documents-section {
  background: white;
  border-radius: 12px;
  padding: 2rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.documents-section h2 {
  font-size: 1.25rem;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 1.5rem;
}

.loading-state,
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 3rem;
  color: #6b7280;
  min-height: 120px;
}

.loading-state p,
.empty-state p {
  margin: 0;
  display: block;
  white-space: nowrap;
  font-size: 0.875rem;
}

.spinner {
  display: block;
  width: 32px;
  height: 32px;
  border: 3px solid #e5e7eb;
  border-top-color: #3b82f6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 1rem;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.document-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.document-card {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  transition: all 0.2s;
}

.document-card:hover {
  border-color: #d1d5db;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.doc-icon {
  width: 40px;
  height: 40px;
  background: #eff6ff;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #3b82f6;
  flex-shrink: 0;
}

.doc-icon svg {
  width: 24px;
  height: 24px;
}

.doc-info {
  flex: 1;
  min-width: 0;
}

.doc-info h3 {
  font-size: 1rem;
  font-weight: 500;
  color: #1f2937;
  margin-bottom: 0.25rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.doc-meta {
  font-size: 0.75rem;
  color: #9ca3af;
  display: flex;
  gap: 0.75rem;
}

.doc-stats {
  margin-top: 0.25rem;
}

.stat {
  font-size: 0.75rem;
  color: #059669;
  background: #ecfdf5;
  padding: 0.125rem 0.5rem;
  border-radius: 4px;
  margin-right: 0.5rem;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.75rem;
  padding: 0.25rem 0.75rem;
  border-radius: 9999px;
  white-space: nowrap;
}

.status-spinner {
  display: inline-block;
  width: 12px;
  height: 12px;
  border: 2px solid currentColor;
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  flex-shrink: 0;
}

.status-text {
  flex-shrink: 0;
}

.status-pending {
  background: #fef3c7;
  color: #92400e;
}

.status-processing {
  background: #dbeafe;
  color: #1e40af;
}

.status-completed {
  background: #dcfce7;
  color: #166534;
}

.status-failed {
  background: #fee2e2;
  color: #991b1b;
}

.doc-actions {
  display: flex;
  gap: 0.5rem;
}

.btn-action {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: none;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  color: #6b7280;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-action svg {
  width: 16px;
  height: 16px;
}

.btn-delete:hover {
  color: #ef4444;
  border-color: #fecaca;
  background: #fef2f2;
}

@media (max-width: 640px) {
  .knowledge-upload {
    padding: 1rem;
  }
  
  .form-row {
    grid-template-columns: 1fr;
  }
  
  .document-card {
    flex-wrap: wrap;
  }
  
  .doc-status {
    order: 3;
    width: 100%;
    margin-top: 0.5rem;
  }
}

/* 上传进度条 */
.upload-progress {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-top: 0.75rem;
}

.progress-bar {
  flex: 1;
  height: 8px;
  background: #e5e7eb;
  border-radius: 4px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, #3b82f6, #2563eb);
  border-radius: 4px;
  transition: width 0.3s ease;
}

.progress-text {
  font-size: 0.75rem;
  font-weight: 500;
  color: #3b82f6;
  min-width: 36px;
  text-align: right;
}
</style>
