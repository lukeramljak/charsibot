import type { Bot } from '@/bot/bot';
import type { EventSubChannelChatMessageEvent } from '@twurple/eventsub-base';

export interface Command {
  moderatorOnly?: boolean;
  shouldTrigger(command: string): boolean;
  execute(bot: Bot, event: EventSubChannelChatMessageEvent): Promise<void>;
}
