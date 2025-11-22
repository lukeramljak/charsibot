import type { Bot } from '@/bot/bot';
import type { Redemption } from '@/redemptions/redemption';
import { formatStats, getRandomStat, getRandomStatDelta } from '@/stats/stats';
import type { EventSubChannelRedemptionAddEvent } from '@twurple/eventsub-base';

export class PotionRedemption implements Redemption {
  shouldTrigger(event: EventSubChannelRedemptionAddEvent): boolean {
    return event.rewardTitle === 'Drink a Potion';
  }

  async execute(bot: Bot, event: EventSubChannelRedemptionAddEvent): Promise<void> {
    const stat = getRandomStat();
    const delta = getRandomStatDelta();
    const outcome = delta < 0 ? 'lost' : 'gained';

    const userId = event.userId;
    const username = event.userDisplayName;

    await bot.store.modifyStat(userId, username, stat.column, delta);

    const message = `A shifty looking merchant hands ${username} a glittering potion. Without hesitation, they sink the whole drink. ${username} ${outcome} ${stat.display}`;
    await bot.sendMessage(message);

    const stats = await bot.store.getStats(userId, username);
    await bot.sendMessage(formatStats(username, stats));
  }
}
