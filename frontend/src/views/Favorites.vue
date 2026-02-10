<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import { useLessonStore } from '@/stores/lesson';
import { Star, StarFilled, Delete } from '@element-plus/icons-vue';

const router = useRouter();
const lessonStore = useLessonStore();
const loading = ref(false);
const favorites = ref<string[]>([]);

const lessons = computed(() =>
  lessonStore.lessons.filter(l => favorites.value.includes(l.id))
);

async function loadFavorites() {
  loading.value = true;
  try {
    const stored = localStorage.getItem('favorites');
    if (stored) {
      favorites.value = JSON.parse(stored);
    }
    await lessonStore.fetchLessons();
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
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">我的收藏</h1>
      <p class="page-subtitle">收藏的教案，方便快速查找</p>
    </div>

    <el-card class="surface-card" shadow="never">
      <el-skeleton v-if="loading" :rows="4" animated />

      <el-empty v-else-if="lessons.length === 0" description="暂无收藏">
        <template #image>
          <el-icon :size="48" class="app-icon-muted"><Star /></el-icon>
        </template>
        <el-button type="primary" @click="router.push('/lessons')">浏览教案</el-button>
      </el-empty>

      <el-row v-else :gutter="16">
        <el-col v-for="lesson in lessons" :key="lesson.id" :xs="24" :sm="12" :lg="8" class="mb-4">
          <el-card class="surface-card card-hover" shadow="never">
            <div class="flex items-start justify-between gap-2">
              <div class="min-w-0">
                <div class="flex gap-1 mb-2">
                  <el-tag size="small" effect="plain">{{ lesson.subject }}</el-tag>
                  <el-tag size="small" effect="plain">{{ lesson.grade }}</el-tag>
                </div>
                <div class="font-semibold app-text-primary line-clamp-1">{{ lesson.title }}</div>
              </div>
              <el-button text circle @click="removeFavorite(lesson.id)">
                <el-icon class="app-icon-danger"><StarFilled /></el-icon>
              </el-button>
            </div>

            <div class="mt-3 text-xs app-text-muted">
              <div>时长：{{ lesson.duration }}分钟</div>
              <div class="mt-1">更新：{{ formatDate(lesson.updatedAt) }}</div>
            </div>

            <div class="mt-4 flex gap-2">
              <el-button class="!w-full" @click="router.push(`/lessons/${lesson.id}`)">查看</el-button>
              <el-button class="!w-full" @click="router.push(`/lessons/${lesson.id}/edit`)">编辑</el-button>
              <el-button type="danger" plain :icon="Delete" @click="removeFavorite(lesson.id)" />
            </div>
          </el-card>
        </el-col>
      </el-row>
    </el-card>
  </div>
</template>
