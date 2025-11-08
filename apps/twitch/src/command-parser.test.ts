import { describe, expect, it } from 'bun:test';
import { parseModifyStatCommand } from './command-parser';

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
