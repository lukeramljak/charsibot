import type { EventHandler } from '@/events/event-handler';
import { TriggerHandler } from '@/events/trigger-handler';
import { Message } from 'discord.js';

export class MessageHandler implements EventHandler {
  constructor(private triggerHandler: TriggerHandler) {}

  public async process(message: Message): Promise<void> {
    // Ignore system messages and messages from ourselves
    if (message.system || message.author.id === message.client.user?.id) {
      return;
    }

    await this.triggerHandler.process(message);
  }
}
