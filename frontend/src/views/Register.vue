<script setup lang="ts">
import { ref, computed } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import { EyeIcon, EyeSlashIcon } from '@heroicons/vue/24/outline';

const router = useRouter();
const authStore = useAuthStore();

const form = ref({
  username: '',
  email: '',
  password: '',
  confirmPassword: '',
});

const showPassword = ref(false);
const showConfirmPassword = ref(false);
const isLoading = computed(() => authStore.loading);
const error = computed(() => authStore.error);
const localError = ref('');

async function handleSubmit() {
  authStore.clearError();
  localError.value = '';

  // 验证密码
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
    
    // 注册成功，跳转到登录页
    router.push('/login');
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
            <p class="mt-2 text-sm text-gray-600">创建新账户</p>
          </div>

          <!-- Error message -->
          <div
            v-if="error || localError"
            class="mb-4 p-4 rounded-lg bg-red-50 text-red-700 text-sm"
          >
            {{ error || localError }}
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
              <label for="email" class="label">邮箱</label>
              <input
                id="email"
                v-model="form.email"
                type="email"
                required
                class="input"
                placeholder="请输入邮箱"
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
                  placeholder="请输入密码（至少6位）"
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

            <div>
              <label for="confirmPassword" class="label">确认密码</label>
              <div class="relative">
                <input
                  id="confirmPassword"
                  v-model="form.confirmPassword"
                  :type="showConfirmPassword ? 'text' : 'password'"
                  required
                  class="input pr-10"
                  placeholder="请再次输入密码"
                />
                <button
                  type="button"
                  class="absolute inset-y-0 right-0 flex items-center pr-3"
                  @click="showConfirmPassword = !showConfirmPassword"
                >
                  <EyeSlashIcon v-if="showConfirmPassword" class="h-5 w-5 text-gray-400" />
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
                注册中...
              </span>
              <span v-else>注册</span>
            </button>
          </form>

          <!-- Login link -->
          <div class="mt-6 text-center">
            <span class="text-sm text-gray-600">已有账户？</span>
            <RouterLink to="/login" class="text-sm font-medium text-primary-600 hover:text-primary-500 ml-1">
              立即登录
            </RouterLink>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
