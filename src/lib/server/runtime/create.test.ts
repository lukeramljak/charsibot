import { describe, expect, it } from 'vitest';

import { readRuntimeConfig } from './create';

describe('readRuntimeConfig', () => {
  it('throws when DB_PATH is missing', () => {
    expect(() => readRuntimeConfig({})).toThrow('DB_PATH environment variable is required');
  });

  it('reads DB_PATH', () => {
    const config = readRuntimeConfig({ DB_PATH: '/data/charsibot.db', TWITCH_MOCK_MODE: 'true' });

    expect(config.dbPath).toBe('/data/charsibot.db');
    expect(config.mockTwitch).toBe(true);
  });

  it('permits missing Twitch configuration only in explicit mock mode', () => {
    const config = readRuntimeConfig({ DB_PATH: '/data/charsibot.db', TWITCH_MOCK_MODE: 'true' });

    expect(config.twitch).toBeUndefined();
  });

  it('fails when Twitch configuration is missing', () => {
    expect(() => readRuntimeConfig({ DB_PATH: '/data/charsibot.db' })).toThrow(
      'TWITCH_CLIENT_ID, TWITCH_CLIENT_SECRET, TWITCH_BOT_USER_ID, and TWITCH_CHANNEL_USER_ID are required',
    );
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

  it('treats empty strings as missing and fails closed', () => {
    expect(() =>
      readRuntimeConfig({
      DB_PATH: '/data/charsibot.db',
      TWITCH_CLIENT_ID: '',
      TWITCH_CLIENT_SECRET: 'secret',
      TWITCH_BOT_USER_ID: 'bot',
      TWITCH_CHANNEL_USER_ID: 'channel',
      }),
    ).toThrow('TWITCH_CLIENT_ID, TWITCH_CLIENT_SECRET, TWITCH_BOT_USER_ID, and TWITCH_CHANNEL_USER_ID are required');
  });
});
