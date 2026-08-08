import { describe, expect, it, vi } from 'vitest';

import type { SendChatMessageParams, TwitchChatClient } from '$lib/server/twitch/types';

import { createTwitchChatSender } from './chat-sender';

const createClientFake = () => {
  const calls: Array<{ params: SendChatMessageParams; signal?: AbortSignal }> = [];

  const client: TwitchChatClient = {
    sendMessage: vi.fn(async (params, signal) => {
      calls.push({ params, signal });

      return { messageId: 'msg-1', isSent: true };
    }),
  };

  return { client, calls };
};

describe('createTwitchChatSender', () => {
  it('maps ChatMessage to SendChatMessageParams', async () => {
    const { client, calls } = createClientFake();
    const sender = createTwitchChatSender({
      client,
      broadcasterId: 'channel-123',
      senderId: 'bot-456',
    });

    await sender.send({ message: 'hello world' });

    expect(calls).toHaveLength(1);
    expect(calls[0].params).toEqual({
      broadcasterId: 'channel-123',
      senderId: 'bot-456',
      message: 'hello world',
      replyParentMessageId: undefined,
    });
  });

  it('forwards replyParentMessageID', async () => {
    const { client, calls } = createClientFake();
    const sender = createTwitchChatSender({
      client,
      broadcasterId: 'channel-123',
      senderId: 'bot-456',
    });

    await sender.send({ message: 'reply', replyParentMessageID: 'parent-789' });

    expect(calls[0].params.replyParentMessageId).toBe('parent-789');
  });

  it('forwards the abort signal', async () => {
    const { client, calls } = createClientFake();
    const sender = createTwitchChatSender({
      client,
      broadcasterId: 'channel-123',
      senderId: 'bot-456',
    });

    const controller = new AbortController();
    await sender.send({ message: 'test' }, controller.signal);

    expect(calls[0].signal).toBe(controller.signal);
  });

  it('propagates errors from the underlying client', async () => {
    const client: TwitchChatClient = {
      sendMessage: vi.fn(async () => {
        throw new Error('rate limited');
      }),
    };
    const sender = createTwitchChatSender({
      client,
      broadcasterId: 'channel-123',
      senderId: 'bot-456',
    });

    await expect(sender.send({ message: 'test' })).rejects.toThrow('rate limited');
  });
});
