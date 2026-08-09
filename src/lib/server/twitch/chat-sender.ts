import type { ChatSender } from '$lib/server/application/ports';
import type { TwitchChatClient } from '$lib/server/twitch/types';

export interface TwitchChatSenderOptions {
  client: TwitchChatClient;
  broadcasterId: string;
  senderId: string;
}

export const createTwitchChatSender = (options: TwitchChatSenderOptions): ChatSender => ({
  send: async (message, signal) => {
    await options.client.sendMessage(
      {
        broadcasterId: options.broadcasterId,
        senderId: options.senderId,
        message: message.message,
        replyParentMessageId: message.replyParentMessageID,
      },
      signal,
    );
  },
});
