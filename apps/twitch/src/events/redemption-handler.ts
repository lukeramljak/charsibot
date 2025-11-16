import type { Bot } from '@/bot/bot';
import { log } from '@/logger';
import type { Redemption } from '@/redemptions/redemption';
import type { EventSubChannelRedemptionAddEvent } from '@twurple/eventsub-base';

export class RedemptionHandler {
  constructor(private redemptions: Redemption[]) {}

  public async process(bot: Bot, event: EventSubChannelRedemptionAddEvent): Promise<void> {
    const redemptions = this.redemptions.filter((redemption) => {
      if (!redemption.shouldTrigger(event)) {
        return false;
      }

      return true;
    });

    if (redemptions.length === 0) {
      return;
    }

    for (const redemption of redemptions) {
      await redemption.execute(bot, event);
      log.info({ userId: event.userId, username: event.userDisplayName }, 'redemption executed');
    }
  }
}
