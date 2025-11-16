export type CollectionType = 'coobubu' | 'olliepop';

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
