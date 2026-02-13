<script setup lang="ts">
import { useRouter } from 'vue-router';
import {
  Reading,
  InfoFilled,
  MagicStick,
  Upload,
  Share,
  EditPen,
  Key,
} from '@element-plus/icons-vue';

const router = useRouter();

const quickStartSteps = [
  {
    title: '登录系统并初始化信息',
    description: '使用账号登录后，先到个人中心补充学科与年级信息，便于生成更贴合教学场景的教案。',
  },
  {
    title: '上传知识文档',
    description: '在“知识库管理”上传教材、讲义或试题文档，系统会自动解析并构建知识图谱。',
  },
  {
    title: '生成教案初稿',
    description: '在“生成教案”中选择学科、年级、课题和课时，结合知识库内容生成完整初稿。',
  },
  {
    title: '编辑、版本管理与发布',
    description: '进入教案详情进行优化，支持历史版本回滚、导出和后续复用。',
  },
];

const featureCards = [
  {
    title: '智能生成教案',
    icon: MagicStick,
    description: '基于提示词与知识上下文，自动生成目标、重难点、教学过程、评价与作业。',
  },
  {
    title: '知识图谱检索',
    icon: Share,
    description: '将教学文档结构化为节点和关系，支持关键词检索、范围过滤与图谱可视化。',
  },
  {
    title: '知识库文档管理',
    icon: Upload,
    description: '统一管理文档上传、状态查看与清理，作为生成教案时的知识来源。',
  },
  {
    title: '教案编辑与版本',
    icon: EditPen,
    description: '支持迭代编辑、查看历史版本并回滚，减少误修改导致的内容丢失。',
  },
  {
    title: 'Token 与密钥配置',
    icon: Key,
    description: '展示 Token 使用趋势，并支持单独配置生成模型与 Embedding 的 API Key。',
  },
  {
    title: 'AI 智能问答',
    icon: Reading,
    description: '快速咨询项目使用方法、模板生成建议和排障步骤，提升上手与运营效率。',
  },
];

const faqItems = [
  {
    question: '生成教案内容不够贴合，应该怎么优化？',
    answer:
      '优先补充知识库文档；在生成页增加更明确的课题范围、课时目标和教学风格；必要时使用智能问答先生成模板再微调。',
  },
  {
    question: '知识图谱节点太多看不清怎么办？',
    answer:
      '建议先使用关键词检索，再结合“仅命中 / 一跳 / 两跳”范围切换，逐步缩小可视化范围。',
  },
  {
    question: '服务启动后页面一直加载怎么办？',
    answer:
      '先确认数据库容器健康，再检查 `/tmp/lesson-plan` 目录中的 backend.log、agent.log、frontend.log。',
  },
  {
    question: 'API Key 配置后什么时候生效？',
    answer:
      '保存后对新请求立即生效，不会影响已经开始执行的生成任务。',
  },
];

function goAssistant() {
  router.push('/ai-assistant');
}
</script>

<template>
  <div class="page-container">
    <div class="page-header flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
      <div>
        <h1 class="page-title">帮助文档</h1>
        <p class="page-subtitle">快速了解系统使用流程、核心功能与使用说明</p>
      </div>
      <el-space>
        <el-button :icon="Reading" @click="goAssistant">打开智能问答</el-button>
      </el-space>
    </div>

    <el-row :gutter="16">
      <el-col :xs="24" :lg="14">
        <el-card class="surface-card" shadow="never">
          <template #header>
            <div class="flex items-center gap-2">
              <el-icon><InfoFilled /></el-icon>
              <span class="font-semibold">快速开始</span>
            </div>
          </template>

          <el-timeline>
            <el-timeline-item
              v-for="(step, index) in quickStartSteps"
              :key="step.title"
              :timestamp="`步骤 ${index + 1}`"
              placement="top"
            >
              <div class="font-medium app-text-primary">{{ step.title }}</div>
              <div class="text-sm app-text-muted mt-1 leading-6">{{ step.description }}</div>
            </el-timeline-item>
          </el-timeline>
        </el-card>
      </el-col>

      <el-col :xs="24" :lg="10">
        <el-card class="surface-card h-full" shadow="never">
          <template #header>
            <div class="flex items-center gap-2">
              <el-icon><MagicStick /></el-icon>
              <span class="font-semibold">使用建议</span>
            </div>
          </template>

          <el-alert title="推荐工作流" type="success" :closable="false" show-icon>
            <p class="text-sm leading-6">
              上传知识文档 → 生成教案初稿 → 在知识图谱中复核要点 → 进入详情页精修并保存版本。
            </p>
          </el-alert>

          <el-divider />

          <el-alert title="排障优先级" type="warning" :closable="false" show-icon>
            <p class="text-sm leading-6">
              先看日志、再看环境变量、最后检查模型 API Key 与网络连通性。
            </p>
          </el-alert>
        </el-card>
      </el-col>
    </el-row>

    <el-card class="surface-card" shadow="never">
      <template #header>
        <span class="font-semibold">核心功能说明</span>
      </template>

      <el-row :gutter="16">
        <el-col v-for="item in featureCards" :key="item.title" :xs="24" :sm="12" :lg="8" class="mb-4">
          <el-card class="surface-card card-hover h-full" shadow="never">
            <div class="flex items-start gap-3">
              <el-icon :size="20" class="app-icon-primary">
                <component :is="item.icon" />
              </el-icon>
              <div>
                <div class="font-semibold app-text-primary">{{ item.title }}</div>
                <div class="text-sm app-text-muted mt-1 leading-6">{{ item.description }}</div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </el-card>

    <el-card class="surface-card" shadow="never">
      <template #header>
        <span class="font-semibold">常见问题</span>
      </template>

      <el-collapse accordion>
        <el-collapse-item v-for="(item, index) in faqItems" :key="item.question" :name="String(index)">
          <template #title>
            <span class="font-medium app-text-primary">{{ item.question }}</span>
          </template>
          <div class="text-sm app-text-secondary leading-7">{{ item.answer }}</div>
        </el-collapse-item>
      </el-collapse>
    </el-card>
  </div>
</template>
