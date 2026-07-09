# Wave 5.5 Staging Preflight

Status: implemented.

## Goal

Add a local-only readiness check so staging operators can verify required env and unsafe flag state before running Discord command sync or gateway smoke.

## Behavior

`mhcat-staging-preflight` performs no network calls and no writes. It checks:

- `MHCAT_DISCORD_TOKEN` is present;
- `MHCAT_DISCORD_APPLICATION_ID` is present;
- `MHCAT_STAGING_GUILD_ID` is present;
- `MHCAT_MONGODB_URI` is present;
- `MHCAT_MONGODB_DATABASE` is present;
- `MHCAT_COMMAND_SYNC_SCOPE` is unset or `guild`;
- command deletion and bulk overwrite flags are disabled;
- Message Content intent is disabled;
- optional application ID pin matches when configured.

Secrets are redacted in output. Private IDs are reported only as present/missing.

## Usage

```bash
go run ./cmd/mhcat-staging-preflight --format text
go run ./cmd/mhcat-staging-preflight --format json
```

Exit code:

- `0`: all required checks pass;
- non-zero: at least one required or unsafe check fails.

## Verification

Run from `MHCAT-REFACTOR` on 2026-07-04:

```bash
go fmt ./...
go test ./...
go vet ./...
go build ./cmd/mhcat-bot
go build ./cmd/mhcat-command-sync
go build ./cmd/mhcat-mongo-audit
go build ./cmd/mhcat-mongo-index
go build ./cmd/mhcat-staging-preflight
make check
```

Result: all commands passed.

Missing-env safety checks:

```bash
go run ./cmd/mhcat-staging-preflight --format text
go run ./cmd/mhcat-staging-preflight --format json
go run ./cmd/mhcat-command-sync
go run ./cmd/mhcat-bot
```

Result: all commands exited non-zero with clear missing-config errors and no panic. The preflight output reported missing Discord token, application ID, staging guild ID, Mongo URI, and Mongo database while confirming unsafe delete, bulk overwrite, and Message Content flags were disabled.

Boundary scans:

- Legacy `MHCAT/` git status was clean on `main`.
- `MHCAT-REFACTOR/` is not a git repository in this workspace; it remains a separate generated refactor tree.
- No command mutation or Mongo write/index calls were found in `cmd/mhcat-bot`, `cmd/mhcat-staging-preflight`, app runtime, Discord runtime, feature, interaction, or core packages.
- No DiscordGo or MongoDB driver imports were found in `internal/core/**`, utility feature services, or `cmd/mhcat-staging-preflight`.
- No obvious token, webhook, or Mongo URI password literals were found in the staging preflight tool, scripts, README, `.env.example`, or Wave 5.3-5.5 staging docs.
- No `migrations/` directory exists.

## Boundary Notes

- No `cmd/mhcat-bot` behavior changed.
- No command sync apply behavior changed.
- No Discord or Mongo network call is made by preflight.
- No Mongo feature write or index creation is possible from preflight.
- Real staging smoke remains pending until staging env is available.
