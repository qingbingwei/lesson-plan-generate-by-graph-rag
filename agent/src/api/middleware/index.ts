import { Request, Response, NextFunction } from 'express';
import { v4 as uuidv4 } from 'uuid';
import logger from '../../shared/utils/logger';
import { withTraceContext } from '../../shared/context/traceContext';
import { recordHttpRequest } from '../../shared/observability/metrics';

export const TRACE_ID_HEADER = 'X-Trace-ID';
export const REQUEST_ID_HEADER = 'X-Request-ID';

type StandardErrorResponse = {
  success: false;
  code: number;
  message: string;
  error: string;
  error_detail: {
    code: string;
    details?: unknown;
  };
  trace_id: string;
};

function defaultErrorCode(statusCode: number): string {
  if (statusCode === 400) {
    return 'BAD_REQUEST';
  }
  if (statusCode === 401) {
    return 'UNAUTHORIZED';
  }
  if (statusCode === 403) {
    return 'FORBIDDEN';
  }
  if (statusCode === 404) {
    return 'NOT_FOUND';
  }
  if (statusCode === 429) {
    return 'RATE_LIMITED';
  }
  if (statusCode >= 500) {
    return 'INTERNAL_SERVER_ERROR';
  }
  return 'REQUEST_FAILED';
}

function resolveTraceId(req: Request): string {
  const traceId = (req.header(TRACE_ID_HEADER) || '').trim();
  if (traceId) {
    return traceId;
  }

  const requestId = (req.header(REQUEST_ID_HEADER) || '').trim();
  if (requestId) {
    return requestId;
  }

  return uuidv4();
}

function setTraceHeaders(req: Request, res: Response, traceId: string) {
  res.setHeader(TRACE_ID_HEADER, traceId);
  res.setHeader(REQUEST_ID_HEADER, traceId);
  req.headers[TRACE_ID_HEADER.toLowerCase()] = traceId;
  req.headers[REQUEST_ID_HEADER.toLowerCase()] = traceId;
  res.locals.traceId = traceId;
}

function getRoutePath(req: Request): string {
  const routePath = req.route?.path;
  if (typeof routePath === 'string') {
    return `${req.baseUrl || ''}${routePath}`;
  }

  return req.path || req.originalUrl || 'UNKNOWN';
}

function buildErrorResponse(
  statusCode: number,
  errorCode: string,
  message: string,
  traceId: string,
  details?: unknown
): StandardErrorResponse {
  return {
    success: false,
    code: statusCode,
    message,
    error: message,
    error_detail: {
      code: errorCode,
      details,
    },
    trace_id: traceId,
  };
}

function normalizeJsonPayload(body: unknown, statusCode: number, traceId: string): unknown {
  if (!body || typeof body !== 'object' || Array.isArray(body)) {
    return body;
  }

  const payload = { ...(body as Record<string, unknown>) };

  if (typeof payload.trace_id !== 'string' || !payload.trace_id.trim()) {
    payload.trace_id = traceId;
  }

  const inferredSuccess = statusCode < 400;
  if (typeof payload.success !== 'boolean') {
    payload.success = inferredSuccess;
  }

  if (typeof payload.code !== 'number') {
    payload.code = payload.success ? 0 : statusCode;
  }

  if (typeof payload.message !== 'string' || !payload.message.trim()) {
    if (payload.success) {
      payload.message = 'success';
    } else if (typeof payload.error === 'string' && payload.error.trim()) {
      payload.message = payload.error;
    } else {
      payload.message = statusCode >= 500 ? 'Internal server error' : 'request failed';
    }
  }

  if (!payload.success) {
    if (typeof payload.error !== 'string' || !payload.error.trim()) {
      payload.error = String(payload.message);
    }

    if (!payload.error_detail || typeof payload.error_detail !== 'object') {
      payload.error_detail = {
        code: defaultErrorCode(statusCode),
      };
    }
  }

  return payload;
}

export function getTraceId(req: Request, res?: Response): string {
  const fromLocals = res?.locals?.traceId;
  if (typeof fromLocals === 'string' && fromLocals.trim()) {
    return fromLocals;
  }

  const fromHeader = req.header(TRACE_ID_HEADER) || req.header(REQUEST_ID_HEADER);
  if (typeof fromHeader === 'string' && fromHeader.trim()) {
    return fromHeader.trim();
  }

  return '';
}

/**
 * trace 透传中间件：接收上游 trace_id，若缺失则生成并写回响应头。
 */
export function traceMiddleware(req: Request, res: Response, next: NextFunction) {
  const traceId = resolveTraceId(req);
  setTraceHeaders(req, res, traceId);

  const originalJson = res.json.bind(res);
  res.json = ((body: unknown) => {
    const normalized = normalizeJsonPayload(body, res.statusCode, traceId);
    return originalJson(normalized);
  }) as Response['json'];

  withTraceContext(traceId, () => next());
}

/**
 * 请求日志中间件
 */
export function requestLogger(req: Request, res: Response, next: NextFunction) {
  const startedAt = Date.now();
  const traceId = getTraceId(req, res);

  logger.info('Incoming request', {
    trace_id: traceId,
    method: req.method,
    path: req.path,
    ip: req.ip,
  });

  res.on('finish', () => {
    const durationMs = Date.now() - startedAt;
    const route = getRoutePath(req);
    const statusCode = res.statusCode;
    recordHttpRequest(req.method, route, statusCode, durationMs);

    const logPayload = {
      trace_id: traceId,
      method: req.method,
      path: req.path,
      route,
      ip: req.ip,
      status: statusCode,
      latency_ms: durationMs,
    };

    if (statusCode >= 500) {
      logger.error('Request completed', logPayload);
      return;
    }
    if (statusCode >= 400) {
      logger.warn('Request completed', logPayload);
      return;
    }

    logger.info('Request completed', logPayload);
  });

  next();
}

/**
 * 全局错误处理中间件
 */
export function errorHandler(err: Error, req: Request, res: Response, _next: NextFunction) {
  const traceId = getTraceId(req, res);

  logger.error('Unhandled error', {
    trace_id: traceId,
    message: err.message,
    stack: err.stack,
  });

  res.status(500).json(buildErrorResponse(500, 'INTERNAL_SERVER_ERROR', 'Internal server error', traceId));
}
