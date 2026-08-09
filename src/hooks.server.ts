import type { Handle } from '@sveltejs/kit';

import { createApplicationRuntime } from '$lib/server/runtime/create';
import { ApplicationLifecycle } from '$lib/server/runtime/lifecycle';

// In production the custom server entrypoint (server/index.ts) owns the lifecycle.
// In dev, SvelteKit's dev server doesn't run that entrypoint, so we boot it here
// using $env/dynamic/private since Vite doesn't populate process.env from .env files.
const init = import.meta.env.DEV
	? (async () => {
			const { env } = await import('$env/dynamic/private');
			await new ApplicationLifecycle(() => createApplicationRuntime(env as NodeJS.ProcessEnv))
				.start()
				.catch((error: unknown) => {
					console.error('failed to start application runtime', error);
					process.exitCode = 1;
				});
		})()
	: undefined;

export const handle: Handle = async ({ event, resolve }) => {
	await init;

	return resolve(event);
};
