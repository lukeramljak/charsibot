import { Bot } from '@/bot';
import { CustomClient } from '@/client';
import type { Command } from '@/commands/command';
import { CommandRegistrationService } from '@/commands/command-registration-service';
import { ViewDateJoinedCommand } from '@/commands/user/view-date-joined';
import { MessageHandler } from '@/events/message-handler';
import { TriggerHandler } from '@/events/trigger-handler';
import { ButtTrigger } from '@/triggers/butt';
import { ComeTrigger } from '@/triggers/come';
import { CowTrigger } from '@/triggers/cow';
import { DogTrigger } from '@/triggers/dog';
import { EggTrigger } from '@/triggers/egg';
import { PingTrigger } from '@/triggers/ping';
import type { Trigger } from '@/triggers/trigger';
import { UserJoinTrigger } from '@/triggers/user-join';
import { GatewayIntentBits, REST } from 'discord.js';

async function main() {
  const clientId = process.env.DISCORD_CLIENT_ID!;
  const guildId = process.env.DISCORD_GUILD_ID!;
  const token = process.env.DISCORD_TOKEN!;

  const client = new CustomClient({
    intents: [
      GatewayIntentBits.Guilds,
      GatewayIntentBits.GuildMessages,
      GatewayIntentBits.GuildMembers,
      GatewayIntentBits.GuildMessageReactions,
      GatewayIntentBits.MessageContent,
    ],
  });

  const commands: Command[] = [new ViewDateJoinedCommand()];
  const triggers: Trigger[] = [
    new ButtTrigger(),
    new ComeTrigger(),
    new CowTrigger(),
    new DogTrigger(),
    new EggTrigger(),
    new PingTrigger(),
    new UserJoinTrigger(),
  ];

  const triggerHandler = new TriggerHandler(triggers);
  const messageHandler = new MessageHandler(triggerHandler);

  const bot = new Bot(token, client, commands, messageHandler);

  if (process.argv[2] == 'commands') {
    try {
      const rest = new REST({ version: '10' }).setToken(token);
      const commandRegistrationService = new CommandRegistrationService(rest);
      await commandRegistrationService.process(clientId, guildId, commands, process.argv);
    } catch (error) {
      console.error('Failed to register commands with Discord API:', error);
    }
    // Wait for any final logs to be written.
    await new Promise((resolve) => setTimeout(resolve, 1000));
    process.exit();
  }

  await bot.start();
}

main().catch((error) => {
  console.error('Fatal error', error);
});
