import type { EventSubChannelChatMessageEvent } from '@twurple/eventsub-base';
import type { Command } from '@/commands/command';
import type { Bot } from '@/bot/bot';
import { log } from '@/logger';
import { formatStats, parseModifyStatCommand } from '@/stats/stats';

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

    await bot.store.modifyStat(
      event.chatterId,
      mentionedLogin,
      statColumn,
      isRemove ? -amount : amount,
    );

    const stats = await bot.store.getStats(event.chatterId, mentionedLogin);
    await bot.sendMessage(formatStats(mentionedLogin, stats));
  }
}
