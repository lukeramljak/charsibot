import type { CollectionType } from '$lib/types';

export interface PlushieData {
  /** Unique identifier for the plushie */
  key: string;
  /** Display name of the plushie */
  name: string;
  /** Path to the plushie image asset */
  image: string;
}

export interface BlindBoxOverlayConfig {
  /** Collection type identifier */
  collectionType: CollectionType;

  /** Name of the blind box collection to be displayed (e.g. "Coobubus") */
  collectionName: string;

  revealSound: string;

  boxFrontFace: string;
  boxSideFace: string;
  emptyPlushieImage: string;

  displayColor: string;
  textColor: string;

  /** Array of plushies available in this blind box */
  plushies: PlushieData[];
}
