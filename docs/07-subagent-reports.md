# Subagent Reports

Status: Phase 1 consolidated. All required read-only agents completed. Reports below are summaries normalized into project terminology; no code was modified by subagents.

## Repository Cartographer Agent

Scope: repo tree, entrypoints, handlers, events, commands, models, assets, deployment, dead/duplicate code, global state.

Files inspected: 190 non-git legacy files, including 74 slash commands, 18 events, 47 models, 5 handlers, 5 functions, assets/fonts/config/runtime files.

Commands run: `rg --files`, `find`, `sed`, `rg -n`, `wc -l`, `du -ah`, `git status/log`, read-only `node` scripts, `node --check`.

Legacy behavior found: `shard.js` starts `index.js`; `index.js` owns singleton client and Mongo connection; handlers register commands, events, cron, channel status, and work jobs. No in-repo dashboard/server/worker found.

Mongo collections / fields / indexes found: 47 Mongoose models; no explicit indexes; only `join_message.guild` and `join_message.enable` unique flags with `autoIndex:false`.

Discord commands / events / components found: 74 slash command files, 18 event files, many custom IDs for polls, tickets, voice locks, verification, cron, ranks, lottery, sign/profile, role add/delete.

Config / env found: `TOKEN`, `MONGOOSE_CONNECTION_STRING`, webhook vars, `DASHBOARD_URL` template; Docker, PM2, Crowdin config.

External APIs / side effects found: Discord gateway/REST/webhooks, MongoDB, Google Translate, captcha, canvas/chart, Excel/text attachments.

Bugs / risks / performance issues: broken `functions/delete.js`, empty `lang.js`, missing `commands/`, circular singleton imports, global slash registration on every shard, large files, duplicate XP/rank logic, unsafe regex, unindexed hot paths, hardcoded webhook.

Recommended refactor approach: freeze behavior specs and collection catalog; use app composition root, repositories, one interaction router, feature domains, explicit command registration tool.

Required tests: command snapshots, router/permission tests, repository/index tests, feature parity tests, fake timer cron/voice tests.

Open questions: dashboard/worker status, commented features, prefix commands, production indexes, deployment topology, webhook rotation.

Confidence: Medium

## Behavior Archaeologist Agent

Scope: user-visible commands, events, components, scheduled jobs, Mongo usage, side effects.

Files inspected: all slash commands, all events, all models, entrypoints, handlers, config/runtime files.

Commands run: read-only `rg`, `find`, `sed`, `git status`, static `node -e` extractors.

Legacy behavior found: sharded Discord.js bot; Mongo-first startup; global command registration; slash dispatch; metadata cooldowns not enforced; ad hoc permissions; restart text command; major domains include moderation, logging, welcome, verification, voice rooms, XP, economy, work, polls, cron, stats, chatbot, anti-scam.

Mongo collections / fields / indexes found: all 47 models and fields; no explicit indexes; `join_message` unique flags with auto-index disabled.

Discord commands / events / components found: compact 74-command matrix; message, reaction, voice, member, logging, and interaction events; broad custom ID patterns.

Config / env found: config colors/emojis/descriptions; token/Mongo/webhook env; hardcoded webhook; gateway intents.

External APIs / side effects found: Discord REST/webhooks, Google Translate, captcha, Canvas/ChartJS, Excel, drop-table, is-url, dashboard/docs links.

Bugs / risks / performance issues: no cooldown enforcement, inconsistent permission checks, global command registration, inactive birthday scheduler, disabled lottery creation, cron lifecycle issues, unindexed Mongo, `voice_xp` save typo, component collisions, regex risk, gacha races, inline rendering.

Recommended refactor approach: typed command registry, central middleware, domain services, repositories, typed custom ID router, scheduler service, atomic economy updates.

Required tests: command snapshots, permission/cooldown tests, Mongo race tests, component router tests, fake-timer scheduler tests, feature/event parity tests.

Open questions: birthday/lottery status, external chatgpt worker, dashboard-moved commands, owner/admin semantics, command scope, schema cleanup.

Confidence: Medium

## Discord API / DiscordGo Specialist Agent

Scope: map Discord.js behavior to DiscordGo, interaction/defer strategy, intents, sharding, webhooks, voice state.

Files inspected: package/runtime/config/handlers/events/models/slash commands.

Commands run: read-only `rg`, `find`, `sed`, `awk`, `wc`.

Legacy behavior found: global slash commands, multiple `interactionCreate` listeners, message-content features, reaction roles, temporary voice channels, XP, tickets, join/leave, logging, cron.

Mongo collections / fields / indexes found: all models; no explicit indexes; `join_message.enable` unique risk.

Discord commands / events / components found: 74 commands; active gateway events; components for roles/tickets/lottery/poll/rank/profile/verification/work/announcement.

