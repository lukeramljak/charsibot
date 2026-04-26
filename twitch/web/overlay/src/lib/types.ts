export type OverlayEvent =
  | ChatCommandEvent
  | CollectionDisplayEvent
  | BlindBoxRedemptionEvent;

export interface ChatCommandEvent {
  type: 'chat_command';
  message: string;
}

export interface CollectionDisplayEvent {
  type: 'collection_display';
  data: {
    username: string;
    series: string;
    collection: string[];
  };
}

export interface BlindBoxRedemptionEvent {
  type: 'blindbox_redemption';
  data: {
    username: string;
    series: string;
    plushie: string;
    isNew: boolean;
    collection: string[];
  };
}
