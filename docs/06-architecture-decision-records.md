# Architecture Decision Records

Status: Phase 1.5 Gate B review. Phase 1 ADRs remain planning context; Phase 1.5 ADR updates are accepted for Wave 1/Wave 2 implementation constraints.

## ADR-001 Go Project Layout

- Status: Proposed
- Context: Legacy has 74 slash command files, 47 model files, 18 events, cron jobs, rendering, and many Discord side effects.
- Decision: Use a Go module under `MHCAT-REFACTOR/` with `cmd/mhcat-bot`, `cmd/mhcat-tools`, `internal/app`, `internal/config`, `internal/core`, `internal/discord`, `internal/adapters`, `internal/jobs`, and `internal/observability`. Group implementation by feature/domain rather than one package per legacy JS file.
- Options considered: mirror JS tree; flat package; ports-and-adapters.
- Consequences: More upfront boundaries, less Discord/Mongo coupling.
- Risks: Over-abstraction if features are split too early.
- Tests: import-boundary tests, package compile tests.

## ADR-002 DiscordGo Adapter Boundary

- Status: Proposed
- Context: Legacy Discord.js types leak everywhere through singleton `client`.
- Decision: `discordgo` may appear only in Discord adapter/session/router packages. Core services receive internal DTOs and use ports.
- Options considered: use `discordgo` directly in handlers/services; full DTO translation.
- Consequences: More mapping code, but testable services.
- Risks: DTO drift from Discord payloads.
- Tests: static import checks; fake Discord adapter tests.

## ADR-003 MongoDB Repository Design

- Status: Proposed
- Context: Legacy directly imports Mongoose models across commands/events.
- Decision: Use adapter repositories with BSON documents matching legacy tags. Repositories expose feature-oriented methods and map Mongo errors to domain errors.
- Options considered: generic DAO; per-collection repositories; feature repositories.
- Consequences: Better atomic operations and context handling.
- Risks: Large initial catalog; duplicate repository methods.
- Tests: repository contract tests and legacy fixture decode tests.

## ADR-004 MongoDB Index Bootstrap Strategy

- Status: Proposed
- Context: Legacy has no explicit indexes and Mongoose `autoIndex:false`.
- Decision: No high-risk index writes on bot startup. Use `mhcat-tools` check/dry-run/apply, with duplicate audits before unique indexes.
- Options considered: startup auto-index; no index tooling; explicit operational tool.
- Consequences: Safer production rollout.
- Risks: Operators may forget to apply low-risk indexes.
- Tests: dry-run output, existing-index diff, duplicate audit, context cancellation.

## ADR-005 Data Compatibility and Schema Change Strategy

- Status: Proposed
- Context: MongoDB data is schemaless and may contain legacy drift. User permits schema changes if needed.
- Decision: Default to legacy-compatible reads and rollback-compatible writes. Schema changes are allowed only with ADR, fixtures, audit evidence, dry-run repair/backfill, and rollback notes.
- Options considered: preserve all fields forever; immediate canonical schema rewrite; compatibility-first with controlled improvements.
- Consequences: Safer canary and rollback while allowing cleanup.
- Risks: Dual-read/write complexity if new schema is introduced.
- Tests: BSON decode, Node-readable write fixtures, repair dry-run tests.

## ADR-006 Command Router Design

- Status: Proposed
- Context: Legacy registers and dispatches global slash commands dynamically, with metadata not consistently enforced.
- Decision: Build a typed command registry with generated Discord command JSON, central middleware for permission/cooldown/defer/logging/recovery, and feature handlers.
- Options considered: manual switch; dynamic file loading; typed registry.
- Consequences: Snapshot-friendly behavior.
- Risks: Command metadata extraction must be exact.
- Tests: command JSON snapshots, router table tests, permission/cooldown tests.

## ADR-007 Component / Modal Router Design

- Status: Proposed
- Context: Legacy uses many broad `customId.includes(...)` checks and multiple interaction listeners.
- Decision: Use a single component/modal router with typed namespaces and encoded payloads. Legacy custom IDs must be supported or bridged where existing messages may still be clicked.
- Options considered: preserve broad matching; typed router only; compatibility router plus typed new IDs.
- Consequences: Reduced collision/spoofing risk.
- Risks: Existing live components may use old IDs.
- Tests: custom ID golden tests, collision tests, legacy ID compatibility tests.

## ADR-008 Interaction Response / Defer Strategy