Config / env found: token/Mongo/webhook vars, prefix/colors/emojis, PM2/Docker.

External APIs / side effects found: Discord Gateway/REST/webhooks, MongoDB, translation, captcha, rendering, cron.

Bugs / risks / performance issues: ACK/defer risk, component collisions, non-shard-safe command registration, unawaited command run, cache reliance, old channel types, per-member voice intervals, MessageContent required for several features.

Recommended refactor approach: DiscordGo only in adapter; one interaction router; explicit responder state machine; command registration via tool or controlled deploy; shard ownership and scheduler lease; cache abstraction.

Required tests: ACK timing, command registration idempotency, custom ID routing, intents, reaction roles, temp voice, poll/lottery concurrency.

Open questions: retaining MessageContent features, presences, reaction roles, command registration mode, leaked webhook rotation.

Confidence: High

## Go Architecture Agent

Scope: Go layout, ports/adapters boundaries, DTO/domain split, lifecycle risks.

Files inspected: package/runtime/config/handlers/events/models/slash commands and docs skeleton.

Commands run: read-only `find`, `sed`, `rg`.

Legacy behavior found: large sharded bot with Mongo persistence, 74 slash commands, 47 models, interactions, cron, rendering, many Discord mutations.

Mongo collections / fields / indexes found: all model fields; no explicit indexes; only `join_message` uniqueness flags.

Discord commands / events / components found: major domains and component patterns.

Config / env found: token/Mongo/webhook vars, hardcoded admin IDs/webhook, intents.

External APIs / side effects found: Discord, MongoDB, Translate, CAPTCHA, canvas/chart, Excel, PM2/Docker.

Bugs / risks / performance issues: singleton client/global Mongoose, side-effect imports, command registration on ready, cron ownership, voice intervals, read-modify-write races, hot unindexed paths, unsafe regex, raw logs.

Recommended refactor approach: layered Go layout, feature services, narrow ports, adapter DTOs for Discord/rendering/cron payloads, pure core entities for stable business concepts.

Required tests: Mongo fixtures, repository contracts, domain table tests, Discord adapter tests, scheduler tests, golden response tests.

Open questions: feature wave priority, MessageContent replacement, production collection names/indexes, scheduler topology, hardcoded webhook rotation.

Confidence: High

## MongoDB Data Agent

Scope: all Mongoose models, query/write patterns, collection names, index recommendations, compatibility plan.

Files inspected: all models plus events/slash commands/handlers/functions/config.

Commands run: read-only `find`, `rg`, targeted `sed`.

Legacy behavior found: Mongoose with `strictQuery=false`, `autoIndex=false`, `bufferCommands=false`; no aggregate/populate/limit/sort; ranks sort in process; many raw update patterns.

Mongo collections / fields / indexes found: inferred Mongoose default collection names; all fields; no explicit indexes; `join_message` uniqueness flags.

Discord commands / events / components found: feature groups and component IDs related to Mongo state.

Config / env found: token/Mongo/webhook/dashboard vars and dependencies.

External APIs / side effects found: Discord REST/state, MongoDB, webhooks, Translate, rendering, cron, chatbot handoff.

Bugs / risks / performance issues: collection-name verification needed, loose/mixed types, non-atomic updates, rank scans, unsafe unique Boolean, duplicates from delete/recreate, cron scan.

Recommended refactor approach: no SQL-style migrations; use catalog, index bootstrap, audit, dry-run repair; exact BSON tags; compatibility decoders; repository groups; audit before unique indexes.

Required tests: BSON fixtures, collection-name tests, repository query/update tests, duplicate audits, atomic concurrency tests, rank ordering, cron reset.

Open questions: live DB catalog, singleton truth, retention policy, dashboard writes, external ChatGPT worker.

Confidence: High

## Security and Abuse Prevention Agent

Scope: secrets, webhooks, dashboard URLs, permissions, input validation, cooldown/rate limiting, MessageContent.

Files inspected: runtime/config/README/test, all models, key handlers/events/functions, all slash commands via static inventory.

Commands run: read-only `sed`, `rg`, `find`, redacted `node -e` scans.

Legacy behavior found: env-based token/Mongo, env webhooks plus one hardcoded webhook, hardcoded admin IDs, message restart, cooldown metadata not enforced, ad hoc permissions, hardcoded dashboard URLs, raw errors/logged content.

Mongo collections / fields / indexes found: all models; no explicit indexes; `join_message` unique flags.

Discord commands / events / components found: high-risk admin/moderation/reset/report/logging/role/lottery paths.

Config / env found: token/Mongo/webhook/DASHBOARD vars; README old config keys; no secret validation/redaction layer.

External APIs / side effects found: Discord, MongoDB, Translate, dashboard links, PM2 logs.

