import type { EventSubChannelChatMessageEvent } from '@twurple/eventsub-base';
import type { Bot } from '@/bot/bot';
import type { Command } from '@/commands/command';
import { redeemBlindBox } from '@/blind-box/redeem-blind-box';
import type { CollectionType } from '@/blind-box/types';

export class RedeemBlindBoxCommand implements Command {
  moderatorOnly = true;

  constructor(private type: CollectionType) {}

  shouldTrigger(command: string): boolean {
    return command === `${this.type}-redeem`;
  }

  async execute(bot: Bot, event: EventSubChannelChatMessageEvent): Promise<void> {
    const userId = event.chatterId;
    const username = event.chatterDisplayName;

    await redeemBlindBox(bot, {
      type: this.type,
      userId,
      username,
    });
  }
}
