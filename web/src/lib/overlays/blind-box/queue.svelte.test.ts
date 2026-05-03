import type { BlindBoxRedemptionEvent, CollectionDisplayEvent } from '$lib/types';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { BlindBoxQueue, type QueueItem } from './queue.svelte';

const flushPromises = () => new Promise<void>((r) => setTimeout(r));

function deferred() {
  let resolve!: () => void;
  const promise = new Promise<void>((r) => (resolve = r));
  return { promise, resolve };
}

function makeRedemption(overrides: Partial<BlindBoxRedemptionEvent> = {}): BlindBoxRedemptionEvent {
  return {
    type: 'blindbox_redemption',
    username: 'user',
    plushie: {
      id: 1,
      series: 'test',
      key: 'plushie-01',
      sortOrder: 0,
      weight: 1,
      name: 'Plushie',
      image: '',
      emptyImage: '',
    },
    isNew: true,
    collection: [],
    config: {
      redemptionTitle: '',
      series: 'test',
      name: 'Test Series',
      plushies: [],
      boxFrontFace: '',
      boxSideFace: '',
      revealSound: '',
      displayColor: '',
      textColor: '',
    },
    ...overrides,
  };
}

function makeDisplay(overrides: Partial<CollectionDisplayEvent> = {}): CollectionDisplayEvent {
  return {
    type: 'blindbox_display',
    username: 'user',
    collection: [],
    config: {
      redemptionTitle: '',
      series: 'test',
      name: 'Test Series',
      plushies: [],
      boxFrontFace: '',
      boxSideFace: '',
      revealSound: '',
      displayColor: '',
      textColor: '',
    },
    ...overrides,
  };
}

describe('BlindBoxQueue', () => {
  let onRedemption: (item: BlindBoxRedemptionEvent) => Promise<void>;
  let onDisplay: (item: CollectionDisplayEvent) => Promise<void>;
  let queue: BlindBoxQueue;

  beforeEach(() => {
    onRedemption = vi.fn().mockResolvedValue(undefined);
    onDisplay = vi.fn().mockResolvedValue(undefined);
    queue = new BlindBoxQueue({ onRedemption, onDisplay });
  });

  it('calls onRedemption for a redemption item', () => {
    const item = makeRedemption();
    queue.add(item);
    expect(onRedemption).toHaveBeenCalledOnce();
    expect(onRedemption).toHaveBeenCalledWith(item);
  });

  it('calls onDisplay for a display item', () => {
    const item = makeDisplay();
    queue.add(item);
    expect(onDisplay).toHaveBeenCalledOnce();
    expect(onDisplay).toHaveBeenCalledWith(item);
  });

  it('is a no-op if the queue is empty', () => {
    expect(onRedemption).not.toHaveBeenCalled();
    expect(onDisplay).not.toHaveBeenCalled();
  });

  it('is a no-op if already processing', () => {
    const { promise, resolve } = deferred();
    onRedemption = vi.fn().mockReturnValueOnce(promise).mockResolvedValue(undefined);
    queue = new BlindBoxQueue({ onRedemption, onDisplay });

    queue.add(makeRedemption());
    queue.add(makeRedemption());

    expect(onRedemption).toHaveBeenCalledTimes(1);
    resolve();
  });

  it('automatically processes the next item after the current one finishes', async () => {
    queue.add(makeRedemption({ username: 'first' }));
    queue.add(makeRedemption({ username: 'second' }));
    await flushPromises();
    expect(onRedemption).toHaveBeenCalledTimes(2);
  });

  it('picks up items added to the queue while a handler is executing', async () => {
    const { promise, resolve } = deferred();
    onRedemption = vi.fn().mockReturnValueOnce(promise).mockResolvedValue(undefined);
    queue = new BlindBoxQueue({ onRedemption, onDisplay });

    queue.add(makeRedemption({ username: 'first' }));
    queue.add(makeRedemption({ username: 'second' }));
    resolve();
    await flushPromises();

    expect(onRedemption).toHaveBeenCalledTimes(2);
  });

  it('processes redemptions in FIFO order', async () => {
    const order: string[] = [];
    onRedemption = vi.fn().mockImplementation(async (item: BlindBoxRedemptionEvent) => {
      order.push(item.username);
    });
    queue = new BlindBoxQueue({ onRedemption, onDisplay });

    queue.add(makeRedemption({ username: 'first' }));
    queue.add(makeRedemption({ username: 'second' }));
    await flushPromises();

    expect(order).toEqual(['first', 'second']);
  });

  it('processes display items in FIFO order', async () => {
    const order: string[] = [];
    onDisplay = vi.fn().mockImplementation(async (item: CollectionDisplayEvent) => {
      order.push(item.username);
    });
    queue = new BlindBoxQueue({ onRedemption, onDisplay });

    queue.add(makeDisplay({ username: 'first' }));
    queue.add(makeDisplay({ username: 'second' }));
    await flushPromises();

    expect(order).toEqual(['first', 'second']);
  });

  it('processes redemptions before display items when both are queued', async () => {
    const order: QueueItem['type'][] = [];
    const { promise, resolve } = deferred();
    onRedemption = vi
      .fn()
      .mockReturnValueOnce(promise)
      .mockImplementation(async () => order.push('blindbox_redemption'));
    onDisplay = vi.fn().mockImplementation(async () => order.push('blindbox_display'));
    queue = new BlindBoxQueue({ onRedemption, onDisplay });

    queue.add(makeRedemption()); // starts processing (blocked)
    queue.add(makeDisplay());
    queue.add(makeRedemption());
    resolve();
    await flushPromises();

    expect(order).toEqual(['blindbox_redemption', 'blindbox_display']);
  });

  it('processes display items after all redemptions are drained', async () => {
    const order: QueueItem['type'][] = [];
    onRedemption = vi.fn().mockImplementation(async () => order.push('blindbox_redemption'));
    onDisplay = vi.fn().mockImplementation(async () => order.push('blindbox_display'));
    queue = new BlindBoxQueue({ onRedemption, onDisplay });

    queue.add(makeRedemption());
    queue.add(makeDisplay());
    await flushPromises();

    expect(order).toEqual(['blindbox_redemption', 'blindbox_display']);
  });

  it('does not lock isProcessing permanently if a handler throws', async () => {
    onRedemption = vi.fn().mockRejectedValue(new Error('handler error'));
    queue = new BlindBoxQueue({ onRedemption, onDisplay });
    queue.add(makeRedemption());
    await flushPromises();

    onRedemption = vi.fn().mockResolvedValue(undefined);
    queue = new BlindBoxQueue({ onRedemption, onDisplay });
    queue.add(makeRedemption());
    expect(onRedemption).toHaveBeenCalledOnce();
  });

  it('clears the queue so pending items are not processed', async () => {
    const { promise, resolve } = deferred();
    onRedemption = vi.fn().mockReturnValueOnce(promise);
    queue = new BlindBoxQueue({ onRedemption, onDisplay });

    queue.add(makeRedemption()); // starts processing (blocked)
    queue.add(makeDisplay()); // queued
    queue.clear();
    resolve();
    await flushPromises();

    expect(onRedemption).toHaveBeenCalledTimes(1);
    expect(onDisplay).not.toHaveBeenCalled();
  });
});
