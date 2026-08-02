import type { ApplicationRuntime } from '$lib/server/runtime/contracts';

/**
 * The dependency container is filled in by the data and Twitch checkpoints.
 * Keeping the empty runtime here lets the Node server lifecycle be exercised
 * without starting a second bot beside the Go reference implementation.
 */
export const createApplicationRuntime = async (): Promise<ApplicationRuntime> => {
  return {
    stop: async () => {},
  };
};
