import { describe, expect, it } from 'vitest';

import { loadCatalog, loadCatalogFromSources, type CatalogSources } from '$lib/server/catalog/load';

const stat = {
  name: 'strength',
  shortName: 'STR',
  longName: 'Strength',
  defaultValue: 3,
  sortOrder: 1,
  emoji: 'strong',
};

const series = {
  series: 'test',
  redemptionTitle: 'Test box',
  name: 'Tests',
  revealSound: 'reveal.mp3',
  boxFrontFace: '/absolute/front.png',
  boxSideFace: 'side.png',
  displayColor: '#000000',
  textColor: '#ffffff',
  plushies: [
    {
      key: 'one',
      sortOrder: 1,
      weight: 1,
      name: 'One',
      image: 'one.png',
      emptyImage: 'empty.png',
    },
  ],
};

const sources = (overrides: Partial<CatalogSources> = {}): CatalogSources => ({
  stats: JSON.stringify([stat]),
  series: { 'test.json': JSON.stringify(series) },
  ...overrides,
});

describe('catalog loading', () => {
  it('loads the production catalog in deterministic order with expanded assets', () => {
    const catalog = loadCatalog();
    expect(catalog.stats).toHaveLength(6);
    expect(catalog.stats.map((definition) => definition.name)).toEqual([
      'strength',
      'intelligence',
      'charisma',
      'luck',
      'dexterity',
      'penis',
    ]);
    expect(catalog.series.map((config) => config.series)).toEqual([
      'coobubu',
      'easter',
      'halloween',
      'olliepop',
      'pixel',
      'valentines',
      'xmas',
    ]);
    expect(catalog.series.find((config) => config.series === 'olliepop')?.revealSound).toBe(
      '/assets/blind-box/olliepops/reveal.mp3',
    );
    const pixel = catalog.series.find((config) => config.series === 'pixel');
    expect(pixel).toMatchObject({
      name: 'Pixel Pups',
      revealSound: '/assets/blind-box/pixel/reveal.mp3',
    });
    expect(pixel?.plushies.map(({ key, image }) => ({ key, image }))).toEqual([
      { key: 'ollie-of-rivia', image: '/assets/blind-box/pixel/ollie-of-rivia.png' },
      { key: 'cooper-silverpaw', image: '/assets/blind-box/pixel/cooper-silverpaw.png' },
      { key: 'lethal-coopany', image: '/assets/blind-box/pixel/lethal-coopany.png' },
      { key: 'cult-of-ollie', image: '/assets/blind-box/pixel/cult-of-ollie.png' },
      { key: 'legend-of-cooper', image: '/assets/blind-box/pixel/legend-of-cooper.png' },
      { key: 'cuber', image: '/assets/blind-box/pixel/cuber.png' },
      { key: 'repollie', image: '/assets/blind-box/pixel/repollie.png' },
      { key: 'secret', image: '/assets/blind-box/pixel/secret.png' },
    ]);
  });

  it('preserves absolute asset paths and defaults assetDir to series', () => {
    const catalog = loadCatalogFromSources(sources());
    expect(catalog.series[0].boxFrontFace).toBe('/absolute/front.png');
    expect(catalog.series[0].boxSideFace).toBe('/assets/blind-box/test/side.png');
    expect(catalog.series[0].plushies[0].image).toBe('/assets/blind-box/test/one.png');
  });

  it('rejects unknown fields and trailing JSON values', () => {
    expect(() =>
      loadCatalogFromSources(sources({ stats: JSON.stringify([{ ...stat, unexpected: true }]) })),
    ).toThrow();
    expect(() =>
      loadCatalogFromSources(sources({ stats: `${JSON.stringify([stat])}\n{}` })),
    ).toThrow(/single valid JSON value/);
  });

  it('rejects empty catalogs, duplicate identifiers, and invalid weights', () => {
    expect(() => loadCatalogFromSources(sources({ stats: '[]' }))).toThrow(/at least one stat/);
    expect(() => loadCatalogFromSources(sources({ series: {} }))).toThrow(/at least one blind-box/);
    expect(() =>
      loadCatalogFromSources(
        sources({
          series: {
            'a.json': JSON.stringify(series),
            'b.json': JSON.stringify(series),
          },
        }),
      ),
    ).toThrow(/duplicate blind-box series/);
    expect(() =>
      loadCatalogFromSources(
        sources({
          series: {
            'test.json': JSON.stringify({
              ...series,
              plushies: [{ ...series.plushies[0], weight: 0 }],
            }),
          },
        }),
      ),
    ).toThrow(/positive safe weight/);
  });

  it('rejects duplicate stat and plushie sort orders', () => {
    expect(() =>
      loadCatalogFromSources(
        sources({
          stats: JSON.stringify([stat, { ...stat, name: 'luck' }]),
        }),
      ),
    ).toThrow(/duplicate stat sortOrder/);
    expect(() =>
      loadCatalogFromSources(
        sources({
          series: {
            'test.json': JSON.stringify({
              ...series,
              plushies: [series.plushies[0], { ...series.plushies[0], key: 'two' }],
            }),
          },
        }),
      ),
    ).toThrow(/duplicate plushie sortOrder/);
  });
});
