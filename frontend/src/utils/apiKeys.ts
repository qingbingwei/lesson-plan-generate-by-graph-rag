export const GENERATION_API_KEY_HEADER = 'X-Generation-Api-Key';
export const EMBEDDING_API_KEY_HEADER = 'X-Embedding-Api-Key';

const API_KEY_STORAGE = 'lesson-plan:api-keys';

export interface ApiKeySettings {
  generationApiKey: string;
  embeddingApiKey: string;
  updatedAt: string;
}

function emptySettings(): ApiKeySettings {
  return {
    generationApiKey: '',
    embeddingApiKey: '',
    updatedAt: '',
  };
}

export function getApiKeySettings(): ApiKeySettings {
  if (typeof window === 'undefined') {
    return emptySettings();
  }

  try {
    const raw = localStorage.getItem(API_KEY_STORAGE);
    if (!raw) {
      return emptySettings();
    }

    const parsed = JSON.parse(raw) as Partial<ApiKeySettings>;
    return {
      generationApiKey: String(parsed.generationApiKey || ''),
      embeddingApiKey: String(parsed.embeddingApiKey || ''),
      updatedAt: String(parsed.updatedAt || ''),
    };
  } catch {
    return emptySettings();
  }
}

export function saveApiKeySettings(next: Partial<ApiKeySettings>): ApiKeySettings {
  const current = getApiKeySettings();
  const merged: ApiKeySettings = {
    generationApiKey: String(next.generationApiKey ?? current.generationApiKey).trim(),
    embeddingApiKey: String(next.embeddingApiKey ?? current.embeddingApiKey).trim(),
    updatedAt: new Date().toISOString(),
  };

  if (typeof window !== 'undefined') {
    localStorage.setItem(API_KEY_STORAGE, JSON.stringify(merged));
  }

  return merged;
}

export function clearApiKeySettings(): ApiKeySettings {
  const cleared = emptySettings();
  if (typeof window !== 'undefined') {
    localStorage.setItem(API_KEY_STORAGE, JSON.stringify(cleared));
  }
  return cleared;
}

export function getApiKeyHeaders(): Record<string, string> {
  const settings = getApiKeySettings();
  const headers: Record<string, string> = {};

  if (settings.generationApiKey) {
    headers[GENERATION_API_KEY_HEADER] = settings.generationApiKey;
  }
  if (settings.embeddingApiKey) {
    headers[EMBEDDING_API_KEY_HEADER] = settings.embeddingApiKey;
  }

  return headers;
}

export function maskApiKey(raw: string): string {
  const value = raw.trim();
  if (!value) {
    return '未设置';
  }
  if (value.length <= 10) {
    return `${value.slice(0, 2)}***${value.slice(-2)}`;
  }
  return `${value.slice(0, 4)}***${value.slice(-4)}`;
}
