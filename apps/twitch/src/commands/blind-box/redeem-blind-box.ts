import type { EventSubChannelChatMessageEvent } from '@twurple/eventsub-base';
import type { Bot } from '../../bot/bot';
import type { Command } from '../command';
import { log } from '../../logger';
import { redeemBlindBox } from '../../blind-box/redeem-blind-box';
import { commandToCollectionType } from '../../blind-box/blind-box';

export class RedeemBlindBoxCommand implements Command {
  moderatorOnly = true;

  shouldTrigger(command: string) {
    const triggers = ['coobubu-redeem', 'olliepop-redeem'];
    return triggers.includes(command);
  }

  async execute(bot: Bot, event: EventSubChannelChatMessageEvent) {
    const [command] = event.messageText.toLowerCase().split('-');
    const userId = event.chatterId;
    const username = event.chatterDisplayName;

    const type = commandToCollectionType[command];
    if (!type) {
      log.error({ command, type }, 'Invalid blind box type');
      return;
    }

    await redeemBlindBox(bot, {
      type,
      userId,
      username
    });
  }
}
