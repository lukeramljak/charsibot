import type { ChatSender } from '$lib/server/application/ports';
import { createBot } from '$lib/server/bot';
import { describe, expect, it } from 'vitest';

import { createBotHarness, createStatsFake } from './testing/fakes';
import { chatMessageEvent, raidEvent, redemptionEvent } from './testing/fixtures';

const flushMicrotasks = async (): Promise<void> => {
  for (let count = 0; count < 8; count += 1) {
    await Promise.resolve();
  }
};

describe('bot dispatch', () => {
  it('suppresses the configured bot user before activity, commands, and triggers', async () => {
    const activity: string[] = [];
    const harness = createBotHarness({
      stats: createStatsFake({
        recordActivity: async (userID) => {
          activity.push(userID);
        },
      }),
    });
    const bot = createBot(harness.dependencies);

    await bot.handleChatMessage(
      chatMessageEvent({ chatterUserID: 'bot-user', text: '!stats coming' }),
    );

    expect(activity).toEqual([]);
    expect(harness.chat.messages).toEqual([]);
    expect(harness.overlay.events).toEqual([]);
  });

  it('records activity before command work and continues when activity fails', async () => {
    const phases: string[] = [];
    const chat: ChatSender = {
      send: async () => {
        phases.push('chat');
      },
    };
    const stats = createStatsFake({
      recordActivity: async () => {
        phases.push('activity');
        throw new Error('activity unavailable');
      },
      getOrCreate: async () => {
        phases.push('command');

        return [{ name: 'strength', shortName: 'STR', longName: 'Strength', value: 3 }];
      },
    });
    const harness = createBotHarness({ stats, chat });
    const bot = createBot(harness.dependencies);

    await bot.handleChatMessage(chatMessageEvent({ text: '!stats' }));

    expect(phases).toEqual(['activity', 'command', 'chat']);
    expect(harness.logger.entries.some((entry) => entry.message === 'record chat activity')).toBe(
      true,
    );
  });

  it('starts separate message handlers concurrently', async () => {
    const resolvers: (() => void)[] = [];
    const stats = createStatsFake({
      recordActivity: () =>
        new Promise<void>((resolve) => {
          resolvers.push(resolve);
        }),
    });
    const harness = createBotHarness({ stats });
    const bot = createBot(harness.dependencies);

    const first = bot.handleChatMessage(chatMessageEvent({ chatterUserID: 'one' }));
    const second = bot.handleChatMessage(chatMessageEvent({ chatterUserID: 'two' }));
    expect(resolvers).toHaveLength(2);

    resolvers.forEach((resolve) => resolve());
    await Promise.all([first, second]);
  });

  it('records activity for unknown and case-mismatched redemptions', async () => {
    const activity: string[] = [];
    const stats = createStatsFake({
      recordActivity: async (userID) => {
        activity.push(userID);
      },
    });
    const harness = createBotHarness({ stats });
    const bot = createBot(harness.dependencies);

    await bot.handleRedemption(redemptionEvent({ rewardTitle: 'Unknown Reward' }));
    await bot.handleRedemption(redemptionEvent({ rewardTitle: 'drink a potion' }));

    expect(activity).toEqual(['viewer-1', 'viewer-1']);
    expect(harness.chat.messages).toEqual([]);
  });

  it('aborts at the deadline but tracks an abort-ignoring operation until it settles', async () => {
    let releaseActivity: (() => void) | undefined;
    const stats = createStatsFake({
      recordActivity: () =>
        new Promise<void>((resolve) => {
          releaseActivity = resolve;
        }),
    });
    const harness = createBotHarness({ stats });
    const bot = createBot(harness.dependencies);

    const task = bot.handleChatMessage(chatMessageEvent());
    let taskSettled = false;
    void task.then(() => {
      taskSettled = true;
    });
    expect(harness.clock.pending()).toContain(10_000);
    harness.clock.resolveNext(10_000);
    await flushMicrotasks();

    expect(
      harness.logger.entries.some(
        (entry) =>
          entry.level === 'error' &&
          (entry.attributes?.error as Error | undefined)?.message.includes('exceeded 10000ms'),
      ),
    ).toBe(true);
    expect(taskSettled).toBe(false);

    let stopSettled = false;
    const stop = bot.stop().then(() => {
      stopSettled = true;
    });
    await flushMicrotasks();
    expect(stopSettled).toBe(false);

    releaseActivity?.();
    await Promise.all([task, stop]);
    expect(taskSettled).toBe(true);
    expect(stopSettled).toBe(true);
  });

  it('cancels a losing deadline sleep without a handler failure', async () => {
    const harness = createBotHarness();
    const bot = createBot(harness.dependencies);

    await bot.handleChatMessage(chatMessageEvent());
    await flushMicrotasks();

    expect(harness.clock.pending()).toEqual([]);
    expect(
      harness.logger.entries.some((entry) => entry.message === 'chat message handler failed'),
    ).toBe(false);
  });

  it('delays raid shoutouts for five seconds', async () => {
    const harness = createBotHarness();
    const bot = createBot(harness.dependencies);

    const task = bot.handleRaid(raidEvent({ fromBroadcasterUsername: 'some_raider' }));
    expect(harness.chat.messages).toEqual([]);
    expect(harness.clock.pending()).toContain(5_000);

    harness.clock.resolveNext(5_000);
    await task;
    expect(harness.chat.messages).toEqual([{ message: '!so @some_raider' }]);
  });

  it('aborts and awaits tracked chat and raid tasks during stop', async () => {
    const sentSignals: AbortSignal[] = [];
    const chat: ChatSender = {
      send: (_message, signal) =>
        new Promise<void>((resolve, reject) => {
          if (!signal) {
            resolve();

            return;
          }
          sentSignals.push(signal);
          signal.addEventListener('abort', () => reject(signal.reason), { once: true });
        }),
    };
    const harness = createBotHarness({ chat });
    const bot = createBot(harness.dependencies);

    const messageTask = bot.handleChatMessage(chatMessageEvent({ text: '!stats' }));
    const raidTask = bot.handleRaid(raidEvent());
    await flushMicrotasks();
    expect(sentSignals).toHaveLength(1);
    expect(harness.clock.pending()).toContain(5_000);

    await bot.stop('test shutdown');
    await Promise.all([messageTask, raidTask]);
    expect(sentSignals[0].aborted).toBe(true);
    expect(harness.clock.pending()).toEqual([]);

    await bot.handleChatMessage(chatMessageEvent({ text: '!stats' }));
    expect(sentSignals).toHaveLength(1);
  });
});
