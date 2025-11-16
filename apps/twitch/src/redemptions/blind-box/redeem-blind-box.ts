import type { EventSubChannelRedemptionAddEvent } from '@twurple/eventsub-base';
import type { Redemption } from '../redemption';
import type { Bot } from '../../bot/bot';
import { log } from '../../logger';
import { redeemBlindBox } from '../../blind-box/redeem-blind-box';
import { redemptionToCollectionType } from '../../blind-box/blind-box';

export class BlindBoxRedemption implements Redemption {
  shouldTrigger(event: EventSubChannelRedemptionAddEvent) {
    const triggers = ['Cooper Series Blind Box', 'Ollie Series Blind Box'];
    return triggers.includes(event.rewardTitle);
  }

  async execute(bot: Bot, event: EventSubChannelRedemptionAddEvent) {
    const userId = event.userId;
    const username = event.userDisplayName;

    const type = redemptionToCollectionType[event.rewardTitle];
    if (!type) {
      log.error({ redemption: event.rewardTitle, type }, 'Invalid blind box type');
      return;
    }

    await redeemBlindBox(bot, {
      type,
      userId,
      username
    });
  }
}
