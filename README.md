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

2. Set `DB_PATH` in `.env` and, when connecting Twitch chat, set all four `TWITCH_*` credentials. Set `TWITCH_OAUTH_REDIRECT_URI` only when using the OAuth pages.

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

The Compose configuration maps port `8081`, stores SQLite data in the `twitch-data` volume, and includes a daily backup sidecar. Run `docker compose down` to stop the stack; omit `-v` to preserve database data.

## Environment Variables

```bash
cp .env.example .env
```

`DB_PATH` is required. `PORT`, `HOST`, and `SHUTDOWN_TIMEOUT` are optional. The included `.env` example and Docker Compose deployment set `PORT=8081`; adapter-node otherwise defaults to port `3000`, host `0.0.0.0`, and a 30-second shutdown timeout.

## Quality checks

```bash
pnpm check
pnpm lint
pnpm test
pnpm build
```

## Database

Charsibot uses a SQLite database at `DB_PATH`; it is created automatically when absent. The application requires the existing Goose v7 schema and does not alter migration history.

## Catalog config

Stat definitions and blind-box series are versioned JSON files under `catalog/config`.
SQLite stores only viewer state (stat values and collected plushies).
Catalog JSON is the runtime source of truth.

Blind-box images and sounds live under `static/assets/blind-box/<series>/`.
JSON files use filenames such as `cutey.png` and the app expands them to public paths like `/assets/blind-box/coobubu/cutey.png`.
