import type { BlindBoxSeries } from '$lib/contracts/catalog';
import type {
  BlindBoxService,
  Logger,
  OverlayBus,
  StatsService,
} from '$lib/server/application/ports';
import { formatStats } from '$lib/server/domain/stats/format';

import { splitFields } from './text';
import type { Command, SendChat } from './types';

export interface CommandDependencies {
  series: readonly BlindBoxSeries[];
  stats: StatsService;
  blindBox: BlindBoxService;
  overlay: OverlayBus;
  logger: Logger;
  sendChat: SendChat;
}

export const createCommands = ({
  series,
  stats,
  blindBox,
  overlay,
  logger,
  sendChat,
}: CommandDependencies): ReadonlyMap<string, Command> => {
  const commands = new Map<string, Command>();

  commands.set('collections', async (_event, signal) => {
    let collections;
    try {
      collections = await blindBox.completed();
    } catch (error) {
      logger.error('failed to get completed collections', { error });

      return;
    }

    if (signal.aborted) {
      return;
    }

    await sendChat(
      'The following chatters have completed the below blind box collections:',
      signal,
    );
    for (const collection of collections) {
      if (signal.aborted) {
        return;
      }

      await sendChat(`${collection.seriesName}: ${collection.usernames}`, signal);
    }
  });

  commands.set('leaderboard', async (_event, signal) => {
    let rows;
    try {
      rows = await stats.leaderboard();
    } catch (error) {
      logger.error('failed to get leaderboard', { error });
      await sendChat('Failed to get leaderboard', signal);

      return;
    }

    if (signal.aborted) {
      return;
    }

    const message = rows.map((row) => `${row.emoji} ${row.username} (${row.value})`).join(' | ');
    await sendChat(message, signal);
  });

  commands.set('stats', async (event, signal) => {
    if (splitFields(event.text).length !== 1) {
      return;
    }

    let values;
    try {
      values = await stats.getOrCreate(event.chatterUserID, event.chatterUsername);
    } catch (error) {
      logger.error('failed to get stats', { error, user: event.chatterUsername });

      return;
    }

    if (signal.aborted) {
      return;
    }

    await sendChat(formatStats(event.chatterUsername, values), signal, event.messageID);
  });

  for (const config of series) {
    commands.set(config.series, async (event, signal) => {
      if (splitFields(event.text).length !== 1) {
        return;
      }

      let collection;
      try {
        collection = await blindBox.getCollection(event.chatterUserID, config.series);
      } catch (error) {
        logger.error('failed to get collection', { error, user: event.chatterUsername });
        await sendChat(`Failed to get ${event.chatterUsername}'s collection`, signal);

        return;
      }

      if (signal.aborted) {
        return;
      }

      overlay.publish({
        type: 'blindbox_display',
        username: event.chatterUsername,
        collection,
        config,
      });
      logger.info('displaying collection', {
        user: event.chatterUsername,
        series: config.series,
        size: collection.length,
      });
    });
  }

  return commands;
};

export const commandName = (text: string): string | undefined => {
  if (!text.startsWith('!')) {
    return undefined;
  }

  const fields = splitFields(text.toLowerCase());
  if (fields.length === 0) {
    return undefined;
  }

  const name = fields[0].startsWith('!') ? fields[0].slice(1) : fields[0];

  return name || undefined;
};
