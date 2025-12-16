import type { Bot } from '@/bot/bot';
import type { Trigger } from '@/triggers/trigger';
import type { EventSubChannelChatMessageEvent } from '@twurple/eventsub-base';

export class ComeTrigger implements Trigger {
  triggerChance = 20;

  shouldTrigger(event: EventSubChannelChatMessageEvent): boolean {
    const triggers = ['come', 'coming', 'cum', 'came'];
    const words = event.messageText.toLowerCase().split(/\W+/).filter(Boolean)
    return words.some((word) => triggers.includes(word))
  }

  async execute(bot: Bot, event: EventSubChannelChatMessageEvent): Promise<void> {
    await bot.sendMessage('no coming', event.messageId);
  }
}
