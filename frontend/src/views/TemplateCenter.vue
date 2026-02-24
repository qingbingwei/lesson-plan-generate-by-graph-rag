<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { ElMessage, ElMessageBox } from 'element-plus';
import type { CreateLessonTemplateRequest, LessonTemplate } from '@/types';
import { createLessonTemplate, deleteLessonTemplate, listLessonTemplates } from '@/api/template';

const router = useRouter();

const loading = ref(false);
const templates = ref<LessonTemplate[]>([]);
const keyword = ref('');
const activeTab = ref<'all' | 'builtin' | 'mine'>('all');

const createDialogVisible = ref(false);
const createSubmitting = ref(false);
const previewDialogVisible = ref(false);
const previewTemplate = ref<LessonTemplate | null>(null);

const createForm = ref<CreateLessonTemplateRequest>({
  name: '',
  description: '',
  category: '',
  subject: '',
  grade: '',
  duration: 45,
  topic_hint: '',
  style: '',
  requirements: '',
  objectives: '',
  content_outline: '',
  activities: '',
  assessment: '',
  resources: '',
  tags: [],
});

const tagInput = ref('');

const filteredTemplates = computed(() => {
  const search = keyword.value.trim().toLowerCase();
  return templates.value.filter((tpl) => {
    if (activeTab.value === 'builtin' && !tpl.built_in) {
      return false;
    }
    if (activeTab.value === 'mine' && tpl.built_in) {
      return false;
    }
    if (!search) {
      return true;
    }
    const haystack = [
      tpl.name,
      tpl.description || '',
      tpl.subject || '',
      tpl.grade || '',
      (tpl.tags || []).join(' '),
    ].join(' ').toLowerCase();
    return haystack.includes(search);
  });
});

async function loadTemplates() {
  loading.value = true;
  try {
    templates.value = await listLessonTemplates();
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : '加载模板失败');
  } finally {
    loading.value = false;
  }
}

function openCreateDialog() {
  createForm.value = {
    name: '',
    description: '',
    category: '',
    subject: '',
    grade: '',
    duration: 45,
    topic_hint: '',
    style: '',
    requirements: '',
    objectives: '',
    content_outline: '',
    activities: '',
    assessment: '',
    resources: '',
    tags: [],
  };
  tagInput.value = '';
  createDialogVisible.value = true;
}

function addTag() {
  const value = tagInput.value.trim();
  if (!value) {
    return;
  }
  if (!createForm.value.tags) {
    createForm.value.tags = [];
  }
  if (!createForm.value.tags.includes(value)) {
    createForm.value.tags.push(value);
  }
  tagInput.value = '';
}

function removeTag(tag: string) {
  createForm.value.tags = (createForm.value.tags || []).filter((item) => item !== tag);
}

async function submitCreate() {
  if (!createForm.value.name.trim()) {
    ElMessage.warning('请输入模板名称');
    return;
  }

  createSubmitting.value = true;
  try {
    await createLessonTemplate({
      ...createForm.value,
      tags: createForm.value.tags || [],
    });
    ElMessage.success('模板创建成功');
    createDialogVisible.value = false;
    await loadTemplates();
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : '模板创建失败');
  } finally {
    createSubmitting.value = false;
  }
}

function applyTemplate(template: LessonTemplate) {
  router.push(`/generate?templateId=${encodeURIComponent(template.id)}`);
}

function preview(template: LessonTemplate) {
  previewTemplate.value = template;
  previewDialogVisible.value = true;
}

async function removeTemplate(template: LessonTemplate) {
  if (template.built_in) {
    ElMessage.warning('内置模板不支持删除');
    return;
  }

  try {
    await ElMessageBox.confirm(`确定删除模板「${template.name}」吗？`, '删除确认', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消',
    });
  } catch {
    return;
  }

  try {
    await deleteLessonTemplate(template.id);
    ElMessage.success('模板已删除');
    await loadTemplates();
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : '删除模板失败');
  }
}

onMounted(() => {
  loadTemplates();
});
</script>

