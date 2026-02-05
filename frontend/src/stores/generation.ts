import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { GenerateLessonRequest, GeneratedLesson, GenerationProgress } from '@/types';
import * as generationApi from '@/api/generation';

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

  // 生成教案
  async function generateLesson(request: GenerateLessonRequest) {
    isGenerating.value = true;
    generatedLesson.value = null;
    error.value = null;
    initProgress();

    // 使用定时器模拟进度更新，后端生成需要时间
    let progressInterval: ReturnType<typeof setInterval> | null = null;
    let currentStepIndex = 0;
    const steps = ['inputAnalysis', 'knowledgeQuery', 'objectiveDesign', 'contentDesign', 'activityDesign', 'outputFormat'];
    
    // 启动进度动画
    const startProgressAnimation = () => {
      updateProgress(steps[0], 'running');
      progressInterval = setInterval(() => {
        if (currentStepIndex < steps.length - 1) {
          // 完成当前步骤
          updateProgress(steps[currentStepIndex], 'completed');
          currentStepIndex++;
          // 开始下一步骤
          updateProgress(steps[currentStepIndex], 'running');
        }
      }, 8000); // 每8秒推进一步，给足够时间让API返回
    };

    try {
      startProgressAnimation();
      const result = await generationApi.generateLesson(request);
      
      // 清除定时器
      if (progressInterval) {
        clearInterval(progressInterval);
        progressInterval = null;
      }
      
      if (result.status === 'completed') {
        // 解析教学目标（后端返回的是格式化字符串）
        const parseObjectives = (text: string) => {
          const knowledge = text.match(/【知识与技能】\n([\s\S]*?)(?=\n\n【|$)/)?.[1]?.trim() || '';
          const process = text.match(/【过程与方法】\n([\s\S]*?)(?=\n\n【|$)/)?.[1]?.trim() || '';
          const emotion = text.match(/【情感态度价值观】\n([\s\S]*?)$/)?.[1]?.trim() || '';
          return { knowledge, process, emotion };
        };

        // 解析列表格式的字符串（如 "1. xxx\n2. yyy"）
        const parseList = (text: string): string[] => {
          if (!text) return [];
          return text.split('\n').map(line => line.replace(/^\d+\.\s*/, '').trim()).filter(Boolean);
        };

        const objectives = result.objectives ? parseObjectives(result.objectives) : { knowledge: '', process: '', emotion: '' };

        // 转换后端返回的简单结构为前端期望的复杂结构
        generatedLesson.value = {
          title: result.title || request.topic,
          objectives,
          keyPoints: parseList(result.key_points || ''),
          difficultPoints: parseList(result.difficult_points || ''),
          teachingMethods: parseList(result.teaching_methods || ''),
          content: {
            sections: [
              {
                title: '新课讲授',
                duration: request.duration - 10,
                teacherActivity: result.content || '',
                studentActivity: result.activities || '',
                content: result.content || '',
              },
            ],
            materials: result.resources ? [result.resources] : [],
            homework: result.assessment || '',
          },
          evaluation: result.assessment || '',
        };
        // 标记所有节点为完成
        progress.value.forEach(p => {
          p.status = 'completed';
        });
      } else {
        error.value = result.error_message || '生成失败';
      }
    } catch (err) {
      // 清除定时器
      if (progressInterval) {
        clearInterval(progressInterval);
        progressInterval = null;
      }
      error.value = err instanceof Error ? err.message : '生成失败';
    } finally {
      isGenerating.value = false;
    }
  }

  // 流式生成教案（改用普通接口，因为后端暂不支持流式）
  async function streamGenerateLesson(request: GenerateLessonRequest) {
    isGenerating.value = true;
    generatedLesson.value = null;
    error.value = null;
    initProgress();

    // 使用定时器模拟进度更新
    let progressInterval: ReturnType<typeof setInterval> | null = null;
    let currentStepIndex = 0;
    const steps = ['inputAnalysis', 'knowledgeQuery', 'objectiveDesign', 'contentDesign', 'activityDesign', 'outputFormat'];
    
    // 启动进度动画
    const startProgressAnimation = () => {
      updateProgress(steps[0], 'running');
      progressInterval = setInterval(() => {
        if (currentStepIndex < steps.length - 1) {
          updateProgress(steps[currentStepIndex], 'completed');
          currentStepIndex++;
          updateProgress(steps[currentStepIndex], 'running');
        }
      }, 8000);
    };

    try {
      startProgressAnimation();
      
      // 调用后端生成接口（会等待完成）
      const response = await generationApi.generateLesson(request);
      
      // 清除定时器
      if (progressInterval) {
        clearInterval(progressInterval);
        progressInterval = null;
      }
      
      // 生成完成，标记所有步骤为完成
      updateProgress('inputAnalysis', 'completed');
      updateProgress('knowledgeQuery', 'completed');
      updateProgress('objectiveDesign', 'completed');
      updateProgress('contentDesign', 'completed');
      updateProgress('activityDesign', 'completed');
      updateProgress('outputFormat', 'completed');
      
      // 解析教学目标和列表
      const parseObjectives = (text: string) => {
        const knowledge = text.match(/【知识与技能】\n([\s\S]*?)(?=\n\n【|$)/)?.[1]?.trim() || '';
        const process = text.match(/【过程与方法】\n([\s\S]*?)(?=\n\n【|$)/)?.[1]?.trim() || '';
        const emotion = text.match(/【情感态度价值观】\n([\s\S]*?)$/)?.[1]?.trim() || '';
        return { knowledge, process, emotion };
      };

      const parseList = (text: string): string[] => {
        if (!text) return [];
        return text.split('\n').map(line => line.replace(/^\d+\.\s*/, '').trim()).filter(Boolean);
      };

      const objectives = response.objectives ? parseObjectives(response.objectives) : { knowledge: '', process: '', emotion: '' };

      // 转换后端返回的简单结构为前端期望的复杂结构
      generatedLesson.value = {
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
        reflection: '',
      };
    } catch (err) {
      // 清除定时器
      if (progressInterval) {
        clearInterval(progressInterval);
        progressInterval = null;
      }
      error.value = err instanceof Error ? err.message : '生成失败';
      // 标记当前步骤为错误
      if (currentStepIndex < steps.length) {
        updateProgress(steps[currentStepIndex], 'error', error.value);
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
