import { EventSubProtocolError } from '$lib/server/twitch/errors';
import { noopTwitchLogger } from '$lib/server/twitch/logger';
import type {
  ConduitSessionManager,
  EventSubConnection,
  EventSubConnector,
  EventSubMessage,
  EventSubNotificationMessage,
  EventSubReadiness,
  EventSubReconnectMessage,
  EventSubRevocationMessage,
  EventSubSession,
  EventSubTransport,
  EventSubTransportState,
  EventSubWelcomeMessage,
  TwitchLogger,
} from '$lib/server/twitch/types';

export interface EventSubTransportOptions {
  connector: EventSubConnector;
  sessionManager: ConduitSessionManager;
  onNotification: (
    message: EventSubNotificationMessage,
    signal: AbortSignal,
  ) => Promise<void> | void;
  onRevocation?: (message: EventSubRevocationMessage, signal: AbortSignal) => Promise<void> | void;
  onReadinessChange?: (readiness: EventSubReadiness) => void;
  logger?: TwitchLogger;
  websocketUrl?: string;
  reconnectDelayMilliseconds?: number;
  welcomeTimeoutMilliseconds?: number;
  reconnectTimeoutMilliseconds?: number;
  keepaliveGraceMilliseconds?: number;
  dedupeCapacity?: number;
  dedupeTtlMilliseconds?: number;
  now?: () => number;
}

interface ActiveSession {
  connection: EventSubConnection;
  session: EventSubSession;
}

interface Deferred<T> {
  promise: Promise<T>;
  resolve: (value: T | PromiseLike<T>) => void;
  reject: (reason?: unknown) => void;
}

interface HandoffRead {
  source: 'old' | 'new' | 'timeout';
  raw?: string;
  error?: unknown;
}

const DEFAULT_WEBSOCKET_URL = 'wss://eventsub.wss.twitch.tv/ws';
const DEFAULT_RECONNECT_DELAY_MILLISECONDS = 10_000;
const DEFAULT_WELCOME_TIMEOUT_MILLISECONDS = 10_000;
const DEFAULT_RECONNECT_TIMEOUT_MILLISECONDS = 25_000;
const DEFAULT_KEEPALIVE_GRACE_MILLISECONDS = 1_000;
const DEFAULT_DEDUPE_CAPACITY = 2_048;
const DEFAULT_DEDUPE_TTL_MILLISECONDS = 10 * 60 * 1_000;

const createDeferred = <T>(): Deferred<T> => {
  let resolve: Deferred<T>['resolve'] = () => {};
  let reject: Deferred<T>['reject'] = () => {};
  const promise = new Promise<T>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise;
    reject = rejectPromise;
  });

  return { promise, resolve, reject };
};

const isRecord = (value: unknown): value is Record<string, unknown> => {
  return typeof value === 'object' && value !== null;
};

const validateSubscription = (payload: Record<string, unknown>, messageType: string): void => {
  const subscription = payload.subscription;

  if (!isRecord(subscription)) {
    throw new EventSubProtocolError(
      `EventSub ${messageType} message did not contain a subscription`,
    );
  }

  if (
    typeof subscription.id !== 'string' ||
    typeof subscription.status !== 'string' ||
    typeof subscription.type !== 'string' ||
    typeof subscription.version !== 'string' ||
    typeof subscription.created_at !== 'string' ||
    typeof subscription.cost !== 'number' ||
    !isRecord(subscription.condition) ||
    !isRecord(subscription.transport) ||
    typeof subscription.transport.method !== 'string' ||
    !Object.values(subscription.condition).every((value) => typeof value === 'string')
  ) {
    throw new EventSubProtocolError(
      `EventSub ${messageType} message contained an invalid subscription`,
    );
  }
};

const validateSession = (payload: Record<string, unknown>, messageType: string): void => {
  const session = payload.session;

  if (
    !isRecord(session) ||
    typeof session.id !== 'string' ||
    typeof session.status !== 'string' ||
    typeof session.connected_at !== 'string' ||
    (typeof session.keepalive_timeout_seconds !== 'number' &&
      session.keepalive_timeout_seconds !== null) ||
    (typeof session.reconnect_url !== 'string' && session.reconnect_url !== null)
  ) {
    throw new EventSubProtocolError(`EventSub ${messageType} message contained an invalid session`);
  }
};

