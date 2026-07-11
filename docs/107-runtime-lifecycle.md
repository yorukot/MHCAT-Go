# Runtime Lifecycle Contract

Status: parity/operations-audited against `shard.js`, `index.js`, active legacy handler imports, current config validation, app/session/Mongo lifecycle, interaction/event dispatchers, recurring workers, signal handling, failure cleanup, smoke mode, and shutdown tests. Live Gateway smoke and multi-process production topology remain operator-gated.

## Startup Order

Legacy connects Mongo before loading handlers and logging into Discord. Go preserves that dependency order: validate config, create/connect Mongo, create the Discord session, build interaction and event dispatchers, register handlers/workers, and only then open Gateway. Any factory, dispatcher, or Gateway-open failure unregisters handlers, shuts down created workers, closes Discord, and disconnects Mongo before returning an error.

Gateway is disabled by default. With Gateway off, the app can validate/build non-live dependencies and shuts down without opening Discord. Gateway smoke is separately allowed only in staging, waits for the real ready signal under a bounded timeout, logs readiness without secrets, and exits through normal shutdown.

Bot startup never creates, updates, or deletes Discord application commands and never creates Mongo indexes. Command registration is owned by the explicit dry-run-first `mhcat-command-sync` CLI; indexes are owned by `mhcat-mongo-index`. This intentionally removes legacy per-shard `ready` races.

## Session And Shard Scope

The currently supported bot deployment is one DiscordGo session with shard ID `0` and count `1`. Guild interactions/events, status metrics, and worker lifecycle are correct in that scope. Legacy automatic process sharding/respawn is not reproduced inside the bot binary. Raising shard count requires an explicit session manager, cross-process component-state policy, aggregated bot-info metrics, and verified scheduler ownership.

Global recurring jobs do not rely on shard zero. Auto-notification, daily reset, and work payout use independent feature/write gates and Mongo leases; local overlap is suppressed where documented. Node jobs and Go jobs still require exclusive ownership because Node does not honor Go leases/markers.

## Shutdown And Restart

`mhcat-bot` derives a cancellation context from termination signals. Shutdown is idempotent and unregisters interaction/Gateway handlers, drains interaction and event dispatcher shutdown hooks, stops timers/workers, closes Discord, and disconnects Mongo. Startup and shutdown errors are joined and returned without credential leakage.

Go intentionally does not preserve the legacy hardcoded-admin MessageCreate restart backdoor or exit webhook. Restarts are owned by the process supervisor/orchestrator: send a normal termination signal, wait for bounded graceful shutdown, then start the replacement process. No privileged Discord user IDs or webhook credentials are embedded in runtime code.

## Verification

```bash
go test ./internal/app ./internal/adapters/discordgo ./internal/config ./cmd/mhcat-bot
go test -race ./internal/app ./internal/adapters/discordgo
go vet ./...
make check
```

Coverage includes Mongo/Discord factory failures, Gateway timeout, ready smoke, no-Gateway run, handler registration/removal, event and interaction shutdown hooks, idempotent shutdown, worker stop behavior, and signal-context cancellation.

## Staging And Rollback

1. Keep every feature/write gate off initially. Run preflight and command-sync dry-run without printing environment values.
2. Start one Go process/session against isolated staging Mongo and a staging guild. Verify ready, one interaction, representative enabled events, graceful `SIGTERM`, and clean restart under the intended supervisor.
3. Enable each recurring worker separately with unique lease ownership and stop its Node equivalent first.
4. Observe Discord reconnect/rate-limit logs, Mongo disconnects, worker lease release, and process exit timing without exposing secrets.

Rollback by disabling Go runtime/worker gates, gracefully stopping Go, and only then restoring reviewed Node owners. Open process-local component state may expire during rollback as documented by each feature contract. Do not overlap command/event/job writers or use Discord messages as a privileged restart channel.

Production remains gated on live Gateway smoke, supervisor restart rehearsal, confirmed single-session capacity, exclusive worker ownership, and an explicit architecture change before any multi-session shard rollout.
