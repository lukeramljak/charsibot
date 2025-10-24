import { AccessTokenWithUserId, RefreshingAuthProvider } from "@twurple/auth";
import { ApiClient } from "@twurple/api";
import { EventSubWsConfig, EventSubWsListener } from "@twurple/eventsub-ws";
import { Config } from "./config";
import { Store, formatStats, statList, Stat } from "./store";
import { parseModifyStatCommand } from "./command-parser";
import type { EventSubChannelChatMessageEvent } from "@twurple/eventsub-base";
import { log } from "./logger";
import { MockEventSubListener } from "./mock-eventsub";

export class Bot {
  private api: ApiClient;
  private listener: EventSubWsListener;
  private mockListener?: MockEventSubListener;
  private store: Store;
  private authProvider: RefreshingAuthProvider;

  constructor(private config: Config, store: Store) {
    this.store = store;

    this.authProvider = new RefreshingAuthProvider({
      clientId: config.clientId,
      clientSecret: config.clientSecret,
    });

    this.authProvider.onRefresh(async (userId, tokenData) => {
      if (!tokenData.refreshToken) return;

      if (userId === this.config.channelUserId) {
        await this.store.saveTokens(
          "streamer",
          tokenData.accessToken,
          tokenData.refreshToken
        );
        log.info({ userId }, "streamer tokens refreshed");
      }

      if (userId === this.config.botUserId) {
        await this.store.saveTokens(
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
            await this.onDrinkPotion(data.userId, data.userName);
          } else if (data.rewardTitle === "Tempt the Dice") {
            await this.onTemptDice(data.userId, data.userName);
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
              this.handleStats(e.chatterId, e.chatterDisplayName);
              break;
            case "!addstat":
              this.handleModifyStat(e.messageText, e, false);
              break;
            case "!rmstat":
              this.handleModifyStat(e.messageText, e, true);
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

          if (e.rewardTitle === "Drink a Potion") {
            await this.onDrinkPotion(e.userId, e.userName);
          } else if (e.rewardTitle === "Tempt the Dice") {
            await this.onTemptDice(e.userId, e.userName);
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

  private async handleStats(userId: string, username: string) {
    const stats = await this.store.getStats(userId, username);
    await this.sendMessage(formatStats(username, stats));
    log.info({ userId, username }, "stats command handled");
  }

  private async handleModifyStat(
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

    await this.store.modifyStat(
      event.chatterId,
      mentionedLogin,
      statColumn,
      isRemove ? -amount : amount
    );

    const stats = await this.store.getStats(event.chatterId, mentionedLogin);
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

  private async onDrinkPotion(userId: string, userName: string) {
    const stat = this.randStat();
    const delta = this.randDelta();
    const outcome = delta < 0 ? "lost" : "gained";

    await this.store.modifyStat(userId, userName, stat.column, delta);

    await this.sendMessage(
      `A shifty looking merchant hands ${userName} a glittering potion. Without hesitation, they sink the whole drink. ${userName} ${outcome} ${stat.display}`
    );

    const stats = await this.store.getStats(userId, userName);
    await this.sendMessage(formatStats(userName, stats));

    log.info(
      { userId, userName, stat: stat.column, delta, outcome },
      "drink potion reward handled"
    );
  }

  private async onTemptDice(userId: string, userName: string) {
    await this.sendMessage(`${userName} has rolled with initiative.`);

    const stats = await this.store.getStats(userId, userName);
    await this.sendMessage(formatStats(userName, stats));

    log.info({ userId, userName }, "tempt dice reward handled");
  }
}
