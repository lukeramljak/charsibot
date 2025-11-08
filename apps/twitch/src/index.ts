import { loadConfig } from './config';
import { Store } from './store';
import { Bot } from './bot';
import { Database } from 'bun:sqlite';

const main = async () => {
  const config = loadConfig();
  const sqlite = new Database(config.dbPath);
  const store = new Store(sqlite);
  const bot = new Bot(config, store);
  await bot.init();
};

main().catch(err => {
  console.error('Fatal error', err);
  process.exit(1);
});
