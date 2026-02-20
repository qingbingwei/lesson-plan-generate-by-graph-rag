// 用户相关类型
export interface User {
  id: number;
  username: string;
  email: string;
  role: UserRole;
  profile: UserProfile;
  createdAt: string;
  updatedAt: string;
}

export type UserRole = 'admin' | 'teacher' | 'guest';

export interface UserProfile {
  name?: string;
  avatar?: string;
  school?: string;
  subject?: string;
  grade?: string;
  phone?: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  expires_at: number;
  user: User;
}

// 教案相关类型
export interface Lesson {
  id: string;
  userId: string;
  title: string;
  subject: string;
  grade: string;
  duration: number;
  objectives: LessonObjectives;
  keyPoints: string[];
  difficultPoints: string[];
  teachingMethods: string[];
  content: LessonContent;
  evaluation: string;
  reflection?: string;
  status: LessonStatus;
  version: number;
  metadata?: LessonMetadata;
  createdAt: string;
  updatedAt: string;
}

export type LessonStatus = 'draft' | 'published' | 'archived';

export interface LessonObjectives {
  knowledge: string;
  process: string;
  emotion: string;
}

export interface LessonContent {
  sections: LessonSection[];
  materials: string[];
  homework: string;
}

export interface LessonSection {
  title: string;
  duration: number;
  teacherActivity: string;
  studentActivity: string;
  content: string;
  designIntent?: string;
}

export interface LessonMetadata {
  generatedBy?: string;
  tokenUsage?: TokenUsage;
  generationTime?: number;
  knowledgeIds?: string[];
}

export interface TokenUsage {
  promptTokens: number;
  completionTokens: number;
  totalTokens: number;
}

// 教案生成相关类型
export interface GenerateLessonRequest {
  subject: string;
  grade: string;
  topic: string;
  duration: number;
  style?: string;
  requirements?: string;
}

export interface GenerateLessonResponse {
  success: boolean;
  data?: GeneratedLesson;
  error?: string;
  usage?: TokenUsage;
}

export interface GeneratedLesson {
  title: string;
  objectives: LessonObjectives;
  keyPoints: string[];
  difficultPoints: string[];
  teachingMethods: string[];
  content: LessonContent;
  evaluation: string;
  reflection?: string;
}

export interface GenerationProgress {
  node: string;
  status: 'pending' | 'running' | 'completed' | 'error';
  message?: string;
  currentNode?: string;
  nodes?: Record<string, { status: 'pending' | 'running' | 'completed' | 'error'; message?: string }>;
}

// 版本相关类型
export interface LessonVersion {
  id: string;
  lessonId: string;
  version: number;
  content: string;
  changeLog?: string;
  createdBy?: string;
  createdAt: string;
}

// 评论相关类型
export interface LessonComment {
  id: number;
  lessonId: number;
  userId: number;
  user?: User;
  content: string;
  parentId?: number;
  replies?: LessonComment[];
  createdAt: string;
  updatedAt: string;
}

// 收藏相关类型
export interface LessonFavorite {
  id: number;
  lessonId: number;
  userId: number;
  lesson?: Lesson;
  createdAt: string;
}

// 知识图谱相关类型
export interface KnowledgeNode {
  id: string;
  name: string;
  type: KnowledgeNodeType;
  properties?: Record<string, unknown>;
}

export type KnowledgeNodeType = 
  | 'Subject' 
  | 'Chapter' 
  | 'KnowledgePoint' 
  | 'Skill' 
  | 'Concept'
  | 'Principle'
  | 'Formula'
  | 'Example';

export interface KnowledgeLink {
  source: string;
  target: string;
  type: string;
  properties?: Record<string, unknown>;
}

export interface KnowledgeGraphData {
  nodes: KnowledgeNode[];
  links: KnowledgeLink[];
}

export interface KnowledgePoint {
  id: string;
  name: string;
  description: string;
  difficulty: string;
  grade: string;
  importance: number;
  content: string;
  examples?: string[];
}

// API 响应类型
export interface ApiResponse<T = unknown> {
  success?: boolean;
  code: number;
  message: string;
  data: T;
  error?: string | { code?: string; details?: unknown };
  trace_id?: string;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

export interface PaginationParams {
  page?: number;
  pageSize?: number;
  sort?: string;
  order?: 'asc' | 'desc';
}

// 表单相关类型
export interface FormField {
  label: string;
  name: string;
  type: 'text' | 'number' | 'select' | 'textarea' | 'checkbox' | 'radio';
  placeholder?: string;
  required?: boolean;
  options?: { label: string; value: string | number }[];
  validation?: (value: unknown) => string | undefined;
}

// 通知相关类型
export interface Notification {
  id: string;
  type: 'success' | 'error' | 'warning' | 'info';
  title: string;
  message?: string;
  duration?: number;
}

// 模态框相关类型
export interface ModalState {
  isOpen: boolean;
  title?: string;
  content?: string;
  onConfirm?: () => void;
  onCancel?: () => void;
}
