import type { Logger } from '$lib/server/application/ports';

import { commandName, createCommands } from './commands';
import { createRedemptions } from './redemptions';
import { createTaskRunner } from './tasks';
import { createTriggers, passesTriggerChance } from './triggers';
import type { Bot, BotDependencies, ChatMessageEvent, SendChat } from './types';

const handlerTimeoutMilliseconds = 10_000;
const raidDelayMilliseconds = 5_000;

const createSendChat =
  (chat: BotDependencies['chat'], logger: Logger): SendChat =>
  async (message, signal, replyParentMessageID) => {
    if (signal.aborted) {
      return;
    }

    try {
      await chat.send(
        {
          message,
          ...(replyParentMessageID ? { replyParentMessageID } : {}),
        },
        signal,
      );
    } catch (error) {
      if (signal.aborted) {
        logger.debug('chat message cancelled', { message });
      } else {
        logger.error('failed to send message', { error, message });
      }
    }
  };

export const createBot = ({
  botUserID,
  series,
  stats,
  blindBox,
  overlay,
  chat,
  clock,
  random,
  logger,
}: BotDependencies): Bot => {
  const sendChat = createSendChat(chat, logger);
  const commands = createCommands({ series, stats, blindBox, overlay, logger, sendChat });
  const redemptions = createRedemptions({
    series,
    stats,
    blindBox,
    overlay,
    random,
    logger,
    sendChat,
  });
  const triggers = createTriggers(sendChat);
  const tasks = createTaskRunner(clock, logger);

  const processCommand = async (event: ChatMessageEvent, signal: AbortSignal): Promise<void> => {
    const name = commandName(event.text);

    if (!name) {
      return;
    }

    const command = commands.get(name);

    if (!command) {
      return;
    }

    logger.info('chat command received', {
      command: name,
      user: event.chatterUsername,
      message: event.text,
    });
    logger.info('executing command', { command: name, user: event.chatterUsername });
    await command(event, signal);
  };

  const processTriggers = async (event: ChatMessageEvent, signal: AbortSignal): Promise<void> => {
    for (const trigger of triggers) {
      if (signal.aborted || !trigger.matches(event)) {
        continue;
      }

      if (!passesTriggerChance(trigger.chance, random)) {
        logger.debug('trigger failed chance roll', { chance: trigger.chance });

        continue;
      }

      logger.info('executing trigger', {
        user: event.chatterUsername,
        message: event.text,
      });
      await trigger.execute(event, signal);
    }
  };

  const handleChatMessage: Bot['handleChatMessage'] = (event) => {
    if (event.chatterUserID === botUserID) {
      return Promise.resolve();
    }

    return tasks.run(
      'chat message handler',
      async (signal) => {
        try {
          await stats.recordActivity(event.chatterUserID, event.chatterUsername, clock.now());
        } catch (error) {
          logger.error('record chat activity', { error, user: event.chatterUsername });
        }

        if (signal.aborted) {
          return;
        }

        logger.debug('processing message', {
          user: event.chatterUsername,
          message: event.text,
        });
        await processCommand(event, signal);

        if (signal.aborted) {
          return;
        }

        await processTriggers(event, signal);
      },
      handlerTimeoutMilliseconds,
    );
  };

  const handleRedemption: Bot['handleRedemption'] = (event) =>
    tasks.run(
      'channel point redemption handler',
      async (signal) => {
        logger.info('channel point redemption', {
          user: event.username,
          reward: event.rewardTitle,
        });

        try {
          await stats.recordActivity(event.userID, event.username, clock.now());
        } catch (error) {
          logger.error('record redemption activity', { error, user: event.username });
        }

        if (signal.aborted) {
          return;
        }

        const redemption = redemptions.get(event.rewardTitle);

        if (!redemption) {
          return;
        }

        await redemption(event, signal);
      },
      handlerTimeoutMilliseconds,
    );

  const handleRaid: Bot['handleRaid'] = (event) =>
    tasks.run('raid handler', async (signal) => {
      await clock.sleep(raidDelayMilliseconds, signal);

      if (signal.aborted) {
        return;
      }

      await sendChat(`!so @${event.fromBroadcasterUsername}`, signal);
    });

  const stop: Bot['stop'] = async (reason = 'bot stopped') => {
    await tasks.stop(reason);
  };

  return { handleChatMessage, handleRedemption, handleRaid, stop };
};
