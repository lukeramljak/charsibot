import { describe, expect, it } from 'vitest';

import { readRuntimeConfig } from './create';

describe('readRuntimeConfig', () => {
  it('throws when DB_PATH is missing', () => {
    expect(() => readRuntimeConfig({})).toThrow('DB_PATH environment variable is required');
  });

  it('reads DB_PATH', () => {
    const config = readRuntimeConfig({ DB_PATH: '/data/charsibot.db' });

    expect(config.dbPath).toBe('/data/charsibot.db');
  });

  it('returns undefined twitch config when env vars are missing', () => {
    const config = readRuntimeConfig({ DB_PATH: '/data/charsibot.db' });

    expect(config.twitch).toBeUndefined();
  });

  it('returns undefined twitch config when only some vars are set', () => {
    const config = readRuntimeConfig({
      DB_PATH: '/data/charsibot.db',
      TWITCH_CLIENT_ID: 'id',
      TWITCH_CLIENT_SECRET: 'secret',
    });

    expect(config.twitch).toBeUndefined();
  });

  it('reads twitch config when all vars are set', () => {
    const config = readRuntimeConfig({
      DB_PATH: '/data/charsibot.db',
      TWITCH_CLIENT_ID: 'my-client-id',
      TWITCH_CLIENT_SECRET: 'my-secret',
      TWITCH_BOT_USER_ID: 'bot-123',
      TWITCH_CHANNEL_USER_ID: 'channel-456',
    });

    expect(config.twitch).toEqual({
      clientId: 'my-client-id',
      clientSecret: 'my-secret',
      botUserId: 'bot-123',
      channelUserId: 'channel-456',
    });
  });

  it('treats empty strings as missing', () => {
    const config = readRuntimeConfig({
      DB_PATH: '/data/charsibot.db',
      TWITCH_CLIENT_ID: '',
      TWITCH_CLIENT_SECRET: 'secret',
      TWITCH_BOT_USER_ID: 'bot',
      TWITCH_CHANNEL_USER_ID: 'channel',
    });

    expect(config.twitch).toBeUndefined();
  });
});
