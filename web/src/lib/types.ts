export interface PlushieData {
  id: number;
  series: string;
  /** Unique identifier for the plushie (matches key from DB) */
  key: string;
  sortOrder: number;
  weight: number;
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

  redemptionTitle: string;

  revealSound: string;

  boxFrontFace: string;
  boxSideFace: string;

  displayColor: string;
  textColor: string;

  /** Array of plushies available in this blind box */
  plushies: PlushieData[];
}

export type OverlayEvent = ChatCommandEvent | CollectionDisplayEvent | BlindBoxRedemptionEvent;

export type OverlayEventType = OverlayEvent['type'];

export interface ChatCommandEvent {
  type: 'chat_command';
  message: string;
}

export interface CollectionDisplayEvent {
  type: 'blindbox_display';
  username: string;
  collection: string[];
  config: BlindBoxOverlayConfig;
}

export interface BlindBoxRedemptionEvent {
  type: 'blindbox_redemption';
  username: string;
  plushie: PlushieData;
  isNew: boolean;
  collection: string[];
  config: BlindBoxOverlayConfig;
}
