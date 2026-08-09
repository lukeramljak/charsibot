import type { BlindBoxSeries, Plushie } from '$lib/contracts/catalog';

export const overlayEventTypes = [
  'chat_command',
  'blindbox_display',
  'blindbox_redemption',
] as const;

export type OverlayEventType = (typeof overlayEventTypes)[number];

export interface ChatCommandEvent {
  type: 'chat_command';
  message: string;
}

export interface CollectionDisplayEvent {
  type: 'blindbox_display';
  username: string;
  collection: string[];
  config: BlindBoxSeries;
}

export interface BlindBoxRedemptionEvent {
  type: 'blindbox_redemption';
  username: string;
  plushie: Plushie;
  isNew: boolean;
  collection: string[];
  config: BlindBoxSeries;
}

export type OverlayEvent = ChatCommandEvent | CollectionDisplayEvent | BlindBoxRedemptionEvent;
