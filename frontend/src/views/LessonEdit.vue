<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useLessonStore } from '@/stores/lesson';
import MarkdownRenderer from '@/components/common/MarkdownRenderer.vue';
import {
  CheckCircleIcon,
  BookOpenIcon,
  AcademicCapIcon,
  ClipboardDocumentCheckIcon,
  FolderOpenIcon,
  EyeIcon,
  PencilSquareIcon,
} from '@heroicons/vue/24/outline';

const route = useRoute();
const router = useRouter();
const lessonStore = useLessonStore();

const lessonId = computed(() => route.params.id as string);
const loading = computed(() => lessonStore.loading);
const saving = ref(false);

// 当前激活的 Tab
const activeTab = ref('basic');

// 各模块的编辑/预览模式
const editModes = ref({
  objectives: true,
  content: true,
  activities: true,
  assessment: true,
  resources: true,
});

// Tab 配置
const tabs = [
  { id: 'basic', name: '基本信息', icon: BookOpenIcon },
  { id: 'objectives', name: '教学目标', icon: CheckCircleIcon },
  { id: 'content', name: '教学内容', icon: AcademicCapIcon },
  { id: 'activities', name: '教学活动', icon: PencilSquareIcon },
  { id: 'assessment', name: '教学评价', icon: ClipboardDocumentCheckIcon },
  { id: 'resources', name: '教学资源', icon: FolderOpenIcon },
];

// 表单数据
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

// 学科选项
const subjects = ['语文', '数学', '英语', '物理', '化学', '生物', '历史', '地理', '政治', '信息技术', '音乐', '美术', '体育'];

// 年级选项
const grades = ['一年级', '二年级', '三年级', '四年级', '五年级', '六年级', '初一', '初二', '初三', '高一', '高二', '高三'];

// 解析 JSON 字符串，提取纯文本
function parseJsonToText(value: any): string {
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

// 加载教案数据
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

// 保存教案
async function handleSave() {
  if (!form.value.title?.trim()) {
    alert('请填写标题');
    activeTab.value = 'basic';
    return;
  }
  if (!form.value.subject?.trim()) {
    alert('请选择学科');
    activeTab.value = 'basic';
    return;
  }
  if (!form.value.grade?.trim()) {
    alert('请选择年级');
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
    
    router.push(`/lessons/${lessonId.value}`);
  } catch (err) {
    console.error('保存失败:', err);
    alert('保存失败，请重试');
  } finally {
    saving.value = false;
  }
}

// 切换编辑/预览模式
function toggleEditMode(field: keyof typeof editModes.value) {
  editModes.value[field] = !editModes.value[field];
}

// 导航到下一个 Tab
function nextTab() {
  const currentIndex = tabs.findIndex(t => t.id === activeTab.value);
  if (currentIndex < tabs.length - 1) {
    activeTab.value = tabs[currentIndex + 1].id;
  }
}

// 导航到上一个 Tab
function prevTab() {
  const currentIndex = tabs.findIndex(t => t.id === activeTab.value);
  if (currentIndex > 0) {
    activeTab.value = tabs[currentIndex - 1].id;
  }
}

// 判断是否是第一个/最后一个 Tab
const isFirstTab = computed(() => activeTab.value === tabs[0].id);
const isLastTab = computed(() => activeTab.value === tabs[tabs.length - 1].id);

onMounted(() => {
  loadLesson();
});
</script>

<template>
  <div class="max-w-5xl mx-auto">
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">编辑教案</h1>
        <p class="mt-1 text-sm text-gray-500">
          {{ form.title || '未命名教案' }}
        </p>
      </div>
      <div class="flex items-center gap-2">
        <RouterLink
          :to="`/lessons/${lessonId}`"
          class="btn-outline"
        >
          取消
        </RouterLink>
        <button
          type="button"
          class="btn-primary"
          :disabled="saving"
          @click="handleSave"
        >
          {{ saving ? '保存中...' : '保存' }}
        </button>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center py-12">
      <div class="loading loading-lg" />
    </div>

    <!-- Editor -->
    <template v-else>
      <div class="flex gap-6">
        <!-- Side Navigation -->
        <div class="w-48 flex-shrink-0">
          <nav class="space-y-1 sticky top-4">
            <button
              v-for="tab in tabs"
              :key="tab.id"
              type="button"
              class="w-full flex items-center gap-3 px-3 py-2 text-sm font-medium rounded-lg transition-colors text-left"
              :class="[
                activeTab === tab.id
                  ? 'bg-primary-50 text-primary-700 border-l-4 border-primary-500'
                  : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
              ]"
              @click="activeTab = tab.id"
            >
              <component :is="tab.icon" class="h-5 w-5" />
              {{ tab.name }}
            </button>
          </nav>
        </div>

        <!-- Content Area -->
        <div class="flex-1 min-w-0">
          <!-- 基本信息 -->
          <div v-show="activeTab === 'basic'" class="card">
            <div class="card-header">
              <h3 class="font-medium flex items-center gap-2">
                <BookOpenIcon class="h-5 w-5 text-gray-500" />
                基本信息
              </h3>
            </div>
            <div class="card-body space-y-4">
              <div>
                <label class="label">标题 <span class="text-red-500">*</span></label>
                <input 
                  v-model="form.title" 
                  type="text" 
                  class="input" 
                  placeholder="请输入教案标题" 
                />
              </div>
              <div class="grid grid-cols-3 gap-4">
                <div>
                  <label class="label">学科 <span class="text-red-500">*</span></label>
                  <select v-model="form.subject" class="input">
                    <option value="">请选择</option>
                    <option v-for="s in subjects" :key="s" :value="s">{{ s }}</option>
                  </select>
                </div>
                <div>
                  <label class="label">年级 <span class="text-red-500">*</span></label>
                  <select v-model="form.grade" class="input">
                    <option value="">请选择</option>
                    <option v-for="g in grades" :key="g" :value="g">{{ g }}</option>
                  </select>
                </div>
                <div>
                  <label class="label">课时（分钟）</label>
                  <input 
                    v-model.number="form.duration" 
                    type="number" 
                    class="input" 
                    min="1" 
                    max="180" 
                  />
                </div>
              </div>
            </div>
          </div>

          <!-- 教学目标 -->
          <div v-show="activeTab === 'objectives'" class="card">
            <div class="card-header flex items-center justify-between">
              <h3 class="font-medium flex items-center gap-2">
                <CheckCircleIcon class="h-5 w-5 text-gray-500" />
                教学目标
              </h3>
              <button
                type="button"
                class="text-sm text-gray-500 hover:text-gray-700 flex items-center gap-1"
                @click="toggleEditMode('objectives')"
              >
                <EyeIcon v-if="editModes.objectives" class="h-4 w-4" />
                <PencilSquareIcon v-else class="h-4 w-4" />
                {{ editModes.objectives ? '预览' : '编辑' }}
              </button>
            </div>
            <div class="card-body">
              <textarea 
                v-if="editModes.objectives"
                v-model="form.objectives" 
                rows="12" 
                class="input font-mono text-sm resize-y min-h-[200px]"
                placeholder="请输入教学目标，支持 Markdown 格式：

