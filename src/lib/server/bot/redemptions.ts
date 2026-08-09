import type { BlindBoxSeries } from '$lib/contracts/catalog';
import type {
  BlindBoxService,
  Logger,
  OverlayBus,
  Random,
  StatsService,
} from '$lib/server/application/ports';
import { pickWeightedPlushie } from '$lib/server/domain/blind-box/random';
import { formatStats } from '$lib/server/domain/stats/format';

import { randomInteger } from './random';
import type { Redemption, SendChat } from './types';

export interface RedemptionDependencies {
  series: readonly BlindBoxSeries[];
  stats: StatsService;
  blindBox: BlindBoxService;
  overlay: OverlayBus;
  random: Random;
  logger: Logger;
  sendChat: SendChat;
}

const redemptionFailureMessage = (username: string): string =>
  `@${username} sorry, the redemption failed. Please ping @modservo.`;

export const createRedemptions = ({
  series,
  stats,
  blindBox,
  overlay,
  random,
  logger,
  sendChat,
}: RedemptionDependencies): ReadonlyMap<string, Redemption> => {
  const redemptions = new Map<string, Redemption>();

  redemptions.set('Drink a Potion', async (event, signal) => {
    try {
      await stats.getOrCreate(event.userID, event.username);
    } catch (error) {
      logger.error('failed to get or create stats', { error, user: event.username });

      return;
    }

    if (signal.aborted) {
      return;
    }

    let definition;
    try {
      const index = randomInteger(random, stats.definitions.length, 'stat definition');
      definition = stats.definitions[index];
    } catch (error) {
      logger.error('failed to get random stat definition', { error });

      return;
    }

    const lost = randomInteger(random, 100, 'potion outcome') < 5;
    const amount = lost ? -1 : 1;

    try {
      await stats.adjust(event.userID, definition.name, amount);
    } catch (error) {
      logger.error('failed to modify stat', { error, user: event.username });

      return;
    }

    if (signal.aborted) {
      return;
    }

    const outcome = lost ? 'lost' : 'gained';
    await sendChat(
      `A shifty looking merchant hands ${event.username} a glittering potion. Without hesitation, they sink the whole drink. ${event.username} ${outcome} ${definition.longName}`,
      signal,
    );

    if (signal.aborted) {
      return;
    }

    let values;
    try {
      values = await stats.get(event.userID);
    } catch (error) {
      logger.error('failed to get stats', { error, user: event.username });

      return;
    }

    if (signal.aborted) {
      return;
    }

    await sendChat(formatStats(event.username, values), signal);
  });

  redemptions.set('Tempt the Dice', async (event, signal) => {
    await sendChat(`${event.username} has rolled with initiative.`, signal);

    if (signal.aborted) {
      return;
    }

    let values;
    try {
      values = await stats.getOrCreate(event.userID, event.username);
    } catch (error) {
      logger.error('failed to get stats', { error, user: event.username });

      return;
    }

    if (signal.aborted) {
      return;
    }

    await sendChat(formatStats(event.username, values), signal);
  });

  for (const config of series) {
    redemptions.set(config.redemptionTitle, async (event, signal) => {
      let plushie;
      try {
        plushie = pickWeightedPlushie(config.plushies, random);
      } catch (error) {
        logger.error('failed to pick plushie', { error, series: config.series });
        await sendChat(redemptionFailureMessage(event.username), signal);

        return;
      }

      let result;
      try {
        result = await blindBox.grant(event.userID, event.username, config.series, plushie.key);
      } catch (error) {
        logger.error('failed to redeem blind box', { error, user: event.username });
        await sendChat(redemptionFailureMessage(event.username), signal);

        return;
      }

      if (signal.aborted) {
        return;
      }

      overlay.publish({
        type: 'blindbox_redemption',
        username: event.username,
        plushie,
        isNew: result.isNew,
        collection: result.collection,
        config,
      });
      logger.info('blind box redeemed', {
        user: event.username,
        series: config.series,
        plushie: plushie.key,
        isNew: result.isNew,
      });
    });
  }

  return redemptions;
};
