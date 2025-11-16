import type { Trigger } from '@/triggers/trigger';
import { Message } from 'discord.js';

export class CowTrigger implements Trigger {
  shouldTrigger(message: Message): boolean {
    return message.content.toLowerCase().includes('cow');
  }

  async execute(message: Message): Promise<void> {
    await message.reply(`MOOOOO! <:rage:1302882593339084851>`);
  }
}
