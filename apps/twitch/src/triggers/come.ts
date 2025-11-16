import type { Bot } from '@/bot/bot';
import type { Trigger } from '@/triggers/trigger';
import type { EventSubChannelChatMessageEvent } from '@twurple/eventsub-base';

export class ComeTrigger implements Trigger {
  shouldTrigger(event: EventSubChannelChatMessageEvent): boolean {
    const triggers = ['come', 'coming', 'cum', 'came'];
    return triggers.some((trigger) => event.messageText.toLowerCase().includes(trigger));
  }

  async execute(bot: Bot, event: EventSubChannelChatMessageEvent): Promise<void> {
    await bot.sendMessage('no coming', event.messageId);
  }
}
