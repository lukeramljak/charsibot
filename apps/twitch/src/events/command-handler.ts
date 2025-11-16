import type { Bot } from '@/bot/bot';
import type { Command } from '@/commands/command';
import type { EventSubChannelChatMessageEvent } from '@twurple/eventsub-base';
import { log } from '@/logger';

export class CommandHandler {
  constructor(private commands: Command[]) {}

  public async process(bot: Bot, event: EventSubChannelChatMessageEvent): Promise<void> {
    if (!event.messageText.startsWith('!')) return;

    const [cmd] = event.messageText.toLowerCase().replace('!', '').split(' ');
    log.info(
      {
        command: cmd,
        user: event.chatterDisplayName,
        message: event.messageText,
      },
      'chat command received',
    );

    const commands = this.commands.filter((command) => {
      if (!command.shouldTrigger(cmd)) {
        return false;
      }

      return true;
    });

    if (commands.length === 0) {
      return;
    }

    for (const command of commands) {
      const isMod = event.badges.moderator || event.badges.broadcaster;
      if (command.moderatorOnly && !isMod) {
        log.warn(
          { user: event.chatterDisplayName, command: event.messageText },
          'non-moderator attempted to use mod command',
        );

        await bot.sendMessage('You must be a moderator to use this command');
        continue;
      }

      await command.execute(bot, event);
      log.info(
        { command: cmd, userId: event.chatterId, username: event.chatterDisplayName },
        'chat command executed',
      );
    }
  }
}
