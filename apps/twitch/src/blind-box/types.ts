export type CollectionType = 'coobubu' | 'olliepop';
export type BlindBoxRedemptionTitle = 'Cooper Series Blind Box' | 'Ollie Series Blind Box';

export type RewardColumn =
  | 'reward1'
  | 'reward2'
  | 'reward3'
  | 'reward4'
  | 'reward5'
  | 'reward6'
  | 'reward7'
  | 'reward8';

export interface PlushieData {
  /** Unique identifier for the plushie */
  key: RewardColumn;
  /** Display name of the plushie */
  name: string;
  /** Rarity weight (higher = more common) */
  weight: number;
}

export interface BlindBoxConfig {
  /** Collection type identifier for database storage */
  collectionType: CollectionType;
  /** Channel point reward name in Twitch */
  rewardTitle: BlindBoxRedemptionTitle;
  /** Array of plushies available in this blind box */
  plushies: PlushieData[];
}
