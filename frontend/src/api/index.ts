import axios, { AxiosInstance, AxiosError, InternalAxiosRequestConfig, AxiosResponse } from 'axios';
import { useAuthStore } from '@/stores/auth';
import type { ApiResponse } from '@/types';

function normalizeApiBaseUrl(rawValue?: string): string {
  const fallback = '/api/v1';
  if (!rawValue) {
    return fallback;
  }

  const value = rawValue.trim().replace(/\/+$/, '');
  if (!value) {
    return fallback;
  }

  if (/\/api\/v1$/i.test(value)) {
    return value;
  }

  if (/\/api$/i.test(value)) {
    return `${value}/v1`;
  }

  return `${value}/api/v1`;
}

const resolvedApiBaseUrl = normalizeApiBaseUrl(import.meta.env.VITE_API_BASE_URL);

// 创建 axios 实例
const api: AxiosInstance = axios.create({
  baseURL: resolvedApiBaseUrl,
  timeout: 600000, // 10分钟超时，生成教案可能需要较长时间
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器
api.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const authStore = useAuthStore();
    
    if (authStore.token) {
      config.headers.Authorization = `Bearer ${authStore.token}`;
    }
    
    return config;
  },
  (error: AxiosError) => {
    return Promise.reject(error);
  }
);

// 响应拦截器
// 标记是否正在刷新 token
let isRefreshing = false;
let refreshSubscribers: Array<(token: string) => void> = [];

function subscribeTokenRefresh(cb: (token: string) => void) {
  refreshSubscribers.push(cb);
}

function onRefreshed(token: string) {
  refreshSubscribers.forEach(cb => cb(token));
  refreshSubscribers = [];
}

api.interceptors.response.use(
  (response: AxiosResponse) => {
    return response;
  },
  async (error: AxiosError<ApiResponse>) => {
    const authStore = useAuthStore();
    const originalRequest = error.config;
    
    if (error.response?.status === 401 && originalRequest) {
      // 如果已经在刷新 token，将请求加入队列等待
      if (isRefreshing) {
        return new Promise((resolve) => {
          subscribeTokenRefresh((token: string) => {
            originalRequest.headers.Authorization = `Bearer ${token}`;
            resolve(api(originalRequest));
          });
        });
      }
      
      // Token 过期，尝试刷新
      if (authStore.refreshToken) {
        isRefreshing = true;
        try {
          await authStore.refreshAccessToken();
          isRefreshing = false;
          onRefreshed(authStore.token!);
          // 重试原请求
          originalRequest.headers.Authorization = `Bearer ${authStore.token}`;
          return api(originalRequest);
        } catch {
          isRefreshing = false;
          refreshSubscribers = [];
          // 刷新失败，退出登录但不立即跳转
          authStore.logout();
          // 使用 router 而不是 location.href，避免闪烁
          window.location.href = '/login';
        }
      } else {
        // 没有 refresh token，只是静默登出，不强制跳转
        // 由路由守卫处理跳转
        authStore.logout();
      }
    }
    
    return Promise.reject(error);
  }
);

export default api;

// Knowledge API
export { knowledgeApi } from './knowledge';

// Agent API
export const agentApi: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_AGENT_BASE_URL || '/agent',
  timeout: 600000, // 10分钟超时，Agent生成请求需要较长时间
  headers: {
    'Content-Type': 'application/json',
  },
});

agentApi.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const authStore = useAuthStore();
    
    if (authStore.token) {
      config.headers.Authorization = `Bearer ${authStore.token}`;
    }
    
    return config;
  },
  (error: AxiosError) => {
    return Promise.reject(error);
  }
);
