import { beforeEach, describe, expect, it, vi } from 'vitest';
import { BlindBoxQueue } from './queue.svelte';
import type { BlindBoxRedemptionEvent, CollectionDisplayEvent } from '$lib/types';

function makeRedemption(overrides: Partial<BlindBoxRedemptionEvent> = {}): BlindBoxRedemptionEvent {
  return {
    type: 'blindbox_redemption',
    username: 'user',
    plushie: { id: 1, series: 'test', key: 'plushie-01', sortOrder: 0, weight: 1, name: 'Plushie', image: '', emptyImage: '' },
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

  it('calls onRedemption for a redemption item', async () => {
    const item = makeRedemption();
    queue.add(item);
    await queue.processNext();
    expect(onRedemption).toHaveBeenCalledOnce();
    expect(onRedemption).toHaveBeenCalledWith(item);
  });

  it('calls onDisplay for a display item', async () => {
    const item = makeDisplay();
    queue.add(item);
    await queue.processNext();
    expect(onDisplay).toHaveBeenCalledOnce();
    expect(onDisplay).toHaveBeenCalledWith(item);
  });

  it('processes redemptions before display items', async () => {
    const order: string[] = [];
    onDisplay = vi.fn().mockImplementation(async () => {
      order.push('display');
    });
    onRedemption = vi.fn().mockImplementation(async () => {
      order.push('redemption');
    });
    queue = new BlindBoxQueue({ onRedemption, onDisplay });

    queue.add(makeDisplay());
    queue.add(makeRedemption());
    await queue.processNext();

    expect(order[0]).toBe('redemption');
  });

  it('processes redemptions in FIFO order', async () => {
    const order: string[] = [];
    onRedemption = vi.fn().mockImplementation(async (item: BlindBoxRedemptionEvent) => {
      order.push(item.username);
    });
    queue = new BlindBoxQueue({ onRedemption, onDisplay });

    queue.add(makeRedemption({ username: 'first' }));
    queue.add(makeRedemption({ username: 'second' }));
    await queue.processNext();

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
    await queue.processNext();

    expect(order).toEqual(['first', 'second']);
  });

  it('is a no-op if already processing', async () => {
    let resolve!: () => void;
    onRedemption = vi.fn().mockReturnValue(
      new Promise<void>((r) => {
        resolve = r;
      }),
    );
    queue = new BlindBoxQueue({ onRedemption, onDisplay });

    queue.add(makeRedemption());
    queue.add(makeRedemption());

    const first = queue.processNext();
    await queue.processNext(); // should be a no-op

    expect(onRedemption).toHaveBeenCalledTimes(1);
    resolve();
    await first;
  });

  it('automatically processes the next item after the current one finishes', async () => {
    queue.add(makeRedemption({ username: 'first' }));
    queue.add(makeRedemption({ username: 'second' }));
    await queue.processNext();
    expect(onRedemption).toHaveBeenCalledTimes(2);
  });

  it('does not lock isProcessing permanently if a handler throws', async () => {
    onRedemption = vi.fn().mockRejectedValue(new Error('handler error'));
    queue = new BlindBoxQueue({ onRedemption, onDisplay });

    queue.add(makeRedemption());
    await queue.processNext();

    // Should be able to process again after the error
    onRedemption = vi.fn().mockResolvedValue(undefined);
    queue = new BlindBoxQueue({ onRedemption, onDisplay });
    queue.add(makeRedemption());
    await queue.processNext();
    expect(onRedemption).toHaveBeenCalledOnce();
  });

  it('picks up items added to the queue while a handler is executing', async () => {
    let resolveFirst!: () => void;
    onRedemption = vi
      .fn()
      .mockImplementationOnce(
        () =>
          new Promise<void>((r) => {
            resolveFirst = r;
          }),
      )
      .mockResolvedValue(undefined);
    queue = new BlindBoxQueue({ onRedemption, onDisplay });

    queue.add(makeRedemption({ username: 'first' }));
    const processing = queue.processNext();

    // Add a second item while the first handler is still running
    queue.add(makeRedemption({ username: 'second' }));
    resolveFirst();
    await processing;

    expect(onRedemption).toHaveBeenCalledTimes(2);
  });

  it('is a no-op if the queue is empty', async () => {
    await queue.processNext();
    expect(onRedemption).not.toHaveBeenCalled();
    expect(onDisplay).not.toHaveBeenCalled();
  });

  it('processes display items after all redemptions are drained', async () => {
    const order: string[] = [];
    onRedemption = vi.fn().mockImplementation(async () => {
      order.push('redemption');
    });
    onDisplay = vi.fn().mockImplementation(async () => {
      order.push('display');
    });
    queue = new BlindBoxQueue({ onRedemption, onDisplay });

    queue.add(makeRedemption());
    queue.add(makeDisplay());
    await queue.processNext();

    expect(order).toEqual(['redemption', 'display']);
  });

  it('clears both queues and allows processing again', async () => {
    queue.add(makeRedemption());
    queue.add(makeDisplay());
    queue.clear();

    await queue.processNext();

    expect(onRedemption).not.toHaveBeenCalled();
    expect(onDisplay).not.toHaveBeenCalled();
  });
});
