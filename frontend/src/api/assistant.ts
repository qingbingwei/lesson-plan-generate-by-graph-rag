import axios from 'axios';
import api, { agentApi } from './index';
import type { ApiResponse, TokenUsage } from '@/types';
import { useAuthStore } from '@/stores/auth';

export type AssistantRole = 'user' | 'assistant';

export interface AssistantHistoryMessage {
  role: AssistantRole;
  content: string;
}

export interface AssistantChatRequest {
  question: string;
  history?: AssistantHistoryMessage[];
}

export interface AssistantChatPayload {
  answer: string;
  suggestions?: string[];
  usage?: TokenUsage;
}

interface AgentAssistantResponse {
  success?: boolean;
  error?: string;
  data?: AssistantChatPayload;
  usage?: TokenUsage;
}

function shouldFallbackToAgent(error: unknown): boolean {
  if (!axios.isAxiosError(error)) {
    return true;
  }

  const statusCode = error.response?.status;
  return statusCode === 404 || (typeof statusCode === 'number' && statusCode >= 500);
}



function buildAgentAssistantPaths(): string[] {
  const baseURL = String(agentApi.defaults.baseURL || '').trim();

  const preferShortPath = /\/agent\/?$/i.test(baseURL) || /\/api\/?$/i.test(baseURL);
  const first = preferShortPath ? '/assistant/chat' : '/api/assistant/chat';
  const second = preferShortPath ? '/api/assistant/chat' : '/assistant/chat';

  return [first, second].filter((path, index, array) => array.indexOf(path) === index);
}

function normalizePayload(payload: AssistantChatPayload | undefined): AssistantChatPayload {
  if (!payload) {
    return {
      answer: '',
      suggestions: [],
    };
  }

  return {
    answer: payload.answer || '',
    suggestions: payload.suggestions || [],
    usage: payload.usage,
  };
}

export async function chatWithAssistant(payload: AssistantChatRequest): Promise<AssistantChatPayload> {
  try {
    const response = await api.post<ApiResponse<AssistantChatPayload>>('/generate/assistant/chat', payload);
    return normalizePayload(response.data.data);
  } catch (error) {
    if (!shouldFallbackToAgent(error)) {
      throw error;
    }

    const authStore = useAuthStore();
    const userId = authStore.user?.id;

    const fallbackPayload = {
      question: payload.question,
      history: payload.history || [],
      userId: userId ? String(userId) : undefined,
    };

    const fallbackPaths = buildAgentAssistantPaths();
    let lastError: unknown = null;

    for (const path of fallbackPaths) {
      try {
        const fallback = await agentApi.post<AgentAssistantResponse>(path, fallbackPayload);

        if (fallback.data && fallback.data.success === false) {
          throw new Error(fallback.data.error || '智能问答请求失败');
        }

        const normalized = normalizePayload(fallback.data?.data);

        if (fallback.data?.usage) {
          normalized.usage = fallback.data.usage;
        }

        return normalized;
      } catch (fallbackError) {
        lastError = fallbackError;

        const fallbackStatus = axios.isAxiosError(fallbackError) ? fallbackError.response?.status : null;
        const shouldTryNextPath = fallbackStatus === 404;
        if (!shouldTryNextPath) {
          throw fallbackError;
        }
      }
    }

    throw lastError || new Error('智能问答请求失败');
  }
}
