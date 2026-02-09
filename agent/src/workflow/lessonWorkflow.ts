import { StateGraph, END, START, Annotation } from '@langchain/langgraph';
import {
  inputAnalysisNode,
  knowledgeQueryNode,
  objectiveDesignNode,
  contentDesignNode,
  activityDesignNode,
  outputFormatNode,
} from '../nodes';
import logger from '../utils/logger';
import type { WorkflowState, GenerateLessonRequest, GenerateLessonResponse, KnowledgeContext, LessonObjectives, LessonSection, GeneratedLesson, TokenUsage } from '../types';

// 定义状态注解 - LangGraph 需要这个来正确传递和合并状态
const WorkflowStateAnnotation = Annotation.Root({
  input: Annotation<GenerateLessonRequest>({
    reducer: (_, b) => b,
    default: () => ({ subject: '', grade: '', topic: '', duration: 45 }),
  }),
  knowledgeContext: Annotation<KnowledgeContext[] | undefined>({
    reducer: (_, b) => b,
    default: () => undefined,
  }),
  lessonObjectives: Annotation<LessonObjectives | undefined>({
    reducer: (_, b) => b,
    default: () => undefined,
  }),
  keyPoints: Annotation<string[] | undefined>({
    reducer: (_, b) => b,
    default: () => undefined,
  }),
  difficultPoints: Annotation<string[] | undefined>({
    reducer: (_, b) => b,
    default: () => undefined,
  }),
  teachingMethods: Annotation<string[] | undefined>({
    reducer: (_, b) => b,
    default: () => undefined,
  }),
  sections: Annotation<LessonSection[] | undefined>({
    reducer: (_, b) => b,
    default: () => undefined,
  }),
  materials: Annotation<string[] | undefined>({
    reducer: (_, b) => b,
    default: () => undefined,
  }),
  homework: Annotation<string | undefined>({
    reducer: (_, b) => b,
    default: () => undefined,
  }),
  evaluation: Annotation<string | undefined>({
    reducer: (_, b) => b,
    default: () => undefined,
  }),
  output: Annotation<GeneratedLesson | undefined>({
    reducer: (_, b) => b,
    default: () => undefined,
  }),
  error: Annotation<string | undefined>({
    reducer: (_, b) => b,
    default: () => undefined,
  }),
  usage: Annotation<TokenUsage | undefined>({
    reducer: (a, b) => {
      if (!a && !b) return undefined;
      if (!a) return b;
      if (!b) return a;
      return {
        promptTokens: (a.promptTokens || 0) + (b.promptTokens || 0),
        completionTokens: (a.completionTokens || 0) + (b.completionTokens || 0),
        totalTokens: (a.totalTokens || 0) + (b.totalTokens || 0),
      };
    },
    default: () => undefined,
  }),
  startTime: Annotation<number | undefined>({
    reducer: (a, b) => b ?? a,
    default: () => undefined,
  }),
  endTime: Annotation<number | undefined>({
    reducer: (_, b) => b,
    default: () => undefined,
  }),
});

/**
 * 创建教案生成工作流
 */
export function createLessonWorkflow() {
  // 检查是否有错误的路由函数
  const shouldContinue = (state: WorkflowState) => {
    return state.error ? 'outputFormat' : undefined;
  };

  // 创建状态图 - 使用 Annotation 定义的状态模式
  const workflow = new StateGraph(WorkflowStateAnnotation)
    // 添加节点
    .addNode('inputAnalysis', inputAnalysisNode)
    .addNode('knowledgeQuery', knowledgeQueryNode)
    .addNode('objectiveDesign', objectiveDesignNode)
    .addNode('contentDesign', contentDesignNode)
    .addNode('activityDesign', activityDesignNode)
    .addNode('outputFormat', outputFormatNode)
    // 添加边 - 使用条件边实现错误短路
    .addEdge(START, 'inputAnalysis')
    .addConditionalEdges('inputAnalysis', (state) => {
      return (state as unknown as WorkflowState).error ? 'outputFormat' : 'knowledgeQuery';
    })
    .addConditionalEdges('knowledgeQuery', (state) => {
      return (state as unknown as WorkflowState).error ? 'outputFormat' : 'objectiveDesign';
    })
    .addConditionalEdges('objectiveDesign', (state) => {
      return (state as unknown as WorkflowState).error ? 'outputFormat' : 'contentDesign';
    })
    .addConditionalEdges('contentDesign', (state) => {
      return (state as unknown as WorkflowState).error ? 'outputFormat' : 'activityDesign';
    })
    .addEdge('activityDesign', 'outputFormat')
    .addEdge('outputFormat', END);

  // 编译工作流
  return workflow.compile();
}

/**
 * 运行教案生成工作流
 */
export async function runLessonWorkflow(
  request: GenerateLessonRequest
): Promise<GenerateLessonResponse> {
  const startTime = Date.now();
  logger.info('Starting lesson generation workflow', {
    subject: request.subject,
    grade: request.grade,
    topic: request.topic,
    duration: request.duration,
  });

  try {
    const workflow = createLessonWorkflow();

    // 初始化状态
    const initialState: Partial<WorkflowState> = {
      input: request,
      startTime: Date.now(),
    };

    // 运行工作流
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const finalState = await workflow.invoke(initialState as WorkflowState) as any as WorkflowState;

    const duration = Date.now() - startTime;
    logger.info('Lesson generation workflow completed', {
      duration,
      success: !finalState.error,
      usage: finalState.usage,
    });

    // 返回结果
    if (finalState.error) {
      return {
        success: false,
        error: finalState.error,
        usage: finalState.usage,
      };
    }

    return {
      success: true,
      data: finalState.output,
      usage: finalState.usage,
    };
  } catch (error) {
    const duration = Date.now() - startTime;
    logger.error('Lesson generation workflow failed', {
      error,
      duration,
    });

    return {
      success: false,
      error: error instanceof Error ? error.message : 'Workflow execution failed',
    };
  }
}

/**
 * 流式运行教案生成工作流
 */
export async function* streamLessonWorkflow(
  request: GenerateLessonRequest
): AsyncGenerator<{ node: string; state: Partial<WorkflowState> }> {
  logger.info('Starting streaming lesson generation workflow', {
    subject: request.subject,
    topic: request.topic,
  });

  try {
    const workflow = createLessonWorkflow();

    const initialState: Partial<WorkflowState> = {
      input: request,
      startTime: Date.now(),
    };

    // 流式运行
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    for await (const event of await workflow.stream(initialState as WorkflowState) as any) {
      const entries = Object.entries(event);
      const firstEntry = entries[0];
      if (!firstEntry) continue;
      
      const [nodeName, nodeOutput] = firstEntry;
      
      logger.debug('Workflow node completed', {
        node: nodeName,
        hasError: !!(nodeOutput as Partial<WorkflowState>).error,
      });

      yield {
        node: nodeName,
        state: nodeOutput as Partial<WorkflowState>,
      };
    }
  } catch (error) {
    logger.error('Streaming workflow failed', { error });
    throw error;
  }
}

export default {
  createLessonWorkflow,
  runLessonWorkflow,
  streamLessonWorkflow,
};
