import type { Random } from '$lib/server/application/ports';

import { randomInteger } from './random';
import { tokenizeTriggerWords } from './text';
import type { SendChat, Trigger } from './types';

const comeTriggerWords = new Set(['come', 'coming', 'cum', 'came']);

export const createTriggers = (sendChat: SendChat): readonly Trigger[] => [
  {
    chance: 20,
    matches: (event) => {
      if (event.text.toLowerCase() === 'no coming') {
        return false;
      }

      return tokenizeTriggerWords(event.text).some((word) => comeTriggerWords.has(word));
    },
    execute: async (event, signal) => {
      await sendChat('no coming', signal, event.messageID);
    },
  },
];

export const passesTriggerChance = (chance: number, random: Random): boolean => {
  if (chance <= 0 || chance >= 100) {
    return true;
  }

  const roll = randomInteger(random, 100, 'trigger chance') + 1;

  return roll <= chance;
};
