import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { GenerateLessonRequest, GeneratedLesson, GenerationProgress } from '@/types';
import * as generationApi from '@/api/generation';
import { getApiKeyHeaders } from '@/utils/apiKeys';

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
  const cancelFn = ref<(() => void) | null>(null);

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

  // 流式生成教案（通过 Agent SSE 获取真实节点进度）
  async function streamGenerateLesson(request: GenerateLessonRequest) {
    isGenerating.value = true;
    generatedLesson.value = null;
    error.value = null;
    initProgress();

    // 标记第一步为运行中
    updateProgress(WORKFLOW_STEPS[0], 'running');

    try {
      // 获取 auth token
      let token = '';
      try {
        const authData = localStorage.getItem('auth');
        if (authData) {
          const parsed = JSON.parse(authData);
          token = parsed.token || '';
        }
      } catch { /* ignore */ }

      const controller = new AbortController();
      cancelFn.value = () => controller.abort();

      const response = await fetch('/agent/generate/stream', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...getApiKeyHeaders(),
          ...(token ? { 'Authorization': `Bearer ${token}` } : {}),
        },
        body: JSON.stringify(request),
        signal: controller.signal,
      });

      if (!response.ok) {
        throw new Error(`生成失败: HTTP ${response.status}`);
      }

      const reader = response.body?.getReader();
      if (!reader) throw new Error('无响应流');

      const decoder = new TextDecoder();
      let buffer = '';
      let lastCompletedNode = '';
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      let finalState: Record<string, any> = {};

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() || '';

        for (const line of lines) {
          if (!line.startsWith('data: ')) continue;
          const data = line.slice(6).trim();
          if (data === '[DONE]') continue;

          try {
            const event = JSON.parse(data) as { node: string; state: Record<string, unknown> };
            const nodeName = event.node;

            // 标记已完成节点和下一个运行中的节点
            if (WORKFLOW_STEPS.includes(nodeName as typeof WORKFLOW_STEPS[number])) {
              // 标记之前运行中的节点为完成
              if (lastCompletedNode) {
                updateProgress(lastCompletedNode, 'completed');
              }
              // 当前节点完成
              updateProgress(nodeName, 'completed');
              lastCompletedNode = nodeName;

              // 标记下一个节点为运行中
              const idx = WORKFLOW_STEPS.indexOf(nodeName as typeof WORKFLOW_STEPS[number]);
              if (idx < WORKFLOW_STEPS.length - 1) {
                updateProgress(WORKFLOW_STEPS[idx + 1], 'running');
              }

              // 合并状态用于最终结果
              if (event.state) {
                finalState = { ...finalState, ...event.state };
              }
            }

            // 检查错误
            if (event.state?.error) {
              error.value = event.state.error as string;
              updateProgress(nodeName, 'error', error.value);
            }
          } catch { /* ignore parse errors */ }
        }
      }

      // 所有步骤完成
      if (!error.value) {
        WORKFLOW_STEPS.forEach(step => updateProgress(step, 'completed'));
        
        // 从最终状态的 output 字段构建结果
        if (finalState.output) {
          generatedLesson.value = {
            title: finalState.output.title || request.topic,
            objectives: finalState.output.objectives || { knowledge: '', process: '', emotion: '' },
            keyPoints: finalState.output.keyPoints || [],
            difficultPoints: finalState.output.difficultPoints || [],
            teachingMethods: finalState.output.teachingMethods || [],
            content: finalState.output.content || { sections: [], materials: [], homework: '' },
            evaluation: finalState.output.evaluation || '',
            reflection: finalState.output.reflection || '',
          };
        } else {
          error.value = '生成完成但未收到结果数据';
        }
      }
    } catch (err) {
      if ((err as Error).name === 'AbortError') {
        error.value = '已取消生成';
      } else {
        error.value = err instanceof Error ? err.message : '生成失败';
      }
      // 标记当前运行中的步骤为错误
      const currentStep = progress.value.find(p => p.status === 'running');
      if (currentStep) {
        updateProgress(currentStep.node, 'error', error.value);
      }
    } finally {
      isGenerating.value = false;
      cancelFn.value = null;
    }
  }

  // 取消生成
  function cancelGeneration() {
    if (cancelFn.value) {
      cancelFn.value();
      cancelFn.value = null;
    }
    isGenerating.value = false;
    error.value = '已取消生成';
  }

  // 重新生成某个环节
  async function regenerateSection(
    lessonId: string,
    section: string,
    context: {
      subject: string;
      grade: string;
      topic: string;
      duration: number;
      current: Record<string, unknown>;
    }
  ) {
    try {
      const result = await generationApi.regenerateSection(lessonId, section, context);
      return result;
    } catch (err) {
      error.value = err instanceof Error ? err.message : '重新生成失败';
      throw err;
    }
  }

  // 重置状态
  function reset() {
    isGenerating.value = false;
    generatedLesson.value = null;
    progress.value = [];
    error.value = null;
    cancelFn.value = null;
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
    streamGenerateLesson,
    cancelGeneration,
    regenerateSection,
    reset,
    getNodeLabel,
  };
});
