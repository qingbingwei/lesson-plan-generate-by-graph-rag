<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { RouterLink } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import { useLessonStore } from '@/stores/lesson';
import { getGenerationStats, type DashboardStats } from '@/api/generation';
import {
  SparklesIcon,
  DocumentTextIcon,
  AcademicCapIcon,
  ChartBarIcon,
  PlusIcon,
} from '@heroicons/vue/24/outline';
import { HeartIcon as HeartOutlineIcon } from '@heroicons/vue/24/outline';
import { HeartIcon as HeartSolidIcon } from '@heroicons/vue/24/solid';

const authStore = useAuthStore();
const lessonStore = useLessonStore();

// 新手引导
const showOnboarding = ref(false);

function dismissOnboarding() {
  showOnboarding.value = false;
  localStorage.setItem('onboarding_dismissed', 'true');
}

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

// 统计数据
const statsData = ref<DashboardStats | null>(null);
const statsLoading = ref(true);

// 计算属性：动态统计
const stats = computed(() => [
  { 
    name: '我的教案', 
    value: statsData.value?.total_lessons?.toString() || lessonStore.total.toString() || '0', 
    icon: DocumentTextIcon, 
    color: 'bg-blue-500' 
  },
  { 
    name: '本月生成', 
    value: statsData.value?.this_month_generations?.toString() || '0', 
    icon: SparklesIcon, 
    color: 'bg-green-500' 
  },
  { 
    name: '总生成次数', 
    value: statsData.value?.total_count?.toString() || '0', 
    icon: AcademicCapIcon, 
    color: 'bg-purple-500' 
  },
  { 
    name: 'Token 使用', 
    value: formatTokenCount(statsData.value?.total_tokens || 0), 
    icon: ChartBarIcon, 
    color: 'bg-orange-500' 
  },
]);

// 格式化Token数量
function formatTokenCount(count: number): string {
  if (count >= 1000000) {
    return (count / 1000000).toFixed(1) + 'M';
  } else if (count >= 1000) {
    return (count / 1000).toFixed(1) + 'K';
  }
  return count.toString();
}

const quickActions = [
  {
    name: '快速生成教案',
    description: '使用 AI 智能生成教案',
    href: '/generate',
    icon: SparklesIcon,
    color: 'bg-gradient-to-r from-primary-500 to-primary-600',
  },
  {
    name: '查看我的教案',
    description: '管理和编辑您的教案',
    href: '/lessons',
    icon: DocumentTextIcon,
    color: 'bg-gradient-to-r from-secondary-500 to-secondary-600',
  },
  {
    name: '探索知识图谱',
    description: '可视化知识点关系',
    href: '/knowledge',
    icon: AcademicCapIcon,
    color: 'bg-gradient-to-r from-purple-500 to-purple-600',
  },
];

onMounted(async () => {
  // 新手引导检测
  if (!localStorage.getItem('onboarding_dismissed')) {
    showOnboarding.value = true;
  }
  
  // 加载收藏列表
  loadFavorites();
  
  // 获取教案列表
  lessonStore.fetchLessons({ page: 1, pageSize: 5 });
  
  // 获取统计数据
  try {
    statsData.value = await getGenerationStats();
  } catch (error) {
    console.error('获取统计数据失败:', error);
  } finally {
    statsLoading.value = false;
  }
});
</script>

