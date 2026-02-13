<script setup lang="ts">
import { computed, nextTick, ref } from 'vue';
import { ElMessage } from 'element-plus';
import {
  Delete,
  Promotion,
  QuestionFilled,
  DocumentCopy,
  InfoFilled,
  ArrowRight,
} from '@element-plus/icons-vue';
import { chatWithAssistant, type AssistantHistoryMessage } from '@/api/assistant';
import MarkdownRenderer from '@/components/common/MarkdownRenderer.vue';
import type { TokenUsage } from '@/types';

type ChatRole = 'user' | 'assistant';

interface ChatMessage {
  id: string;
  role: ChatRole;
  content: string;
  createdAt: string;
  usage?: TokenUsage;
}

const INITIAL_GREETING =
  '你好！我是项目智能助手。你可以问我项目使用方法、教案模板生成、功能说明或常见故障排查。';

const DEFAULT_PROMPTS = [
  '如何快速开始使用这个项目？',
  '给我一份 45 分钟高中数学教案模板',
  '知识图谱和教案生成如何配合使用？',
  '服务启动失败时怎么排查？',
];

const capabilityList = [
  '指导项目启动、登录、知识库上传与教案生成流程',
  '根据学科、年级、课时输出可复制的教案模板',
  '解释知识图谱、历史版本、Token 与 API Key 等功能',
  '提供常见报错与日志定位建议',
];


const UI_TEXT = {
  chatPanelTitle: '问答对话',
  contextMemoryLabel: '上下文记忆',
};

const DEFAULT_CONTEXT_HISTORY_TURNS = 12;
const CONTEXT_MESSAGES_PER_ROUND = 2;
const contextTurnsFromEnv = Number.parseInt(import.meta.env.VITE_ASSISTANT_CONTEXT_TURNS || '', 10);
const contextHistoryTurns = Number.isFinite(contextTurnsFromEnv) && contextTurnsFromEnv > 0
  ? contextTurnsFromEnv
  : DEFAULT_CONTEXT_HISTORY_TURNS;
const contextMessageLimit = contextHistoryTurns * CONTEXT_MESSAGES_PER_ROUND;

const draftQuestion = ref('');
const loading = ref(false);
const suggestedPrompts = ref<string[]>([...DEFAULT_PROMPTS]);
const bottomAnchorRef = ref<HTMLElement | null>(null);

const messages = ref<ChatMessage[]>([
  createMessage('assistant', INITIAL_GREETING),
]);

const latestAssistantAnswer = computed(() => {
  for (let index = messages.value.length - 1; index >= 0; index -= 1) {
    const message = messages.value[index];
    if (message.role === 'assistant' && message.content.trim()) {
      return message.content;
    }
  }

  return '';
});

const contextHistoryMessages = computed(() => messages.value.slice(1));

const currentLoadedRounds = computed(() => {
  const userTurns = contextHistoryMessages.value.filter((message) => message.role === 'user').length;
  return Math.min(userTurns, contextHistoryTurns);
});

const contextMemoryTagText = computed(
  () => `${UI_TEXT.contextMemoryLabel}：已加载 ${currentLoadedRounds.value} 轮 / 最大 ${contextHistoryTurns} 轮`
);

function createMessage(role: ChatRole, content: string, usage?: TokenUsage): ChatMessage {
  return {
    id: `${Date.now()}-${Math.random().toString(36).slice(2, 9)}`,
    role,
    content,
    createdAt: new Date().toISOString(),
    usage,
  };
}

function formatTime(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return '--:--:--';
  }

  return date.toLocaleTimeString('zh-CN', { hour12: false });
}

function formatUsage(usage?: TokenUsage): string {
  if (!usage) {
    return '';
  }

  const promptTokens = usage.promptTokens || 0;
  const completionTokens = usage.completionTokens || 0;
  const totalTokens = usage.totalTokens || promptTokens + completionTokens;
  return `Token ${totalTokens}`;
}

function buildHistory(): AssistantHistoryMessage[] {
  return contextHistoryMessages.value
    .map((message) => ({
      role: message.role,
      content: message.content,
    }))
    .slice(-contextMessageLimit);
}

async function scrollToBottom() {
  await nextTick();
  bottomAnchorRef.value?.scrollIntoView({ behavior: 'smooth', block: 'end' });
}

