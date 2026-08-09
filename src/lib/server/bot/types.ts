import type { BlindBoxSeries } from '$lib/contracts/catalog';
import type {
  BlindBoxService,
  ChatSender,
  Clock,
  Logger,
  OverlayBus,
  Random,
  StatsService,
} from '$lib/server/application/ports';

export interface ChatMessageEvent {
  chatterUserID: string;
  chatterUsername: string;
  messageID: string;
  text: string;
}

export interface ChannelPointRedemptionEvent {
  userID: string;
  username: string;
  rewardTitle: string;
}

export interface RaidEvent {
  fromBroadcasterUsername: string;
}

export interface Bot {
  handleChatMessage: (event: ChatMessageEvent) => Promise<void>;
  handleRedemption: (event: ChannelPointRedemptionEvent) => Promise<void>;
  handleRaid: (event: RaidEvent) => Promise<void>;
  stop: (reason?: string) => Promise<void>;
}

export interface BotDependencies {
  botUserID: string;
  series: readonly BlindBoxSeries[];
  stats: StatsService;
  blindBox: BlindBoxService;
  overlay: OverlayBus;
  chat: ChatSender;
  clock: Clock;
  random: Random;
  logger: Logger;
}

export type SendChat = (
  message: string,
  signal: AbortSignal,
  replyParentMessageID?: string,
) => Promise<void>;

export type Command = (event: ChatMessageEvent, signal: AbortSignal) => Promise<void>;

export interface Trigger {
  chance: number;
  matches: (event: ChatMessageEvent) => boolean;
  execute: Command;
}

export type Redemption = (event: ChannelPointRedemptionEvent, signal: AbortSignal) => Promise<void>;
