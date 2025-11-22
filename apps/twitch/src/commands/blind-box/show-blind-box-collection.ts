import type { BlindBoxConfig } from '@/blind-box/types';
import type { Bot } from '@/bot/bot';
import type { Command } from '@/commands/command';
import { log } from '@/logger';
import type { EventSubChannelChatMessageEvent } from '@twurple/eventsub-base';

export class ShowBlindBoxCollectionCommand implements Command {
  constructor(private config: BlindBoxConfig) {}

  shouldTrigger(command: string): boolean {
    return command === this.config.collectionType;
  }

  async execute(bot: Bot, event: EventSubChannelChatMessageEvent): Promise<void> {
    const userId = event.chatterId;
    const username = event.chatterDisplayName;
    const collection = await bot.store.getUserCollections(userId, this.config.collectionType);

    if (!collection) {
      log.error({ collection, userId, username }, 'failed to get collection');
      await bot.sendMessage(`Failed to get ${username}'s collection'`);
      return;
    }

    bot.wsServer.broadcast({
      type: 'collection_display',
      data: {
        userId,
        username,
        collectionType: this.config.collectionType,
        collection: collection,
        collectionSize: collection.length,
      },
    });
  }
}
