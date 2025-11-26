import type { Bot } from '@/bot/bot';
import type { Command } from '@/commands/command';

export class CompletedCollectionsCommand implements Command {
  shouldTrigger(command: string): boolean {
    return command === 'collections';
  }

  async execute(bot: Bot): Promise<void> {
    const collections = await bot.store.getCompletedCollections();

    for (const collection of collections) {
      await Promise.all([
        bot.sendMessage(`${collection.collectionType}: ${collection.usernames.join(', ')}`),
      ]);
    }
  }
}
