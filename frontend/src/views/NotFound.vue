<script setup lang="ts">
import { RouterLink } from 'vue-router';
import { ExclamationTriangleIcon } from '@heroicons/vue/24/outline';
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50 px-4">
    <div class="text-center">
      <ExclamationTriangleIcon class="mx-auto h-16 w-16 text-yellow-500" />
      <h1 class="mt-4 text-4xl font-bold text-gray-900">404</h1>
      <p class="mt-2 text-lg text-gray-600">页面未找到</p>
      <p class="mt-1 text-sm text-gray-500">
        您访问的页面不存在或已被移除
      </p>
      <div class="mt-6 flex items-center justify-center gap-4">
        <RouterLink to="/" class="btn-primary">
          返回首页
        </RouterLink>
        <button
          type="button"
          class="btn-outline"
          @click="$router.back()"
        >
          返回上一页
        </button>
      </div>
    </div>
  </div>
</template>
