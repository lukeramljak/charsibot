import type { PlushiesMap, RewardColumn } from '@/blind-box/types';

export const buildWeightedPlushieList = (plushies: PlushiesMap): RewardColumn[] => {
  const weighted: RewardColumn[] = [];

  for (const p of Object.values(plushies)) {
    for (let i = 0; i < p.weight; i++) {
      weighted.push(p.key);
    }
  }

  return weighted;
};

export const getWeightedRandomPlushie = (plushies: PlushiesMap): RewardColumn => {
  const weighted = buildWeightedPlushieList(plushies);
  const i = Math.floor(Math.random() * weighted.length);
  return weighted[i];
};
