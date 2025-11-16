import type { EventSubChannelChatMessageEvent } from '@twurple/eventsub-base';
import type { Bot } from '../../bot/bot';
import type { Command } from '../command';
import { commandToCollectionType } from '../../blind-box/blind-box';

export class ShowBlindBoxCollectionCommand implements Command {
  shouldTrigger(command: string) {
    const collections = ['coobubu', 'olliepop'];
    return collections.includes(command);
  }

  async execute(bot: Bot, event: EventSubChannelChatMessageEvent) {
    const [command] = event.messageText.toLowerCase().split(' ');
    const userId = event.chatterId;
    const username = event.chatterDisplayName;

    const type = commandToCollectionType[command];

    const collection = await bot.store.getUserCollections(userId, type);

    bot.wsServer.broadcast({
      type: 'collection_display',
      data: {
        userId,
        username,
        collectionType: type,
        collection: collection || [],
        collectionSize: collection?.length || 0
      }
    });
  }
}
