/**
 * 教案生成相关类型定义
 */

// 教案生成请求
export interface GenerateLessonRequest {
  subject: string;
  grade: string;
  topic: string;
  duration: number;
  style?: string;
  requirements?: string;
  context?: KnowledgeContext[];
  userId?: string; // 用户ID，用于过滤个人知识库
}

// 知识上下文
export interface KnowledgeContext {
  id: string;
  name: string;
  content: string;
  relevanceScore?: number;
  source?: string;
}

// 教学目标
export interface LessonObjectives {
  knowledge: string;   // 知识与技能
  process: string;     // 过程与方法
  emotion: string;     // 情感态度价值观
}

// 教学环节
export interface LessonSection {
  title: string;
  duration: number;    // 分钟
  teacherActivity: string;
  studentActivity: string;
  content: string;
  designIntent?: string;
}

// 生成的教案内容
export interface GeneratedLesson {
  title: string;
  objectives: LessonObjectives;
  keyPoints: string[];
  difficultPoints: string[];
  teachingMethods: string[];
  content: {
    sections: LessonSection[];
    materials: string[];
    homework: string;
  };
  evaluation: string;
  reflection?: string;
}

// 教案生成响应
export interface GenerateLessonResponse {
  success: boolean;
  data?: GeneratedLesson;
  error?: string;
  usage?: TokenUsage;
}

// Token使用情况
export interface TokenUsage {
  promptTokens: number;
  completionTokens: number;
  totalTokens: number;
}

// 重新生成环节请求
export interface RegenerateSectionRequest {
  lessonId: string;
  section: string;
  context: {
    subject: string;
    grade: string;
    topic: string;
    duration: number;
    current: Record<string, unknown>;
  };
}

// 重新生成环节响应
export interface RegenerateSectionResponse {
  section: string;
  content: Record<string, unknown>;
  message?: string;
}

// 知识点
export interface KnowledgePoint {
  id: string;
  name: string;
  description: string;
  difficulty: string;
  grade: string;
  importance: number;
  content: string;
  examples?: string[];
  embedding?: number[];
}

// 知识图谱节点
export interface KnowledgeNode {
  id: string;
  name: string;
  type: string;
  properties?: Record<string, unknown>;
}

// 知识图谱连线
export interface KnowledgeLink {
  source: string;
  target: string;
  type: string;
  properties?: Record<string, unknown>;
}

// 知识图谱数据
export interface KnowledgeGraphData {
  nodes: KnowledgeNode[];
  links: KnowledgeLink[];
}

// 搜索结果
export interface SearchResult {
  node: KnowledgeNode;
  score: number;
  vectorScore?: number;
  graphScore?: number;
  relatedNodes?: KnowledgeNode[];
}

// Graph RAG配置
export interface GraphRAGConfig {
  vectorWeight: number;
  graphWeight: number;
  maxResults: number;
  searchDepth: number;
}

// 工作流状态
export interface WorkflowState {
  // 输入
  input: GenerateLessonRequest;
  
  // 中间状态
  knowledgeContext?: KnowledgeContext[];
  lessonObjectives?: LessonObjectives;
  keyPoints?: string[];
  difficultPoints?: string[];
  teachingMethods?: string[];
  sections?: LessonSection[];
  materials?: string[];
  homework?: string;
  evaluation?: string;
  
  // 输出
  output?: GeneratedLesson;
  error?: string;
  
  // 元数据
  usage?: TokenUsage;
  startTime?: number;
  endTime?: number;
}

// DeepSeek消息类型
export interface DeepSeekMessage {
  role: 'system' | 'user' | 'assistant';
  content: string;
}

// DeepSeek请求
export interface DeepSeekRequest {
  model: string;
  messages: DeepSeekMessage[];
  temperature?: number;
  max_tokens?: number;
  top_p?: number;
  stream?: boolean;
}

// DeepSeek响应
export interface DeepSeekResponse {
  id: string;
  object: string;
  created: number;
  model: string;
  choices: {
    index: number;
    message: DeepSeekMessage;
    finish_reason: string;
  }[];
  usage: {
    prompt_tokens: number;
    completion_tokens: number;
    total_tokens: number;
  };
}

// Embedding请求
export interface EmbeddingRequest {
  text: string;
}

// Embedding响应
export interface EmbeddingResponse {
  embedding: number[];
}
