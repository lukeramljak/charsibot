import type { Trigger } from '@/triggers/trigger';
import { Message } from 'discord.js';
import { RateLimiter } from 'discord.js-rate-limiter';

export class TriggerHandler {
  private rateLimiter = new RateLimiter(1, 5000);

  constructor(private triggers: Trigger[]) {}

  public async process(message: Message): Promise<void> {
    // Check if user is rate limited
    const limited = this.rateLimiter.take(message.author.id);
    if (limited) {
      return;
    }

    const triggers = this.triggers.filter((trigger) => {
      if (!trigger.shouldTrigger(message)) {
        return false;
      }

      return true;
    });

    if (triggers.length === 0) {
      return;
    }

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
