import type { TwitchLogger } from '$lib/server/twitch/types';

const discard = (): void => {};

export const noopTwitchLogger: TwitchLogger = {
  debug: discard,
  info: discard,
  warn: discard,
  error: discard,
};
