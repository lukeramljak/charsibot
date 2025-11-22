import { redeemBlindBox } from '@/blind-box/redeem-blind-box';
import { type BlindBoxConfig } from '@/blind-box/types';
import type { Bot } from '@/bot/bot';
import type { Redemption } from '@/redemptions/redemption';
import type { EventSubChannelRedemptionAddEvent } from '@twurple/eventsub-base';

export class BlindBoxRedemption implements Redemption {
  constructor(private config: BlindBoxConfig) {}

  shouldTrigger(event: EventSubChannelRedemptionAddEvent): boolean {
    return event.rewardTitle === this.config.rewardTitle;
  }

  async execute(bot: Bot, event: EventSubChannelRedemptionAddEvent): Promise<void> {
    const userId = event.userId;
    const username = event.userDisplayName;

    await redeemBlindBox(bot, {
      config: this.config,
      userId,
      username,
    });
  }
}
