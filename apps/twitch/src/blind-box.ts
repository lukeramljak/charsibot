import type { PlushieData, RewardColumn } from "./types";

/**
 * Selects a random plushie based on weighted probabilities
 * Higher weight = higher chance of selection
 */
export const getWeightedRandomPlushie = (
  plushies: PlushieData[]
): RewardColumn => {
  const weightedList: RewardColumn[] = [];

  // Build weighted list by repeating items based on their weight
  plushies.forEach((plushie) => {
    for (let i = 0; i < plushie.weight; i++) {
      weightedList.push(plushie.key);
    }
  });

  // Select random item from weighted list
  const randomIndex = Math.floor(Math.random() * weightedList.length);
  return weightedList[randomIndex];
};
