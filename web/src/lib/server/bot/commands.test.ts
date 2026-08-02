import { commandName, createBot, splitFields } from '$lib/server/bot';
import {
  createBlindBoxFake,
  createBotHarness,
  createStatsFake,
} from '$lib/server/bot/testing/fakes';
import { chatMessageEvent } from '$lib/server/bot/testing/fixtures';
import { describe, expect, it } from 'vitest';

describe('command parsing', () => {
  it.each([
    ['!TEST arg', 'test'],
    ['!!test', '!test'],
    ['!', undefined],
    ['hello', undefined],
    [' !stats', undefined],
  ])('parses %j as %j', (text, expected) => {
    expect(commandName(text)).toBe(expected);
  });

  it('uses whitespace fields for command argument checks', () => {
    expect(splitFields('  !stats\t\n')).toEqual(['!stats']);
  });
});

describe('commands', () => {
  it('formats completed collections as one header plus one message per series', async () => {
    const blindBox = createBlindBoxFake({
      completed: async () => [
        { seriesName: 'Coobubus', usernames: 'alice, bob' },
        { seriesName: 'Lil Helpers', usernames: 'carol' },
      ],
    });
    const harness = createBotHarness({ blindBox });
    const bot = createBot(harness.dependencies);

    await bot.handleChatMessage(chatMessageEvent({ text: '!COLLECTIONS' }));

    expect(harness.chat.messages).toMatchInlineSnapshot(`
      [
        {
          "message": "The following chatters have completed the below blind box collections:",
        },
        {
          "message": "Coobubus: alice, bob",
        },
        {
          "message": "Lil Helpers: carol",
        },
      ]
    `);
  });

  it('formats leaderboard rows and preserves the failure message', async () => {
    const success = createBotHarness({
      stats: createStatsFake({
        leaderboard: async () => [
          { emoji: 'strong', username: 'alice', value: 8 },
          { emoji: 'lucky', username: 'bob', value: -2 },
        ],
      }),
    });
    await createBot(success.dependencies).handleChatMessage(
      chatMessageEvent({ text: '!leaderboard' }),
    );
    expect(success.chat.messages).toEqual([{ message: 'strong alice (8) | lucky bob (-2)' }]);

    const failure = createBotHarness({
      stats: createStatsFake({
        leaderboard: async () => {
          throw new Error('database failed');
        },
      }),
    });
    await createBot(failure.dependencies).handleChatMessage(
      chatMessageEvent({ text: '!leaderboard' }),
    );
    expect(failure.chat.messages).toEqual([{ message: 'Failed to get leaderboard' }]);
  });

  it('replies to an exact stats command and ignores commands with arguments', async () => {
    const harness = createBotHarness();
    const bot = createBot(harness.dependencies);

    await bot.handleChatMessage(chatMessageEvent({ text: '!stats   ' }));
    await bot.handleChatMessage(chatMessageEvent({ text: '!stats now', messageID: 'ignored' }));

    expect(harness.chat.messages).toEqual([
      {
        message: "alice's stats: STR: 3 | LUCK: 3",
        replyParentMessageID: 'message-1',
      },
    ]);
  });

  it('publishes the exact catalog collection overlay payload', async () => {
    const harness = createBotHarness({
      blindBox: createBlindBoxFake({ getCollection: async () => ['cutey'] }),
    });
    const bot = createBot(harness.dependencies);

    await bot.handleChatMessage(chatMessageEvent({ text: '!coobubu' }));

    expect(harness.overlay.events).toMatchInlineSnapshot(`
      [
        {
          "collection": [
            "cutey",
          ],
          "config": {
            "boxFrontFace": "/front.png",
            "boxSideFace": "/side.png",
            "displayColor": "#000000",
            "name": "Coobubus",
            "plushies": [
              {
                "emptyImage": "/empty.png",
                "image": "/cutey.png",
                "key": "cutey",
                "name": "Cutey",
                "series": "coobubu",
                "sortOrder": 1,
                "weight": 2,
              },
              {
                "emptyImage": "/empty.png",
                "image": "/secret.png",
                "key": "secret",
                "name": "Secret",
                "series": "coobubu",
                "sortOrder": 2,
                "weight": 1,
              },
            ],
            "redemptionTitle": "Cooper Series Blind Box",
            "revealSound": "/reveal.mp3",
            "series": "coobubu",
            "textColor": "#ffffff",
          },
          "type": "blindbox_display",
          "username": "alice",
        },
      ]
    `);
  });

  it('preserves the catalog command failure string', async () => {
    const harness = createBotHarness({
      blindBox: createBlindBoxFake({
        getCollection: async () => {
          throw new Error('database failed');
        },
      }),
    });
    const bot = createBot(harness.dependencies);

    await bot.handleChatMessage(chatMessageEvent({ text: '!coobubu' }));

    expect(harness.chat.messages).toEqual([{ message: "Failed to get alice's collection" }]);
  });
});
