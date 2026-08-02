# Full SvelteKit migration

Status: in progress  
Baseline: `b9c73b1` (`main`, admin panel merged)  
Implementation branch: `codex/full-sveltekit`  
Checklist owner: integration lead

## Goal

Replace the Go binary, Go HTTP server, generated admin API, and static SvelteKit build with one long-lived SvelteKit application running on `@sveltejs/adapter-node`.

The finished application must preserve the current Twitch bot, overlay, admin panel, SQLite data, deployment, and backup behavior. It remains a stateful, single-replica application.

## Fixed architectural decisions

- [ ] Build a modular monolith: application use-cases in the center, with SvelteKit/admin, Twitch, SQLite, and SSE as adapters.
- [ ] Preserve external behavior and stored data, not the current Go package structure or incidental implementation quirks.
- [ ] Keep the application under `web/` during the rewrite; consider moving it to the repository root only after Go removal.
- [ ] Use `@sveltejs/adapter-node` and one explicit production process.
- [ ] Use a custom Node entrypoint for deterministic startup and shutdown.
- [ ] Keep `ssr = false` initially to avoid unrelated overlay/admin UI churn.
- [ ] Use SvelteKit remote functions for the admin panel.
- [ ] Keep explicit HTTP routes for SSE, OAuth, health, and readiness.
- [ ] Present OAuth as a thin SvelteKit page, with a server load/endpoint performing Twitch redirects and callback exchange.
- [ ] Use a browser-native `EventSource` plus a SvelteKit streaming response for overlay delivery; do not add a realtime framework.
- [ ] Do not replace the ordered overlay SSE stream with `query.live`.
- [ ] Use `better-sqlite3`, prepared statements, and explicit transactions.
- [ ] Keep the existing final v7 SQLite schema unchanged for the first Node release.
- [ ] Use direct Twitch REST and WebSocket clients with app-token conduit semantics; do not substitute a user-token listener abstraction.
- [ ] Keep catalog JSON as the runtime source of truth.
- [ ] Keep the Go implementation until all parity gates pass.
- [ ] Never run the Go and Node bots concurrently against Twitch or the same database.
- [ ] Pin production to exactly one replica.
- [ ] Keep a narrow explicit HTTP boundary for future Twitch Extension clients; remote functions remain the admin-only transport.

Remote functions and async Svelte are currently experimental. Pin Svelte/SvelteKit versions for the migration and do not perform dependency upgrades during the parity phase.

References:

- <https://svelte.dev/docs/kit/remote-functions>
- <https://svelte.dev/docs/kit/adapter-node>
- <https://dev.twitch.tv/docs/eventsub/handling-websocket-events>
- <https://dev.twitch.tv/docs/eventsub/handling-conduit-events>

## Target architecture

```text
Svelte pages/components
        |
admin remote functions       Twitch EventSub handlers
        |                              |
        +-------- application use-cases --------+
                           |
          stats / collections / viewers domain
                   |                 |
             SQLite adapter     typed overlay bus
                                      |
                                 /events SSE
```

Suggested server layout:

```text
web/src/lib/server/
  application/       shared admin/Twitch use-cases and orchestration
  domain/            pure stats, blind-box, viewer types and rules
  db/                SQLite connection, compatibility checks, repositories
  catalog/           JSON loading and validation
  twitch/            token, Helix, conduit and EventSub transport
  bot/               command, trigger and redemption dispatch
  events/            typed process-local event bus
  runtime/           configuration, dependency container and lifecycle
```

Remote functions and Twitch handlers must call the same application use-cases. They must not duplicate SQL, chat formatting, overlay construction, or grant/reset behavior.

## Future Twitch Extension boundary

The planned viewer collection Extension does not change the modular-monolith or overlay decisions, but it means the application must retain a small authenticated HTTP surface:

```text
Twitch Extension static client -> Extension JWT endpoint -> collection use-case -> SQLite
Admin SvelteKit client          -> remote function         -> same collection use-case
OBS overlay                     -> SSE                     -> application event bus
```

