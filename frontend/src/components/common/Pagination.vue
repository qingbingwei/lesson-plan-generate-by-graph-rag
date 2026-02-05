<script setup lang="ts">
import { computed } from 'vue';

const props = withDefaults(
  defineProps<{
    currentPage: number;
    totalPages: number;
    showPages?: number;
  }>(),
  {
    showPages: 5,
  }
);

const emit = defineEmits<{
  'update:currentPage': [page: number];
}>();

const pages = computed(() => {
  const result: (number | string)[] = [];
  const total = props.totalPages;
  const current = props.currentPage;
  const show = props.showPages;

  if (total <= show + 2) {
    for (let i = 1; i <= total; i++) {
      result.push(i);
    }
  } else {
    result.push(1);

    if (current > Math.ceil(show / 2) + 1) {
      result.push('...');
    }

    let start = Math.max(2, current - Math.floor(show / 2));
    let end = Math.min(total - 1, current + Math.floor(show / 2));

    if (current <= Math.ceil(show / 2) + 1) {
      end = show;
    }

    if (current >= total - Math.ceil(show / 2)) {
      start = total - show + 1;
    }

    for (let i = start; i <= end; i++) {
      result.push(i);
    }

    if (current < total - Math.ceil(show / 2)) {
      result.push('...');
    }

    result.push(total);
  }

  return result;
});

function goToPage(page: number) {
  if (page >= 1 && page <= props.totalPages && page !== props.currentPage) {
    emit('update:currentPage', page);
  }
}
</script>

<template>
  <nav class="flex items-center justify-center gap-1" aria-label="分页导航">
    <button
      type="button"
      class="p-2 rounded-lg text-gray-500 hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
      :disabled="currentPage <= 1"
      @click="goToPage(currentPage - 1)"
    >
      <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
      </svg>
    </button>

    <template v-for="page in pages" :key="page">
      <span
        v-if="page === '...'"
        class="px-3 py-2 text-gray-500"
      >
        ...
      </span>
      <button
        v-else
        type="button"
        class="px-3 py-2 rounded-lg text-sm font-medium transition-colors"
        :class="[
          page === currentPage
            ? 'bg-primary-600 text-white'
            : 'text-gray-700 hover:bg-gray-100',
        ]"
        @click="goToPage(page as number)"
      >
        {{ page }}
      </button>
    </template>

    <button
      type="button"
      class="p-2 rounded-lg text-gray-500 hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
      :disabled="currentPage >= totalPages"
      @click="goToPage(currentPage + 1)"
    >
      <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
      </svg>
    </button>
  </nav>
</template>
