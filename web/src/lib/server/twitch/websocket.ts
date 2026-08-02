import WebSocket, { type RawData } from 'ws';

import { EventSubSocketClosedError } from '$lib/server/twitch/errors';
import type {
  EventSubConnection,
  EventSubConnector,
  EventSubSocketClose,
} from '$lib/server/twitch/types';

interface Waiter {
  resolve: (message: string) => void;
  reject: (error: unknown) => void;
  cleanup: () => void;
}

const abortError = (signal?: AbortSignal): unknown => {
  return signal?.reason ?? new DOMException('The operation was aborted', 'AbortError');
};

const messageText = (data: RawData): string => {
  if (Array.isArray(data)) {
    return Buffer.concat(data).toString('utf8');
  }

  return Buffer.from(data as ArrayBuffer).toString('utf8');
};

const wrapSocket = (socket: WebSocket): EventSubConnection => {
  const messages: string[] = [];
  const waiters: Waiter[] = [];
  let terminalError: unknown;
  let closeInfo: EventSubSocketClose | undefined;
  let resolveClosed: (close: EventSubSocketClose) => void = () => {};
  const closed = new Promise<EventSubSocketClose>((resolve) => {
    resolveClosed = resolve;
  });

  const rejectWaiters = (error: unknown): void => {
    for (const waiter of waiters.splice(0)) {
      waiter.cleanup();
      waiter.reject(error);
    }
  };

  socket.on('message', (data) => {
    const message = messageText(data);
    const waiter = waiters.shift();

    if (waiter) {
      waiter.cleanup();
      waiter.resolve(message);
      return;
    }

    messages.push(message);
  });

  socket.on('error', (error) => {
    terminalError = error;
    rejectWaiters(error);
  });

  socket.on('close', (code, reason) => {
    closeInfo = { code, reason: reason.toString('utf8') };
    resolveClosed(closeInfo);
    rejectWaiters(terminalError ?? new EventSubSocketClosedError(closeInfo));
  });

  const next = (signal?: AbortSignal): Promise<string> => {
    const queued = messages.shift();

    if (queued !== undefined) {
      return Promise.resolve(queued);
    }

    if (terminalError) {
      return Promise.reject(terminalError);
    }

    if (closeInfo) {
      return Promise.reject(new EventSubSocketClosedError(closeInfo));
    }

    if (signal?.aborted) {
      return Promise.reject(abortError(signal));
    }

    return new Promise<string>((resolve, reject) => {
      const onAbort = (): void => {
        const index = waiters.indexOf(waiter);

        if (index >= 0) {
          waiters.splice(index, 1);
        }

        reject(abortError(signal));
      };
      const waiter: Waiter = {
        resolve,
        reject,
        cleanup: () => signal?.removeEventListener('abort', onAbort),
      };

      signal?.addEventListener('abort', onAbort, { once: true });
      waiters.push(waiter);
    });
  };

  const close = (code = 1_000, reason = 'Stopping EventSub connection'): void => {
    if (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING) {
      socket.close(code, reason);
    }
  };

  return { next, close, closed };
};

export const createNodeEventSubConnector = (): EventSubConnector => {
  return (url: string, signal?: AbortSignal): Promise<EventSubConnection> => {
    if (signal?.aborted) {
      return Promise.reject(abortError(signal));
    }

    return new Promise<EventSubConnection>((resolve, reject) => {
      const socket = new WebSocket(url, { handshakeTimeout: 10_000 });
      let settled = false;

      const cleanup = (): void => {
        signal?.removeEventListener('abort', onAbort);
        socket.off('open', onOpen);
        socket.off('error', onError);
      };

      const onOpen = (): void => {
        settled = true;
        cleanup();
        resolve(wrapSocket(socket));
      };

      const onError = (error: Error): void => {
        if (!settled) {
          settled = true;
          cleanup();
          reject(error);
        }
      };

      const onAbort = (): void => {
        if (!settled) {
          settled = true;
          cleanup();
          socket.close();
          reject(abortError(signal));
        }
      };

      socket.once('open', onOpen);
      socket.once('error', onError);
      signal?.addEventListener('abort', onAbort, { once: true });
    });
  };
};