const parseMessage = (raw: string): EventSubMessage => {
  let value: unknown;
  try {
    value = JSON.parse(raw);
  } catch (error) {
    throw new EventSubProtocolError(`EventSub message was not valid JSON: ${String(error)}`);
  }

  if (!isRecord(value) || !isRecord(value.metadata)) {
    throw new EventSubProtocolError('EventSub message did not contain metadata');
  }

  if (
    typeof value.metadata.message_id !== 'string' ||
    typeof value.metadata.message_type !== 'string' ||
    typeof value.metadata.message_timestamp !== 'string'
  ) {
    throw new EventSubProtocolError('EventSub message metadata was invalid');
  }

  if (!isRecord(value.payload)) {
    throw new EventSubProtocolError('EventSub message did not contain a payload');
  }

  if (
    value.metadata.message_type === 'notification' ||
    value.metadata.message_type === 'revocation'
  ) {
    validateSubscription(value.payload, value.metadata.message_type);
  }

  if (
    value.metadata.message_type === 'session_welcome' ||
    value.metadata.message_type === 'session_reconnect'
  ) {
    validateSession(value.payload, value.metadata.message_type);
  }

  return value as unknown as EventSubMessage;
};

const welcomeFrom = (message: EventSubMessage): EventSubWelcomeMessage => {
  if (message.metadata.message_type !== 'session_welcome') {
    throw new EventSubProtocolError(
      `Expected EventSub welcome, received ${message.metadata.message_type}`,
    );
  }

  const welcome = message as EventSubWelcomeMessage;
  const session = welcome.payload.session;

  if (
    !isRecord(session) ||
    typeof session.id !== 'string' ||
    (typeof session.keepalive_timeout_seconds !== 'number' &&
      session.keepalive_timeout_seconds !== null)
  ) {
    throw new EventSubProtocolError('EventSub welcome contained an invalid session');
  }

  return welcome;
};

const abortReason = (signal: AbortSignal): unknown => {
  return signal.reason ?? new DOMException('The operation was aborted', 'AbortError');
};

const combinedSignal = (...signals: Array<AbortSignal | undefined>): AbortSignal => {
  return AbortSignal.any(signals.filter((signal): signal is AbortSignal => signal !== undefined));
};

const delay = (milliseconds: number, signal: AbortSignal): Promise<void> => {
  if (signal.aborted) {
    return Promise.reject(abortReason(signal));
  }

  return new Promise<void>((resolve, reject) => {
    const timeout = setTimeout(() => {
      signal.removeEventListener('abort', onAbort);
      resolve();
    }, milliseconds);
    const onAbort = (): void => {
      clearTimeout(timeout);
      reject(abortReason(signal));
    };

    signal.addEventListener('abort', onAbort, { once: true });
  });
};

const nextWithTimeout = async (
  connection: EventSubConnection,
  milliseconds: number,
  signal: AbortSignal,
  message: string,
): Promise<string> => {
  const timeoutController = new AbortController();
  const timeout = setTimeout(() => timeoutController.abort(new Error(message)), milliseconds);

  try {
    return await connection.next(combinedSignal(signal, timeoutController.signal));
  } finally {
    clearTimeout(timeout);
  }
};

const runWithTimeout = async <T>(
  operation: (signal: AbortSignal) => Promise<T>,
  milliseconds: number,
  signal: AbortSignal,
  message: string,
): Promise<T> => {
  const timeoutController = new AbortController();
  let timeout: ReturnType<typeof setTimeout> | undefined;
  const timedOut = new Promise<never>((_resolve, reject) => {
    timeout = setTimeout(() => {
      const error = new Error(message);

      timeoutController.abort(error);
      reject(error);
    }, milliseconds);
  });

  try {
    return await Promise.race([
      operation(combinedSignal(signal, timeoutController.signal)),
      timedOut,
    ]);
  } finally {
    clearTimeout(timeout);
  }
};

