import { describe, expect, it } from 'vitest';

import { readServerConfig } from './config';

describe('readServerConfig', () => {
  it('uses the existing Charsibot port by default', () => {
    expect(readServerConfig({})).toEqual({
      host: '0.0.0.0',
      port: 8081,
      shutdownTimeoutMilliseconds: 30_000,
    });
  });

  it('prefers the adapter-node PORT variable', () => {
    const config = readServerConfig({ PORT: '4000', SERVER_PORT: '8081' });
    expect(config.port).toBe(4000);
  });

  it('supports the existing SERVER_PORT variable during migration', () => {
    expect(readServerConfig({ SERVER_PORT: '9000' }).port).toBe(9000);
  });

  it.each(['0', '-1', '1.5', 'not-a-number'])('rejects invalid ports: %s', (port) => {
    expect(() => readServerConfig({ PORT: port })).toThrow('PORT must be a positive integer');
  });
});
