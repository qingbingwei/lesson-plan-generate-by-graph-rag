import { AsyncLocalStorage } from 'node:async_hooks';

export const GENERATION_API_KEY_HEADER = 'x-generation-api-key';
export const EMBEDDING_API_KEY_HEADER = 'x-embedding-api-key';

export interface RequestApiKeyOverrides {
  generationApiKey?: string;
  embeddingApiKey?: string;
}

const apiKeyStorage = new AsyncLocalStorage<RequestApiKeyOverrides>();

export function withRequestApiKeys<T>(
  overrides: RequestApiKeyOverrides,
  callback: () => Promise<T>
): Promise<T> {
  return apiKeyStorage.run(overrides, callback);
}

export function getRequestApiKeys(): RequestApiKeyOverrides {
  return apiKeyStorage.getStore() || {};
}
