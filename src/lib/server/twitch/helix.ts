import { TwitchChatMessageDroppedError, TwitchHttpError } from '$lib/server/twitch/errors';
import { noopTwitchLogger } from '$lib/server/twitch/logger';
import type {
  AppTokenProvider,
  HelixClient,
  HelixRequestOptions,
  SendChatMessageParams,
  SendChatMessageResult,
  TwitchChatClient,
  TwitchLogger,
} from '$lib/server/twitch/types';

export interface HelixClientOptions {
  clientId: string;
  tokenProvider: AppTokenProvider;
  fetch?: typeof fetch;
  apiBaseUrl?: string;
}

export interface TwitchChatClientOptions {
  helix: HelixClient;
  logger?: TwitchLogger;
}

interface SendChatMessageResponse {
  data: Array<{
    message_id: string;
    is_sent: boolean;
    drop_reason?: { code: string; message: string };
  }>;
}

const DEFAULT_API_BASE_URL = 'https://api.twitch.tv/helix';

const combineHeaders = (clientId: string, token: string, init?: RequestInit): Headers => {
  const headers = new Headers(init?.headers);
  headers.set('Client-Id', clientId);
  headers.set('Authorization', `Bearer ${token}`);

  if (init?.body !== undefined && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }

  return headers;
};

const isAccepted = (response: Response, options?: HelixRequestOptions): boolean => {
  return response.ok || options?.acceptedStatuses?.includes(response.status) === true;
};

const decodeResponse = <T>(body: string): T => {
  if (body === '') {
    return undefined as T;
  }

  try {
    return JSON.parse(body) as T;
  } catch (error) {
    throw new Error('Twitch API response was not valid JSON', { cause: error });
  }
};

export const createHelixClient = (options: HelixClientOptions): HelixClient => {
  const requestImplementation = options.fetch ?? globalThis.fetch;
  const apiBaseUrl = options.apiBaseUrl ?? DEFAULT_API_BASE_URL;

  const execute = async <T>(
    path: string,
    init: RequestInit | undefined,
    requestOptions: HelixRequestOptions | undefined,
    token: string,
  ): Promise<{ response: Response; body: string; decoded?: T }> => {
    const url = new URL(path.replace(/^\//, ''), `${apiBaseUrl.replace(/\/$/, '')}/`);
    const response = await requestImplementation(url, {
      ...init,
      headers: combineHeaders(options.clientId, token, init),
      signal: requestOptions?.signal,
    });
    const body = await response.text();

    return {
      response,
      body,
      decoded: isAccepted(response, requestOptions) ? decodeResponse<T>(body) : undefined,
    };
  };

  const request = async <T>(
    path: string,
    init?: RequestInit,
    requestOptions?: HelixRequestOptions,
  ): Promise<T> => {
    let token = await options.tokenProvider.get(requestOptions?.signal);
    let result = await execute<T>(path, init, requestOptions, token);

    if (result.response.status === 401) {
      options.tokenProvider.invalidate();
      token = await options.tokenProvider.get(requestOptions?.signal);
      result = await execute<T>(path, init, requestOptions, token);
    }

    if (!isAccepted(result.response, requestOptions)) {
      throw new TwitchHttpError(
        `Twitch API request to ${path} failed with status ${result.response.status}`,
        result.response.status,
        result.body,
      );
    }

    return result.decoded as T;
  };

  return { request };
};

export const createTwitchChatClient = (options: TwitchChatClientOptions): TwitchChatClient => {
  const logger = options.logger ?? noopTwitchLogger;

  const sendMessage = async (
    params: SendChatMessageParams,
    signal?: AbortSignal,
  ): Promise<SendChatMessageResult> => {
    const body: Record<string, unknown> = {
      broadcaster_id: params.broadcasterId,
      sender_id: params.senderId,
      message: params.message,
    };

    if (params.replyParentMessageId) {
      body.reply_parent_message_id = params.replyParentMessageId;
    }

    body.for_source_only = params.forSourceOnly ?? true;

    const response = await options.helix.request<SendChatMessageResponse>(
      '/chat/messages',
      { method: 'POST', body: JSON.stringify(body) },
      { signal },
    );
    const result = response.data[0];

    if (!result || typeof result.message_id !== 'string' || typeof result.is_sent !== 'boolean') {
      throw new Error('Twitch chat response did not contain a send result');
    }

    if (!result.is_sent) {
      logger.warn('Twitch dropped chat message', {
        message: params.message,
        dropReason: result.drop_reason,
      });

      throw new TwitchChatMessageDroppedError(
        result.drop_reason?.code ?? 'unknown',
        result.drop_reason?.message ?? 'Twitch did not provide a reason',
      );
    }

    return {
      messageId: result.message_id,
      isSent: result.is_sent,
      dropReason: result.drop_reason,
    };
  };

  return { sendMessage };
};
