import type { Trigger } from '@/triggers/trigger';
import { Message } from 'discord.js';

export class EggTrigger implements Trigger {
  shouldTrigger(message: Message): boolean {
    return message.content.toLowerCase().includes('egg');
  }

  async execute(message: Message): Promise<void> {
    await message.reply('egg');
  }
}
