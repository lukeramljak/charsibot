import { createBot, passesTriggerChance, tokenizeTriggerWords } from '$lib/server/bot';
import { createBotHarness, createRandomFake } from '$lib/server/bot/testing/fakes';
import { chatMessageEvent } from '$lib/server/bot/testing/fixtures';
import { describe, expect, it } from 'vitest';

describe('come trigger matching', () => {
  it.each([
    ['come', true],
    ['COMING!', true],
    ['well...cum?', true],
    ['came2', false],
    ['become', false],
    ['no coming', false],
    ['No Coming', false],
    [' no coming', true],
  ])('handles %j', async (text, expected) => {
    const harness = createBotHarness({ random: createRandomFake([0]) });
    const bot = createBot(harness.dependencies);

    await bot.handleChatMessage(chatMessageEvent({ text }));

    expect(harness.chat.messages.length > 0).toBe(expected);
    if (expected) {
      expect(harness.chat.messages[0]).toEqual({
        message: 'no coming',
        replyParentMessageID: 'message-1',
      });
    }
  });

  it('splits on every non-ASCII-letter or digit like the Go tokenizer', () => {
    expect(tokenizeTriggerWords('abc-coming_cum42/camé')).toEqual([
      'abc',
      'coming',
      'cum42',
      'cam',
    ]);
  });
});

describe('trigger probability', () => {
  it('executes on roll 20 and skips on roll 21 for the 20 percent trigger', async () => {
    const passing = createBotHarness({ random: createRandomFake([19]) });
    await createBot(passing.dependencies).handleChatMessage(chatMessageEvent({ text: 'coming' }));
    expect(passing.chat.messages).toHaveLength(1);

    const failing = createBotHarness({ random: createRandomFake([20]) });
    await createBot(failing.dependencies).handleChatMessage(chatMessageEvent({ text: 'coming' }));
    expect(failing.chat.messages).toEqual([]);
  });

  it.each([-10, 0, 100, 150])('treats chance %i as unconditional', (chance) => {
    const random = createRandomFake([99]);
    expect(passesTriggerChance(chance, random)).toBe(true);
    expect(random.maximums).toEqual([]);
  });

  it('rejects invalid injected random values', () => {
    expect(() => passesTriggerChance(20, createRandomFake([100]))).toThrow(/outside/);
  });
});
