import { defineEnvVars } from '@sveltejs/kit/env';
import * as v from 'valibot';

export const variables = defineEnvVars({
  TWITCH_CLIENT_ID: {
    schema: v.optional(v.string()),
  },
  TWITCH_CLIENT_SECRET: {
    schema: v.optional(v.string()),
  },
  TWITCH_BOT_USER_ID: {
    schema: v.optional(v.string()),
  },
  TWITCH_CHANNEL_USER_ID: {
    schema: v.optional(v.string()),
  },
  TWITCH_OAUTH_REDIRECT_URI: {
    schema: v.optional(v.string()),
  },
  PORT: {
    schema: v.pipe(v.string(), v.transform(parseInt), v.number()),
  },
  DB_PATH: {
    schema: v.string(),
  },
});
