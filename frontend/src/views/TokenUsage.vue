<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
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
  getGenerationHistory,
  getGenerationStats,
  type DashboardStats,
  type GenerationHistoryItem,
} from '@/api/generation';
import {
  clearApiKeySettings,
  getApiKeySettings,
  maskApiKey,
  saveApiKeySettings,
} from '@/utils/apiKeys';

const statsLoading = ref(false);
const historyLoading = ref(false);

const stats = ref<DashboardStats | null>(null);
const records = ref<GenerationHistoryItem[]>([]);

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

async function loadStats() {
  statsLoading.value = true;
  try {
    stats.value = await getGenerationStats();
  } catch {
    ElMessage.error('加载 Token 统计失败');
  } finally {
    statsLoading.value = false;
  }
}

async function loadHistory() {
  historyLoading.value = true;
  try {
    const result = await getGenerationHistory(page.value, pageSize.value);
    records.value = result.items;
    total.value = result.total;
  } catch {
    ElMessage.error('加载生成记录失败');
  } finally {
    historyLoading.value = false;
  }
}

function handlePageChange(nextPage: number) {
  page.value = nextPage;
  loadHistory();
}

function handlePageSizeChange(nextPageSize: number) {
  pageSize.value = nextPageSize;
  page.value = 1;
  loadHistory();
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
  loadStats();
  loadHistory();
});
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">Token 使用与 API Key 配置</h1>
      <p class="page-subtitle">查看 Token 使用量，并手动配置生成与 Embedding 的 API Key</p>
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
          <span class="font-semibold">生成记录 Token 明细</span>
          <el-button text :icon="RefreshRight" @click="loadHistory">刷新</el-button>
        </div>
      </template>

      <el-table :data="records" stripe v-loading="historyLoading">
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
