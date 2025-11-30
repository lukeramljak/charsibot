import { userCollectionsTable } from '@/storage/schema';
import { Store } from '@/storage/store';
import { Database } from 'bun:sqlite';
import { beforeEach, describe, expect, it } from 'bun:test';

describe('Store', () => {
  let store: Store;
  let sqlite: Database;

  beforeEach(async () => {
    sqlite = new Database(':memory:');
    store = new Store(sqlite);
    await store.init();
  });

  describe('tokens', () => {
    it('saves and retrieves streamer tokens', async () => {
      store.saveTokens('streamer', 'access123', 'refresh456');
      const tokens = await store.getTokens('streamer');

      expect(tokens).not.toBeNull();
      expect(tokens?.accessToken).toBe('access123');
      expect(tokens?.refreshToken).toBe('refresh456');
    });

    it('saves and retrieves bot tokens', async () => {
      await store.saveTokens('bot', 'bot_access', 'bot_refresh');
      const tokens = await store.getTokens('bot');

      expect(tokens).not.toBeNull();
      expect(tokens?.accessToken).toBe('bot_access');
      expect(tokens?.refreshToken).toBe('bot_refresh');
    });

    it('updates existing tokens', async () => {
      await store.saveTokens('streamer', 'old_access', 'old_refresh');
      await store.saveTokens('streamer', 'new_access', 'new_refresh');
      const tokens = await store.getTokens('streamer');

      expect(tokens?.accessToken).toBe('new_access');
      expect(tokens?.refreshToken).toBe('new_refresh');
    });

    it('returns null for non-existent token type', async () => {
      const tokens = await store.getTokens('streamer');
      expect(tokens).toBeNull();
    });
  });

  describe('stats', () => {
    it('creates new user stats with defaults', async () => {
      const stats = await store.getStats('user123', 'testuser');

      expect(stats.id).toBe('user123');
      expect(stats.username).toBe('testuser');
      expect(stats.strength).toBe(3);
      expect(stats.intelligence).toBe(3);
      expect(stats.charisma).toBe(3);
      expect(stats.luck).toBe(3);
      expect(stats.dexterity).toBe(3);
      expect(stats.penis).toBe(3);
    });

    it('retrieves existing user stats', async () => {
      await store.getStats('user123', 'testuser');
      const stats = await store.getStats('user123', 'testuser');

      expect(stats.id).toBe('user123');
      expect(stats.username).toBe('testuser');
    });

    it('modifies strength stat', async () => {
      await store.modifyStat('user123', 'testuser', 'strength', 5);
      const stats = await store.getStats('user123', 'testuser');

      expect(stats.strength).toBe(8);
      expect(stats.intelligence).toBe(3);
    });

    it('modifies intelligence stat', async () => {
      await store.modifyStat('user123', 'testuser', 'intelligence', 3);
      const stats = await store.getStats('user123', 'testuser');

      expect(stats.intelligence).toBe(6);
    });

    it('accumulates stat modifications', async () => {
      await store.modifyStat('user123', 'testuser', 'luck', 2);
      await store.modifyStat('user123', 'testuser', 'luck', 3);
      const stats = await store.getStats('user123', 'testuser');

      expect(stats.luck).toBe(8);
    });

    it('handles negative stat modifications', async () => {
      await store.modifyStat('user123', 'testuser', 'charisma', 10);
      await store.modifyStat('user123', 'testuser', 'charisma', -3);
      const stats = await store.getStats('user123', 'testuser');

      expect(stats.charisma).toBe(10);
    });

    it('allows stats to go negative', async () => {
      await store.modifyStat('user123', 'testuser', 'dexterity', -5);
      const stats = await store.getStats('user123', 'testuser');

      expect(stats.dexterity).toBe(-2);
    });

    it('tracks multiple users independently', async () => {
      await store.modifyStat('user1', 'alice', 'strength', 10);
      await store.modifyStat('user2', 'bob', 'strength', 20);

      const aliceStats = await store.getStats('user1', 'alice');
      const bobStats = await store.getStats('user2', 'bob');

      expect(aliceStats.strength).toBe(13);
      expect(bobStats.strength).toBe(23);
    });
  });

  describe('collections', () => {
    it('returns collectionType -> usernames[] only when all rewards are true', async () => {
      await store.db.insert(userCollectionsTable).values([
        {
          username: 'alice',
          collectionType: 'coobubu',
          reward1: 1,
          reward2: 1,
          reward3: 1,
          reward4: 1,
          reward5: 1,
          reward6: 1,
          reward7: 1,
          reward8: 1,
        },
        {
          username: 'bob',
          collectionType: 'coobubu',
          reward1: 1,
          reward2: 1,
          reward3: 1,
          reward4: 1,
          reward5: 1,
          reward6: 1,
          reward7: 1,
          reward8: 1,
        },
        {
          username: 'clown',
          collectionType: 'coobubu',
          reward1: 1,
          reward2: 1,
          reward3: 1,
          reward4: 1,
          reward5: 1,
          reward6: 1,
          reward7: 1,
          reward8: 0,
        }, // fails last reward
        {
          username: 'xena',
          collectionType: 'olliepop',
          reward1: 1,
          reward2: 1,
          reward3: 1,
          reward4: 1,
          reward5: 1,
          reward6: 1,
          reward7: 1,
          reward8: 1,
        },
        {
          username: 'yuri',
          collectionType: 'olliepop',
          reward1: 1,
          reward2: 1,
          reward3: 0,
          reward4: 1,
          reward5: 1,
          reward6: 1,
          reward7: 1,
          reward8: 1,
        }, // fails reward3
      ]);

      const rows = await store.getCompletedCollections();
      expect(rows.length).toBe(2);
      const coobubu = rows.find((r) => r.collectionType === 'coobubu');
      const olliepop = rows.find((r) => r.collectionType === 'olliepop');
      expect(coobubu?.usernames.sort()).toEqual(['alice', 'bob']);
      expect(olliepop?.usernames).toEqual(['xena']);
    });

    it('resets user collections', async () => {
      await store.db.insert(userCollectionsTable).values([
        {
          userId: '1',
          username: 'alice',
          collectionType: 'coobubu',
          reward1: 1,
          reward2: 1,
          reward3: 1,
          reward4: 1,
          reward5: 1,
          reward6: 1,
          reward7: 1,
          reward8: 1,
        },
        {
          userId: '2',
          username: 'bob',
          collectionType: 'coobubu',
          reward1: 1,
          reward2: 1,
          reward3: 1,
          reward4: 1,
          reward5: 1,
          reward6: 1,
          reward7: 1,
          reward8: 1,
        },
      ]);

      const collectionBefore = await store.getUserCollections('1', 'coobubu');
      expect(collectionBefore).toEqual([
        'reward1',
        'reward2',
        'reward3',
        'reward4',
        'reward5',
        'reward6',
        'reward7',
        'reward8',
      ]);

      await store.resetUserCollection('1', 'coobubu');

      const collectionAfter = await store.getUserCollections('1', 'coobubu');
      expect(collectionAfter).toEqual([]);
    });

    it('adds plushie to collection - new reward', async () => {
      await store.db.insert(userCollectionsTable).values({
        userId: '1',
        username: 'alice',
        collectionType: 'coobubu',
        reward1: 1,
        reward2: 0,
        reward3: 0,
        reward4: 0,
        reward5: 0,
        reward6: 0,
        reward7: 0,
        reward8: 0,
      });

      const result = await store.addPlushieToCollection('1', 'alice', 'coobubu', 'reward2');

      expect(result).toBeDefined();
      expect(result?.isNew).toBe(true);
      expect(result?.collection).toEqual(['reward1', 'reward2']);
    });

    it('adds plushie to collection - existing reward', async () => {
      await store.db.insert(userCollectionsTable).values({
        userId: '1',
        username: 'alice',
        collectionType: 'coobubu',
        reward1: 1,
        reward2: 1,
        reward3: 0,
        reward4: 0,
        reward5: 0,
        reward6: 0,
        reward7: 0,
        reward8: 0,
      });

      const result = await store.addPlushieToCollection('1', 'alice', 'coobubu', 'reward1');

      expect(result).toBeDefined();
      expect(result?.isNew).toBe(false);
      expect(result?.collection).toEqual(['reward1', 'reward2']);
    });

    it('adds plushie to collection - first reward for user', async () => {
      const result = await store.addPlushieToCollection('999', 'newuser', 'coobubu', 'reward3');

      expect(result).toBeDefined();
      expect(result?.isNew).toBe(true);
      expect(result?.collection).toEqual(['reward3']);
    });

    it('gets user collections - existing user', async () => {
      await store.db.insert(userCollectionsTable).values({
        userId: '1',
        username: 'alice',
        collectionType: 'coobubu',
        reward1: 1,
        reward2: 0,
        reward3: 1,
        reward4: 0,
        reward5: 1,
        reward6: 0,
        reward7: 0,
        reward8: 0,
      });

      const collection = await store.getUserCollections('1', 'coobubu');
      expect(collection).toEqual(['reward1', 'reward3', 'reward5']);
    });

    it('gets user collections - non-existent user', async () => {
      const collection = await store.getUserCollections('999', 'coobubu');
      expect(collection).toEqual([]);
    });
  });
});
