import { MockEventSubListener } from '@/bot/mock-eventsub';
import type { Config } from '@/config';
import type { CommandHandler } from '@/events/command-handler';
import type { MessageHandler } from '@/events/message-handler';
import type { RedemptionHandler } from '@/events/redemption-handler';
import { log } from '@/logger';
import { Store, type TokenType } from '@/storage/store';
import { WebSocketServer } from '@/websocket/websocket';
import { ApiClient } from '@twurple/api';
import { RefreshingAuthProvider, type AccessToken } from '@twurple/auth';
import type {
  EventSubChannelChatMessageEvent,
  EventSubChannelRedemptionAddEvent,
} from '@twurple/eventsub-base';
import { EventSubWsListener } from '@twurple/eventsub-ws';

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
  public api: ApiClient;
  private listener?: EventSubWsListener;
  private mockListener?: MockEventSubListener;
  private authProvider: RefreshingAuthProvider;
  public wsServer: WebSocketServer;
  private isRunning = false;

  constructor({ config, store, commandHandler, messageHandler, redemptionHandler }: BotConfig) {
    this.config = config;
    this.store = store;
    this.commandHandler = commandHandler;
    this.messageHandler = messageHandler;
    this.redemptionHandler = redemptionHandler;
    this.wsServer = new WebSocketServer(config.wsPort);
    this.authProvider = this.createAuthProvider();
    this.api = this.createApiClient();
  }

  async init(): Promise<void> {
    if (this.isRunning) {
      log.warn('bot already running');
      return;
    }

    try {
      await this.store.init();
      await this.refreshTokens();
      await this.initialiseEventSub();
      this.wsServer.start();

      this.isRunning = true;
      log.info('bot ready');
    } catch (error) {
      log.error({ error }, 'bot initialisation failed');
      await this.shutdown();
      throw error;
    }
  }

  async shutdown(): Promise<void> {
    if (!this.isRunning) {
      return;
    }

    log.info('shutting down bot');

    try {
      this.listener?.stop();
      this.mockListener?.stop();
      this.wsServer.stop();
    } catch (error) {
      log.error({ error }, 'shutdown error');
    } finally {
      this.isRunning = false;
      log.info('bot stopped');
    }
  }

  private async initialiseEventSub(): Promise<void> {
    if (this.config.useMockServer) {
      await this.startMockEventSub();
    } else {
      await this.startEventSub();
    }
  }

  private createAuthProvider(): RefreshingAuthProvider {
    const provider = new RefreshingAuthProvider({
      clientId: this.config.clientId,
      clientSecret: this.config.clientSecret,
    });

    provider.onRefresh(async (userId, tokenData) => {
      await this.handleTokenRefresh(userId, tokenData);
    });

    provider.addUser(this.config.channelUserId, {
      accessToken: this.config.streamerAccessToken,
      refreshToken: this.config.streamerRefreshToken,
      expiresIn: 0,
      obtainmentTimestamp: 0,
    });

    provider.addUser(this.config.botUserId, {
      accessToken: this.config.botAccessToken,
      refreshToken: this.config.botRefreshToken,
      expiresIn: 0,
      obtainmentTimestamp: 0,
    });

    return provider;
  }

  private createApiClient(): ApiClient {
    return new ApiClient({
      authProvider: this.authProvider,
      logger: {
        custom: {
          log: (_, message) => log.info(message),
          debug: (message) => log.debug(message),
          info: (message) => log.info(message),
          warn: (message) => log.warn(message),
          error: (message) => log.error(message),
        },
      },
    });
  }

  private async handleTokenRefresh(userId: string, tokenData: AccessToken): Promise<void> {
    if (!tokenData.refreshToken) {
      log.warn({ userId }, 'no refresh token provided');
      return;
    }

    const tokenType: TokenType = userId === this.config.channelUserId ? 'streamer' : 'bot';
    await this.store.saveTokens(tokenType, tokenData.accessToken, tokenData.refreshToken);
    log.info({ userId }, `${tokenType} tokens refreshed`);
  }

  private async refreshTokens(): Promise<void> {
    Promise.all([
      await this.authProvider.getAccessTokenForUser(this.config.channelUserId),
      await this.authProvider.getAccessTokenForUser(this.config.botUserId),
    ]);
    log.info('tokens initialised');
  }

  private async startMockEventSub(): Promise<void> {
    this.mockListener = new MockEventSubListener({
      url: 'ws://127.0.0.1:8080/ws',
      onRedemption: async (event) => {
        await this.onChannelPointRedemption(event);
      },
    });

    this.mockListener.start();
    log.info('mock eventsub started');
  }

  private async startEventSub(): Promise<void> {
    this.listener = new EventSubWsListener({
      apiClient: this.api,
      logger: {
        custom: {
          log: (_, message) => log.info(message),
          debug: (message) => log.debug(message),
          info: (message) => log.info(message),
          warn: (message) => log.warn(message),
          error: (message) => log.error(message),
        },
      },
    });

    this.listener.onChannelChatMessage(
      this.config.channelUserId,
      this.config.channelUserId,
      async (event) => {
        await this.onMessage(event);
      },
    );

    this.listener.onChannelRedemptionAdd(this.config.channelUserId, async (event) => {
      await this.onChannelPointRedemption(event);
    });

    this.listener.start();
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

  public getId(): string {
    return this.config.botUserId;
  }

  async sendMessage(message: string, replyParentMessageId?: string): Promise<void> {
    if (!this.isRunning) {
      throw new Error('bot not running');
    }

    await this.api.asUser(this.config.botUserId, async (ctx) => {
      await ctx.chat.sendChatMessage(this.config.channelUserId, message, {
        replyParentMessageId,
      });
    });

    log.debug({ message }, 'message sent');
  }
}
