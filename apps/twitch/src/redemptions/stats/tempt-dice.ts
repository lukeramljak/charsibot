import type { EventSubChannelRedemptionAddEvent } from '@twurple/eventsub-base';
import type { Redemption } from '../redemption';
import type { Bot } from '../../bot/bot';
import { formatStats } from '../../stats/stats';

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
