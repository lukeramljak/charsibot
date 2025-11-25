import type { BlindBoxOverlayConfig } from './types';

export const blindBoxConfigs: BlindBoxOverlayConfig[] = [
  {
    collectionType: 'coobubu',
    collectionName: 'Coobubus',

    boxFrontFace: '/coobubu/box-front.png',
    boxSideFace: '/coobubu/box-side.png',
    emptyPlushieImage: '/coobubu/empty-slot.png',

    displayColor: '#ff8c82',
    textColor: '#ffffff',

    plushies: [
      { key: 'reward1', name: 'Cutey', image: '/coobubu/cutey.png' },
      { key: 'reward2', name: 'Blueberry', image: '/coobubu/blueberry.png' },
      { key: 'reward3', name: 'Lemony', image: '/coobubu/lemony.png' },
      { key: 'reward4', name: 'Bibi', image: '/coobubu/bibi.png' },
      { key: 'reward5', name: 'Pinky', image: '/coobubu/pinky.png' },
      { key: 'reward6', name: 'Minty', image: '/coobubu/minty.png' },
      { key: 'reward7', name: 'Cherry', image: '/coobubu/cherry.png' },
      { key: 'reward8', name: 'Secret', image: '/coobubu/secret.png' },
    ],
  },
  {
    collectionType: 'olliepop',
    collectionName: 'Olliepops',

    boxFrontFace: '/olliepops/box-front.png',
    boxSideFace: '/olliepops/box-side.png',
    emptyPlushieImage: '/olliepops/empty-slot.png',

    displayColor: '#ff8c82',
    textColor: '#ffffff',

    plushies: [
      { key: 'reward1', name: 'Berry', image: '/olliepops/berry.png' },
      { key: 'reward2', name: 'Tangerine', image: '/olliepops/tangerine.png' },
      { key: 'reward3', name: 'Bibble', image: '/olliepops/bibble.png' },
      { key: 'reward4', name: 'Kiwi', image: '/olliepops/kiwi.png' },
      { key: 'reward5', name: 'Crunchy', image: '/olliepops/crunchy.png' },
      { key: 'reward6', name: 'Caramel', image: '/olliepops/caramel.png' },
      { key: 'reward7', name: 'Grape', image: '/olliepops/grape.png' },
      { key: 'reward8', name: 'Secret', image: '/olliepops/secret.png' },
    ],
  },
];