## 知识与技能
- 目标1
- 目标2

## 过程与方法
- 目标1

## 情感态度价值观
- 目标1"
              />
              <div v-else class="prose prose-sm max-w-none min-h-[200px] p-4 bg-gray-50 rounded-lg">
                <MarkdownRenderer v-if="form.objectives" :content="form.objectives" />
                <p v-else class="text-gray-400 italic">暂无内容，点击"编辑"添加教学目标</p>
              </div>
            </div>
          </div>

          <!-- 教学内容 -->
          <div v-show="activeTab === 'content'" class="card">
            <div class="card-header flex items-center justify-between">
              <h3 class="font-medium flex items-center gap-2">
                <AcademicCapIcon class="h-5 w-5 text-gray-500" />
                教学内容
              </h3>
              <button
                type="button"
                class="text-sm text-gray-500 hover:text-gray-700 flex items-center gap-1"
                @click="toggleEditMode('content')"
              >
                <EyeIcon v-if="editModes.content" class="h-4 w-4" />
                <PencilSquareIcon v-else class="h-4 w-4" />
                {{ editModes.content ? '预览' : '编辑' }}
              </button>
            </div>
            <div class="card-body">
              <textarea 
                v-if="editModes.content"
                v-model="form.content" 
                rows="16" 
                class="input font-mono text-sm resize-y min-h-[200px]"
                placeholder="请输入教学内容，支持 Markdown 格式：

## 教学重点
1. 重点内容1
2. 重点内容2

## 教学难点
1. 难点内容1

## 知识点
### 知识点1
详细说明..."
              />
              <div v-else class="prose prose-sm max-w-none min-h-[200px] p-4 bg-gray-50 rounded-lg">
                <MarkdownRenderer v-if="form.content" :content="form.content" />
                <p v-else class="text-gray-400 italic">暂无内容，点击"编辑"添加教学内容</p>
              </div>
            </div>
          </div>

          <!-- 教学活动 -->
          <div v-show="activeTab === 'activities'" class="card">
            <div class="card-header flex items-center justify-between">
              <h3 class="font-medium flex items-center gap-2">
                <PencilSquareIcon class="h-5 w-5 text-gray-500" />
                教学活动
              </h3>
              <button
                type="button"
                class="text-sm text-gray-500 hover:text-gray-700 flex items-center gap-1"
                @click="toggleEditMode('activities')"
              >
                <EyeIcon v-if="editModes.activities" class="h-4 w-4" />
                <PencilSquareIcon v-else class="h-4 w-4" />
                {{ editModes.activities ? '预览' : '编辑' }}
              </button>
            </div>
            <div class="card-body">
              <textarea 
                v-if="editModes.activities"
                v-model="form.activities" 
                rows="16" 
                class="input font-mono text-sm resize-y min-h-[200px]"
                placeholder="请输入教学活动设计，支持 Markdown 格式：

