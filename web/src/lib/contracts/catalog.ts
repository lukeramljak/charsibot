export interface StatDefinition {
  name: string;
  shortName: string;
  longName: string;
  defaultValue: number;
  sortOrder: number;
  emoji: string;
}

export interface Plushie {
  series: string;
  key: string;
  sortOrder: number;
  weight: number;
  name: string;
  image: string;
  emptyImage: string;
}

export interface BlindBoxSeries {
  series: string;
  redemptionTitle: string;
  name: string;
  revealSound: string;
  boxFrontFace: string;
  boxSideFace: string;
  displayColor: string;
  textColor: string;
  plushies: Plushie[];
}

export interface Catalog {
  stats: StatDefinition[];
  series: BlindBoxSeries[];
}
