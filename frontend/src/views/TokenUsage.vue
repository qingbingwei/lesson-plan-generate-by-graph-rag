<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import axios from 'axios';
import { ElMessage } from 'element-plus';
import {
  DataAnalysis,
  Timer,
  Warning,
  RefreshRight,
  Key,
  CircleCheck,
} from '@element-plus/icons-vue';
import {
  getLangSmithUsage,
  type DashboardStats,
  type GenerationHistoryItem,
  type LangSmithUsageResponse,
} from '@/api/generation';
import {
  clearApiKeySettings,
  getApiKeySettings,
  maskApiKey,
  saveApiKeySettings,
} from '@/utils/apiKeys';

const TOKEN_USAGE_CACHE_KEY = 'token_usage_snapshot_v1';
const TOKEN_USAGE_CACHE_TTL_MS = 5 * 60 * 1000;

type TokenUsageCacheSnapshot = {
  savedAt: string;
  page: number;
  pageSize: number;
  payload: LangSmithUsageResponse;
};

const statsLoading = ref(false);
const historyLoading = ref(false);

const stats = ref<DashboardStats | null>(null);
const records = ref<GenerationHistoryItem[]>([]);
const dataSource = ref('langsmith');
const projectName = ref('');
const isUsingCachedData = ref(false);
const lastSyncedAt = ref('');

const total = ref(0);
const page = ref(1);
const pageSize = ref(10);

const keyForm = ref({
  generationApiKey: '',
  embeddingApiKey: '',
});
const lastSavedAt = ref('');

const maskedGenerationKey = computed(() => maskApiKey(keyForm.value.generationApiKey));
const maskedEmbeddingKey = computed(() => maskApiKey(keyForm.value.embeddingApiKey));

const statCards = computed(() => [
  {
    name: '累计 Token',
    value: formatNumber(stats.value?.total_tokens || 0),
    icon: DataAnalysis,
    color: '#3b82f6',
  },
  {
    name: '平均耗时',
    value: formatDuration(stats.value?.avg_duration_ms || 0),
    icon: Timer,
    color: '#8b5cf6',
  },
  {
    name: '失败次数',
    value: String(stats.value?.failed_count || 0),
    icon: Warning,
    color: '#f59e0b',
  },
]);

function formatNumber(value: number): string {
  return new Intl.NumberFormat('zh-CN').format(value);
}

function formatDuration(value: number): string {
  if (!value || value < 1) {
    return '-';
  }

  if (value < 1000) {
    return `${Math.round(value)} ms`;
  }

  return `${(value / 1000).toFixed(2)} s`;
}

function formatDate(value?: string): string {
  if (!value) {
    return '-';
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return '-';
  }

  return date.toLocaleString('zh-CN', { hour12: false });
}

function initApiKeyForm() {
  const saved = getApiKeySettings();
  keyForm.value.generationApiKey = saved.generationApiKey;
  keyForm.value.embeddingApiKey = saved.embeddingApiKey;
  lastSavedAt.value = saved.updatedAt;
}

function applyUsagePayload(result: LangSmithUsageResponse) {
  stats.value = result.stats;
  records.value = result.history.items;
  total.value = result.history.total;
  dataSource.value = result.source || 'langsmith';
  projectName.value = result.project || '';

  if (result.history.page > 0) {
    page.value = result.history.page;
  }
  if (result.history.pageSize > 0) {
    pageSize.value = result.history.pageSize;
  }
}

function readUsageCache(): boolean {
  try {
    const raw = localStorage.getItem(TOKEN_USAGE_CACHE_KEY);
    if (!raw) {
      return false;
    }

    const snapshot = JSON.parse(raw) as TokenUsageCacheSnapshot;
    if (!snapshot || !snapshot.savedAt || !snapshot.payload || !snapshot.payload.history) {
      localStorage.removeItem(TOKEN_USAGE_CACHE_KEY);
      return false;
    }

    const savedAtMs = new Date(snapshot.savedAt).getTime();
    if (!Number.isFinite(savedAtMs) || Date.now() - savedAtMs > TOKEN_USAGE_CACHE_TTL_MS) {
      localStorage.removeItem(TOKEN_USAGE_CACHE_KEY);
      return false;
    }

    if (snapshot.page > 0) {
      page.value = snapshot.page;
    }
    if (snapshot.pageSize > 0) {
      pageSize.value = snapshot.pageSize;
    }

    applyUsagePayload(snapshot.payload);
    isUsingCachedData.value = true;
    lastSyncedAt.value = snapshot.savedAt;
    return true;
  } catch {
    localStorage.removeItem(TOKEN_USAGE_CACHE_KEY);
    return false;
  }
}

