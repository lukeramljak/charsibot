export interface Config {
  clientId: string;
  clientSecret: string;
  streamerAccessToken: string;
  streamerRefreshToken: string;
  botAccessToken: string;
  botRefreshToken: string;
  botUserId: string;
  botLogin?: string;
  channelUserId: string;
  dbUrl: string;
  dbAuthToken: string;
  useMockServer: boolean;
}

const REQUIRED_VARS = [
  "TWITCH_CLIENT_ID",
  "TWITCH_CLIENT_SECRET",
  "TWITCH_OAUTH_TOKEN",
  "TWITCH_REFRESH_TOKEN",
  "TWITCH_BOT_OAUTH_TOKEN",
  "TWITCH_BOT_REFRESH_TOKEN",
  "TWITCH_BOT_USER_ID",
  "TWITCH_CHANNEL_USER_ID",
  "TURSO_DATABASE_URL",
  "TURSO_AUTH_TOKEN",
];

export const loadConfig = (): Config => {
  for (const key of REQUIRED_VARS) {
    if (!process.env[key]) {
      console.error(`Missing required env var ${key}`);
      process.exit(1);
    }
  }
  return {
    clientId: process.env.TWITCH_CLIENT_ID!,
    clientSecret: process.env.TWITCH_CLIENT_SECRET!,
    streamerAccessToken: process.env.TWITCH_OAUTH_TOKEN!,
    streamerRefreshToken: process.env.TWITCH_REFRESH_TOKEN!,
    botAccessToken: process.env.TWITCH_BOT_OAUTH_TOKEN!,
    botRefreshToken: process.env.TWITCH_BOT_REFRESH_TOKEN!,
    botUserId: process.env.TWITCH_BOT_USER_ID!,
    channelUserId: process.env.TWITCH_CHANNEL_USER_ID!,
    dbUrl: process.env.TURSO_DATABASE_URL!,
    dbAuthToken: process.env.TURSO_AUTH_TOKEN!,
    useMockServer: process.env.USE_MOCK_SERVER === "true",
  };
};
