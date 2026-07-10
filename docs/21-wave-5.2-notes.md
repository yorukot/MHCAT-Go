# Wave 5.2 Notes

Status: implemented.

## Gate C Review for Wave 5.1

- Legacy source status: clean. `MHCAT/` reports `## main...origin/main`.
- Bot command registration: `cmd/mhcat-bot` and `internal/app` have no command registration path.
- Discord Gateway default: `DefaultDiscordEnableGateway=false`; gateway open remains config-gated.
- Mongo feature writes: none implemented.
- Mongo index creation: only the Wave 3 index CLI has apply-gated `CreateOne`; no index apply was run.
- Implemented feature scope: only `help`, `ping`, and safe `info bot` utility commands.
- High-risk feature groups: ticket, poll, verification, role button, coin, XP, gacha, work, cron, moderation, ChatGPT, external APIs, and dashboard integration remain unimplemented.
- Usage tracking: production implementation is no-op and does not write Mongo.
- Feature services: no DiscordGo or MongoDB driver imports.
- Core boundary: `internal/core/**` has no DiscordGo or MongoDB driver imports.
- Wave 5.1 issues requiring fix before Wave 5.2: none found.

## Files Created

- `internal/app/wiring.go`
- `internal/app/wiring_test.go`
- `internal/app/runtime.go`
- `internal/app/runtime_test.go`
- `cmd/mhcat-bot/main_test.go`
- `internal/discord/runtime/dispatcher.go`
- `internal/discord/runtime/dispatcher_test.go`
- `internal/discord/runtime/gateway.go`
- `internal/discord/runtime/gateway_test.go`
- `internal/discord/runtime/events.go`
- `internal/discord/runtime/events_test.go`
- `internal/discord/interactions/command_options.go`
- `internal/discord/interactions/command_options_test.go`
- `internal/discord/interactions/runtime_adapter.go`
- `internal/discord/interactions/runtime_adapter_test.go`
- `internal/adapters/discordgo/gateway.go`
- `internal/adapters/discordgo/gateway_test.go`
- `internal/adapters/discordgo/interaction_runtime.go`
- `internal/adapters/discordgo/interaction_runtime_test.go`
- `internal/adapters/discordgo/responder_runtime.go`
- `testdata/interactions/slash_help.json`
- `testdata/interactions/slash_help_detail.json`
- `testdata/interactions/slash_ping.json`
- `testdata/interactions/slash_info_bot.json`
- `testdata/interactions/component_help_menu.json`
- `testdata/interactions/invalid_unknown_command.json`

## Files Updated

- `.env.example`
- `Makefile`
- `README.md`
- `internal/app/app.go`
- `internal/app/app_test.go`
- `internal/adapters/discordgo/interaction_adapter.go`
- `internal/adapters/discordgo/session.go`
- `internal/config/config.go`
- `internal/config/config_test.go`
- `internal/config/env.go`
- `internal/config/validation.go`
- `internal/discord/interactions/interaction.go`
- `internal/discord/interactions/middleware.go`
- `internal/discord/interactions/middleware_test.go`
- `docs/09-risk-register.md`
- `docs/10-feature-parity-checklist.md`
- `docs/11-operational-runbook.md`
- `docs/20-wave-5.1-notes.md`

## Runtime Interaction Pipeline

The Wave 5.2 runtime path is:

```txt
DiscordGo InteractionCreate
  -> internal/adapters/discordgo.RuntimeInteraction
  -> internal/discord/interactions.Interaction
  -> internal/discord/runtime.Dispatcher
  -> middleware chain
  -> internal/discord/interactions.Router
  -> utility feature handler
  -> internal/discord/responses.Responder
  -> DiscordGo interaction response adapter
```

The runtime router currently wires only the Wave 5.1 utility commands: `/help`, `/help <command>`, `/ping`, and `/info bot`. Feature handlers receive internal interaction and responder abstractions only; DiscordGo interaction objects and tokens remain inside the DiscordGo adapter.

## Gateway Config Behavior

- `MHCAT_DISCORD_ENABLE_GATEWAY=false` remains the default.
- `MHCAT_DISCORD_GATEWAY_CONNECT_TIMEOUT=15s` bounds gateway `Open`.
- `MHCAT_DISCORD_INTERACTION_TIMEOUT=2500ms` bounds each routed interaction.
- `MHCAT_DISCORD_GATEWAY_SMOKE_TEST=false` remains the default.
- `MHCAT_DISCORD_GATEWAY_SMOKE_TEST_TIMEOUT=30s` bounds one-shot smoke mode.
- `MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=false` remains the default.

