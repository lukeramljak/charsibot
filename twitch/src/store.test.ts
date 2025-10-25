import { describe, expect, it, beforeEach, afterEach } from "bun:test";
import { Store, formatStats } from "./store";

describe("Store", () => {
  let store: Store;

  beforeEach(async () => {
    store = new Store(":memory:");
    await store.init();
  });

  afterEach(async () => {
    store.db.close();
  });

  describe("tokens", () => {
    it("saves and retrieves streamer tokens", async () => {
      await store.saveTokens("streamer", "access123", "refresh456");
      const tokens = await store.getTokens("streamer");

      expect(tokens).not.toBeNull();
      expect(tokens?.access_token).toBe("access123");
      expect(tokens?.refresh_token).toBe("refresh456");
    });

    it("saves and retrieves bot tokens", async () => {
      await store.saveTokens("bot", "bot_access", "bot_refresh");
      const tokens = await store.getTokens("bot");

      expect(tokens).not.toBeNull();
      expect(tokens?.access_token).toBe("bot_access");
      expect(tokens?.refresh_token).toBe("bot_refresh");
    });

    it("updates existing tokens", async () => {
      await store.saveTokens("streamer", "old_access", "old_refresh");
      await store.saveTokens("streamer", "new_access", "new_refresh");
      const tokens = await store.getTokens("streamer");

      expect(tokens?.access_token).toBe("new_access");
      expect(tokens?.refresh_token).toBe("new_refresh");
    });

    it("returns null for non-existent token type", async () => {
      const tokens = await store.getTokens("streamer");
      expect(tokens).toBeNull();
    });
  });

  describe("stats", () => {
    it("creates new user stats with defaults", async () => {
      const stats = await store.getStats("user123", "testuser");

      expect(stats.id).toBe("user123");
      expect(stats.username).toBe("testuser");
      expect(stats.strength).toBe(0);
      expect(stats.intelligence).toBe(0);
      expect(stats.charisma).toBe(0);
      expect(stats.luck).toBe(0);
      expect(stats.dexterity).toBe(0);
      expect(stats.penis).toBe(0);
    });

    it("retrieves existing user stats", async () => {
      await store.getStats("user123", "testuser");
      const stats = await store.getStats("user123", "testuser");

      expect(stats.id).toBe("user123");
      expect(stats.username).toBe("testuser");
    });

    it("modifies strength stat", async () => {
      await store.modifyStat("user123", "testuser", "strength", 5);
      const stats = await store.getStats("user123", "testuser");

      expect(stats.strength).toBe(5);
      expect(stats.intelligence).toBe(0);
    });

    it("modifies intelligence stat", async () => {
      await store.modifyStat("user123", "testuser", "intelligence", 3);
      const stats = await store.getStats("user123", "testuser");

      expect(stats.intelligence).toBe(3);
    });

    it("accumulates stat modifications", async () => {
      await store.modifyStat("user123", "testuser", "luck", 2);
      await store.modifyStat("user123", "testuser", "luck", 3);
      const stats = await store.getStats("user123", "testuser");

      expect(stats.luck).toBe(5);
    });

    it("handles negative stat modifications", async () => {
      await store.modifyStat("user123", "testuser", "charisma", 10);
      await store.modifyStat("user123", "testuser", "charisma", -3);
      const stats = await store.getStats("user123", "testuser");

      expect(stats.charisma).toBe(7);
    });

    it("allows stats to go negative", async () => {
      await store.modifyStat("user123", "testuser", "dexterity", -5);
      const stats = await store.getStats("user123", "testuser");

      expect(stats.dexterity).toBe(-5);
    });

    it("tracks multiple users independently", async () => {
      await store.modifyStat("user1", "alice", "strength", 10);
      await store.modifyStat("user2", "bob", "strength", 20);

      const aliceStats = await store.getStats("user1", "alice");
      const bobStats = await store.getStats("user2", "bob");

      expect(aliceStats.strength).toBe(10);
      expect(bobStats.strength).toBe(20);
    });
  });

  describe("formatStats", () => {
    it("formats stats with positive values", async () => {
      await store.modifyStat("user123", "testuser", "strength", 5);
      await store.modifyStat("user123", "testuser", "intelligence", 3);

      const stats = await store.getStats("user123", "testuser");
      const formatted = formatStats("testuser", stats);

      expect(formatted).toContain("testuser");
      expect(formatted).toContain("STR: 5");
      expect(formatted).toContain("INT: 3");
    });

    it("formats stats with negative values", async () => {
      await store.modifyStat("user123", "testuser", "luck", -2);

      const stats = await store.getStats("user123", "testuser");
      const formatted = formatStats("testuser", stats);

      expect(formatted).toContain("LUCK: -2");
    });
  });
});
