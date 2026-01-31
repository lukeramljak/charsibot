import type { Trigger } from '@/triggers/trigger';
import { Message } from 'discord.js';

export class ComeTrigger implements Trigger {
  triggerChance = 20;

  shouldTrigger(message: Message): boolean {
    const triggers = ['come', 'coming', 'cum', 'came'];
    const words = message.content.toLowerCase().split(/\W+/).filter(Boolean);
    return words.some((word) => triggers.includes(word));
  }

  async execute(message: Message): Promise<void> {
    await message.reply('no coming');
  }
}
