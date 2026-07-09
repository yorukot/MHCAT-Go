# Wave 2 Notes

Status: Wave 2 complete.

## Gate C Review for Wave 1

- Wave 1 boundary check result: passed before Wave 2 coding. `cmd/mhcat-bot` has no command registration path, no feature handlers, and no component/modal routing.
- Config redaction status: still safe. Redaction helpers and logger wrapper redact Discord tokens, webhook URLs, Mongo URI passwords, API-key-like values, and long secret-looking strings.
- Message Content default: disabled by default through `MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=false`.
- Bot command registration: `cmd/mhcat-bot` performs no command registration.
- Mongo behavior: Mongo adapter still performs client creation, connect, ping, and disconnect only. No indexes, feature reads, or feature writes.
- Core boundary: `internal/core/**` has no DiscordGo or MongoDB driver imports.
- Wave 1 issues requiring fix before Wave 2: none found.

## Files Created

- `cmd/mhcat-command-sync/main.go`
- `cmd/mhcat-command-sync/main_test.go`
- `internal/config/command_sync.go`
- `internal/config/command_sync_test.go`
- `internal/discord/commands/definition.go`
- `internal/discord/commands/registry.go`
- `internal/discord/commands/registry_test.go`
- `internal/discord/commands/diff.go`
- `internal/discord/commands/diff_test.go`
- `internal/discord/commands/sync.go`
- `internal/discord/commands/sync_test.go`
- `internal/discord/commands/validate.go`
- `internal/discord/commands/validate_test.go`
- `internal/discord/interactions/interaction.go`
- `internal/discord/interactions/router.go`
- `internal/discord/interactions/router_test.go`
- `internal/discord/interactions/middleware.go`
- `internal/discord/interactions/middleware_test.go`
- `internal/discord/responses/responder.go`
- `internal/discord/responses/state.go`
- `internal/discord/responses/state_test.go`
- `internal/discord/responses/errors.go`
- `internal/adapters/discordgo/command_sync.go`
- `internal/adapters/discordgo/responder.go`
- `internal/adapters/discordgo/interaction_adapter.go`
- `internal/testutil/fakediscord/command_sync_client.go`
- `internal/testutil/fakediscord/responder.go`
- `internal/testutil/fakediscord/interactions.go`
- `testdata/commands/valid_registry.json`
- `testdata/commands/invalid_registry_duplicate.json`
- `testdata/commands/remote_commands.json`

## Files Updated

- `.env.example`
- `Makefile`
- `README.md`
- `docs/10-feature-parity-checklist.md`
- `docs/11-operational-runbook.md`
- `docs/17-wave-2-notes.md`

## Command Sync Design

- `mhcat-command-sync` is a separate CLI and is not called by `cmd/mhcat-bot`.
- Production registry is intentionally empty in Wave 2. Real legacy command definitions are deferred to feature parity waves.
- CLI loads env, applies explicit flags, validates command sync config, fetches remote commands, computes a diff plan, prints the plan, and only writes when `--apply` is set.
- DiscordGo command sync adapter wraps REST application command list/create/edit/delete/bulk overwrite methods and never opens the Gateway.
- DiscordGo REST command APIs used here are not context-aware; the adapter checks `context.Context` before and after calls and isolates the limitation behind the `commands.SyncClient` interface.

## Diff Algorithm

- Desired local registry and remote command list are compared by command type plus name.
- Stable hashes are JSON/SHA-256 over normalized command definitions.
- Hashing ignores remote-only fields: ID, application ID, guild ID, version.
- Hashing ignores local-only fields: disabled, hidden, internal, docs URL.
- Planned operations are deterministic and sorted for stable text/JSON output.
- Unknown remote commands are skipped by default.
- Owned remote deletion is only planned when `AllowDelete=true`.
- Bulk overwrite is not a normal diff path and is always marked dangerous.

## Safety Flags

- CLI defaults to dry-run, regardless of `MHCAT_COMMAND_SYNC_DRY_RUN=false`; writes require explicit `--apply`.
- Deletion requires both `--apply` and `--allow-delete`.
- Bulk overwrite requires both `--apply` and `--allow-bulk-overwrite`; no automatic bulk overwrite path is implemented.
- `MHCAT_DISCORD_APPLICATION_ID` is required for command sync.
- Guild scope requires `MHCAT_COMMAND_SYNC_GUILD_ID`; global scope rejects a guild ID to avoid accidental mixed-scope sync.
- Application ID and guild ID are treated as non-secret config; Discord token is never printed.

