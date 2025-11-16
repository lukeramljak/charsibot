import type { Bot } from '@/bot/bot';
import type { EventHandler } from '@/events/event-handler';
import { TriggerHandler } from '@/events/trigger-handler';
import type { EventSubChannelChatMessageEvent } from '@twurple/eventsub-base';

export class MessageHandler implements EventHandler {
  constructor(private triggerHandler: TriggerHandler) {}

  public async process(bot: Bot, event: EventSubChannelChatMessageEvent): Promise<void> {
    // Ignore system messages and messages from ourselves
    if (event.chatterId === bot.getId()) {
      return;
    }

    await this.triggerHandler.process(bot, event);
  }
}