function persistUsageCache(payload: LangSmithUsageResponse) {
  const snapshot: TokenUsageCacheSnapshot = {
    savedAt: new Date().toISOString(),
    page: page.value,
    pageSize: pageSize.value,
    payload,
  };

  try {
    localStorage.setItem(TOKEN_USAGE_CACHE_KEY, JSON.stringify(snapshot));
  } catch {
    // ignore cache write failures
  }
}

function extractErrorMessage(error: unknown): string {
  let message = '加载 LangSmith Token 数据失败';

  if (axios.isAxiosError(error)) {
    const responseData = error.response?.data as { message?: string; data?: unknown } | undefined;

    if (typeof responseData?.data === 'string' && responseData.data.trim()) {
      message = responseData.data;
    } else if (typeof responseData?.message === 'string' && responseData.message.trim()) {
      message = responseData.message;
    } else if (error.message) {
      message = error.message;
    }

    return message;
  }

  if (error instanceof Error && error.message) {
    return error.message;
  }

  return message;
}

async function loadLangSmithUsage(options: { silent?: boolean } = {}) {
  const silent = options.silent === true;

  if (!silent) {
    statsLoading.value = true;
  }
  historyLoading.value = true;

  try {
    const result = await getLangSmithUsage(page.value, pageSize.value);
    applyUsagePayload(result);
    persistUsageCache(result);
    isUsingCachedData.value = false;
    lastSyncedAt.value = new Date().toISOString();
  } catch (error) {
    const message = extractErrorMessage(error);

    if (isUsingCachedData.value || stats.value !== null) {
      console.warn('[TokenUsage] refresh failed:', message);
      ElMessage.info('LangSmith 同步较慢，当前先展示最近缓存数据');
    } else {
      ElMessage.error(message);
    }
  } finally {
    statsLoading.value = false;
    historyLoading.value = false;
  }
}

function handlePageChange(nextPage: number) {
  page.value = nextPage;
  loadLangSmithUsage({ silent: true });
}

function handlePageSizeChange(nextPageSize: number) {
  pageSize.value = nextPageSize;
  page.value = 1;
  loadLangSmithUsage({ silent: true });
}

function saveApiKeys() {
  const saved = saveApiKeySettings({
    generationApiKey: keyForm.value.generationApiKey,
    embeddingApiKey: keyForm.value.embeddingApiKey,
  });

  keyForm.value.generationApiKey = saved.generationApiKey;
  keyForm.value.embeddingApiKey = saved.embeddingApiKey;
  lastSavedAt.value = saved.updatedAt;

  ElMessage.success('API Key 配置已保存，新请求会自动生效');
}

function clearApiKeys() {
  const cleared = clearApiKeySettings();
  keyForm.value.generationApiKey = cleared.generationApiKey;
  keyForm.value.embeddingApiKey = cleared.embeddingApiKey;
  lastSavedAt.value = '';
  ElMessage.success('API Key 配置已清空');
}

onMounted(() => {
  initApiKeyForm();
  const hasCachedSnapshot = readUsageCache();
  loadLangSmithUsage({ silent: hasCachedSnapshot });
});
</script>


