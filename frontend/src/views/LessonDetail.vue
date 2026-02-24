<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useLessonStore } from '@/stores/lesson';
import {
  getLessonVersion,
  getLessonVersions,
  rollbackToVersion,
  getLessonQualityReview,
  getLessonVersionDiff,
  getExportLayouts,
} from '@/api/lesson';
import type { LessonVersion, LessonQualityReview, LessonVersionDiff, ExportLayout } from '@/types';
import MarkdownRenderer from '@/components/common/MarkdownRenderer.vue';
import {
  Edit,
  Delete,
  Download,
  Share,
  Clock,
  Reading,
  Star,
  StarFilled,
  Upload,
  Back,
} from '@element-plus/icons-vue';
import { ElMessage, ElMessageBox } from 'element-plus';

const route = useRoute();
const router = useRouter();
const lessonStore = useLessonStore();

const lessonId = computed(() => route.params.id as string);
const lesson = computed(() => lessonStore.currentLesson);
const loading = computed(() => lessonStore.loading);
const publishing = ref(false);

const showVersionPanel = ref(false);
const versions = ref<LessonVersion[]>([]);
const versionsLoading = ref(false);
const previewLoading = ref(false);
const previewDialogVisible = ref(false);
const previewVersion = ref<LessonVersion | null>(null);
const diffDialogVisible = ref(false);
const diffLoading = ref(false);
const versionDiff = ref<LessonVersionDiff | null>(null);

const showExportMenu = ref(false);
const exporting = ref(false);
const exportLayouts = ref<ExportLayout[]>([]);
const selectedExportLayout = ref('standard');

const qualityLoading = ref(false);
const qualityError = ref<string | null>(null);
const qualityReview = ref<LessonQualityReview | null>(null);

const favorites = ref<string[]>([]);

const isFavorite = computed(() => favorites.value.includes(lessonId.value));
const currentVersion = computed(() => lesson.value?.version ?? 0);

type LessonSnapshotData = {
  title?: string;
  subject?: string;
  grade?: string;
  duration?: number;
  objectives?: string;
  content?: string;
  activities?: string;
  assessment?: string;
  resources?: string;
  status?: string;
  tags?: string;
};

function loadFavorites() {
  const stored = localStorage.getItem('favorites');
  if (stored) {
    favorites.value = JSON.parse(stored);
  }
}

function toggleFavorite() {
  if (isFavorite.value) {
    favorites.value = favorites.value.filter(id => id !== lessonId.value);
    ElMessage.success('已取消收藏');
  } else {
    favorites.value.push(lessonId.value);
    ElMessage.success('收藏成功');
  }
  localStorage.setItem('favorites', JSON.stringify(favorites.value));
}

async function loadVersions() {
  if (versionsLoading.value) return;
  versionsLoading.value = true;
  try {
    versions.value = await getLessonVersions(lessonId.value);
  } catch {
    versions.value = [];
  } finally {
    versionsLoading.value = false;
  }
}

function toggleVersionPanel() {
  showVersionPanel.value = !showVersionPanel.value;
  if (showVersionPanel.value && versions.value.length === 0) {
    loadVersions();
  }
}

async function handleRollback(version: number) {
  try {
    await ElMessageBox.confirm(
      `确定要回滚到版本 ${version} 吗？当前版本将被保存为历史记录。`,
      '回滚确认',
      { type: 'warning' }
    );

    await rollbackToVersion(lessonId.value, version);
    await lessonStore.fetchLesson(lessonId.value);
    await loadVersions();
    previewDialogVisible.value = false;
    ElMessage.success('回滚成功');
  } catch (err) {
    if (err === 'cancel' || err === 'close') {
      return;
    }

    const message =
      (err as any)?.response?.data?.message ||
      (err instanceof Error ? err.message : '回滚失败，请查看后端日志');
    ElMessage.error(message);
  }
}

