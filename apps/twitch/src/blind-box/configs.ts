import type { BlindBoxConfig } from '@/blind-box/types';

export const blindBoxConfigs: BlindBoxConfig[] = [
  {
    collectionType: 'coobubu',
    moderatorCommand: 'coobubu-redeem',
    collectionDisplayCommand: 'coobubu',
    rewardTitle: 'Cooper Series Blind Box',
    plushies: {
      reward1: { key: 'reward1', weight: 12 },
      reward2: { key: 'reward2', weight: 12 },
      reward3: { key: 'reward3', weight: 12 },
      reward4: { key: 'reward4', weight: 12 },
      reward5: { key: 'reward5', weight: 12 },
      reward6: { key: 'reward6', weight: 12 },
      reward7: { key: 'reward7', weight: 12 },
      reward8: { key: 'reward8', weight: 1 },
    },
  },
  {
    collectionType: 'olliepop',
    moderatorCommand: 'olliepop-redeem',
    collectionDisplayCommand: 'olliepop',
    rewardTitle: 'Ollie Series Blind Box',
    plushies: {
      reward1: { key: 'reward1', weight: 12 },
      reward2: { key: 'reward2', weight: 12 },
      reward3: { key: 'reward3', weight: 12 },
      reward4: { key: 'reward4', weight: 12 },
      reward5: { key: 'reward5', weight: 12 },
      reward6: { key: 'reward6', weight: 12 },
      reward7: { key: 'reward7', weight: 12 },
      reward8: { key: 'reward8', weight: 1 },
    },
  },
  {
    collectionType: 'christmas',
    moderatorCommand: 'xmas-redeem',
    collectionDisplayCommand: 'xmas',
    rewardTitle: 'Christmas Series Blind Box',
    plushies: {
      reward1: { key: 'reward1', weight: 3 },
      reward2: { key: 'reward2', weight: 3 },
      reward3: { key: 'reward3', weight: 3 },
      reward4: { key: 'reward4', weight: 3 },
      reward5: { key: 'reward5', weight: 3 },
      reward6: { key: 'reward6', weight: 3 },
      reward7: { key: 'reward7', weight: 3 },
      reward8: { key: 'reward8', weight: 1 },
    },
  },
];
