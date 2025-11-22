import { redeemBlindBox } from '@/blind-box/redeem-blind-box';
import type { BlindBoxConfig } from '@/blind-box/types';
import type { Bot } from '@/bot/bot';
import type { Command } from '@/commands/command';
import type { EventSubChannelChatMessageEvent } from '@twurple/eventsub-base';

export class RedeemBlindBoxCommand implements Command {
  moderatorOnly = true;

  constructor(private config: BlindBoxConfig) {}

  shouldTrigger(command: string): boolean {
    return command === this.config.moderatorCommand;
  }

  async execute(bot: Bot, event: EventSubChannelChatMessageEvent): Promise<void> {
    const userId = event.chatterId;
    const username = event.chatterDisplayName;

    await redeemBlindBox(bot, {
      config: this.config,
      userId,
      username,
    });
  }
}
