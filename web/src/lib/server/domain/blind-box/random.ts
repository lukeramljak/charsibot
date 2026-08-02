import type { Plushie } from '$lib/contracts/catalog';
import type { Random } from '$lib/server/application/ports';

export const pickWeightedPlushie = (plushies: readonly Plushie[], random: Random): Plushie => {
  let totalWeight = 0;

  for (const plushie of plushies) {
    if (!Number.isSafeInteger(plushie.weight) || plushie.weight <= 0) {
      continue;
    }

    totalWeight += plushie.weight;

    if (!Number.isSafeInteger(totalWeight)) {
      throw new RangeError('total plushie weight exceeds JavaScript safe integer range');
    }
  }

  if (totalWeight === 0) {
    throw new Error('no plushies with positive weight');
  }

  const selection = random.integer(totalWeight);

  if (!Number.isSafeInteger(selection) || selection < 0 || selection >= totalWeight) {
    throw new RangeError('random integer is outside the requested range');
  }

  let boundary = 0;

  for (const plushie of plushies) {
    if (!Number.isSafeInteger(plushie.weight) || plushie.weight <= 0) {
      continue;
    }

    boundary += plushie.weight;

    if (selection < boundary) {
      return plushie;
    }
  }

  throw new Error('weighted plushie selection failed');
};
