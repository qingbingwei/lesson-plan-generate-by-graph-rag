import dotenv from 'dotenv';
import path from 'path';

const rootEnvPath = path.resolve(__dirname, '../../../.env');
dotenv.config({ path: rootEnvPath });

function parseBoolean(value: string | undefined, defaultValue: boolean): boolean {
  if (!value) {
    return defaultValue;
  }

  const normalized = value.trim().toLowerCase();
  return ['1', 'true', 'yes', 'on'].includes(normalized);
}

function parseInteger(value: string | undefined, fallback: number, name: string): number {
  const source = value && value.trim() ? value.trim() : String(fallback);
  const parsed = Number.parseInt(source, 10);
  if (!Number.isInteger(parsed)) {
    throw new Error(`${name} 必须是整数，当前值: ${source}`);
  }
  return parsed;
}

function parseFloatNumber(value: string | undefined, fallback: number, name: string): number {
  const source = value && value.trim() ? value.trim() : String(fallback);
  const parsed = Number.parseFloat(source);
  if (!Number.isFinite(parsed)) {
    throw new Error(`${name} 必须是数字，当前值: ${source}`);
  }
  return parsed;
}

function ensureUrl(name: string, raw: string, protocols?: string[]): string {
  let parsed: URL;
  try {
    parsed = new URL(raw);
  } catch {
    throw new Error(`${name} 不是合法 URL: ${raw}`);
  }

  if (!parsed.protocol || !parsed.hostname) {
    throw new Error(`${name} 缺少协议或主机: ${raw}`);
  }

  if (protocols && protocols.length > 0) {
    const protocol = parsed.protocol.replace(':', '');
    if (!protocols.includes(protocol)) {
      throw new Error(`${name} 协议不合法（${protocol}），期望: ${protocols.join(', ')}`);
    }
  }

  return raw;
}

function looksLikePlaceholder(value: string): boolean {
  const lower = value.trim().toLowerCase();
  return (
    lower.startsWith('your-') ||
    lower.includes('change-in-production') ||
    lower.includes('replace-me') ||
    lower.includes('changeme')
  );
}

export interface Config {
  env: string;
  port: number;
  
  deepseek: {
    apiKey: string;
    baseUrl: string;
    model: string;
    temperature: number;
    maxTokens: number;
  };

  qwen: {
    apiKey: string;
    embeddingModel: string;
    embeddingUrl: string;
  };
  
  neo4j: {
    uri: string;
    username: string;
    password: string;
  };

  langsmith: {
    enabled: boolean;
    apiKey: string;
    endpoint: string;
    project: string;
  };
  
  log: {
    level: string;
    format: string;
  };
  
  graphRag: {
    vectorWeight: number;
    graphWeight: number;
    maxResults: number;
    searchDepth: number;
    embeddingDimension: number;
  };
}

const allowRuntimeApiKeys = parseBoolean(process.env.ALLOW_RUNTIME_API_KEYS, false);

const config: Config = {
  env: process.env.NODE_ENV || 'development',
  port: parseInteger(process.env.AGENT_PORT || process.env.PORT, 13001, 'AGENT_PORT/PORT'),
  
  deepseek: {
    apiKey: process.env.DEEPSEEK_API_KEY || '',
    baseUrl: ensureUrl(
      'DEEPSEEK_BASE_URL',
      process.env.DEEPSEEK_BASE_URL || 'https://api.deepseek.com/v1',
      ['http', 'https']
    ),
    model: process.env.DEEPSEEK_MODEL || 'deepseek-chat',
    temperature: parseFloatNumber(process.env.DEEPSEEK_TEMPERATURE, 0.7, 'DEEPSEEK_TEMPERATURE'),
    maxTokens: parseInteger(process.env.DEEPSEEK_MAX_TOKENS, 4096, 'DEEPSEEK_MAX_TOKENS'),
  },

  qwen: {
    apiKey: process.env.QWEN_API_KEY || '',
    embeddingModel: process.env.QWEN_EMBEDDING_MODEL || 'text-embedding-v4',
    embeddingUrl: ensureUrl(
      'QWEN_EMBEDDING_URL',
      process.env.QWEN_EMBEDDING_URL || 'https://dashscope.aliyuncs.com/compatible-mode/v1/embeddings',
      ['http', 'https']
    ),
  },
  
  neo4j: {
    uri: ensureUrl(process.env.NEO4J_URI ? 'NEO4J_URI' : 'NEO4J_URI(default)', process.env.NEO4J_URI || 'bolt://localhost:7687', [
      'bolt',
      'neo4j',
      'neo4j+s',
      'neo4j+ssc',
    ]),
    username: process.env.NEO4J_USERNAME || 'neo4j',
    password: process.env.NEO4J_PASSWORD || 'password',
  },

  langsmith: {
    enabled: parseBoolean(
      process.env.LANGSMITH_TRACING || process.env.LANGCHAIN_TRACING_V2,
      false
    ),
    apiKey: process.env.LANGSMITH_API_KEY || '',
    endpoint: ensureUrl(
      'LANGSMITH_ENDPOINT',
      process.env.LANGSMITH_ENDPOINT || 'https://api.smith.langchain.com',
      ['http', 'https']
    ),
    project:
      process.env.LANGSMITH_PROJECT ||
      process.env.LANGCHAIN_PROJECT ||
      'lesson-plan-agent',
  },
  
  log: {
    level: process.env.LOG_LEVEL || 'debug',
    format: process.env.LOG_FORMAT || 'json',
  },
  
  graphRag: {
    vectorWeight: parseFloatNumber(process.env.VECTOR_WEIGHT, 0.6, 'VECTOR_WEIGHT'),
    graphWeight: parseFloatNumber(process.env.GRAPH_WEIGHT, 0.4, 'GRAPH_WEIGHT'),
    maxResults: parseInteger(process.env.MAX_RESULTS, 10, 'MAX_RESULTS'),
    searchDepth: parseInteger(process.env.SEARCH_DEPTH, 2, 'SEARCH_DEPTH'),
    embeddingDimension: parseInteger(process.env.EMBEDDING_DIMENSION, 1536, 'EMBEDDING_DIMENSION'),
  },
};

