<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import {
  Document,
  Delete,
  UploadFilled,
  RefreshRight,
  CircleCheckFilled,
  CircleCloseFilled,
  WarningFilled,
  Loading,
} from '@element-plus/icons-vue';
import type { UploadFile, UploadFiles } from 'element-plus';
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
const selectedFile = ref<File | null>(null);
let pollingTimer: number | null = null;

const uploadForm = ref({
  title: '',
  subject: '',
  grade: '',
});

const subjectOptions = ['语文', '数学', '英语', '物理', '化学', '生物', '历史', '地理', '政治', '科学', '信息技术', '音乐', '美术', '体育'];
const gradeOptions = [
  { label: '七年级', value: '7' },
  { label: '八年级', value: '8' },
  { label: '九年级', value: '9' },
  { label: '高一', value: '10' },
  { label: '高二', value: '11' },
  { label: '高三', value: '12' },
];

const statusText: Record<string, string> = {
  pending: '待处理',
  processing: '处理中',
  completed: '已完成',
  failed: '处理失败',
};

const hasProcessing = computed(() => documents.value.some((item) => item.status === 'processing' || item.status === 'pending'));

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function formatDate(dateStr: string): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  });
}

function statusTagType(status: string): 'warning' | 'primary' | 'success' | 'danger' | 'info' {
  if (status === 'pending') return 'warning';
  if (status === 'processing') return 'primary';
  if (status === 'completed') return 'success';
  if (status === 'failed') return 'danger';
  return 'info';
}

function statusIcon(status: string) {
  if (status === 'completed') return CircleCheckFilled;
  if (status === 'failed') return CircleCloseFilled;
  if (status === 'pending' || status === 'processing') return Loading;
  return WarningFilled;
}

function isStatusLoading(status: string): boolean {
  return status === 'pending' || status === 'processing';
}

function resetUploadForm() {
  selectedFile.value = null;
  uploadForm.value = {
    title: '',
    subject: '',
    grade: '',
  };
  uploadProgress.value = 0;
}

function validateSelectedFile(file: File): boolean {
  const ext = file.name.split('.').pop()?.toLowerCase();
  if (ext !== 'txt' && ext !== 'md') {
    ElMessage.error('仅支持 .txt 和 .md 格式文档');
    return false;
  }

  if (file.size > 5 * 1024 * 1024) {
    ElMessage.error('文档大小不能超过 5MB');
    return false;
  }

  return true;
}

function handleUploadChange(uploadFile: UploadFile, _uploadFiles: UploadFiles) {
  const raw = uploadFile.raw;
  if (!raw) return;

  if (!validateSelectedFile(raw)) {
    return;
  }

  selectedFile.value = raw;
  uploadForm.value.title = raw.name.replace(/\.(txt|md)$/i, '');
}

async function loadDocuments() {
  loading.value = true;
  try {
    const response = await knowledgeApi.listDocuments();
    documents.value = response.data.data || [];
  } catch (error) {
    console.error('Failed to load documents:', error);
    ElMessage.error('文档列表加载失败');
  } finally {
    loading.value = false;
  }
}

async function uploadDocument() {
  if (!selectedFile.value) {
    ElMessage.warning('请先选择文档');
    return;
  }

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

    ElMessage.success('文档上传成功，正在后台处理');
    resetUploadForm();
    await loadDocuments();
  } catch (error) {
    console.error('Upload failed:', error);
    ElMessage.error('上传失败，请重试');
  } finally {
    uploading.value = false;
  }
}

async function deleteDocument(id: string) {
  try {
    await ElMessageBox.confirm('确定要删除该文档吗？相关的知识图谱数据也会被删除。', '删除确认', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消',
    });

    await knowledgeApi.deleteDocument(id);
    ElMessage.success('删除成功');

    if (selectedFile.value) {
      resetUploadForm();
    }

    await loadDocuments();
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Delete failed:', error);
      ElMessage.error('删除失败，请稍后重试');
    }
  }
}

onMounted(() => {
  loadDocuments();

  pollingTimer = window.setInterval(() => {
    if (hasProcessing.value) {
      loadDocuments();
    }
  }, 5000);
});

onUnmounted(() => {
  if (pollingTimer) {
    window.clearInterval(pollingTimer);
    pollingTimer = null;
  }
});
</script>

