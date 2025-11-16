import type { EventSubChannelChatMessageEvent } from '@twurple/eventsub-base';
import type { Bot } from '../../bot/bot';
import type { Command } from '../command';
import type { CollectionType } from '../../blind-box/types';

export class ShowBlindBoxCollectionCommand implements Command {
  constructor(private type: CollectionType) {}

  shouldTrigger(command: string) {
    return command === this.type;
  }

  async execute(bot: Bot, event: EventSubChannelChatMessageEvent) {
    const userId = event.chatterId;
    const username = event.chatterDisplayName;
    const collection = await bot.store.getUserCollections(userId, this.type);

    bot.wsServer.broadcast({
      type: 'collection_display',
      data: {
        userId,
        username,
        collectionType: this.type,
        collection: collection || [],
        collectionSize: collection?.length || 0
      }
    });
  }
}
