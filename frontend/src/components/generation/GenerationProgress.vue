<script setup lang="ts">
import { computed } from 'vue';
import type { GenerationProgress } from '@/types';
import {
  CheckCircleIcon,
  ClockIcon,
  ExclamationCircleIcon,
} from '@heroicons/vue/24/solid';

const props = defineProps<{
  progress: GenerationProgress;
}>();

const nodes = [
  { key: 'inputAnalysis', label: '输入分析' },
  { key: 'knowledgeQuery', label: '知识检索' },
  { key: 'objectiveDesign', label: '目标设计' },
  { key: 'contentDesign', label: '内容设计' },
  { key: 'activityDesign', label: '活动设计' },
  { key: 'outputFormat', label: '输出格式化' },
];

const progressPercent = computed(() => {
  const nodesData = props.progress.nodes || {};
  const completed = Object.values(nodesData).filter(
    (n) => n.status === 'completed'
  ).length;
  return Math.round((completed / nodes.length) * 100);
});

function getNodeStatus(key: string) {
  return props.progress.nodes?.[key]?.status || 'pending';
}

function getNodeIcon(status: string) {
  switch (status) {
    case 'completed':
      return CheckCircleIcon;
    case 'running':
      return ClockIcon;
    case 'error':
      return ExclamationCircleIcon;
    default:
      return ClockIcon;
  }
}

function getNodeClass(status: string) {
  switch (status) {
    case 'completed':
      return 'text-green-500';
    case 'running':
      return 'text-blue-500 animate-pulse';
    case 'error':
      return 'text-red-500';
    default:
      return 'text-gray-300';
  }
}
</script>

<template>
  <div class="space-y-4">
    <!-- Progress bar -->
    <div>
      <div class="flex items-center justify-between text-sm mb-2">
        <span class="font-medium text-gray-700">生成进度</span>
        <span class="text-gray-500">{{ progressPercent }}%</span>
      </div>
      <div class="h-2 bg-gray-200 rounded-full overflow-hidden">
        <div
          class="h-full bg-primary-600 transition-all duration-500"
          :style="{ width: `${progressPercent}%` }"
        />
      </div>
    </div>

    <!-- Nodes -->
    <div class="grid grid-cols-2 sm:grid-cols-3 gap-3">
      <div
        v-for="node in nodes"
        :key="node.key"
        class="flex items-center gap-2 p-3 bg-gray-50 rounded-lg"
      >
        <component
          :is="getNodeIcon(getNodeStatus(node.key))"
          class="h-5 w-5 flex-shrink-0"
          :class="getNodeClass(getNodeStatus(node.key))"
        />
        <span class="text-sm text-gray-700">{{ node.label }}</span>
      </div>
    </div>

    <!-- Current status -->
    <div v-if="progress.currentNode" class="text-sm text-gray-500 text-center">
      正在处理: {{ nodes.find(n => n.key === progress.currentNode)?.label }}
    </div>
  </div>
</template>
