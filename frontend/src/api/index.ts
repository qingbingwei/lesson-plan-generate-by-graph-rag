import axios, { AxiosInstance, AxiosError, InternalAxiosRequestConfig, AxiosResponse, AxiosHeaders } from 'axios';
import { useAuthStore } from '@/stores/auth';
import { getApiKeyHeaders } from '@/utils/apiKeys';
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
const DEFAULT_TIMEOUT_MS = 600000; // 10分钟
const MAX_RETRY_ATTEMPTS = 2;
const RETRY_BASE_DELAY_MS = 250;
const TRACE_ID_HEADER = 'X-Trace-ID';
const REQUEST_ID_HEADER = 'X-Request-ID';

type RetryRequestConfig = InternalAxiosRequestConfig & {
  _retryCount?: number;
  _traceId?: string;
};

export function createTraceId(): string {
  if (typeof globalThis.crypto !== 'undefined' && typeof globalThis.crypto.randomUUID === 'function') {
    return globalThis.crypto.randomUUID();
  }

  return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`;
}

export function buildTraceHeaders(traceId?: string): Record<string, string> {
  const value = (traceId || createTraceId()).trim();
  return {
    [TRACE_ID_HEADER]: value,
    [REQUEST_ID_HEADER]: value,
  };
}

function readHeader(config: RetryRequestConfig, headerName: string): string | undefined {
  if (!config.headers) {
    return undefined;
  }

  const headers = AxiosHeaders.from(config.headers);
  const value = headers.get(headerName);
  if (typeof value === 'string' && value.trim()) {
    return value.trim();
  }

  return undefined;
}

function writeHeader(config: RetryRequestConfig, headerName: string, value: string) {
  const headers = AxiosHeaders.from(config.headers);
  headers.set(headerName, value);
  config.headers = headers;
}

function isRetryableMethod(method?: string): boolean {
  const normalized = (method || 'GET').toUpperCase();
  return normalized === 'GET' || normalized === 'HEAD' || normalized === 'OPTIONS' || normalized === 'PUT' || normalized === 'DELETE';
}

function parseRetryAfterToMs(value: string | undefined): number | null {
  if (!value) {
    return null;
  }

  const trimmed = value.trim();
  if (!trimmed) {
    return null;
  }

  const seconds = Number.parseInt(trimmed, 10);
  if (Number.isFinite(seconds) && seconds >= 0) {
    return seconds * 1000;
  }

  const timestamp = Date.parse(trimmed);
  if (!Number.isFinite(timestamp)) {
    return null;
  }

  return Math.max(0, timestamp - Date.now());
}

function getRetryAfterHeader(error: AxiosError): string | undefined {
  const header = error.response?.headers?.['retry-after'];
  if (Array.isArray(header)) {
    return header[0];
  }
  return header;
}

function getRetryDelayMs(error: AxiosError<ApiResponse>, nextAttempt: number): number {
  const retryAfterMs = parseRetryAfterToMs(getRetryAfterHeader(error));
  if (retryAfterMs !== null) {
    return Math.min(5000, retryAfterMs);
  }

  return RETRY_BASE_DELAY_MS * (2 ** Math.max(0, nextAttempt - 1));
}

function wait(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function shouldRetry(error: AxiosError<ApiResponse>, config?: RetryRequestConfig): boolean {
  if (!config) {
    return false;
  }

  if (error.code === 'ERR_CANCELED') {
    return false;
  }

  if (!isRetryableMethod(config.method)) {
    return false;
  }

  const retries = config._retryCount ?? 0;
  if (retries >= MAX_RETRY_ATTEMPTS) {
    return false;
  }

  const status = error.response?.status;
  if (typeof status === 'number') {
    return status === 429 || status >= 500;
  }

  return true;
}

function applyRequestDefaults(config: RetryRequestConfig): RetryRequestConfig {
  const authStore = useAuthStore();

  if (authStore.token) {
    writeHeader(config, 'Authorization', `Bearer ${authStore.token}`);
  }

  const apiKeyHeaders = getApiKeyHeaders();
  Object.entries(apiKeyHeaders).forEach(([key, value]) => {
    writeHeader(config, key, value);
  });

  if (typeof config.timeout !== 'number' || config.timeout <= 0) {
    config.timeout = DEFAULT_TIMEOUT_MS;
  }
  config._retryCount = config._retryCount ?? 0;

  const traceId = readHeader(config, TRACE_ID_HEADER) || config._traceId || createTraceId();
  config._traceId = traceId;
  writeHeader(config, TRACE_ID_HEADER, traceId);
  writeHeader(config, REQUEST_ID_HEADER, traceId);

  return config;
}

async function retryRequest(instance: AxiosInstance, error: AxiosError<ApiResponse>): Promise<AxiosResponse> {
  const originalRequest = error.config as RetryRequestConfig | undefined;
  if (!originalRequest) {
    return Promise.reject(error);
  }

  if (!shouldRetry(error, originalRequest)) {
    return Promise.reject(error);
  }

  const nextAttempt = (originalRequest._retryCount ?? 0) + 1;
  originalRequest._retryCount = nextAttempt;
  await wait(getRetryDelayMs(error, nextAttempt));

  return instance(originalRequest);
}

// 创建 axios 实例
const api: AxiosInstance = axios.create({
  baseURL: resolvedApiBaseUrl,
  timeout: DEFAULT_TIMEOUT_MS,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器
api.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    return applyRequestDefaults(config as RetryRequestConfig);
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
    const originalRequest = error.config as RetryRequestConfig | undefined;

    if (error.response?.status === 401 && originalRequest) {
      const requestUrl = originalRequest.url || '';
      if (requestUrl.includes('/auth/refresh')) {
        authStore.logout();
        return Promise.reject(error);
      }

      // 如果已经在刷新 token，将请求加入队列等待
      if (isRefreshing) {
        return new Promise((resolve) => {
          subscribeTokenRefresh((token: string) => {
            writeHeader(originalRequest, 'Authorization', `Bearer ${token}`);
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
          writeHeader(originalRequest, 'Authorization', `Bearer ${authStore.token}`);
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

    return retryRequest(api, error);
  }
);

export default api;

// Knowledge API
export { knowledgeApi } from './knowledge';

// Agent API
export const agentApi: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_AGENT_BASE_URL || '/agent',
  timeout: DEFAULT_TIMEOUT_MS,
  headers: {
    'Content-Type': 'application/json',
  },
});

agentApi.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    return applyRequestDefaults(config as RetryRequestConfig);
  },
  (error: AxiosError) => {
    return Promise.reject(error);
  }
);

agentApi.interceptors.response.use(
  (response: AxiosResponse) => response,
  async (error: AxiosError<ApiResponse>) => retryRequest(agentApi, error)
);
