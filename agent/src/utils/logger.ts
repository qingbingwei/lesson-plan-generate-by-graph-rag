import winston from 'winston';
import config from '../config';

const { combine, timestamp, printf, colorize, json } = winston.format;

// 自定义日志格式 - 使用 winston.Logform.TransformableInfo 兼容类型
const customFormat = printf((info) => {
  const { level, message, timestamp: ts, ...meta } = info;
  const metaStr = Object.keys(meta).length ? JSON.stringify(meta) : '';
  return `${ts as string} [${level}]: ${message} ${metaStr}`;
});

// 创建日志记录器
const logger = winston.createLogger({
  level: config.log.level,
  format: combine(
    timestamp({ format: 'YYYY-MM-DD HH:mm:ss' }),
    config.log.format === 'json' ? json() : customFormat
  ),
  transports: [
    new winston.transports.Console({
      format: combine(
        colorize(),
        timestamp({ format: 'YYYY-MM-DD HH:mm:ss' }),
        customFormat
      ),
    }),
  ],
});

// 生产环境添加文件日志
if (config.env === 'production') {
  logger.add(
    new winston.transports.File({ filename: 'logs/error.log', level: 'error' })
  );
  logger.add(
    new winston.transports.File({ filename: 'logs/combined.log' })
  );
}

export default logger;