<template>
  <div class="page-container">
    <div class="page-header flex items-center justify-between gap-3">
      <div>
        <h1 class="page-title">模板库与复用中心</h1>
        <p class="page-subtitle">沉淀高质量模板并一键复用到生成流程</p>
      </div>
      <el-button type="primary" @click="openCreateDialog">新建模板</el-button>
    </div>

    <el-card class="surface-card" shadow="never">
      <div class="flex flex-wrap items-center gap-3">
        <el-radio-group v-model="activeTab">
          <el-radio-button label="all">全部模板</el-radio-button>
          <el-radio-button label="builtin">内置模板</el-radio-button>
          <el-radio-button label="mine">我的模板</el-radio-button>
        </el-radio-group>

        <el-input
          v-model="keyword"
          class="max-w-[360px]"
          clearable
          placeholder="搜索模板名称/学科/标签"
        />
      </div>
    </el-card>

    <el-card class="surface-card" shadow="never">
      <el-skeleton v-if="loading" :rows="6" animated />
      <el-empty v-else-if="filteredTemplates.length === 0" description="暂无匹配模板" />
      <div v-else class="template-grid">
        <el-card
          v-for="tpl in filteredTemplates"
          :key="tpl.id"
          shadow="hover"
          class="template-card"
        >
          <template #header>
            <div class="flex items-start justify-between gap-2">
              <div class="font-semibold">{{ tpl.name }}</div>
              <el-tag size="small" :type="tpl.built_in ? 'success' : 'info'">
                {{ tpl.built_in ? '内置' : '自定义' }}
              </el-tag>
            </div>
          </template>

          <div class="text-sm app-text-muted min-h-[48px]">
            {{ tpl.description || '暂无描述' }}
          </div>

          <div class="mt-3 flex flex-wrap gap-2 text-xs">
            <el-tag size="small" effect="plain">{{ tpl.subject || '未指定学科' }}</el-tag>
            <el-tag size="small" effect="plain">{{ tpl.grade || '未指定年级' }}</el-tag>
            <el-tag size="small" effect="plain">{{ (tpl.duration || 45) + ' 分钟' }}</el-tag>
          </div>

          <div class="mt-3 flex flex-wrap gap-1">
            <el-tag
              v-for="tag in (tpl.tags || []).slice(0, 4)"
              :key="tag"
              size="small"
              type="info"
              effect="plain"
            >
              {{ tag }}
            </el-tag>
          </div>

          <div class="mt-4 flex flex-wrap gap-2">
            <el-button size="small" @click="preview(tpl)">预览</el-button>
            <el-button size="small" type="primary" @click="applyTemplate(tpl)">一键复用</el-button>
            <el-button v-if="!tpl.built_in" size="small" type="danger" plain @click="removeTemplate(tpl)">
              删除
            </el-button>
          </div>
        </el-card>
      </div>
    </el-card>

    <el-dialog v-model="createDialogVisible" width="760px" destroy-on-close>
      <template #header>
        <div class="font-semibold">新建模板</div>
      </template>

      <el-form label-position="top">
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="模板名称">
              <el-input v-model="createForm.name" placeholder="例如：初中语文任务驱动模板" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="分类">
              <el-input v-model="createForm.category" placeholder="例如：阅读课/实验课/复习课" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="12">
          <el-col :span="8">
            <el-form-item label="学科">
              <el-input v-model="createForm.subject" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="年级">
              <el-input v-model="createForm.grade" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="课时">
              <el-input-number v-model="createForm.duration" :min="20" :max="120" :step="5" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="模板描述">
          <el-input v-model="createForm.description" type="textarea" :rows="2" />
        </el-form-item>

        <el-form-item label="课题提示">
          <el-input v-model="createForm.topic_hint" placeholder="例如：分数混合运算的应用" />
        </el-form-item>

        <el-form-item label="风格与要求">
          <el-input v-model="createForm.requirements" type="textarea" :rows="3" placeholder="例如：强调小组合作与形成性评价" />
        </el-form-item>

        <el-form-item label="标签">
          <div class="flex flex-wrap items-center gap-2">
            <el-input
              v-model="tagInput"
              class="w-[240px]"
              placeholder="输入标签后回车"
              @keyup.enter="addTag"
            />
            <el-button @click="addTag">添加</el-button>
            <el-tag
              v-for="tag in createForm.tags || []"
              :key="tag"
              closable
              @close="removeTag(tag)"
            >
              {{ tag }}
            </el-tag>
          </div>
        </el-form-item>
      </el-form>

      <template #footer>
        <div class="flex justify-end gap-2">
          <el-button @click="createDialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="createSubmitting" @click="submitCreate">保存模板</el-button>
        </div>
      </template>
    </el-dialog>

    <el-dialog v-model="previewDialogVisible" width="760px" destroy-on-close>
      <template #header>
        <div class="font-semibold">{{ previewTemplate?.name }}</div>
      </template>
      <template v-if="previewTemplate">
        <el-descriptions :column="2" border size="small">
          <el-descriptions-item label="学科">{{ previewTemplate.subject || '-' }}</el-descriptions-item>
          <el-descriptions-item label="年级">{{ previewTemplate.grade || '-' }}</el-descriptions-item>
          <el-descriptions-item label="课时">{{ previewTemplate.duration || 45 }} 分钟</el-descriptions-item>
          <el-descriptions-item label="使用次数">{{ previewTemplate.usage_count }}</el-descriptions-item>
          <el-descriptions-item label="描述" :span="2">{{ previewTemplate.description || '-' }}</el-descriptions-item>
          <el-descriptions-item label="课题提示" :span="2">{{ previewTemplate.topic_hint || '-' }}</el-descriptions-item>
          <el-descriptions-item label="模板要求" :span="2">{{ previewTemplate.requirements || '-' }}</el-descriptions-item>
        </el-descriptions>
      </template>
      <template #footer>
        <div class="flex justify-between items-center">
          <span class="text-xs app-text-muted">模板预览仅展示核心参数，应用后可在生成页继续修改。</span>
          <el-button type="primary" @click="previewTemplate && applyTemplate(previewTemplate)">立即复用</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.template-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 14px;
}

.template-card {
  border: 1px solid var(--el-border-color-lighter);
}
</style>

