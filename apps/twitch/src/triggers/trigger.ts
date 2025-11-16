import type { Bot } from '@/bot/bot';
import type { EventSubChannelChatMessageEvent } from '@twurple/eventsub-base';

export interface Trigger {
  triggerChance?: number;
  shouldTrigger(event: EventSubChannelChatMessageEvent): boolean;
  execute(bot: Bot, event: EventSubChannelChatMessageEvent): Promise<void>;
}
