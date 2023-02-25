const { SlashCommandBuilder } = require('discord.js');

module.exports = {
	data: new SlashCommandBuilder()
		.setName('burrito')
		.setDescription('tuck someone into bed!')
		.addUserOption((option) =>
			option.setName('name').setDescription("who's getting tucked in?"),
		),
	async execute(interaction) {
		const member = interaction.options.getMember('name');
		return interaction.reply({
			content: `${interaction.user.username} has tucked ${member.user.username} into a burrito blanket. Awww goodnight ${member.user.username}`,
		});
	},
};
