import { formatComment, formatEvent, HEARTBEAT_INTERVAL_MS } from '$lib/server/events/sse';
import { getServices } from '$lib/server/runtime/container';

import type { RequestHandler } from './$types';

export const GET: RequestHandler = ({ request }) => {
  const { overlay } = getServices();
  const subscription = overlay.subscribe();

  const stream = new ReadableStream({
    start: (controller) => {
      const encoder = new TextEncoder();
      const enqueue = (chunk: string): boolean => {
        try {
          controller.enqueue(encoder.encode(chunk));

          return true;
        } catch {
          return false;
        }
      };

      const cleanup = (): void => {
        subscription.close();
        clearInterval(heartbeat);
        request.signal.removeEventListener('abort', cleanup);
      };

      request.signal.addEventListener('abort', cleanup, { once: true });

      enqueue(formatComment('ping'));

      const heartbeat = setInterval(() => {
        if (!enqueue(formatComment('ping'))) {
          cleanup();
        }
      }, HEARTBEAT_INTERVAL_MS);

      (async () => {
        for await (const event of subscription) {
          if (!enqueue(formatEvent(event))) {
            break;
          }
        }

        cleanup();
      })();
    },
  });

  return new Response(stream, {
    headers: {
      'Content-Type': 'text/event-stream',
      'Cache-Control': 'no-cache',
      Connection: 'keep-alive',
    },
  });
};
