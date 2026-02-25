import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { GenerateLessonRequest, GeneratedLesson, GenerationProgress } from '@/types';
import * as generationApi from '@/api/generation';

// --- 共享解析工具 ---

/** 解析教学目标（后端返回的是格式化字符串） */
function parseObjectives(text: string) {
  const knowledge = text.match(/【知识与技能】\n([\s\S]*?)(?=\n\n【|$)/)?.[1]?.trim() || '';
  const process = text.match(/【过程与方法】\n([\s\S]*?)(?=\n\n【|$)/)?.[1]?.trim() || '';
  const emotion = text.match(/【情感态度价值观】\n([\s\S]*?)$/)?.[1]?.trim() || '';
  return { knowledge, process, emotion };
}

/** 解析列表格式的字符串（如 "1. xxx\n2. yyy"） */
function parseList(text: string): string[] {
  if (!text) return [];
  return text.split('\n').map(line => line.replace(/^\d+\.\s*/, '').trim()).filter(Boolean);
}

/** 将后端响应转换为前端 GeneratedLesson 结构 */
function toGeneratedLesson(response: Record<string, any>, request: GenerateLessonRequest): GeneratedLesson {
  const objectives = response.objectives ? parseObjectives(response.objectives) : { knowledge: '', process: '', emotion: '' };
  return {
    title: response.title || request.topic,
    objectives,
    keyPoints: parseList(response.key_points || ''),
    difficultPoints: parseList(response.difficult_points || ''),
    teachingMethods: parseList(response.teaching_methods || ''),
    content: {
      sections: [
        {
          title: '新课讲授',
          duration: request.duration - 10,
          teacherActivity: response.content || '',
          studentActivity: response.activities || '',
          content: response.content || '',
        },
      ],
      materials: response.resources ? [response.resources] : [],
      homework: response.assessment || '',
    },
    evaluation: response.assessment || '',
  };
}

// --- 进度动画管理 ---

const WORKFLOW_STEPS = ['inputAnalysis', 'knowledgeQuery', 'objectiveDesign', 'contentDesign', 'activityDesign', 'outputFormat'] as const;

export const useGenerationStore = defineStore('generation', () => {
  // 状态
  const isGenerating = ref(false);
  const generatedLesson = ref<GeneratedLesson | null>(null);
  const progress = ref<GenerationProgress[]>([]);
  const error = ref<string | null>(null);

  // 节点映射
  const nodeLabels: Record<string, string> = {
    inputAnalysis: '输入分析',
    knowledgeQuery: '知识检索',
    objectiveDesign: '目标设计',
    contentDesign: '内容设计',
    activityDesign: '活动设计',
    outputFormat: '输出格式化',
  };

  // 初始化进度
  function initProgress() {
    progress.value = Object.keys(nodeLabels).map(node => ({
      node,
      status: 'pending' as const,
    }));
  }

  // 更新节点进度
  function updateProgress(node: string, status: GenerationProgress['status'], message?: string) {
    const index = progress.value.findIndex(p => p.node === node);
    if (index !== -1) {
      progress.value[index] = { node, status, message };
    }
  }

  // 生成教案（非流式 - 用后端API，显示等待状态）
  async function generateLesson(request: GenerateLessonRequest) {
    isGenerating.value = true;
    generatedLesson.value = null;
    error.value = null;
    initProgress();

    // 标记第一步为运行中
    updateProgress(WORKFLOW_STEPS[0], 'running');

    try {
      const result = await generationApi.generateLesson(request);
      
      if (result.status === 'completed') {
        generatedLesson.value = toGeneratedLesson(result, request);
        // 标记所有节点为完成
        progress.value.forEach(p => { p.status = 'completed'; });
      } else {
        error.value = result.error_message || '生成失败';
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : '生成失败';
    } finally {
      isGenerating.value = false;
    }
  }

  // 重置状态
  function reset() {
    isGenerating.value = false;
    generatedLesson.value = null;
    progress.value = [];
    error.value = null;
  }

  // 获取节点标签
  function getNodeLabel(node: string): string {
    return nodeLabels[node] || node;
  }

  return {
    // 状态
    isGenerating,
    generatedLesson,
    progress,
    error,
    
    // 方法
    generateLesson,
    reset,
    getNodeLabel,
  };
});
