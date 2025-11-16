import type { Bot } from '../bot/bot';
import type { Trigger } from '../triggers/trigger';
import type { EventSubChannelChatMessageEvent } from '@twurple/eventsub-base';

export class TriggerHandler {
  constructor(private triggers: Trigger[]) {}

  public async process(bot: Bot, event: EventSubChannelChatMessageEvent): Promise<void> {
    const triggers = this.triggers.filter(trigger => {
      if (!trigger.shouldTrigger(event)) {
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
      await trigger.execute(bot);
    }
  }
}
