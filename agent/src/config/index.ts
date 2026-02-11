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

const config: Config = {
  env: process.env.NODE_ENV || 'development',
  port: parseInt(process.env.AGENT_PORT || process.env.PORT || '13001', 10),
  
  deepseek: {
    apiKey: process.env.DEEPSEEK_API_KEY || '',
    baseUrl: process.env.DEEPSEEK_BASE_URL || 'https://api.deepseek.com/v1',
    model: process.env.DEEPSEEK_MODEL || 'deepseek-chat',
    temperature: parseFloat(process.env.DEEPSEEK_TEMPERATURE || '0.7'),
    maxTokens: parseInt(process.env.DEEPSEEK_MAX_TOKENS || '4096', 10),
  },

  qwen: {
    apiKey: process.env.QWEN_API_KEY || '',
    embeddingModel: process.env.QWEN_EMBEDDING_MODEL || 'text-embedding-v4',
    embeddingUrl: process.env.QWEN_EMBEDDING_URL || 'https://dashscope.aliyuncs.com/compatible-mode/v1/embeddings',
  },
  
  neo4j: {
    uri: process.env.NEO4J_URI || 'bolt://localhost:7687',
    username: process.env.NEO4J_USERNAME || 'neo4j',
    password: process.env.NEO4J_PASSWORD || 'password',
  },

  langsmith: {
    enabled: parseBoolean(
      process.env.LANGSMITH_TRACING || process.env.LANGCHAIN_TRACING_V2,
      false
    ),
    apiKey: process.env.LANGSMITH_API_KEY || '',
    endpoint: process.env.LANGSMITH_ENDPOINT || 'https://api.smith.langchain.com',
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
    vectorWeight: parseFloat(process.env.VECTOR_WEIGHT || '0.6'),
    graphWeight: parseFloat(process.env.GRAPH_WEIGHT || '0.4'),
    maxResults: parseInt(process.env.MAX_RESULTS || '10', 10),
    searchDepth: parseInt(process.env.SEARCH_DEPTH || '2', 10),
    embeddingDimension: parseInt(process.env.EMBEDDING_DIMENSION || '1536', 10),
  },
};

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

applyLangSmithRuntimeEnv(config);

export default config;
