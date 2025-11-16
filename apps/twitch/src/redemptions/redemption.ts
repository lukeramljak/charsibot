import type { Bot } from '@/bot/bot';
import type { EventSubChannelRedemptionAddEvent } from '@twurple/eventsub-base';

export interface Redemption {
  shouldTrigger(event: EventSubChannelRedemptionAddEvent): boolean;
  execute(bot: Bot, event: EventSubChannelRedemptionAddEvent): Promise<void>;
}
