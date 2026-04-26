import { beforeEach, describe, expect, it, vi } from 'vitest';
import {
  BlindBoxQueue,
  type PlushieDisplayQueueItem,
  type PlushieRedemptionQueueItem,
} from './queue.svelte';

function makeRedemption(
  overrides: Partial<PlushieRedemptionQueueItem> = {},
): PlushieRedemptionQueueItem {
  return {
    type: 'redemption',
    username: 'user',
    plushie: { key: 'plushie-01', name: 'Plushie', image: '', emptyImage: '' },
    isNew: true,
    collection: [],
    config: {
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

function makeDisplay(overrides: Partial<PlushieDisplayQueueItem> = {}): PlushieDisplayQueueItem {
  return {
    type: 'display',
    username: 'user',
    collection: [],
    config: {
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
  let onRedemption: (item: PlushieRedemptionQueueItem) => Promise<void>;
  let onDisplay: (item: PlushieDisplayQueueItem) => Promise<void>;
  let queue: BlindBoxQueue;

  beforeEach(() => {
    onRedemption = vi.fn().mockResolvedValue(undefined);
    onDisplay = vi.fn().mockResolvedValue(undefined);
    queue = new BlindBoxQueue({ onRedemption, onDisplay });
  });

  it('calls onRedemption for a redemption item', async () => {
    const item = makeRedemption();
    queue.addRedemption(item);
    await queue.processNext();
    expect(onRedemption).toHaveBeenCalledOnce();
    expect(onRedemption).toHaveBeenCalledWith(item);
  });

  it('calls onDisplay for a display item', async () => {
    const item = makeDisplay();
    queue.addDisplay(item);
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

    queue.addDisplay(makeDisplay());
    queue.addRedemption(makeRedemption());
    await queue.processNext();

    expect(order[0]).toBe('redemption');
  });

  it('processes redemptions in FIFO order', async () => {
    const order: string[] = [];
    onRedemption = vi.fn().mockImplementation(async (item: PlushieRedemptionQueueItem) => {
      order.push(item.username);
    });
    queue = new BlindBoxQueue({ onRedemption, onDisplay });

    queue.addRedemption(makeRedemption({ username: 'first' }));
    queue.addRedemption(makeRedemption({ username: 'second' }));
    await queue.processNext();

    expect(order).toEqual(['first', 'second']);
  });

  it('processes display items in FIFO order', async () => {
    const order: string[] = [];
    onDisplay = vi.fn().mockImplementation(async (item: PlushieDisplayQueueItem) => {
      order.push(item.username);
    });
    queue = new BlindBoxQueue({ onRedemption, onDisplay });

    queue.addDisplay(makeDisplay({ username: 'first' }));
    queue.addDisplay(makeDisplay({ username: 'second' }));
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

    queue.addRedemption(makeRedemption());
    queue.addRedemption(makeRedemption());

    const first = queue.processNext();
    await queue.processNext(); // should be a no-op

    expect(onRedemption).toHaveBeenCalledTimes(1);
    resolve();
    await first;
  });

  it('automatically processes the next item after the current one finishes', async () => {
    queue.addRedemption(makeRedemption({ username: 'first' }));
    queue.addRedemption(makeRedemption({ username: 'second' }));
    await queue.processNext();
    expect(onRedemption).toHaveBeenCalledTimes(2);
  });

  it('does not lock isProcessing permanently if a handler throws', async () => {
    onRedemption = vi.fn().mockRejectedValue(new Error('handler error'));
    queue = new BlindBoxQueue({ onRedemption, onDisplay });

    queue.addRedemption(makeRedemption());
    await queue.processNext();

    // Should be able to process again after the error
    onRedemption = vi.fn().mockResolvedValue(undefined);
    queue = new BlindBoxQueue({ onRedemption, onDisplay });
    queue.addRedemption(makeRedemption());
    await queue.processNext();
    expect(onRedemption).toHaveBeenCalledOnce();
  });

  it('clears both queues and allows processing again', async () => {
    queue.addRedemption(makeRedemption());
    queue.addDisplay(makeDisplay());
    queue.clear();

    await queue.processNext();

    expect(onRedemption).not.toHaveBeenCalled();
    expect(onDisplay).not.toHaveBeenCalled();
  });
});
