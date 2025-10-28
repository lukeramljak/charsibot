import type { BlindBoxOverlayConfig } from '../types';
import { olliepopsConfig as sharedConfig } from 'shared/blind-box-configs';

export const olliepopsConfig: BlindBoxOverlayConfig = {
  ...sharedConfig,
  collectionName: 'Olliepops',

  boxFrontFace: '/olliepops/box-front.png',
  boxSideFace: '/olliepops/box-side.png',
  emptyPlushieImage: '/olliepops/empty-slot.png',
  revealSound: '/olliepops/reveal.mp3',

  displayColor: '#ff8c82',
  textColor: '#ffffff',

  audioVolume: 50,

  plushies: sharedConfig.plushies.map((plushie) => ({
    ...plushie,
    image: `/olliepops/${plushie.name.toLowerCase()}.png`
  }))
};
