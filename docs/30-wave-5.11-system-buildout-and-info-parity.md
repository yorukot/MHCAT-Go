# Wave 5.11 System Buildout Fan-Out And Info Parity

## Gate C Review

- Legacy `MHCAT/` source remained read-only.
- `cmd/mhcat-bot` still does not register or sync Discord application commands.
- Discord Gateway remains disabled by default.
- Message Content remains disabled by default.
- No Mongo feature write or index creation was added.
- Usage tracking remains no-op.
- Feature services and `internal/core/**` remain free of DiscordGo and MongoDB driver imports.

## Subagent Fan-Out

The broad feature-parity buildout was split into read-only system agents. Each agent inspected one system, reported UI/UX parity requirements, old bugs to fix, Mongo/index implications, performance risks, tests, and proposed Go ownership. No subagent modified files.

| Agent scope | Status | Main outcome |
| --- | --- | --- |
| Economy/sign/profile/shop/game | complete | Requires plural Mongo catalog fixes, atomic coin operations, canvas golden tests, and staged implementation before games/shop/gacha. |
| XP/ranking | complete | Requires message/voice gateway event dispatch, attachment support, normalized rank strategy, and atomic XP updates before feature port. |
| Ticket/private channels | complete | Preserve panel/open/close UI but fix premature config writes, duplicate detection, permissions, and allowed-mentions safety. |
| Poll/vote | complete | Preserve button/select UI but fix broad `poll_` routing, unawaited writes, embedded vote-array races, and chart/export bugs. |
| Join/leave/verification/logging | complete | Preserve onboarding/logging UI but fix role loop returns, unsafe safe-server regex, logging attribution, and permissions. |
| Moderation/admin/safety | complete | High-risk permission bugs found; warning/admin/safety features need stricter policies and atomic repositories. |
| Cron/announcement/birthday | complete | Durable scheduler ownership, leases, and idempotency are required before Go owns scheduled sends. |
| Voice rooms/locks | complete | Requires `GuildVoiceStates` intent/config, centralized voice event dispatch, reconciliation, and rate-limit queues. |
| Utility/info/chat | complete | `/info user` and `/info guild` were the next safe parity gap; translate/chat need provider abstractions and Message Content gates. |
| Mongo repository/index | complete | Current catalog is a partial scaffold and must be corrected to production plural collection names before feature writes. |

## Files Created

- `internal/core/ports/discordinfo.go`
- `internal/adapters/discordgo/discordinfo.go`
- `internal/adapters/discordgo/discordinfo_test.go`
- `docs/30-wave-5.11-system-buildout-and-info-parity.md`

## Files Updated

- `internal/adapters/discordgo/interaction_adapter.go`
- `internal/adapters/discordgo/interaction_runtime_test.go`
- `internal/adapters/discordgo/responder.go`
- `internal/app/wiring.go`
- `internal/discord/commands/builtin_registry.go`
- `internal/discord/commands/builtin_registry_test.go`
- `internal/discord/features/utility/legacy_info.go`
- `internal/discord/features/utility/module.go`
- `internal/discord/features/utility/status_handler.go`
- `internal/discord/features/utility/status_handler_test.go`
- `internal/discord/interactions/command_options.go`
- `internal/discord/interactions/command_options_test.go`
- `internal/discord/responses/responder.go`
- `internal/testutil/fakebotinfo/botinfo.go`
- `README.md`
- `docs/09-risk-register.md`
- `docs/10-feature-parity-checklist.md`

## Info User/Guild Parity

Implemented read-only `/info user` and `/info guild` handlers.

- `/info user` defers, reads a driver-agnostic Discord user snapshot, and edits the original response with the legacy-style embed title, avatar thumbnail, user ID, account creation time, and guild join time.
- `/info guild` defers, reads a driver-agnostic Discord guild snapshot, and edits the original response with the legacy-style embed title, icon thumbnail, banner image, member count, boost state, creation time, owner mention, emoji count, preferred locale, and verification level.
- Lookup failures return a red safe error embed. Internal error strings are not surfaced.
- The DiscordGo adapter reads from session state first and uses context-aware REST fallback where needed.
- Bot startup still does not sync command definitions. Since `/info` command shape changed, command sync must be run manually through the existing guarded CLI.

## Intentional Bug Fixes

- The old `/info shard` empty initial embed remains fixed: Go responds with shard fields immediately.
- `/info user` and `/info guild` do not copy fragile nil/undefined behavior. Missing timestamps, locale, owner, or images render safe fallback text or omit optional media.
- Lookup errors are safe and do not leak raw internal errors.

## Platform Gaps Before Stateful Systems

The agents found shared prerequisites before implementing high-risk systems:

1. Correct `DefaultCollectionCatalog()` to production plural collection names and expand it beyond the current high-risk scaffold.
2. Add collection-name contract tests before any feature repository writes.
3. Add command model support for choices, channel type constraints, permissions metadata, and role/channel/user runtime option values.
4. Add gateway event dispatch for message, reaction, guild member, and voice-state events with explicit intent config.
5. Add response attachment support for rank/profile/poll images and exports.
6. Add allowed-mentions support to preserve visible legacy content without accidental pings.
7. Add Discord side-effect ports for channels, roles, members, messages, and audit-log reads.
8. Add per-guild rate-limit queues/singleflight for channel creation, role grants, audit-log reads, and message edits.
9. Add scheduler lease/idempotency infrastructure before cron, birthday, work payout, or daily reset jobs.
10. Keep Node and Go mutually exclusive per event-owning feature during rollout to avoid duplicate side effects.

## Recommended Buildout Queue

1. Platform Wave A: command model gaps, response attachments, allowed mentions, event dispatch, Discord side-effect ports.
2. Platform Wave B: production-plural Mongo catalog, repository contracts per collection group, duplicate-audit fixtures, non-unique index plan fixes.
3. Utility Wave: `/translate` provider abstraction and `/chat` config commands only after Message Content and external worker contracts are confirmed.
4. Ticket Wave: ticket setup/delete/open/close with config persisted after valid modal submit and safe channel permission checks.
5. Onboarding Wave: join/leave messages, join roles, verification captcha/button/modal with fixed role-loop and missing-config behavior.
6. Moderation/Safety Wave: warnings/admin/safe-server with permission tightening and URL normalization.
7. Economy Wave: coin/sign/settings first, then rank/profile renderers, then shop/gacha/games with atomic balance and inventory flows.
8. Poll Wave: poll create/vote/result/export with atomic vote storage and legacy component compatibility.
9. XP/Voice Wave: message/voice event ownership, rank renderer, voice session reconciliation, and bounded workers.
10. Scheduler Wave: cron/birthday/work payout with Mongo leases and idempotent runs.

## Known Limitations

- No high-risk feature group was implemented in this wave.
- No Mongo repository writes were added.
- No real command sync apply was run after adding `/info user` and `/info guild`.
- Real Discord runtime for `/info user` and `/info guild` should be staged after guild-scoped command sync updates the `/info` definition.

## Commands Run

- `GOCACHE=... go fmt ./...`
- `GOCACHE=... go test ./internal/discord/features/utility ./internal/discord/commands ./internal/discord/interactions ./internal/adapters/discordgo ./internal/app`
- Boundary scans for core/feature imports and bot/runtime command/Mongo mutation calls.

## Next Recommended Step

Run the full acceptance suite, then start Platform Wave A. Do not begin ticket/economy/XP/poll repositories until the Mongo catalog is corrected to production collection names and repository contract tests exist.
