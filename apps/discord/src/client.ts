import { Client, type ClientOptions, Collection } from 'discord.js';
import type { Command } from '@/commands/command';

export class CustomClient extends Client {
  public commands: Collection<string, Command>;

  constructor(clientOptions: ClientOptions) {
    super(clientOptions);
    this.commands = new Collection<string, Command>();
  }
}