Bugs / risks / performance issues: hardcoded webhook secret, hardcoded IDs, no central cooldown, regex injection/ReDoS, raw error leakage, no default member permissions, no central allowed mentions, admin invite permissions.

Recommended refactor approach: central config/secret validation, centralized auth/cooldown, disable MessageContent where possible, validation helpers, redacted logging, safe errors, least-privilege invite.

Required tests: secret scan, env validation, permission matrix, owner-only restart, cooldown/rate-limit, regex safety, allowed mentions, redaction, MessageContent-disabled mode.

Open questions: retained MessageContent features, operator source, restart control, dashboard status, webhook need, least-privilege permissions.

Confidence: Medium

## Performance and Reliability Agent

Scope: blocking operations, Mongo slow queries, Discord REST rate limits, cache, sharding, cron, rendering, shutdown.

Files inspected: runtime/deploy/test, cron/channel/gift handlers, high-risk events/commands, all models.

Commands run: read-only `rg`, `sed`, `find`, schema/index searches.

Legacy behavior found: auto sharded bot, Mongo before login, every shard loads handlers, command registration on ready, Mongoose pool options, unbounded member cache, only SIGINT Mongo close.

Mongo collections / fields / indexes found: hot fields include `guild`, `member/user`, `messageid`, `id`, `channel`, `state`, `end_time`, `today`, `time`.

Discord commands / events / components found: high-risk poll, translate, rank/profile, gacha, game, cron, stats, voice, clear, shop, work.

Config / env found: token/Mongo/webhooks, PM2/Docker, Node memory.

External APIs / side effects found: Discord REST, Translate, Discord CDN image load, chart/Excel rendering.

Bugs / risks / performance issues: unindexed scans, per-message XP writes, full guild member fetches, in-memory rank sorting, per-user voice intervals, non-atomic updates, REST bursts, duplicate cron, channel rename loops, inline rendering, command registration rate limits, shutdown not draining.

Recommended refactor approach: explicit indexes, projections, atomic updates, bounded REST queue, worker pools/timeouts for rendering/translation, persisted voice sessions, scheduler lease, bounded config caches, explicit command registration, graceful shutdown.

Required tests: index dry-run/presence, atomic repository tests, scheduler lease, REST 429 queue, voice restart, load/timeout/cancellation, shutdown drain.

Open questions: production guild/member counts, shard counts, current indexes, chatgpt worker, freshness expectations, command registration policy.

Confidence: High

## Testing Strategy Agent

Scope: legacy tests, testability gaps, parity tests, Mongo contracts, fake Discord, fixtures, race/benchmark/fuzz.

Files inspected: package/test/config/runtime/handlers/models/all command inventory/key events.

Commands run: read-only `find`, `sed`, `rg`, `node -e` static extraction.

Legacy behavior found: only `npm test -> node test-startup.js`, a live Discord login smoke test; no unit/integration/fake tests.

Mongo collections / fields / indexes found: all model fields; no explicit indexes; type/typo fields.

Discord commands / events / components found: 74 slash modules; key events/components/custom IDs.

Config / env found: token/Mongo/webhook vars, PM2/Docker.

External APIs / side effects found: Discord, Mongo, Translate, canvas/chart, cron, webhooks, generated attachments.

Bugs / risks / performance issues: singleton client and direct models hurt testability; races in counters/arrays; timer leaks; known bugs in voice/chatbot/delete-data/NaN/channel types/slash usage.

Recommended refactor approach: ports/adapters, pure domain services, injected clock/timer/RNG, command/component snapshots, atomic repositories, cancellable long operations.

Required tests: behavior parity, Mongo contracts, Discord fake tests, golden fixtures, race tests, benchmarks, fuzz tests, cancellation/timeout tests.

Open questions: bug preservation vs fixes, ephemeral Mongo availability, external chatgpt worker, image golden tolerance.

Confidence: High

## Compatibility and Rollout Agent

Scope: Node-to-Go rollout, Mongo compatibility, canary/shadow, rollback, deployment/runbook.

Files inspected: runtime/deploy/config/handlers/events/models/slash inventory/docs skeleton.

Commands run: read-only `rg`, `find`, `sed`, `perl`, `git -C MHCAT`.

Legacy behavior found: Node/discord.js sharded bot, Mongo-first startup, global command registration, restart text command, scheduled work and channel stats.

Mongo collections / fields / indexes found: 47 model files, inferred collection names, no explicit indexes, recommended index candidates.

Discord commands / events / components found: 74 slash commands, 18 events, broad component/custom ID evidence.

Config / env found: legacy envs, Go alias plan, Docker/PM2, no systemd/compose.

External APIs / side effects found: Discord, MongoDB, cron, Translate, rendering, webhooks, chatbot local state.

Bugs / risks / performance issues: Node and Go same token/guild would double-handle; command registration races; scheduler singleton needed; non-atomic writes; type drift; MessageContent dependence; hardcoded IDs/webhook.

