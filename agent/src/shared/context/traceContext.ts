import { AsyncLocalStorage } from 'node:async_hooks';

type TraceStore = {
  traceId: string;
};

const traceStorage = new AsyncLocalStorage<TraceStore>();

export function withTraceContext<T>(traceId: string, callback: () => T): T {
  return traceStorage.run({ traceId }, callback);
}

export function getTraceIdFromContext(): string {
  return traceStorage.getStore()?.traceId || '';
}

