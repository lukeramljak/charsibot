import type { EventSubChannelChatMessageEvent } from '@twurple/eventsub-base';
import type { Command } from '../command';
import type { Bot } from '../../bot/bot';
import { formatStats } from '../../stats/stats';

export class StatsCommand implements Command {
  shouldTrigger(command: string) {
    return command === 'stats';
  }

  async execute(bot: Bot, event: EventSubChannelChatMessageEvent) {
    const id = event.chatterId;
    const username = event.chatterDisplayName;
    const stats = await bot.store.getStats(id, username);
    const message = formatStats(username, stats);

    await bot.sendMessage(message);
  }
}
