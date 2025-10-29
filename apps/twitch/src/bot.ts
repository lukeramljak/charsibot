import { ApiClient } from "@twurple/api";
import { RefreshingAuthProvider } from "@twurple/auth";
import type { EventSubChannelChatMessageEvent } from "@twurple/eventsub-base";
import { EventSubWsListener } from "@twurple/eventsub-ws";
import { parseModifyStatCommand } from "./command-parser";
import type { Config } from "./config";
import { log } from "./logger";
import { MockEventSubListener } from "./mock-eventsub";
import { type Stat, Store, formatStats, statList } from "./store";
import { WebSocketServer } from "./websocket";
import { getWeightedRandomPlushie } from "./blind-box";
import type { CollectionType } from "./types";
import { blindBoxConfigs } from "./blind-box-configs";

export class Bot {
  private api: ApiClient;
  private listener: EventSubWsListener;
  private mockListener?: MockEventSubListener;
  private store: Store;
  private authProvider: RefreshingAuthProvider;
  private wsServer: WebSocketServer;

  constructor(private config: Config, store: Store) {
    this.store = store;
    this.wsServer = new WebSocketServer(config.wsPort);

    this.authProvider = new RefreshingAuthProvider({
      clientId: config.clientId,
      clientSecret: config.clientSecret,
    });

    this.authProvider.onRefresh(async (userId, tokenData) => {
      if (!tokenData.refreshToken) return;

      if (userId === this.config.channelUserId) {
        this.store.saveTokens(
          "streamer",
          tokenData.accessToken,
          tokenData.refreshToken
        );
        log.info({ userId }, "streamer tokens refreshed");
      }

      if (userId === this.config.botUserId) {
        this.store.saveTokens(
          "bot",
          tokenData.accessToken,
          tokenData.refreshToken
        );
        log.info({ userId }, "bot tokens refreshed");
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
    log.info("store ready");

    this.wsServer.start();
    log.info("websocket server ready");

    await this.refreshTokens();
    log.info("tokens initialised");

    if (this.config.useMockServer) {
      this.mockListener = new MockEventSubListener({
        url: "ws://127.0.0.1:8080/ws",
        onRedemption: async (data) => {
          log.info(
            {
              reward: data.rewardTitle,
              user: data.userName,
              userId: data.userId,
            },
            "channel point reward redeemed"
          );

          if (data.rewardTitle === "Drink a Potion") {
            await this.handleDrinkPotionReward(data.userId, data.userName);
          } else if (data.rewardTitle === "Tempt the Dice") {
            await this.handleTemptDiceReward(data.userId, data.userName);
          }
        },
      });
      this.mockListener.start();
      log.info("mock eventsub started");
    } else {
      this.listener.onChannelChatMessage(
        this.config.channelUserId,
        this.config.channelUserId,
        (e) => {
          if (!e.messageText.startsWith("!")) return;

          const [cmd] = e.messageText.toLowerCase().split(" ");
          log.info(
            {
              command: cmd,
              user: e.chatterDisplayName,
              message: e.messageText,
            },
            "chat command received"
          );

          switch (cmd) {
            case "!stats":
              this.handleStatsCommand(e.chatterId, e.chatterDisplayName);
              break;
            case "!addstat":
              this.handleModifyStatCommand(e.messageText, e, false);
              break;
            case "!rmstat":
              this.handleModifyStatCommand(e.messageText, e, true);
              break;
            case "!coobubu":
              this.handleShowCollectionCommand(
                "coobubu",
                e.chatterId,
                e.chatterDisplayName
              );
              break;
            case "!coobubu-redeem":
              this.handleRedeemBlindBoxCommand(
                "coobubu",
                e.chatterId,
                e.chatterDisplayName
              );
              break;
            case "!olliepop":
              this.handleShowCollectionCommand(
                "olliepops",
                e.chatterId,
                e.chatterDisplayName
              );
              break;
            case "!olliepop-redeem":
              this.handleRedeemBlindBoxCommand(
                "olliepops",
                e.chatterId,
                e.chatterDisplayName
              );
              break;
          }
        }
      );

      this.listener.onChannelRedemptionAdd(
        this.config.channelUserId,
        async (e) => {
          log.info(
            { reward: e.rewardTitle, user: e.userName, userId: e.userId },
            "channel point reward redeemed"
          );

          if (e.rewardTitle === "Cooper Series Blind Box") {
            await this.handleRedeemBlindBoxCommand(
              "coobubu",
              e.userId,
              e.userName
            );
          } else if (e.rewardTitle === "Ollie Series Blind Box") {
            await this.handleRedeemBlindBoxCommand(
              "olliepops",
              e.userId,
              e.userName
            );
          }

          if (e.rewardTitle === "Drink a Potion") {
            await this.handleDrinkPotionReward(e.userId, e.userName);
          } else if (e.rewardTitle === "Tempt the Dice") {
            await this.handleTemptDiceReward(e.userId, e.userName);
          }
        }
      );

      try {
        this.listener.start();
        log.info("eventsub started");
      } catch (err) {
        log.error({ err }, "eventsub start failed");
      }
    }

    log.info("bot ready");
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

  private async sendMessage(message: string) {
    await this.api.asUser(this.config.botUserId, async (ctx) => {
      await ctx.chat.sendChatMessage(this.config.channelUserId, message);
    });
    log.debug({ message }, "message sent");
  }

  private async handleStatsCommand(userId: string, username: string) {
    const stats = this.store.getStats(userId, username);
    await this.sendMessage(formatStats(username, stats));
    log.info({ userId, username }, "stats command handled");
  }

  private async handleModifyStatCommand(
    command: string,
    event: EventSubChannelChatMessageEvent,
    isRemove: boolean
  ) {
    const isMod = event.badges.moderator || event.badges.broadcaster;
    if (!isMod) {
      log.warn(
        { user: event.chatterDisplayName, command },
        "non-moderator attempted to use mod command"
      );
      await this.sendMessage("You must be a moderator to use this command");
      return;
    }

    const parsed = parseModifyStatCommand(command, isRemove);
    if (parsed.error) {
      log.warn({ command, error: parsed.error }, "invalid modify stat command");
      await this.sendMessage(parsed.error);
      return;
    }

    const { mentionedLogin, statColumn, amount } = parsed;

    this.store.modifyStat(
      event.chatterId,
      mentionedLogin,
      statColumn,
      isRemove ? -amount : amount
    );

    const stats = this.store.getStats(event.chatterId, mentionedLogin);
    await this.sendMessage(formatStats(mentionedLogin, stats));

    log.info(
      {
        moderator: event.chatterDisplayName,
        target: mentionedLogin,
        stat: statColumn,
        amount: isRemove ? -amount : amount,
      },
      "modify stat command handled"
    );
  }

  private randStat(): Stat {
    return statList[Math.floor(Math.random() * statList.length)];
  }

  private randDelta(): number {
    return Math.random() < 0.05 ? -1 : 1;
  }

  private async handleDrinkPotionReward(userId: string, username: string) {
    const stat = this.randStat();
    const delta = this.randDelta();
    const outcome = delta < 0 ? "lost" : "gained";

    this.store.modifyStat(userId, username, stat.column, delta);

    const message = `A shifty looking merchant hands ${username} a glittering potion. Without hesitation, they sink the whole drink. ${username} ${outcome} ${stat.display}`;
    await this.sendMessage(message);

    const stats = this.store.getStats(userId, username);
    await this.sendMessage(formatStats(username, stats));

    log.info(
      { userId, username, stat: stat.column, delta, outcome },
      "drink potion reward handled"
    );
  }

  private async handleTemptDiceReward(userId: string, username: string) {
    await this.sendMessage(`${username} has rolled with initiative.`);

    const stats = this.store.getStats(userId, username);
    await this.sendMessage(formatStats(username, stats));

    log.info({ userId, username }, "tempt dice reward handled");
  }

  private async handleShowCollectionCommand(
    type: CollectionType,
    userId: string,
    username: string
  ) {
    const collection = this.store.getUserCollections(userId, type);

    this.wsServer.broadcast({
      type: "collection_display",
      data: {
        userId,
        username,
        collectionType: type,
        collection: collection || [],
        collectionSize: collection?.length || 0,
      },
    });

    log.info(
      {
        userId,
        username,
        collectionType: type,
        collectionSize: collection?.length || 0,
      },
      "collection command handled"
    );
  }

  private async handleRedeemBlindBoxCommand(
    type: CollectionType,
    userId: string,
    username: string
  ) {
    const seriesConfig = blindBoxConfigs[type];
    const plushieWeights = seriesConfig.plushies;
    const seriesName = seriesConfig.rewardTitle;

    const plushieKey = getWeightedRandomPlushie(plushieWeights);

    const result = this.store.addPlushieToCollection(
      userId,
      username,
      type,
      plushieKey
    );

    // Find the plushie data for the redeemed item
    const plushieData = plushieWeights.find((p) => p.key === plushieKey);

    this.wsServer.broadcast({
      type: "blindbox_redemption",
      data: {
        userId,
        username,
        collectionType: type,
        seriesName,
        plushie: {
          key: plushieKey,
          name: plushieData?.name || "Unknown",
          weight: plushieData?.weight || 0,
        },
        isNew: result?.isNew || false,
        collectionSize: result?.collection.length || 0,
        collection: result?.collection || [],
      },
    });

    log.info(
      {
        userId,
        username,
        collectionType: type,
        reward: plushieKey,
        isNew: result?.isNew,
        collectionSize: result?.collection.length || 0,
      },
      "redeem handled"
    );
  }
}
