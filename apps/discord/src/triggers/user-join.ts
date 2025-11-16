import { Message, MessageType } from 'discord.js';
import type { Trigger } from './trigger';

export class UserJoinTrigger implements Trigger {
  shouldTrigger(message: Message): boolean {
    return message.type === MessageType.UserJoin;
  }

  async execute(message: Message): Promise<void> {
    await Promise.all([
      message.react('a:catJAM:1111234741639848026'),
      message.react('a:hooray:1057490323561001042'),
      message.react('a:pedro:1057490323561001042'),
    ]);
  }
}
