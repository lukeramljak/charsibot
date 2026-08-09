import { createServer } from 'node:http';

import { createApplicationRuntime } from '../src/lib/server/runtime/create';
import { ApplicationLifecycle } from '../src/lib/server/runtime/lifecycle';
import { readServerConfig } from './config';

// adapter-node creates this module before the custom server bundle is built.
import { handler } from './handler.js';

const config = readServerConfig(process.env);
const lifecycle = new ApplicationLifecycle(createApplicationRuntime);
const server = createServer(handler);

let shutdownPromise: Promise<void> | undefined;

const listen = async (): Promise<void> => {
  await lifecycle.start();
  await new Promise<void>((resolve, reject) => {
    const onError = (error: Error) => {
      server.off('listening', onListening);
      reject(error);
    };
    const onListening = () => {
      server.off('error', onError);
      resolve();
    };

    server.once('error', onError);
    server.once('listening', onListening);
    server.listen(config.port, config.host);
  });
};

const shutdown = (reason: string): Promise<void> => {
  if (shutdownPromise) return shutdownPromise;

  shutdownPromise = (async () => {
    const forceClose = setTimeout(
      () => server.closeAllConnections(),
      config.shutdownTimeoutMilliseconds,
    );
    forceClose.unref();

    try {
      if (server.listening) {
        await new Promise<void>((resolve, reject) => {
          server.close((error) => (error ? reject(error) : resolve()));
        });
      }
      await lifecycle.stop(reason);
    } finally {
      clearTimeout(forceClose);
    }
  })();

  return shutdownPromise;
};

for (const signal of ['SIGINT', 'SIGTERM'] as const) {
  process.once(signal, () => {
    void shutdown(signal).catch((error: unknown) => {
      console.error('failed to stop application', error);
      process.exitCode = 1;
    });
  });
}

try {
  await listen();
  console.info(`charsibot listening on http://${config.host}:${config.port}`);
} catch (error) {
  console.error('failed to start application', error);
  await shutdown('startup failure').catch(() => undefined);
  process.exitCode = 1;
}
