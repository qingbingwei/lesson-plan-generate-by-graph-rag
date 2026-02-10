<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import { useLessonStore } from '@/stores/lesson';
import { useDebounceFn } from '@/composables';
import { Plus, Search, Star, StarFilled } from '@element-plus/icons-vue';

const router = useRouter();
const lessonStore = useLessonStore();

const lessons = computed(() => lessonStore.lessons);
const loading = computed(() => lessonStore.loading);
const filters = computed(() => lessonStore.filters);

const debouncedSearch = useDebounceFn(() => {
  lessonStore.fetchLessons();
}, 400);

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
  <div class="page-container">
    <div class="page-header flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
      <div>
        <h1 class="page-title">我的教案</h1>
        <p class="page-subtitle">管理和编辑您创建的所有教案</p>
      </div>
      <el-button type="primary" :icon="Plus" @click="router.push('/generate')">生成新教案</el-button>
    </div>

    <el-card class="surface-card" shadow="never">
      <el-row :gutter="12">
        <el-col :xs="24" :md="10">
          <el-input
            v-model="filters.keyword"
            :prefix-icon="Search"
            placeholder="搜索教案..."
            clearable
            @input="debouncedSearch"
            @keyup.enter="handleSearch"
          />
        </el-col>
        <el-col :xs="24" :md="14">
          <div class="flex flex-wrap gap-2 justify-start md:justify-end mt-2 md:mt-0">
            <el-select v-model="filters.subject" style="width: 120px" @change="handleFilterChange">
              <el-option v-for="s in subjects" :key="s.value" :label="s.label" :value="s.value" />
            </el-select>
            <el-select v-model="filters.grade" style="width: 120px" @change="handleFilterChange">
              <el-option v-for="g in grades" :key="g.value" :label="g.label" :value="g.value" />
            </el-select>
            <el-select v-model="filters.status" style="width: 120px" @change="handleFilterChange">
              <el-option v-for="s in statuses" :key="s.value" :label="s.label" :value="s.value" />
            </el-select>
            <el-button @click="handleSearch">筛选</el-button>
          </div>
        </el-col>
      </el-row>
    </el-card>

    <el-card class="surface-card" shadow="never">
      <el-table v-loading="loading" :data="lessons" stripe>
        <el-table-column prop="title" label="标题" min-width="240" show-overflow-tooltip />
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
        <el-table-column label="创建时间" width="130">
          <template #default="{ row }">
            {{ new Date(row.createdAt).toLocaleDateString() }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button text type="primary" @click="router.push(`/lessons/${row.id}`)">查看</el-button>
            <el-button text @click="router.push(`/lessons/${row.id}/edit`)">编辑</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && lessons.length === 0" description="暂无教案" class="py-8" />

      <div v-if="lessonStore.totalPages > 1" class="mt-4 flex justify-between items-center">
        <el-text type="info">共 {{ lessonStore.total }} 条记录</el-text>
        <el-pagination
          background
          :current-page="lessonStore.page"
          :page-size="1"
          :total="lessonStore.totalPages"
          layout="prev, pager, next"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>
  </div>
</template>
