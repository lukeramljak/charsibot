import { describe, expect, it, vi } from 'vitest';

import { createConduitSessionManager } from '$lib/server/twitch/conduits';
import type { HelixClient, HelixRequestOptions } from '$lib/server/twitch/types';

interface StoredSubscription {
  id: string;
  status: string;
  type: string;
  version: string;
  condition: Record<string, string>;
  transport: { method: string; conduit_id: string };
}

type HelixRequestImplementation = (
  path: string,
  init?: RequestInit,
  options?: HelixRequestOptions,
) => Promise<unknown>;

const helixClient = (request: HelixRequestImplementation): HelixClient => {
  return { request: request as HelixClient['request'] };
};

describe('createConduitSessionManager', () => {
  it('reuses a one-shard conduit, enables shard zero and reconciles all subscriptions', async () => {
    const stored = new Map<string, StoredSubscription>();
    const request = vi.fn(async (path: string, init?: RequestInit) => {
      if (path === '/eventsub/conduits') {
        return {
          data: [
            { id: 'wrong-size', shard_count: 2 },
            { id: 'chosen', shard_count: 1 },
          ],
        };
      }

      if (path === '/eventsub/conduits/shards') {
        return { data: [{ id: '0', status: 'enabled' }], errors: [] };
      }

      if (path.startsWith('/eventsub/subscriptions?')) {
        const type = new URL(`https://mock.invalid${path}`).searchParams.get('type') as string;
        const existing = stored.get(type);
        return { data: existing ? [existing] : [] };
      }

      if (path === '/eventsub/subscriptions' && init?.body) {
        const body = JSON.parse(String(init.body)) as Omit<StoredSubscription, 'id' | 'status'>;
        stored.set(body.type, { ...body, id: `id-${body.type}`, status: 'enabled' });
        return undefined;
      }

      throw new Error(`Unexpected request: ${path}`);
    });
    const manager = createConduitSessionManager({
      helix: helixClient(request),
      clientId: 'client',
      botUserId: 'bot',
      channelUserId: 'channel',
    });

    await manager.prepareSession('session');

    const shardCall = request.mock.calls.find(([path]) => path === '/eventsub/conduits/shards');
    expect(JSON.parse(String(shardCall?.[1]?.body))).toEqual({
      conduit_id: 'chosen',
      shards: [{ id: '0', transport: { method: 'websocket', session_id: 'session' } }],
    });
    expect(stored).toHaveLength(4);
    expect(stored.get('conduit.shard.disabled')?.condition).toEqual({
      client_id: 'client',
      conduit_id: 'chosen',
    });
    expect([...stored.values()].every((entry) => entry.transport.conduit_id === 'chosen')).toBe(
      true,
    );
  });

  it('creates a one-shard conduit when only multi-shard conduits exist', async () => {
    const request = vi.fn(async (path: string, init?: RequestInit) => {
      if (path === '/eventsub/conduits' && init?.method === 'POST') {
        return { data: [{ id: 'new', shard_count: 1 }] };
      }

      if (path === '/eventsub/conduits') {
        return { data: [{ id: 'wrong-size', shard_count: 3 }] };
      }

      if (path === '/eventsub/conduits/shards') {
        return { data: [{ id: '0', status: 'enabled' }] };
      }

      if (path.startsWith('/eventsub/subscriptions?')) {
        const type = new URL(`https://mock.invalid${path}`).searchParams.get('type') as string;
        return {
          data: [
            {
              id: `id-${type}`,
              status: 'enabled',
              type,
              version: '1',
              condition:
                type === 'channel.chat.message'
                  ? { broadcaster_user_id: 'channel', user_id: 'bot' }
                  : type === 'channel.raid'
                    ? { to_broadcaster_user_id: 'channel' }
                    : type === 'conduit.shard.disabled'
                      ? { client_id: 'client', conduit_id: 'new' }
                      : { broadcaster_user_id: 'channel' },
              transport: { method: 'conduit', conduit_id: 'new' },
            },
          ],
        };
      }

      throw new Error(`Unexpected request: ${path}`);
    });
    const manager = createConduitSessionManager({
      helix: helixClient(request),
      clientId: 'client',
      botUserId: 'bot',
      channelUserId: 'channel',
    });

    await manager.prepareSession('session');

    expect(request).toHaveBeenCalledWith(
      '/eventsub/conduits',
      { method: 'POST', body: JSON.stringify({ shard_count: 1 }) },
      { signal: undefined },
    );
  });

  it('rejects an identity-matching subscription on a different conduit', async () => {
    const request = vi.fn(async (path: string) => {
      if (path === '/eventsub/conduits') {
        return { data: [{ id: 'chosen', shard_count: 1 }] };
      }

      if (path === '/eventsub/conduits/shards') {
        return { data: [{ id: '0', status: 'enabled' }] };
      }

      return {
        data: [
          {
            id: 'conflict',
            status: 'enabled',
            type: 'channel.chat.message',
            version: '1',
            condition: { broadcaster_user_id: 'channel', user_id: 'bot' },
            transport: { method: 'conduit', conduit_id: 'other' },
          },
        ],
      };
    });
    const manager = createConduitSessionManager({
      helix: helixClient(request),
      clientId: 'client',
      botUserId: 'bot',
      channelUserId: 'channel',
    });

    await expect(manager.prepareSession('session')).rejects.toThrow(
      'different transport or conduit',
    );
  });

  it('rejects an identity-matching subscription that is not enabled', async () => {
    const request = vi.fn(async (path: string) => {
      if (path === '/eventsub/conduits') {
        return { data: [{ id: 'chosen', shard_count: 1 }] };
      }

      if (path === '/eventsub/conduits/shards') {
        return { data: [{ id: '0', status: 'enabled' }] };
      }

      return {
        data: [
          {
            id: 'pending',
            status: 'websocket_disconnected',
            type: 'channel.chat.message',
            version: '1',
            condition: { broadcaster_user_id: 'channel', user_id: 'bot' },
            transport: { method: 'conduit', conduit_id: 'chosen' },
          },
        ],
      };
    });
    const manager = createConduitSessionManager({
      helix: helixClient(request),
      clientId: 'client',
      botUserId: 'bot',
      channelUserId: 'channel',
    });

    await expect(manager.prepareSession('session')).rejects.toThrow('is not enabled');
  });

  it('requires PATCH to report shard zero as enabled', async () => {
    const request = vi.fn(async (path: string) => {
      if (path === '/eventsub/conduits') {
        return { data: [{ id: 'chosen', shard_count: 1 }] };
      }

      return { data: [{ id: '0', status: 'websocket_disconnected' }] };
    });
    const manager = createConduitSessionManager({
      helix: helixClient(request),
      clientId: 'client',
      botUserId: 'bot',
      channelUserId: 'channel',
    });

    await expect(manager.prepareSession('session')).rejects.toThrow('did not enable shard 0');
  });
});
