import type { Handle } from '@sveltejs/kit';

import { createApplicationRuntime } from '$lib/server/runtime/create';
import { ApplicationLifecycle } from '$lib/server/runtime/lifecycle';

// In production the custom server entrypoint (server/index.ts) owns the lifecycle.
// In dev, SvelteKit's dev server doesn't run that entrypoint, so we boot it here.
const init = import.meta.env.DEV
	? (async () => {
			const env = await import('$app/env/private');
			await new ApplicationLifecycle(() => createApplicationRuntime(env))
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
