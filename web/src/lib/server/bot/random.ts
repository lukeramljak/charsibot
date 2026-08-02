import type { Random } from '$lib/server/application/ports';

export const randomInteger = (random: Random, maxExclusive: number, label: string): number => {
  const value = random.integer(maxExclusive);

  if (!Number.isSafeInteger(value) || value < 0 || value >= maxExclusive) {
    throw new RangeError(`${label} random integer is outside the requested range`);
  }

  return value;
};
