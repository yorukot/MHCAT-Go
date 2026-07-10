# Economy Daily Reset

Status: dry-run/apply tooling and the lease-backed `00:00 Asia/Taipei` bot worker are implemented behind disabled-by-default gates.

## Legacy Behavior

Legacy daily reset is in `MHCAT/handler/cron.js`.

- Runs at `00:00 Asia/Taipei`.
- Schedules only when `client.shard && client.shard.ids[0] === 0`.
- Reads `gift_change.distinct('guild', { time: { $ne: 0 } })`.
- Resets `coin.today` to `0` for guilds outside that excluded guild list.
- Reads all `work_set` rows.
- For each row, increments `work_user.energi` by `get_energy` where below `max_energy`.
- Clamps `work_user.energi` back to `max_energy` where above the maximum.

## Go Implementation

The Go paths share `internal/adapters/mongo/repositories.DailyResetRepository` and lease name `daily-reset`:

- `cmd/mhcat-economy-reset` for operator dry-run/apply;
- `internal/core/services/economy.DailyResetWorker` for recurring bot execution;
- `internal/core/ports.DailyResetRepository`;
- `internal/core/domain.DailyResetResult`.

The scheduler uses five-field cron `0 0 * * *` in fixed UTC+8 named `Asia/Taipei`. Each Go replica has the cron entry, but a replica acquires `daily-reset` only when the tick fires. Exactly one lease holder runs the writes; contenders skip. The lease is released after success, repository failure, or canceled shutdown.

There is no catch-up run after downtime. This matches the legacy cron process: a bot that is down at midnight waits until the next scheduled midnight.

## One-Shot Tool

Dry-run performs no writes and does not acquire a lease:

```bash
go run ./cmd/mhcat-economy-reset --dry-run
```

Apply requires the write and lease gates, a unique owner, and an explicit flag:

```bash
MHCAT_JOBS_DAILY_RESET_ENABLED=true \
MHCAT_SCHEDULER_LEASE_ENABLED=true \
MHCAT_SCHEDULER_LEASE_OWNER=staging-reset-cli \
MHCAT_SCHEDULER_LEASE_TTL=2m \
MHCAT_SCHEDULER_LEASE_TIMEOUT=10s \
MHCAT_JOBS_DAILY_RESET_TIMEOUT=60s \
go run ./cmd/mhcat-economy-reset --apply
```

Apply acquires `daily-reset`, exits with code `2` without writes when another owner holds it, and releases after the run. `--dry-run=false` remains rejected.

## Recurring Worker

Recurring execution requires all of these:

```bash
MHCAT_FEATURE_DAILY_RESET_SCHEDULER_ENABLED=true
MHCAT_JOBS_DAILY_RESET_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_SCHEDULER_LEASE_ENABLED=true
MHCAT_SCHEDULER_LEASE_OWNER=staging-reset-bot-a
MHCAT_SCHEDULER_LEASE_TTL=2m
MHCAT_SCHEDULER_LEASE_TIMEOUT=10s
MHCAT_JOBS_DAILY_RESET_TIMEOUT=60s
```

The lease TTL must be greater than the reset timeout plus the lease-operation timeout. The existing `MHCAT_JOBS_DAILY_RESET_DRY_RUN` setting controls the CLI default only; the explicit scheduler feature gate always means the midnight worker can write.

Gateway is required because the worker currently shares the bot event-runtime lifecycle, although the reset itself performs no Discord API call and needs no privileged intent.

## Compatibility Fix

Go does not copy the raw legacy `time: {$ne: 0}` query. It reads `gift_changes` and normalizes `time` through the same BSON conversion used by sign-in. A guild is excluded from midnight reset only when normalized `gift_changes.time != 0`.

This keeps missing, null, and loose-typed legacy values consistent with Go sign-in daily-reset mode.

## Safety Boundaries

- Stop every Node process that loads `handler/cron.js` before enabling Go scheduling or running apply.
- Give every Go process and CLI invocation a unique `MHCAT_SCHEDULER_LEASE_OWNER`; do not reuse a bot replica's owner for manual apply because same-owner reacquisition is allowed by the lease primitive.
- Do not run manual apply while investigating a partial midnight failure without checking which work guilds were already incremented.
- Repeating the coin reset is harmless, but repeating work-energy refill can add energy again up to `max_energy`.
- No path creates indexes, repairs/backfills documents, syncs commands, or sends Discord messages.
- Duplicate `work_sets` rows retain legacy repeated-increment behavior; audit them before production.

## Staging Smoke

1. Use an isolated database with disposable `coins`, `gift_changes`, `work_sets`, and `work_users` rows.
2. Stop Node `handler/cron.js` and record before-values for `coins.today` and `work_users.energi`.
3. Run CLI dry-run and review all reported counts.
4. Run one-shot apply with lease gates; verify the expected writes, `lease_acquired=true`, and a released `daily-reset` lease.
5. Hold `daily-reset` with another owner and verify apply exits `2` with no writes.
6. Restore fresh fixtures, enable the recurring worker on two Go replicas with distinct owners, and keep both running across `00:00 Asia/Taipei`.
7. Verify one replica logs completion, the other logs lease contention, and energy changes exactly once.
8. Gracefully stop both replicas and confirm no lease remains held.

## Rollback

1. Set `MHCAT_FEATURE_DAILY_RESET_SCHEDULER_ENABLED=false` and stop Go.
2. Confirm `daily-reset` is released or expired.
3. Restore Node cron ownership only after Go can no longer run.
4. Do not reverse successful reset values automatically; use the captured staging/production audit and backup if a data rollback is required.

## Remaining Work

- Recurring completed-work payout now has crash-safe repository idempotency but still needs separately gated worker wiring, lease lifecycle tests, and Node-to-Go ownership smoke.
- Optional stronger dry-run estimation for energy rows that are below max but will exceed max after increment.