function getVersionTime(version: LessonVersion): string {
  const raw = (version as any).createdAt ?? (version as any).created_at;
  if (!raw) return '-';
  const date = new Date(raw);
  if (Number.isNaN(date.getTime())) return '-';
  return date.toLocaleString('zh-CN');
}

function getVersionSummary(version: LessonVersion): string {
  const summary = version.changeLog || (version as any).change_log;
  if (typeof summary === 'string' && summary.trim().length > 0) {
    return summary;
  }
  return '自动保存的历史快照';
}

function parseSnapshot(content: string): LessonSnapshotData {
  try {
    const parsed = JSON.parse(content);
    if (parsed && typeof parsed === 'object') {
      return parsed as LessonSnapshotData;
    }
  } catch {
    // ignore
  }
  return {};
}

const previewSnapshot = computed<LessonSnapshotData>(() => {
  if (!previewVersion.value?.content || typeof previewVersion.value.content !== 'string') {
    return {};
  }
  return parseSnapshot(previewVersion.value.content);
});

function previewFieldText(value: unknown): string {
  if (typeof value !== 'string') {
    return '';
  }
  return parseJsonText(value);
}

const previewTagList = computed<string[]>(() => {
  const tagsRaw = previewSnapshot.value.tags;
  if (!tagsRaw) return [];

  try {
    const parsed = JSON.parse(tagsRaw);
    if (Array.isArray(parsed)) {
      return parsed
        .map(tag => (typeof tag === 'string' ? tag : String(tag)))
        .filter(tag => tag.trim().length > 0);
    }
  } catch {
    // ignore
  }

  return [];
});

async function handlePreview(version: LessonVersion) {
  previewLoading.value = true;
  previewDialogVisible.value = true;

  try {
    previewVersion.value = await getLessonVersion(lessonId.value, version.version);
  } catch (err) {
    previewVersion.value = version;
    ElMessage.warning(err instanceof Error ? err.message : '读取版本详情失败，已展示列表中的版本信息');
  } finally {
    previewLoading.value = false;
  }
}

async function handleCompare(version: number) {
  diffLoading.value = true;
  diffDialogVisible.value = true;
  versionDiff.value = null;

  try {
    versionDiff.value = await getLessonVersionDiff(lessonId.value, version, 'current');
  } catch (err) {
    diffDialogVisible.value = false;
    ElMessage.error(err instanceof Error ? err.message : '版本对比失败');
  } finally {
    diffLoading.value = false;
  }
}

async function loadExportLayouts() {
  try {
    exportLayouts.value = await getExportLayouts();
    if (!exportLayouts.value.some(item => item.id === selectedExportLayout.value)) {
      selectedExportLayout.value = exportLayouts.value[0]?.id || 'standard';
    }
  } catch {
    exportLayouts.value = [
      { id: 'standard', name: '标准模板', description: '通用教学文档结构' },
      { id: 'compact', name: '紧凑模板', description: '关键要点优先展示' },
      { id: 'research', name: '教研模板', description: '适合教研交流场景' },
    ];
  }
}

async function runQualityReview() {
  qualityLoading.value = true;
  qualityError.value = null;

  try {
    qualityReview.value = await getLessonQualityReview(lessonId.value);
  } catch (err) {
    qualityError.value = err instanceof Error ? err.message : '质量审查失败';
    ElMessage.error(qualityError.value);
  } finally {
    qualityLoading.value = false;
  }
}

function parseDownloadFilename(contentDisposition: string | null, fallback: string): string {
  if (!contentDisposition) {
    return fallback;
  }

  const utf8Match = contentDisposition.match(/filename\*=UTF-8''([^;]+)/i);
  if (utf8Match?.[1]) {
    try {
      return decodeURIComponent(utf8Match[1]);
    } catch {
      // ignore decode error
    }
  }

  const plainMatch = contentDisposition.match(/filename="?([^";]+)"?/i);
  if (plainMatch?.[1]) {
    return plainMatch[1];
  }

  return fallback;
}

