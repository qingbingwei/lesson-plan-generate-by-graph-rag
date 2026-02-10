<script setup lang="ts">
withDefaults(
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
</script>

<template>
  <el-pagination
    background
    :current-page="currentPage"
    :page-size="1"
    :total="totalPages"
    :pager-count="showPages"
    layout="prev, pager, next"
    @current-change="emit('update:currentPage', $event)"
  />
</template>