## Responder Design

- `internal/discord/responses` defines the responder interface and state machine.
- State prevents double initial responses, defer-after-reply, edit-before-initial-response, follow-up-before-initial-response, and ephemeral state changes after defer.
- Safe error responses use a generic ephemeral message and do not expose internal errors.
- DiscordGo responder shell maps reply, defer, edit original, follow-up, and safe error behavior to DiscordGo interaction methods.

## Router Design

- `internal/discord/interactions` routes slash and autocomplete interactions by exact command name.
- Component and modal routing use structured keys, not broad substring matching.
- Legacy custom ID parsing is intentionally not implemented in Wave 2; it belongs to Wave 4.
- Middleware supports deterministic chaining, timeout context, panic recovery, logging, and a permission checker shell.
- Permission shell has no hardcoded owner, guild, role, or user IDs.

## Tests Added

- Command sync config tests: required application ID, guild scope requirements, defaults, global-scope validation, legacy token alias.
- Command registry/validation tests: empty registry, valid chat input/user commands, duplicate names, invalid names, missing descriptions, option count, duplicate options, required option order, deterministic ordering.
- Diff tests: create, unchanged, update, skipped delete, allowed delete, unknown remote skip, local-only/remote-only hash ignoring, deterministic output, dangerous bulk overwrite.
- Sync tests: dry-run no writes, apply create/update with fake client, no delete without allow-delete, no token-like plan output.
- CLI tests: missing application ID, missing guild ID, dry-run no writes, apply create/update, no delete without allow-delete, raw token not printed.
- Responder tests: reply once, double reply fail, defer/edit, defer/follow-up, invalid pre-initial edit/follow-up, safe error response, ephemeral preservation.
- Router/middleware tests: exact slash route, unknown route, component/modal key routing, deterministic middleware order, panic recovery, timeout context, permission checker invocation.

## Commands Run

- `GOCACHE=/private/tmp/mhcat-refactor-gocache GOMODCACHE=/private/tmp/mhcat-refactor-gomodcache go fmt ./...`
- `GOCACHE=/private/tmp/mhcat-refactor-gocache GOMODCACHE=/private/tmp/mhcat-refactor-gomodcache go test ./...`
- `GOCACHE=/private/tmp/mhcat-refactor-gocache GOMODCACHE=/private/tmp/mhcat-refactor-gomodcache go vet ./...`
- `GOCACHE=/private/tmp/mhcat-refactor-gocache GOMODCACHE=/private/tmp/mhcat-refactor-gomodcache go build ./cmd/mhcat-bot`
- `GOCACHE=/private/tmp/mhcat-refactor-gocache GOMODCACHE=/private/tmp/mhcat-refactor-gomodcache go build ./cmd/mhcat-command-sync`
- `GOCACHE=/private/tmp/mhcat-refactor-gocache GOMODCACHE=/private/tmp/mhcat-refactor-gomodcache make check`
- `MHCAT_DISCORD_TOKEN= MHCAT_DISCORD_APPLICATION_ID= MHCAT_COMMAND_SYNC_GUILD_ID= GOCACHE=/private/tmp/mhcat-refactor-gocache GOMODCACHE=/private/tmp/mhcat-refactor-gomodcache go run ./cmd/mhcat-command-sync`

## Known Limitations

- No real MHCAT feature commands are defined yet.
- No legacy command behavior, component grammar, modal grammar, or autocomplete behavior is implemented yet.
- No command ownership marker exists for remote Discord commands; unknown remote commands are skipped.
- No command sync bulk overwrite execution path is wired from CLI.
- DiscordGo REST command sync methods do not accept context directly; context is checked around each call.
- Mongo live audit has not run, and no Mongo indexes/repositories/data writes were added.

## Next Recommended Step

Start Wave 3 only after reviewing this Wave 2 output. Wave 3 should add Mongo repository infrastructure, index bootstrap planning code, audit CLI, atomic update helpers, repository contract tests, and no feature data mutation by default.
