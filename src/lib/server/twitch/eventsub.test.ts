import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { EventSubSocketClosedError } from '$lib/server/twitch/errors';
import { createEventSubTransport, parseEventSubMessage } from '$lib/server/twitch/eventsub';
import type {
  ConduitSessionManager,
  EventSubConnection,
  EventSubConnector,
  EventSubSocketClose,
} from '$lib/server/twitch/types';

interface FakeWaiter {
  resolve: (value: string) => void;
  reject: (error: unknown) => void;
  cleanup: () => void;
}

class FakeConnection implements EventSubConnection {
  readonly messages: string[] = [];
  readonly waiters: FakeWaiter[] = [];
  readonly closeCalls: Array<{ code?: number; reason?: string }> = [];
  readonly closed: Promise<EventSubSocketClose>;
  #resolveClosed: (close: EventSubSocketClose) => void = () => {};
  #close: EventSubSocketClose | undefined;

  constructor() {
    this.closed = new Promise((resolve) => {
      this.#resolveClosed = resolve;
    });
  }

  next = (signal?: AbortSignal): Promise<string> => {
    const message = this.messages.shift();
    if (message !== undefined) {
      return Promise.resolve(message);
    }

    if (this.#close) {
      return Promise.reject(new EventSubSocketClosedError(this.#close));
    }

    if (signal?.aborted) {
      return Promise.reject(signal.reason);
    }

    return new Promise<string>((resolve, reject) => {
      const onAbort = (): void => {
        const index = this.waiters.indexOf(waiter);
        if (index >= 0) {
          this.waiters.splice(index, 1);
        }

        reject(signal?.reason);
      };
      const waiter: FakeWaiter = {
        resolve,
        reject,
        cleanup: () => signal?.removeEventListener('abort', onAbort),
      };

      signal?.addEventListener('abort', onAbort, { once: true });
      this.waiters.push(waiter);
    });
  };

  close = (code = 1_000, reason = ''): void => {
    this.closeCalls.push({ code, reason });
    if (this.#close) {
      return;
    }

    this.#close = { code, reason };
    this.#resolveClosed(this.#close);
    for (const waiter of this.waiters.splice(0)) {
      waiter.cleanup();
      waiter.reject(new EventSubSocketClosedError(this.#close));
    }
  };

  emit = (message: unknown): void => {
    const raw = typeof message === 'string' ? message : JSON.stringify(message);
    const waiter = this.waiters.shift();
    if (waiter) {
      waiter.cleanup();
      waiter.resolve(raw);
      return;
    }

    this.messages.push(raw);
  };
}

const metadata = (messageType: string, messageId: string) => ({
  message_id: messageId,
  message_type: messageType,
  message_timestamp: '2026-08-02T00:00:00Z',
});

const welcome = (id: string, keepaliveSeconds = 10) => ({
  metadata: metadata('session_welcome', `welcome-${id}`),
  payload: {
    session: {
      id,
      status: 'connected',
      keepalive_timeout_seconds: keepaliveSeconds,
      reconnect_url: null,
      connected_at: '2026-08-02T00:00:00Z',
    },
  },
});

const notification = (id: string, type = 'channel.chat.message') => ({
  metadata: metadata('notification', id),
  payload: {
    subscription: {
      id: `subscription-${type}`,
      status: 'enabled',
      type,
      version: '1',
      condition: {},
      transport: { method: 'conduit', conduit_id: 'conduit' },
      created_at: '2026-08-02T00:00:00Z',
      cost: 0,
    },
    event: { id },
  },
});

const reconnect = (url: string) => ({
  metadata: metadata('session_reconnect', 'reconnect'),
  payload: {
    session: {
      id: 'old',
      status: 'reconnecting',
      keepalive_timeout_seconds: null,
      reconnect_url: url,
      connected_at: '2026-08-02T00:00:00Z',
    },
  },
});

const revocation = () => ({
  metadata: metadata('revocation', 'revocation'),
  payload: {
    subscription: {
      ...notification('unused').payload.subscription,
      id: 'revoked-id',
      status: 'authorization_revoked',
    },
  },
});

const flush = async (): Promise<void> => {
  for (let index = 0; index < 20; index += 1) {
    await Promise.resolve();
  }
};

const setup = (connections: FakeConnection[], overrides: Record<string, unknown> = {}) => {
  const connector = vi.fn<EventSubConnector>(async () => {
    const connection = connections.shift();
    if (!connection) {
      throw new Error('No fake connection available');
    }

    return connection;
  });
  const prepareSession = vi.fn<ConduitSessionManager['prepareSession']>(async () => {});
  const onNotification = vi.fn(async () => {});
  const transport = createEventSubTransport({
    connector,
    sessionManager: { prepareSession },
    onNotification,
    reconnectDelayMilliseconds: 100,
    welcomeTimeoutMilliseconds: 1_000,
    reconnectTimeoutMilliseconds: 1_000,
    keepaliveGraceMilliseconds: 0,
    ...overrides,
  });

  return { connector, prepareSession, onNotification, transport };
};

describe('createEventSubTransport', () => {
  beforeEach(() => vi.useFakeTimers());
  afterEach(() => vi.useRealTimers());

  it('prepares a fresh session, reports readiness and suppresses duplicate notifications', async () => {
    const connection = new FakeConnection();
    const { onNotification, prepareSession, transport } = setup([connection]);
    const started = transport.start();

    connection.emit(welcome('session-1'));
    await started;
    expect(prepareSession).toHaveBeenCalledWith('session-1', expect.any(AbortSignal));
    expect(transport.readiness()).toMatchObject({
      ready: true,
      state: 'ready',
      sessionId: 'session-1',
    });

    connection.emit(notification('same-id'));
    connection.emit(notification('same-id'));
    await flush();
    expect(onNotification).toHaveBeenCalledOnce();

    await transport.stop();
  });

  it('bounds deduplication history by capacity', async () => {
    const connection = new FakeConnection();
    const { onNotification, transport } = setup([connection], { dedupeCapacity: 2 });
    const started = transport.start();
    connection.emit(welcome('session-1'));
    await started;

    connection.emit(notification('one'));
    connection.emit(notification('two'));
    connection.emit(notification('three'));
    connection.emit(notification('one'));
    await flush();

    expect(onNotification).toHaveBeenCalledTimes(4);
    await transport.stop();
  });

  it('expires duplicate IDs after the configured TTL', async () => {
    let now = 0;
    const connection = new FakeConnection();
    const { onNotification, transport } = setup([connection], {
      dedupeTtlMilliseconds: 600_000,
      now: () => now,
    });
    const started = transport.start();
    connection.emit(welcome('session-1'));
    await started;

    connection.emit(notification('same'));
    connection.emit(notification('same'));
    await flush();
    expect(onNotification).toHaveBeenCalledOnce();

    now = 600_001;
    connection.emit(notification('same'));
    await flush();
    expect(onNotification).toHaveBeenCalledTimes(2);

    await transport.stop();
  });

  it('rejects malformed notification subscription envelopes at the parser boundary', () => {
    expect(() =>
      parseEventSubMessage(
        JSON.stringify({
          metadata: metadata('notification', 'bad'),
          payload: { subscription: { id: 'incomplete' }, event: {} },
        }),
      ),
    ).toThrow('invalid subscription');
  });

  it('performs a seamless reconnect while continuing to process the old socket', async () => {
    const oldConnection = new FakeConnection();
    const newConnection = new FakeConnection();
    const { connector, onNotification, prepareSession, transport } = setup([
      oldConnection,
      newConnection,
    ]);
    const started = transport.start();
    oldConnection.emit(welcome('old'));
    await started;

    oldConnection.emit(reconnect('wss://reconnect.invalid/session'));
    await flush();
    oldConnection.emit(notification('during-handoff'));
    await flush();
    newConnection.emit(welcome('new'));
    await flush();

    expect(connector).toHaveBeenNthCalledWith(
      2,
      'wss://reconnect.invalid/session',
      expect.any(AbortSignal),
    );
    expect(onNotification).toHaveBeenCalledOnce();
    expect(prepareSession).toHaveBeenCalledTimes(1);
    expect(oldConnection.closeCalls).toContainEqual({
      code: 1_000,
      reason: 'EventSub reconnect handoff complete',
    });
    expect(transport.readiness()).toMatchObject({ ready: true, sessionId: 'new' });

    await transport.stop();
  });

  it('bounds reconnect socket establishment by the handoff deadline', async () => {
    const oldConnection = new FakeConnection();
    const connector = vi.fn<EventSubConnector>(async (_url, signal) => {
      if (connector.mock.calls.length === 1) {
        return oldConnection;
      }

      return new Promise<EventSubConnection>((_resolve, reject) => {
        signal?.addEventListener('abort', () => reject(signal.reason), { once: true });
      });
    });
    const transport = createEventSubTransport({
      connector,
      sessionManager: { prepareSession: async () => {} },
      onNotification: async () => {},
      reconnectDelayMilliseconds: 100,
      welcomeTimeoutMilliseconds: 1_000,
      reconnectTimeoutMilliseconds: 1_000,
    });
    const started = transport.start();
    oldConnection.emit(welcome('old'));
    await started;

    oldConnection.emit(reconnect('wss://reconnect.invalid/session'));
    await flush();
    await vi.advanceTimersByTimeAsync(1_000);

    expect(transport.readiness()).toMatchObject({
      ready: false,
      state: 'reconnecting',
      lastError: 'Timed out connecting the Twitch EventSub reconnect socket',
    });

    await transport.stop();
  });

  it('reconnects fresh after a keepalive timeout', async () => {
    const first = new FakeConnection();
    const second = new FakeConnection();
    const { connector, prepareSession, transport } = setup([first, second]);
    const started = transport.start();
    first.emit(welcome('first', 1));
    await started;

    await vi.advanceTimersByTimeAsync(1_100);
    second.emit(welcome('second', 1));
    await flush();

    expect(connector).toHaveBeenCalledTimes(2);
    expect(prepareSession).toHaveBeenCalledTimes(2);
    expect(transport.readiness()).toMatchObject({ ready: true, sessionId: 'second' });

    await transport.stop();
  });

  it.each([
    ['revocation', revocation()],
    [
      'unknown message',
      { metadata: metadata('future_message', 'future'), payload: { future: true } },
    ],
  ])('does not reset the keepalive watchdog for %s frames', async (_label, frame) => {
    const first = new FakeConnection();
    const second = new FakeConnection();
    const { connector, transport } = setup([first, second]);
    const started = transport.start();
    first.emit(welcome('first', 1));
    await started;

    await vi.advanceTimersByTimeAsync(800);
    first.emit(frame);
    await flush();
    await vi.advanceTimersByTimeAsync(201);

    expect(transport.readiness()).toMatchObject({
      ready: false,
      state: 'reconnecting',
      lastError: 'Twitch EventSub keepalive timed out',
    });
    expect(connector).toHaveBeenCalledOnce();

    await transport.stop();
  });

  it('resets the keepalive watchdog for notification frames', async () => {
    const connection = new FakeConnection();
    const { transport } = setup([connection]);
    const started = transport.start();
    connection.emit(welcome('session', 1));
    await started;

    await vi.advanceTimersByTimeAsync(800);
    connection.emit(notification('activity'));
    await flush();
    await vi.advanceTimersByTimeAsync(300);

    expect(transport.readiness()).toMatchObject({ ready: true, state: 'ready' });
    await transport.stop();
  });

  it('takes the ordinary reconnect path when Twitch disables the conduit shard', async () => {
    const first = new FakeConnection();
    const second = new FakeConnection();
    const { connector, onNotification, transport } = setup([first, second]);
    const started = transport.start();
    first.emit(welcome('first'));
    await started;

    first.emit(notification('disabled', 'conduit.shard.disabled'));
    await flush();
    await vi.advanceTimersByTimeAsync(100);
    second.emit(welcome('second'));
    await flush();

    expect(connector).toHaveBeenCalledTimes(2);
    expect(onNotification).not.toHaveBeenCalled();
    expect(transport.readiness()).toMatchObject({ ready: true, sessionId: 'second' });

    await transport.stop();
  });

  it('degrades readiness for a revocation without reconnecting', async () => {
    const connection = new FakeConnection();
    const onRevocation = vi.fn(async () => {});
    const { connector, transport } = setup([connection], { onRevocation });
    const started = transport.start();
    connection.emit(welcome('session'));
    await started;

    connection.emit(revocation());
    await flush();

    expect(onRevocation).toHaveBeenCalledOnce();
    expect(connector).toHaveBeenCalledOnce();
    expect(transport.readiness()).toMatchObject({
      ready: false,
      state: 'ready',
      lastError: expect.stringContaining('authorization_revoked'),
    });

    await transport.stop();
  });

  it('cancels a welcome wait and awaits in-flight notification handlers on shutdown', async () => {
    const connection = new FakeConnection();
    let finishHandler: () => void = () => {};
    const handler = new Promise<void>((resolve) => {
      finishHandler = resolve;
    });
    const { transport } = setup([connection], {
      onNotification: () => handler,
    });
    const started = transport.start();
    connection.emit(welcome('session'));
    await started;
    connection.emit(notification('working'));
    await flush();

    let stopped = false;
    const stopping = transport.stop().then(() => {
      stopped = true;
    });
    await flush();
    expect(stopped).toBe(false);

    finishHandler();
    await stopping;
    expect(stopped).toBe(true);
    expect(transport.readiness()).toMatchObject({ ready: false, state: 'stopped' });
  });

  it('fails the start wait when stopped before a welcome arrives', async () => {
    const connection = new FakeConnection();
    const { transport } = setup([connection]);
    const started = transport.start();

    await transport.stop('test stop');
    await expect(started).rejects.toMatchObject({ name: 'AbortError' });
  });
});
