import { command, getRequestEvent, query } from '$app/server';
import type { AdminUserDetail, GrantResult } from '$lib/admin/types';
import type { StatDefinition } from '$lib/contracts/catalog';
import { pickWeightedPlushie } from '$lib/server/domain/blind-box/random';
import { formatStats } from '$lib/server/domain/stats/format';
import { getServices } from '$lib/server/runtime/container';
import { error } from '@sveltejs/kit';
import * as v from 'valibot';

const EXPLODED_PENIS_VALUE = -1000;

const isLoopback = (address: string): boolean =>
  address === '127.0.0.1' || address === '::1' || address === '::ffff:127.0.0.1';

const requireLocalAdmin = (): void => {
  const event = getRequestEvent();
  const address = event.getClientAddress();

  if (!isLoopback(address)) {
    throw error(403, 'admin is available only on localhost');
  }
};

const buildUserDetail = async (userID: string, grant?: GrantResult): Promise<AdminUserDetail> => {
  const { stats, blindBox } = getServices();

  const viewer = await stats.getViewer(userID);
  const userStats = await stats.get(userID);
  const collections = await blindBox.getViewerCollections(userID);

  return { user: viewer, stats: userStats, collections, ...(grant ? { grant } : {}) };
};

const sendChatStats = async (userID: string, username: string): Promise<void> => {
  const { stats, chat } = getServices();
  const userStats = await stats.get(userID);

  await chat.send({ message: formatStats(username, userStats) });
};

const findStatDefinition = (name: string): StatDefinition => {
  const { stats } = getServices();
  const definition = stats.definitions.find((d) => d.name === name);

  if (!definition) {
    throw error(503, `${name} stat is unavailable`);
  }

  return definition;
};

// --- Queries ---

export const listViewers = query(async () => {
  requireLocalAdmin();

  const { stats } = getServices();

  return stats.listViewers();
});

export const getViewer = query(
  v.object({ userID: v.pipe(v.string(), v.nonEmpty()) }),
  async ({ userID }) => {
    requireLocalAdmin();

    return buildUserDetail(userID);
  },
);

// --- Commands ---

export const deleteViewer = command(
  v.object({ userID: v.pipe(v.string(), v.nonEmpty()) }),
  async ({ userID }) => {
    requireLocalAdmin();

    const { stats } = getServices();

    await stats.getViewer(userID);
    await stats.deleteViewers([userID]);
  },
);

export const deleteViewers = command(
  v.object({ userIDs: v.pipe(v.array(v.pipe(v.string(), v.nonEmpty())), v.nonEmpty()) }),
  async ({ userIDs }) => {
    requireLocalAdmin();

    const { stats } = getServices();

    await stats.deleteViewers(userIDs);
  },
);

export const updateStat = command(
  v.object({
    userID: v.pipe(v.string(), v.nonEmpty()),
    statName: v.pipe(v.string(), v.nonEmpty()),
    mode: v.picklist(['set', 'adjust']),
    value: v.pipe(v.number(), v.integer()),
  }),
  async ({ userID, statName, mode, value }) => {
    requireLocalAdmin();

    const { stats } = getServices();
    const viewer = await stats.getViewer(userID);

    await stats.getOrCreate(viewer.id, viewer.username);

    if (mode === 'set') {
      await stats.set(userID, statName, value);
    } else {
      await stats.adjust(userID, statName, value);
    }

    return buildUserDetail(userID);
  },
);

export const displayStats = command(
  v.object({ userID: v.pipe(v.string(), v.nonEmpty()) }),
  async ({ userID }) => {
    requireLocalAdmin();

    const { stats } = getServices();
    const viewer = await stats.getViewer(userID);

    await stats.getOrCreate(viewer.id, viewer.username);
    await sendChatStats(viewer.id, viewer.username);

    return buildUserDetail(userID);
  },
);

export const grantRandomStat = command(
  v.object({
    userID: v.pipe(v.string(), v.nonEmpty()),
    displayInChat: v.boolean(),
  }),
  async ({ userID, displayInChat }) => {
    requireLocalAdmin();

    const { stats, random } = getServices();
    const viewer = await stats.getViewer(userID);

    await stats.getOrCreate(viewer.id, viewer.username);

    const definitions = stats.definitions;
    const definition = definitions[random.integer(definitions.length)];
    await stats.adjust(userID, definition.name, 1);

    if (displayInChat) {
      await sendChatStats(viewer.id, viewer.username);
    }

    const grant: GrantResult = { kind: 'stat', statName: definition.name };

    return buildUserDetail(userID, grant);
  },
);

export const explode = command(
  v.object({ userID: v.pipe(v.string(), v.nonEmpty()) }),
  async ({ userID }) => {
    requireLocalAdmin();

    findStatDefinition('penis');

    const { stats } = getServices();
    const viewer = await stats.getViewer(userID);

    await stats.getOrCreate(viewer.id, viewer.username);
    await stats.set(userID, 'penis', EXPLODED_PENIS_VALUE);
    await sendChatStats(viewer.id, viewer.username);

    return buildUserDetail(userID);
  },
);

