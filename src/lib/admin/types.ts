import type { ViewerCollection } from '$lib/contracts/collections';
import type { UserStat, Viewer } from '$lib/contracts/viewer';

export interface GrantResult {
  kind: 'stat' | 'plushie';
  statName?: string;
  plushieName?: string;
  isDuplicate?: boolean;
}

export interface AdminUserDetail {
  user: Viewer;
  stats: UserStat[];
  collections: ViewerCollection[];
  grant?: GrantResult;
}

export type ActivityFilter = 'all' | 'unknown' | 'inactive30' | 'inactive90' | 'recent';
