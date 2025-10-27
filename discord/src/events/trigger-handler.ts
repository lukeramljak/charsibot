import { RateLimiter } from "discord.js-rate-limiter";
import { Message } from "discord.js";
import { Trigger } from "../triggers/trigger";

export class TriggerHandler {
  private rateLimiter = new RateLimiter(1, 5000);

  constructor(private triggers: Trigger[]) {}

  public async process(message: Message): Promise<void> {
    // Check if user is rate limited
    const limited = this.rateLimiter.take(message.author.id);
    if (limited) {
      return;
    }

    // Find triggers caused by this message
    const triggers = this.triggers.filter((trigger) => {
      if (!trigger.shouldTrigger(message)) {
        return false;
      }

      return true;
    });

    // If this message causes no triggers then return
    if (triggers.length === 0) {
      return;
    }

    // Execute triggers
    for (const trigger of triggers) {
      if (trigger.triggerChance) {
        const roll = Math.random() * 100;
        if (roll > trigger.triggerChance) {
          continue;
        }
      }
      await trigger.execute(message);
    }
  }
}
