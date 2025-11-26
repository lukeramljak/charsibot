import { blindBoxConfigs } from '@/blind-box/configs';
import { Bot } from '@/bot/bot';
import { CompletedCollectionsCommand } from '@/commands/blind-box/completed-collections';
import { RedeemBlindBoxCommand } from '@/commands/blind-box/redeem-blind-box';
import { ShowBlindBoxCollectionCommand } from '@/commands/blind-box/show-blind-box-collection';
import type { Command } from '@/commands/command';
import { ModifyStatCommand } from '@/commands/stats/modify-stat';
import { StatsCommand } from '@/commands/stats/stats';
import { loadConfig } from '@/config';
import { CommandHandler } from '@/events/command-handler';
import { MessageHandler } from '@/events/message-handler';
import { RedemptionHandler } from '@/events/redemption-handler';
import { TriggerHandler } from '@/events/trigger-handler';
import { BlindBoxRedemption } from '@/redemptions/blind-box/redeem-blind-box';
import type { Redemption } from '@/redemptions/redemption';
import { PotionRedemption } from '@/redemptions/stats/potion';
import { TemptDiceRedemption } from '@/redemptions/stats/tempt-dice';
import { Store } from '@/storage/store';
import { ComeTrigger } from '@/triggers/come';
import type { Trigger } from '@/triggers/trigger';
import { Database } from 'bun:sqlite';

const createBlindBoxCommands = (): Command[] => {
  const commands: Command[] = [];

  for (const config of blindBoxConfigs) {
    commands.push(new RedeemBlindBoxCommand(config), new ShowBlindBoxCollectionCommand(config));
  }

  return commands;
};

const createBlindBoxRedemptions = (): Redemption[] => {
  const redemptions: Redemption[] = [];

  for (const config of blindBoxConfigs) {
    redemptions.push(new BlindBoxRedemption(config));
  }

  return redemptions;
};

const main = async () => {
  const config = loadConfig();
  const sqlite = new Database(config.dbPath);
  const store = new Store(sqlite);

  const commands: Command[] = [
    ...createBlindBoxCommands(),
    new CompletedCollectionsCommand(),
    new ModifyStatCommand(),
    new StatsCommand(),
  ];
  const redemptions: Redemption[] = [
    ...createBlindBoxRedemptions(),
    new TemptDiceRedemption(),
    new PotionRedemption(),
  ];
  const triggers: Trigger[] = [new ComeTrigger()];

  const commandHandler = new CommandHandler(commands);
  const triggerHandler = new TriggerHandler(triggers);
  const messageHandler = new MessageHandler(triggerHandler);
  const redemptionHandler = new RedemptionHandler(redemptions);

  const bot = new Bot({
    config,
    store,
    commandHandler,
    messageHandler,
    redemptionHandler,
  });
  await bot.init();
};

main().catch((err) => {
  console.error('Fatal error', err);
  process.exit(1);
});
