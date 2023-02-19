const { SlashCommandBuilder } = require('discord.js');

module.exports = {
	data: new SlashCommandBuilder()
		.setName('dog')
		.setDescription('what IS he doing?'),
	async execute(interaction) {
		await interaction.reply("what the dog doin'?");
	},
};
