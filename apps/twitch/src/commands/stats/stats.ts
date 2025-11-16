import type { Bot } from '@/bot/bot';
import type { Command } from '@/commands/command';
import { formatStats } from '@/stats/stats';
import type { EventSubChannelChatMessageEvent } from '@twurple/eventsub-base';

export class StatsCommand implements Command {
  shouldTrigger(command: string): boolean {
    return command === 'stats';
  }

  async execute(bot: Bot, event: EventSubChannelChatMessageEvent): Promise<void> {
    const id = event.chatterId;
    const username = event.chatterDisplayName;
    const stats = await bot.store.getStats(id, username);
    const message = formatStats(username, stats);

    await bot.sendMessage(message, event.messageId);
  }
}
