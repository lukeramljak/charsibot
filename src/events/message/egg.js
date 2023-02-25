const { Client, GatewayIntentBits } = require('discord.js');
const client = new Client({
	intents: GatewayIntentBits.MessageContent,
});

client.on('message', (message) => {
	if (message.content.toLowerCase() === 'egg') {
		message.channel.send('egg');
	}
});
