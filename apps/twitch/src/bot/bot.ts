import { ApiClient } from '@twurple/api';
import { RefreshingAuthProvider } from '@twurple/auth';
import type {
  EventSubChannelChatMessageEvent,
  EventSubChannelRedemptionAddEvent,
} from '@twurple/eventsub-base';
import { EventSubWsListener } from '@twurple/eventsub-ws';
import type { Config } from '../config';
import { log } from '../logger';
import { MockEventSubListener } from './mock-eventsub';
import { Store } from '../storage/store';
import { WebSocketServer } from '../websocket/websocket';
import type { MessageHandler } from '../events/message-handler';
import type { CommandHandler } from '../events/command-handler';
import type { RedemptionHandler } from '../events/redemption-handler';

interface BotConfig {
  config: Config;
  store: Store;
  commandHandler: CommandHandler;
  messageHandler: MessageHandler;
  redemptionHandler: RedemptionHandler;
}

export class Bot {
  private config: Config;
  public store: Store;
  private commandHandler: CommandHandler;
  private messageHandler: MessageHandler;
  private redemptionHandler: RedemptionHandler;
  private api: ApiClient;
  private listener: EventSubWsListener;
  private mockListener?: MockEventSubListener;
  private authProvider: RefreshingAuthProvider;
  public wsServer: WebSocketServer;

  constructor({ config, store, commandHandler, messageHandler, redemptionHandler }: BotConfig) {
    this.config = config;
    this.store = store;
    this.commandHandler = commandHandler;
    this.messageHandler = messageHandler;
    this.redemptionHandler = redemptionHandler;
    this.wsServer = new WebSocketServer(config.wsPort);
    this.authProvider = new RefreshingAuthProvider({
      clientId: config.clientId,
      clientSecret: config.clientSecret,
    });

    this.authProvider.onRefresh(async (userId, tokenData) => {
      if (!tokenData.refreshToken) return;

      if (userId === this.config.channelUserId) {
        await this.store.saveTokens('streamer', tokenData.accessToken, tokenData.refreshToken);
        log.info({ userId }, 'streamer tokens refreshed');
      }

      if (userId === this.config.botUserId) {
        await this.store.saveTokens('bot', tokenData.accessToken, tokenData.refreshToken);
        log.info({ userId }, 'bot tokens refreshed');
      }
    });

    this.authProvider.addUserForToken({
      accessToken: config.streamerAccessToken,
      refreshToken: config.streamerRefreshToken,
      expiresIn: 0,
      obtainmentTimestamp: 0,
    });

    if (config.botUserId !== config.channelUserId) {
      this.authProvider.addUserForToken({
        accessToken: config.botAccessToken,
        refreshToken: config.botRefreshToken,
        expiresIn: 0,
        obtainmentTimestamp: 0,
      });
    }

    this.api = new ApiClient({ authProvider: this.authProvider });
    this.listener = new EventSubWsListener({ apiClient: this.api });
  }

  async init() {
    await this.store.init();
    log.info('store ready');

    this.wsServer.start();
    log.info('websocket server ready');

    await this.refreshTokens();
    log.info('tokens initialised');

    if (this.config.useMockServer) {
      this.mockListener = new MockEventSubListener({
        url: 'ws://127.0.0.1:8080/ws',
        onRedemption: async (event) => {
          await this.onChannelPointRedemption(event);
        },
      });
      this.mockListener.start();
      log.info('mock eventsub started');
    } else {
      this.listener.onChannelChatMessage(
        this.config.channelUserId,
        this.config.channelUserId,
        async (event) => {
          return this.onMessage(event);
        },
      );

      this.listener.onChannelRedemptionAdd(this.config.channelUserId, async (event) => {
        return this.onChannelPointRedemption(event);
      });

      try {
        this.listener.start();
        log.info('eventsub started');
      } catch (err) {
        log.error({ err }, 'eventsub start failed');
      }
    }

    log.info('bot ready');
  }

  public getId(): string {
    return this.config.botUserId;
  }

  private async onMessage(event: EventSubChannelChatMessageEvent): Promise<void> {
    try {
      await this.commandHandler.process(this, event);
      await this.messageHandler.process(this, event);
    } catch (error) {
      log.error({ error }, 'message processing failed');
    }
  }

  private async onChannelPointRedemption(event: EventSubChannelRedemptionAddEvent): Promise<void> {
    try {
      await this.redemptionHandler.process(this, event);
    } catch (error) {
      log.error({ error }, 'redemption processing failed');
    }
  }

  private async refreshTokens() {
    const userIds = [this.config.channelUserId];
    if (this.config.botUserId !== this.config.channelUserId) {
      userIds.push(this.config.botUserId);
    }

    const promises = userIds.map((userId) => {
      return new Promise<void>((resolve) => {
        const checkToken = async () => {
          const token = await this.authProvider.getAccessTokenForUser(userId);
          if (token) {
            resolve();
          } else {
            setTimeout(checkToken, 100);
          }
        };
        checkToken();
      });
    });

    await Promise.all(promises);
  }

  async sendMessage(message: string, replyParentMessageId?: string) {
    await this.api.asUser(this.config.botUserId, async (ctx) => {
      await ctx.chat.sendChatMessage(this.config.channelUserId, message, {
        replyParentMessageId,
      });
    });
    log.debug({ message }, 'message sent');
  }
}
