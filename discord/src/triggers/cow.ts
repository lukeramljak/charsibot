import { Message } from "discord.js";
import { Trigger } from "./trigger";

export class CowTrigger implements Trigger {
  triggerChance = 20;

  shouldTrigger(message: Message): boolean {
    return message.content.toLowerCase().includes("cow");
  }

  async execute(message: Message): Promise<void> {
    await message.reply(`MOOOOO! <:rage:1302882593339084851>`);
  }
}