<template>
  <div class="space-y-8">
    <!-- Welcome section -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">
          欢迎回来，{{ authStore.userName }}！
        </h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
          今天想要生成什么教案呢？
        </p>
      </div>
      <RouterLink to="/generate" class="btn-primary inline-flex items-center gap-2">
        <PlusIcon class="h-5 w-5" />
        生成新教案
      </RouterLink>
    </div>

    <!-- 新手引导 -->
    <transition name="slide-up">
      <div v-if="showOnboarding" class="card bg-gradient-to-r from-primary-50 to-blue-50 border-primary-200">
        <div class="card-body">
          <div class="flex items-start justify-between gap-4">
            <div class="flex items-start gap-3">
              <div class="p-2 bg-primary-100 rounded-lg flex-shrink-0">
                <SparklesIcon class="h-6 w-6 text-primary-600" />
              </div>
              <div>
                <h3 class="font-medium text-primary-900">快速上手指南</h3>
                <div class="mt-2 text-sm text-primary-700 space-y-1">
                  <p><strong>1.</strong> 前往「知识库管理」上传教学文档 (.txt/.md)，系统将自动构建知识图谱</p>
                  <p><strong>2.</strong> 在「生成教案」页选择学科、年级和课题，一键 AI 生成完整教案</p>
                  <p><strong>3.</strong> 在「知识图谱」中探索知识点关系，支持缩放、高亮、筛选等交互</p>
                  <p><strong>4.</strong> 生成的教案可保存、编辑，并导出为 Markdown / PDF / Word 格式</p>
                </div>
              </div>
            </div>
            <button
              type="button"
              class="btn-icon flex-shrink-0 text-primary-400 hover:text-primary-600"
              @click="dismissOnboarding"
              title="关闭引导"
            >
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="h-5 w-5">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>
      </div>
    </transition>

    <!-- Stats -->
    <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
      <div
        v-for="stat in stats"
        :key="stat.name"
        class="card p-6"
      >
        <div class="flex items-center gap-4">
          <div :class="[stat.color, 'flex-shrink-0 p-3 rounded-lg']">
            <component :is="stat.icon" class="h-6 w-6 text-white" />
          </div>
          <div>
            <p class="text-2xl font-semibold text-gray-900 dark:text-gray-100">{{ stat.value }}</p>
            <p class="text-sm text-gray-500 dark:text-gray-400">{{ stat.name }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Quick actions -->
    <div>
      <h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">快捷操作</h2>
      <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <RouterLink
          v-for="action in quickActions"
          :key="action.name"
          :to="action.href"
          class="card p-6 hover:shadow-md transition-shadow group"
        >
          <div class="flex items-start gap-4">
            <div :class="[action.color, 'flex-shrink-0 p-3 rounded-lg']">
              <component :is="action.icon" class="h-6 w-6 text-white" />
            </div>
            <div>
              <h3 class="font-medium text-gray-900 dark:text-gray-100 group-hover:text-primary-600 transition-colors">
                {{ action.name }}
              </h3>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ action.description }}</p>
            </div>
          </div>
        </RouterLink>
      </div>
    </div>

    <!-- Recent lessons -->
    <div>
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">最近教案</h2>
        <RouterLink to="/lessons" class="text-sm font-medium text-primary-600 hover:text-primary-500">
          查看全部
        </RouterLink>
      </div>
      
      <div class="card overflow-hidden">
        <div v-if="lessonStore.loading" class="p-8 text-center">
          <div class="loading mx-auto" />
          <p class="mt-2 text-sm text-gray-500">加载中...</p>
        </div>
        
        <div v-else-if="lessonStore.lessons.length === 0" class="p-8 text-center">
          <DocumentTextIcon class="mx-auto h-12 w-12 text-gray-400" />
          <p class="mt-2 text-sm text-gray-500">暂无教案</p>
          <RouterLink to="/generate" class="mt-4 btn-primary inline-flex items-center gap-2">
            <PlusIcon class="h-5 w-5" />
            生成第一个教案
          </RouterLink>
        </div>
        
        <ul v-else class="divide-y divide-gray-200 dark:divide-gray-700">
          <li
            v-for="lesson in lessonStore.lessons"
            :key="lesson.id"
            class="p-4 hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors"
          >
            <RouterLink :to="`/lessons/${lesson.id}`" class="block">
              <div class="flex items-center justify-between">
                <div class="flex-1 min-w-0">
                  <p class="text-sm font-medium text-gray-900 dark:text-gray-100 truncate">
                    {{ lesson.title }}
                  </p>
                  <div class="mt-1 flex items-center gap-2">
                    <span class="badge-secondary">{{ lesson.subject }}</span>
                    <span class="badge-secondary">{{ lesson.grade }}</span>
                    <span class="text-xs text-gray-500">
                      {{ lesson.duration }}分钟
                    </span>
                  </div>
                </div>
                <div class="ml-4 flex-shrink-0 flex items-center gap-2">
                  <button
                    type="button"
                    class="p-1 rounded-full transition-colors"
                    :class="isFavorite(lesson.id) ? 'text-red-500 hover:text-red-600' : 'text-gray-400 hover:text-red-500'"
                    :title="isFavorite(lesson.id) ? '取消收藏' : '添加收藏'"
                    @click="toggleFavorite($event, lesson.id)"
                  >
                    <HeartSolidIcon v-if="isFavorite(lesson.id)" class="h-4 w-4" />
                    <HeartOutlineIcon v-else class="h-4 w-4" />
                  </button>
                  <span
                    :class="[
                      lesson.status === 'published' ? 'badge-success' : 'badge-secondary',
                    ]"
                  >
                    {{ lesson.status === 'published' ? '已发布' : '草稿' }}
                  </span>
                </div>
              </div>
            </RouterLink>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>