export const undoExplode = command(
  v.object({ userID: v.pipe(v.string(), v.nonEmpty()) }),
  async ({ userID }) => {
    requireLocalAdmin();

    const definition = findStatDefinition('penis');

    const { stats } = getServices();
    const viewer = await stats.getViewer(userID);

    await stats.getOrCreate(viewer.id, viewer.username);
    await stats.set(userID, 'penis', definition.defaultValue);
    await sendChatStats(viewer.id, viewer.username);

    return buildUserDetail(userID);
  },
);

export const resetStats = command(
  v.object({
    userID: v.pipe(v.string(), v.nonEmpty()),
    displayInChat: v.boolean(),
  }),
  async ({ userID, displayInChat }) => {
    requireLocalAdmin();

    const { stats } = getServices();
    const viewer = await stats.getViewer(userID);

    await stats.getOrCreate(viewer.id, viewer.username);
    await stats.reset(userID);

    if (displayInChat) {
      await sendChatStats(viewer.id, viewer.username);
    }

    return buildUserDetail(userID);
  },
);

export const grantPlushie = command(
  v.object({
    userID: v.pipe(v.string(), v.nonEmpty()),
    series: v.pipe(v.string(), v.nonEmpty()),
    key: v.pipe(v.string(), v.nonEmpty()),
    triggerOverlay: v.boolean(),
  }),
  async ({ userID, series, key, triggerOverlay }) => {
    requireLocalAdmin();

    const { stats, blindBox, overlay, catalog } = getServices();
    const viewer = await stats.getViewer(userID);
    const seriesConfig = catalog.series.find((s) => s.series === series);

    if (!seriesConfig) {
      throw error(400, 'unknown series');
    }

    const plushie = seriesConfig.plushies.find((p) => p.key === key);

    if (!plushie) {
      throw error(400, 'unknown series or plushie');
    }

    const result = await blindBox.grant(viewer.id, viewer.username, series, key);

    if (triggerOverlay) {
      const collection = await blindBox.getCollection(viewer.id, series);

      overlay.publish({
        type: 'blindbox_redemption',
        username: viewer.username,
        plushie,
        isNew: result.isNew,
        collection,
        config: seriesConfig,
      });
    }

    return buildUserDetail(userID);
  },
);

export const grantRandomPlushie = command(
  v.object({
    userID: v.pipe(v.string(), v.nonEmpty()),
    series: v.pipe(v.string(), v.nonEmpty()),
    triggerOverlay: v.boolean(),
  }),
  async ({ userID, series, triggerOverlay }) => {
    requireLocalAdmin();

    const { stats, blindBox, overlay, random, catalog } = getServices();
    const viewer = await stats.getViewer(userID);
    const seriesConfig = catalog.series.find((s) => s.series === series);

    if (!seriesConfig) {
      throw error(400, 'unknown series');
    }

    const plushie = pickWeightedPlushie(seriesConfig.plushies, random);
    const result = await blindBox.grant(viewer.id, viewer.username, series, plushie.key);

    if (triggerOverlay) {
      const collection = await blindBox.getCollection(viewer.id, series);

      overlay.publish({
        type: 'blindbox_redemption',
        username: viewer.username,
        plushie,
        isNew: result.isNew,
        collection,
        config: seriesConfig,
      });
    }

    const grant: GrantResult = {
      kind: 'plushie',
      plushieName: plushie.name,
      isDuplicate: !result.isNew,
    };

    return buildUserDetail(userID, grant);
  },
);

export const removePlushie = command(
  v.object({
    userID: v.pipe(v.string(), v.nonEmpty()),
    series: v.pipe(v.string(), v.nonEmpty()),
    key: v.pipe(v.string(), v.nonEmpty()),
  }),
  async ({ userID, series, key }) => {
    requireLocalAdmin();

    const { blindBox } = getServices();

    await blindBox.remove(userID, series, key);

    return buildUserDetail(userID);
  },
);

export const resetCollection = command(
  v.object({
    userID: v.pipe(v.string(), v.nonEmpty()),
    series: v.pipe(v.string(), v.nonEmpty()),
  }),
  async ({ userID, series }) => {
    requireLocalAdmin();

    const { blindBox } = getServices();

    await blindBox.reset(userID, series);

    return buildUserDetail(userID);
  },
);

export const displayCollection = command(
  v.object({
    userID: v.pipe(v.string(), v.nonEmpty()),
    series: v.pipe(v.string(), v.nonEmpty()),
  }),
  async ({ userID, series }) => {
    requireLocalAdmin();

    const { stats, blindBox, overlay, catalog } = getServices();
    const viewer = await stats.getViewer(userID);
    const seriesConfig = catalog.series.find((s) => s.series === series);

    if (!seriesConfig) {
      throw error(400, 'unknown series');
    }

    const collection = await blindBox.getCollection(viewer.id, series);

    overlay.publish({
      type: 'blindbox_display',
      username: viewer.username,
      collection,
      config: seriesConfig,
    });

    return buildUserDetail(userID);
  },
);