export const createEventSubTransport = (options: EventSubTransportOptions): EventSubTransport => {
  const logger = options.logger ?? noopTwitchLogger;
  const websocketUrl = options.websocketUrl ?? DEFAULT_WEBSOCKET_URL;
  const reconnectDelay = options.reconnectDelayMilliseconds ?? DEFAULT_RECONNECT_DELAY_MILLISECONDS;
  const welcomeTimeout = options.welcomeTimeoutMilliseconds ?? DEFAULT_WELCOME_TIMEOUT_MILLISECONDS;
  const reconnectTimeout =
    options.reconnectTimeoutMilliseconds ?? DEFAULT_RECONNECT_TIMEOUT_MILLISECONDS;
  const keepaliveGrace = options.keepaliveGraceMilliseconds ?? DEFAULT_KEEPALIVE_GRACE_MILLISECONDS;
  const dedupeCapacity = options.dedupeCapacity ?? DEFAULT_DEDUPE_CAPACITY;
  const dedupeTtl = options.dedupeTtlMilliseconds ?? DEFAULT_DEDUPE_TTL_MILLISECONDS;
  const now = options.now ?? Date.now;

  let state: EventSubTransportState = 'idle';
  let ready = false;
  let sessionId: string | undefined;
  let lastError: string | undefined;
  let controller: AbortController | undefined;
  let runPromise: Promise<void> | undefined;
  let startPromise: Promise<void> | undefined;
  let stopPromise: Promise<void> | undefined;
  let initialReady: Deferred<void> | undefined;
  const activeConnections = new Set<EventSubConnection>();
  const handlerTasks = new Set<Promise<void>>();
  const seenMessageIds = new Map<string, number>();
  const revokedSubscriptions = new Set<string>();

  const readiness = (): EventSubReadiness => ({
    ready,
    state,
    ...(sessionId ? { sessionId } : {}),
    ...(lastError ? { lastError } : {}),
  });

  const publishReadiness = (): void => {
    options.onReadinessChange?.(readiness());
  };

  const errorMessage = (error: unknown): string | undefined => {
    if (error === undefined) {
      return undefined;
    }

    return error instanceof Error ? error.message : String(error);
  };

  const setState = (
    nextState: EventSubTransportState,
    nextReady: boolean,
    error?: unknown,
  ): void => {
    state = nextState;
    ready = nextReady;
    lastError = errorMessage(error);

    publishReadiness();
  };

  const rememberMessage = (id: string): boolean => {
    const timestamp = now();

    for (const [seenId, seenAt] of seenMessageIds) {
      if (timestamp - seenAt <= dedupeTtl) {
        break;
      }

      seenMessageIds.delete(seenId);
    }

    const seenAt = seenMessageIds.get(id);
    if (seenAt !== undefined && timestamp - seenAt <= dedupeTtl) {
      seenMessageIds.delete(id);
      seenMessageIds.set(id, timestamp);
      return false;
    }

    seenMessageIds.delete(id);
    seenMessageIds.set(id, timestamp);
    while (seenMessageIds.size > dedupeCapacity) {
      const oldest = seenMessageIds.keys().next().value;

      if (oldest === undefined) {
        break;
      }

      seenMessageIds.delete(oldest);
    }

    return true;
  };

  const trackHandler = (operation: Promise<void> | void, label: string): void => {
    const task = Promise.resolve(operation)
      .catch((error: unknown) => {
        logger.error(`Twitch ${label} handler failed`, { error });
      })
      .finally(() => {
        handlerTasks.delete(task);
      });

    handlerTasks.add(task);
  };

  const dispatchMessage = (
    message: EventSubMessage,
    signal: AbortSignal,
  ): 'continue' | 'reconnect' => {
    const type = message.metadata.message_type;

    if (type === 'session_keepalive') {
      return 'continue';
    }

    if (type === 'notification') {
      if (!rememberMessage(message.metadata.message_id)) {
        logger.debug('Ignoring duplicate Twitch EventSub notification', {
          messageId: message.metadata.message_id,
        });
        return 'continue';
      }

      const notification = message as EventSubNotificationMessage;

      if (notification.payload.subscription.type === 'conduit.shard.disabled') {
        logger.warn('Twitch conduit shard disabled', { event: notification.payload.event });

        return 'reconnect';
      }

      trackHandler(options.onNotification(notification, signal), 'notification');

      return 'continue';
    }

    if (type === 'revocation') {
      if (!rememberMessage(message.metadata.message_id)) {
        return 'continue';
      }

      const revocation = message as EventSubRevocationMessage;
      revokedSubscriptions.add(revocation.payload.subscription.id);
      logger.warn('Twitch EventSub subscription revoked', {
        type: revocation.payload.subscription.type,
        status: revocation.payload.subscription.status,
      });
      setState(
        'ready',
        false,
        new Error(
          `Twitch subscription ${revocation.payload.subscription.type} revoked: ${revocation.payload.subscription.status}`,
        ),
      );

      if (options.onRevocation) {
        trackHandler(options.onRevocation(revocation, signal), 'revocation');
      }

      return 'continue';
    }

    if (type === 'session_welcome') {
      throw new EventSubProtocolError('Received an unexpected EventSub welcome message');
    }

    logger.warn('Ignoring unknown Twitch EventSub message type', { type });
    return 'continue';
  };

  const connect = async (url: string, signal: AbortSignal): Promise<EventSubConnection> => {
    const connection = await options.connector(url, signal);

    activeConnections.add(connection);
    void connection.closed.finally(() => activeConnections.delete(connection));

    return connection;
  };

  const openFreshSession = async (signal: AbortSignal): Promise<ActiveSession> => {
    setState('connecting', false);

    const connection = await connect(websocketUrl, signal);

    try {
      const raw = await nextWithTimeout(
        connection,
        welcomeTimeout,
        signal,
        'Timed out waiting for Twitch EventSub welcome',
      );
      const welcome = welcomeFrom(parseMessage(raw));

      await runWithTimeout(
        (prepareSignal) =>
          options.sessionManager.prepareSession(welcome.payload.session.id, prepareSignal),
        welcomeTimeout,
        signal,
        'Timed out preparing Twitch conduit session',
      );

      revokedSubscriptions.clear();
      sessionId = welcome.payload.session.id;
      setState('ready', true);
      initialReady?.resolve();
      logger.info('Twitch EventSub session ready', { sessionId });

      return { connection, session: welcome.payload.session };
    } catch (error) {
      connection.close(1_011, 'EventSub session setup failed');
      throw error;
    }
  };

  const readForHandoff = async (
    oldConnection: EventSubConnection | undefined,
    newConnection: EventSubConnection,
    milliseconds: number,
    signal: AbortSignal,
  ): Promise<HandoffRead> => {
    const oldController = new AbortController();
    const newController = new AbortController();
    const timeoutController = new AbortController();
    const timeout = delay(milliseconds, combinedSignal(signal, timeoutController.signal)).then(
      (): HandoffRead => ({ source: 'timeout' }),
    );
    const reads: Array<Promise<HandoffRead>> = [
      newConnection
        .next(combinedSignal(signal, newController.signal))
        .then((raw): HandoffRead => ({ source: 'new', raw }))
        .catch((error: unknown): HandoffRead => ({ source: 'new', error })),
      timeout,
    ];

    if (oldConnection) {
      reads.push(
        oldConnection
          .next(combinedSignal(signal, oldController.signal))
          .then((raw): HandoffRead => ({ source: 'old', raw }))
          .catch((error: unknown): HandoffRead => ({ source: 'old', error })),
      );
    }

    try {
      return await Promise.race(reads);
    } finally {
      oldController.abort();
      newController.abort();
      timeoutController.abort();
    }
  };

  const handoff = async (
    oldSession: ActiveSession,
    reconnect: EventSubReconnectMessage,
    signal: AbortSignal,
  ): Promise<ActiveSession> => {
    const reconnectUrl = reconnect.payload.session.reconnect_url;

    if (!reconnectUrl) {
      throw new EventSubProtocolError('Twitch reconnect message did not contain a reconnect URL');
    }

    setState('reconnecting', false);

    const startedAt = now();
    const newConnection = await runWithTimeout(
      (handoffSignal) => connect(reconnectUrl, handoffSignal),
      reconnectTimeout,
      signal,
      'Timed out connecting the Twitch EventSub reconnect socket',
    );
    let oldConnection: EventSubConnection | undefined = oldSession.connection;

    try {
      while (true) {
        const elapsed = now() - startedAt;
        const remaining = reconnectTimeout - elapsed;

        if (remaining <= 0) {
          throw new Error('Timed out during Twitch EventSub reconnect handoff');
        }

        const read = await readForHandoff(oldConnection, newConnection, remaining, signal);
        if (read.source === 'timeout') {
          throw new Error('Timed out during Twitch EventSub reconnect handoff');
        }

        if (read.error) {
          if (read.source === 'old') {
            oldConnection = undefined;
            continue;
          }

          throw read.error;
        }

        const message = parseMessage(read.raw as string);
        if (read.source === 'new') {
          const welcome = welcomeFrom(message);
          oldSession.connection.close(1_000, 'EventSub reconnect handoff complete');
          sessionId = welcome.payload.session.id;
          setState('ready', revokedSubscriptions.size === 0);
          logger.info('Twitch EventSub reconnect handoff complete', { sessionId });

          return { connection: newConnection, session: welcome.payload.session };
        }

        if (message.metadata.message_type === 'session_reconnect') {
          continue;
        }

        if (dispatchMessage(message, signal) === 'reconnect') {
          throw new Error('Twitch conduit shard was disabled during reconnect handoff');
        }
      }
    } catch (error) {
      newConnection.close(1_011, 'EventSub reconnect handoff failed');
      throw error;
    }
  };

  const runSession = async (initialSession: ActiveSession, signal: AbortSignal): Promise<void> => {
    let active = initialSession;
    let watchdogDeadline =
      now() + (active.session.keepalive_timeout_seconds ?? 10) * 1_000 + keepaliveGrace;

    while (!signal.aborted) {
      const remaining = Math.max(0, watchdogDeadline - now());
      const raw = await nextWithTimeout(
        active.connection,
        remaining,
        signal,
        'Twitch EventSub keepalive timed out',
      );
      const message = parseMessage(raw);

      if (message.metadata.message_type === 'session_reconnect') {
        active = await handoff(active, message as EventSubReconnectMessage, signal);
        watchdogDeadline =
          now() + (active.session.keepalive_timeout_seconds ?? 10) * 1_000 + keepaliveGrace;
        continue;
      }

      if (
        message.metadata.message_type === 'notification' ||
        message.metadata.message_type === 'session_keepalive'
      ) {
        watchdogDeadline =
          now() + (active.session.keepalive_timeout_seconds ?? 10) * 1_000 + keepaliveGrace;
      }

      if (dispatchMessage(message, signal) === 'reconnect') {
        active.connection.close(1_011, 'Twitch conduit shard disabled');
        throw new Error('Twitch conduit shard disabled');
      }
    }
  };

  const closeConnections = (reason: string): void => {
    for (const connection of [...activeConnections]) {
      connection.close(1_000, reason);
    }
  };

  const run = async (signal: AbortSignal): Promise<void> => {
    while (!signal.aborted) {
      try {
        const active = await openFreshSession(signal);
        await runSession(active, signal);
      } catch (error) {
        if (signal.aborted) {
          break;
        }

        logger.error('Twitch EventSub disconnected; reconnecting', {
          error,
          delayMilliseconds: reconnectDelay,
        });
        setState('reconnecting', false, error);
        closeConnections('Reconnecting EventSub');
        await delay(reconnectDelay, signal);
      }
    }
  };

  const start = (signal?: AbortSignal): Promise<void> => {
    if (startPromise) {
      return startPromise;
    }

    if (state === 'stopping' || state === 'stopped') {
      return Promise.reject(new Error('Twitch EventSub transport has already stopped'));
    }

    controller = new AbortController();
    const runSignal = combinedSignal(controller.signal, signal);
    initialReady = createDeferred<void>();
    startPromise = initialReady.promise;
    runPromise = run(runSignal)
      .catch((error: unknown) => {
        if (!runSignal.aborted) {
          initialReady?.reject(error);
          logger.error('Twitch EventSub transport failed', { error });
        }
      })
      .finally(() => {
        if (!ready) {
          initialReady?.reject(abortReason(runSignal));
        }

        sessionId = undefined;
        setState('stopped', false);
      });

    return startPromise;
  };

  const stop = (reason = 'Stopping Twitch EventSub'): Promise<void> => {
    if (stopPromise) {
      return stopPromise;
    }

    stopPromise = (async () => {
      setState('stopping', false);
      controller?.abort(new DOMException(reason, 'AbortError'));
      closeConnections(reason);

      await runPromise;
      await Promise.allSettled([...handlerTasks]);
      sessionId = undefined;
      setState('stopped', false);
    })();

    return stopPromise;
  };

  return { start, stop, readiness };
};

export { parseMessage as parseEventSubMessage };