## 一、导入环节（5分钟）
活动描述...

## 二、新授环节（20分钟）
### 活动1：xxx
- 活动目标：
- 活动步骤：
- 预期效果：

## 三、练习环节（15分钟）
..."
              />
              <div v-else class="prose prose-sm max-w-none min-h-[200px] p-4 bg-gray-50 rounded-lg">
                <MarkdownRenderer v-if="form.activities" :content="form.activities" />
                <p v-else class="text-gray-400 italic">暂无内容，点击"编辑"添加教学活动</p>
              </div>
            </div>
          </div>

          <!-- 教学评价 -->
          <div v-show="activeTab === 'assessment'" class="card">
            <div class="card-header flex items-center justify-between">
              <h3 class="font-medium flex items-center gap-2">
                <ClipboardDocumentCheckIcon class="h-5 w-5 text-gray-500" />
                教学评价
              </h3>
              <button
                type="button"
                class="text-sm text-gray-500 hover:text-gray-700 flex items-center gap-1"
                @click="toggleEditMode('assessment')"
              >
                <EyeIcon v-if="editModes.assessment" class="h-4 w-4" />
                <PencilSquareIcon v-else class="h-4 w-4" />
                {{ editModes.assessment ? '预览' : '编辑' }}
              </button>
            </div>
            <div class="card-body">
              <textarea 
                v-if="editModes.assessment"
                v-model="form.assessment" 
                rows="12" 
                class="input font-mono text-sm resize-y min-h-[200px]"
                placeholder="请输入教学评价方案，支持 Markdown 格式：

## 评价方式
- 课堂观察
- 作业检查
- 测验评估

## 评价标准
| 等级 | 标准 |
|------|------|
| 优秀 | ... |
| 良好 | ... |"
              />
              <div v-else class="prose prose-sm max-w-none min-h-[200px] p-4 bg-gray-50 rounded-lg">
                <MarkdownRenderer v-if="form.assessment" :content="form.assessment" />
                <p v-else class="text-gray-400 italic">暂无内容，点击"编辑"添加教学评价</p>
              </div>
            </div>
          </div>

          <!-- 教学资源 -->
          <div v-show="activeTab === 'resources'" class="card">
            <div class="card-header flex items-center justify-between">
              <h3 class="font-medium flex items-center gap-2">
                <FolderOpenIcon class="h-5 w-5 text-gray-500" />
                教学资源
              </h3>
              <button
                type="button"
                class="text-sm text-gray-500 hover:text-gray-700 flex items-center gap-1"
                @click="toggleEditMode('resources')"
              >
                <EyeIcon v-if="editModes.resources" class="h-4 w-4" />
                <PencilSquareIcon v-else class="h-4 w-4" />
                {{ editModes.resources ? '预览' : '编辑' }}
              </button>
            </div>
            <div class="card-body">
              <textarea 
                v-if="editModes.resources"
                v-model="form.resources" 
                rows="10" 
                class="input font-mono text-sm resize-y min-h-[200px]"
                placeholder="请输入所需教学资源，支持 Markdown 格式：

## 教具
- 教具1
- 教具2

## 多媒体资源
- PPT课件
- 教学视频

## 参考资料
- 教材：xxx"
              />
              <div v-else class="prose prose-sm max-w-none min-h-[200px] p-4 bg-gray-50 rounded-lg">
                <MarkdownRenderer v-if="form.resources" :content="form.resources" />
                <p v-else class="text-gray-400 italic">暂无内容，点击"编辑"添加教学资源</p>
              </div>
            </div>
          </div>

          <!-- 底部导航 -->
          <div class="flex items-center justify-between mt-4">
            <button
              type="button"
              class="btn-outline"
              :disabled="isFirstTab"
              @click="prevTab"
            >
              ← 上一步
            </button>
            <div class="text-sm text-gray-500">
              {{ tabs.findIndex(t => t.id === activeTab) + 1 }} / {{ tabs.length }}
            </div>
            <button
              v-if="!isLastTab"
              type="button"
              class="btn-primary"
              @click="nextTab"
            >
              下一步 →
            </button>
            <button
              v-else
              type="button"
              class="btn-primary"
              :disabled="saving"
              @click="handleSave"
            >
              {{ saving ? '保存中...' : '保存教案' }}
            </button>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.card-header {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid #e5e7eb;
  background-color: #f9fafb;
}

.card-body {
  padding: 1rem;
}
</style>
