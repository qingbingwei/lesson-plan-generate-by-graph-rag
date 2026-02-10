<script setup lang="ts">
import { computed } from 'vue';
import { useRouter } from 'vue-router';
import type { Lesson } from '@/types';
import { Star, StarFilled } from '@element-plus/icons-vue';

const props = defineProps<{
  lesson: Lesson;
  isFavorite?: boolean;
}>();

const emit = defineEmits<{
  favorite: [id: string];
  delete: [id: string];
}>();

const router = useRouter();

const statusType = computed<'success' | 'warning'>(() =>
  props.lesson.status === 'published' ? 'success' : 'warning'
);

const statusText = computed(() =>
  props.lesson.status === 'published' ? '已发布' : '草稿'
);

function formatDate(dateStr: string): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
}

function goDetail() {
  router.push(`/lessons/${props.lesson.id}`);
}

function goEdit() {
  router.push(`/lessons/${props.lesson.id}/edit`);
}
</script>

<template>
  <el-card class="surface-card card-hover" shadow="never">
    <div class="flex items-start justify-between gap-3">
      <div class="flex-1 min-w-0">
        <div class="flex items-center gap-2 mb-2">
          <el-tag size="small" effect="plain">{{ lesson.subject }}</el-tag>
          <el-tag size="small" effect="plain">{{ lesson.grade }}</el-tag>
          <el-tag size="small" :type="statusType">{{ statusText }}</el-tag>
        </div>
        <h3 class="text-base font-semibold app-text-primary line-clamp-1">{{ lesson.title }}</h3>
      </div>
      <el-button text circle @click="emit('favorite', lesson.id)">
        <el-icon :color="isFavorite ? '#ef4444' : '#94a3b8'">
          <StarFilled v-if="isFavorite" />
          <Star v-else />
        </el-icon>
      </el-button>
    </div>

    <div class="mt-3 flex flex-wrap items-center gap-3 text-xs app-text-muted">
      <span>时长：{{ lesson.duration }}分钟</span>
      <span>版本：v{{ lesson.version }}</span>
      <span>{{ formatDate(lesson.updatedAt) }}</span>
    </div>

    <div class="mt-4 flex gap-2">
      <el-button type="primary" plain class="!w-full" @click="goDetail">查看</el-button>
      <el-button class="!w-full" @click="goEdit">编辑</el-button>
    </div>
  </el-card>
</template>
