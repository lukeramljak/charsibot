import type { BlindBoxConfig, CollectionType, RewardColumn } from './types';

export const coobubuConfig: BlindBoxConfig = {
  collectionType: 'coobubu',
  rewardTitle: 'Cooper Series Blind Box',
  plushies: [
    {
      key: 'reward1',
      name: 'Cutey',
      weight: 12,
    },
    {
      key: 'reward2',
      name: 'Blueberry',
      weight: 12,
    },
    {
      key: 'reward3',
      name: 'Lemony',
      weight: 12,
    },
    {
      key: 'reward4',
      name: 'Bibi',
      weight: 12,
    },
    {
      key: 'reward5',
      name: 'Pinky',
      weight: 12,
    },
    {
      key: 'reward6',
      name: 'Minty',
      weight: 12,
    },
    {
      key: 'reward7',
      name: 'Cherry',
      weight: 12,
    },
    {
      key: 'reward8',
      name: 'Secret',
      weight: 1,
    },
  ],
};

export const olliepopsConfig: BlindBoxConfig = {
  collectionType: 'olliepop',
  rewardTitle: 'Ollie Series Blind Box',
  plushies: [
    {
      key: 'reward1',
      name: 'Berry',
      weight: 12,
    },
    {
      key: 'reward2',
      name: 'Tangerine',
      weight: 12,
    },
    {
      key: 'reward3',
      name: 'Bibble',
      weight: 12,
    },
    {
      key: 'reward4',
      name: 'Kiwi',
      weight: 12,
    },
    {
      key: 'reward5',
      name: 'Crunchy',
      weight: 12,
    },
    {
      key: 'reward6',
      name: 'Caramel',
      weight: 12,
    },
    {
      key: 'reward7',
      name: 'Grape',
      weight: 12,
    },
    {
      key: 'reward8',
      name: 'Secret',
      weight: 1,
    },
  ],
};

export const blindBoxConfigs: Record<CollectionType, BlindBoxConfig> = {
  coobubu: coobubuConfig,
  olliepop: olliepopsConfig,
} as const;

/**
 * Helper to validate that all 8 reward slots are filled
 */
const validateConfig = (config: BlindBoxConfig): void => {
  const keys = config.plushies.map((p) => p.key);
  const expectedKeys: RewardColumn[] = [
    'reward1',
    'reward2',
    'reward3',
    'reward4',
    'reward5',
    'reward6',
    'reward7',
    'reward8',
  ];

  for (const expected of expectedKeys) {
    if (!keys.includes(expected)) {
      throw new Error(`Config for ${config.collectionType} is missing ${expected}`);
    }
  }

  if (keys.length !== 8) {
    throw new Error(
      `Config for ${config.collectionType} must have exactly 8 plushies, got ${keys.length}`,
    );
  }
};

// Validate configs at module load time
validateConfig(coobubuConfig);
validateConfig(olliepopsConfig);
