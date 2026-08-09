import type { OverlayEvent } from '$lib/contracts/overlay';

export const HEARTBEAT_INTERVAL_MS = 30_000;

export const formatComment = (text: string): string => `: ${text}\n\n`;

export const formatEvent = (event: OverlayEvent): string => {
  const { type, ...data } = event;

  return `event: ${type}\ndata: ${JSON.stringify(data)}\n\n`;
};
