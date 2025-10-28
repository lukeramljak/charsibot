import { Message, type TextBasedChannel } from "discord.js";
import type { Trigger } from "./trigger";

export class ComeTrigger implements Trigger {
  triggerChance = 20;

  shouldTrigger(message: Message): boolean {
    const triggers = ["come", "coming", "cum", "came"];
    return triggers.some((trigger) =>
      message.content.toLowerCase().includes(trigger)
    );
  }

  async execute(message: Message): Promise<void> {
    await message.reply("no coming");
  }
}
