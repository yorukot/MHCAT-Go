# Wave 5.1 Notes

Status: implemented and verified.

## Gate C Review for Wave 4

- Legacy source status: clean. `MHCAT/` reports `## main...origin/main`.
- Bot command registration: `cmd/mhcat-bot` and `internal/app` have no command registration path.
- Discord Gateway default: `DefaultDiscordEnableGateway=false`; gateway open remains config-gated.
- Mongo feature writes: none implemented.
- Mongo index creation: only the Wave 3 index CLI has apply-gated `CreateOne`; no index apply was run.
- Router broad matching: `internal/discord/interactions` has no runtime broad custom ID routing; `strings.Contains` only appears in tests that assert safe errors do not leak raw IDs.
- Parser: `mhcat:v1:<feature>:<action>:<payload>` is implemented and length-checked.
- Core boundary: `internal/core/**` has no DiscordGo or MongoDB driver imports.
- Wave 4 issues requiring fix before Wave 5.1: none found.

## Legacy Re-check

Read-only files inspected:

- `docs/01-legacy-behavior-spec.md`
- `docs/10-feature-parity-checklist.md`
- `docs/12-component-modal-grammar.md`
- `MHCAT/slashCommands/實用工具/ping.js`
- `MHCAT/slashCommands/實用工具/help.js`
- `MHCAT/slashCommands/實用工具/info.js`
- `MHCAT/slashCommands/實用工具/translate.js`
- `MHCAT/slashCommands/實用工具/chat.js`
- `MHCAT/slashCommands/實用工具/chat_delete.js`
- `MHCAT/functions/menu.js`

`MHCAT/commands/` is absent in the local checkout.

## Selected Features

### ping

- Legacy command name: `ping`
- Legacy file: `MHCAT/slashCommands/實用工具/ping.js`
- Trigger type: slash command
- Options: none
- Permissions: public
- Response shape: direct text reply containing `Pong!` and latency in milliseconds
- Components: none
- Mongo reads/writes: none
- External API usage: none
- Reason selected: stateless, public, no Mongo, no external API, no components

### help

- Legacy command name: `help`
- Legacy file: `MHCAT/slashCommands/實用工具/help.js`
- Trigger type: slash command
- Options: optional string `指令名稱`
- Permissions: public
- Response shape: deferred help overview or command detail; legacy output is embed/select-menu based
- Components: legacy select menu `helphelphelphelpmenu`; Wave 5.8 restores the legacy select/menu UI for runtime `/help`
- Mongo reads/writes: none
- External API usage: none
- Reason selected: read-only registry-backed command; validates feature pipeline

### info bot subset

- Legacy command name: `info`
- Legacy file: `MHCAT/slashCommands/實用工具/info.js`
- Trigger type: slash command subcommand `bot`
- Options: `bot` subcommand only in Wave 5.1
- Permissions: public
- Response shape: deferred bot/system status summary
- Components: legacy refresh button `botinfoupdate` is documented; Wave 5.9 restores the runtime refresh button route and legacy bot-info embed UI
- Mongo reads/writes: none
- External API usage: none
- Reason selected: can be backed by a `BotInfoProvider` port and tested without Discord Gateway

## Excluded Features

- `info user`: excluded because legacy fetches guild members and needs richer Discord guild/member adapter behavior.
- `info guild`: excluded because legacy reads guild details, icons, banners, owner IDs, and locale data from Discord gateway/cache objects.
- `info shard`: excluded because legacy includes refresh components and shard broadcast behavior.
- `translate` / `翻譯`: excluded because it calls Google Translate libraries and needs timeout/defer/external API policy.
- `自動聊天頻道` and `自動聊天頻道刪除`: excluded because they read/write Mongo `chat`, require ManageMessages permissions, and connect to Message Content/chat worker behavior.
- Ticket, poll, verification, coin, XP, gacha, work, lottery, cron, moderation, ChatGPT, dashboard, and external API feature groups: explicitly out of Wave 5.1 scope.

## Files Created

- `internal/core/features/feature.go`
- `internal/core/features/registry.go`
- `internal/core/features/registry_test.go`
- `internal/core/features/boundary_test.go`
- `internal/core/ports/usage.go`
- `internal/core/ports/botinfo.go`
- `internal/core/services/utility/help.go`
- `internal/core/services/utility/help_test.go`
- `internal/core/services/utility/ping.go`
- `internal/core/services/utility/ping_test.go`
- `internal/core/services/utility/status.go`
- `internal/core/services/utility/status_test.go`
- `internal/core/services/utility/golden_test.go`
- `internal/discord/commands/builtin_registry.go`
- `internal/discord/commands/builtin_registry_test.go`
- `internal/discord/components/help_menu.go`
- `internal/discord/components/help_menu_test.go`
- `internal/discord/features/utility/module.go`
- `internal/discord/features/utility/handlers.go`
- `internal/discord/features/utility/help_handler.go`
- `internal/discord/features/utility/help_handler_test.go`
- `internal/discord/features/utility/ping_handler.go`
- `internal/discord/features/utility/ping_handler_test.go`
- `internal/discord/features/utility/status_handler.go`
- `internal/discord/features/utility/status_handler_test.go`
- `internal/discord/features/utility/module_test.go`
- `internal/discord/features/utility/handlers_test.go`
- `internal/discord/features/utility/boundary_test.go`
- `internal/adapters/discordgo/botinfo.go`
- `internal/adapters/discordgo/botinfo_test.go`
- `internal/adapters/usage/noop.go`
- `internal/adapters/usage/noop_test.go`
- `internal/testutil/fakefeatures/registry.go`
- `internal/testutil/fakebotinfo/botinfo.go`
- `internal/testutil/fakeusage/usage.go`
- `testdata/features/utility_help_golden.json`
- `testdata/features/utility_ping_golden.json`
- `testdata/features/utility_status_golden.json`

