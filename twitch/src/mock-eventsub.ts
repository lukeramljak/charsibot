import { log } from "./logger";

interface MockEventSubConfig {
  url: string;
  onRedemption: (data: {
    userId: string;
    userName: string;
    rewardTitle: string;
  }) => void;
}

export class MockEventSubListener {
  private ws: WebSocket | null = null;
  private sessionId: string | null = null;

  constructor(private config: MockEventSubConfig) {}

  start() {
    this.ws = new WebSocket(this.config.url);

    this.ws.addEventListener("open", () => {
      log.info("Mock WebSocket connected");
    });

    this.ws.addEventListener("message", (event) => {
      const message = JSON.parse(event.data as string);
      log.debug({ message }, "Mock WebSocket message received");

      if (message.metadata?.message_type === "session_welcome") {
        this.sessionId = message.payload?.session?.id;
        log.info(
          { sessionId: this.sessionId },
          "Mock WebSocket session established"
        );
      }

      if (message.metadata?.message_type === "notification") {
        const subscriptionType = message.metadata?.subscription_type;
        const event = message.payload?.event;

        log.info(
          { subscriptionType, event },
          "Mock event notification received"
        );

        if (
          subscriptionType ===
          "channel.channel_points_custom_reward_redemption.add"
        ) {
          this.config.onRedemption({
            userId: event.user_id,
            userName: event.user_name,
            rewardTitle: event.reward?.title || "",
          });
        }
      }

      if (message.metadata?.message_type === "session_keepalive") {
        log.debug("Mock WebSocket keepalive received");
      }
    });

    this.ws.addEventListener("error", (event) => {
      log.error({ error: event }, "Mock WebSocket error");
    });

    this.ws.addEventListener("close", () => {
      log.info("Mock WebSocket closed");
    });
  }

  stop() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }
}
