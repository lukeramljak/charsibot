import { describe, expect, it, vi } from 'vitest';

import { TwitchHttpError } from '$lib/server/twitch/errors';
import { createAppTokenProvider } from '$lib/server/twitch/token';

describe('createAppTokenProvider', () => {
  it('caches a token until its expiry window and encodes client credentials', async () => {
    let now = 1_000;
    const request = vi
      .fn<typeof fetch>()
      .mockResolvedValueOnce(
        Response.json({ access_token: 'first', expires_in: 120, token_type: 'bearer' }),
      )
      .mockResolvedValueOnce(Response.json({ client_id: 'client id', expires_in: 120 }))
      .mockResolvedValueOnce(
        Response.json({ access_token: 'second', expires_in: 120, token_type: 'bearer' }),
      )
      .mockResolvedValueOnce(Response.json({ client_id: 'client id', expires_in: 120 }));
    const provider = createAppTokenProvider({
      clientId: 'client id',
      clientSecret: 'secret&value',
      fetch: request,
      now: () => now,
    });

    await expect(provider.get()).resolves.toBe('first');
    await expect(provider.get()).resolves.toBe('first');
    expect(request).toHaveBeenCalledTimes(2);

    const requestBody = request.mock.calls[0][1]?.body as URLSearchParams;
    expect(String(request.mock.calls[0][0])).toBe('https://id.twitch.tv/oauth2/token');
    expect(new Headers(request.mock.calls[0][1]?.headers).get('Content-Type')).toBe(
      'application/x-www-form-urlencoded',
    );
    expect(requestBody.get('client_id')).toBe('client id');
    expect(requestBody.get('client_secret')).toBe('secret&value');
    expect(requestBody.get('grant_type')).toBe('client_credentials');

    now += 60_001;
    await expect(provider.get()).resolves.toBe('second');
    expect(request).toHaveBeenCalledTimes(4);
  });

  it('coalesces concurrent refreshes and supports explicit invalidation', async () => {
    const request = vi
      .fn<typeof fetch>()
      .mockResolvedValueOnce(
        Response.json({ access_token: 'first', expires_in: 3_600, token_type: 'bearer' }),
      )
      .mockResolvedValueOnce(Response.json({ client_id: 'client', expires_in: 3_600 }))
      .mockResolvedValueOnce(
        Response.json({ access_token: 'second', expires_in: 3_600, token_type: 'bearer' }),
      )
      .mockResolvedValueOnce(Response.json({ client_id: 'client', expires_in: 3_600 }));
    const provider = createAppTokenProvider({
      clientId: 'client',
      clientSecret: 'secret',
      fetch: request,
    });

    await expect(Promise.all([provider.get(), provider.get()])).resolves.toEqual([
      'first',
      'first',
    ]);
    expect(request).toHaveBeenCalledTimes(2);

    provider.invalidate();
    await expect(provider.get()).resolves.toBe('second');
  });

  it('surfaces non-success responses and permits a later retry', async () => {
    const request = vi
      .fn<typeof fetch>()
      .mockResolvedValueOnce(new Response('nope', { status: 500 }))
      .mockResolvedValueOnce(
        Response.json({ access_token: 'recovered', expires_in: 3_600, token_type: 'bearer' }),
      )
      .mockResolvedValueOnce(Response.json({ client_id: 'client', expires_in: 3_600 }));
    const provider = createAppTokenProvider({
      clientId: 'client',
      clientSecret: 'secret',
      fetch: request,
    });

    await expect(provider.get()).rejects.toBeInstanceOf(TwitchHttpError);
    await expect(provider.get()).resolves.toBe('recovered');
  });

  it('validates hourly and reacquires a token when validation returns 401', async () => {
    let now = 0;
    const request = vi
      .fn<typeof fetch>()
      .mockResolvedValueOnce(
        Response.json({ access_token: 'first', expires_in: 7_200, token_type: 'bearer' }),
      )
      .mockResolvedValueOnce(Response.json({ client_id: 'client', expires_in: 7_200 }))
      .mockResolvedValueOnce(new Response('', { status: 401 }))
      .mockResolvedValueOnce(
        Response.json({ access_token: 'second', expires_in: 7_200, token_type: 'bearer' }),
      )
      .mockResolvedValueOnce(Response.json({ client_id: 'client', expires_in: 7_200 }));
    const provider = createAppTokenProvider({
      clientId: 'client',
      clientSecret: 'secret',
      fetch: request,
      now: () => now,
    });

    await expect(provider.get()).resolves.toBe('first');
    now += 3_600_001;
    await expect(provider.get()).resolves.toBe('second');
    expect(request).toHaveBeenCalledTimes(5);
  });

  it('maintains validation hourly until aborted', async () => {
    vi.useFakeTimers();
    let now = 0;
    const request = vi
      .fn<typeof fetch>()
      .mockResolvedValueOnce(
        Response.json({ access_token: 'token', expires_in: 10_800, token_type: 'bearer' }),
      )
      .mockResolvedValueOnce(Response.json({ client_id: 'client', expires_in: 10_800 }))
      .mockResolvedValueOnce(Response.json({ client_id: 'client', expires_in: 7_200 }));
    const provider = createAppTokenProvider({
      clientId: 'client',
      clientSecret: 'secret',
      fetch: request,
      now: () => now,
    });
    const controller = new AbortController();
    const maintenance = provider.maintain(controller.signal);
    await vi.advanceTimersByTimeAsync(0);
    expect(request).toHaveBeenCalledTimes(2);

    now = 3_600_000;
    await vi.advanceTimersByTimeAsync(3_600_000);
    expect(request).toHaveBeenCalledTimes(3);

    controller.abort();
    await expect(maintenance).resolves.toBeUndefined();
    vi.useRealTimers();
  });
});
