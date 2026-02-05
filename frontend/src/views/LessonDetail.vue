<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useLessonStore } from '@/stores/lesson';
import MarkdownRenderer from '@/components/common/MarkdownRenderer.vue';
import {
  PencilIcon,
  TrashIcon,
  ArrowDownTrayIcon,
  ShareIcon,
  ClockIcon,
  BookOpenIcon,
} from '@heroicons/vue/24/outline';
import { HeartIcon as HeartOutlineIcon } from '@heroicons/vue/24/outline';
import { HeartIcon as HeartSolidIcon } from '@heroicons/vue/24/solid';

const route = useRoute();
const router = useRouter();
const lessonStore = useLessonStore();

const lessonId = computed(() => route.params.id as string);
const lesson = computed(() => lessonStore.currentLesson);
const loading = computed(() => lessonStore.loading);
const publishing = ref(false);
const publishError = ref('');

// 收藏功能
const favorites = ref<string[]>([]);

function loadFavorites() {
  const stored = localStorage.getItem('favorites');
  if (stored) {
    favorites.value = JSON.parse(stored);
  }
}

const isFavorite = computed(() => favorites.value.includes(lessonId.value));

function toggleFavorite() {
  if (isFavorite.value) {
    favorites.value = favorites.value.filter(id => id !== lessonId.value);
  } else {
    favorites.value.push(lessonId.value);
  }
  localStorage.setItem('favorites', JSON.stringify(favorites.value));
}

// 点击外部关闭下拉菜单
function handleClickOutside(event: MouseEvent) {
  const target = event.target as HTMLElement;
  if (!target.closest('.relative')) {
    showExportMenu.value = false;
  }
}

onMounted(() => {
  loadFavorites();
  lessonStore.fetchLesson(lessonId.value);
  document.addEventListener('click', handleClickOutside);
});

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside);
});

async function handleDelete() {
  if (!confirm('确定要删除这个教案吗？')) return;
  
  try {
    await lessonStore.deleteLesson(lessonId.value);
    router.push('/lessons');
  } catch {
    alert('删除失败，请重试');
  }
}

async function handlePublish() {
  publishing.value = true;
  publishError.value = '';
  try {
    await lessonStore.publishLesson(lessonId.value);
    alert('发布成功！');
  } catch (err) {
    publishError.value = err instanceof Error ? err.message : '发布失败';
    alert(publishError.value);
  } finally {
    publishing.value = false;
  }
}

// 导出状态
const showExportMenu = ref(false);
const exporting = ref(false);

