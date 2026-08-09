import { TwitchHttpError } from '$lib/server/twitch/errors';
import type { AppTokenProvider } from '$lib/server/twitch/types';

export interface AppTokenProviderOptions {
  clientId: string;
  clientSecret: string;
  fetch?: typeof fetch;
  tokenUrl?: string;
  validateUrl?: string;
  now?: () => number;
  expirySkewMilliseconds?: number;
  validationIntervalMilliseconds?: number;
}

interface CachedToken {
  value: string;
  expiresAt: number;
  validatedAt: number;
}

const DEFAULT_TOKEN_URL = 'https://id.twitch.tv/oauth2/token';
const DEFAULT_VALIDATE_URL = 'https://id.twitch.tv/oauth2/validate';
const DEFAULT_EXPIRY_SKEW_MILLISECONDS = 60_000;
const DEFAULT_VALIDATION_INTERVAL_MILLISECONDS = 60 * 60 * 1_000;

const isRecord = (value: unknown): value is Record<string, unknown> => {
  return typeof value === 'object' && value !== null;
};

const parseTokenResponse = (value: unknown): { accessToken: string; expiresIn: number } => {
  if (!isRecord(value) || typeof value.access_token !== 'string') {
    throw new Error('Twitch token response did not contain an access token');
  }

  if (
    typeof value.expires_in !== 'number' ||
    !Number.isFinite(value.expires_in) ||
    value.expires_in <= 0
  ) {
    throw new Error('Twitch token response did not contain a valid expiry');
  }

  if (value.token_type !== 'bearer') {
    throw new Error('Twitch token response did not identify a bearer token');
  }

  return { accessToken: value.access_token, expiresIn: value.expires_in };
};

const abortReason = (signal: AbortSignal): unknown => {
  return signal.reason ?? new DOMException('The operation was aborted', 'AbortError');
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

export const createAppTokenProvider = (options: AppTokenProviderOptions): AppTokenProvider => {
  const request = options.fetch ?? globalThis.fetch;
  const tokenUrl = options.tokenUrl ?? DEFAULT_TOKEN_URL;
  const validateUrl = options.validateUrl ?? DEFAULT_VALIDATE_URL;
  const now = options.now ?? Date.now;
  const expirySkew = options.expirySkewMilliseconds ?? DEFAULT_EXPIRY_SKEW_MILLISECONDS;
  const validationInterval =
    options.validationIntervalMilliseconds ?? DEFAULT_VALIDATION_INTERVAL_MILLISECONDS;

  let cached: CachedToken | undefined;
  let pending: Promise<string> | undefined;

  const validateToken = async (
    value: string,
    signal?: AbortSignal,
  ): Promise<CachedToken | undefined> => {
    const response = await request(validateUrl, {
      headers: { Authorization: `OAuth ${value}` },
      signal,
    });

    if (response.status === 401) {
      return undefined;
    }

    const body = await response.text();
    if (!response.ok) {
      throw new TwitchHttpError(
        `Twitch app token validation failed with status ${response.status}`,
        response.status,
        body,
      );
    }

    let decoded: unknown;
    try {
      decoded = JSON.parse(body);
    } catch (error) {
      throw new Error('Twitch token validation response was not valid JSON', { cause: error });
    }

    if (!isRecord(decoded) || decoded.client_id !== options.clientId) {
      throw new Error('Twitch token validation returned the wrong client ID');
    }

    if (
      typeof decoded.expires_in !== 'number' ||
      !Number.isFinite(decoded.expires_in) ||
      decoded.expires_in <= 0
    ) {
      throw new Error('Twitch token validation response did not contain a valid expiry');
    }

    const expiresAt = now() + Math.max(0, decoded.expires_in * 1_000 - expirySkew);

    return { value, expiresAt, validatedAt: now() };
  };

  const requestToken = async (signal?: AbortSignal): Promise<string> => {
    const requestBody = new URLSearchParams({
      client_id: options.clientId,
      client_secret: options.clientSecret,
      grant_type: 'client_credentials',
    });
    const response = await request(tokenUrl, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: requestBody,
      signal,
    });

    const responseBody = await response.text();

    if (!response.ok) {
      throw new TwitchHttpError(
        `Twitch app token request failed with status ${response.status}`,
        response.status,
        responseBody,
      );
    }

    let decoded: unknown;
    try {
      decoded = JSON.parse(responseBody);
    } catch (error) {
      throw new Error('Twitch token response was not valid JSON', { cause: error });
    }

    const token = parseTokenResponse(decoded);
    const validated = await validateToken(token.accessToken, signal);

    if (!validated) {
      throw new Error('Twitch rejected a newly issued app token');
    }

    cached = validated;

    return token.accessToken;
  };

  const get = (signal?: AbortSignal): Promise<string> => {
    if (cached && cached.expiresAt > now() && cached.validatedAt + validationInterval > now()) {
      return Promise.resolve(cached.value);
    }

    if (!pending) {
      pending = (async () => {
        if (cached && cached.expiresAt > now()) {
          const validated = await validateToken(cached.value, signal);

          if (validated) {
            cached = validated;
            return cached.value;
          }

          cached = undefined;
        }

        return requestToken(signal);
      })().finally(() => {
        pending = undefined;
      });
    }

    return pending;
  };

  const invalidate = (): void => {
    cached = undefined;
  };

  const maintain = async (signal: AbortSignal): Promise<void> => {
    try {
      await get(signal);

      while (!signal.aborted) {
        await delay(validationInterval, signal);
        await get(signal);
      }
    } catch (error) {
      if (!signal.aborted) {
        throw error;
      }
    }
  };

  return { get, invalidate, maintain };
};
