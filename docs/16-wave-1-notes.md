# Wave 1 Notes

Status: implemented as skeleton only.

## Files Created

- `go.mod`
- `cmd/mhcat-bot/main.go`
- `internal/app/app.go`
- `internal/app/lifecycle.go`
- `internal/app/shutdown.go`
- `internal/app/app_test.go`
- `internal/config/config.go`
- `internal/config/env.go`
- `internal/config/redaction.go`
- `internal/config/validation.go`
- `internal/config/config_test.go`
- `internal/config/redaction_test.go`
- `internal/observability/logger.go`
- `internal/observability/logger_test.go`
- `internal/adapters/mongo/client.go`
- `internal/adapters/mongo/client_test.go`
- `internal/adapters/discordgo/session.go`
- `internal/adapters/discordgo/intents.go`
- `internal/adapters/discordgo/intents_test.go`
- `internal/core/ports/clock.go`
- `internal/core/ports/health.go`
- `.env.example`
- `README.md`
- `Makefile`
- `docs/16-wave-1-notes.md`

## Module Path

Chosen module path:

```txt
github.com/yorukot/MHCAT/MHCAT-REFACTOR
```

This is the requested path. If future release tooling requires a lowercase path, change it in an ADR before package imports spread further.

## Dependencies Added

- `github.com/bwmarrin/discordgo`
- `go.mongodb.org/mongo-driver/v2`

The standard library is used for config loading and logging.

## Config Decisions

- Required env: `MHCAT_DISCORD_TOKEN`, `MHCAT_MONGODB_URI`, `MHCAT_MONGODB_DATABASE`.
- Legacy aliases: `TOKEN`, `MONGOOSE_CONNECTION_STRING`.
- New `MHCAT_*` env vars take precedence.
- Differing alias values are represented as warnings with redacted values only.
- Gateway, Message Content, and Guild Members intents are disabled by default.

## Mongo Behavior

Wave 1 Mongo adapter can only:

- create a client;
- connect;
- ping;
- disconnect.

It does not select feature collections, create indexes, read feature documents, write documents, or infer schemas.

## Discord Behavior

Wave 1 Discord adapter can only:

- build intents;
- create a DiscordGo session;
- optionally open the gateway if `MHCAT_DISCORD_ENABLE_GATEWAY=true`;
- close the session.

Default intents are `Guilds` only.

## Intentionally Not Implemented

- Slash command features.
- Prefix command features.
- Command registration.
- Component/modal routing.
- Legacy custom ID decoders.
- Mongo repositories.
- Mongo index bootstrap.
- Data repair/backfill.
- Scheduler/jobs.
- Feature parity handlers.

## Test Results

Commands run from `MHCAT-REFACTOR`:

```txt
go fmt ./...
go test ./...
go vet ./...
go build ./cmd/mhcat-bot
go run ./cmd/mhcat-bot
```

Results:

- `go fmt ./...`: passed.
- `go test ./...`: passed; Mongo integration test skipped by default.
- `go vet ./...`: passed.
- `go build ./cmd/mhcat-bot`: passed.
- `make check`: passed.
- clean no-env `go run ./cmd/mhcat-bot`: exited non-zero with missing config error, no panic, no secret output.

Sandbox note: verification used `GOCACHE` and `GOMODCACHE` under `/private/tmp` because the managed filesystem sandbox does not allow writes to the default Go cache under the user home directory.

## Known Limitations

- No live Mongo audit has run yet.
- No real Discord gateway smoke test is included.
- `go run ./cmd/mhcat-bot` with valid config requires a reachable MongoDB because Wave 1 performs connect and ping.
- With gateway disabled, the app exits after the Mongo ping health check and cleanup.

## Next Recommended Step

Wave 2 should add command registry modeling, command sync dry-run/apply mechanics, and interaction responder/router interfaces without registering commands from shard startup.
