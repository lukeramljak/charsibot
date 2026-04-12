export interface PlushieData {
  /** Unique identifier for the plushie (matches key from DB) */
  key: string;
  name: string;
  /** Path to the plushie image asset */
  image: string;
  emptyImage: string;
}

export interface BlindBoxOverlayConfig {
  /** Series identifier (e.g. 'coobubu', 'xmas') */
  series: string;

  /** Display name of the blind box collection (e.g. "Coobubus") */
  name: string;

  revealSound: string;

  boxFrontFace: string;
  boxSideFace: string;

  displayColor: string;
  textColor: string;

  /** Array of plushies available in this blind box */
  plushies: PlushieData[];
}
