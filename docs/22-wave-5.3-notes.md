# Wave 5.3 Notes

Status: implemented; real staging smoke not run in this environment.

## Gate C Review for Wave 5.2

- Legacy source status: clean. `MHCAT/` reports `## main...origin/main`.
- Bot command sync/apply path: `cmd/mhcat-bot` has no command sync import or application command mutation path.
- Discord Gateway default: disabled by default.
- Message Content default: disabled by default.
- Runtime path: DiscordGo `InteractionCreate` is wired through adapter, dispatcher, router, middleware, handlers, and responder.
- Runtime-routable commands: only `/help`, `/help <command>`, `/ping`, and `/info bot`.
- High-risk feature groups: still unimplemented.
- Usage tracking: no-op default; no Mongo write.
- Mongo writes/indexes: no feature writes or index creation in bot startup.
- Feature services: no DiscordGo or MongoDB driver imports.
- Core boundary: `internal/core/**` has no DiscordGo or MongoDB driver imports.
- Wave 5.2 issue requiring fix before staging work: gateway smoke existed but was not staging-guarded; Wave 5.3 now requires staging mode plus explicit smoke allow.

## Files Created

- `.gitignore`
- `internal/config/staging.go`
- `internal/config/staging_test.go`
- `internal/discord/commands/ownership.go`
- `internal/discord/commands/ownership_test.go`
- `internal/discord/commands/staging_sync.go`
- `internal/discord/commands/staging_sync_test.go`
- `internal/app/smoke.go`
- `internal/app/smoke_test.go`
- `internal/testutil/fakestaging/command_sync.go`
- `internal/testutil/fakestaging/gateway.go`
- `internal/testutil/fakestaging/smoke.go`
- `scripts/staging/command-sync-dry-run.sh`
- `scripts/staging/command-sync-apply-guild.sh`
- `scripts/staging/gateway-smoke.sh`
- `scripts/staging/smoke-checklist.md`
- `docs/23-staging-smoke-runbook.md`

## Files Updated

- `.env.example`
- `Makefile`
- `README.md`
- `cmd/mhcat-command-sync/main.go`
- `cmd/mhcat-command-sync/main_test.go`
- `internal/config/config.go`
- `internal/config/env.go`
- `internal/config/validation.go`
- `internal/config/command_sync.go`
- `internal/config/config_test.go`
- `internal/app/app.go`
- `internal/app/app_test.go`
- `internal/discord/commands/definition.go`
- `internal/discord/commands/registry.go`
- `internal/discord/commands/builtin_registry.go`
- `internal/discord/commands/builtin_registry_test.go`
- `docs/06-architecture-decision-records.md`
- `docs/09-risk-register.md`
- `docs/10-feature-parity-checklist.md`
- `docs/11-operational-runbook.md`
- `docs/15-gate-b-architecture-freeze.md`
- `docs/21-wave-5.2-notes.md`

## Staging Safety Model

- Staging mode defaults off.
- Staging command apply defaults off.
- Staging gateway smoke defaults off.
- Staging command sync requires guild scope when staging mode is enabled.
- Apply mode now requires staging mode, explicit apply allow, guild scope, no delete, and no bulk overwrite.
- The staging command registry check allows only managed `help`, `ping`, and `info`.
- Command ownership metadata is local-only and stripped before stable hashing and Discord payload conversion.

## Command Sync Staging Rules

Dry-run:

```bash
scripts/staging/command-sync-dry-run.sh
```

Apply:

```bash
MHCAT_STAGING_MODE=true \
MHCAT_STAGING_ALLOW_COMMAND_APPLY=true \
scripts/staging/command-sync-apply-guild.sh
```

Apply can create/update managed utility commands in the configured staging guild. It cannot delete commands, bulk overwrite, or run global sync.

## Gateway Smoke Behavior

Gateway smoke now requires all of:

- `MHCAT_STAGING_MODE=true`
- `MHCAT_STAGING_ALLOW_GATEWAY_SMOKE=true`
- `MHCAT_DISCORD_ENABLE_GATEWAY=true`
- `MHCAT_DISCORD_GATEWAY_SMOKE_TEST=true`

`scripts/staging/gateway-smoke.sh` sets the gateway flags and relies on env for token and Mongo connectivity. The app still connects/pings Mongo during startup; Mongo is not optional for this smoke mode.

## Manual Smoke Checklist Result

Not run. No staging Discord token/application ID/guild ID and staging Mongo URI were present in this environment, and no approval was given to run a real Discord smoke.

## Tests Added

- Staging config defaults and guardrails.
- Staging apply rejects global scope, deletion, and bulk overwrite.
- Gateway smoke requires staging allow flag.
- Command ownership metadata is stripped from enabled definitions and stable hashes.
- Staging sync accepts only managed `help`, `ping`, and `info`.
- Staging sync skips unknown remote commands by default.
- CLI apply requires staging mode and staging allow.
- App smoke timeout uses staging timeout only when staging smoke is explicitly allowed.
- Scripts are scanned for secret-like literals.

## Commands Run

- `go fmt ./...`
- `go test ./...`
- `go vet ./...`
- `go build ./cmd/mhcat-bot`
- `go build ./cmd/mhcat-command-sync`
- `go build ./cmd/mhcat-mongo-audit`
- `go build ./cmd/mhcat-mongo-index`
- `make check`
- `go run ./cmd/mhcat-command-sync` without env: exited non-zero with missing config error, no panic, no secret output.
- `go run ./cmd/mhcat-bot` without env: exited non-zero with missing config error, no panic, no secret output.
- Staging env probe for token/application/guild/Mongo required vars: failed, so real staging dry-run/apply/gateway smoke were not run.

## Known Limitations

- Real staging smoke was not run.
- Manual interaction validation for `/help`, `/ping`, and `/info bot` remains pending.
- No Mongo feature write verification against live staging was possible without staging Mongo credentials.
- No global command sync path was exercised.

## Next Recommended Step

Run the staging checklist in `docs/23-staging-smoke-runbook.md` with a staging Discord application, staging guild, and staging Mongo URI. Record the result in a local `.smoke/` file or update this note with sanitized pass/fail facts only.
