import { describe, expect, it, vi } from 'vitest';

import { ApplicationLifecycle } from '$lib/server/runtime/lifecycle';

describe('ApplicationLifecycle', () => {
  it('starts once for concurrent callers', async () => {
    const stop = vi.fn(async () => undefined);
    const createRuntime = vi.fn(async () => ({ stop }));
    const lifecycle = new ApplicationLifecycle(createRuntime);

    const [first, second] = await Promise.all([lifecycle.start(), lifecycle.start()]);

    expect(first).toBe(second);
    expect(createRuntime).toHaveBeenCalledOnce();
    expect(lifecycle.state).toBe('running');
  });

  it('stops once for concurrent callers', async () => {
    const stop = vi.fn(async () => undefined);
    const lifecycle = new ApplicationLifecycle(async () => ({ stop }));
    await lifecycle.start();

    await Promise.all([lifecycle.stop('SIGTERM'), lifecycle.stop('SIGINT')]);

    expect(stop).toHaveBeenCalledOnce();
    expect(stop).toHaveBeenCalledWith('SIGTERM');
    expect(lifecycle.state).toBe('stopped');
  });

  it('waits for startup before stopping', async () => {
    let finishStartup: (() => void) | undefined;
    const stop = vi.fn(async () => undefined);
    const lifecycle = new ApplicationLifecycle(
      () =>
        new Promise((resolve) => {
          finishStartup = () => resolve({ stop });
        }),
    );

    const starting = lifecycle.start();
    const stopping = lifecycle.stop('test');
    finishStartup?.();
    await Promise.all([starting, stopping]);

    expect(stop).toHaveBeenCalledOnce();
    expect(lifecycle.state).toBe('stopped');
  });

  it('can stop safely after startup fails', async () => {
    const startupError = new Error('startup failed');
    const lifecycle = new ApplicationLifecycle(async () => {
      throw startupError;
    });

    await expect(lifecycle.start()).rejects.toBe(startupError);
    expect(lifecycle.state).toBe('failed');
    await expect(lifecycle.stop('startup failure')).resolves.toBeUndefined();
    expect(lifecycle.state).toBe('stopped');
  });

  it('does not restart after stopping', async () => {
    const lifecycle = new ApplicationLifecycle(async () => ({ stop: async () => undefined }));
    await lifecycle.stop('test');

    await expect(lifecycle.start()).rejects.toThrow('application runtime has already stopped');
  });
});
