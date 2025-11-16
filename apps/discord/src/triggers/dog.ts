import type { Trigger } from '@/triggers/trigger';
import { Message } from 'discord.js';

export class DogTrigger implements Trigger {
  triggerChance = 20;

  shouldTrigger(message: Message): boolean {
    return message.content.toLowerCase().includes('dog');
  }

  async execute(message: Message): Promise<void> {
    await message.reply("what the dog doin'?");
  }
}
