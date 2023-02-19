const { SlashCommandBuilder } = require("discord.js");

module.exports = {
  data: new SlashCommandBuilder().setName("egg").setDescription("egg"),
  async execute(interaction) {
    await interaction.reply("egg");
  },
};