- [ ] Build the Extension as a separate static Svelte client that can be uploaded to Twitch's CDN.
- [ ] Share catalog, collection DTOs, and presentation components with the main SvelteKit application where practical.
- [ ] Add a read-only endpoint such as `GET /api/extension/me/collections`.
- [ ] Verify the Twitch-signed Extension JWT, expiry, role, and configured `channel_id` server-side.
- [ ] Derive the numeric Twitch viewer ID only from a verified identity-linked JWT; never accept an arbitrary viewer ID from the client.
- [ ] Render a clear identity-sharing prompt for viewers who have not linked their Twitch identity.
- [ ] Restrict CORS to the submitted Extension origin and keep the Extension shared secret server-only.
- [ ] Start with request/response collection loading; add Twitch Extension PubSub invalidation only if live refresh is later required.
- [ ] Keep the first release explicitly single-channel. If multi-channel installation becomes a goal, add broadcaster tenancy as a deliberate later schema migration.

## Rearchitecture policy

### Improve during the rewrite

These changes reduce risk or remove duplicated behavior without breaking rollback compatibility:

- [ ] Make multi-row mutations transactional.
- [ ] Add Twitch notification deduplication by message ID.
- [ ] Add app-token expiry handling and one authorization refresh/retry.
- [ ] Make every Twitch endpoint injectable so mock mode is fully offline.
- [ ] Add explicit readiness separate from process liveness.
- [ ] Replace client-side repeated deletion requests with one transactional bulk-delete use-case.
- [ ] Define deterministic ordering/tie-breaking where current behavior is unspecified, then snapshot the chosen behavior.
- [ ] Centralize exact chat messages and overlay payload construction in application use-cases.
- [ ] Validate environment, Twitch payloads, remote-function inputs, and catalog data at their boundaries.
- [ ] Track background tasks and make shutdown cancellation explicit.

### Defer until after the rollback window

These changes expand the failure surface or make the Go rollback incompatible:

- [ ] Redesigning the SQLite schema or replacing Goose version 7.
- [ ] Replacing the app-token conduit transport with a different Twitch authorization model.
- [ ] Supporting multiple application replicas or moving SQLite to a network database.
- [ ] Replacing localhost-only admin access with a new authentication system.
- [ ] Enabling SSR throughout the overlay/admin UI.
- [ ] Moving `web/` to the repository root.
- [ ] Changing public routes, OBS URLs, catalog identifiers, or stored values.

### Simplification test

Before introducing an abstraction, it must either:

1. allow admin and Twitch to share a use-case,
2. isolate an external system for deterministic testing, or
3. own application lifecycle/resource cleanup.

Avoid framework-style repository/service/interface layers that have only one caller and provide no testing or boundary benefit.

### TypeScript function style

- Use arrow-function syntax for all new TypeScript and JavaScript functions, including named helpers, handlers, closures, and class fields.
- Express callable interface members as function properties rather than method signatures.
- Convert existing function declarations when their file is otherwise being changed; do not create a standalone whole-repository style rewrite.
- Constructors and language-level accessors are the only syntax exceptions because JavaScript does not provide arrow equivalents.

## Overlay transport decision

The overlay listens to processed application events, not directly to Twitch:

- Twitch redemption notifications do not contain the final weighted selection, duplicate status, updated collection, or catalog-expanded overlay payload.
- Collection display commands require database state.
- Admin-triggered grants and display actions have no Twitch event to observe.
- Browser code must not receive the Twitch client secret or app access token, and it cannot share the backend conduit WebSocket.

Therefore the server remains the authority that consumes Twitch, runs application use-cases, updates SQLite, and publishes typed overlay events.

### Selected: native SSE