Recommended refactor approach: staging bot first, shadow read-only mode without side effects, canary guild with exclusive feature ownership, rollback by stopping Go/restarting Node with compatible Mongo docs.

Required tests: Mongo fixtures/contracts, index dry-run, Discord fake, scheduler singleton, golden parity, canary smoke, rollback drill, race/benchmark/fuzz.

Open questions: production collections/indexes, canary guild, staging token, restart command, launch feature order, supervisor target, dashboard status, command registration scope.

Confidence: Medium

## Code Quality Review Agent

Scope: review constraints, docs skeleton, architecture/code-quality risks, Gate A/Gate B blockers.

Files inspected: docs skeleton, runtime/config/deploy, all models/handlers/events/functions/slash commands.

Commands run: read-only `rg`, `find`, `sed`, `head`, `perl`, `git status`.

Legacy behavior found: singleton client; runtime state attached to client; global command registration; multiple interaction/message handlers; cron split; restart text command.

Mongo collections / fields / indexes found: all model fields; no explicit indexes; unsafe `join_message.enable`.

Discord commands / events / components found: command names/events/custom IDs; broad component matching.

Config / env found: env vars, hardcoded admin IDs/webhook, intents, Mongo options, Docker/PM2.

External APIs / side effects found: Discord, MongoDB, Translate, captcha, Canvas/chart, Excel, cron.

Bugs / risks / performance issues: type leakage risk, non-atomic updates, unawaited writes, raw errors, global command registration, custom ID routing, per-message/voice hot paths, unsafe regex.

Recommended refactor approach: Gate A complete matrices/catalogs/ADRs; Gate B import-boundary checks, context everywhere, config/redaction, typed error mapping, custom ID codec, responder, scheduler, Mongo dry-run tooling.

Required tests: fixture decode, repository contracts, custom ID tests, Discord fake tests, env/redaction tests, concurrency, scheduler, import-boundary, golden parity.

Open questions: production collections/indexes, active/dead features, restart replacement, command registration mode, chatgpt worker, topology, intents, docs/dashboard links.

Confidence: Medium

## Gate A Review

1. All entrypoints found: yes
   - `shard.js` and `index.js` are the active bot entrypoints.
   - No in-repo dashboard/server/worker entrypoint was found.

2. All commands found: yes
   - 74 slash command files found.
   - Prefix dispatcher exists, but no `commands/` directory exists in this checkout.
   - Direct restart message commands found in `index.js`.

3. All events found: yes
   - 18 event files found.

4. All components/modals found: partial
   - Broad custom ID patterns are inventoried.
   - Exact payload grammar for every dynamic custom ID is not yet frozen.
   - Implementation should not start until command/component snapshot tests or a targeted component grammar pass is complete.

5. All Mongo models found: yes
   - 47 model files found.
   - Collection names are inferred from Mongoose defaults and still need live DB verification.

6. All env vars found: yes for local code/template
   - `TOKEN`, `MONGOOSE_CONNECTION_STRING`, webhook vars, and `DASHBOARD_URL` were found.
   - Go env aliases are planned.

7. All external APIs found: partial
   - Discord, MongoDB, webhooks, Google Translate, captcha, rendering, Excel/text attachments, dashboard/docs links found.
   - External dashboard and possible `chatgpt` worker ownership are unknown.

8. Unclear areas:
   - Actual production Mongo collection names and indexes.
   - Exact custom ID payload grammar.
   - External dashboard/worker writes to Mongo.
   - Whether birthday notifications and lottery creation should remain inactive.
   - Canonical owner/admin IDs.
   - Target Go deployment topology.
   - Whether all Message Content features must be retained.

9. Highest behavior parity risks:
   - MessageContent features: XP, chatbot, announcement relay, anti-scam, logging, restart.
   - Inactive/commented birthday and lottery behavior.
   - Broad custom ID matching and existing live components.
   - Permission/cooldown enforcement differs from declared metadata.

10. Highest Mongo compatibility risks:
   - Inferred collection names may differ from production.
   - Mixed string/number/bool field types.
   - Duplicate singleton/per-user records.
   - `join_message.enable` unique Boolean.
   - Non-atomic counter/inventory/vote updates.

11. Highest DiscordGo mapping risks:
   - Interaction ACK/defer deadlines.
   - Command registration strategy and global rate limits.
   - Sharding/scheduler ownership.
   - Discord.js cache assumptions versus DiscordGo REST/cache behavior.
   - Privileged intents and MessageContent dependency.

12. Recommendation:
   - Do not enter main implementation yet.
   - Next targeted research should freeze exact command JSON and custom ID/modal payload grammars, then verify live Mongo collection/index state if credentials are available.
   - Wave 1 skeleton can start only after Gate B accepts the architecture constraints and implementation scope.
