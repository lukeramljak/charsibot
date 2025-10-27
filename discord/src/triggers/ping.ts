import { Message } from "discord.js";
import { Trigger } from "./trigger";

export class PingTrigger implements Trigger {
  shouldTrigger(message: Message): boolean {
    return message.content.toLowerCase().includes("ping");
  }

  async execute(message: Message): Promise<void> {
    await message.reply({ content: "pong" });
  }
}
