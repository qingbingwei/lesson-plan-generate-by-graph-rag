<script setup lang="ts">
import { computed } from 'vue';
import { marked } from 'marked';

const props = defineProps<{
  content: string;
}>();

// 配置marked
marked.setOptions({
  breaks: true,
  gfm: true,
});

const renderedContent = computed(() => {
  if (!props.content) return '';
  return marked(props.content);
});
</script>

<template>
  <div class="markdown-body prose prose-sm max-w-none" v-html="renderedContent" />
</template>

<style scoped>
.markdown-body :deep(h1) {
  @apply text-xl font-bold mb-4 mt-6 first:mt-0;
}
.markdown-body :deep(h2) {
  @apply text-lg font-semibold mb-3 mt-5;
}
.markdown-body :deep(h3) {
  @apply text-base font-semibold mb-2 mt-4;
}
.markdown-body :deep(h4) {
  @apply text-sm font-semibold mb-2 mt-3;
}
.markdown-body :deep(p) {
  @apply mb-3 leading-relaxed;
}
.markdown-body :deep(ul) {
  @apply list-disc list-inside mb-3 space-y-1;
}
.markdown-body :deep(ol) {
  @apply list-decimal list-inside mb-3 space-y-1;
}
.markdown-body :deep(li) {
  @apply text-gray-700;
}
.markdown-body :deep(strong) {
  @apply font-semibold text-gray-900;
}
.markdown-body :deep(table) {
  @apply w-full border-collapse mb-4;
}
.markdown-body :deep(th) {
  @apply border border-gray-300 px-3 py-2 bg-gray-50 text-left font-semibold;
}
.markdown-body :deep(td) {
  @apply border border-gray-300 px-3 py-2;
}
.markdown-body :deep(blockquote) {
  @apply border-l-4 border-gray-300 pl-4 italic text-gray-600 my-4;
}
.markdown-body :deep(code) {
  @apply bg-gray-100 px-1 py-0.5 rounded text-sm;
}
.markdown-body :deep(pre) {
  @apply bg-gray-100 p-4 rounded-lg overflow-x-auto mb-4;
}
.markdown-body :deep(hr) {
  @apply border-t border-gray-200 my-6;
}
</style>
