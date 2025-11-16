import { describe, it, expect } from 'bun:test';
import { formatStats, parseModifyStatCommand } from './stats';
import type { Stats } from './types';

describe('formatStats', () => {
  it('formats stats with positive values', async () => {
    const stats: Stats = {
      id: 'user123',
      username: 'testuser',
      strength: 5,
      intelligence: 5,
      charisma: 3,
      luck: 3,
      dexterity: 3,
      penis: 3,
    };

    const formatted = formatStats('testuser', stats);
    expect(formatted).toEqual(
      "testuser's stats: STR: 5 | INT: 5 | CHA: 3 | LUCK: 3 | DEX: 3 | PENIS: 3",
    );
  });

  it('formats stats with negative values', async () => {
    const stats: Stats = {
      id: 'user123',
      username: 'testuser',
      strength: 3,
      intelligence: 3,
      charisma: 9,
      luck: -2,
      dexterity: 3,
      penis: 3,
    };

    const formatted = formatStats('testuser', stats);
    expect(formatted).toBe(
      "testuser's stats: STR: 3 | INT: 3 | CHA: 9 | LUCK: -2 | DEX: 3 | PENIS: 3",
    );
  });
});

describe('parseModifyStatCommand', () => {
  it('parses addstat command', () => {
    const res = parseModifyStatCommand('!addstat @foo strength 3', false);
    expect(res.error).toBeUndefined();
    expect(res.mentionedLogin).toBe('foo');
    expect(res.statColumn).toBe('strength');
    expect(res.amount).toBe(3);
    expect(res.remove).toBe(false);
  });

  it('parses rmstat command', () => {
    const res = parseModifyStatCommand('!rmstat @bar luck 2', true);
    expect(res.error).toBeUndefined();
    expect(res.mentionedLogin).toBe('bar');
    expect(res.statColumn).toBe('luck');
    expect(res.amount).toBe(2);
    expect(res.remove).toBe(true);
  });

  it('errors on missing mention', () => {
    const res = parseModifyStatCommand('!addstat strength 3', false);
    expect(res.error).toMatch(/expected format/i);
  });

  it('errors on missing stat and amount', () => {
    const res = parseModifyStatCommand('!addstat @user', false);
    expect(res.error).toMatch(/expected format/i);
  });

  it('errors on missing amount', () => {
    const res = parseModifyStatCommand('!addstat @user strength', false);
    expect(res.error).toMatch(/expected format/i);
  });

  it('errors on invalid number', () => {
    const res = parseModifyStatCommand('!addstat @user strength abc', false);
    expect(res.error).toMatch(/invalid number/i);
  });

  it('handles mention with @ symbol', () => {
    const res = parseModifyStatCommand('!addstat @username strength 5', false);
    expect(res.mentionedLogin).toBe('username');
  });

  it('converts username to lowercase', () => {
    const res = parseModifyStatCommand('!addstat @UserName strength 5', false);
    expect(res.mentionedLogin).toBe('username');
  });

  it('handles extra whitespace', () => {
    const res = parseModifyStatCommand('!addstat  @user   strength   5', false);
    expect(res.error).toBeUndefined();
    expect(res.mentionedLogin).toBe('user');
    expect(res.statColumn).toBe('strength');
    expect(res.amount).toBe(5);
  });

  it('parses negative numbers', () => {
    const res = parseModifyStatCommand('!addstat @user strength -3', false);
    expect(res.error).toBeUndefined();
    expect(res.amount).toBe(-3);
  });

  it('parses zero', () => {
    const res = parseModifyStatCommand('!addstat @user strength 0', false);
    expect(res.error).toBeUndefined();
    expect(res.amount).toBe(0);
  });

  it('handles mention in different position', () => {
    const res = parseModifyStatCommand('!addstat strength @user 5', false);
    expect(res.error).toBeUndefined();
    expect(res.mentionedLogin).toBe('user');
  });
});