async function sendQuestion(presetQuestion?: string) {
  const question = (presetQuestion ?? draftQuestion.value).trim();
  if (!question) {
    ElMessage.warning('请输入问题后再发送');
    return;
  }

  if (loading.value) {
    return;
  }

  const history = buildHistory();
  messages.value.push(createMessage('user', question));
  draftQuestion.value = '';
  await scrollToBottom();

  loading.value = true;
  try {
    const response = await chatWithAssistant({
      question,
      history,
    });

    const answer = response.answer?.trim() || '暂未生成回答，请稍后重试。';
    messages.value.push(createMessage('assistant', answer, response.usage));

    if (response.suggestions && response.suggestions.length > 0) {
      suggestedPrompts.value = response.suggestions.slice(0, 4);
    }
  } catch (error) {
    const message = error instanceof Error ? error.message : '智能问答请求失败';
    ElMessage.error(message);
    messages.value.push(
      createMessage('assistant', '抱歉，当前问答服务暂时不可用。请稍后重试，或先查看帮助文档页面。')
    );
  } finally {
    loading.value = false;
    await scrollToBottom();
  }
}

function clearConversation() {
  messages.value = [createMessage('assistant', INITIAL_GREETING)];
  suggestedPrompts.value = [...DEFAULT_PROMPTS];
  draftQuestion.value = '';
}

async function copyLatestAnswer() {
  const content = latestAssistantAnswer.value.trim();
  if (!content) {
    ElMessage.warning('暂无可复制的回答内容');
    return;
  }

  try {
    await navigator.clipboard.writeText(content);
    ElMessage.success('最新回答已复制');
  } catch {
    ElMessage.error('复制失败，请手动复制');
  }
}
</script>

