import { describe, expect, it, vi } from 'vitest';

import type { Bot } from '$lib/server/bot';
import { createBotNotificationHandler } from '$lib/server/twitch/bot-adapter';
import type { EventSubNotificationMessage } from '$lib/server/twitch/types';

const createBotFake = (): Bot => ({
  handleChatMessage: vi.fn(async () => {}),
  handleRedemption: vi.fn(async () => {}),
  handleRaid: vi.fn(async () => {}),
  stop: vi.fn(async () => {}),
});

const message = (type: string, event: unknown): EventSubNotificationMessage => {
  return {
    metadata: {
      message_id: 'message',
      message_type: 'notification',
      message_timestamp: '2026-08-02T00:00:00Z',
    },
    payload: {
      subscription: {
        id: 'subscription',
        status: 'enabled',
        type,
        version: '1',
        condition: {},
        transport: {},
        created_at: '2026-08-02T00:00:00Z',
        cost: 0,
      },
      event,
    },
  };
};

describe('createBotNotificationHandler', () => {
  it('maps supported Twitch events onto normalized bot events', async () => {
    const bot = createBotFake();
    const handle = createBotNotificationHandler(bot);

    await handle(
      message('channel.chat.message', {
        chatter_user_id: 'user',
        chatter_user_name: 'Viewer',
        message_id: 'chat-message',
        message: { text: '!stats' },
      }),
    );
    await handle(
      message('channel.channel_points_custom_reward_redemption.add', {
        user_id: 'user',
        user_name: 'Viewer',
        reward: { title: 'Drink a Potion' },
      }),
    );
    await handle(message('channel.raid', { from_broadcaster_user_name: 'Raider' }));

    expect(bot.handleChatMessage).toHaveBeenCalledWith({
      chatterUserID: 'user',
      chatterUsername: 'Viewer',
      messageID: 'chat-message',
      text: '!stats',
    });
    expect(bot.handleRedemption).toHaveBeenCalledWith({
      userID: 'user',
      username: 'Viewer',
      rewardTitle: 'Drink a Potion',
    });
    expect(bot.handleRaid).toHaveBeenCalledWith({ fromBroadcasterUsername: 'Raider' });
  });

  it('ignores unsupported subscription types and rejects malformed supported events', async () => {
    const bot = createBotFake();
    const handle = createBotNotificationHandler(bot);

    await expect(handle(message('stream.online', {}))).resolves.toBeUndefined();
    await expect(handle(message('channel.chat.message', {}))).rejects.toThrow(
      'did not contain message details',
    );
    expect(bot.handleChatMessage).not.toHaveBeenCalled();
  });
});
