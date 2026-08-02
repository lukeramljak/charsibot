import type { Bot } from '$lib/server/bot';
import { EventSubProtocolError } from '$lib/server/twitch/errors';
import type { EventSubNotificationMessage } from '$lib/server/twitch/types';

const isRecord = (value: unknown): value is Record<string, unknown> => {
  return typeof value === 'object' && value !== null;
};

const requiredString = (
  record: Record<string, unknown>,
  key: string,
  eventType: string,
): string => {
  const value = record[key];

  if (typeof value !== 'string') {
    throw new EventSubProtocolError(`Twitch ${eventType} event did not contain ${key}`);
  }

  return value;
};

export const createBotNotificationHandler = (bot: Bot) => {
  return async (message: EventSubNotificationMessage): Promise<void> => {
    const type = message.payload.subscription.type;
    const event = message.payload.event;

    if (!isRecord(event)) {
      throw new EventSubProtocolError(`Twitch ${type} notification did not contain an event`);
    }

    if (type === 'channel.chat.message') {
      if (!isRecord(event.message)) {
        throw new EventSubProtocolError('Twitch chat event did not contain message details');
      }

      await bot.handleChatMessage({
        chatterUserID: requiredString(event, 'chatter_user_id', type),
        chatterUsername: requiredString(event, 'chatter_user_name', type),
        messageID: requiredString(event, 'message_id', type),
        text: requiredString(event.message, 'text', type),
      });

      return;
    }

    if (type === 'channel.channel_points_custom_reward_redemption.add') {
      if (!isRecord(event.reward)) {
        throw new EventSubProtocolError('Twitch redemption event did not contain reward details');
      }

      await bot.handleRedemption({
        userID: requiredString(event, 'user_id', type),
        username: requiredString(event, 'user_name', type),
        rewardTitle: requiredString(event.reward, 'title', type),
      });

      return;
    }

    if (type === 'channel.raid') {
      await bot.handleRaid({
        fromBroadcasterUsername: requiredString(event, 'from_broadcaster_user_name', type),
      });
    }
  };
};
