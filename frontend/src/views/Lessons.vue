<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { RouterLink } from 'vue-router';
import { useLessonStore } from '@/stores/lesson';
import {
  PlusIcon,
  MagnifyingGlassIcon,
  FunnelIcon,
  DocumentTextIcon,
} from '@heroicons/vue/24/outline';
import { HeartIcon as HeartOutlineIcon } from '@heroicons/vue/24/outline';
import { HeartIcon as HeartSolidIcon } from '@heroicons/vue/24/solid';

const lessonStore = useLessonStore();

const lessons = computed(() => lessonStore.lessons);
const loading = computed(() => lessonStore.loading);
const filters = computed(() => lessonStore.filters);

// 收藏功能
const favorites = ref<string[]>([]);

function loadFavorites() {
  const stored = localStorage.getItem('favorites');
  if (stored) {
    favorites.value = JSON.parse(stored);
  }
}

function isFavorite(lessonId: string): boolean {
  return favorites.value.includes(lessonId);
}

function toggleFavorite(event: MouseEvent, lessonId: string) {
  event.preventDefault();
  event.stopPropagation();
  
  if (isFavorite(lessonId)) {
    favorites.value = favorites.value.filter(id => id !== lessonId);
  } else {
    favorites.value.push(lessonId);
  }
  localStorage.setItem('favorites', JSON.stringify(favorites.value));
}

// 选项
const subjects = [
  { value: '', label: '全部学科' },
  { value: '语文', label: '语文' },
  { value: '数学', label: '数学' },
  { value: '英语', label: '英语' },
  { value: '物理', label: '物理' },
  { value: '化学', label: '化学' },
  { value: '生物', label: '生物' },
];

const grades = [
  { value: '', label: '全部年级' },
  { value: '七年级', label: '七年级' },
  { value: '八年级', label: '八年级' },
  { value: '九年级', label: '九年级' },
  { value: '高一', label: '高一' },
  { value: '高二', label: '高二' },
  { value: '高三', label: '高三' },
];

const statuses = [
  { value: '', label: '全部状态' },
  { value: 'draft', label: '草稿' },
  { value: 'published', label: '已发布' },
];

function handleSearch() {
  lessonStore.fetchLessons();
}

function handleFilterChange() {
  lessonStore.fetchLessons();
}

function handlePageChange(page: number) {
  lessonStore.fetchLessons({ page });
}

