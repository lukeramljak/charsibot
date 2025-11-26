import type { Bot } from '@/bot/bot';
import type { Command } from '@/commands/command';

export class StatLeaderboardCommand implements Command {
  shouldTrigger(command: string): boolean {
    return command === 'leaderboard';
  }

  async execute(bot: Bot): Promise<void> {
    const [stats] = await bot.store.getStatLeaderboard();

    const statMap: Record<string, { username: string; value: number }> = {
      STR: { username: stats.top_strength_username, value: stats.top_strength_value },
      INT: { username: stats.top_intelligence_username, value: stats.top_intelligence_value },
      CHA: { username: stats.top_charisma_username, value: stats.top_charisma_value },
      LUCK: { username: stats.top_luck_username, value: stats.top_luck_value },
      DEX: { username: stats.top_dexterity_username, value: stats.top_dexterity_value },
      PENIS: { username: stats.top_penis_username, value: stats.top_penis_value },
    };

    const emojiMap: Record<string, string> = {
      STR: '💪',
      INT: '🧠',
      CHA: '✨',
      LUCK: '🍀',
      DEX: '🤸',
      PENIS: '🍆',
    };

    const message = Object.entries(statMap)
      .map(([label, { username, value }]) => `${emojiMap[label]} ${username}(${value})`)
      .join(' | ');

    await bot.sendMessage(`Stats leaderboard: ${message}`);
  }
}
