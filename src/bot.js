const dotenv = require('dotenv');
dotenv.config();
const fs = require('fs');
const path = require('path');
const { Client, Events, GatewayIntentBits, Collection } = require('discord.js');

const client = new Client({
	intents: [
		GatewayIntentBits.Guilds,
		GatewayIntentBits.GuildMessages,
		GatewayIntentBits.GuildMembers,
		GatewayIntentBits.DirectMessages,
		GatewayIntentBits.MessageContent,
	],
});

client.login(process.env.token);

client.on(Events.ClientReady, (c) => {
	console.log(`${c.user.username} is ready!`);
});

client.on('message', (message) => {
	if (message.content.toLowerCase() === 'egg') {
		message.channel.send('egg');
	}
});
