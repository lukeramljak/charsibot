import type { Bot } from '@/bot/bot';
import type { Redemption } from '@/redemptions/redemption';
import { formatStats } from '@/stats/stats';
import type { EventSubChannelRedemptionAddEvent } from '@twurple/eventsub-base';

export class TemptDiceRedemption implements Redemption {
  shouldTrigger(event: EventSubChannelRedemptionAddEvent) {
    return event.rewardTitle === 'Tempt the Dice';
  }

  async execute(bot: Bot, event: EventSubChannelRedemptionAddEvent) {
    const userId = event.userId;
    const username = event.userDisplayName;

    await bot.sendMessage(`${username} has rolled with initiative.`);

    const stats = await bot.store.getStats(userId, username);
    await bot.sendMessage(formatStats(username, stats));
  }
}