- Status: Proposed
- Context: Discord interactions require initial response or defer quickly; legacy long paths are inconsistent.
- Decision: All DB/REST/render/member-fetch/external API commands defer by default. Responder tracks state: respond, defer, follow-up, edit original, delete original, safe error.
- Options considered: handler-managed responses; centralized responder.
- Consequences: Fewer timeout failures.
- Risks: Public/ephemeral mismatches if metadata is wrong.
- Tests: ACK timing tests, duplicate response tests, safe error tests.

## ADR-009 Gateway Intent Strategy

- Status: Proposed
- Context: Legacy enables Guilds, GuildMembers, GuildMessages, GuildMessageReactions, GuildVoiceStates, and MessageContent.
- Decision: Start with current required intents for parity, behind explicit config flags. Remove `MessageContent` only after replacing or disabling message-content features.
- Options considered: least-intent immediately; parity-first intents; feature-flag intents.
- Consequences: Honest parity; allows later reduction.
- Risks: Privileged intent approval/privacy burden remains.
- Tests: disabled-intent behavior tests, event coverage tests.

## ADR-010 Sharding and Scheduler Strategy

- Status: Proposed
- Context: Legacy uses discord.js `ShardingManager`; cron is partly shard-0 guarded and partly per-process.
- Decision: Use DiscordGo shard ID/count options for gateway sessions. Guild-scoped work runs on owning shard. Global schedulers use one explicit scheduler process or Mongo-backed lease.
- Options considered: one process all shards; one process per shard; separate scheduler worker.
- Consequences: Prevents duplicate cron/events in multi-replica rollout.
- Risks: Lease failure or split-brain if implemented poorly.
- Tests: shard ownership tests, scheduler singleton tests, restart recovery tests.

## ADR-011 Config and Secret Strategy

- Status: Proposed
- Context: Legacy env vars exist but admin IDs/dashboard URLs and one webhook are hardcoded.
- Decision: Central config module validates env, supports legacy aliases, redacts secrets, and moves owner/admin IDs and webhooks to env/config.
- Options considered: copy legacy env; config file; typed env config.
- Consequences: Safer startup and operations.
- Risks: Env transition mistakes.
- Tests: env alias tests, validation tests, secret redaction tests, hardcoded secret scan.

## ADR-012 Error Handling Strategy

- Status: Proposed
- Context: Legacy logs raw errors and can send raw errors to Discord.
- Decision: Use typed/domain errors, wrapped internal errors, safe user-facing messages, correlation IDs, and redacted logs.
- Options considered: return raw errors; centralized error mapper.
- Consequences: Better security and debuggability.
- Risks: Less detail for users unless logs are accessible.
- Tests: error mapping, safe response, redaction tests.

## ADR-013 Logging / Metrics Strategy

- Status: Proposed
- Context: Legacy uses console/PM2 logs and raw stacks.
- Decision: Use structured logging with redaction and metrics for Mongo latency, Discord REST errors/429s, render duration, scheduler leases, shard reconnects, and interaction latency.
- Options considered: stdout only; structured logger.
- Consequences: Operational visibility.
- Risks: Metrics sink not yet chosen.
- Tests: logger redaction tests, metrics no-op tests.

## ADR-014 Testing Strategy

- Status: Proposed
- Context: Legacy only has live Discord login smoke test.
- Decision: Build table-driven unit tests, Discord fake adapter tests, Mongo repository contract tests, BSON fixtures, command/component snapshots, scheduler fake-time tests, race/benchmark/fuzz tests where justified.
- Options considered: live integration only; layered test strategy.
- Consequences: Enables incremental refactor.
- Risks: Fixture maintenance cost.
- Tests: see `08-test-plan.md`.

## ADR-015 Rollout / Rollback Strategy

- Status: Proposed
- Context: Running Node and Go with same token/guild would double-handle events and writes.
- Decision: Use staging bot/token first, then shadow read-only mode, then canary guild with exclusive feature ownership, then staged rollout. Rollback is stopping Go and restarting Node with compatible Mongo documents.
- Options considered: big-bang swap; parallel same-token shadow; staged canary.
- Consequences: Lower blast radius.
- Risks: Requires feature flags and Node/Go ownership coordination.
- Tests: canary smoke tests, rollback drill, Node-readable Go writes.

## ADR-016 Command Registration Strategy

