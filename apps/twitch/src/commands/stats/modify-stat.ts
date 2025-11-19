import type { Bot } from '@/bot/bot';
import type { Command } from '@/commands/command';
import { log } from '@/logger';
import { formatStats, parseModifyStatCommand } from '@/stats/stats';
import type { EventSubChannelChatMessageEvent } from '@twurple/eventsub-base';

export class ModifyStatCommand implements Command {
  moderatorOnly = true;

  shouldTrigger(command: string): boolean {
    return command === 'addstat' || command === 'rmstat';
  }

  async execute(bot: Bot, event: EventSubChannelChatMessageEvent): Promise<void> {
    const [command] = event.messageText.toLowerCase().split(' ');
    const isRemove = command === '!rmstat';

    const parsed = parseModifyStatCommand(event.messageText, isRemove);
    if (parsed.error) {
      log.warn(
        { command, error: parsed.error, msg: event.messageText },
        'invalid modify stat command',
      );
      await bot.sendMessage(parsed.error);
      return;
    }

    const { mentionedLogin, statColumn, amount } = parsed;

    const mentionedUser = await bot.api.users.getUserByName(mentionedLogin);
    if (!mentionedUser) {
      log.warn({ mentionedLogin }, 'failed to find user via helix api');
      await bot.sendMessage('Failed to find user');
      return;
    }

    await bot.store.modifyStat(
      mentionedUser.id,
      mentionedLogin,
      statColumn,
      isRemove ? -amount : amount,
    );

    const stats = await bot.store.getStats(mentionedUser.id, mentionedLogin);
    await bot.sendMessage(formatStats(mentionedLogin, stats));
  }
}