- One-way server-to-overlay delivery matches the requirement exactly.
- `EventSource` provides browser/OBS reconnection without a client dependency.
- SvelteKit can return the stream from a normal `+server.ts` route.
- A small typed in-process bus keeps transport concerns out of bot/admin use-cases.
- The existing event names and overlay client can remain stable.

### Not selected: SvelteKit `query.live`

Remote functions still execute on the server and use a generated network endpoint, so this would not make the overlay listen to Twitch directly. Live queries retain only the latest pending value when a consumer falls behind and explicitly are not event logs. Charsibot must not collapse two rapid redemptions into one. Making `query.live` reliable here would require adding sequence numbers, replay buffers, cursors, and acknowledgement semantics.

### Not selected: `svelte-realtime`

The library is a reasonable future choice for bidirectional RPC, collaborative state, presence, many realtime topics, or very high connection counts. For the current single broadcast stream it adds more architecture than it removes:

- It replaces `@sveltejs/adapter-node` with `svelte-adapter-uws`.
- It adds uWebSockets.js as a native C++ dependency installed from GitHub.
- It requires WebSocket hooks and multiple Vite plugins, plus a separate development WebSocket implementation.
- Its documented Docker runtime requires a newer glibc image than the current Bookworm base.
- Core streams do not automatically provide the ordered durable event log Charsibot would need; replay is an additional feature.

Revisit this decision only if Charsibot gains substantial bidirectional realtime features beyond the OBS overlay.

## Definition of done

- [ ] One Node process serves the web UI, admin remote functions, OAuth, health/readiness, and SSE while running the Twitch bot.
- [ ] A copied production v7 database opens without schema or data changes.
- [ ] A fresh database is compatible with both the Node app and the final Go binary.
- [ ] The admin panel provides every current operation and remains local-only.
- [ ] Overlay URLs, assets, event names, payloads, ordering, and reconnect behavior remain compatible with OBS.
- [ ] Twitch commands, triggers, redemptions, chat output, reconnects, and shutdown pass mock and live validation.
- [ ] Docker deployment, persistent volume, health check, and backup/restore work.
- [ ] The final Go image can be restarted against the unchanged database throughout the observation window.
- [ ] Go, sqlc, Huma, OpenAPI generation, and obsolete CI are removed only after the live soak.

## Execution model

This is one rewrite branch, but not one giant commit. Work proceeds in waves.

### Ownership rules

- The integration lead alone edits this file, shared contracts, package configuration, lockfiles, runtime wiring, entrypoints, Docker, Compose, Taskfile, README, and CI.
- Agents do not edit outside their assigned paths without handing the work back to the integration lead.
- Shared interfaces are frozen before parallel implementation begins.
- Every workstream lands with tests and a checklist update.
- Deleting old code is a final integration task, not an agent task.

### Agent allocation

| Owner             | Exclusive paths                                                                     | Responsibility                                                       |
| ----------------- | ----------------------------------------------------------------------------------- | -------------------------------------------------------------------- |
| Integration lead  | shared config, contracts, runtime/container, event bus, explicit routes, deployment | Foundation, integration, lifecycle, SSE/OAuth/health, CI and cutover |
| Data/domain agent | `web/src/lib/server/db/**`, `catalog/**`, `domain/stats/**`, `domain/blind-box/**`  | SQLite, catalog, repositories, domain services and parity tests      |
| Twitch agent      | `web/src/lib/server/twitch/**`, `bot/**`                                            | Helix, conduit WebSocket, commands, triggers, redemptions and tests  |
| Admin agent       | `web/src/lib/admin/**`, `web/src/routes/admin/**`                                   | Remote functions, admin UI conversion and admin tests                |

## Wave 0: freeze the baseline

