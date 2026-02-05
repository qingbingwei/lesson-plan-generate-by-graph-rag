<script setup lang="ts">
import { ref, computed } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import { EyeIcon, EyeSlashIcon } from '@heroicons/vue/24/outline';

const router = useRouter();
const route = useRoute();
const authStore = useAuthStore();

const form = ref({
  username: '',
  password: '',
});

const showPassword = ref(false);
const isLoading = computed(() => authStore.loading);
const error = computed(() => authStore.error);

async function handleSubmit() {
  authStore.clearError();
  
  try {
    await authStore.login(form.value.username, form.value.password);
    
    // 跳转到之前的页面或首页
    const redirect = route.query.redirect as string;
    router.push(redirect || '/dashboard');
  } catch {
    // 错误已经在 store 中处理
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-primary-50 to-secondary-50 py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full">
      <div class="card">
        <div class="card-body p-8">
          <!-- Logo -->
          <div class="text-center mb-8">
            <h1 class="text-3xl font-bold text-gradient">智能教案生成系统</h1>
            <p class="mt-2 text-sm text-gray-600">登录您的账户</p>
          </div>

          <!-- Error message -->
          <div
            v-if="error"
            class="mb-4 p-4 rounded-lg bg-red-50 text-red-700 text-sm"
          >
            {{ error }}
          </div>

          <!-- Form -->
          <form @submit.prevent="handleSubmit" class="space-y-6">
            <div>
              <label for="username" class="label">用户名</label>
              <input
                id="username"
                v-model="form.username"
                type="text"
                required
                class="input"
                placeholder="请输入用户名"
              />
            </div>

            <div>
              <label for="password" class="label">密码</label>
              <div class="relative">
                <input
                  id="password"
                  v-model="form.password"
                  :type="showPassword ? 'text' : 'password'"
                  required
                  class="input pr-10"
                  placeholder="请输入密码"
                />
                <button
                  type="button"
                  class="absolute inset-y-0 right-0 flex items-center pr-3"
                  @click="showPassword = !showPassword"
                >
                  <EyeSlashIcon v-if="showPassword" class="h-5 w-5 text-gray-400" />
                  <EyeIcon v-else class="h-5 w-5 text-gray-400" />
                </button>
              </div>
            </div>

            <button
              type="submit"
              :disabled="isLoading"
              class="btn-primary w-full"
            >
              <span v-if="isLoading" class="flex items-center justify-center gap-2">
                <svg class="loading h-4 w-4" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none" />
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
                登录中...
              </span>
              <span v-else>登录</span>
            </button>
          </form>

          <!-- Register link -->
          <div class="mt-6 text-center">
            <span class="text-sm text-gray-600">还没有账户？</span>
            <RouterLink to="/register" class="text-sm font-medium text-primary-600 hover:text-primary-500 ml-1">
              立即注册
            </RouterLink>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
