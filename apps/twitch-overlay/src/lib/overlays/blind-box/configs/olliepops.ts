import type { BlindBoxOverlayConfig } from '../types';

export const olliepopsConfig: BlindBoxOverlayConfig = {
  collectionType: 'olliepops',
  collectionName: 'Olliepops',

  boxFrontFace: '/olliepops/box-front.png',
  boxSideFace: '/olliepops/box-side.png',
  emptyPlushieImage: '/olliepops/empty-slot.png',
  revealSound: '/olliepops/reveal.mp3',

  displayColor: '#ff8c82',
  textColor: '#ffffff',

  audioVolume: 50,

  plushies: [
    { key: 'reward1', name: 'Berry', image: '/olliepops/berry.png' },
    { key: 'reward2', name: 'Tangerine', image: '/olliepops/tangerine.png' },
    { key: 'reward3', name: 'Bibble', image: '/olliepops/bibble.png' },
    { key: 'reward4', name: 'Kiwi', image: '/olliepops/kiwi.png' },
    { key: 'reward5', name: 'Crunchy', image: '/olliepops/crunchy.png' },
    { key: 'reward6', name: 'Caramel', image: '/olliepops/caramel.png' },
    { key: 'reward7', name: 'Grape', image: '/olliepops/grape.png' },
    { key: 'reward8', name: 'Secret', image: '/olliepops/secret.png' }
  ]
};
