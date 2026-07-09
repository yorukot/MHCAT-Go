# Platform Wave A Shared Discord Runtime Infrastructure

## Gate C Review

- Legacy `MHCAT/` source remained read-only.
- `cmd/mhcat-bot` still does not register or sync Discord application commands.
- Discord Gateway remains disabled by default.
- Message Content remains disabled by default.
- No Mongo feature write or index creation was added.
- Usage tracking remains no-op.
- Runtime-routable feature commands remain the utility slice: `help`, `ping`, `info bot`, `info shard`, `info user`, and `info guild`.
- Feature services and `internal/core/**` remain free of DiscordGo and MongoDB driver imports.

## Files Created

- `internal/discord/events/events.go`
- `internal/discord/events/events_test.go`
- `internal/core/ports/discord_actions.go`
- `internal/adapters/discordgo/gateway_events.go`
- `internal/adapters/discordgo/gateway_events_test.go`
- `internal/adapters/discordgo/side_effects.go`
- `internal/adapters/discordgo/side_effects_test.go`
- `docs/31-platform-wave-a-notes.md`

## Files Updated

- `.env.example`
- `README.md`
- `internal/app/app.go`
- `internal/app/app_test.go`
- `internal/config/config.go`
- `internal/config/config_test.go`
- `internal/config/env.go`
- `internal/discord/commands/definition.go`
- `internal/discord/commands/validate.go`
- `internal/discord/commands/validate_test.go`
- `internal/discord/interactions/command_options.go`
- `internal/discord/interactions/command_options_test.go`
- `internal/discord/responses/responder.go`
- `internal/adapters/discordgo/command_sync.go`
- `internal/adapters/discordgo/command_sync_test.go`
- `internal/adapters/discordgo/intents.go`
- `internal/adapters/discordgo/intents_test.go`
- `internal/adapters/discordgo/interaction_adapter.go`
- `internal/adapters/discordgo/responder.go`
- `internal/adapters/discordgo/responder_test.go`

## Command Model Gaps Closed

- Added application command option choices.
- Added channel type constraints for channel options.
- Added validation for choice count, duplicate choice names, invalid choice-bearing option types, and invalid channel type usage.
- Added deterministic DiscordGo conversion round-trip coverage for choices and channel types.
- Added runtime slash option parsing support for number, channel, role, and mentionable values.

## Response Infrastructure

- Added attachment support through internal responder messages and DiscordGo interaction responses.
- Added allowed-mentions support with safe default suppression when no explicit policy is provided.
- Added disabled component and default select-option mapping.
- Existing responder state machine remains the path for all runtime responses.

## Gateway Event Dispatch

- Added a driver-agnostic event dispatcher for ready/resumed, message create/update/delete, reaction add/remove, guild member add/remove, and voice state events.
- Added DiscordGo gateway event adapters behind explicit event options.
- Gateway event handlers are registered only when the gateway is explicitly enabled.
- Message, reaction, guild-member, and voice-state event subscriptions are opt-in through config:
  - `MHCAT_DISCORD_GUILD_MESSAGES_INTENT`
  - `MHCAT_DISCORD_GUILD_MESSAGE_REACTIONS_INTENT`
  - `MHCAT_DISCORD_GUILD_MEMBERS_INTENT`
  - `MHCAT_DISCORD_VOICE_STATE_INTENT`
- Message Content remains a separate privileged flag and stays disabled by default.

## Discord Side-Effect Ports

- Added core-safe ports for future feature actions:
  - message send/edit/delete;
  - channel create/delete;
  - role add/remove;
  - member move/nickname/kick/ban;
  - audit-log reads.
- Added a DiscordGo adapter shell using context-aware REST calls where DiscordGo exposes them.
- No feature handler uses these ports yet. They exist so future ticket, onboarding, role, moderation, and logging waves do not import DiscordGo directly.

## Tests Added

- Command option choice and channel type validation tests.
- DiscordGo command conversion round-trip tests.
- Runtime command option parsing tests for number/channel/role/mentionable values.
- Responder attachment and allowed-mentions tests.
- Intent default and explicit event-intent tests.
- App gateway event registration tests.
- Event dispatcher and DiscordGo gateway event adapter tests.
- Side-effect adapter shell tests for missing-session and safe mention defaults.

## Known Limitations

- No high-risk feature group was implemented.
- No feature currently owns message, reaction, guild-member, or voice-state events.
- Side-effect ports are infrastructure only and do not change runtime behavior yet.
- No rate-limit queue or per-guild work scheduler was added in this wave.
- Mongo collection catalog corrections and stateful repository contracts remain for Platform Wave B.
- Command registration remains manual through `mhcat-command-sync`; bot startup still never syncs commands.

## Commands Run

- `GOCACHE=... go fmt ./...`
- `GOCACHE=... go test ./internal/adapters/discordgo ./internal/discord/commands ./internal/discord/interactions ./internal/discord/responses ./internal/discord/events ./internal/config ./internal/app`
- `GOCACHE=... go test ./...`
- `GOCACHE=... go vet ./...`
- `GOCACHE=... go build ./cmd/mhcat-bot`
- `GOCACHE=... go build ./cmd/mhcat-command-sync`
- `GOCACHE=... go build ./cmd/mhcat-mongo-audit`
- `GOCACHE=... go build ./cmd/mhcat-mongo-index`
- `GOCACHE=... go build ./cmd/mhcat-staging-preflight`
- `GOCACHE=... make check`

## Next Recommended Step

Start Platform Wave B: correct the Mongo collection catalog to production plural names, add collection-name contract tests, expand repository contracts by collection group, and keep every write/index path dry-run or explicitly gated until the live audit findings are reconciled.
