import { describe, expect, it } from 'vitest';

import type { OverlayEvent } from '$lib/contracts/overlay';
import { createLoggerFake } from '$lib/server/bot/testing/fakes';

import { createOverlayBus } from './overlay-bus';

const chatEvent = (message: string): OverlayEvent => ({
  type: 'chat_command',
  message,
});

describe('createOverlayBus', () => {
  it('delivers published events to a subscriber', async () => {
    const bus = createOverlayBus(createLoggerFake());
    const subscription = bus.subscribe();

    bus.publish(chatEvent('hello'));
    subscription.close();

    const collected: OverlayEvent[] = [];
    for await (const event of subscription) {
      collected.push(event);
    }

    expect(collected).toEqual([chatEvent('hello')]);
  });

  it('delivers events to multiple subscribers', async () => {
    const bus = createOverlayBus(createLoggerFake());
    const sub1 = bus.subscribe();
    const sub2 = bus.subscribe();

    bus.publish(chatEvent('broadcast'));
    sub1.close();
    sub2.close();

    const collected1: OverlayEvent[] = [];
    for await (const event of sub1) collected1.push(event);

    const collected2: OverlayEvent[] = [];
    for await (const event of sub2) collected2.push(event);

    expect(collected1).toEqual([chatEvent('broadcast')]);
    expect(collected2).toEqual([chatEvent('broadcast')]);
  });

  it('drops events when a client buffer is full', async () => {
    const logger = createLoggerFake();
    const bus = createOverlayBus(logger);
    const subscription = bus.subscribe();

    for (let i = 0; i < 12; i++) {
      bus.publish(chatEvent(`msg-${i}`));
    }

    subscription.close();

    const collected: OverlayEvent[] = [];
    for await (const event of subscription) collected.push(event);

    expect(collected).toHaveLength(10);
    expect(collected[0]).toEqual(chatEvent('msg-0'));
    expect(collected[9]).toEqual(chatEvent('msg-9'));

    const warnings = logger.entries.filter((e) => e.level === 'warn');
    expect(warnings).toHaveLength(2);
    expect(warnings[0].message).toBe('SSE client buffer full, dropping event');
  });

  it('does not deliver events after subscription is closed', async () => {
    const bus = createOverlayBus(createLoggerFake());
    const subscription = bus.subscribe();

    bus.publish(chatEvent('before'));
    subscription.close();
    bus.publish(chatEvent('after'));

    const collected: OverlayEvent[] = [];
    for await (const event of subscription) collected.push(event);

    expect(collected).toEqual([chatEvent('before')]);
  });

  it('does not deliver events to unsubscribed clients', async () => {
    const bus = createOverlayBus(createLoggerFake());
    const sub1 = bus.subscribe();
    sub1.close();

    const sub2 = bus.subscribe();
    bus.publish(chatEvent('only-sub2'));
    sub2.close();

    const collected1: OverlayEvent[] = [];
    for await (const event of sub1) collected1.push(event);

    const collected2: OverlayEvent[] = [];
    for await (const event of sub2) collected2.push(event);

    expect(collected1).toEqual([]);
    expect(collected2).toEqual([chatEvent('only-sub2')]);
  });

  it('closes all subscriptions when bus is closed', async () => {
    const bus = createOverlayBus(createLoggerFake());
    const sub1 = bus.subscribe();
    const sub2 = bus.subscribe();

    bus.publish(chatEvent('before-close'));
    bus.close();

    const collected1: OverlayEvent[] = [];
    for await (const event of sub1) collected1.push(event);

    const collected2: OverlayEvent[] = [];
    for await (const event of sub2) collected2.push(event);

    expect(collected1).toEqual([chatEvent('before-close')]);
    expect(collected2).toEqual([chatEvent('before-close')]);
  });

  it('ignores publishes after bus is closed', () => {
    const logger = createLoggerFake();
    const bus = createOverlayBus(logger);
    bus.close();

    bus.publish(chatEvent('ignored'));

    expect(logger.entries).toEqual([]);
  });

  it('rejects subscribe after bus is closed', () => {
    const bus = createOverlayBus(createLoggerFake());
    bus.close();

    expect(() => bus.subscribe()).toThrow('overlay bus is closed');
  });

  it('wakes a waiting subscriber when an event arrives', async () => {
    const bus = createOverlayBus(createLoggerFake());
    const subscription = bus.subscribe();

    const eventPromise = (async () => {
      for await (const event of subscription) {
        return event;
      }
    })();

    await Promise.resolve();
    bus.publish(chatEvent('wake'));
    subscription.close();

    expect(await eventPromise).toEqual(chatEvent('wake'));
  });

  it('does not block the publisher when a client buffer is full', () => {
    const logger = createLoggerFake();
    const bus = createOverlayBus(logger);
    const subscription = bus.subscribe();

    for (let i = 0; i < 12; i++) {
      bus.publish(chatEvent(`msg-${i}`));
    }

    subscription.close();

    expect(logger.entries.filter((e) => e.level === 'warn')).toHaveLength(2);
  });

  it('preserves FIFO order under rapid publishing', async () => {
    const bus = createOverlayBus(createLoggerFake());
    const subscription = bus.subscribe();

    for (let i = 0; i < 10; i++) {
      bus.publish(chatEvent(`rapid-${i}`));
    }

    subscription.close();

    const collected: OverlayEvent[] = [];
    for await (const event of subscription) collected.push(event);

    expect(collected).toHaveLength(10);
    for (let i = 0; i < 10; i++) {
      expect(collected[i]).toEqual(chatEvent(`rapid-${i}`));
    }
  });

  it('maintains insertion order across interleaved event types', async () => {
    const bus = createOverlayBus(createLoggerFake());
    const subscription = bus.subscribe();

    const display: OverlayEvent = {
      type: 'blindbox_display',
      username: 'alice',
      collection: [],
      config: {
        series: 's',
        redemptionTitle: 't',
        name: 'n',
        revealSound: '',
        boxFrontFace: '',
        boxSideFace: '',
        displayColor: '',
        textColor: '',
        plushies: [],
      },
    };

    bus.publish(chatEvent('first'));
    bus.publish(display);
    bus.publish(chatEvent('third'));
    subscription.close();

    const collected: OverlayEvent[] = [];
    for await (const event of subscription) collected.push(event);

    expect(collected[0].type).toBe('chat_command');
    expect(collected[1].type).toBe('blindbox_display');
    expect(collected[2].type).toBe('chat_command');
  });

  it('delivers events to a slow consumer without blocking a fast one', async () => {
    const bus = createOverlayBus(createLoggerFake());
    const slow = bus.subscribe();
    const fast = bus.subscribe();

    bus.publish(chatEvent('msg-1'));
    bus.publish(chatEvent('msg-2'));

    const fastCollected: OverlayEvent[] = [];
    fast.close();
    for await (const event of fast) fastCollected.push(event);
    expect(fastCollected).toHaveLength(2);

    bus.publish(chatEvent('msg-3'));
    slow.close();

    const slowCollected: OverlayEvent[] = [];
    for await (const event of slow) slowCollected.push(event);
    expect(slowCollected).toHaveLength(3);
    expect(slowCollected).toEqual([chatEvent('msg-1'), chatEvent('msg-2'), chatEvent('msg-3')]);
  });

  it('delivers events published concurrently with async iteration', async () => {
    const bus = createOverlayBus(createLoggerFake());
    const subscription = bus.subscribe();

    const collected: OverlayEvent[] = [];
    const reader = (async () => {
      for await (const event of subscription) {
        collected.push(event);

        if (collected.length === 3) {
          subscription.close();
        }
      }
    })();

    await Promise.resolve();
    bus.publish(chatEvent('a'));
    await Promise.resolve();
    bus.publish(chatEvent('b'));
    await Promise.resolve();
    bus.publish(chatEvent('c'));

    await reader;

    expect(collected).toEqual([chatEvent('a'), chatEvent('b'), chatEvent('c')]);
  });
});