## Files Updated

- `internal/discord/commands/registry.go`
- `internal/discord/commands/validate.go`
- `internal/discord/commands/validate_test.go`
- `internal/discord/interactions/interaction.go`
- `internal/adapters/discordgo/interaction_adapter.go`
- `internal/testutil/fakediscord/interactions.go`
- `README.md`
- `Makefile`
- `docs/06-architecture-decision-records.md`
- `docs/09-risk-register.md`
- `docs/10-feature-parity-checklist.md`
- `docs/11-operational-runbook.md`
- `docs/20-wave-5.1-notes.md`

## Feature Module Design

- `internal/core/features.Registry` collects feature modules deterministically by module name.
- Feature modules expose local command definitions and register handlers on the internal interaction router.
- Duplicate command definitions fail through the existing command registry validator.
- `internal/discord/commands.DefaultRegistry` now returns only the Wave 5.1 utility definitions: `help`, `info` with `bot`, and `ping`.
- `cmd/mhcat-command-sync` automatically sees those definitions through the existing registry loader, but `cmd/mhcat-bot` still does not sync or register commands.

## Handler/Responder Design

- `ping` replies immediately through the responder state machine.
- `help` defers and edits the original response because it renders registry-backed help output.
- `info bot` defers and follows up with the legacy bot-info embed because it reads through a provider port. Wave 5.9 also restores the legacy `botinfoupdate` refresh button route.
- The help component route uses parsed `mhcat:v1:help:category:<payload>` route keys and returns the registry-backed overview.
- Unsupported `info` subcommands return a safe “not implemented” response instead of claiming parity.
- Handlers are tested with fake interactions and fake responders only.

## Usage Tracking Temporary Behavior

- Added `ports.UsageTracker`.
- Added production-safe `adapters/usage.NoopTracker`.
- Added fake usage tracker for tests.
- Wave 5.1 does not increment legacy `all_use_count`.
- This is a temporary parity gap tracked in `docs/09-risk-register.md`; a later Wave 5.x Mongo repository must implement an atomic usage counter after Mongo audit/catalog approval.

## Behavior Parity Notes

- `ping`: preserves the public command name and `Pong! <ms>` response shape. Emoji thresholds follow legacy good/idle/bad timing boundaries.
- `help`: preserves command name, description, and optional legacy option name `指令名稱`. Wave 5.1 initially used a registry-backed text placeholder; Wave 5.8 replaces that runtime output with the legacy embed/select menu and command detail interface from `help.js` and `functions/menu.js`.
- `info`: preserves command name/localizations and implements only the `bot` subcommand. `user`, `guild`, and `shard` are intentionally excluded until Discord cache/fetch/shard behavior is modeled.
- `translate`, `自動聊天頻道`, and `自動聊天頻道刪除` are excluded because they involve external APIs, Mongo writes, permissions, or Message Content/chat worker behavior.

## Tests Added

- Feature registry tests for empty registry, deterministic module order, duplicate command failure, and route registration.
- Command registry tests for built-in definitions, deterministic ordering, dry-run create planning, local-only stripping, and Unicode legacy help option validation.
- Utility service tests and golden fixtures for help, ping, and status output.
- Handler tests for responder state usage, context cancellation, usage tracking, safe not-found/degraded responses, permission middleware, and panic recovery.
- Help component builder tests for versioned custom ID use and hidden command exclusion.
- DiscordGo bot info provider tests for disconnected/degraded and cached-count behavior.
- Boundary tests for core and utility feature import rules.

## Commands Run

- `go fmt ./...`: passed. One final run required approved Go build-cache access after the sandbox blocked cache reads.
- `go test ./...`: passed.
- `go vet ./...`: passed.
- `go build ./cmd/mhcat-bot`: passed.
- `go build ./cmd/mhcat-command-sync`: passed.
- `go build ./cmd/mhcat-mongo-audit`: passed.
- `go build ./cmd/mhcat-mongo-index`: passed.
- `make check`: passed.
- `go test ./internal/core/features ./internal/core/services/utility ./internal/discord/features/utility ./internal/discord/commands ./internal/discord/interactions`: passed.
- `go run ./cmd/mhcat-command-sync` without env: exited non-zero with missing `MHCAT_DISCORD_TOKEN`, no panic, no Discord mutation.

## Known Limitations

- The command registry contains only Wave 5.1 utility definitions, not all 74 legacy commands.
- `help` output was text-based in the initial Wave 5.1 implementation; Wave 5.8 restores the legacy embed/select UI for runtime responses.
- `info` is partial; only `info bot` is implemented. Wave 5.9 restores the legacy bot-info embed/button UI for that subset.
- Usage tracking is no-op in production and does not write Mongo `all_use_count`.
- Runtime gateway event-loop integration for feature handlers is still not wired into `cmd/mhcat-bot`; the feature pipeline is tested in-process.
- No production Mongo feature repositories or writes exist.

## Next Recommended Step

Wave 5.2 has wired the runtime `InteractionCreate` path for the existing utility subset. The next wave should either run an explicit staging gateway smoke test with a staging token/guild, expand the utility slice with `info shard` only after shard/gateway cache behavior is modeled, or start a similarly low-risk read-only feature with complete behavior fixtures. Do not start Mongo-writing features until live audit results and repository contracts for that collection are approved.
