<script setup lang="ts">
import { useRouter } from 'vue-router';
import { Warning } from '@element-plus/icons-vue';

const router = useRouter();
</script>

<template>
  <div class="min-h-screen flex items-center justify-center px-4 py-10">
    <el-card class="surface-card w-full max-w-lg" shadow="never">
      <div class="text-center py-6">
        <el-icon :size="56" class="app-icon-warning">
          <Warning />
        </el-icon>
        <h1 class="mt-3 text-4xl font-bold app-text-primary">404</h1>
        <p class="mt-2 text-lg app-text-secondary">页面未找到</p>
        <p class="mt-1 text-sm app-text-muted">您访问的页面不存在或已被移除</p>
        <div class="mt-6 flex items-center justify-center gap-2">
          <el-button type="primary" @click="router.push('/')">返回首页</el-button>
          <el-button @click="router.back()">返回上一页</el-button>
        </div>
      </div>
    </el-card>
  </div>
</template>
