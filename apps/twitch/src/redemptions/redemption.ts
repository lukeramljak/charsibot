import type { EventSubChannelRedemptionAddEvent } from '@twurple/eventsub-base';
import type { Bot } from '@/bot/bot';

export interface Redemption {
  shouldTrigger(event: EventSubChannelRedemptionAddEvent): boolean;
  execute(bot: Bot, event: EventSubChannelRedemptionAddEvent): Promise<void>;
}
