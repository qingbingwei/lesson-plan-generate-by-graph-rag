import { Request, Response, NextFunction } from 'express';
import logger from '../../shared/utils/logger';

/**
 * 请求日志中间件
 */
export function requestLogger(req: Request, _res: Response, next: NextFunction) {
  logger.info('Incoming request', {
    method: req.method,
    path: req.path,
    ip: req.ip,
  });
  next();
}

/**
 * 全局错误处理中间件
 */
export function errorHandler(err: Error, _req: Request, res: Response, _next: NextFunction) {
  logger.error('Unhandled error', { error: err });
  res.status(500).json({
    success: false,
    error: 'Internal server error',
  });
}
