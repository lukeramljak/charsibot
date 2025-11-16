import type { Trigger } from '@/triggers/trigger';
import { Message } from 'discord.js';

export class ComeTrigger implements Trigger {
  triggerChance = 20;

  shouldTrigger(message: Message): boolean {
    const triggers = ['come', 'coming', 'cum', 'came'];
    return triggers.some((trigger) => message.content.toLowerCase().includes(trigger));
  }

  async execute(message: Message): Promise<void> {
    await message.reply('no coming');
  }
}
