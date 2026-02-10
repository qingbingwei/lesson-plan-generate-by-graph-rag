<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';
import { useLessonStore } from '@/stores/lesson';
import MarkdownRenderer from '@/components/common/MarkdownRenderer.vue';

type MarkdownSectionKey = 'objectives' | 'content' | 'activities' | 'assessment' | 'resources';

const route = useRoute();
const router = useRouter();
const lessonStore = useLessonStore();

const lessonId = computed(() => route.params.id as string);
const loading = computed(() => lessonStore.loading);
const saving = ref(false);
const activeTab = ref('basic');

const editModes = ref<Record<MarkdownSectionKey, boolean>>({
  objectives: true,
  content: true,
  activities: true,
  assessment: true,
  resources: true,
});

const tabs = [
  { id: 'basic', name: '基本信息' },
  { id: 'objectives', name: '教学目标' },
  { id: 'content', name: '教学内容' },
  { id: 'activities', name: '教学活动' },
  { id: 'assessment', name: '教学评价' },
  { id: 'resources', name: '教学资源' },
];

const form = ref({
  title: '',
  subject: '',
  grade: '',
  duration: 45,
  objectives: '',
  content: '',
  activities: '',
  assessment: '',
  resources: '',
});

const subjects = [
  '语文',
  '数学',
  '英语',
  '物理',
  '化学',
  '生物',
  '历史',
  '地理',
  '政治',
  '信息技术',
  '音乐',
  '美术',
  '体育',
];

const grades = [
  '一年级',
  '二年级',
  '三年级',
  '四年级',
  '五年级',
  '六年级',
  '初一',
  '初二',
  '初三',
  '高一',
  '高二',
  '高三',
];

const markdownPlaceholders: Record<MarkdownSectionKey, string> = {
  objectives: `请输入教学目标，支持 Markdown 格式：

## 知识与技能
- 目标1
- 目标2

## 过程与方法
- 目标1

## 情感态度价值观
- 目标1`,
  content: `请输入教学内容，支持 Markdown 格式：

## 教学重点
1. 重点内容1
2. 重点内容2

## 教学难点
1. 难点内容1`,
  activities: `请输入教学活动设计，支持 Markdown 格式：

## 一、导入环节（5分钟）
活动描述...

## 二、新授环节（20分钟）
### 活动1：xxx
- 活动目标：`,
  assessment: `请输入教学评价方案，支持 Markdown 格式：

## 评价方式
- 课堂观察
- 作业检查
- 测验评估

## 评价标准
| 等级 | 标准 |
|------|------|`,
  resources: `请输入所需教学资源，支持 Markdown 格式：

## 教具
- 教具1
- 教具2

## 多媒体资源
- PPT课件
- 教学视频`,
};

function parseJsonToText(value: unknown): string {
  if (!value) return '';
  if (typeof value !== 'string') return String(value);

  try {
    const parsed = JSON.parse(value);
    if (typeof parsed === 'string') {
      return parsed;
    }
    if (parsed.text) {
      return parsed.text;
    }
    return JSON.stringify(parsed, null, 2);
  } catch {
    return value;
  }
}

async function loadLesson() {
  await lessonStore.fetchLesson(lessonId.value);

  if (lessonStore.currentLesson) {
    const lesson = lessonStore.currentLesson as any;
    form.value = {
      title: lesson.title || '',
      subject: lesson.subject || '',
      grade: lesson.grade || '',
      duration: lesson.duration || 45,
      objectives: parseJsonToText(lesson.objectives),
      content: parseJsonToText(lesson.content),
      activities: parseJsonToText(lesson.activities),
      assessment: parseJsonToText(lesson.assessment),
      resources: parseJsonToText(lesson.resources),
    };
  }
}

