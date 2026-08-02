import { describe, expect, it, vi } from 'vitest';

import { TwitchHttpError } from '$lib/server/twitch/errors';
import { createHelixClient, createTwitchChatClient } from '$lib/server/twitch/helix';
import type { AppTokenProvider, HelixClient } from '$lib/server/twitch/types';

const helixClient = (request: (...args: never[]) => Promise<unknown>): HelixClient => {
  return { request: request as HelixClient['request'] };
};

const tokenProvider = (tokens: string[]): AppTokenProvider => {
  let index = 0;

  return {
    get: vi.fn(async () => tokens[Math.min(index++, tokens.length - 1)]),
    invalidate: vi.fn(),
    maintain: vi.fn(async () => {}),
  };
};

describe('createHelixClient', () => {
  it('adds Twitch headers and refreshes exactly once after a 401', async () => {
    const tokens = tokenProvider(['expired', 'fresh']);
    const request = vi
      .fn<typeof fetch>()
      .mockResolvedValueOnce(new Response('expired', { status: 401 }))
      .mockResolvedValueOnce(Response.json({ data: [{ id: 'ok' }] }));
    const helix = createHelixClient({
      clientId: 'client',
      tokenProvider: tokens,
      fetch: request,
      apiBaseUrl: 'https://mock.invalid/helix',
    });

    await expect(helix.request('/example')).resolves.toEqual({ data: [{ id: 'ok' }] });
    expect(tokens.invalidate).toHaveBeenCalledOnce();
    expect(request).toHaveBeenCalledTimes(2);
    expect(new Headers(request.mock.calls[0][1]?.headers).get('Authorization')).toBe(
      'Bearer expired',
    );
    expect(new Headers(request.mock.calls[1][1]?.headers).get('Authorization')).toBe(
      'Bearer fresh',
    );
    expect(new Headers(request.mock.calls[1][1]?.headers).get('Client-Id')).toBe('client');
  });

  it('does not retry a second 401', async () => {
    const request = vi
      .fn<typeof fetch>()
      .mockResolvedValueOnce(new Response('expired', { status: 401 }))
      .mockResolvedValueOnce(new Response('still expired', { status: 401 }));
    const helix = createHelixClient({
      clientId: 'client',
      tokenProvider: tokenProvider(['first', 'second']),
      fetch: request,
    });

    await expect(helix.request('/example')).rejects.toBeInstanceOf(TwitchHttpError);
    expect(request).toHaveBeenCalledTimes(2);
  });
});

describe('createTwitchChatClient', () => {
  it('sends app-token chat messages source-only with optional reply IDs', async () => {
    const request = vi.fn(async () => ({
      data: [{ message_id: 'message-1', is_sent: true }],
    }));
    const chat = createTwitchChatClient({ helix: helixClient(request) });

    await expect(
      chat.sendMessage({
        broadcasterId: 'channel',
        senderId: 'bot',
        message: 'hello',
        replyParentMessageId: 'parent',
      }),
    ).resolves.toEqual({ messageId: 'message-1', isSent: true, dropReason: undefined });

    expect(request).toHaveBeenCalledWith(
      '/chat/messages',
      {
        method: 'POST',
        body: JSON.stringify({
          broadcaster_id: 'channel',
          sender_id: 'bot',
          message: 'hello',
          reply_parent_message_id: 'parent',
          for_source_only: true,
        }),
      },
      { signal: undefined },
    );
  });

  it('surfaces a successful HTTP response whose chat message was dropped', async () => {
    const request = vi.fn(async () => ({
      data: [
        {
          message_id: 'message-1',
          is_sent: false,
          drop_reason: { code: 'automod_held', message: 'held by AutoMod' },
        },
      ],
    }));
    const chat = createTwitchChatClient({ helix: helixClient(request) });

    await expect(
      chat.sendMessage({ broadcasterId: 'channel', senderId: 'bot', message: 'hello' }),
    ).rejects.toEqual(
      expect.objectContaining({
        code: 'automod_held',
        detail: 'held by AutoMod',
      }),
    );
  });
});
