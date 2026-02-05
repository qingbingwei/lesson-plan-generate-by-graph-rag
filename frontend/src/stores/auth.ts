import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import type { User } from '@/types';
import * as authApi from '@/api/auth';

export const useAuthStore = defineStore('auth', () => {
  // 状态
  const user = ref<User | null>(null);
  const token = ref<string | null>(null);
  const refreshToken = ref<string | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);

  // 计算属性
  const isAuthenticated = computed(() => !!token.value);
  const isAdmin = computed(() => user.value?.role === 'admin');
  const isTeacher = computed(() => user.value?.role === 'teacher');
  const userName = computed(() => user.value?.profile?.name || user.value?.username || '');

  // 登录
  async function login(username: string, password: string) {
    loading.value = true;
    error.value = null;
    
    try {
      const response = await authApi.login({ username, password });
      token.value = response.access_token;
      refreshToken.value = response.refresh_token;
      user.value = response.user;
    } catch (err) {
      error.value = err instanceof Error ? err.message : '登录失败';
      throw err;
    } finally {
      loading.value = false;
    }
  }

  // 注册
  async function register(username: string, email: string, password: string) {
    loading.value = true;
    error.value = null;
    
    try {
      await authApi.register({ username, email, password });
    } catch (err) {
      error.value = err instanceof Error ? err.message : '注册失败';
      throw err;
    } finally {
      loading.value = false;
    }
  }

  // 刷新 Token
  async function refreshAccessToken() {
    if (!refreshToken.value) {
      throw new Error('No refresh token');
    }
    
    try {
      const response = await authApi.refreshToken(refreshToken.value);
      token.value = response.access_token;
      refreshToken.value = response.refresh_token;
    } catch (err) {
      logout();
      throw err;
    }
  }

  // 获取用户信息
  async function fetchUser() {
    if (!token.value) return;
    
    try {
      user.value = await authApi.getCurrentUser();
    } catch (err) {
      console.error('Failed to fetch user:', err);
    }
  }

  // 更新用户信息
  async function updateUser(data: Partial<User>) {
    try {
      user.value = await authApi.updateProfile(data);
    } catch (err) {
      error.value = err instanceof Error ? err.message : '更新失败';
      throw err;
    }
  }

  // 退出登录
  function logout() {
    user.value = null;
    token.value = null;
    refreshToken.value = null;
    error.value = null;
  }

  // 清除错误
  function clearError() {
    error.value = null;
  }

  return {
    // 状态
    user,
    token,
    refreshToken,
    loading,
    error,
    
    // 计算属性
    isAuthenticated,
    isAdmin,
    isTeacher,
    userName,
    
    // 方法
    login,
    register,
    refreshAccessToken,
    fetchUser,
    updateUser,
    logout,
    clearError,
  };
}, {
  persist: {
    key: 'auth',
    paths: ['token', 'refreshToken', 'user'],
  },
});
