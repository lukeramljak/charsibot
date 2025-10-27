import {
  ActivityType,
  Events,
  GuildMember,
  type Interaction,
  Message,
  MessageFlags,
  type PartialGuildMember,
  TextChannel,
} from "discord.js";
import { CustomClient } from "./client";
import type { Command } from "./commands/command";
import { MessageHandler } from "./events/message-handler";

export class Bot {
  private ready = false;

  constructor(
    private token: string,
    private client: CustomClient,
    private commands: Command[],
    private messageHandler: MessageHandler
  ) {}

  public async start(): Promise<void> {
    this.registerListeners();
    this.registerCommands();
    await this.login(this.token);
  }

  private registerListeners(): void {
    this.client.on(Events.ClientReady, () => this.onReady());
    this.client.on(
      Events.GuildMemberRemove,
      (member: GuildMember | PartialGuildMember) =>
        this.onGuildMemberRemove(member)
    );
    this.client.on(Events.MessageCreate, (message: Message) =>
      this.onMessage(message)
    );
    this.client.on(Events.InteractionCreate, (interaction: Interaction) =>
      this.onInteraction(interaction)
    );
  }

  private registerCommands(): void {
    for (const command of this.commands) {
      this.client.commands.set(command.data.name, command);
    }
  }

  private async login(token: string): Promise<void> {
    try {
      await this.client.login(token);
    } catch (error) {
      console.error("Failed to login:", error);
    }
  }

  private async onReady(): Promise<void> {
    const userTag = this.client.user?.tag;
    this.ready = true;
    this.client.user?.setActivity({
      name: "Big Chungus",
      type: ActivityType.Listening,
    });
    console.log(`Ready! Logged in as ${userTag}`);
  }

  private async onMessage(message: Message): Promise<void> {
    if (!this.ready) {
      return;
    }

    try {
      await this.messageHandler.process(message);
    } catch (error) {
      console.error("Error processing message:", error);
    }
  }

  private async onInteraction(interaction: Interaction): Promise<void> {
    if (!this.ready) {
      return;
    }

    if (!interaction.isChatInputCommand()) {
      return;
    }

    const command = this.client.commands.get(interaction.commandName);
    if (!command) {
      console.error(`No matching ${interaction.commandName} was found`);
      return;
    }

    try {
      await command.execute(interaction);
    } catch (error) {
      console.error(error);
      if (interaction.replied || interaction.deferred) {
        await interaction.followUp({
          content: "There was an error while executing this command",
          flags: MessageFlags.Ephemeral,
        });
      } else {
        await interaction.reply({
          content: "There was an error while executing this command",
          flags: MessageFlags.Ephemeral,
        });
      }
    }
  }

  private async onGuildMemberRemove(
    member: GuildMember | PartialGuildMember
  ): Promise<void> {
    const channel = this.client.channels.cache.get(
      "1018070065423335437"
    ) as TextChannel;
    if (channel?.isTextBased()) {
      await channel.send(
        `${member.user.tag} has left the server. <:periodt:1302882591552307240>`
      );
    }
  }
}
