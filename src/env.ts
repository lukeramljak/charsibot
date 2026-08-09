import { defineEnvVars } from '@sveltejs/kit/env';
import * as v from 'valibot';

export const variables = defineEnvVars({
  TWITCH_CLIENT_ID: {
    schema: v.string(),
  },
  TWITCH_CLIENT_SECRET: {
    schema: v.string(),
  },
  TWITCH_BOT_USER_ID: {
    schema: v.pipe(
      v.string(),
      v.transform(parseInt),
      v.number(),
    ),
  },
  TWITCH_CHANNEL_USER_ID: {
    schema: v.pipe(
      v.string(),
      v.transform(parseInt),
      v.number(),
    ),
  },
  TWITCH_OAUTH_REDIRECT_URI: {
    schema: v.pipe(v.string(), v.url()),
  },
  SERVER_PORT: {
    schema: v.pipe(
      v.string(),
      v.transform(parseInt),
      v.number(),
    ),
  },
  DB_PATH: {
    schema: v.string()
  },
});