<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">Token 使用与 API Key 配置</h1>
      <p class="page-subtitle">以下 Token 统计全部来自 LangSmith Trace，并支持手动配置生成与 Embedding 的 API Key</p>
      <p v-if="isUsingCachedData" class="text-xs app-text-muted mt-1">已优先展示最近缓存数据，正在后台同步 LangSmith...</p>
      <p v-else-if="lastSyncedAt" class="text-xs app-text-muted mt-1">最近同步：{{ formatDate(lastSyncedAt) }}</p>
    </div>

    <el-row :gutter="16">
      <el-col v-for="card in statCards" :key="card.name" :xs="12" :md="8">
        <el-card class="surface-card card-hover" shadow="never">
          <div class="flex items-center gap-3">
            <el-icon :size="22" :color="card.color"><component :is="card.icon" /></el-icon>
            <div>
              <div class="text-xl font-semibold app-text-primary">
                <el-skeleton v-if="statsLoading" :rows="1" animated />
                <template v-else>{{ card.value }}</template>
              </div>
              <div class="text-xs app-text-muted">{{ card.name }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card class="surface-card" shadow="never">
      <template #header>
        <div class="flex items-center justify-between">
          <div class="inline-flex items-center gap-2">
            <el-icon class="app-icon-primary"><Key /></el-icon>
            <span class="font-semibold">API Key 手动配置</span>
          </div>
          <el-button text :icon="RefreshRight" @click="initApiKeyForm">重置到已保存</el-button>
        </div>
      </template>

      <el-row :gutter="16">
        <el-col :xs="24" :lg="12">
          <el-form label-position="top">
            <el-form-item label="生成教案 API Key (DeepSeek)">
              <el-input
                v-model="keyForm.generationApiKey"
                type="password"
                show-password
                clearable
                placeholder="输入用于生成教案的 API Key"
              />
            </el-form-item>

            <el-form-item label="Embedding API Key (Qwen)">
              <el-input
                v-model="keyForm.embeddingApiKey"
                type="password"
                show-password
                clearable
                placeholder="输入用于 Embedding 的 API Key"
              />
            </el-form-item>

            <div class="flex gap-2">
              <el-button type="primary" :icon="CircleCheck" @click="saveApiKeys">保存配置</el-button>
              <el-button @click="clearApiKeys">清空配置</el-button>
            </div>
          </el-form>
        </el-col>

        <el-col :xs="24" :lg="12">
          <el-card class="surface-card" shadow="never">
            <div class="space-y-3 text-sm">
              <div>
                <div class="app-text-muted">当前生成 Key</div>
                <div class="font-medium app-text-primary break-all">{{ maskedGenerationKey }}</div>
              </div>
              <div>
                <div class="app-text-muted">当前 Embedding Key</div>
                <div class="font-medium app-text-primary break-all">{{ maskedEmbeddingKey }}</div>
              </div>
              <div>
                <div class="app-text-muted">最近保存时间</div>
                <div class="font-medium app-text-primary">{{ formatDate(lastSavedAt) }}</div>
              </div>
              <div class="text-xs app-text-muted">
                密钥保存在当前浏览器本地存储，仅用于你的请求头透传，不会写入后端数据库。
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </el-card>

    <el-card class="surface-card" shadow="never">
      <template #header>
        <div class="flex items-center justify-between">
          <div class="inline-flex items-center gap-2">
            <span class="font-semibold">LangSmith Token 明细</span>
            <el-tag size="small" effect="plain">{{ dataSource }}</el-tag>
            <el-tag v-if="projectName" size="small" type="info" effect="plain">{{ projectName }}</el-tag>
            <el-tag v-if="isUsingCachedData" size="small" type="warning" effect="plain">缓存数据</el-tag>
          </div>
          <el-button text :icon="RefreshRight" @click="loadLangSmithUsage({ silent: true })">刷新</el-button>
        </div>
      </template>

      <el-table v-loading="historyLoading" :data="records" stripe>
        <el-table-column label="时间" min-width="170">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>

        <el-table-column label="状态" width="110">
          <template #default="{ row }">
            <el-tag
              :type="row.status === 'completed' ? 'success' : row.status === 'failed' ? 'danger' : 'info'"
              size="small"
            >
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="token_count" label="Token" width="110" />

        <el-table-column label="耗时" width="120">
          <template #default="{ row }">
            {{ formatDuration(row.duration_ms || 0) }}
          </template>
        </el-table-column>

        <el-table-column label="错误信息" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.error_msg || '-' }}
          </template>
        </el-table-column>
      </el-table>

      <div class="mt-4 flex justify-end">
        <el-pagination
          background
          layout="total, sizes, prev, pager, next"
          :total="total"
          :page-size="pageSize"
          :current-page="page"
          :page-sizes="[10, 20, 50]"
          @current-change="handlePageChange"
          @size-change="handlePageSizeChange"
        />
      </div>
    </el-card>
  </div>
</template>
