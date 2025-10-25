import type { BlindBoxOverlayConfig } from '../types';
import { coobubuConfig as sharedConfig } from '@charsibot/shared/blind-box-configs';

export const coobubuConfig: BlindBoxOverlayConfig = {
  ...sharedConfig,
  collectionName: 'Coobubus',

  boxFrontFace: '/coobubu/box-front.png',
  boxSideFace: '/coobubu/box-side.png',
  emptyPlushieImage: '/coobubu/empty-slot.png',
  revealSound: '/coobubu/reveal.mp3',

  displayColor: '#ff8c82',
  textColor: '#ffffff',

  audioVolume: 50,

  plushies: sharedConfig.plushies.map((plushie) => ({
    ...plushie,
    image: `/coobubu/${plushie.name.toLowerCase()}.png`
  }))
};
