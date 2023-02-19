const { SlashCommandBuilder } = require("discord.js");

module.exports = {
  data: new SlashCommandBuilder().setName("dog").setDescription("what is he doing?"),
  async execute(interaction) {
    await interaction.reply("What the dog doin'?");
  },
};
