import app from './app';
import config from './config';
import logger from './shared/utils/logger';

const PORT = config.port;

if (config.langsmith.enabled) {
  if (config.langsmith.apiKey) {
    logger.info('LangSmith tracing enabled', {
      project: config.langsmith.project,
      endpoint: config.langsmith.endpoint,
    });
  } else {
    logger.warn('LangSmith tracing is enabled but LANGSMITH_API_KEY is missing; tracing disabled at runtime');
  }
}

app.listen(PORT, () => {
  logger.info('Agent service started', {
    port: PORT,
    env: config.env,
  });
});
