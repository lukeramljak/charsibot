import { redemptionToCollectionType } from '@/blind-box/blind-box';
import { redeemBlindBox } from '@/blind-box/redeem-blind-box';
import { type BlindBoxRedemptionTitle } from '@/blind-box/types';
import type { Bot } from '@/bot/bot';
import { log } from '@/logger';
import type { Redemption } from '@/redemptions/redemption';
import type { EventSubChannelRedemptionAddEvent } from '@twurple/eventsub-base';

export class BlindBoxRedemption implements Redemption {
  constructor(private name: BlindBoxRedemptionTitle) {}

  shouldTrigger(event: EventSubChannelRedemptionAddEvent) {
    return event.rewardTitle === this.name;
  }

  async execute(bot: Bot, event: EventSubChannelRedemptionAddEvent) {
    const userId = event.userId;
    const username = event.userDisplayName;

    const type = redemptionToCollectionType[event.rewardTitle as BlindBoxRedemptionTitle];
    if (!type) {
      log.error({ redemption: event.rewardTitle, type }, 'Invalid blind box type');
      return;
    }

    await redeemBlindBox(bot, {
      type,
      userId,
      username,
    });
  }
}
