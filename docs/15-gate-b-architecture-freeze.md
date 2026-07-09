# Gate B Architecture Freeze

Status: Passed for Wave 1 skeleton only. Not passed for production data writes, feature repositories, command registration apply, index creation, or feature parity work.

Date: 2026-07-03

## 1. Final Target File Tree

Baseline tree for implementation:

```txt
MHCAT-REFACTOR/
  cmd/
    mhcat-bot/
    mhcat-command-sync/
    mhcat-mongo-audit/
    mhcat-mongo-repair/
  internal/
    app/
    config/
    discord/
      session/
      commands/
      events/
      components/
      modals/
      responses/
      permissions/
    core/
      domain/
      services/
      ports/
    adapters/
      discordgo/
      mongo/
        documents/
        indexes/
        repositories/
      cache/
      external/
    jobs/
    observability/
    security/
    errors/
  docs/
  testdata/
  tools/
  README.md
  .env.example
  go.mod
  go.sum
  Makefile
```

Allowed adjustment:

- `tools/mongo-audit-readonly.mjs` exists as a temporary read-only Phase 1.5 audit helper. Wave 3 may replace it with `cmd/mhcat-mongo-audit` after the Go module exists.

## 2. Package Boundary Rules

### DiscordGo Imports

May import `github.com/bwmarrin/discordgo`:

- `internal/adapters/discordgo/**`
- `internal/discord/session/**`
- thin Discord infrastructure packages under `internal/discord/**` only where they adapt DiscordGo payloads
- `cmd/mhcat-bot` wiring only
- `cmd/mhcat-command-sync` Discord command sync adapter only

Must not import DiscordGo:

- `internal/core/**`
- `internal/core/domain/**`
- `internal/core/services/**`
- `internal/core/ports/**`
- `internal/adapters/mongo/**`
- repository tests except adapter-level fakes that explicitly test Discord adapter behavior

### MongoDB Driver Imports

May import the official MongoDB Go Driver:

- `internal/adapters/mongo/**`
- `cmd/mhcat-mongo-audit`
- `cmd/mhcat-mongo-repair`
- `cmd/mhcat-bot` wiring only if unavoidable for dependency construction

Must not import MongoDB driver:

- `internal/core/**`
- `internal/discord/**`
- Discord handlers/routers except through repository ports

### Business Logic

May contain business logic:

- `internal/core/services/**`
- feature-specific core packages under `internal/core/**`
- pure domain helpers under `internal/core/domain/**`

Must not contain business logic:

- `cmd/**`
- `internal/adapters/**`
- `internal/discord/session/**`
- `internal/observability/**`
- `internal/config/**`

Discord handlers stay thin: parse DTO, authorize/defer, call service, format response.

### Config Access

May access parsed config:

- `cmd/**`
- `internal/app/**`
- `internal/config/**`
- infrastructure constructors

Core services receive explicit dependencies and scalar options. They do not read environment variables.

### Logging

May log:

- `cmd/**`
- `internal/app/**`
- `internal/discord/**`
- `internal/adapters/**`
- `internal/jobs/**`
- `internal/observability/**`

Core services may accept a narrow logger port only where useful. Logs must redact tokens, webhook URLs, Mongo URIs, access tokens, raw stack traces in user-facing paths, and high-risk user content samples.

### Goroutines

May spawn goroutines:

- `internal/app/**` lifecycle orchestration
- `internal/discord/session/**` Discord session lifecycle
- `internal/jobs/**` scheduler/workers
- adapter internals with explicit context cancellation

Goroutines must have:

- parent `context.Context`;
- shutdown path;
- bounded concurrency or ownership;
- panic recovery where they own a process boundary.

Handlers must not start unbounded goroutines per event.

### Command Registration

May register commands:

- `cmd/mhcat-command-sync` only, or a single explicitly configured leader mode after ADR.

Must not register commands:

- shard `Ready` handler;
- every `cmd/mhcat-bot` process by default;
- feature handlers.

All registration is diff-based. Deletes require dry-run plus explicit apply/delete flag.

### Mongo Index Writes

May write indexes:

- `cmd/mhcat-mongo-audit` only if it later grows an explicit `ensure-indexes --apply` command, or a dedicated operational command.

Must not write indexes:

- default bot startup;
- shard ready;
- repository constructors;
- feature handlers.

Safe startup may compare planned vs existing indexes only if read-only.

