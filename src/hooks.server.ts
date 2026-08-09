import type { ServerInit } from '@sveltejs/kit';

import { createApplicationRuntime } from '$lib/server/runtime/create';
import { ApplicationLifecycle } from '$lib/server/runtime/lifecycle';

const lifecycle = new ApplicationLifecycle(async () => {
  const env = await import('$app/env/private');
  return createApplicationRuntime(env);
});

process.once('sveltekit:shutdown', (reason: string) =>
  lifecycle.stop(reason).catch((error: unknown) => {
    console.error('failed to stop application runtime', error);
  }),
);

export const init: ServerInit = async () => {
  try {
    await lifecycle.start();
  } catch (error) {
    console.error('failed to start application runtime', error);
    throw error;
  }
};