async function handleSave() {
  if (!form.value.title?.trim()) {
    ElMessage.warning('请填写标题');
    activeTab.value = 'basic';
    return;
  }

  if (!form.value.subject?.trim()) {
    ElMessage.warning('请选择学科');
    activeTab.value = 'basic';
    return;
  }

  if (!form.value.grade?.trim()) {
    ElMessage.warning('请选择年级');
    activeTab.value = 'basic';
    return;
  }

  saving.value = true;

  try {
    await lessonStore.updateLesson(lessonId.value, {
      title: form.value.title,
      subject: form.value.subject,
      grade: form.value.grade,
      duration: form.value.duration,
      objectives: form.value.objectives,
      content: form.value.content,
      activities: form.value.activities,
      assessment: form.value.assessment,
      resources: form.value.resources,
    } as any);

    ElMessage.success('保存成功');
    router.push(`/lessons/${lessonId.value}`);
  } catch (err) {
    console.error('保存失败:', err);
    ElMessage.error('保存失败，请重试');
  } finally {
    saving.value = false;
  }
}

function toggleEditMode(field: MarkdownSectionKey) {
  editModes.value[field] = !editModes.value[field];
}

function nextTab() {
  const currentIndex = tabs.findIndex((tab) => tab.id === activeTab.value);
  if (currentIndex < tabs.length - 1) {
    activeTab.value = tabs[currentIndex + 1].id;
  }
}

function prevTab() {
  const currentIndex = tabs.findIndex((tab) => tab.id === activeTab.value);
  if (currentIndex > 0) {
    activeTab.value = tabs[currentIndex - 1].id;
  }
}

const isFirstTab = computed(() => activeTab.value === tabs[0].id);
const isLastTab = computed(() => activeTab.value === tabs[tabs.length - 1].id);

onMounted(() => {
  loadLesson();
});
</script>

