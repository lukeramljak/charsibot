import type { BlindBoxConfig } from '@/blind-box/types';
import type { Bot } from '@/bot/bot';
import type { Command } from '@/commands/command';
import type { EventSubChannelChatMessageEvent } from '@twurple/eventsub-base';

export class ResetCollectionCommand implements Command {
  moderatorOnly = true;

  constructor(private config: BlindBoxConfig) {}

  shouldTrigger(command: string): boolean {
    return command === this.config.resetCommand;
  }

  async execute(bot: Bot, event: EventSubChannelChatMessageEvent): Promise<void> {
    await bot.store.resetUserCollection(event.chatterId, this.config.collectionType);
  }
}
