import type { CollectionType } from '@charsibot/shared/types';

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

  boxFrontFace: string;
  boxSideFace: string;
  emptyPlushieImage: string;
  revealSound: string;

  displayColor: string;
  textColor: string;

  /** Audio volume (0-100) */
  audioVolume: number;

  /** Array of plushies available in this blind box */
  plushies: PlushieData[];
}
