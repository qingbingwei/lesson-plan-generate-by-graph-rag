<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useLessonStore } from '@/stores/lesson';
import { RouterLink } from 'vue-router';
import {
  TrashIcon,
  ClockIcon,
} from '@heroicons/vue/24/outline';
import { HeartIcon as HeartSolidIcon } from '@heroicons/vue/24/solid';

const lessonStore = useLessonStore();
const loading = ref(false);
const favorites = ref<string[]>([]);

const lessons = computed(() => 
  lessonStore.lessons.filter(l => favorites.value.includes(l.id))
);

async function loadFavorites() {
  loading.value = true;
  try {
    // 从本地存储获取收藏列表
    const stored = localStorage.getItem('favorites');
    if (stored) {
      favorites.value = JSON.parse(stored);
    }
    // 加载所有教案
    await lessonStore.fetchLessons();
  } catch (error) {
    console.error('Failed to load favorites:', error);
  } finally {
    loading.value = false;
  }
}

function removeFavorite(lessonId: string) {
  favorites.value = favorites.value.filter(id => id !== lessonId);
  localStorage.setItem('favorites', JSON.stringify(favorites.value));
}

function formatDate(dateStr: string): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });
}

onMounted(() => {
  loadFavorites();
});
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div>
      <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">我的收藏</h1>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
        收藏的教案，方便快速查找
      </p>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center py-12">
      <div class="loading loading-lg" />
    </div>

    <!-- Empty -->
    <div
      v-else-if="lessons.length === 0"
      class="text-center py-12 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg"
    >
      <HeartSolidIcon class="mx-auto h-12 w-12 text-gray-400" />
      <h3 class="mt-2 text-sm font-medium text-gray-900 dark:text-gray-100">暂无收藏</h3>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
        浏览教案时点击收藏按钮添加到收藏夹
      </p>
      <div class="mt-6">
        <RouterLink to="/lessons" class="btn-primary">
          浏览教案
        </RouterLink>
      </div>
    </div>

    <!-- List -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <div
        v-for="lesson in lessons"
        :key="lesson.id"
        class="card group hover:shadow-lg transition-shadow"
      >
        <div class="card-body">
          <div class="flex items-start justify-between">
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2 mb-2">
                <span class="badge-secondary text-xs">{{ lesson.subject }}</span>
                <span class="badge-secondary text-xs">{{ lesson.grade }}</span>
              </div>
              <RouterLink
                :to="`/lessons/${lesson.id}`"
                class="font-medium text-gray-900 dark:text-gray-100 hover:text-primary-600 dark:hover:text-primary-400 line-clamp-1"
              >
                {{ lesson.title }}
              </RouterLink>
            </div>
            <button
              type="button"
              class="p-1 text-red-500 hover:text-red-600 transition-colors"
              @click="removeFavorite(lesson.id)"
            >
              <HeartSolidIcon class="h-5 w-5" />
            </button>
          </div>

          <div class="mt-3 flex items-center gap-4 text-xs text-gray-500 dark:text-gray-400">
            <span class="flex items-center gap-1">
              <ClockIcon class="h-4 w-4" />
              {{ lesson.duration }}分钟
            </span>
            <span>{{ formatDate(lesson.updatedAt) }}</span>
          </div>

          <div class="mt-4 flex items-center gap-2">
            <RouterLink
              :to="`/lessons/${lesson.id}`"
              class="btn-outline btn-sm flex-1 text-center"
            >
              查看
            </RouterLink>
            <RouterLink
              :to="`/lessons/${lesson.id}/edit`"
              class="btn-outline btn-sm flex-1 text-center"
            >
              编辑
            </RouterLink>
            <button
              type="button"
              class="btn-ghost btn-sm text-red-500"
              @click="removeFavorite(lesson.id)"
            >
              <TrashIcon class="h-4 w-4" />
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
