import express from 'express';
import cors from 'cors';
import helmet from 'helmet';
import compression from 'compression';
import routes from './api/routes';
import { requestLogger, errorHandler } from './api/middleware';

const app = express();

app.use(helmet());
app.use(cors());
app.use(compression());
app.use(express.json({ limit: '10mb' }));
app.use(requestLogger);
app.use(routes);
app.use(errorHandler);

export default app;
