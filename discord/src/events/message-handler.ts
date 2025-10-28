import { Message } from "discord.js";
import type { EventHandler } from "./event-handler";
import { TriggerHandler } from "./trigger-handler";

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
