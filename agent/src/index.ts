import express from 'express';
import cors from 'cors';
import helmet from 'helmet';
import compression from 'compression';
import config from './config';
import logger from './utils/logger';
import routes from './routes';
import { requestLogger, errorHandler } from './middleware';

const app = express();

// 全局中间件
app.use(helmet());
app.use(cors());
app.use(compression());
app.use(express.json({ limit: '10mb' }));
app.use(requestLogger);

// 路由
app.use(routes);

// 全局错误处理
app.use(errorHandler);

// 启动服务器
const PORT = config.port;

app.listen(PORT, () => {
  logger.info(`Agent service started`, {
    port: PORT,
    env: config.env,
  });
});

export default app;