- Status: Proposed
- Context: Legacy registers/deletes global commands on every shard `ready`.
- Decision: Move command registration to `mhcat-tools commands register --dry-run/--apply` or one controlled deploy step. Bot startup should not mutate global commands by default.
- Options considered: startup registration; one-shard registration; explicit tool.
- Consequences: Avoids global command REST races/rate limits.
- Risks: Command deploy becomes a separate operation.
- Tests: registration diff tests, idempotency tests, shard race tests.

## ADR-017 Security and Abuse Strategy

- Status: Proposed
- Context: Hardcoded webhook/admin IDs, no central cooldowns, unsafe regex, raw errors, and no central allowed mentions.
- Decision: Centralize authorization, owner/admin policy, cooldown/rate-limit, input validation, `allowedMentions`, and webhook sending. Replace user-controlled regex with normalized domain matching where behavior change is approved.
- Options considered: preserve ad hoc checks; central security middleware.
- Consequences: Safer behavior, some intentional behavior changes.
- Risks: Permission parity differences.
- Tests: permission matrix, cooldown, regex, allowed mentions, owner-only tests.

## Gate B Blockers

- Docs must identify unresolved behavior and data uncertainties.
- Go skeleton must compile.
- Import-boundary checks must prevent DiscordGo/Mongo leakage into core.
- Config validation and redaction must exist before feature ports.
- Responder state machine and custom ID router must be tested before complex commands.
- Mongo repository contracts and index dry-run tooling must exist before production data writes.

## Phase 1.5 Gate B ADR Updates

These updates supersede conflicting Phase 1 wording for implementation waves.

## ADR-016 Command Registration and Sharding Strategy

- Status: Accepted for Wave 1/Wave 2.
- Context: Legacy `handler/slash_commands.js` runs from `ready` and calls `client.application.commands.create(data)` for every slash command, then deletes unknown global commands. In a sharded bot, every shard can race to mutate global application commands.
- Decision: Production command registration must not happen independently on every shard. Build a dedicated `cmd/mhcat-command-sync` CLI that loads the typed command registry, fetches Discord application commands, computes a diff, logs create/update/delete decisions, and only mutates with explicit `--apply`. Development may support guild-scoped command sync for fast iteration. Bot shard startup may validate registry availability but must not create/update/delete application commands.
- Options considered: register at app startup from every shard; register from shard 0 only; register from a single leader process; separate command sync CLI; manual dashboard-only command registration.
- Consequences: Deploys gain an explicit command-sync step. Shards no longer fight global command state or hit avoidable REST rate limits.
- Risks: Operators can forget to run command sync after command changes. Global command propagation remains delayed by Discord.
- Tests: registry snapshot tests, diff tests, dry-run output tests, deletion safety tests, idempotent sync tests, no-command-sync-on-shard-ready test.

Deletion policy:

- Never delete unknown global commands by default.
- `--dry-run` must show unknown commands separately.
- `--apply --delete-unknown` must be explicit and log every deletion target.
- Rollback: rerun sync from the Node-compatible command snapshot or restart Node after disabling Go command sync.

Sharding policy:

- `cmd/mhcat-bot` owns one configured shard or shard range.
- Guild-scoped events are processed only by the owning shard.
- Global jobs are run by a scheduler process or by a Mongo lease, not by every shard.
- No command registration from `Ready`.

## ADR-018 Component and Modal Custom ID Strategy

- Status: Accepted for Wave 4 design.
- Context: Legacy component routing uses multiple `interactionCreate` listeners, broad `includes()` checks, generic modal ID `nal`, raw user/admin-controlled custom IDs, and delimiter parsing. See `docs/12-component-modal-grammar.md`.
- Decision: Go uses a single component/modal router. Existing live legacy IDs are supported by explicit compatibility decoders. Newly generated IDs use `mhcat:v1:<feature>:<action>:<payload>` with bounded, validated payloads. Sensitive or oversized state is stored server-side with only an opaque state ID in the custom ID.
- Options considered: preserve legacy parser exactly; preserve parser behind wrappers; versioned new IDs plus legacy compatibility; full replacement.
- Consequences: Existing messages remain clickable while new messages are safer and testable.
- Risks: Some dead/rare legacy IDs may be missed; weak raw-text IDs need scoped compatibility handling.
- Tests: golden encode/decode tests, collision tests, malformed legacy ID rejection tests, live-message compatibility fixtures.

No-Go:

