import { Message } from 'discord.js';

export interface Trigger {
  triggerChance?: number;
  shouldTrigger(msg: Message): boolean;
  execute(msg: Message): Promise<void>;
}
