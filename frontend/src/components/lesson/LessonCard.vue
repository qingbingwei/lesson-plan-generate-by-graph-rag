<script setup lang="ts">
import { computed } from 'vue';
import { RouterLink } from 'vue-router';
import type { Lesson } from '@/types';
import { ClockIcon, BookOpenIcon } from '@heroicons/vue/24/outline';
import { HeartIcon as HeartSolidIcon } from '@heroicons/vue/24/solid';
import { HeartIcon as HeartOutlineIcon } from '@heroicons/vue/24/outline';

const props = defineProps<{
  lesson: Lesson;
  isFavorite?: boolean;
}>();

const emit = defineEmits<{
  favorite: [id: string];
  delete: [id: string];
}>();

const statusClass = computed(() => {
  return props.lesson.status === 'published'
    ? 'badge-success'
    : 'badge-warning';
});

const statusText = computed(() => {
  return props.lesson.status === 'published' ? '已发布' : '草稿';
});

function formatDate(dateStr: string): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
}
</script>

<template>
  <div class="card hover:shadow-lg transition-shadow">
    <div class="card-body">
      <div class="flex items-start justify-between">
        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-2 mb-2">
            <span class="badge-secondary text-xs">{{ lesson.subject }}</span>
            <span class="badge-secondary text-xs">{{ lesson.grade }}</span>
            <span :class="statusClass" class="text-xs">{{ statusText }}</span>
          </div>
          <RouterLink
            :to="`/lessons/${lesson.id}`"
            class="text-lg font-medium text-gray-900 hover:text-primary-600 line-clamp-1"
          >
            {{ lesson.title }}
          </RouterLink>
        </div>
        <button
          type="button"
          class="p-1 transition-colors"
          :class="isFavorite ? 'text-red-500' : 'text-gray-400 hover:text-red-500'"
          @click="emit('favorite', lesson.id)"
        >
          <HeartSolidIcon v-if="isFavorite" class="h-5 w-5" />
          <HeartOutlineIcon v-else class="h-5 w-5" />
        </button>
      </div>

      <div class="mt-3 flex items-center gap-4 text-xs text-gray-500">
        <span class="flex items-center gap-1">
          <ClockIcon class="h-4 w-4" />
          {{ lesson.duration }}分钟
        </span>
        <span class="flex items-center gap-1">
          <BookOpenIcon class="h-4 w-4" />
          v{{ lesson.version }}
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
      </div>
    </div>
  </div>
</template>
