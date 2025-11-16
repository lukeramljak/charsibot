import 'dotenv/config';
import { defineConfig } from 'drizzle-kit';

export default defineConfig({
  out: './drizzle',
  schema: './src/storage/schema.ts',
  dialect: 'sqlite',
  casing: 'snake_case',
  dbCredentials: {
    url: process.env.DB_PATH!
  }
});
