const MAX_LATENCY_SAMPLES = 5000;
const QPS_WINDOW_SECONDS = 60;

type MetricBucket = {
  count: number;
  errorCount: number;
  totalLatencyMs: number;
  latencySamples: number[];
};

type MetricsState = {
  startedAtMs: number;
  httpSummary: MetricBucket;
  routes: Map<string, MetricBucket>;
  downstream: Map<string, MetricBucket>;
  requestTimestamps: number[];
};

export type BucketSnapshot = {
  count: number;
  error_count: number;
  error_rate: number;
  avg_latency_ms: number;
  p95_latency_ms: number;
  p99_latency_ms: number;
};

export type MetricsSnapshot = {
  timestamp: string;
  uptime_sec: number;
  summary: {
    total_requests: number;
    total_errors: number;
    error_rate: number;
    qps_avg: number;
    qps_1m: number;
    avg_latency_ms: number;
    p95_latency_ms: number;
    p99_latency_ms: number;
  };
  routes: Record<string, BucketSnapshot>;
  downstream: Record<string, BucketSnapshot>;
};

function createBucket(): MetricBucket {
  return {
    count: 0,
    errorCount: 0,
    totalLatencyMs: 0,
    latencySamples: [],
  };
}

const state: MetricsState = {
  startedAtMs: Date.now(),
  httpSummary: createBucket(),
  routes: new Map<string, MetricBucket>(),
  downstream: new Map<string, MetricBucket>(),
  requestTimestamps: [],
};

function isErrorStatus(statusCode: number): boolean {
  return statusCode >= 400 || statusCode <= 0;
}

function addSample(bucket: MetricBucket, latencyMs: number, isError: boolean) {
  bucket.count += 1;
  if (isError) {
    bucket.errorCount += 1;
  }

  const safeLatency = Number.isFinite(latencyMs) && latencyMs >= 0 ? latencyMs : 0;
  bucket.totalLatencyMs += safeLatency;
  bucket.latencySamples.push(safeLatency);

  if (bucket.latencySamples.length > MAX_LATENCY_SAMPLES) {
    bucket.latencySamples.splice(0, bucket.latencySamples.length - MAX_LATENCY_SAMPLES);
  }
}

function percentile(samples: number[], target: number): number {
  if (samples.length === 0) {
    return 0;
  }

  const copy = [...samples].sort((a, b) => a - b);
  const rank = Math.ceil((target / 100) * copy.length) - 1;
  const index = Math.max(0, Math.min(copy.length - 1, rank));
  return copy[index] ?? 0;
}

function round(value: number, digits: number): number {
  const factor = 10 ** digits;
  return Math.round(value * factor) / factor;
}

function formatBucket(bucket: MetricBucket): BucketSnapshot {
  if (bucket.count <= 0) {
    return {
      count: 0,
      error_count: 0,
      error_rate: 0,
      avg_latency_ms: 0,
      p95_latency_ms: 0,
      p99_latency_ms: 0,
    };
  }

  const avg = bucket.totalLatencyMs / bucket.count;
  return {
    count: bucket.count,
    error_count: bucket.errorCount,
    error_rate: round(bucket.errorCount / bucket.count, 4),
    avg_latency_ms: round(avg, 2),
    p95_latency_ms: round(percentile(bucket.latencySamples, 95), 2),
    p99_latency_ms: round(percentile(bucket.latencySamples, 99), 2),
  };
}

function routeKey(method: string, route: string): string {
  return `${method.toUpperCase()} ${route}`;
}

function downstreamKey(service: string, operation: string): string {
  return `${service}:${operation}`;
}

function purgeQPSWindow(nowSec: number) {
  const cutoff = nowSec - QPS_WINDOW_SECONDS + 1;
  let idx = 0;

  while (idx < state.requestTimestamps.length && (state.requestTimestamps[idx] ?? 0) < cutoff) {
    idx += 1;
  }

  if (idx > 0) {
    state.requestTimestamps = state.requestTimestamps.slice(idx);
  }
}

export function recordHttpRequest(method: string, route: string, statusCode: number, latencyMs: number): void {
  const normalizedRoute = route || 'UNKNOWN';
  const nowSec = Math.floor(Date.now() / 1000);
  const isError = isErrorStatus(statusCode);

  addSample(state.httpSummary, latencyMs, isError);

  const key = routeKey(method, normalizedRoute);
  let bucket = state.routes.get(key);
  if (!bucket) {
    bucket = createBucket();
    state.routes.set(key, bucket);
  }
  addSample(bucket, latencyMs, isError);

  state.requestTimestamps.push(nowSec);
  purgeQPSWindow(nowSec);
}

export function recordDownstream(service: string, operation: string, statusCode: number, latencyMs: number): void {
  const normalizedService = service || 'unknown';
  const normalizedOperation = operation || 'unknown';
  const key = downstreamKey(normalizedService, normalizedOperation);

  let bucket = state.downstream.get(key);
  if (!bucket) {
    bucket = createBucket();
    state.downstream.set(key, bucket);
  }

  addSample(bucket, latencyMs, isErrorStatus(statusCode));
}

export function snapshotMetrics(): MetricsSnapshot {
  const nowMs = Date.now();
  const nowSec = Math.floor(nowMs / 1000);
  purgeQPSWindow(nowSec);

  const uptimeSec = Math.max(1, Math.floor((nowMs - state.startedAtMs) / 1000));
  const summaryBucket = formatBucket(state.httpSummary);
  const qpsAvg = state.httpSummary.count / uptimeSec;
  const qps1m = state.requestTimestamps.length / QPS_WINDOW_SECONDS;

  const routeSnapshot: Record<string, BucketSnapshot> = {};
  for (const [key, bucket] of state.routes.entries()) {
    routeSnapshot[key] = formatBucket(bucket);
  }

  const downstreamSnapshot: Record<string, BucketSnapshot> = {};
  for (const [key, bucket] of state.downstream.entries()) {
    downstreamSnapshot[key] = formatBucket(bucket);
  }

  return {
    timestamp: new Date(nowMs).toISOString(),
    uptime_sec: uptimeSec,
    summary: {
      total_requests: state.httpSummary.count,
      total_errors: state.httpSummary.errorCount,
      error_rate: summaryBucket.error_rate,
      qps_avg: round(qpsAvg, 4),
      qps_1m: round(qps1m, 4),
      avg_latency_ms: summaryBucket.avg_latency_ms,
      p95_latency_ms: summaryBucket.p95_latency_ms,
      p99_latency_ms: summaryBucket.p99_latency_ms,
    },
    routes: routeSnapshot,
    downstream: downstreamSnapshot,
  };
}

