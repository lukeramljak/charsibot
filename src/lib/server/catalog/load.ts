import type { BlindBoxSeries, Catalog, Plushie, StatDefinition } from '$lib/contracts/catalog';
import { posix } from 'node:path';
import * as v from 'valibot';

import statsJSON from '$catalog/config/stats.json?raw';

const seriesJSON = import.meta.glob<string>('$catalog/config/blind-box/*.json', {
  eager: true,
  import: 'default',
  query: '?raw',
});

const statSchema = v.strictObject({
  name: v.string(),
  shortName: v.string(),
  longName: v.string(),
  defaultValue: v.optional(v.number(), 0),
  sortOrder: v.optional(v.number(), 0),
  emoji: v.optional(v.string(), ''),
});

const plushieSchema = v.strictObject({
  key: v.string(),
  sortOrder: v.number(),
  weight: v.number(),
  name: v.optional(v.string(), ''),
  image: v.optional(v.string(), ''),
  emptyImage: v.optional(v.string(), ''),
});

const seriesSchema = v.strictObject({
  series: v.string(),
  assetDir: v.optional(v.string(), ''),
  redemptionTitle: v.string(),
  name: v.string(),
  revealSound: v.optional(v.string(), ''),
  boxFrontFace: v.optional(v.string(), ''),
  boxSideFace: v.optional(v.string(), ''),
  displayColor: v.optional(v.string(), ''),
  textColor: v.optional(v.string(), ''),
  plushies: v.array(plushieSchema),
});

const statsSchema = v.array(statSchema);

export interface CatalogSources {
  stats: string;
  series: Readonly<Record<string, string>>;
}

const compareStrings = (left: string, right: string): number => {
  if (left < right) {
    return -1;
  }

  if (left > right) {
    return 1;
  }

  return 0;
};

const decodeJSON = (name: string, source: string): unknown => {
  try {
    return JSON.parse(source) as unknown;
  } catch (error) {
    throw new Error(`decode ${name}: expected a single valid JSON value`, { cause: error });
  }
};

const parseStats = (name: string, source: string): StatDefinition[] => {
  const result = v.safeParse(statsSchema, decodeJSON(name, source));
  if (!result.success) {
    throw new Error(`decode ${name}: ${result.issues.map((issue) => issue.message).join('; ')}`);
  }

  if (result.output.length === 0) {
    throw new Error('at least one stat is required');
  }

  const names = new Set<string>();
  const sortOrders = new Set<number>();

  const definitions = result.output.map((stat) => {
    if (!stat.name.trim() || !stat.shortName.trim() || !stat.longName.trim()) {
      throw new Error('stat name, shortName, and longName are required');
    }

    if (!Number.isSafeInteger(stat.defaultValue)) {
      throw new Error(`stat ${JSON.stringify(stat.name)} defaultValue must be a safe integer`);
    }

    if (!Number.isSafeInteger(stat.sortOrder)) {
      throw new Error(`stat ${JSON.stringify(stat.name)} sortOrder must be a safe integer`);
    }

    if (names.has(stat.name)) {
      throw new Error(`duplicate stat ${JSON.stringify(stat.name)}`);
    }

    if (sortOrders.has(stat.sortOrder)) {
      throw new Error(`duplicate stat sortOrder ${stat.sortOrder}`);
    }

    names.add(stat.name);
    sortOrders.add(stat.sortOrder);

    return { ...stat };
  });

  return definitions.sort((left, right) => left.sortOrder - right.sortOrder);
};

const assetURL = (assetDirectory: string, filename: string): string => {
  if (filename.startsWith('/')) {
    return filename;
  }

  return posix.join('/assets/blind-box', assetDirectory, filename);
};

const parseSeries = (name: string, source: string): BlindBoxSeries => {
  const result = v.safeParse(seriesSchema, decodeJSON(name, source));
  if (!result.success) {
    throw new Error(`decode ${name}: ${result.issues.map((issue) => issue.message).join('; ')}`);
  }

  const raw = result.output;

  if (!raw.series.trim()) {
    throw new Error('series is required');
  }

  if (!raw.redemptionTitle.trim() || !raw.name.trim()) {
    throw new Error('redemptionTitle and name are required');
  }

  if (raw.plushies.length === 0) {
    throw new Error('at least one plushie is required');
  }

  const keys = new Set<string>();
  const sortOrders = new Set<number>();
  const assetDirectory = raw.assetDir || raw.series;

  const plushies: Plushie[] = raw.plushies.map((plushie) => {
    if (!plushie.key.trim()) {
      throw new Error('plushie key is required');
    }

    if (!Number.isSafeInteger(plushie.sortOrder)) {
      throw new Error(`plushie ${JSON.stringify(plushie.key)} sortOrder must be a safe integer`);
    }

    if (!Number.isSafeInteger(plushie.weight) || plushie.weight <= 0) {
      throw new Error(`plushie ${JSON.stringify(plushie.key)} must have a positive safe weight`);
    }

    if (keys.has(plushie.key)) {
      throw new Error(`duplicate plushie ${JSON.stringify(plushie.key)}`);
    }

    if (sortOrders.has(plushie.sortOrder)) {
      throw new Error(`duplicate plushie sortOrder ${plushie.sortOrder}`);
    }

    keys.add(plushie.key);
    sortOrders.add(plushie.sortOrder);

    return {
      series: raw.series,
      key: plushie.key,
      sortOrder: plushie.sortOrder,
      weight: plushie.weight,
      name: plushie.name,
      image: assetURL(assetDirectory, plushie.image),
      emptyImage: assetURL(assetDirectory, plushie.emptyImage),
    };
  });

  plushies.sort((left, right) => left.sortOrder - right.sortOrder);

  return {
    series: raw.series,
    redemptionTitle: raw.redemptionTitle,
    name: raw.name,
    revealSound: assetURL(assetDirectory, raw.revealSound),
    boxFrontFace: assetURL(assetDirectory, raw.boxFrontFace),
    boxSideFace: assetURL(assetDirectory, raw.boxSideFace),
    displayColor: raw.displayColor,
    textColor: raw.textColor,
    plushies,
  };
};

export const loadCatalogFromSources = (sources: CatalogSources): Catalog => {
  const stats = parseStats('stats.json', sources.stats);
  const identifiers = new Set<string>();

  const series = Object.entries(sources.series).map(([name, source]) => {
    const config = parseSeries(name, source);

    if (identifiers.has(config.series)) {
      throw new Error(`duplicate blind-box series ${JSON.stringify(config.series)}`);
    }

    identifiers.add(config.series);

    return config;
  });

  if (series.length === 0) {
    throw new Error('at least one blind-box series is required');
  }

  series.sort((left, right) => compareStrings(left.series, right.series));

  return { stats, series };
};

export const loadCatalog = (): Catalog =>
  loadCatalogFromSources({ stats: statsJSON, series: seriesJSON });
