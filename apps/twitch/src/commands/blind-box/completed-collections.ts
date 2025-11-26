import type { Bot } from '@/bot/bot';
import type { Command } from '@/commands/command';

export class CompletedCollectionsCommand implements Command {
  shouldTrigger(command: string): boolean {
    return command === 'collections';
  }

  async execute(bot: Bot): Promise<void> {
    const collections = await bot.store.getCompletedCollections();

    await bot.sendMessage('The following chatters have completed the below blind box collections:');

    for (const collection of collections) {
      const collectionType = collection.collectionType
        ? collection.collectionType.trim()[0].toUpperCase() +
          collection.collectionType.trim().slice(1)
        : '';
      bot.sendMessage(`${collectionType}: ${collection.usernames.join(', ')}`);
    }
  }
}
