<script setup lang="ts">
import { ref, computed } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useAuthStore } from '@/stores/auth';

const router = useRouter();
const route = useRoute();
const authStore = useAuthStore();

const form = ref({
  account: '',
  password: '',
});

const isLoading = computed(() => authStore.loading);
const error = computed(() => authStore.error);

async function handleSubmit() {
  authStore.clearError();

  try {
    await authStore.login(form.value.account, form.value.password);
    const redirect = route.query.redirect as string;
    router.push(redirect || '/dashboard');
  } catch {
    // 错误已经在 store 中处理
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center px-4 py-10">
    <el-card class="surface-card w-full max-w-md" shadow="never">
      <div class="text-center mb-8">
        <div class="text-xs font-semibold tracking-[0.2em] uppercase text-primary-600">Hero Classroom</div>
        <h1 class="mt-2 text-3xl font-bold app-text-primary">智能教案生成系统</h1>
        <p class="mt-2 text-sm app-text-muted">登录您的账户</p>
      </div>

      <el-alert v-if="error" :title="error" type="error" show-icon class="mb-4" />

      <el-form :model="form" label-position="top" @submit.prevent="handleSubmit">
        <el-form-item label="用户名 / 邮箱">
          <el-input v-model="form.account" placeholder="请输入用户名或邮箱" clearable />
        </el-form-item>

        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" placeholder="请输入密码" show-password clearable />
        </el-form-item>

        <el-button type="primary" class="w-full" :loading="isLoading" @click="handleSubmit">
          登录
        </el-button>
      </el-form>

      <div class="mt-6 text-center">
        <el-text type="info">还没有账户？</el-text>
        <RouterLink to="/register" class="ml-1 text-sm font-semibold text-primary-600 hover:text-primary-500">
          立即注册
        </RouterLink>
      </div>
    </el-card>
  </div>
</template>
