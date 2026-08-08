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
});