- No broad `customId.includes(...)` router in Go.
- No raw captcha answer, password, webhook, token, or user-provided free text in new custom IDs.
- No multiple independent global component routers.

## ADR-019 Privileged Intent Strategy

- Status: Accepted for Wave 1 config and Wave 5 feature flags.
- Context: Legacy enables `GuildMembers`, `GuildMessages`, `GuildMessageReactions`, `GuildVoiceStates`, and `MessageContent`. Some are required for parity but carry privacy/approval and memory risks.
- Decision: `Guilds` is always enabled. All privileged or high-volume intents are disabled by default and enabled by explicit config/feature dependencies. Message Content remains off by default. Restart-by-message and prefix commands are not carried into Go by default; restart becomes owner-only slash command or out-of-band deployment action.
- Options considered: parity-first enable all intents; least privilege immediately; feature-flagged minimal intents.
- Consequences: Wave 1 skeleton runs with minimal intents. Feature parity must declare and validate its required intents.
- Risks: A feature may appear implemented but remain disabled if operators do not enable its required intent.
- Tests: intent builder tests, feature-to-intent validation tests, disabled-intent behavior tests.

See `docs/14-discord-intents-plan.md`.

## ADR-020 External Dashboard and Worker Compatibility

- Status: Accepted for data design.
- Context: Local `../mhcat-mono/mhcat-dashboard` reads/writes `join_messages`, `guilds`, and `work_somethings`, reads `warndbs`, and exports many guild collections. A ChatGPT worker is inferred from `chatgpts` handoff behavior but no local worker code was found.
- Decision: Treat shared Mongo collections as public contracts. Existing collection names and fields remain readable. Schema changes must be additive and must document dashboard impact. Go repositories must use patch-style writes for dashboard-shared collections. The ChatGPT handoff schema is preserved until the external worker is confirmed retired.
- Options considered: bot-private data assumption; dashboard-aware compatibility; immediate dashboard rewrite.
- Consequences: Safer rollout and rollback, with more compatibility constraints on schema cleanup.
- Risks: Dashboard production DB name and deployment status are still manually unconfirmed.
- Tests: BSON fixture tests for dashboard-shared docs, repository patch-write tests, backup/export compatibility checklist.

See `docs/13-external-compatibility.md`.

## ADR-021 Mongo Schema Change Policy

- Status: Accepted.
- Context: User permits schema changes if justified, but MongoDB does not require SQL-style migrations and Node rollback remains important.
- Decision: Wave 1 makes no production schema changes. Later schema changes must be additive first, ADR-backed, live-audited or fixture-backed, dry-run repair/backfill capable, and rollback-compatible. No SQL-style migration runner, migration directory, or version table will be introduced by default.
- Options considered: preserve all legacy fields forever; immediate canonical rewrite; additive compatibility-first changes.
- Consequences: Existing data and dashboard flows remain stable while allowing targeted cleanup such as component state, scheduler leases, normalized anti-scam fields, or canonical numeric fields.
- Risks: Dual-read/write may add complexity; unresolved live data drift can block unique indexes and strict types.
- Tests: legacy fixture decode tests, compatibility write tests, repair dry-run tests, rollback checklist.

Allowed future additive collections or fields, only after ADR:

- `mhcat_component_states` for large/sensitive component state.
- `mhcat_scheduler_locks` for scheduler ownership.
- optional canonical numeric fields for XP/economy/work/gacha after type audit.
- optional normalized anti-scam URL/domain fields after security ADR.

Blocked without audit/ADR:

- in-place type rewrites;
- destructive dedupe;
- unique indexes on singleton/user keys;
- TTL indexes;
- dashboard-shared schema changes.

## ADR-022 Mongo Audit and Index Tooling Boundary

- Status: Accepted for Wave 3.
- Context: Feature waves need Mongo visibility and index planning, but production data must not be mutated by default and MongoDB is not using SQL-style migrations.
- Decision: Add separate `mhcat-mongo-audit` and `mhcat-mongo-index` CLIs. Audit is read-only. Index tooling defaults to dry-run, never drops indexes, and only creates safe missing indexes with explicit `--apply`. Unique indexes require `--allow-unique` plus clean duplicate audit. TTL indexes require `--allow-ttl` plus a retention ADR/note.
- Options considered: app startup auto-indexing; SQL-style migration runner; separate operational tools with dry-run defaults.
- Consequences: Operators can inspect live collection/index compatibility before feature writes. Bot startup remains free of index mutation.
- Risks: The partial catalog may miss collections until expanded; live audit still requires valid Mongo env; index apply must be treated as an operational change.
- Tests: pure audit analyzer tests, index diff tests, atomic update builder tests, config safety tests, missing-env CLI safety checks.

