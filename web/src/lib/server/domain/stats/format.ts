import type { UserStat } from '$lib/contracts/viewer';

export const formatStats = (username: string, stats: readonly UserStat[]): string => {
  if (stats.length === 0) {
    return `No stats found for ${username}`;
  }

  const values = stats.map((stat) => `${stat.shortName}: ${stat.value}`).join(' | ');

  return `${username}'s stats: ${values}`;
};