function handleClickOutside(event: MouseEvent) {
  const target = event.target as HTMLElement;
  if (!target.closest('.export-menu')) {
    showExportMenu.value = false;
  }
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm('确定要删除这个教案吗？', '删除确认', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消',
    });

    await lessonStore.deleteLesson(lessonId.value);
    ElMessage.success('删除成功');
    router.push('/lessons');
  } catch {
    // cancel or fail
  }
}

async function handlePublish() {
  publishing.value = true;
  try {
    await lessonStore.publishLesson(lessonId.value);
    ElMessage.success('发布成功');
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : '发布失败');
  } finally {
    publishing.value = false;
  }
}

async function handleExport(format: 'md' | 'pdf' | 'docx') {
  if (!lesson.value) return;

  showExportMenu.value = false;
  exporting.value = true;

  try {
    const params = new URLSearchParams({
      format,
      layout: selectedExportLayout.value,
    });
    const response = await fetch(`/api/v1/lessons/${lessonId.value}/export?${params.toString()}`, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem('auth') ? JSON.parse(localStorage.getItem('auth')!).token : ''}`,
      },
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || '导出失败');
    }

    const contentDisposition = response.headers.get('Content-Disposition');
    const filename = parseDownloadFilename(contentDisposition, `${lesson.value.title}.${format}`);

    const blob = await response.blob();
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);

    ElMessage.success('导出成功');
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : '导出失败，请重试');
  } finally {
    exporting.value = false;
  }
}

function parseJsonText(value: any): string {
  if (!value) return '';
  if (typeof value !== 'string') return String(value);

  try {
    const parsed = JSON.parse(value);
    if (typeof parsed === 'string') return parsed;
    if (parsed.text) return parsed.text;
    return JSON.stringify(parsed, null, 2);
  } catch {
    return value;
  }
}

onMounted(() => {
  loadFavorites();
  lessonStore.fetchLesson(lessonId.value);
  loadExportLayouts();
  document.addEventListener('click', handleClickOutside);
});

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside);
});
</script>

<template>
  <div class="page-container max-w-5xl mx-auto">
    <el-skeleton v-if="loading" :rows="8" animated />

    <template v-else-if="lesson">
      <el-card class="surface-card" shadow="never">
        <div class="flex flex-col gap-4">
          <div class="flex flex-wrap items-center gap-2">
            <el-tag>{{ lesson.subject }}</el-tag>
            <el-tag>{{ lesson.grade }}</el-tag>
            <el-tag :type="lesson.status === 'published' ? 'success' : 'info'">
              {{ lesson.status === 'published' ? '已发布' : '草稿' }}
            </el-tag>
          </div>

          <div class="flex flex-col xl:flex-row xl:items-start xl:justify-between gap-4">
            <div class="min-w-0">
              <h1 class="page-title">{{ lesson.title }}</h1>
              <div class="mt-2 text-sm app-text-muted flex items-center gap-4">
                <span class="inline-flex items-center gap-1"><el-icon><Clock /></el-icon>{{ lesson.duration }}分钟</span>
                <span class="inline-flex items-center gap-1"><el-icon><Reading /></el-icon>版本 {{ lesson.version }}</span>
              </div>
            </div>

            <div class="action-panel">
              <div class="action-row action-row--primary">
                <el-button :icon="isFavorite ? StarFilled : Star" @click="toggleFavorite">
                  {{ isFavorite ? '已收藏' : '收藏' }}
                </el-button>
                <el-button :icon="Edit" @click="router.push(`/lessons/${lesson.id}/edit`)">编辑</el-button>
                <el-button :loading="qualityLoading" @click="runQualityReview">质量审查</el-button>
                <el-button v-if="lesson.status === 'draft'" type="success" :icon="Upload" :loading="publishing" @click="handlePublish">
                  发布
                </el-button>
              </div>

              <div class="action-row action-row--secondary">
                <div class="export-menu export-controls">
                  <el-select v-model="selectedExportLayout" class="export-layout-select" size="default">
                    <el-option
                      v-for="item in exportLayouts"
                      :key="item.id"
                      :label="item.name"
                      :value="item.id"
                    />
                  </el-select>

                  <el-dropdown trigger="click" @visible-change="(v:boolean)=>showExportMenu=v">
                    <el-button :icon="Download" :loading="exporting">导出</el-button>
                    <template #dropdown>
                      <el-dropdown-menu>
                        <el-dropdown-item @click="handleExport('md')">Markdown (.md)</el-dropdown-item>
                        <el-dropdown-item @click="handleExport('docx')">Word (.docx)</el-dropdown-item>
                        <el-dropdown-item @click="handleExport('pdf')">PDF (.pdf)</el-dropdown-item>
                      </el-dropdown-menu>
                    </template>
                  </el-dropdown>
                </div>

                <el-button :icon="Clock" @click="toggleVersionPanel">版本历史</el-button>
                <el-button :icon="Share">分享</el-button>
                <el-button type="danger" :icon="Delete" @click="handleDelete">删除</el-button>
              </div>
            </div>
          </div>
        </div>
      </el-card>

      <el-card v-if="showVersionPanel" class="surface-card" shadow="never">
        <template #header>
          <div class="flex items-center justify-between gap-3">
            <span class="font-semibold">版本历史</span>
            <el-tag size="small" type="info">当前版本：v{{ currentVersion }}</el-tag>
          </div>
        </template>

        <el-skeleton v-if="versionsLoading" :rows="4" animated />
        <el-empty v-else-if="versions.length === 0" description="暂无版本记录（编辑后将自动保存版本）" />

        <div v-else class="version-list">
          <el-card
            v-for="v in versions"
            :key="v.id"
            class="version-item"
            shadow="never"
          >
            <div class="version-item__header">
              <div class="version-item__meta">
                <div class="font-semibold">版本 v{{ v.version }}</div>
                <div class="text-xs app-text-muted">{{ getVersionTime(v) }}</div>
              </div>
              <el-tag v-if="v.version === currentVersion" size="small" type="success">当前</el-tag>
            </div>

            <div class="text-sm app-text-muted">{{ getVersionSummary(v) }}</div>

            <div class="version-item__actions">
              <el-button size="small" @click="handlePreview(v)">查看版本</el-button>
              <el-button size="small" type="primary" @click="handleCompare(v.version)">对比当前</el-button>
              <el-button
                size="small"
                type="warning"
                :disabled="v.version === currentVersion"
                @click="handleRollback(v.version)"
              >
                回滚到此版本
              </el-button>
            </div>
          </el-card>
        </div>
      </el-card>

      <el-card v-if="qualityReview" class="surface-card" shadow="never">
        <template #header>
          <div class="flex items-center justify-between gap-3">
            <span class="font-semibold">质量审查结果</span>
            <div class="flex items-center gap-2">
              <el-tag :type="qualityReview.grade === 'A' ? 'success' : qualityReview.grade === 'B' ? 'warning' : 'danger'">
                等级 {{ qualityReview.grade }}
              </el-tag>
              <el-tag size="small" effect="plain">
                {{ qualityReview.total_score }} / {{ qualityReview.max_score }}
              </el-tag>
              <el-tag v-if="qualityReview.auto_approved" type="success" size="small">建议可发布</el-tag>
            </div>
          </div>
        </template>

        <el-alert
          v-if="qualityError"
          :title="qualityError"
          type="warning"
          show-icon
          class="mb-3"
        />

        <el-table :data="qualityReview.dimensions" size="small" border>
          <el-table-column prop="name" label="维度" min-width="120" />
          <el-table-column label="得分" width="120">
            <template #default="{ row }">
              {{ row.score }} / {{ row.max_score }}
            </template>
          </el-table-column>
          <el-table-column prop="comment" label="说明" min-width="220" />
        </el-table>

        <div class="mt-4 grid md:grid-cols-2 gap-3">
          <el-card shadow="never">
            <template #header><span class="text-sm font-semibold">发现问题</span></template>
            <ul class="list-disc pl-5 text-sm app-text-muted">
              <li v-for="(issue, idx) in qualityReview.issues" :key="`issue-${idx}`">{{ issue }}</li>
            </ul>
          </el-card>
          <el-card shadow="never">
            <template #header><span class="text-sm font-semibold">优化建议</span></template>
            <ul class="list-disc pl-5 text-sm app-text-muted">
              <li v-for="(item, idx) in qualityReview.suggestions" :key="`suggestion-${idx}`">{{ item }}</li>
            </ul>
          </el-card>
        </div>
      </el-card>

      <el-dialog v-model="diffDialogVisible" width="1000px" destroy-on-close>
        <template #header>
          <div class="font-semibold">
            版本差异对比
            <span v-if="versionDiff" class="text-xs app-text-muted ml-2">
              {{ versionDiff.from_version }} -> {{ versionDiff.to_version }}（变更字段 {{ versionDiff.changed_fields }}）
            </span>
          </div>
        </template>

        <el-skeleton v-if="diffLoading" :rows="6" animated />
        <el-empty v-else-if="!versionDiff" description="暂无差异数据" />
        <div v-else class="space-y-3 max-h-[560px] overflow-y-auto pr-1">
          <el-card
            v-for="field in versionDiff.fields"
            :key="field.field"
            shadow="never"
            class="border border-[var(--el-border-color-light)]"
          >
            <template #header>
              <div class="flex items-center justify-between">
                <span class="font-medium">{{ field.label }}</span>
                <el-tag :type="field.changed ? 'warning' : 'success'" size="small">
                  {{ field.changed ? '已变更' : '无变化' }}
                </el-tag>
              </div>
            </template>

            <div v-if="field.changed" class="grid md:grid-cols-2 gap-3">
              <div>
                <div class="text-xs app-text-muted mb-1">旧版本</div>
                <pre class="diff-box">{{ field.before || '-' }}</pre>
              </div>
              <div>
                <div class="text-xs app-text-muted mb-1">当前版本</div>
                <pre class="diff-box">{{ field.after || '-' }}</pre>
              </div>
            </div>
            <div v-else class="text-sm app-text-muted">该字段未变化</div>
          </el-card>
        </div>
      </el-dialog>

      <el-dialog
        v-model="previewDialogVisible"
        width="900px"
        destroy-on-close
        class="version-preview-dialog"
      >
        <template #header>
          <div class="font-semibold">
            历史版本预览
            <span v-if="previewVersion">- v{{ previewVersion.version }}</span>
          </div>
        </template>

        <el-skeleton v-if="previewLoading" :rows="6" animated />

        <template v-else>
          <el-descriptions :column="3" border size="small" class="mb-4">
            <el-descriptions-item label="标题">
              {{ previewSnapshot.title || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="学科">
              {{ previewSnapshot.subject || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="年级">
              {{ previewSnapshot.grade || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="时长">
              {{ previewSnapshot.duration ? `${previewSnapshot.duration} 分钟` : '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="状态">
              {{ previewSnapshot.status || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="标签">
              <div class="flex flex-wrap gap-1">
                <el-tag v-for="tag in previewTagList" :key="tag" size="small">{{ tag }}</el-tag>
                <span v-if="previewTagList.length === 0">-</span>
              </div>
            </el-descriptions-item>
          </el-descriptions>

          <el-collapse>
            <el-collapse-item v-if="previewFieldText(previewSnapshot.objectives)" title="教学目标" name="objectives">
              <div class="markdown-prose">
                <MarkdownRenderer :content="previewFieldText(previewSnapshot.objectives)" />
              </div>
            </el-collapse-item>

            <el-collapse-item v-if="previewFieldText(previewSnapshot.content)" title="教学内容" name="content">
              <div class="markdown-prose">
                <MarkdownRenderer :content="previewFieldText(previewSnapshot.content)" />
              </div>
            </el-collapse-item>

            <el-collapse-item v-if="previewFieldText(previewSnapshot.activities)" title="教学活动" name="activities">
              <div class="markdown-prose">
                <MarkdownRenderer :content="previewFieldText(previewSnapshot.activities)" />
              </div>
            </el-collapse-item>

            <el-collapse-item v-if="previewFieldText(previewSnapshot.assessment)" title="教学评价" name="assessment">
              <div class="markdown-prose">
                <MarkdownRenderer :content="previewFieldText(previewSnapshot.assessment)" />
              </div>
            </el-collapse-item>

            <el-collapse-item v-if="previewFieldText(previewSnapshot.resources)" title="教学资源" name="resources">
              <div class="markdown-prose">
                <MarkdownRenderer :content="previewFieldText(previewSnapshot.resources)" />
              </div>
            </el-collapse-item>
          </el-collapse>
        </template>

        <template #footer>
          <div class="flex items-center justify-between">
            <span class="text-xs app-text-muted">可先查看版本内容，再决定是否回滚</span>
            <div class="flex gap-2">
              <el-button @click="previewDialogVisible = false">关闭</el-button>
              <el-button
                v-if="previewVersion"
                type="warning"
                :disabled="previewVersion.version === currentVersion"
                @click="handleRollback(previewVersion.version)"
              >
                回滚到该版本
              </el-button>
            </div>
          </div>
        </template>
      </el-dialog>

      <el-card v-if="(lesson as any).objectives" class="surface-card" shadow="never">
        <template #header><span class="font-semibold">教学目标</span></template>
        <div class="markdown-prose"><MarkdownRenderer :content="parseJsonText((lesson as any).objectives)" /></div>
      </el-card>

      <el-card v-if="(lesson as any).content" class="surface-card" shadow="never">
        <template #header><span class="font-semibold">教学内容</span></template>
        <div class="markdown-prose"><MarkdownRenderer :content="parseJsonText((lesson as any).content)" /></div>
      </el-card>

      <el-card v-if="(lesson as any).activities" class="surface-card" shadow="never">
        <template #header><span class="font-semibold">教学活动</span></template>
        <div class="markdown-prose"><MarkdownRenderer :content="(lesson as any).activities" /></div>
      </el-card>

      <el-card v-if="(lesson as any).assessment" class="surface-card" shadow="never">
        <template #header><span class="font-semibold">教学评价</span></template>
        <div class="markdown-prose"><MarkdownRenderer :content="(lesson as any).assessment" /></div>
      </el-card>

      <el-card v-if="(lesson as any).resources" class="surface-card" shadow="never">
        <template #header><span class="font-semibold">教学资源</span></template>
        <div class="markdown-prose"><MarkdownRenderer :content="(lesson as any).resources" /></div>
      </el-card>
    </template>

    <el-empty v-else description="教案不存在">
      <el-button type="primary" :icon="Back" @click="router.push('/lessons')">返回列表</el-button>
    </el-empty>
  </div>
</template>

<style scoped>
.version-list {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 12px;
}

.action-panel {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 10px;
}

.action-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: flex-end;
}

.export-controls {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  border: 1px solid var(--el-border-color-light);
  border-radius: 12px;
  background: var(--el-fill-color-lighter);
}

.export-layout-select {
  width: 168px;
}

.version-item {
  border: 1px solid var(--el-border-color-lighter);
}

.version-item__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
  margin-bottom: 8px;
}

.version-item__meta {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.version-item__actions {
  margin-top: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.version-preview-dialog :deep(.el-dialog__body) {
  padding-top: 8px;
}

.diff-box {
  margin: 0;
  min-height: 120px;
  max-height: 260px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 12px;
  line-height: 1.45;
  border: 1px solid var(--el-border-color);
  border-radius: 6px;
  padding: 10px;
  background: var(--el-fill-color-light);
}

@media (max-width: 1280px) {
  .action-panel {
    align-items: stretch;
  }

  .action-row {
    justify-content: flex-start;
  }

  .export-controls {
    width: 100%;
    justify-content: space-between;
  }

  .export-layout-select {
    min-width: 150px;
    width: 60%;
  }
}
</style>
