import { Bot } from '@/bot/bot';
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

const main = async () => {
  const config = loadConfig();
  const sqlite = new Database(config.dbPath);
  const store = new Store(sqlite);

  const commands: Command[] = [
    new ModifyStatCommand(),
    new RedeemBlindBoxCommand('coobubu'),
    new RedeemBlindBoxCommand('olliepop'),
    new ShowBlindBoxCollectionCommand('coobubu'),
    new ShowBlindBoxCollectionCommand('olliepop'),
    new StatsCommand(),
  ];
  const redemptions: Redemption[] = [
    new BlindBoxRedemption('Cooper Series Blind Box'),
    new BlindBoxRedemption('Ollie Series Blind Box'),
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
