import type { BlindBoxConfig } from '@/blind-box/types';

export const blindBoxConfigs: BlindBoxConfig[] = [
  {
    enabled: true,
    collectionType: 'coobubu',
    moderatorCommand: 'coobubu-redeem',
    collectionDisplayCommand: 'coobubu',
    rewardTitle: 'Cooper Series Blind Box',
    plushies: {
      reward1: { key: 'reward1', name: 'Cutey', weight: 12 },
      reward2: { key: 'reward2', name: 'Blueberry', weight: 12 },
      reward3: { key: 'reward3', name: 'Lemony', weight: 12 },
      reward4: { key: 'reward4', name: 'Bibi', weight: 12 },
      reward5: { key: 'reward5', name: 'Pinky', weight: 12 },
      reward6: { key: 'reward6', name: 'Minty', weight: 12 },
      reward7: { key: 'reward7', name: 'Cherry', weight: 12 },
      reward8: { key: 'reward8', name: 'Secret', weight: 1 },
    },
  },
  {
    enabled: true,
    collectionType: 'olliepop',
    moderatorCommand: 'olliepop-redeem',
    collectionDisplayCommand: 'olliepop',
    rewardTitle: 'Ollie Series Blind Box',
    plushies: {
      reward1: { key: 'reward1', name: 'Berry', weight: 12 },
      reward2: { key: 'reward2', name: 'Tangerine', weight: 12 },
      reward3: { key: 'reward3', name: 'Bibble', weight: 12 },
      reward4: { key: 'reward4', name: 'Kiwi', weight: 12 },
      reward5: { key: 'reward5', name: 'Crunchy', weight: 12 },
      reward6: { key: 'reward6', name: 'Caramel', weight: 12 },
      reward7: { key: 'reward7', name: 'Grape', weight: 12 },
      reward8: { key: 'reward8', name: 'Secret', weight: 1 },
    },
  },
];
