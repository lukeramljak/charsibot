import type { EventSubChannelChatMessageEvent } from '@twurple/eventsub-base';
import type { Bot } from '@/bot/bot';

export interface Trigger {
  triggerChance?: number;
  shouldTrigger(event: EventSubChannelChatMessageEvent): boolean;
  execute(bot: Bot, event: EventSubChannelChatMessageEvent): Promise<void>;
}
