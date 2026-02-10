<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import { useLessonStore } from '@/stores/lesson';
import { getGenerationStats, type DashboardStats } from '@/api/generation';
import { MagicStick, Document, DataAnalysis, Histogram, Plus, Star, StarFilled, Close } from '@element-plus/icons-vue';

const router = useRouter();
const authStore = useAuthStore();
const lessonStore = useLessonStore();

const showOnboarding = ref(false);
const favorites = ref<string[]>([]);
const statsData = ref<DashboardStats | null>(null);
const statsLoading = ref(true);

function dismissOnboarding() {
  showOnboarding.value = false;
  localStorage.setItem('onboarding_dismissed', 'true');
}

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
    favorites.value = favorites.value.filter((id) => id !== lessonId);
  } else {
    favorites.value.push(lessonId);
  }
  localStorage.setItem('favorites', JSON.stringify(favorites.value));
}

function formatTokenCount(count: number): string {
  if (count >= 1000000) return `${(count / 1000000).toFixed(1)}M`;
  if (count >= 1000) return `${(count / 1000).toFixed(1)}K`;
  return count.toString();
}

const stats = computed(() => [
  {
    name: '我的教案',
    value: statsData.value?.total_lessons?.toString() || lessonStore.total.toString() || '0',
    icon: Document,
    color: '#3b82f6',
  },
  {
    name: '本月生成',
    value: statsData.value?.this_month_generations?.toString() || '0',
    icon: MagicStick,
    color: '#10b981',
  },
  {
    name: '总生成次数',
    value: statsData.value?.total_count?.toString() || '0',
    icon: DataAnalysis,
    color: '#8b5cf6',
  },
  {
    name: 'Token 使用',
    value: formatTokenCount(statsData.value?.total_tokens || 0),
    icon: Histogram,
    color: '#f59e0b',
  },
]);

const quickActions = [
  {
    name: '快速生成教案',
    description: '使用 AI 智能生成教案',
    href: '/generate',
    icon: MagicStick,
  },
  {
    name: '查看我的教案',
    description: '管理和编辑您的教案',
    href: '/lessons',
    icon: Document,
  },
  {
    name: '探索知识图谱',
    description: '可视化知识点关系',
    href: '/knowledge',
    icon: DataAnalysis,
  },
];

onMounted(async () => {
  if (!localStorage.getItem('onboarding_dismissed')) {
    showOnboarding.value = true;
  }

  loadFavorites();
  lessonStore.fetchLessons({ page: 1, pageSize: 5 });

  try {
    statsData.value = await getGenerationStats();
  } catch {
    // noop
  } finally {
    statsLoading.value = false;
  }
});
</script>

<template>
  <div class="page-container">
    <div class="page-header flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
      <div>
        <h1 class="page-title">欢迎回来，{{ authStore.userName }}！</h1>
        <p class="page-subtitle">今天想要生成什么教案呢？</p>
      </div>
      <el-button type="primary" :icon="Plus" @click="router.push('/generate')">生成新教案</el-button>
    </div>

    <el-alert v-if="showOnboarding" type="success" show-icon :closable="false" class="surface-card">
      <template #title>快速上手指南</template>
      <template #default>
        <div class="text-sm leading-6">
          <p><strong>1.</strong> 前往「知识库管理」上传教学文档，系统自动构建知识图谱</p>
          <p><strong>2.</strong> 在「生成教案」页选择学科、年级和课题，一键 AI 生成完整教案</p>
          <p><strong>3.</strong> 在「知识图谱」中探索知识点关系，支持缩放、高亮、筛选等交互</p>
          <p><strong>4.</strong> 生成后的教案支持保存、编辑和导出</p>
          <div class="mt-2">
            <el-button text :icon="Close" @click="dismissOnboarding">关闭引导</el-button>
          </div>
        </div>
      </template>
    </el-alert>

    <el-row :gutter="16">
      <el-col v-for="stat in stats" :key="stat.name" :xs="12" :sm="12" :md="12" :lg="6">
        <el-card class="surface-card card-hover" shadow="never">
          <div class="flex items-center gap-4">
            <el-icon :size="24" :color="stat.color"><component :is="stat.icon" /></el-icon>
            <div>
              <div class="text-2xl font-bold app-text-primary">
                <el-skeleton v-if="statsLoading" :rows="1" animated />
                <template v-else>{{ stat.value }}</template>
              </div>
              <div class="text-sm app-text-muted">{{ stat.name }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card class="surface-card" shadow="never">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="font-semibold">快捷操作</span>
        </div>
      </template>
      <el-row :gutter="16">
        <el-col v-for="action in quickActions" :key="action.name" :xs="24" :sm="12" :md="8">
          <el-card class="surface-card card-hover cursor-pointer" shadow="never" @click="router.push(action.href)">
            <div class="flex items-start gap-3">
              <el-icon :size="22" class="app-icon-primary"><component :is="action.icon" /></el-icon>
              <div>
                <div class="font-semibold app-text-primary">{{ action.name }}</div>
                <div class="text-sm app-text-muted mt-1">{{ action.description }}</div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </el-card>

    <el-card class="surface-card" shadow="never">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="font-semibold">最近教案</span>
          <el-button text @click="router.push('/lessons')">查看全部</el-button>
        </div>
      </template>

      <el-skeleton v-if="lessonStore.loading" :rows="4" animated />

      <el-empty v-else-if="lessonStore.lessons.length === 0" description="暂无教案">
        <el-button type="primary" :icon="Plus" @click="router.push('/generate')">生成第一个教案</el-button>
      </el-empty>

      <el-table v-else :data="lessonStore.lessons" stripe>
        <el-table-column prop="title" label="标题" min-width="220" show-overflow-tooltip />
        <el-table-column prop="subject" label="学科" width="120" />
        <el-table-column prop="grade" label="年级" width="120" />
        <el-table-column prop="duration" label="时长(分钟)" width="110" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'published' ? 'success' : 'info'" size="small">
              {{ row.status === 'published' ? '已发布' : '草稿' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="收藏" width="90">
          <template #default="{ row }">
            <el-button text circle @click="toggleFavorite($event, row.id)">
              <el-icon :color="isFavorite(row.id) ? '#ef4444' : '#94a3b8'">
                <StarFilled v-if="isFavorite(row.id)" />
                <Star v-else />
              </el-icon>
            </el-button>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-button text type="primary" @click="router.push(`/lessons/${row.id}`)">查看</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>