## ADR-023 Feature Pipeline and Low-Risk Utility Command Strategy

- Status: Accepted for Wave 5.1.
- Context: Feature parity must start with a small, testable slice that proves command definitions, feature modules, interaction routing, responders, and usage hooks can work together without DiscordGo or MongoDB leaking into services.
- Decision: Implement a feature module registry that exposes local command definitions and registers handlers against the internal interaction router. Start with low-risk utility commands only: `help`, `ping`, and the safe `info bot` subset. Command definitions are visible to `mhcat-command-sync` dry-run, but `cmd/mhcat-bot` still does not sync or register commands. Usage tracking is a port with a no-op production implementation until Mongo audit/repository work approves `all_use_count` writes.
- Options considered: implement all utility commands at once; start with command definitions only; implement a narrow end-to-end feature pipeline.
- Consequences: The refactor gains a repeatable feature pattern and test coverage while avoiding high-risk data writes and external API calls.
- Risks: Utility usage statistics temporarily diverge from legacy because command usage writes are disabled. The `info` command is intentionally partial until Discord guild/member/shard behavior is modeled.
- Tests: feature registry tests, command validation/dry-run tests, utility service tests, handler/responder tests, golden response tests, boundary import tests, usage no-op/fake tests.

## ADR-024 Runtime Interaction Gateway Wiring

- Status: Accepted for Wave 5.2.
- Context: Wave 5.1 implemented low-risk utility handlers in-process, but `cmd/mhcat-bot` did not yet connect DiscordGo `InteractionCreate` events to the internal router. Runtime wiring must preserve the command-sync boundary, keep gateway disabled by default, and keep DiscordGo types out of feature services and core.
- Decision: `cmd/mhcat-bot` may register a DiscordGo `InteractionCreate` handler and open Gateway only when `MHCAT_DISCORD_ENABLE_GATEWAY=true`. DiscordGo payloads are translated in `internal/adapters/discordgo` into internal interaction DTOs and dispatched through `internal/discord/runtime.Dispatcher`, middleware, router, and responder. Bot startup never creates/updates/deletes application commands. A one-shot smoke mode is available behind `MHCAT_DISCORD_GATEWAY_SMOKE_TEST=true`.
- Options considered: keep runtime test-only until all utility commands complete; wire gateway immediately for utility subset; combine gateway startup with command sync.
- Consequences: Implemented utility commands can now run from real slash interactions in staging while command registration remains an explicit CLI action. Gateway startup remains opt-in and Message Content remains disabled by default.
- Risks: No live Discord smoke test has been run in this environment; staging may reveal DiscordGo session/event edge cases. Usage tracking remains no-op, so `all_use_count` parity is still deferred.
- Tests: app wiring tests, dispatcher tests, DiscordGo adapter tests, command option parser tests, gateway ready timeout tests, missing-env safety checks, boundary import scans.

## ADR-025 Staging Smoke and Guild Command Sync Guardrails

- Status: Accepted for Wave 5.3.
- Context: Before canary feature work, the refactor needs a controlled staging path to verify guild-scoped command sync, gateway ready, and manual utility interactions without exposing production/global command state or data.
- Decision: Add staging-only config and scripts. `mhcat-command-sync --apply` now requires `MHCAT_STAGING_MODE=true`, `MHCAT_STAGING_ALLOW_COMMAND_APPLY=true`, guild scope, no delete, and no bulk overwrite. Staging sync validates local command ownership metadata and permits only managed `help`, `ping`, and `info`. Gateway smoke requires `MHCAT_STAGING_MODE=true` and `MHCAT_STAGING_ALLOW_GATEWAY_SMOKE=true`.
- Options considered: run staging manually without code guardrails; allow global dry-run/apply; add staging-specific guardrails and scripts.
- Consequences: Staging smoke can verify the current utility slice while default production behavior remains unchanged. Command ownership metadata is local-only and stripped before Discord payload/hash generation.
- Risks: Operators still need valid staging Discord/Mongo env. Real smoke has not been executed in this environment.
- Tests: staging config tests, command ownership/hash tests, staging sync safety tests, CLI apply guard tests, script secret scans, app smoke timeout tests.
