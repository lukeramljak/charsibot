import type { OverlayEvent } from '$lib/contracts/overlay';
import type { Logger, OverlayBus, OverlaySubscription } from '$lib/server/application/ports';

const CLIENT_CAPACITY = 10;

interface SubscriptionState {
  buffer: OverlayEvent[];
  wake: (() => void) | undefined;
  closed: boolean;
}

const createSubscription = (state: SubscriptionState): OverlaySubscription => {
  const close = (): void => {
    state.closed = true;
    state.wake?.();
  };

  const next = async (): Promise<IteratorResult<OverlayEvent>> => {
    while (state.buffer.length === 0 && !state.closed) {
      await new Promise<void>((resolve) => {
        state.wake = resolve;
      });
      state.wake = undefined;
    }

    if (state.buffer.length > 0) {
      return { done: false, value: state.buffer.shift()! };
    }

    return { done: true, value: undefined as never };
  };

  return {
    close,
    [Symbol.asyncIterator]: () => ({ next }),
  };
};

export const createOverlayBus = (logger: Logger): OverlayBus => {
  const clients = new Set<SubscriptionState>();
  let closed = false;

  const publish = (event: OverlayEvent): void => {
    if (closed) return;

    for (const client of clients) {
      if (client.buffer.length >= CLIENT_CAPACITY) {
        logger.warn('SSE client buffer full, dropping event', { type: event.type });
        continue;
      }

      client.buffer.push(event);
      client.wake?.();
    }
  };

  const subscribe = (): OverlaySubscription => {
    if (closed) {
      throw new Error('overlay bus is closed');
    }

    const state: SubscriptionState = {
      buffer: [],
      wake: undefined,
      closed: false,
    };

    clients.add(state);

    const subscription = createSubscription(state);
    const originalClose = subscription.close;

    return {
      ...subscription,
      close: () => {
        clients.delete(state);
        originalClose();
      },
    };
  };

  const close = (): void => {
    closed = true;

    for (const client of clients) {
      client.closed = true;
      client.wake?.();
    }

    clients.clear();
  };

  return { publish, subscribe, close };
};