<template>
  <div class="page-container max-w-6xl mx-auto">
    <div class="page-header flex flex-col md:flex-row md:items-start md:justify-between gap-3">
      <div>
        <h1 class="page-title">编辑教案</h1>
        <p class="page-subtitle">{{ form.title || '未命名教案' }}</p>
      </div>
      <div class="flex items-center gap-2">
        <el-button @click="router.push(`/lessons/${lessonId}`)">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </div>
    </div>

    <el-skeleton v-if="loading" :rows="10" animated />

    <el-card v-else class="surface-card" shadow="never">
      <el-tabs v-model="activeTab" tab-position="left" class="lesson-edit-tabs">
        <el-tab-pane label="基本信息" name="basic">
          <el-form :model="form" label-position="top" class="max-w-3xl">
            <el-form-item label="标题" required>
              <el-input v-model="form.title" placeholder="请输入教案标题" />
            </el-form-item>

            <el-row :gutter="16">
              <el-col :xs="24" :sm="8">
                <el-form-item label="学科" required>
                  <el-select v-model="form.subject" placeholder="请选择学科">
                    <el-option v-for="subject in subjects" :key="subject" :label="subject" :value="subject" />
                  </el-select>
                </el-form-item>
              </el-col>

              <el-col :xs="24" :sm="8">
                <el-form-item label="年级" required>
                  <el-select v-model="form.grade" placeholder="请选择年级">
                    <el-option v-for="grade in grades" :key="grade" :label="grade" :value="grade" />
                  </el-select>
                </el-form-item>
              </el-col>

              <el-col :xs="24" :sm="8">
                <el-form-item label="课时（分钟）">
                  <el-input-number v-model="form.duration" :min="1" :max="180" />
                </el-form-item>
              </el-col>
            </el-row>
          </el-form>
        </el-tab-pane>

        <el-tab-pane label="教学目标" name="objectives">
          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-semibold app-text-primary">教学目标</h3>
              <el-button text type="primary" @click="toggleEditMode('objectives')">
                {{ editModes.objectives ? '预览' : '编辑' }}
              </el-button>
            </div>

            <el-input
              v-if="editModes.objectives"
              v-model="form.objectives"
              type="textarea"
              :rows="14"
              resize="vertical"
              :placeholder="markdownPlaceholders.objectives"
            />

            <el-card v-else class="surface-card" shadow="never">
              <div v-if="form.objectives" class="markdown-prose">
                <MarkdownRenderer :content="form.objectives" />
              </div>
              <el-empty v-else description="暂无内容，点击“编辑”开始编写" />
            </el-card>
          </div>
        </el-tab-pane>

        <el-tab-pane label="教学内容" name="content">
          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-semibold app-text-primary">教学内容</h3>
              <el-button text type="primary" @click="toggleEditMode('content')">
                {{ editModes.content ? '预览' : '编辑' }}
              </el-button>
            </div>

            <el-input
              v-if="editModes.content"
              v-model="form.content"
              type="textarea"
              :rows="14"
              resize="vertical"
              :placeholder="markdownPlaceholders.content"
            />

            <el-card v-else class="surface-card" shadow="never">
              <div v-if="form.content" class="markdown-prose">
                <MarkdownRenderer :content="form.content" />
              </div>
              <el-empty v-else description="暂无内容，点击“编辑”开始编写" />
            </el-card>
          </div>
        </el-tab-pane>

        <el-tab-pane label="教学活动" name="activities">
          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-semibold app-text-primary">教学活动</h3>
              <el-button text type="primary" @click="toggleEditMode('activities')">
                {{ editModes.activities ? '预览' : '编辑' }}
              </el-button>
            </div>

            <el-input
              v-if="editModes.activities"
              v-model="form.activities"
              type="textarea"
              :rows="14"
              resize="vertical"
              :placeholder="markdownPlaceholders.activities"
            />

            <el-card v-else class="surface-card" shadow="never">
              <div v-if="form.activities" class="markdown-prose">
                <MarkdownRenderer :content="form.activities" />
              </div>
              <el-empty v-else description="暂无内容，点击“编辑”开始编写" />
            </el-card>
          </div>
        </el-tab-pane>

        <el-tab-pane label="教学评价" name="assessment">
          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-semibold app-text-primary">教学评价</h3>
              <el-button text type="primary" @click="toggleEditMode('assessment')">
                {{ editModes.assessment ? '预览' : '编辑' }}
              </el-button>
            </div>

            <el-input
              v-if="editModes.assessment"
              v-model="form.assessment"
              type="textarea"
              :rows="14"
              resize="vertical"
              :placeholder="markdownPlaceholders.assessment"
            />

            <el-card v-else class="surface-card" shadow="never">
              <div v-if="form.assessment" class="markdown-prose">
                <MarkdownRenderer :content="form.assessment" />
              </div>
              <el-empty v-else description="暂无内容，点击“编辑”开始编写" />
            </el-card>
          </div>
        </el-tab-pane>

        <el-tab-pane label="教学资源" name="resources">
          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-semibold app-text-primary">教学资源</h3>
              <el-button text type="primary" @click="toggleEditMode('resources')">
                {{ editModes.resources ? '预览' : '编辑' }}
              </el-button>
            </div>

            <el-input
              v-if="editModes.resources"
              v-model="form.resources"
              type="textarea"
              :rows="14"
              resize="vertical"
              :placeholder="markdownPlaceholders.resources"
            />

            <el-card v-else class="surface-card" shadow="never">
              <div v-if="form.resources" class="markdown-prose">
                <MarkdownRenderer :content="form.resources" />
              </div>
              <el-empty v-else description="暂无内容，点击“编辑”开始编写" />
            </el-card>
          </div>
        </el-tab-pane>
      </el-tabs>

      <div class="flex items-center justify-between app-divider-top pt-4">
        <el-button :disabled="isFirstTab" @click="prevTab">上一步</el-button>
        <div class="text-sm app-text-muted">
          {{ tabs.findIndex((tab) => tab.id === activeTab) + 1 }} / {{ tabs.length }}
        </div>
        <el-button v-if="!isLastTab" type="primary" @click="nextTab">下一步</el-button>
        <el-button v-else type="primary" :loading="saving" @click="handleSave">保存教案</el-button>
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.lesson-edit-tabs :deep(.el-tabs__header) {
  margin-right: 20px;
}

.lesson-edit-tabs :deep(.el-tabs__content) {
  min-height: 520px;
}
</style>
