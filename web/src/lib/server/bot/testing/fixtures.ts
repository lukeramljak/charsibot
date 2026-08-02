import type { ChannelPointRedemptionEvent, ChatMessageEvent, RaidEvent } from '$lib/server/bot';

export const chatMessageEvent = (overrides: Partial<ChatMessageEvent> = {}): ChatMessageEvent => ({
  chatterUserID: 'viewer-1',
  chatterUsername: 'alice',
  messageID: 'message-1',
  text: 'hello',
  ...overrides,
});

export const redemptionEvent = (
  overrides: Partial<ChannelPointRedemptionEvent> = {},
): ChannelPointRedemptionEvent => ({
  userID: 'viewer-1',
  username: 'alice',
  rewardTitle: 'Unknown Reward',
  ...overrides,
});

export const raidEvent = (overrides: Partial<RaidEvent> = {}): RaidEvent => ({
  fromBroadcasterUsername: 'raider',
  ...overrides,
});