// 导出教案
async function handleExport(format: 'md' | 'pdf' | 'docx') {
  if (!lesson.value) return;
  
  showExportMenu.value = false;
  exporting.value = true;
  
  try {
    const response = await fetch(`/api/v1/lessons/${lessonId.value}/export?format=${format}`, {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('auth') ? JSON.parse(localStorage.getItem('auth')!).token : ''}`,
      },
    });
    
    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || '导出失败');
    }
    
    // 获取文件名
    const contentDisposition = response.headers.get('Content-Disposition');
    let filename = `${lesson.value.title}.${format}`;
    if (contentDisposition) {
      const match = contentDisposition.match(/filename="(.+)"/);
      if (match) {
        filename = match[1];
      }
    }
    
    // 下载文件
    const blob = await response.blob();
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  } catch (err) {
    alert(err instanceof Error ? err.message : '导出失败，请重试');
  } finally {
    exporting.value = false;
  }
}

// 解析JSON字符串中的文本
function parseJsonText(value: any): string {
  if (!value) return '';
  if (typeof value !== 'string') return String(value);
  
  // 尝试解析 JSON
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
    // 不是JSON，直接返回
    return value;
  }
}
</script>

<template>
  <div class="max-w-4xl mx-auto space-y-6">
    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center py-12">
      <div class="loading loading-lg" />
    </div>

    <!-- Content -->
    <template v-else-if="lesson">
      <!-- Header -->
      <div class="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4">
        <div>
          <div class="flex items-center gap-2 mb-2">
            <span class="badge-secondary">{{ lesson.subject }}</span>
            <span class="badge-secondary">{{ lesson.grade }}</span>
            <span
              :class="[
                lesson.status === 'published' ? 'badge-success' : 'badge-warning',
              ]"
            >
              {{ lesson.status === 'published' ? '已发布' : '草稿' }}
            </span>
          </div>
          <h1 class="text-2xl font-bold text-gray-900">{{ lesson.title }}</h1>
          <div class="mt-2 flex items-center gap-4 text-sm text-gray-500">
            <span class="flex items-center gap-1">
              <ClockIcon class="h-4 w-4" />
              {{ lesson.duration }}分钟
            </span>
            <span class="flex items-center gap-1">
              <BookOpenIcon class="h-4 w-4" />
              版本 {{ lesson.version }}
            </span>
          </div>
        </div>
        <div class="flex items-center gap-2">
          <button
            type="button"
            class="btn-outline btn-sm inline-flex items-center gap-1"
            :class="isFavorite ? 'text-red-500 border-red-500 hover:bg-red-50' : ''"
            @click="toggleFavorite"
          >
            <HeartSolidIcon v-if="isFavorite" class="h-4 w-4" />
            <HeartOutlineIcon v-else class="h-4 w-4" />
            {{ isFavorite ? '已收藏' : '收藏' }}
          </button>
          <button
            v-if="lesson.status === 'draft'"
            type="button"
            class="btn-success btn-sm"
            :disabled="publishing"
            @click="handlePublish"
          >
            {{ publishing ? '发布中...' : '发布' }}
          </button>
          <RouterLink
            :to="`/lessons/${lesson.id}/edit`"
            class="btn-outline btn-sm inline-flex items-center gap-1"
          >
            <PencilIcon class="h-4 w-4" />
            编辑
          </RouterLink>
          
          <!-- 导出下拉菜单 -->
          <div class="relative">
            <button
              type="button"
              class="btn-outline btn-sm inline-flex items-center gap-1"
              :disabled="exporting"
              @click="showExportMenu = !showExportMenu"
            >
              <ArrowDownTrayIcon class="h-4 w-4" />
              {{ exporting ? '导出中...' : '导出' }}
            </button>
            <div
              v-if="showExportMenu"
              class="absolute right-0 mt-2 w-40 bg-white rounded-md shadow-lg ring-1 ring-black ring-opacity-5 z-10"
            >
              <div class="py-1">
                <button
                  type="button"
                  class="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                  @click="handleExport('md')"
                >
                  📝 Markdown (.md)
                </button>
                <button
                  type="button"
                  class="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                  @click="handleExport('docx')"
                >
                  📄 Word (.docx)
                </button>
                <button
                  type="button"
                  class="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                  @click="handleExport('pdf')"
                >
                  📕 PDF (.pdf)
                </button>
              </div>
            </div>
          </div>
          
          <button
            type="button"
            class="btn-outline btn-sm inline-flex items-center gap-1"
          >
            <ShareIcon class="h-4 w-4" />
            分享
          </button>
          <button
            type="button"
            class="btn-danger btn-sm inline-flex items-center gap-1"
            @click="handleDelete"
          >
            <TrashIcon class="h-4 w-4" />
            删除
          </button>
        </div>
      </div>

      <!-- Content -->
      <div class="space-y-6">
        <!-- 教学目标 -->
        <div v-if="(lesson as any).objectives" class="card">
          <div class="card-header">
            <h3 class="font-medium">教学目标</h3>
          </div>
          <div class="card-body prose prose-sm max-w-none">
            <MarkdownRenderer :content="parseJsonText((lesson as any).objectives)" />
          </div>
        </div>

        <!-- 教学内容 -->
        <div v-if="(lesson as any).content" class="card">
          <div class="card-header">
            <h3 class="font-medium">教学内容</h3>
          </div>
          <div class="card-body prose prose-sm max-w-none">
            <MarkdownRenderer :content="parseJsonText((lesson as any).content)" />
          </div>
        </div>

        <!-- 教学活动 -->
        <div v-if="(lesson as any).activities" class="card">
          <div class="card-header">
            <h3 class="font-medium">教学活动</h3>
          </div>
          <div class="card-body prose prose-sm max-w-none">
            <MarkdownRenderer :content="(lesson as any).activities" />
          </div>
        </div>

        <!-- 教学评价 -->
        <div v-if="(lesson as any).assessment" class="card">
          <div class="card-header">
            <h3 class="font-medium">教学评价</h3>
          </div>
          <div class="card-body prose prose-sm max-w-none">
            <MarkdownRenderer :content="(lesson as any).assessment" />
          </div>
        </div>

        <!-- 教学资源 -->
        <div v-if="(lesson as any).resources" class="card">
          <div class="card-header">
            <h3 class="font-medium">教学资源</h3>
          </div>
          <div class="card-body prose prose-sm max-w-none">
            <MarkdownRenderer :content="(lesson as any).resources" />
          </div>
        </div>
      </div>
    </template>

    <!-- Not found -->
    <div v-else class="text-center py-12">
      <h2 class="text-lg font-medium text-gray-900">教案不存在</h2>
      <p class="mt-1 text-sm text-gray-500">该教案可能已被删除</p>
      <RouterLink to="/lessons" class="mt-4 btn-primary inline-block">
        返回列表
      </RouterLink>
    </div>
  </div>
</template>
