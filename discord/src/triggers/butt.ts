import type { Trigger } from '@/triggers/trigger';
import { Message } from 'discord.js';

export class ButtTrigger implements Trigger {
  triggerChance = 20;

  shouldTrigger(message: Message): boolean {
    return message.content.toLowerCase().includes('but');
  }

  async execute(message: Message): Promise<void> {
    await message.reply('butt');
  }
}
