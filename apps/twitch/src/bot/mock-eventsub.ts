import type { EventSubChannelRedemptionAddEvent } from '@twurple/eventsub-base';
import { log } from '@/logger';

interface MockEventSubConfig {
  url: string;
  onRedemption: (event: EventSubChannelRedemptionAddEvent) => void;
}

export class MockEventSubListener {
  private ws: WebSocket | null = null;
  private sessionId: string | null = null;

  constructor(private config: MockEventSubConfig) {}

  start() {
    this.ws = new WebSocket(this.config.url);

    this.ws.addEventListener('open', () => {
      log.info('Mock WebSocket connected');
    });

    this.ws.addEventListener('message', (event) => {
      const message = JSON.parse(event.data as string);
      log.debug({ message }, 'Mock WebSocket message received');

      if (message.metadata?.message_type === 'session_welcome') {
        this.sessionId = message.payload?.session?.id;
        log.info({ sessionId: this.sessionId }, 'Mock WebSocket session established');
      }

      if (message.metadata?.message_type === 'notification') {
        const subscriptionType = message.metadata?.subscription_type;
        const event = message.payload?.event;

        log.info({ subscriptionType, event }, 'Mock event notification received');

        if (subscriptionType === 'channel.channel_points_custom_reward_redemption.add') {
          this.config.onRedemption({
            broadcasterDisplayName: event.broadcaster_user_name,
            broadcasterId: event.broadcaster_user_id,
            id: event.user_id,
            input: event.user_input,
            redemptionDate: event.redeemed_at,
            rewardCost: event.reward.cost,
            rewardId: event.reward.id,
            rewardPrompt: event.reward.prompt,
            rewardTitle: event.reward.title,
            userId: event.user_id,
            userDisplayName: event.user_name,
            userName: event.user_name,
            status: event.status,
          } as EventSubChannelRedemptionAddEvent);
        }
      }

      if (message.metadata?.message_type === 'session_keepalive') {
        log.debug('Mock WebSocket keepalive received');
      }
    });

    this.ws.addEventListener('error', (event) => {
      log.error({ error: event }, 'Mock WebSocket error');
    });

    this.ws.addEventListener('close', () => {
      log.info('Mock WebSocket closed');
    });
  }

  stop() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }
}