<template>
  <div class="page-container max-w-5xl mx-auto">
    <div class="page-header">
      <h1 class="page-title">知识库管理</h1>
      <p class="page-subtitle">上传文档，自动构建个人知识图谱</p>
    </div>

    <el-card class="surface-card" shadow="never">
      <template #header>
        <div class="flex items-center justify-between gap-2 flex-wrap">
          <span class="font-semibold">上传文档</span>
          <el-button :icon="RefreshRight" plain @click="loadDocuments">刷新列表</el-button>
        </div>
      </template>

      <el-upload drag :auto-upload="false" :show-file-list="false" accept=".txt,.md" :on-change="handleUploadChange">
        <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
        <div class="el-upload__text">将文件拖到此处，或 <em>点击上传</em></div>
        <template #tip>
          <div class="el-upload__tip">支持 .txt 和 .md 格式，单个文件不超过 5MB</div>
        </template>
      </el-upload>

      <div v-if="selectedFile" class="mt-5 space-y-4">
        <el-alert type="info" :closable="false" show-icon>
          <template #title>
            已选择文件：{{ selectedFile.name }}（{{ formatFileSize(selectedFile.size) }}）
          </template>
        </el-alert>

        <el-form :model="uploadForm" label-position="top">
          <el-row :gutter="16">
            <el-col :xs="24" :md="24">
              <el-form-item label="文档标题（可选）">
                <el-input v-model="uploadForm.title" placeholder="默认使用文件名" />
              </el-form-item>
            </el-col>

            <el-col :xs="24" :md="12">
              <el-form-item label="学科（可选）">
                <el-select v-model="uploadForm.subject" placeholder="不指定" clearable>
                  <el-option v-for="subject in subjectOptions" :key="subject" :label="subject" :value="subject" />
                </el-select>
              </el-form-item>
            </el-col>

            <el-col :xs="24" :md="12">
              <el-form-item label="年级（可选）">
                <el-select v-model="uploadForm.grade" placeholder="不指定" clearable>
                  <el-option v-for="grade in gradeOptions" :key="grade.value" :label="grade.label" :value="grade.value" />
                </el-select>
              </el-form-item>
            </el-col>
          </el-row>
        </el-form>

        <div class="flex items-center gap-2">
          <el-button type="primary" :loading="uploading" @click="uploadDocument">开始上传</el-button>
          <el-button :disabled="uploading" @click="resetUploadForm">清空</el-button>
        </div>

        <el-progress v-if="uploading || uploadProgress > 0" :percentage="uploadProgress" :status="uploadProgress === 100 ? 'success' : undefined" />
      </div>
    </el-card>

    <el-card class="surface-card" shadow="never">
      <template #header>
        <div class="flex items-center justify-between gap-2 flex-wrap">
          <span class="font-semibold">我的知识文档</span>
          <el-tag effect="plain">共 {{ documents.length }} 个</el-tag>
        </div>
      </template>

      <el-skeleton v-if="loading" :rows="6" animated />

      <el-empty v-else-if="documents.length === 0" description="暂无上传的文档" />

      <el-table v-else :data="documents" stripe>
        <el-table-column label="文档" min-width="280">
          <template #default="{ row }">
            <div class="flex items-center gap-2 min-w-0">
              <el-icon class="app-icon-primary"><Document /></el-icon>
              <div class="min-w-0">
                <div class="font-medium app-text-primary line-clamp-1">{{ row.title }}</div>
                <div class="text-xs app-text-muted line-clamp-1">{{ row.fileName }}</div>
              </div>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="大小" width="120">
          <template #default="{ row }">{{ formatFileSize(row.fileSize) }}</template>
        </el-table-column>

        <el-table-column label="状态" width="140" align="center">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" effect="light" round size="small" class="status-pill">
              <span class="status-pill__content">
                <el-icon :class="{ 'is-loading': isStatusLoading(row.status) }">
                  <component :is="statusIcon(row.status)" />
                </el-icon>
                <span>{{ statusText[row.status] || row.status }}</span>
              </span>
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="图谱统计" width="210">
          <template #default="{ row }">
            <div v-if="row.status === 'completed'" class="kg-stats-v2">
              <div class="kg-stat-row">
                <span class="kg-stat-key">实体</span>
                <span class="kg-stat-value">{{ row.entityCount }}</span>
              </div>
              <div class="kg-stat-row">
                <span class="kg-stat-key">关系</span>
                <span class="kg-stat-value">{{ row.relationCount }}</span>
              </div>
            </div>
            <span v-else class="text-sm app-text-muted">-</span>
          </template>
        </el-table-column>

        <el-table-column label="上传日期" width="130">
          <template #default="{ row }">{{ formatDate(row.createdAt) }}</template>
        </el-table-column>

        <el-table-column label="操作" width="90" fixed="right">
          <template #default="{ row }">
            <el-button circle type="danger" plain :icon="Delete" @click="deleteDocument(row.id)" />
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<style scoped>
.status-pill {
  min-width: 84px;
  justify-content: center;
}

.status-pill :deep(.el-tag__content) {
  width: 100%;
  line-height: 1;
}

.status-pill__content {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  font-weight: 500;
}

.status-pill__content :deep(.el-icon) {
  font-size: 13px;
}

.kg-stats-v2 {
  display: inline-flex;
  flex-direction: column;
  gap: 4px;
  min-width: 108px;
}

.kg-stat-row {
  display: inline-flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 3px 8px;
  border-radius: 8px;
  background: color-mix(in srgb, var(--el-fill-color-light) 70%, transparent);
}

.kg-stat-key {
  font-size: 12px;
  line-height: 1;
  color: var(--app-text-muted);
}

.kg-stat-value {
  font-size: 14px;
  font-weight: 700;
  line-height: 1;
  color: var(--app-text-primary);
  font-variant-numeric: tabular-nums;
}
</style>
