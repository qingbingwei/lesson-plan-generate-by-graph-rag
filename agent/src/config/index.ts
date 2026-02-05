import dotenv from 'dotenv';
import path from 'path';

// 加载环境变量
dotenv.config({ path: path.resolve(__dirname, '../../.env') });

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
  
  neo4j: {
    uri: string;
    username: string;
    password: string;
  };
  
  redis: {
    url: string;
  };
  
  backend: {
    url: string;
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
  port: parseInt(process.env.PORT || '3001', 10),
  
  deepseek: {
    apiKey: process.env.DEEPSEEK_API_KEY || '',
    baseUrl: process.env.DEEPSEEK_BASE_URL || 'https://api.deepseek.com/v1',
    model: process.env.DEEPSEEK_MODEL || 'deepseek-chat',
    temperature: parseFloat(process.env.DEEPSEEK_TEMPERATURE || '0.7'),
    maxTokens: parseInt(process.env.DEEPSEEK_MAX_TOKENS || '4096', 10),
  },
  
  neo4j: {
    uri: process.env.NEO4J_URI || 'bolt://localhost:7687',
    username: process.env.NEO4J_USERNAME || 'neo4j',
    password: process.env.NEO4J_PASSWORD || 'password',
  },
  
  redis: {
    url: process.env.REDIS_URL || 'redis://localhost:6379',
  },
  
  backend: {
    url: process.env.BACKEND_URL || 'http://localhost:8080',
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

export default config;
