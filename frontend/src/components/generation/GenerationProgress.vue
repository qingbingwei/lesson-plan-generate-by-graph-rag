<script setup lang="ts">
import { computed } from 'vue';
import type { GenerationProgress } from '@/types';

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

function getNodeType(status: string): 'success' | 'primary' | 'warning' | 'info' {
  switch (status) {
    case 'completed':
      return 'success';
    case 'running':
      return 'primary';
    case 'error':
      return 'warning';
    default:
      return 'info';
  }
}
</script>

<template>
  <div class="space-y-4">
    <el-progress :percentage="progressPercent" :stroke-width="10" striped striped-flow />

    <el-steps :active="Math.max(0, Math.ceil((progressPercent / 100) * nodes.length) - 1)" finish-status="success" align-center>
      <el-step
        v-for="node in nodes"
        :key="node.key"
        :title="node.label"
        :status="getNodeStatus(node.key) === 'error' ? 'error' : getNodeStatus(node.key) === 'completed' ? 'success' : getNodeStatus(node.key) === 'running' ? 'process' : 'wait'"
      />
    </el-steps>

    <div class="flex flex-wrap gap-2">
      <el-tag
        v-for="node in nodes"
        :key="node.key"
        :type="getNodeType(getNodeStatus(node.key))"
        effect="light"
        round
      >
        {{ node.label }}
      </el-tag>
    </div>

    <el-text v-if="progress.currentNode" type="info">
      正在处理：{{ nodes.find(n => n.key === progress.currentNode)?.label }}
    </el-text>
  </div>
</template>
