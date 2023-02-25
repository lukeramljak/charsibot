const { Client, Message, EmbedBuilder } = require('discord.js');

module.exports = {
	name: 'messageCreate',
	/**
	 * @param {Message} message
	 * @param {Client} client
	 */
	execute(message, client) {
		const embed = new EmbedBuilder().setColor('Blurple').setTimestamp();

		if (message.author.bot) return;
		if (message.mentions.has(client.user.id) && !message.author.bot)
			return message.reply({
				embeds: [
					embed
						.setAuthor({
							name: 'Hiya!',
							iconURL: client.user.avatarURL({ dynamic: true }),
						})
						.setDescription("Hiya, I'm charsibot!"),
				],
			});
	},
};
