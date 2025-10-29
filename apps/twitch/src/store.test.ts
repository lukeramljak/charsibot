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
    it("saves and retrieves streamer tokens", () => {
      store.saveTokens("streamer", "access123", "refresh456");
      const tokens = store.getTokens("streamer");

      expect(tokens).not.toBeNull();
      expect(tokens?.access_token).toBe("access123");
      expect(tokens?.refresh_token).toBe("refresh456");
    });

    it("saves and retrieves bot tokens", () => {
      store.saveTokens("bot", "bot_access", "bot_refresh");
      const tokens = store.getTokens("bot");

      expect(tokens).not.toBeNull();
      expect(tokens?.access_token).toBe("bot_access");
      expect(tokens?.refresh_token).toBe("bot_refresh");
    });

    it("updates existing tokens", () => {
      store.saveTokens("streamer", "old_access", "old_refresh");
      store.saveTokens("streamer", "new_access", "new_refresh");
      const tokens = store.getTokens("streamer");

      expect(tokens?.access_token).toBe("new_access");
      expect(tokens?.refresh_token).toBe("new_refresh");
    });

    it("returns null for non-existent token type", () => {
      const tokens = store.getTokens("streamer");
      expect(tokens).toBeNull();
    });
  });

  describe("stats", () => {
    it("creates new user stats with defaults", () => {
      const stats = store.getStats("user123", "testuser");

      expect(stats.id).toBe("user123");
      expect(stats.username).toBe("testuser");
      expect(stats.strength).toBe(3);
      expect(stats.intelligence).toBe(3);
      expect(stats.charisma).toBe(3);
      expect(stats.luck).toBe(3);
      expect(stats.dexterity).toBe(3);
      expect(stats.penis).toBe(3);
    });

    it("retrieves existing user stats", () => {
      store.getStats("user123", "testuser");
      const stats = store.getStats("user123", "testuser");

      expect(stats.id).toBe("user123");
      expect(stats.username).toBe("testuser");
    });

    it("modifies strength stat", () => {
      store.modifyStat("user123", "testuser", "strength", 5);
      const stats = store.getStats("user123", "testuser");

      expect(stats.strength).toBe(8);
      expect(stats.intelligence).toBe(3);
    });

    it("modifies intelligence stat", () => {
      store.modifyStat("user123", "testuser", "intelligence", 3);
      const stats = store.getStats("user123", "testuser");

      expect(stats.intelligence).toBe(6);
    });

    it("accumulates stat modifications", () => {
      store.modifyStat("user123", "testuser", "luck", 2);
      store.modifyStat("user123", "testuser", "luck", 3);
      const stats = store.getStats("user123", "testuser");

      expect(stats.luck).toBe(8);
    });

    it("handles negative stat modifications", () => {
      store.modifyStat("user123", "testuser", "charisma", 10);
      store.modifyStat("user123", "testuser", "charisma", -3);
      const stats = store.getStats("user123", "testuser");

      expect(stats.charisma).toBe(10);
    });

    it("allows stats to go negative", () => {
      store.modifyStat("user123", "testuser", "dexterity", -5);
      const stats = store.getStats("user123", "testuser");

      expect(stats.dexterity).toBe(-2);
    });

    it("tracks multiple users independently", () => {
      store.modifyStat("user1", "alice", "strength", 10);
      store.modifyStat("user2", "bob", "strength", 20);

      const aliceStats = store.getStats("user1", "alice");
      const bobStats = store.getStats("user2", "bob");

      expect(aliceStats.strength).toBe(13);
      expect(bobStats.strength).toBe(23);
    });
  });

  describe("formatStats", () => {
    it("formats stats with positive values", () => {
      store.modifyStat("user123", "testuser", "strength", 5);
      store.modifyStat("user123", "testuser", "intelligence", 3);

      const stats = store.getStats("user123", "testuser");
      const formatted = formatStats("testuser", stats);

      expect(formatted).toContain("testuser");
      expect(formatted).toContain("STR: 8");
      expect(formatted).toContain("INT: 6");
    });

    it("formats stats with negative values", () => {
      store.modifyStat("user123", "testuser", "luck", -2);

      const stats = store.getStats("user123", "testuser");
      const formatted = formatStats("testuser", stats);

      expect(formatted).toContain("LUCK: 1");
    });
  });
});
