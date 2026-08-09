# charsibot

[![Twitch Bot](https://github.com/lukeramljak/charsibot/actions/workflows/ci.yml/badge.svg)](https://github.com/lukeramljak/charsibot/actions/workflows/ci.yml)

Twitch bot and overlay for [Charsibel](https://twitch.tv/charsibel). It is a single-replica SvelteKit/Node application backed by SQLite.

## Prerequisites

- Node.js 22+
- pnpm 10+
- Docker and Docker Compose (optional, for container deployment)

## Run locally

1. Install dependencies and create your local configuration:

   ```bash
   pnpm install --frozen-lockfile
   cp .env.example .env
   ```

2. For local offline work, leave `TWITCH_MOCK_MODE=true`. To connect Twitch chat, set `TWITCH_MOCK_MODE=false` and all four `TWITCH_*` credentials. Set `TWITCH_OAUTH_REDIRECT_URI` only when using the OAuth pages.

3. Build and run the application:

   ```bash
   pnpm build
   set -a && source .env && set +a
   pnpm start
   ```

   The service listens on `http://localhost:8081` by default. `GET /health` is its liveness probe and `GET /ready` reports catalog, database, and Twitch readiness.

`pnpm dev` starts the Vite development server and the application runtime, including the database and Twitch bot when credentials are configured.

## Docker deployment

Create `.env` as above, then run:

```bash
docker compose up --build -d
```

The Compose configuration maps port `8081`, stores SQLite data in the `twitch-data` volume, and includes a daily backup sidecar. Before a production deployment, set `TWITCH_MOCK_MODE=false` and provide all four Twitch credentials; the application refuses to start otherwise. Run `docker compose down` to stop the stack; omit `-v` to preserve database data.

## Environment Variables

```bash
cp .env.example .env
```

`DB_PATH` is required. `TWITCH_MOCK_MODE=true` is an explicit offline mode for local development and CI. When it is `false` or omitted, all four Twitch credentials are required and startup fails if any are missing. `PORT`, `HOST`, and `SHUTDOWN_TIMEOUT` are optional. The included `.env` example and Docker Compose deployment set `PORT=8081`; adapter-node otherwise defaults to port `3000`, host `0.0.0.0`, and a 30-second shutdown timeout. `ORIGIN` defaults to `http://localhost:8081` in Docker; set it when accessing the app from another device or through a reverse proxy.

### Accessing the admin from another device

Set `ORIGIN` to the exact URL used to open the app in the browser. For a direct LAN deployment, for example:

```bash
ORIGIN=http://192.168.1.50:8081
```

Then restart the container with `docker compose up -d`. Use that same scheme, host, and port in the browser. This lets SvelteKit validate admin updates as same-origin requests without disabling CSRF protection. If the app is behind a TLS reverse proxy, set `ORIGIN` to its public `https://` URL instead.

## Quality checks

```bash
pnpm check
pnpm lint
pnpm test
pnpm build
```

## Database

Charsibot uses a SQLite database at `DB_PATH`. It is created and migrated automatically when absent or behind the current application version. Versioned SQL migrations live in `drizzle/`. Use `pnpm db:generate` after changing the Drizzle schema, review the generated SQL, and commit it with the schema change.

## Catalog config

Stat definitions and blind-box series are versioned JSON files under `catalog/config`.
SQLite stores only viewer state (stat values and collected plushies).
Catalog JSON is the runtime source of truth.

Blind-box images and sounds live under `static/assets/blind-box/<series>/`.
JSON files use filenames such as `cutey.png` and the app expands them to public paths like `/assets/blind-box/coobubu/cutey.png`.
