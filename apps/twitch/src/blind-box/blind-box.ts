import type { CollectionType, PlushieData, RewardColumn } from './types';

export const commandToCollectionType: Record<string, CollectionType> = {
  '!coobubu': 'coobubu',
  '!olliepop': 'olliepops'
};

export const redemptionToCollectionType: Record<string, CollectionType> = {
  'Cooper Series Blind Box': 'coobubu',
  'Ollie Series Blind Box': 'olliepops'
};

/**
 * Selects a random plushie based on weighted probabilities
 * Higher weight = higher chance of selection
 */
export const getWeightedRandomPlushie = (plushies: PlushieData[]): RewardColumn => {
  const weightedList: RewardColumn[] = [];

  // Build weighted list by repeating items based on their weight
  plushies.forEach(plushie => {
    for (let i = 0; i < plushie.weight; i++) {
      weightedList.push(plushie.key);
    }
  });

  // Select random item from weighted list
  const randomIndex = Math.floor(Math.random() * weightedList.length);
  return weightedList[randomIndex];
};
