export type CollectionType = 'coobubu' | 'olliepop' | 'christmas';

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
  /** Rarity weight (higher = more common) */
  weight: number;
}

export type PlushiesMap = {
  [K in RewardColumn]: PlushieData & { key: K };
};

export interface BlindBoxConfig {
  /** Collection type identifier for database storage */
  collectionType: CollectionType;
  /** Channel point reward name in Twitch */
  rewardTitle: string;
  /** Moderator-only chat command trigger (e.g. "blindbox") */
  moderatorCommand: `${string}-redeem`;
  /** Chat command to display collection (e.g. "collection") */
  collectionDisplayCommand: string;
  plushies: PlushiesMap;
}
