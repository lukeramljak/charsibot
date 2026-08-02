import type { UserStat } from '$lib/contracts/viewer';
import { createBot } from '$lib/server/bot';
import {
  createBlindBoxFake,
  createBotHarness,
  createRandomFake,
  createStatsFake,
  testSeries,
} from '$lib/server/bot/testing/fakes';
import { redemptionEvent } from '$lib/server/bot/testing/fixtures';
import { describe, expect, it } from 'vitest';

describe('Drink a Potion', () => {
  it.each([
    [0, -1, 'lost'],
    [5, 1, 'gained'],
  ])('uses outcome roll %i for a %i adjustment', async (outcomeRoll, amount, outcome) => {
    const adjustments: { userID: string; statName: string; amount: number }[] = [];
    const stats = createStatsFake({
      adjust: async (userID, statName, adjustment) => {
        adjustments.push({ userID, statName, amount: adjustment });
      },
    });
    const random = createRandomFake([1, outcomeRoll]);
    const harness = createBotHarness({ stats, random });
    const bot = createBot(harness.dependencies);

    await bot.handleRedemption(redemptionEvent({ rewardTitle: 'Drink a Potion' }));

    expect(random.maximums).toEqual([2, 100]);
    expect(adjustments).toEqual([{ userID: 'viewer-1', statName: 'luck', amount }]);
    expect(harness.chat.messages).toEqual([
      {
        message: `A shifty looking merchant hands alice a glittering potion. Without hesitation, they sink the whole drink. alice ${outcome} Luck`,
      },
      { message: "alice's stats: STR: 4 | LUCK: 3" },
    ]);
  });

  it('stops before chat when stat initialization or adjustment fails', async () => {
    const initializeFailure = createBotHarness({
      stats: createStatsFake({
        getOrCreate: async () => {
          throw new Error('initialize failed');
        },
      }),
    });
    await createBot(initializeFailure.dependencies).handleRedemption(
      redemptionEvent({ rewardTitle: 'Drink a Potion' }),
    );
    expect(initializeFailure.chat.messages).toEqual([]);

    const adjustFailure = createBotHarness({
      stats: createStatsFake({
        adjust: async () => {
          throw new Error('adjust failed');
        },
      }),
    });
    await createBot(adjustFailure.dependencies).handleRedemption(
      redemptionEvent({ rewardTitle: 'Drink a Potion' }),
    );
    expect(adjustFailure.chat.messages).toEqual([]);
  });

  it('keeps the merchant message when the updated stat read fails', async () => {
    const stats = createStatsFake({
      get: async (): Promise<UserStat[]> => {
        throw new Error('read failed');
      },
    });
    const harness = createBotHarness({ stats, random: createRandomFake([0, 50]) });
    await createBot(harness.dependencies).handleRedemption(
      redemptionEvent({ rewardTitle: 'Drink a Potion' }),
    );
    expect(harness.chat.messages).toHaveLength(1);
  });
});

describe('Tempt the Dice', () => {
  it('sends initiative before creating and displaying stats', async () => {
    const phases: string[] = [];
    const stats = createStatsFake({
      getOrCreate: async () => {
        phases.push('stats');

        return [{ name: 'strength', shortName: 'STR', longName: 'Strength', value: 3 }];
      },
    });
    const harness = createBotHarness({ stats });
    harness.dependencies.chat = {
      send: async (message) => {
        phases.push(message.message);
      },
    };

    await createBot(harness.dependencies).handleRedemption(
      redemptionEvent({ rewardTitle: 'Tempt the Dice' }),
    );

    expect(phases).toEqual(['alice has rolled with initiative.', 'stats', "alice's stats: STR: 3"]);
  });
});

describe('catalog redemption', () => {
  it('grants the weighted plushie and publishes the exact overlay payload', async () => {
    const grants: unknown[] = [];
    const blindBox = createBlindBoxFake({
      grant: async (...arguments_) => {
        grants.push(arguments_);

        return { isNew: false, collection: ['cutey'] };
      },
    });
    const harness = createBotHarness({ blindBox, random: createRandomFake([0]) });
    const bot = createBot(harness.dependencies);

    await bot.handleRedemption(redemptionEvent({ rewardTitle: 'Cooper Series Blind Box' }));

    expect(grants).toEqual([['viewer-1', 'alice', 'coobubu', 'cutey']]);
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
          "isNew": false,
          "plushie": {
            "emptyImage": "/empty.png",
            "image": "/cutey.png",
            "key": "cutey",
            "name": "Cutey",
            "series": "coobubu",
            "sortOrder": 1,
            "weight": 2,
          },
          "type": "blindbox_redemption",
          "username": "alice",
        },
      ]
    `);
  });

  it('preserves the exact user-facing failure string for pick and grant errors', async () => {
    const pickFailure = createBotHarness();
    const pickBot = createBot({
      ...pickFailure.dependencies,
      series: [
        {
          ...testSeries,
          redemptionTitle: 'Broken Box',
          plushies: [],
        },
      ],
    });
    await pickBot.handleRedemption(redemptionEvent({ rewardTitle: 'Broken Box' }));

    const grantFailure = createBotHarness({
      blindBox: createBlindBoxFake({
        grant: async () => {
          throw new Error('database failed');
        },
      }),
    });
    await createBot(grantFailure.dependencies).handleRedemption(
      redemptionEvent({ rewardTitle: 'Cooper Series Blind Box' }),
    );

    const expected = {
      message: '@alice sorry, the redemption failed. Please ping @modservo.',
    };
    expect(pickFailure.chat.messages).toEqual([expected]);
    expect(grantFailure.chat.messages).toEqual([expected]);
  });
});