- [x] Create `codex/full-sveltekit` from `b9c73b1`.
- [x] Record successful baseline commands and tool versions.
- [x] Run `go test ./...`.
- [x] Run `golangci-lint run`.
- [x] Resolve the current web check/test hang and record clean `pnpm` checks.
- [ ] Build the existing Docker image.
- [ ] Retain/tag the last known-good Go image and commit.
- [ ] Capture the current `web/openapi.json` as an admin contract fixture.
- [ ] Capture representative SSE frames for all three named event types.
- [ ] Capture exact chat output for every command, trigger, and redemption.
- [ ] Create a sanitized v7 production database fixture.
- [ ] Restore-test the current production backup.
- [ ] Confirm production has run merged Go `main` and is at Goose version 7.

Gate: the merged Go application is green, deployable, backed up, and provides committed parity fixtures.

Baseline recorded 2026-08-02:

- Node `26.5.0`; project package manager pnpm `10.15.1`.
- SvelteKit `2.70.2` and Svelte `5.56.8`, pinned exactly for the migration.
- Go `1.26.5`; golangci-lint `2.12.2`.
- `go test ./...` passes and `golangci-lint run` reports zero issues.
- Web Prettier, ESLint, Svelte check, Vitest, and the static baseline build pass.
- The experimental native TypeScript checker was removed because it hung before producing diagnostics; stable `svelte-check` is clean.
- The initial Node checkpoint has 24 passing Vitest tests, including lifecycle and server configuration coverage.

## Wave 1: lead-owned foundation

### Dependencies and configuration

- [x] Replace `@sveltejs/adapter-static` with `@sveltejs/adapter-node`.
- [x] Add and pin `better-sqlite3`, its types, `ws`, its types, and one Standard Schema validator.
- [x] Enable `kit.experimental.remoteFunctions` and `compilerOptions.experimental.async`.
- [ ] Add server-side Vitest configuration and fixture helpers.
- [x] Add production and development scripts without removing the Go scripts yet.
- [ ] Define how development/HMR prevents duplicate database and bot runtimes.

### Shared contracts

- [x] Freeze TypeScript types/interfaces for catalog definitions and overlay events.
- [x] Freeze `StatsService` and `BlindBoxService` interfaces.
- [x] Freeze `OverlayBus`, `ChatSender`, `Clock`, `Random`, logger, and readiness interfaces.
- [x] Freeze Twitch runtime start/stop and readiness interfaces.
- [x] Define typed application errors used by remote functions and routes.
- [ ] Keep server-only dependencies under `$lib/server`; do not store request/user state in global reactive Svelte modules.

### Process lifecycle

- [x] Add a custom Node entrypoint importing `build/handler.js`.
- [x] Implement idempotent `startPromise` and `stopPromise` lifecycle control.
- [ ] Implement startup order: validate config -> database -> catalog -> services/event bus -> Twitch -> HTTP listen.
- [ ] Implement shutdown order: stop accepting HTTP/SSE -> abort Twitch timers/sockets -> await handlers -> close SQLite.
- [x] Implement `PORT`, `HOST`, and shutdown-timeout behavior explicitly; adapter-node continues to own its documented trusted-proxy variables.
- [ ] Fail startup on invalid configuration, database schema, or catalog.
- [ ] Enforce/document the single-process constraint.

Gate: adapter-node builds and serves the unchanged UI; shared interfaces compile; an empty runtime starts and shuts down exactly once.

## Wave 2A: data and domain

Exclusive ownership: `web/src/lib/server/db/**`, `catalog/**`, `domain/stats/**`, `domain/blind-box/**`, and colocated tests.

### SQLite compatibility

- [ ] Open one SQLite connection at `DB_PATH`.
- [ ] Apply `foreign_keys=ON`, `journal_mode=WAL`, `page_size=4096`, `cache_size=-8000`, `synchronous=NORMAL`, `secure_delete=ON`, and `busy_timeout=30000`.
- [ ] For an existing database, require Goose version exactly 7 and validate all required tables, columns, primary keys, and indexes without modifying schema/history.
- [ ] Fail closed for Goose versions below/above 7, malformed history, and partial schemas.
- [ ] For an empty database, create the consolidated final v7 schema in one transaction.
- [ ] Reproduce the Goose metadata table and applied version rows 0 through 7 for fresh databases.
- [ ] Do not replay historical migrations 1 through 7 and do not introduce v8.
- [ ] Prove the final Go binary can reopen a Node-created fresh database.
- [ ] Use transactions for multi-row stats initialization/reset and viewer deletion.
- [ ] Checkpoint/close WAL cleanly during shutdown and backup preparation.

