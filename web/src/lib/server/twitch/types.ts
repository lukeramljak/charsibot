export interface TwitchLogger {
  debug: (message: string, attributes?: Record<string, unknown>) => void;
  info: (message: string, attributes?: Record<string, unknown>) => void;
  warn: (message: string, attributes?: Record<string, unknown>) => void;
  error: (message: string, attributes?: Record<string, unknown>) => void;
}

export interface AppTokenProvider {
  get: (signal?: AbortSignal) => Promise<string>;
  invalidate: () => void;
  maintain: (signal: AbortSignal) => Promise<void>;
}

export interface HelixResponse<T> {
  data: T[];
}

export interface HelixClient {
  request: <T>(path: string, init?: RequestInit, options?: HelixRequestOptions) => Promise<T>;
}

export interface HelixRequestOptions {
  acceptedStatuses?: readonly number[];
  signal?: AbortSignal;
}

export interface SendChatMessageParams {
  broadcasterId: string;
  senderId: string;
  message: string;
  replyParentMessageId?: string;
  forSourceOnly?: boolean;
}

export interface SendChatMessageResult {
  messageId: string;
  isSent: boolean;
  dropReason?: {
    code: string;
    message: string;
  };
}

export interface TwitchChatClient {
  sendMessage: (
    params: SendChatMessageParams,
    signal?: AbortSignal,
  ) => Promise<SendChatMessageResult>;
}

export interface ConduitSessionManager {
  prepareSession: (sessionId: string, signal?: AbortSignal) => Promise<void>;
}

export interface EventSubMetadata {
  message_id: string;
  message_type:
    | 'session_welcome'
    | 'session_keepalive'
    | 'notification'
    | 'session_reconnect'
    | 'revocation'
    | string;
  message_timestamp: string;
  subscription_type?: string;
  subscription_version?: string;
}

export interface EventSubSession {
  id: string;
  status: string;
  keepalive_timeout_seconds: number | null;
  reconnect_url: string | null;
  connected_at: string;
}

export interface EventSubSubscription {
  id: string;
  status: string;
  type: string;
  version: string;
  condition: Record<string, string>;
  transport: Record<string, unknown>;
  created_at: string;
  cost: number;
}

export interface EventSubWelcomeMessage {
  metadata: EventSubMetadata;
  payload: { session: EventSubSession };
}

export interface EventSubKeepaliveMessage {
  metadata: EventSubMetadata;
  payload: Record<string, never>;
}

export interface EventSubNotificationMessage<TEvent = unknown> {
  metadata: EventSubMetadata;
  payload: {
    subscription: EventSubSubscription;
    event: TEvent;
  };
}

export interface EventSubReconnectMessage {
  metadata: EventSubMetadata;
  payload: { session: EventSubSession };
}

export interface EventSubRevocationMessage {
  metadata: EventSubMetadata;
  payload: { subscription: EventSubSubscription };
}

export type EventSubMessage =
  | EventSubWelcomeMessage
  | EventSubKeepaliveMessage
  | EventSubNotificationMessage
  | EventSubReconnectMessage
  | EventSubRevocationMessage;

export interface EventSubSocketClose {
  code: number;
  reason: string;
}

export interface EventSubConnection {
  next: (signal?: AbortSignal) => Promise<string>;
  close: (code?: number, reason?: string) => void;
  closed: Promise<EventSubSocketClose>;
}

export type EventSubConnector = (url: string, signal?: AbortSignal) => Promise<EventSubConnection>;

export type EventSubTransportState =
  | 'idle'
  | 'connecting'
  | 'ready'
  | 'reconnecting'
  | 'stopping'
  | 'stopped';

export interface EventSubReadiness {
  ready: boolean;
  state: EventSubTransportState;
  sessionId?: string;
  lastError?: string;
}

export interface EventSubTransport {
  start: (signal?: AbortSignal) => Promise<void>;
  stop: (reason?: string) => Promise<void>;
  readiness: () => EventSubReadiness;
}
