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
    series: string;
    collection: string[];
    collectionSize: number;
  };
}

export interface BlindBoxRedemptionEvent {
  type: 'blindbox_redemption';
  data: {
    userId: string;
    username: string;
    series: string;
    seriesName: string;
    plushie: string;
    isNew: boolean;
    collectionSize: number;
    collection: string[];
  };
}

export type Leaderboard = Array<{
  displayName: string;
  username: string;
  value: number;
}>;

export interface LeaderboardEvent {
  type: 'leaderboard';
  data: Leaderboard;
}
