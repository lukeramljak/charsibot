import { ChatInputCommandInteraction, SlashCommandBuilder } from "discord.js";
import type { Command } from "../command";

export class ViewDateJoinedCommand implements Command {
  data = new SlashCommandBuilder()
    .setName("datejoined")
    .setDescription("View the date you joined the server");

  async execute(interaction: ChatInputCommandInteraction): Promise<void> {
    const member = await interaction.guild?.members.fetch(interaction.user.id);
    if (!member) {
      await interaction.reply("Could not fetch your information.");
      return;
    }

    const joinDate = member.joinedAt;
    if (!joinDate) {
      await interaction.reply("Could not determine your join date.");
      return;
    }

    await interaction.reply(
      `You joined the server on: ${joinDate.toLocaleDateString("en-US", {
        weekday: "long",
        year: "numeric",
        month: "long",
        day: "numeric",
      })}`
    );
  }
}
