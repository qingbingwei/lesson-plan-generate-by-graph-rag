<script setup lang="ts">
import { ref, computed } from 'vue';
import { useRouter } from 'vue-router';
import { useGenerationStore } from '@/stores/generation';
import { useLessonStore } from '@/stores/lesson';
import MarkdownRenderer from '@/components/common/MarkdownRenderer.vue';
import { MagicStick, Refresh, DocumentAdd } from '@element-plus/icons-vue';

const router = useRouter();
const generationStore = useGenerationStore();
const lessonStore = useLessonStore();

const isSaving = ref(false);
const saveError = ref<string | null>(null);

const form = ref({
  subject: '',
  grade: '',
  topic: '',
  duration: 45,
  style: '',
  requirements: '',
});

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

const templates = [
  { label: '小学数学 · 分数', subject: '数学', grade: '五年级', topic: '分数的加法和减法', duration: 40, style: 'interactive' },
  { label: '初中语文 · 古诗', subject: '语文', grade: '七年级', topic: '唐诗三百首赏析', duration: 45, style: '' },
  { label: '高中物理 · 力学', subject: '物理', grade: '高一', topic: '牛顿第二定律', duration: 45, style: 'lecture' },
];

function applyTemplate(tpl: typeof templates[number]) {
  form.value.subject = tpl.subject;
  form.value.grade = tpl.grade;
  form.value.topic = tpl.topic;
  form.value.duration = tpl.duration;
  form.value.style = tpl.style;
  form.value.requirements = '';
}

const isGenerating = computed(() => generationStore.isGenerating);
const progress = computed(() => generationStore.progress);
const generatedLesson = computed(() => generationStore.generatedLesson);
const error = computed(() => generationStore.error);

const isValid = computed(() => {
  return form.value.subject && form.value.grade && form.value.topic && form.value.duration > 0;
});

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

function handleCancel() {
  generationStore.cancelGeneration();
}

