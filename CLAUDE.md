# Charsibot

Twitch bot and overlay for viewer stats and blind-box collections. Single-replica stateful app backed by SQLite.

**Active branch:** `codex/full-sveltekit` — migrating from Go + static SvelteKit to a single SvelteKit node process. See `MIGRATION.md` for the full plan and checklist.

## Migration status

Waves 0–4 are done (foundation, data/domain, Twitch/bot, admin remote functions, overlay/SSE/OAuth/health, runtime integration, deployment/CI). **Wave 5 is next:** controlled cutover. Read only the Wave 5 section of `MIGRATION.md` — earlier waves are complete.

## Commands

All commands run from `web/`.

```
pnpm check        # svelte-kit sync + svelte-check (type checking)
pnpm lint         # prettier --check + eslint
pnpm test         # vitest run (165 tests)
pnpm build        # vite build (adapter-node) + server entrypoint bundle
pnpm format       # prettier --write
```

## Project layout

```
catalog/config/          JSON stat + blind-box definitions (shared by Go and Node)
server/                  Go implementation (kept until Wave 6 removal)
web/                     SvelteKit application (all new work goes here)
  server/                Custom Node entrypoint (imports adapter-node handler)
  src/
    lib/
      contracts/         Shared TS types: catalog.ts, viewer.ts, collections.ts, overlay.ts
      admin/
        admin.remote.ts  Remote functions (query/command) for admin panel
        types.ts         Admin-specific types (AdminUserDetail, GrantResult, ActivityFilter)
        *.svelte         Admin UI components
      server/
        application/     Ports (service interfaces) and error types
        bot/             Command, trigger, redemption dispatch + tests
        catalog/         JSON catalog loading + validation
        db/              SQLite client, schema, repositories
        domain/
          stats/         Stats service + formatting
          blind-box/     Blind-box service + weighted random selection
        runtime/
          container.ts   Module-level service registry (getServices/setServices)
          contracts.ts   Readiness, TwitchRuntime, ApplicationRuntime interfaces
          create.ts      Runtime factory (DB, catalog, services, bus, Twitch)
          lifecycle.ts   Idempotent start/stop lifecycle controller
        twitch/          Token, Helix, conduit, EventSub, WebSocket transport
        events/          Typed overlay event bus and SSE formatting
    routes/
      admin/             Admin page (uses remote functions, ssr=false)
      blind-box/         Overlay page
      events/            SSE streaming route
      health/            Liveness probe (200 OK)
      ready/             Readiness probe (database, catalog, Twitch)
      oauth/             Twitch OAuth start/callback pages
```

## Key interfaces

- **Service ports:** `web/src/lib/server/application/ports.ts` — `StatsService`, `BlindBoxService`, `ChatSender`, `OverlayBus`, `Clock`, `Random`, `Logger`
- **Service container:** `web/src/lib/server/runtime/container.ts` — `ApplicationServices`, `getServices()`, `setServices()`
- **Application errors:** `web/src/lib/server/application/errors.ts` — `ApplicationError` with typed codes
- **Runtime contracts:** `web/src/lib/server/runtime/contracts.ts` — `Readiness`, `TwitchRuntime`, `ApplicationRuntime`

## Style rules

- Arrow functions for all new TS/JS (constructors and accessors excepted).
- `$lib/...` alias for internal imports; `$catalog/...` for root catalog. No `../` ladders.
- Blank lines to separate guards, setup, transformation, and return phases.
- No comments unless the WHY is non-obvious.
- Prettier owns formatting. Run `pnpm format` before committing.

## Architecture decisions

- `@sveltejs/adapter-node` with a custom entrypoint (`web/server/index.ts`).
- `ssr = false` globally — no SSR during migration.
- Remote functions (`query`/`command` from `$app/server`) for admin panel; explicit HTTP routes for SSE, OAuth, health.
- `better-sqlite3` with WAL mode, Goose v7 schema unchanged.
- Native SSE (`EventSource`) for overlay delivery — no realtime framework.
- App-token conduit for Twitch EventSub (not user-token listener).
- Catalog JSON is the runtime source of truth; Go and Node share the same catalog files.
- Single process, single replica. No multi-tenant.

## What NOT to change during the migration

- SQLite v7 schema (deferred until after rollback window).
- Catalog identifiers, stored viewer-state values, public routes, OBS URLs.
- Go implementation (stays in-tree until Wave 6).
