export type CollectionType = 'coobubu' | 'olliepops';

export type RewardColumn =
  | 'reward1'
  | 'reward2'
  | 'reward3'
  | 'reward4'
  | 'reward5'
  | 'reward6'
  | 'reward7'
  | 'reward8';

export type OverlayEvent =
  | ChatCommandEvent
  | RedemptionEvent
  | CollectionDisplayEvent
  | BlindBoxRedemptionEvent;

export interface ChatCommandEvent {
  type: 'chat_command';
  message: string;
}

export interface RedemptionEvent {
  type: 'redemption';
  message: string;
}

export interface CollectionDisplayEvent {
  type: 'collection_display';
  data: {
    userId: string;
    username: string;
    collectionType: CollectionType;
    collection: string[];
    collectionSize: number;
  };
}

export interface BlindBoxRedemptionEvent {
  type: 'blindbox_redemption';
  data: {
    userId: string;
    username: string;
    collectionType: CollectionType;
    seriesName: string;
    plushie: {
      key: string;
      name: string;
      weight: number;
    };
    isNew: boolean;
    collectionSize: number;
    collection: string[];
  };
}

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
  rewardTitle: string;
  /** Moderator-only chat command trigger (e.g. "!blindbox") */
  moderatorCommand: string;
  /** Chat command to display collection (e.g. "!collection") */
  collectionCommand: string;
  /** Array of plushies available in this blind box */
  plushies: PlushieData[];
}
