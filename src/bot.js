const { Client, Events, GatewayIntentBits } = require('discord.js');
const dotenv = require('dotenv');
const client = new Client({ intents: [GatewayIntentBits.Guilds] });

dotenv.config();

client.once(Events.ClientReady, (c) => {
	console.log(`${c.user.tag} is ready!`);
});

client.login(process.env.token);