### Catalog

- [ ] Port strict JSON decoding, including unknown-field and trailing-value rejection.
- [ ] Port validation for required fields, positive weights, uniqueness, and non-empty catalogs.
- [ ] Preserve stat/plushie `sortOrder` ordering and alphabetic series ordering.
- [ ] Preserve absolute asset paths and relative `/assets/blind-box/<assetDir>/...` expansion.
- [ ] Preserve the `olliepop` -> `olliepops` asset-directory distinction.

### Repositories and services

- [ ] Port viewer activity upsert, union listing, case-insensitive ordering, lookup, and transactional deletion.
- [ ] Preserve UTC RFC3339-compatible activity timestamps.
- [ ] Port stats create/read/set/adjust/reset, username synchronization, formatting, and leaderboard behavior.
- [ ] Port blind-box grant/duplicate username sync/remove/reset/completion and weighted selection.
- [ ] Inject clock and RNG; do not use global time/randomness in tests.
- [ ] Treat collection keys as membership; do not accidentally depend on current unspecified SQL order.
- [ ] Keep event broadcasting and chat sending outside domain services.

### Data/domain gate

- [ ] Fresh DB and copied v7 DB tests pass.
- [ ] Invalid schema/version rejection tests pass.
- [ ] Every connection pragma is asserted.
- [ ] Go-vs-TypeScript golden outputs match for viewers, stats, leaderboard, formatting, collections, duplicates, and completion.
- [ ] Weighted-selection boundary tests are deterministic and exact.
- [ ] JavaScript number safety for SQLite integer values is explicitly tested or bounded.

## Wave 2B: Twitch and bot runtime

Exclusive ownership: `web/src/lib/server/twitch/**`, `bot/**`, and their tests/fixtures.

### Helix and conduit

- [ ] Port app-token acquisition with expiry caching and one refresh/retry on authorization failure.
- [ ] Port authenticated Helix requests and chat sending with optional reply parent ID.
- [ ] List conduits and use the first, or create one one-shard conduit.
- [ ] Connect to `wss://eventsub.wss.twitch.tv/ws`.
- [ ] On welcome, assign shard `0` within the Twitch deadline.
- [ ] Ensure the four current subscriptions and treat HTTP 409 as success:
  - [ ] `channel.chat.message` v1.
  - [ ] `channel.channel_points_custom_reward_redemption.add` v1.
  - [ ] `channel.raid` v1.
  - [ ] `conduit.shard.disabled` v1.
- [ ] Parse welcome, notification, keepalive, reconnect, revocation, and shard-disabled messages with runtime validation.
- [ ] Implement seamless reconnect handoff using Twitch's supplied reconnect URL.
- [ ] Implement watchdog expiry and abortable 10-second ordinary reconnect delay.
- [ ] Let the WebSocket library handle ping/pong; send no application frames.
- [ ] Add bounded TTL deduplication by Twitch `metadata.message_id` before dispatch.
- [ ] Make token, REST, and WebSocket base URLs injectable in mock mode.

### Bot behavior

- [ ] Ignore messages from the configured bot user.
- [ ] Record activity before commands/triggers and for all redemptions, including unknown rewards.
- [ ] Preserve concurrent handler execution with a 10-second cancellation deadline.
- [ ] Port exact parsing and behavior for `!collections`, `!leaderboard`, `!stats`, and all catalog commands.
- [ ] Port the standalone-word `come`/`coming`/`cum`/`came` trigger, `no coming` exclusion, and current probability boundary semantics.
- [ ] Port `Drink a Potion`, `Tempt the Dice`, all catalog redemptions, and exact failure messages.
- [ ] Port the five-second raid shoutout delay.
- [ ] Preserve exact chat strings, reply IDs, overlay payloads, and case sensitivity.
- [ ] Track all handler/raid tasks so shutdown can await or cancel them.