<template>
  <div class="page-container">
    <div class="page-header flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
      <div>
        <h1 class="page-title">智能 AI 问答</h1>
        <p class="page-subtitle">基于 DeepSeek，支持项目使用指导、教案模板生成与常见问题排查</p>
      </div>
      <el-space>
        <el-button :icon="DocumentCopy" @click="copyLatestAnswer">复制最新回答</el-button>
        <el-button :icon="Delete" @click="clearConversation">清空会话</el-button>
      </el-space>
    </div>

    <el-row :gutter="16">
      <el-col :xs="24" :lg="16">
        <el-card class="surface-card" shadow="never">
          <template #header>
            <div class="flex items-center justify-between gap-3">
              <div class="font-semibold">{{ UI_TEXT.chatPanelTitle }}</div>
              <el-tag type="info" effect="plain">{{ contextMemoryTagText }}</el-tag>
            </div>
          </template>

          <el-scrollbar height="560px" class="chat-scroll">
            <div class="flex flex-col gap-3">
              <div
                v-for="message in messages"
                :key="message.id"
                class="flex"
                :class="message.role === 'user' ? 'justify-end' : 'justify-start'"
              >
                <el-card
                  class="chat-bubble w-full max-w-[92%]"
                  :class="message.role === 'user' ? 'chat-user' : 'chat-assistant'"
                  shadow="never"
                >
                  <div class="mb-2 flex items-center gap-2">
                    <el-tag size="small" :type="message.role === 'user' ? 'primary' : 'success'" effect="plain">
                      {{ message.role === 'user' ? '你' : 'AI 助手' }}
                    </el-tag>
                    <span class="text-xs app-text-muted">{{ formatTime(message.createdAt) }}</span>
                    <span v-if="formatUsage(message.usage)" class="text-xs app-text-muted">
                      {{ formatUsage(message.usage) }}
                    </span>
                  </div>

                  <MarkdownRenderer v-if="message.role === 'assistant'" :content="message.content" />
                  <div v-else class="text-sm leading-7 whitespace-pre-wrap app-text-primary">{{ message.content }}</div>
                </el-card>
              </div>

              <div v-if="loading" class="flex justify-start">
                <el-card class="chat-bubble chat-assistant w-full max-w-[92%]" shadow="never">
                  <div class="flex items-center gap-2 app-text-muted">
                    <el-icon class="is-loading"><Promotion /></el-icon>
                    <span class="text-sm">正在思考中...</span>
                  </div>
                </el-card>
              </div>

              <div ref="bottomAnchorRef" />
            </div>
          </el-scrollbar>

          <el-divider />

          <div class="flex flex-col gap-3">
            <el-input
              v-model="draftQuestion"
              type="textarea"
              :rows="3"
              maxlength="1200"
              show-word-limit
              resize="none"
              placeholder="请输入你的问题，例如：给我一个初二英语阅读课教案模板，包含分层作业设计。"
              @keydown.enter.exact.prevent="sendQuestion()"
            />

            <div class="flex items-center justify-between gap-3">
              <span class="text-xs app-text-muted">提示：按 Enter 发送；Shift + Enter 换行</span>
              <el-button type="primary" :icon="Promotion" :loading="loading" @click="sendQuestion()">
                发送问题
              </el-button>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :xs="24" :lg="8">
        <el-card class="surface-card assistant-side-card mb-4" shadow="never">
          <template #header>
            <div class="assistant-card-header">
              <div class="assistant-header-main">
                <el-icon><QuestionFilled /></el-icon>
                <span class="font-semibold">快捷提问</span>
              </div>
              <el-tag size="small" effect="plain" type="info">一键发送</el-tag>
            </div>
          </template>

          <div class="quick-prompt-list">
            <el-button
              v-for="(prompt, index) in suggestedPrompts"
              :key="prompt"
              class="quick-prompt-btn"
              :disabled="loading"
              @click="sendQuestion(prompt)"
            >
              <span class="quick-prompt-main">
                <span class="quick-prompt-index">{{ index + 1 }}</span>
                <span class="quick-prompt-text">{{ prompt }}</span>
              </span>
              <el-icon class="quick-prompt-arrow"><ArrowRight /></el-icon>
            </el-button>
          </div>
        </el-card>

        <el-card class="surface-card assistant-side-card" shadow="never">
          <template #header>
            <div class="assistant-card-header">
              <div class="assistant-header-main">
                <el-icon><InfoFilled /></el-icon>
                <span class="font-semibold">可回答内容</span>
              </div>
            </div>
          </template>

          <div class="capability-list">
            <div v-for="(item, index) in capabilityList" :key="item" class="capability-item">
              <span class="capability-index">{{ index + 1 }}</span>
              <span class="capability-text">{{ item }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
.chat-scroll :deep(.el-scrollbar__view) {
  padding-right: 4px;
}

.chat-bubble {
  border: 1px solid var(--app-surface-border);
  background: color-mix(in srgb, var(--el-bg-color) 92%, transparent);
}

.chat-user {
  background: color-mix(in srgb, var(--el-color-primary) 12%, var(--el-bg-color));
}

.chat-assistant {
  background: color-mix(in srgb, var(--el-fill-color-light) 66%, var(--el-bg-color));
}

.assistant-side-card :deep(.el-card__header) {
  padding: 14px 16px;
}

.assistant-side-card :deep(.el-card__body) {
  padding: 14px 16px;
}

.assistant-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.assistant-header-main {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: var(--app-text-primary);
}

.quick-prompt-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.quick-prompt-list :deep(.el-button + .el-button) {
  margin-left: 0;
}

.quick-prompt-btn {
  width: 100%;
  height: auto;
  margin: 0;
  padding: 11px 12px;
  border-radius: 14px;
  border: 1px solid var(--app-surface-border);
  background: color-mix(in srgb, var(--el-fill-color-light) 38%, var(--el-bg-color));
  color: var(--app-text-primary);
  display: flex;
  align-items: center;
  justify-content: space-between;
  transition: all 0.2s ease;
}

.quick-prompt-btn:hover {
  border-color: color-mix(in srgb, var(--el-color-primary) 55%, var(--app-surface-border));
  background: color-mix(in srgb, var(--el-color-primary) 8%, var(--el-bg-color));
  transform: translateY(-1px);
}

.quick-prompt-btn:disabled {
  opacity: 0.7;
}

.quick-prompt-main {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.quick-prompt-index {
  width: 20px;
  height: 20px;
  border-radius: 9999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 600;
  color: var(--el-color-primary);
  background: color-mix(in srgb, var(--el-color-primary) 14%, transparent);
  flex-shrink: 0;
}

.quick-prompt-text {
  text-align: left;
  line-height: 1.4;
  white-space: normal;
  word-break: break-word;
  color: var(--app-text-primary);
}

.quick-prompt-arrow {
  color: var(--app-text-muted);
  flex-shrink: 0;
}

.capability-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.capability-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  border-radius: 12px;
  border: 1px solid color-mix(in srgb, var(--app-surface-border) 65%, transparent);
  background: color-mix(in srgb, var(--el-fill-color-light) 35%, var(--el-bg-color));
  padding: 10px 12px;
}

.capability-index {
  width: 22px;
  height: 22px;
  border-radius: 9999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 700;
  color: var(--el-color-primary);
  background: color-mix(in srgb, var(--el-color-primary) 16%, transparent);
  flex-shrink: 0;
}

.capability-text {
  color: var(--app-text-secondary);
  line-height: 1.55;
}
</style>
