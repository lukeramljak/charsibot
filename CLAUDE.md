# Charsibot

Twitch bot and overlay for viewer stats and blind-box collections. Single-replica stateful app backed by SQLite.

## Commands

```
pnpm check        # svelte-kit sync + svelte-check (type checking)
pnpm lint         # prettier --check + eslint
pnpm test         # vitest run
pnpm build        # vite build (adapter-node) + server entrypoint bundle
pnpm format       # prettier --write
```

## Project layout

```
catalog/config/          JSON stat + blind-box definitions
server/                  Custom Node entrypoint (imports adapter-node handler)
src/
  lib/
    contracts/           Shared TS types: catalog.ts, viewer.ts, collections.ts, overlay.ts
    admin/
      admin.remote.ts    Remote functions (query/command) for admin panel
      types.ts           Admin-specific types (AdminUserDetail, GrantResult, ActivityFilter)
      *.svelte           Admin UI components
    server/
      application/       Ports (service interfaces) and error types
      bot/               Command, trigger, redemption dispatch + tests
      catalog/           JSON catalog loading + validation
      db/                SQLite client, schema, repositories
      domain/
        stats/           Stats service + formatting
        blind-box/       Blind-box service + weighted random selection
      runtime/
        container.ts     Module-level service registry (getServices/setServices)
        contracts.ts     Readiness, TwitchRuntime, ApplicationRuntime interfaces
        create.ts        Runtime factory (DB, catalog, services, bus, Twitch)
        lifecycle.ts     Idempotent start/stop lifecycle controller
      twitch/            Token, Helix, conduit, EventSub, WebSocket transport
      events/            Typed overlay event bus and SSE formatting
  routes/
    admin/               Admin page (uses remote functions, ssr=false)
    blind-box/           Overlay page
    events/              SSE streaming route
    health/              Liveness probe (200 OK)
    ready/               Readiness probe (database, catalog, Twitch)
    oauth/               Twitch OAuth start/callback pages
```

## Key interfaces

- **Service ports:** `src/lib/server/application/ports.ts` — `StatsService`, `BlindBoxService`, `ChatSender`, `OverlayBus`, `Clock`, `Random`, `Logger`
- **Service container:** `src/lib/server/runtime/container.ts` — `ApplicationServices`, `getServices()`, `setServices()`
- **Application errors:** `src/lib/server/application/errors.ts` — `ApplicationError` with typed codes
- **Runtime contracts:** `src/lib/server/runtime/contracts.ts` — `Readiness`, `TwitchRuntime`, `ApplicationRuntime`

## Style rules

- Arrow functions for all new TS/JS (constructors and accessors excepted).
- `$lib/...` alias for internal imports; `$catalog/...` for root catalog. No `../` ladders.
- No barrel files (`index.ts` that only re-export). Import directly from the defining module.
- Blank lines to separate guards, setup, transformation, and return phases.
- No comments unless the WHY is non-obvious.
- Prettier owns formatting. Run `pnpm format` before committing.

## Architecture decisions

- `@sveltejs/adapter-node` with a custom entrypoint (`server/index.ts`).
- `ssr = false` globally.
- Remote functions (`query`/`command` from `$app/server`) for admin panel; explicit HTTP routes for SSE, OAuth, health.
- `better-sqlite3` with WAL mode.
- Native SSE (`EventSource`) for overlay delivery — no realtime framework.
- App-token conduit for Twitch EventSub (not user-token listener).
- Catalog JSON is the runtime source of truth.
- Single process, single replica. No multi-tenant.
