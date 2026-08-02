import type { Random } from '$lib/server/application/ports';
import { describe, expect, it } from 'vitest';

import { loadCatalog } from '$lib/server/catalog/load';
import { openDatabase } from '$lib/server/db/client';
import { createCollectionsRepository } from '$lib/server/db/collections.repository';

import { pickWeightedPlushie } from '$lib/server/domain/blind-box/random';
import { createBlindBoxService } from '$lib/server/domain/blind-box/service';

const setup = () => {
  const connection = openDatabase(':memory:');
  const catalog = loadCatalog();
  const repository = createCollectionsRepository(connection.database);
  const service = createBlindBoxService({ repository, series: catalog.series });
  return { connection, catalog, repository, service };
};

describe('blind-box service', () => {
  it('grants in catalog order and synchronizes the username on duplicates', async () => {
    const { connection, service } = setup();
    expect(await service.grant('viewer', 'oldname', 'coobubu', 'blueberry')).toEqual({
      isNew: true,
      collection: ['blueberry'],
    });
    expect(await service.grant('viewer', 'oldname', 'coobubu', 'cutey')).toEqual({
      isNew: true,
      collection: ['cutey', 'blueberry'],
    });
    expect(await service.grant('viewer', 'newname', 'coobubu', 'cutey')).toEqual({
      isNew: false,
      collection: ['cutey', 'blueberry'],
    });
    expect(
      connection.database
        .prepare(
          "SELECT username FROM user_plushies WHERE user_id = 'viewer' AND series = 'coobubu' AND key = 'cutey'",
        )
        .pluck()
        .get(),
    ).toBe('newname');
    connection.close();
  });

  it('returns every configured collection and supports remove and reset', async () => {
    const { connection, service } = setup();
    await service.grant('viewer', 'viewer', 'coobubu', 'cutey');
    const collections = await service.getViewerCollections('viewer');
    expect(collections).toHaveLength(5);
    expect(collections.find((entry) => entry.config.series === 'coobubu')?.collected).toEqual([
      'cutey',
    ]);
    await service.remove('viewer', 'coobubu', 'cutey');
    expect(await service.getCollection('viewer', 'coobubu')).toEqual([]);
    await service.grant('viewer', 'viewer', 'coobubu', 'blueberry');
    await service.reset('viewer', 'coobubu');
    expect(await service.getCollection('viewer', 'coobubu')).toEqual([]);
    connection.close();
  });

  it('groups completed collections by stable user ID and sorts output', async () => {
    const { connection, catalog, repository, service } = setup();
    const config = catalog.series.find((candidate) => candidate.series === 'coobubu');

    if (!config) {
      throw new Error('missing coobubu fixture');
    }

    for (const [userID, username] of [
      ['viewer-b', 'dave'],
      ['viewer-a', 'carol'],
    ]) {
      config.plushies.forEach((plushie, index) => {
        repository.grant(
          userID,
          index < 4 ? `old-${username}` : username,
          config.series,
          plushie.key,
        );
      });
    }
    expect(await service.completed()).toEqual([
      { seriesName: 'Coobubus', usernames: 'old-carol, old-dave' },
    ]);
    connection.close();
  });

  it('does not count partial or over-complete unknown-key collections', async () => {
    const { connection, catalog, repository, service } = setup();
    const config = catalog.series.find((candidate) => candidate.series === 'coobubu');

    if (!config) {
      throw new Error('missing coobubu fixture');
    }

    config.plushies.forEach((plushie) =>
      repository.grant('viewer', 'viewer', config.series, plushie.key),
    );
    repository.grant('viewer', 'viewer', config.series, 'unknown');
    expect(await service.completed()).toEqual([]);
    connection.close();
  });
});

describe('pickWeightedPlushie', () => {
  const plushies = [
    {
      series: 'test',
      key: 'first',
      sortOrder: 1,
      weight: 2,
      name: 'First',
      image: '',
      emptyImage: '',
    },
    {
      series: 'test',
      key: 'second',
      sortOrder: 2,
      weight: 3,
      name: 'Second',
      image: '',
      emptyImage: '',
    },
  ];

  it.each([
    [0, 'first'],
    [1, 'first'],
    [2, 'second'],
    [4, 'second'],
  ])('maps exact integer boundary %i to %s', (selection, expected) => {
    const random: Random = { integer: () => selection };
    expect(pickWeightedPlushie(plushies, random).key).toBe(expected);
  });

  it('rejects invalid random output and empty positive weight', () => {
    expect(() => pickWeightedPlushie(plushies, { integer: () => 5 })).toThrow(/outside/);
    expect(() =>
      pickWeightedPlushie(
        plushies.map((plushie) => ({ ...plushie, weight: 0 })),
        { integer: () => 0 },
      ),
    ).toThrow(/no plushies/);
  });

  it('preserves the production coobubu secret boundary of one in twenty-four', () => {
    const config = loadCatalog().series.find((candidate) => candidate.series === 'coobubu');

    if (!config) {
      throw new Error('missing coobubu fixture');
    }

    expect(pickWeightedPlushie(config.plushies, { integer: () => 160 }).key).toBe('cherry');
    expect(pickWeightedPlushie(config.plushies, { integer: () => 161 }).key).toBe('secret');
    expect(pickWeightedPlushie(config.plushies, { integer: () => 167 }).key).toBe('secret');
  });
});
