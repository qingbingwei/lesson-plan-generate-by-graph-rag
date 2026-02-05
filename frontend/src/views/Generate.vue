<script setup lang="ts">
import { ref, computed } from 'vue';
import { useRouter } from 'vue-router';
import { useGenerationStore } from '@/stores/generation';
import { useLessonStore } from '@/stores/lesson';
import MarkdownRenderer from '@/components/common/MarkdownRenderer.vue';
import {
  SparklesIcon,
  CheckCircleIcon,
  XCircleIcon,
  ClockIcon,
  ArrowPathIcon,
} from '@heroicons/vue/24/outline';

// 保存状态
const isSaving = ref(false);
const saveError = ref<string | null>(null);

const router = useRouter();
const generationStore = useGenerationStore();
const lessonStore = useLessonStore();

// 表单
const form = ref({
  subject: '',
  grade: '',
  topic: '',
  duration: 45,
  style: '',
  requirements: '',
});

// 选项
const subjects = [
  '语文', '数学', '英语', '物理', '化学', '生物',
  '历史', '地理', '政治', '科学', '信息技术',
  '音乐', '美术', '体育',
];

const grades = [
  '一年级', '二年级', '三年级', '四年级', '五年级', '六年级',
  '七年级', '八年级', '九年级',
  '高一', '高二', '高三',
];

const styles = [
  { value: '', label: '默认风格' },
  { value: 'interactive', label: '互动探究' },
  { value: 'lecture', label: '讲授式' },
  { value: 'project', label: '项目式学习' },
  { value: 'flipped', label: '翻转课堂' },
];

// 状态
const isGenerating = computed(() => generationStore.isGenerating);
const progress = computed(() => generationStore.progress);
const generatedLesson = computed(() => generationStore.generatedLesson);
const error = computed(() => generationStore.error);

// 验证
const isValid = computed(() => {
  return form.value.subject && form.value.grade && form.value.topic && form.value.duration > 0;
});

// 生成教案
async function handleGenerate() {
  if (!isValid.value) return;
  
  generationStore.streamGenerateLesson({
    subject: form.value.subject,
    grade: form.value.grade,
    topic: form.value.topic,
    duration: form.value.duration,
    style: form.value.style || undefined,
    requirements: form.value.requirements || undefined,
  });
}

// 取消生成
function handleCancel() {
  generationStore.cancelGeneration();
}

// 保存教案
async function handleSave() {
  if (!generatedLesson.value || isSaving.value) return;
  
  // 验证必填字段
  const title = generatedLesson.value.title || form.value.topic;
  const subject = form.value.subject;
  const grade = form.value.grade;
  
  if (!title || !subject || !grade) {
    saveError.value = '缺少必填信息（标题、学科或年级）';
    return;
  }
  
  isSaving.value = true;
  saveError.value = null;
  
  try {
    // 将前端结构转换为后端期望的结构
    const objectivesText = [
      generatedLesson.value.objectives?.knowledge ? `【知识与技能】\n${generatedLesson.value.objectives.knowledge}` : '',
      generatedLesson.value.objectives?.process ? `【过程与方法】\n${generatedLesson.value.objectives.process}` : '',
      generatedLesson.value.objectives?.emotion ? `【情感态度价值观】\n${generatedLesson.value.objectives.emotion}` : '',
    ].filter(Boolean).join('\n\n');
    
    // 将教学环节转换为内容文本
    const sections = generatedLesson.value.content?.sections || [];
    const contentText = sections.map((section, index) => {
      return `## ${index + 1}. ${section.title || '教学环节'}（${section.duration || 10}分钟）\n\n` +
        `**教师活动：**\n${section.teacherActivity || ''}\n\n` +
        `**学生活动：**\n${section.studentActivity || ''}\n\n` +
        (section.content ? `**教学内容：**\n${section.content}\n\n` : '') +
        (section.designIntent ? `**设计意图：**\n${section.designIntent}` : '');
    }).join('\n\n---\n\n');
    
    const lesson = await lessonStore.createLesson({
      title: title,
      subject: subject,
      grade: grade,
      duration: form.value.duration || 45,
      objectives: objectivesText || '',
      content: contentText || '',
      activities: sections.map(s => s.studentActivity || '').filter(Boolean).join('\n\n'),
      assessment: generatedLesson.value.evaluation || '',
      resources: generatedLesson.value.content?.materials?.join('\n') || '',
      tags: [subject, grade].filter(Boolean),
    } as any);
    
    router.push(`/lessons/${lesson.id}`);
  } catch (err) {
    saveError.value = err instanceof Error ? err.message : '保存失败，请重试';
    console.error('保存教案失败:', err);
  } finally {
    isSaving.value = false;
  }
}

