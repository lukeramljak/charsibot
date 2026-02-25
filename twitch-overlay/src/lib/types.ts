export type CollectionType = 'coobubu' | 'olliepop' | 'christmas' | 'valentines';

export type OverlayEvent =
  | ChatCommandEvent
  | RedemptionEvent
  | CollectionDisplayEvent
  | BlindBoxRedemptionEvent
  | LeaderboardEvent;

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
    plushie: string;
    isNew: boolean;
    collectionSize: number;
    collection: string[];
  };
}

export type Stat = 'STR' | 'INT' | 'CHA' | 'LUCK' | 'DEX' | 'PENIS';

export type Leaderboard = Record<
  Stat,
  {
    username: string;
    value: number;
  }
>;

export interface LeaderboardEvent {
  type: 'leaderboard';
  data: Leaderboard;
}
