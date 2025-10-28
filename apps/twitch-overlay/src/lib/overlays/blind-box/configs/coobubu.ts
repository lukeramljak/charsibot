import type { BlindBoxOverlayConfig } from '../types';

export const coobubuConfig: BlindBoxOverlayConfig = {
  collectionType: 'coobubu',
  collectionName: 'Coobubus',

  boxFrontFace: '/coobubu/box-front.png',
  boxSideFace: '/coobubu/box-side.png',
  emptyPlushieImage: '/coobubu/empty-slot.png',
  revealSound: '/coobubu/reveal.mp3',

  displayColor: '#ff8c82',
  textColor: '#ffffff',

  audioVolume: 50,

  plushies: [
    { key: 'reward1', name: 'Cutey', image: '/coobubu/cutey.png' },
    { key: 'reward2', name: 'Blueberry', image: '/coobubu/blueberry.png' },
    { key: 'reward3', name: 'Lemony', image: '/coobubu/lemony.png' },
    { key: 'reward4', name: 'Bibi', image: '/coobubu/bibi.png' },
    { key: 'reward5', name: 'Pinky', image: '/coobubu/pinky.png' },
    { key: 'reward6', name: 'Minty', image: '/coobubu/minty.png' },
    { key: 'reward7', name: 'Cherry', image: '/coobubu/cherry.png' },
    { key: 'reward8', name: 'Secret', image: '/coobubu/secret.png' }
  ]
};