### Twitch gate

- [ ] Command, trigger, redemption, and chat payload snapshot tests pass.
- [ ] Token/Helix request and error tests pass.
- [ ] Fake-WebSocket tests cover welcome, keepalive timeout, reconnect handoff, unexpected close, revocation, disabled shard, deduplication, and shutdown.
- [ ] Mock mode contacts no real Twitch endpoint.
- [ ] Recorded EventSub fixtures pass.

## Wave 2C: admin remote functions and UI

Exclusive ownership: `web/src/lib/admin/**`, `web/src/routes/admin/**`, and admin tests.

### Remote functions

- [ ] Add `admin.remote.ts` outside `$lib/server`, importing only server-side services from `$lib/server`.
- [ ] Use `query` for viewer listing and viewer detail.
- [ ] Use `command` for mutations and imperative overlay/chat actions.
- [ ] Validate every argument with Standard Schema.
- [ ] Implement `requireLocalAdmin` using `getRequestEvent().getClientAddress()` inside every remote boundary.
- [ ] Do not authorize with client-manipulable route/URL information.
- [ ] Do not trust forwarding headers unless proxy topology is explicitly configured.
- [ ] Add one transactional bulk-delete command rather than looping client requests, while preserving clear failure reporting.
- [ ] Return refreshed user data from mutations where the current UI expects it.

### Operation parity

- [ ] List users.
- [ ] Get user detail.
- [ ] Delete one user.
- [ ] Delete selected users.
- [ ] Set/adjust a stat.
- [ ] Display stats in chat.
- [ ] Grant a random stat.
- [ ] Reset stats.
- [ ] Explode and undo explode.
- [ ] Grant a random plushie.
- [ ] Grant/remove a selected plushie.
- [ ] Display a collection overlay.
- [ ] Reset a collection.

### UI conversion

- [ ] Replace `openapi-fetch`, generated schema imports, `readJSON`, `ensureSuccess`, and API calls with remote functions and domain types.
- [ ] Preserve URL-selected viewers and stale-request suppression.
- [ ] Preserve filters, command palette, dialogs, focus behavior, loading/status/error announcements, and responsive layout.
- [ ] Preserve chat/overlay option choices and grant result messages.
- [ ] Keep existing local rune state and imperative event-handler flow initially.
- [ ] Run the Svelte autofixer on every touched component.

### Admin gate

- [ ] All current operations have remote-function coverage.
- [ ] Local and non-local address authorization tests pass.
- [ ] UI smoke/E2E tests cover select, edit, grant, display, reset, delete, and bulk delete.
- [ ] No browser bundle contains DB clients, credentials, or server-only code.

## Wave 3: lead-owned integration and explicit routes

### Event bus and SSE

- [ ] Implement the process-local overlay event bus with per-client capacity 10.
- [ ] Add `/events/+server.ts` as a streaming SSE response.
- [ ] Preserve named events: `chat_command`, `blindbox_display`, and `blindbox_redemption`.
- [ ] Preserve exact JSON payloads and add snapshot tests.
- [ ] Send an initial heartbeat and 30-second comment heartbeats.
- [ ] Clean up disconnected clients and timers.
- [ ] Preserve the current overflow policy: drop for a full client without blocking the bot.
- [ ] Preserve EventSource reconnection behavior and overlay FIFO/priority tests.

### OAuth and health

