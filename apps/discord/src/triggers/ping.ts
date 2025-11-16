import type { Trigger } from '@/triggers/trigger';
import { Message } from 'discord.js';

export class PingTrigger implements Trigger {
  shouldTrigger(message: Message): boolean {
    return message.content.toLowerCase().includes('ping');
  }

  async execute(message: Message): Promise<void> {
    await message.reply({ content: 'pong' });
  }
}
