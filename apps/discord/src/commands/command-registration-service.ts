import { REST } from '@discordjs/rest';
import { type RESTPostAPIApplicationCommandsJSONBody, Routes } from 'discord.js';
import { type Command } from '../commands/command';

export class CommandRegistrationService {
  constructor(private rest: REST) {}

  public async process(
    clientId: string,
    guildId: string,
    commands: Command[],
    args: string[],
  ): Promise<void> {
    switch (args[3]) {
      case 'register': {
        if (commands.length > 0) {
          console.log(`Started refreshing ${commands.length} application (/) commands.`);

          const cmdsToRegister: RESTPostAPIApplicationCommandsJSONBody[] = [];

          for (const cmd of commands) {
            cmdsToRegister.push(cmd.data.toJSON());
          }

          try {
            await this.rest.put(Routes.applicationGuildCommands(clientId, guildId), {
              body: cmdsToRegister,
            });
            console.log(`Successfully registered ${commands.length} application (/) commands`);
          } catch (error) {
            console.error('Failed to register commands:', error);
          }

          return;
        }
      }
      case 'clear': {
        await this.rest.put(Routes.applicationGuildCommands(clientId, guildId), { body: [] });
        console.log(`Successfully cleared ${commands.length} (/) commands`);
        return;
      }
    }
  }
}