## 3. Implementation Waves

### Wave 1: Skeleton and Infrastructure Shell

Scope:

- Go module.
- Config loading/validation with legacy aliases.
- Secret redaction helpers.
- Logger.
- App lifecycle and shutdown.
- Mongo connection and ping only.
- Discord session shell and minimal intent builder.
- No command registration.
- No feature logic.
- No repository writes.
- No index writes.

Acceptance:

- `go fmt ./...`
- `go test ./...`
- `go vet ./...`
- `go run ./cmd/mhcat-bot` fails safely when required env is missing.
- Hardcoded secret scan passes for Go files.

### Wave 2: Command Sync and Interaction Shell

Scope:

- Typed command registry model.
- `cmd/mhcat-command-sync`.
- Diff-based dry-run/apply design.
- Interaction responder state machine.
- Router interfaces.
- Fake Discord tests.

Acceptance:

- command JSON snapshot tests;
- diff/idempotency tests;
- no registration from shard ready;
- responder tests for respond/defer/follow-up/edit/error.

### Wave 3: Mongo Repositories and Audit

Scope:

- Mongo documents with exact legacy BSON tags.
- Repository ports and adapter implementations for first low-risk features.
- Index plan structures and read-only comparison.
- Go `cmd/mhcat-mongo-audit`.
- Atomic update helpers.
- Repository contract tests.

Blocked until:

- live read-only audit output is reviewed for collection names, indexes, duplicates, mixed types, and dashboard collections.

### Wave 4: Component/Modal Parser

Scope:

- Versioned `mhcat:v1` custom ID codec.
- Legacy compatibility decoders.
- Component/modal router.
- Payload length and secret checks.
- Collision tests.

Acceptance:

- all patterns in `docs/12-component-modal-grammar.md` have golden tests;
- no broad `includes()` dispatch;
- malformed IDs reject safely.

### Wave 5: Feature Parity by Domain Group

Suggested order:

1. Health/help/info/ping with no Mongo writes.
2. Config/dashboard-shared reads with strict compatibility.
3. Logging/webhook abstraction.
4. Welcome/verification/member flows.
5. Tickets and role components.
6. Polls.
7. Economy/work/gacha/shop.
8. XP/rank/voice XP.
9. Cron/scheduler.
10. Chatbot/anti-scam/content features.
11. Rendering/export-heavy features.
12. Owner/admin operations.

Each feature needs:

- legacy behavior reference;
- service/repository/handler tests;
- permission/cooldown/error/defer behavior;
- docs parity update.

### Wave 6: Bug and Performance Fixes

Only after parity tests protect the feature.

Allowed examples:

- replace read-modify-write with atomic updates;
- normalize anti-scam URL matching after ADR;
- move render-heavy paths to worker pool;
- fix scheduler duplicate execution with lease.

Each fix needs behavior-change documentation and tests.

### Wave 7: Operations, Deploy, Rollback

Scope:

- Docker/compose if useful.
- Makefile.
- CI if appropriate.
- operational runbook updates.
- command sync runbook.
- Mongo audit/repair runbook.
- rollback drill.

## 4. No-Go List

- No command registration on every shard ready.
- No broad `customId.includes(...)` router.
- No hardcoded webhook, token, Mongo URI, dashboard URL, owner ID, admin ID, or operator bot ID.
- No direct Mongo calls in Discord handlers.
- No DiscordGo types in domain services or core ports.
- No MongoDB driver types in domain services.
- No read-modify-write for counters when atomic update exists.
- No unbounded goroutine per event.
- No long interaction handler without defer.
- No logging secrets, raw tokens, webhook URLs, Mongo URIs, OAuth access tokens, or raw production user content samples.
- No deleting unknown global commands without dry-run diff and explicit delete flag.
- No production index creation from default bot startup.
- No unique index before duplicate/missing/null audit.
- No TTL index before data-retention ADR.
- No in-place schema rewrite without ADR, audit, dry-run repair/backfill, and rollback guide.
- No dashboard-shared collection schema change without dashboard impact review.
- No SQL-style migration runner, `migrations/` directory, or migration version table by default.

## 5. Gate B Review

Are all blockers resolved?

