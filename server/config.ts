export interface ServerConfig {
  host: string;
  port: number;
  shutdownTimeoutMilliseconds: number;
}

const DEFAULT_PORT = 8081;
const DEFAULT_SHUTDOWN_TIMEOUT_SECONDS = 30;

const positiveInteger = (value: string | undefined, fallback: number, name: string): number => {
  if (value === undefined || value === '') return fallback;

  const parsed = Number(value);
  if (!Number.isSafeInteger(parsed) || parsed <= 0) {
    throw new Error(`${name} must be a positive integer`);
  }
  return parsed;
};

export const readServerConfig = (environment: NodeJS.ProcessEnv): ServerConfig => {
  const port = positiveInteger(environment.PORT ?? environment.SERVER_PORT, DEFAULT_PORT, 'PORT');
  const shutdownTimeoutSeconds = positiveInteger(
    environment.SHUTDOWN_TIMEOUT,
    DEFAULT_SHUTDOWN_TIMEOUT_SECONDS,
    'SHUTDOWN_TIMEOUT',
  );

  return {
    host: environment.HOST || '0.0.0.0',
    port,
    shutdownTimeoutMilliseconds: shutdownTimeoutSeconds * 1_000,
  };
};
