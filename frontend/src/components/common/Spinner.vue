<script setup lang="ts">
import { Loading } from '@element-plus/icons-vue';

withDefaults(
  defineProps<{
    size?: 'sm' | 'md' | 'lg';
  }>(),
  {
    size: 'md',
  }
);

const sizeMap: Record<'sm' | 'md' | 'lg', string> = {
  sm: '16px',
  md: '22px',
  lg: '30px',
};
</script>

<template>
  <el-icon class="is-loading" :size="sizeMap[size]">
    <Loading />
  </el-icon>
</template>
