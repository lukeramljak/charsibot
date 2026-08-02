import type { BlindBoxSeries } from '$lib/contracts/catalog';

export interface ViewerCollection {
  config: BlindBoxSeries;
  collected: string[];
}
