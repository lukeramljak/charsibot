import type { BlindBoxOverlayConfig } from './types';

export const blindBoxConfigs: BlindBoxOverlayConfig[] = [
  {
    collectionType: 'coobubu',
    collectionName: 'Coobubus',

    revealSound: '/blind-box/reveal-default.mp3',

    boxFrontFace: '/blind-box/coobubu/box-front.png',
    boxSideFace: '/blind-box/coobubu/box-side.png',
    emptyPlushieImage: '/blind-box/coobubu/empty-slot.png',

    displayColor: '#ff8c82',
    textColor: '#ffffff',

    plushies: [
      { key: 'reward1', name: 'Cutey', image: '/blind-box/coobubu/cutey.png' },
      { key: 'reward2', name: 'Blueberry', image: '/blind-box/coobubu/blueberry.png' },
      { key: 'reward3', name: 'Lemony', image: '/blind-box/coobubu/lemony.png' },
      { key: 'reward4', name: 'Bibi', image: '/blind-box/coobubu/bibi.png' },
      { key: 'reward5', name: 'Pinky', image: '/blind-box/coobubu/pinky.png' },
      { key: 'reward6', name: 'Minty', image: '/blind-box/coobubu/minty.png' },
      { key: 'reward7', name: 'Cherry', image: '/blind-box/coobubu/cherry.png' },
      { key: 'reward8', name: 'Secret', image: '/blind-box/coobubu/secret.png' },
    ],
  },
  {
    collectionType: 'olliepop',
    collectionName: 'Olliepops',

    revealSound: '/blind-box/reveal-default.mp3',

    boxFrontFace: '/blind-box/olliepops/box-front.png',
    boxSideFace: '/blind-box/olliepops/box-side.png',
    emptyPlushieImage: '/blind-box/olliepops/empty-slot.png',

    displayColor: '#ff8c82',
    textColor: '#ffffff',

    plushies: [
      { key: 'reward1', name: 'Berry', image: '/blind-box/olliepops/berry.png' },
      { key: 'reward2', name: 'Tangerine', image: '/blind-box/olliepops/tangerine.png' },
      { key: 'reward3', name: 'Bibble', image: '/blind-box/olliepops/bibble.png' },
      { key: 'reward4', name: 'Kiwi', image: '/blind-box/olliepops/kiwi.png' },
      { key: 'reward5', name: 'Crunchy', image: '/blind-box/olliepops/crunchy.png' },
      { key: 'reward6', name: 'Caramel', image: '/blind-box/olliepops/caramel.png' },
      { key: 'reward7', name: 'Grape', image: '/blind-box/olliepops/grape.png' },
      { key: 'reward8', name: 'Secret', image: '/blind-box/olliepops/secret.png' },
    ],
  },
];
