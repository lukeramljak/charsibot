const { SlashCommandBuilder } = require("discord.js");

module.exports = {
  data: new SlashCommandBuilder()
    .setName("burrito")
    .setDescription("tuck someone into bed")
    .addUserOption((option) => option.setName("target").setDescription("who are you tucking in?")),

  async execute(interaction) {
    const member = interaction.options.getMember("target");
    return interaction.reply({
      content: `You have tucked ${member.user.username} into a burrito blanket. Awww goodnight ${member.user.username}`,
    });
  },
};