- exact customId/modal payload grammar: resolved for implementation planning in `docs/12-component-modal-grammar.md`.
- live Mongo collection/index compatibility: partially resolved. Read-only audit tool/process exists, but live audit was not executed because no Mongo env was available.
- dashboard/external worker compatibility: partially resolved. Dashboard sharing is confirmed locally; ChatGPT worker remains unconfirmed.
- command registration/sharding strategy: resolved for implementation planning. Use command sync CLI or explicit leader, never every shard.
- privileged intent strategy: resolved for Wave 1 and parity planning in `docs/14-discord-intents-plan.md`.
- final architecture boundary: resolved for Wave 1.

Which blockers remain?

- Run live Mongo audit against staging/prod.
- Confirm production DB name for bot and dashboard.
- Confirm whether `mhcat-dashboard` is currently deployed and backup/export is in use.
- Confirm whether a ChatGPT worker still consumes `chatgpts`.
- Confirm which local/remote bot repo is production.

Is it safe to begin Wave 1?

Yes, with strict scope: Wave 1 skeleton only. Wave 1 must not register commands, write feature data, create indexes, mutate schemas, or start feature parity.

What is the minimum viable skeleton?

- Config with env aliases and validation.
- Redacted logger.
- App lifecycle with signal handling.
- Mongo connect/ping with context timeout.
- DiscordGo session construction with minimal intents, but no login required in tests and no command registration.
- Import-boundary tests or static checks where practical.

What must be tested before feature parity?

- Config validation and secret redaction.
- Intent dependency validation.
- Responder state machine.
- Command sync diff/dry-run.
- Component/modal codec and legacy decoders.
- Mongo BSON fixture decoding and repository contract tests.
- Atomic update behavior for counters.

What manual confirmations are still required?

- Mongo live audit report.
- Dashboard deployment/DB name.
- ChatGPT worker ownership.
- Webhook/reporting workflow.
- Production bot source of truth.

## Gate B Recommendation

Begin Wave 1 only if the implementation prompt explicitly preserves these limits:

- no feature logic;
- no command registration;
- no Mongo writes except optional ping metadata-free connection health is read-only;
- no index writes;
- no legacy source edits.

Do not begin Wave 3 repositories, Wave 4 feature routing, or Wave 5 parity until the live Mongo audit and dashboard/worker confirmations are resolved for the affected feature.

## Wave 3 Freeze Check

Wave 3 implemented Mongo infrastructure without changing the Gate B boundaries:

- Mongo audit/index tooling is separate from `cmd/mhcat-bot`.
- Bot startup still does not register commands or create indexes.
- MongoDB driver imports remain isolated to `internal/adapters/mongo`.
- `internal/core/**` remains free of MongoDB driver and DiscordGo imports.
- No SQL-style migration directory, runner, or version table was added.
- Feature repositories and feature writes remain deferred to later waves.
- Index creation code is reachable only through `mhcat-mongo-index --apply`; dry-run is default, and unique/TTL indexes require additional explicit allow flags.

## Wave 5.2 Freeze Check

Wave 5.2 wired runtime interactions without changing the command registration or Mongo boundaries:

- `cmd/mhcat-bot` can open Discord Gateway only when `MHCAT_DISCORD_ENABLE_GATEWAY=true`.
- `cmd/mhcat-bot` still does not import or call command sync apply logic.
- Runtime `InteractionCreate` events are translated by `internal/adapters/discordgo` and routed through internal dispatcher/router/responder abstractions.
- Runtime-routable commands are limited to `help`, `ping`, and `info bot`.
- Usage tracking remains no-op by default and writes no Mongo feature data.
- Message Content remains disabled by default.
- No feature Mongo repositories, Mongo feature writes, Mongo index creation, or SQL-style migration system were added.

## Wave 5.3 Freeze Check

Wave 5.3 added staging smoke and command-sync guardrails without changing production startup boundaries:

- Command sync apply is staging-only and guild-scoped.
- Staging apply requires `MHCAT_STAGING_MODE=true` and `MHCAT_STAGING_ALLOW_COMMAND_APPLY=true`.
- Staging apply rejects command deletion and bulk overwrite.
- Staging sync validates local command ownership and permits only `help`, `ping`, and `info`.
- Gateway smoke requires `MHCAT_STAGING_MODE=true` and `MHCAT_STAGING_ALLOW_GATEWAY_SMOKE=true`.
- `cmd/mhcat-bot` still does not register or sync commands.
- No Mongo feature writes, Mongo index creation, new feature groups, or SQL-style migration system were added.