When gateway is disabled, app startup still validates config, connects/pings Mongo, creates the Discord session object, registers feature modules in memory, and shuts down without opening Discord Gateway.

When gateway is enabled, `cmd/mhcat-bot` registers the `InteractionCreate` handler and opens the gateway. It does not register, update, delete, or bulk-overwrite Discord application commands.

## Gateway Smoke Test Behavior

Smoke mode requires both `MHCAT_DISCORD_ENABLE_GATEWAY=true` and `MHCAT_DISCORD_GATEWAY_SMOKE_TEST=true`. It opens the gateway, waits for the Discord ready signal or timeout, logs a safe ready message, and shuts down. It does not send messages, register commands, write feature data, or create Mongo indexes.

## Interaction Adapter Design

`internal/adapters/discordgo` converts DiscordGo interaction payloads into internal models:

- slash command name, subcommand group, subcommand, and typed options;
- actor user/guild/role IDs;
- channel ID, locale, and guild locale;
- component custom ID plus selected values;
- modal custom ID plus submitted fields.

The internal interaction model does not expose the Discord interaction token. Unsupported interaction types return typed runtime errors. Component and modal paths use the Wave 4 custom ID parser through router integration instead of broad substring matching.

## Router/Responder Runtime Design

`internal/discord/runtime.Dispatcher` wraps the router and ensures runtime errors are turned into safe responder errors if a handler has not already responded. The router middleware chain includes:

- panic recovery;
- interaction timeout;
- permission checker shell;
- usage tracking hook;
- safe structured route logging.

The DiscordGo responder implementation is still the only layer that calls Discord interaction response APIs. Successful quick utility handlers reply exactly once through the responder state machine.

## Usage Tracking Behavior At Wave 5.2

At Wave 5.2, the runtime invoked usage tracking only after successful slash command handling and the production tracker was `adapters/usage.NoopTracker`, so it wrote no `all_use_count` data. This historical gap is now closed: the current runtime has a disabled-by-default `MHCAT_FEATURE_USAGE_TRACKING_ENABLED` gate backed by an atomic `all_use_counts` tracker in the global pre-handler slash middleware.

## Tests Added

- Config tests for gateway connect timeout, interaction timeout, smoke-test defaults, smoke-test validation, and env overrides.
- App wiring tests for gateway disabled/enabled behavior, interaction handler registration, smoke mode, shutdown, and timeout handling.
- Runtime dispatcher tests for `/help`, `/info bot`, unknown command safe errors, and context cancellation.
- Gateway wait tests for ready and timeout behavior.
- DiscordGo adapter tests for slash commands, subcommands, components, modals, selected values, unsupported interaction types, and token isolation.
- Command option parsing tests for string, integer, boolean, subcommand, subcommand group, missing optional values, unknown types, and malformed options.
- Middleware tests for no-op/fake usage tracking.

## Commands Run

- `go fmt ./...`
- `go test ./...`
- `go vet ./...`
- `go build ./cmd/mhcat-bot`
- `go build ./cmd/mhcat-command-sync`
- `go build ./cmd/mhcat-mongo-audit`
- `go build ./cmd/mhcat-mongo-index`
- `make check`
- `go test ./internal/app ./internal/discord/runtime ./internal/discord/interactions ./internal/adapters/discordgo ./internal/discord/features/utility`
- `go run ./cmd/mhcat-bot` without env: exited non-zero with missing config error, no panic, no secret output.
- `go run ./cmd/mhcat-command-sync` without env: exited non-zero with missing config error, no panic, no secret output.

## Known Limitations

- No live Discord smoke test was run; gateway mode is implemented but remains opt-in and unverified against a real bot token in this environment.
- `info bot` runtime output uses the legacy bot-info embed/button UI; values are degraded only when a concrete bot info provider cannot provide gateway/system state.
- At Wave 5.2, usage tracking was no-op and did not match legacy `all_use_count`; the later gated global tracker closes this limitation.
- Command definitions still cover only `help`, `ping`, and `info bot`.
- The bot still does not perform command sync at startup by design.
- No feature Mongo repositories or index writes were added.

## Next Recommended Step

Wave 5.3 added staging smoke guardrails and scripts. The next step is to run the real staging checklist with staging Discord/Mongo env, or continue with another low-risk read-only utility command only after behavior fixtures exist. Do not start Mongo-writing features until the live Mongo audit and collection-specific repository/index plans are approved.