onMounted(() => {
  loadFavorites();
  lessonStore.fetchLessons();
});
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">我的教案</h1>
        <p class="mt-1 text-sm text-gray-500">
          管理和编辑您创建的所有教案
        </p>
      </div>
      <RouterLink to="/generate" class="btn-primary inline-flex items-center gap-2">
        <PlusIcon class="h-5 w-5" />
        生成新教案
      </RouterLink>
    </div>

    <!-- Filters -->
    <div class="card">
      <div class="card-body">
        <div class="flex flex-col lg:flex-row gap-4">
          <!-- Search -->
          <div class="flex-1">
            <div class="relative">
              <MagnifyingGlassIcon class="absolute left-3 top-1/2 -translate-y-1/2 h-5 w-5 text-gray-400" />
              <input
                v-model="filters.keyword"
                type="text"
                class="input pl-10"
                placeholder="搜索教案..."
                @keyup.enter="handleSearch"
              />
            </div>
          </div>

          <!-- Filter dropdowns -->
          <div class="flex flex-wrap gap-2">
            <select
              v-model="filters.subject"
              class="select w-auto"
              @change="handleFilterChange"
            >
              <option v-for="s in subjects" :key="s.value" :value="s.value">
                {{ s.label }}
              </option>
            </select>

            <select
              v-model="filters.grade"
              class="select w-auto"
              @change="handleFilterChange"
            >
              <option v-for="g in grades" :key="g.value" :value="g.value">
                {{ g.label }}
              </option>
            </select>

            <select
              v-model="filters.status"
              class="select w-auto"
              @change="handleFilterChange"
            >
              <option v-for="s in statuses" :key="s.value" :value="s.value">
                {{ s.label }}
              </option>
            </select>

            <button type="button" class="btn-outline" @click="handleSearch">
              <FunnelIcon class="h-5 w-5" />
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Lessons list -->
    <div class="card overflow-hidden">
      <div v-if="loading" class="p-8 text-center">
        <div class="loading mx-auto" />
        <p class="mt-2 text-sm text-gray-500">加载中...</p>
      </div>

      <div v-else-if="lessons.length === 0" class="p-8 text-center">
        <DocumentTextIcon class="mx-auto h-12 w-12 text-gray-400" />
        <h3 class="mt-2 text-sm font-medium text-gray-900">暂无教案</h3>
        <p class="mt-1 text-sm text-gray-500">开始生成您的第一个教案吧</p>
        <div class="mt-6">
          <RouterLink to="/generate" class="btn-primary inline-flex items-center gap-2">
            <PlusIcon class="h-5 w-5" />
            生成教案
          </RouterLink>
        </div>
      </div>

      <ul v-else class="divide-y divide-gray-200">
        <li
          v-for="lesson in lessons"
          :key="lesson.id"
          class="p-4 hover:bg-gray-50 transition-colors"
        >
          <RouterLink :to="`/lessons/${lesson.id}`" class="block">
            <div class="flex items-start justify-between gap-4">
              <div class="flex-1 min-w-0">
                <h3 class="text-base font-medium text-gray-900 truncate">
                  {{ lesson.title }}
                </h3>
                <div class="mt-2 flex flex-wrap items-center gap-2">
                  <span class="badge-secondary">{{ lesson.subject }}</span>
                  <span class="badge-secondary">{{ lesson.grade }}</span>
                  <span class="text-sm text-gray-500">{{ lesson.duration }}分钟</span>
                </div>
                <p class="mt-2 text-sm text-gray-500 line-clamp-2">
                  {{ lesson.objectives?.knowledge || '暂无描述' }}
                </p>
              </div>
              <div class="flex-shrink-0 text-right flex items-start gap-2">
                <button
                  type="button"
                  class="p-1.5 rounded-full transition-colors"
                  :class="isFavorite(lesson.id) ? 'text-red-500 hover:text-red-600' : 'text-gray-400 hover:text-red-500'"
                  :title="isFavorite(lesson.id) ? '取消收藏' : '添加收藏'"
                  @click="toggleFavorite($event, lesson.id)"
                >
                  <HeartSolidIcon v-if="isFavorite(lesson.id)" class="h-5 w-5" />
                  <HeartOutlineIcon v-else class="h-5 w-5" />
                </button>
                <div>
                  <span
                    :class="[
                      lesson.status === 'published' ? 'badge-success' : 'badge-secondary',
                    ]"
                  >
                    {{ lesson.status === 'published' ? '已发布' : '草稿' }}
                  </span>
                  <p class="mt-2 text-xs text-gray-500">
                    {{ new Date(lesson.createdAt).toLocaleDateString() }}
                  </p>
                </div>
              </div>
            </div>
          </RouterLink>
        </li>
      </ul>

      <!-- Pagination -->
      <div v-if="lessonStore.totalPages > 1" class="card-footer">
        <div class="flex items-center justify-between">
          <p class="text-sm text-gray-700">
            共 <span class="font-medium">{{ lessonStore.total }}</span> 条记录
          </p>
          <nav class="flex gap-1">
            <button
              v-for="page in lessonStore.totalPages"
              :key="page"
              type="button"
              :class="[
                'px-3 py-1 text-sm rounded',
                page === lessonStore.page
                  ? 'bg-primary-600 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              ]"
              @click="handlePageChange(page)"
            >
              {{ page }}
            </button>
          </nav>
        </div>
      </div>
    </div>
  </div>
</template>