async function handleSave() {
  if (!generatedLesson.value || isSaving.value) return;

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
    const objectivesText = [
      generatedLesson.value.objectives?.knowledge ? `【知识与技能】\n${generatedLesson.value.objectives.knowledge}` : '',
      generatedLesson.value.objectives?.process ? `【过程与方法】\n${generatedLesson.value.objectives.process}` : '',
      generatedLesson.value.objectives?.emotion ? `【情感态度价值观】\n${generatedLesson.value.objectives.emotion}` : '',
    ].filter(Boolean).join('\n\n');

    const sections = generatedLesson.value.content?.sections || [];
    const contentText = sections.map((section, index) => {
      return `## ${index + 1}. ${section.title || '教学环节'}（${section.duration || 10}分钟）\n\n` +
        `**教师活动：**\n${section.teacherActivity || ''}\n\n` +
        `**学生活动：**\n${section.studentActivity || ''}\n\n` +
        (section.content ? `**教学内容：**\n${section.content}\n\n` : '') +
        (section.designIntent ? `**设计意图：**\n${section.designIntent}` : '');
    }).join('\n\n---\n\n');

    const lesson = await lessonStore.createLesson({
      title,
      subject,
      grade,
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
  } finally {
    isSaving.value = false;
  }
}

function handleRegenerate() {
  generationStore.reset();
}
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">生成教案</h1>
      <p class="page-subtitle">填写关键信息并使用 AI 智能生成教案</p>
    </div>

    <el-card class="surface-card" shadow="never">
      <template #header>
        <div class="font-semibold">快速模板</div>
      </template>
      <div class="flex flex-wrap gap-2">
        <el-button
          v-for="tpl in templates"
          :key="tpl.label"
          plain
          @click="applyTemplate(tpl)"
        >
          {{ tpl.label }}
        </el-button>
      </div>
    </el-card>

    <el-card class="surface-card" shadow="never">
      <template #header>
        <div class="font-semibold">生成参数</div>
      </template>

      <el-form :model="form" label-position="top">
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
            <el-form-item label="课时（分钟）" required>
              <el-input-number v-model="form.duration" :min="20" :max="120" :step="5" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="课题" required>
          <el-input v-model="form.topic" placeholder="例如：分数的加法和减法" clearable />
        </el-form-item>

        <el-form-item label="教学风格">
          <el-select v-model="form.style" placeholder="请选择风格">
            <el-option v-for="style in styles" :key="style.value" :label="style.label" :value="style.value" />
          </el-select>
        </el-form-item>

        <el-form-item label="额外要求">
          <el-input
            v-model="form.requirements"
            type="textarea"
            :rows="4"
            placeholder="可选：例如希望加入分层练习、小组讨论等"
          />
        </el-form-item>

        <div class="flex flex-wrap gap-2">
          <el-button type="primary" :icon="MagicStick" :loading="isGenerating" :disabled="!isValid" @click="handleGenerate">
            开始生成
          </el-button>
          <el-button v-if="isGenerating" @click="handleCancel">取消</el-button>
        </div>
      </el-form>
    </el-card>

    <el-card v-if="progress.length > 0" class="surface-card" shadow="never">
      <template #header>
        <div class="font-semibold">生成进度</div>
      </template>
      <el-timeline>
        <el-timeline-item
          v-for="(p, index) in progress"
          :key="index"
          :timestamp="generationStore.getNodeLabel(p.node)"
          :type="p.status === 'completed' ? 'success' : p.status === 'running' ? 'primary' : p.status === 'error' ? 'danger' : 'info'"
        >
          <span v-if="p.message">{{ p.message }}</span>
          <span v-else>处理中...</span>
        </el-timeline-item>
      </el-timeline>
    </el-card>

    <el-alert v-if="error && !isGenerating" :title="error" type="error" show-icon />

    <el-card v-if="generatedLesson && !isGenerating" class="surface-card" shadow="never">
      <template #header>
        <div class="flex flex-wrap items-center justify-between gap-2">
          <div class="font-semibold">生成结果</div>
          <div>
            <el-button :icon="Refresh" @click="handleRegenerate" :disabled="isSaving">重新生成</el-button>
            <el-button type="primary" :icon="DocumentAdd" :loading="isSaving" @click="handleSave">保存教案</el-button>
          </div>
        </div>
      </template>

      <el-alert v-if="saveError" :title="saveError" type="warning" show-icon class="mb-4" />

      <div class="space-y-6">
        <div>
          <h2 class="text-xl font-bold app-text-primary">{{ generatedLesson.title }}</h2>
        </div>

        <el-descriptions :column="1" border>
          <el-descriptions-item label="知识与技能">{{ generatedLesson.objectives.knowledge }}</el-descriptions-item>
          <el-descriptions-item label="过程与方法">{{ generatedLesson.objectives.process }}</el-descriptions-item>
          <el-descriptions-item label="情感态度价值观">{{ generatedLesson.objectives.emotion }}</el-descriptions-item>
        </el-descriptions>

        <div>
          <h3 class="font-semibold mb-2">教学方法</h3>
          <div class="flex flex-wrap gap-2">
            <el-tag v-for="(method, index) in generatedLesson.teachingMethods" :key="index" type="primary" effect="plain">
              {{ method }}
            </el-tag>
          </div>
        </div>

        <div>
          <h3 class="font-semibold mb-2">教学过程</h3>
          <el-collapse accordion>
            <el-collapse-item
              v-for="(section, index) in generatedLesson.content.sections"
              :key="index"
              :name="String(index)"
              :title="`${section.title}（${section.duration}分钟）`"
            >
              <div class="space-y-3 text-sm">
                <div>
                  <strong>教师活动：</strong>
                  <MarkdownRenderer :content="section.teacherActivity" />
                </div>
                <div>
                  <strong>学生活动：</strong>
                  <MarkdownRenderer :content="section.studentActivity" />
                </div>
                <div v-if="section.content">
                  <strong>教学内容：</strong>
                  <MarkdownRenderer :content="section.content" />
                </div>
                <div v-if="section.designIntent">
                  <strong>设计意图：</strong>
                  <span>{{ section.designIntent }}</span>
                </div>
              </div>
            </el-collapse-item>
          </el-collapse>
        </div>

        <div>
          <h3 class="font-semibold mb-2">教学评价</h3>
          <MarkdownRenderer :content="generatedLesson.evaluation" />
        </div>

        <div v-if="generatedLesson.content.homework">
          <h3 class="font-semibold mb-2">课后作业</h3>
          <MarkdownRenderer :content="generatedLesson.content.homework" />
        </div>
      </div>
    </el-card>
  </div>
</template>