// 重新生成
function handleRegenerate() {
  generationStore.reset();
}

// 获取进度状态图标
function getProgressIcon(status: string) {
  switch (status) {
    case 'completed':
      return CheckCircleIcon;
    case 'running':
      return ArrowPathIcon;
    case 'error':
      return XCircleIcon;
    default:
      return ClockIcon;
  }
}

// 获取进度状态颜色
function getProgressColor(status: string) {
  switch (status) {
    case 'completed':
      return 'text-green-500';
    case 'running':
      return 'text-primary-500 animate-spin';
    case 'error':
      return 'text-red-500';
    default:
      return 'text-gray-400';
  }
}
</script>

<template>
  <div class="max-w-4xl mx-auto space-y-8">
    <!-- Header -->
    <div>
      <h1 class="text-2xl font-bold text-gray-900">生成教案</h1>
      <p class="mt-1 text-sm text-gray-500">
        填写基本信息，让 AI 为您智能生成教案
      </p>
    </div>

    <!-- Form -->
    <div class="card">
      <div class="card-body space-y-6">
        <div class="grid grid-cols-1 gap-6 sm:grid-cols-2">
          <!-- 学科 -->
          <div>
            <label for="subject" class="label">学科 <span class="text-red-500">*</span></label>
            <select
              id="subject"
              v-model="form.subject"
              class="select"
              :disabled="isGenerating"
            >
              <option value="">请选择学科</option>
              <option v-for="s in subjects" :key="s" :value="s">{{ s }}</option>
            </select>
          </div>

          <!-- 年级 -->
          <div>
            <label for="grade" class="label">年级 <span class="text-red-500">*</span></label>
            <select
              id="grade"
              v-model="form.grade"
              class="select"
              :disabled="isGenerating"
            >
              <option value="">请选择年级</option>
              <option v-for="g in grades" :key="g" :value="g">{{ g }}</option>
            </select>
          </div>

          <!-- 课题 -->
          <div class="sm:col-span-2">
            <label for="topic" class="label">课题 <span class="text-red-500">*</span></label>
            <input
              id="topic"
              v-model="form.topic"
              type="text"
              class="input"
              placeholder="请输入课题名称，如：二次函数的图像与性质"
              :disabled="isGenerating"
            />
          </div>

          <!-- 课时 -->
          <div>
            <label for="duration" class="label">课时（分钟） <span class="text-red-500">*</span></label>
            <input
              id="duration"
              v-model.number="form.duration"
              type="number"
              min="20"
              max="180"
              class="input"
              :disabled="isGenerating"
            />
          </div>

          <!-- 教学风格 -->
          <div>
            <label for="style" class="label">教学风格</label>
            <select
              id="style"
              v-model="form.style"
              class="select"
              :disabled="isGenerating"
            >
              <option v-for="s in styles" :key="s.value" :value="s.value">
                {{ s.label }}
              </option>
            </select>
          </div>

          <!-- 特殊要求 -->
          <div class="sm:col-span-2">
            <label for="requirements" class="label">特殊要求（可选）</label>
            <textarea
              id="requirements"
              v-model="form.requirements"
              rows="3"
              class="input"
              placeholder="请输入任何特殊要求，如：需要包含小组讨论环节、注重培养学生的创新思维等"
              :disabled="isGenerating"
            />
          </div>
        </div>

        <!-- Buttons -->
        <div class="flex items-center gap-4 pt-4">
          <button
            v-if="!isGenerating"
            type="button"
            class="btn-primary inline-flex items-center gap-2"
            :disabled="!isValid"
            @click="handleGenerate"
          >
            <SparklesIcon class="h-5 w-5" />
            开始生成
          </button>
          <button
            v-else
            type="button"
            class="btn-danger inline-flex items-center gap-2"
            @click="handleCancel"
          >
            <XCircleIcon class="h-5 w-5" />
            取消生成
          </button>
        </div>
      </div>
    </div>

    <!-- Progress -->
    <div v-if="progress.length > 0" class="card">
      <div class="card-header">
        <h3 class="font-medium">生成进度</h3>
      </div>
      <div class="card-body">
        <div class="space-y-3">
          <div
            v-for="p in progress"
            :key="p.node"
            class="flex items-center gap-3"
          >
            <component
              :is="getProgressIcon(p.status)"
              :class="['h-5 w-5', getProgressColor(p.status)]"
            />
            <span
              :class="[
                'text-sm',
                p.status === 'completed' ? 'text-gray-900' :
                p.status === 'running' ? 'text-primary-600 font-medium' :
                p.status === 'error' ? 'text-red-600' :
                'text-gray-400'
              ]"
            >
              {{ generationStore.getNodeLabel(p.node) }}
            </span>
            <span v-if="p.message" class="text-xs text-red-500">
              {{ p.message }}
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- Error -->
    <div v-if="error && !isGenerating" class="card border-red-200 bg-red-50">
      <div class="card-body">
        <div class="flex items-start gap-3">
          <XCircleIcon class="h-6 w-6 text-red-500 flex-shrink-0" />
          <div>
            <h3 class="font-medium text-red-800">生成失败</h3>
            <p class="mt-1 text-sm text-red-700">{{ error }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Generated Lesson Preview -->
    <div v-if="generatedLesson && !isGenerating" class="card">
      <div class="card-header flex items-center justify-between">
        <h3 class="font-medium">生成结果</h3>
        <div class="flex items-center gap-2">
          <span v-if="saveError" class="text-sm text-red-500">{{ saveError }}</span>
          <button
            type="button"
            class="btn-outline btn-sm"
            @click="handleRegenerate"
            :disabled="isSaving"
          >
            重新生成
          </button>
          <button
            type="button"
            class="btn-primary btn-sm"
            @click="handleSave"
            :disabled="isSaving"
          >
            <span v-if="isSaving" class="flex items-center gap-2">
              <ArrowPathIcon class="h-4 w-4 animate-spin" />
              保存中...
            </span>
            <span v-else>保存教案</span>
          </button>
        </div>
      </div>
      <div class="card-body space-y-6">
        <!-- 标题 -->
        <div>
          <h2 class="text-xl font-bold text-gray-900">{{ generatedLesson.title }}</h2>
        </div>

        <!-- 教学目标 -->
        <div>
          <h4 class="font-medium text-gray-900 mb-2">教学目标</h4>
          <div class="space-y-2 text-sm text-gray-600">
            <p><strong>知识与技能：</strong>{{ generatedLesson.objectives.knowledge }}</p>
            <p><strong>过程与方法：</strong>{{ generatedLesson.objectives.process }}</p>
            <p><strong>情感态度价值观：</strong>{{ generatedLesson.objectives.emotion }}</p>
          </div>
        </div>

        <!-- 重点难点 -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <h4 class="font-medium text-gray-900 mb-2">教学重点</h4>
            <ul class="list-disc list-inside text-sm text-gray-600 space-y-1">
              <li v-for="(point, index) in generatedLesson.keyPoints" :key="index">
                {{ point }}
              </li>
            </ul>
          </div>
          <div>
            <h4 class="font-medium text-gray-900 mb-2">教学难点</h4>
            <ul class="list-disc list-inside text-sm text-gray-600 space-y-1">
              <li v-for="(point, index) in generatedLesson.difficultPoints" :key="index">
                {{ point }}
              </li>
            </ul>
          </div>
        </div>

        <!-- 教学方法 -->
        <div>
          <h4 class="font-medium text-gray-900 mb-2">教学方法</h4>
          <div class="flex flex-wrap gap-2">
            <span
              v-for="(method, index) in generatedLesson.teachingMethods"
              :key="index"
              class="badge-primary"
            >
              {{ method }}
            </span>
          </div>
        </div>

        <!-- 教学过程 -->
        <div>
          <h4 class="font-medium text-gray-900 mb-2">教学过程</h4>
          <div class="space-y-4">
            <div
              v-for="(section, index) in generatedLesson.content.sections"
              :key="index"
              class="p-4 bg-gray-50 rounded-lg"
            >
              <div class="flex items-center justify-between mb-2">
                <h5 class="font-medium text-gray-900">{{ section.title }}</h5>
                <span class="text-sm text-gray-500">{{ section.duration }}分钟</span>
              </div>
              <div class="text-sm text-gray-600 space-y-4">
                <div>
                  <strong class="block mb-1">教师活动：</strong>
                  <MarkdownRenderer :content="section.teacherActivity" />
                </div>
                <div>
                  <strong class="block mb-1">学生活动：</strong>
                  <MarkdownRenderer :content="section.studentActivity" />
                </div>
                <div v-if="section.content">
                  <strong class="block mb-1">教学内容：</strong>
                  <MarkdownRenderer :content="section.content" />
                </div>
                <div v-if="section.designIntent">
                  <strong class="block mb-1">设计意图：</strong>
                  <p>{{ section.designIntent }}</p>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 教学评价 -->
        <div>
          <h4 class="font-medium text-gray-900 mb-2">教学评价</h4>
          <div class="text-sm text-gray-600">
            <MarkdownRenderer :content="generatedLesson.evaluation" />
          </div>
        </div>

        <!-- 课后作业 -->
        <div v-if="generatedLesson.content.homework">
          <h4 class="font-medium text-gray-900 mb-2">课后作业</h4>
          <div class="text-sm text-gray-600">
            <MarkdownRenderer :content="generatedLesson.content.homework" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
