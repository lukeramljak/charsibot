import { loadConfig } from "./config";
import { Store } from "./store";
import { Bot } from "./bot";

const main = async () => {
  const config = loadConfig();
  const store = new Store(config.dbPath);
  const bot = new Bot(config, store);
  await bot.init();
};

main().catch((err) => {
  console.error("Fatal error", err);
  process.exit(1);
});