- [ ] Add a thin `/oauth` SvelteKit page explaining streamer and bot authorization, with one action/link for each account.
- [ ] Add `/oauth/start` server handling with exact account validation, streamer/bot scopes, and Twitch redirect.
- [ ] Implement `/oauth/callback` as a server-rendered page flow with current state validation, token exchange, error behavior, and intentionally unchanged non-persistence semantics.
- [ ] Render clear success/denial/error states in the callback page instead of returning plain text.
- [ ] Keep Twitch OAuth secrets and token exchange entirely server-side.
- [ ] Preserve `/health` as process liveness.
- [ ] Add `/ready` for database/catalog/Twitch readiness without changing the existing health contract.

### Runtime integration

- [ ] Wire config, database, catalog, services, bus, Twitch, remote functions, and routes through one application container.
- [ ] Verify startup fails atomically and partially created resources close.
- [ ] Verify shutdown during a request, SSE connection, reconnect delay, redemption, and raid delay.
- [ ] Verify only one runtime starts in production, tests, and development/HMR.

Gate: the full Node application passes automated parity tests while the Go reference remains in-tree.

## Wave 4: deployment and CI

- [ ] Replace the Go/static Docker build with pnpm build plus production dependencies and the native SQLite runtime.
- [ ] Keep Debian/glibc-compatible builder and runtime images for `better-sqlite3`.
- [ ] Preserve port 8081 mapping, `/data`, `DB_PATH`, the named volume, backup label, and backup container.
- [ ] Configure Docker for exactly one application replica.
- [ ] Update `.env.example`, Taskfile, README, and operational documentation.
- [ ] Add CI for format, lint, Svelte check, unit tests, build, DB compatibility fixtures, mock Twitch integration, and Docker smoke.
- [ ] Add admin authorization and SSE contract tests to CI.
- [ ] Keep Go build/test/lint and API drift checks until the final parity gate.
- [ ] Test backup and restore with an active WAL database.

## Wave 5: controlled cutover

- [ ] Build and retain the final Go rollback image.
- [ ] Stop Go and take a verified database/volume snapshot.
- [ ] Start Node against a cloned production database first.
- [ ] Validate row counts, schema version, health/readiness, admin, SSE, and mock Twitch behavior.
- [ ] Start Node against production only after the clone passes.
- [ ] Validate inbound chat, threaded `!stats`, chatbot identity, each redemption type, overlay rendering, forced reconnect, and clean shutdown in a controlled channel.
- [ ] Run a live soak while monitoring reconnects, dropped SSE events, SQLite busy errors, memory, and event-loop delay.
- [ ] Roll back by stopping Node and restarting the final Go image against the unchanged v7 database.
- [ ] Keep schema unchanged and the Go image available for the observation window.

## Wave 6: removal and cleanup

- [ ] Confirm all definition-of-done and live-soak gates.
- [ ] Remove Go packages, commands, generated sqlc files, migrations replay tooling, `go.mod`, and `go.sum`.
- [ ] Remove Huma/OpenAPI generation, generated admin client, and obsolete API drift checks.
- [ ] Remove Go build/lint/test tasks and CI jobs.
- [ ] Remove Vite's Go API/SSE proxy.
- [ ] Remove the obsolete Go/static Docker stages.
- [ ] Decide whether to move `web/` to the repository root as a separate mechanical change.
- [ ] Run the complete Node CI and Docker backup/restore smoke test.
- [ ] Update this document to completed and record the final Node release/image.

## Rollback invariants

- The first Node release does not alter the v7 application schema.
- `goose_db_version` remains compatible with the final Go binary.
- Catalog identifiers and stored viewer-state values remain unchanged.
- The last Go image and a tested database backup remain available.
- Only one bot process is active at any time.

## Known high-risk areas

- Twitch seamless reconnect, shard-disabled recovery, token expiry, and duplicate delivery.
- Existing production databases that have not reached Goose v7.
- Native SQLite packaging and WAL-aware backup/restore.
- JavaScript number precision relative to SQLite/Go `int64`.
- Admin loopback semantics through Docker or a reverse proxy.
- Custom-server and development/HMR double starts.
- Synchronous SQLite work blocking the Node event loop.
- Parallel agents editing shared configuration or contracts.
