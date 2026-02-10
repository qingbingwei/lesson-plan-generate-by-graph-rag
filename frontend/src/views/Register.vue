<script setup lang="ts">
import { ref, computed } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';

const router = useRouter();
const authStore = useAuthStore();

const form = ref({
  username: '',
  email: '',
  password: '',
  confirmPassword: '',
});

const isLoading = computed(() => authStore.loading);
const error = computed(() => authStore.error);
const localError = ref('');

async function handleSubmit() {
  authStore.clearError();
  localError.value = '';

  if (form.value.password !== form.value.confirmPassword) {
    localError.value = '两次输入的密码不一致';
    return;
  }

  if (form.value.password.length < 6) {
    localError.value = '密码长度至少为6位';
    return;
  }

  try {
    await authStore.register(
      form.value.username,
      form.value.email,
      form.value.password
    );

    router.push('/login');
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
        <h1 class="mt-2 text-3xl font-bold app-text-primary">创建账户</h1>
        <p class="mt-2 text-sm app-text-muted">开启智能备课体验</p>
      </div>

      <el-alert v-if="error || localError" :title="error || localError" type="error" show-icon class="mb-4" />

      <el-form :model="form" label-position="top" @submit.prevent="handleSubmit">
        <el-form-item label="用户名">
          <el-input v-model="form.username" placeholder="请输入用户名" clearable />
        </el-form-item>

        <el-form-item label="邮箱">
          <el-input v-model="form.email" placeholder="请输入邮箱" clearable />
        </el-form-item>

        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" placeholder="至少6位" show-password clearable />
        </el-form-item>

        <el-form-item label="确认密码">
          <el-input v-model="form.confirmPassword" type="password" placeholder="请再次输入密码" show-password clearable />
        </el-form-item>

        <el-button type="primary" class="w-full" :loading="isLoading" @click="handleSubmit">
          注册
        </el-button>
      </el-form>

      <div class="mt-6 text-center">
        <el-text type="info">已有账户？</el-text>
        <RouterLink to="/login" class="ml-1 text-sm font-semibold text-primary-600 hover:text-primary-500">
          立即登录
        </RouterLink>
      </div>
    </el-card>
  </div>
</template>
