import express from 'express';
import cors from 'cors';
import helmet from 'helmet';
import compression from 'compression';
import routes from './api/routes';
import { traceMiddleware, requestLogger, errorHandler, TRACE_ID_HEADER, REQUEST_ID_HEADER } from './api/middleware';

const app = express();

app.use(helmet());
app.use(
  cors({
    origin: true,
    credentials: true,
    allowedHeaders: [
      'Content-Type',
      'Authorization',
      TRACE_ID_HEADER,
      REQUEST_ID_HEADER,
      'X-Generation-Api-Key',
      'X-Embedding-Api-Key',
    ],
    exposedHeaders: [TRACE_ID_HEADER, REQUEST_ID_HEADER],
  })
);
app.use(compression());
app.use(express.json({ limit: '10mb' }));
app.use(traceMiddleware);
app.use(requestLogger);
app.use(routes);
app.use(errorHandler);

export default app;