function validateConfig(currentConfig: Config) {
  const errors: string[] = [];

  if (currentConfig.port <= 0 || currentConfig.port > 65535) {
    errors.push('AGENT_PORT/PORT 必须在 1~65535');
  }

  if (!currentConfig.deepseek.model.trim()) {
    errors.push('DEEPSEEK_MODEL 不能为空');
  }
  if (currentConfig.deepseek.temperature < 0 || currentConfig.deepseek.temperature > 2) {
    errors.push('DEEPSEEK_TEMPERATURE 必须在 0~2');
  }
  if (currentConfig.deepseek.maxTokens <= 0) {
    errors.push('DEEPSEEK_MAX_TOKENS 必须大于 0');
  }

  if (!currentConfig.qwen.embeddingModel.trim()) {
    errors.push('QWEN_EMBEDDING_MODEL 不能为空');
  }

  if (!currentConfig.neo4j.username.trim()) {
    errors.push('NEO4J_USERNAME 不能为空');
  }
  if (!currentConfig.neo4j.password.trim()) {
    errors.push('NEO4J_PASSWORD 不能为空');
  }

  if (!allowRuntimeApiKeys) {
    if (!currentConfig.deepseek.apiKey.trim()) {
      errors.push('DEEPSEEK_API_KEY 不能为空（可设置 ALLOW_RUNTIME_API_KEYS=true 使用请求头覆盖）');
    }
    if (!currentConfig.qwen.apiKey.trim()) {
      errors.push('QWEN_API_KEY 不能为空（可设置 ALLOW_RUNTIME_API_KEYS=true 使用请求头覆盖）');
    }
  }

  if (currentConfig.deepseek.apiKey && looksLikePlaceholder(currentConfig.deepseek.apiKey)) {
    errors.push('DEEPSEEK_API_KEY 仍是占位值');
  }
  if (currentConfig.qwen.apiKey && looksLikePlaceholder(currentConfig.qwen.apiKey)) {
    errors.push('QWEN_API_KEY 仍是占位值');
  }

  if (currentConfig.langsmith.enabled) {
    if (!currentConfig.langsmith.apiKey.trim()) {
      errors.push('LANGSMITH_TRACING=true 时必须配置 LANGSMITH_API_KEY');
    } else if (looksLikePlaceholder(currentConfig.langsmith.apiKey)) {
      errors.push('LANGSMITH_API_KEY 仍是占位值');
    }
  }

  if (currentConfig.graphRag.vectorWeight < 0 || currentConfig.graphRag.vectorWeight > 1) {
    errors.push('VECTOR_WEIGHT 必须在 0~1');
  }
  if (currentConfig.graphRag.graphWeight < 0 || currentConfig.graphRag.graphWeight > 1) {
    errors.push('GRAPH_WEIGHT 必须在 0~1');
  }
  if (currentConfig.graphRag.vectorWeight+ currentConfig.graphRag.graphWeight <= 0) {
    errors.push('VECTOR_WEIGHT + GRAPH_WEIGHT 必须大于 0');
  }
  if (currentConfig.graphRag.maxResults <= 0) {
    errors.push('MAX_RESULTS 必须大于 0');
  }
  if (currentConfig.graphRag.searchDepth <= 0) {
    errors.push('SEARCH_DEPTH 必须大于 0');
  }
  if (currentConfig.graphRag.embeddingDimension <= 0) {
    errors.push('EMBEDDING_DIMENSION 必须大于 0');
  }

  if (errors.length > 0) {
    throw new Error(`Agent 配置校验失败:\n- ${errors.join('\n- ')}`);
  }
}

function applyLangSmithRuntimeEnv(currentConfig: Config) {
  if (!currentConfig.langsmith.enabled) {
    process.env.LANGSMITH_TRACING = process.env.LANGSMITH_TRACING || 'false';
    process.env.LANGCHAIN_TRACING_V2 = process.env.LANGCHAIN_TRACING_V2 || 'false';
    return;
  }

  if (!currentConfig.langsmith.apiKey) {
    process.env.LANGSMITH_TRACING = 'false';
    process.env.LANGCHAIN_TRACING_V2 = 'false';
    return;
  }

  process.env.LANGSMITH_TRACING = 'true';
  process.env.LANGCHAIN_TRACING_V2 = 'true';
  process.env.LANGSMITH_API_KEY = currentConfig.langsmith.apiKey;
  process.env.LANGSMITH_ENDPOINT = currentConfig.langsmith.endpoint;
  process.env.LANGSMITH_PROJECT = currentConfig.langsmith.project;
  process.env.LANGCHAIN_PROJECT = currentConfig.langsmith.project;
}

validateConfig(config);
applyLangSmithRuntimeEnv(config);

export default config;
