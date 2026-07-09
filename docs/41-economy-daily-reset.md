# Economy Daily Reset One-Shot Tool

Status: implemented as a dry-run-first operational command, not a recurring bot-startup scheduler.

## Legacy Behavior

Legacy daily reset is in `MHCAT/handler/cron.js`.

- Runs at `00:00 Asia/Taipei`.
- Schedules only when `client.shard && client.shard.ids[0] === 0`.
- Reads `gift_change.distinct('guild', { time: { $ne: 0 } })`.
- Resets `coin.today` to `0` for guilds outside that excluded guild list.
- Reads all `work_set` rows.
- For each guild, increments `work_user.energi` by `work_set.get_energy` where below `work_set.max_energy`.
- Clamps `work_user.energi` back to `work_set.max_energy` where above the maximum.

## Go Implementation

The Go refactor adds:

- `cmd/mhcat-economy-reset`
- `internal/adapters/mongo/repositories.DailyResetRepository`
- `internal/core/ports.DailyResetRepository`
- `internal/core/domain.DailyResetResult`

Default command behavior is dry-run:

```bash
go run ./cmd/mhcat-economy-reset --dry-run
```

Apply requires both:

```bash
MHCAT_JOBS_DAILY_RESET_ENABLED=true
go run ./cmd/mhcat-economy-reset --apply
```

`--dry-run=false` is rejected. The write path can only be entered through `--apply`.

## Intentional Compatibility Fix

The Go reset does not copy the legacy raw `time: {$ne: 0}` query directly. It reads `gift_changes` and normalizes `time` through the same BSON conversion strategy used by the sign-in repository. A guild is excluded from midnight reset only when normalized `gift_changes.time != 0`.

Reason: legacy Mongo data can have missing/null/loose typed `gift_changes.time`, and the raw `$ne: 0` query can exclude guilds that Go sign-in treats as daily-reset mode.

## Safety

- The tool does not run from `cmd/mhcat-bot`.
- The tool does not sync Discord commands.
- The tool does not create Mongo indexes.
- The tool does not repair or backfill documents.
- Dry-run uses counts and no writes.
- Apply uses the legacy-compatible write paths only after the explicit env gate and `--apply`.

## Production Preconditions

Before production apply:

- Run `cmd/mhcat-mongo-audit`.
- Confirm no duplicate logical-key risks block planned unique indexes for `coins` and `sign_lists`.
- Review duplicate/mixed-type risks for `gift_changes`.
- Confirm Node.js bot and Go tool are not both trying to own the reset at the same time.
- Capture backup/restore point or at least an operational rollback plan.

## Remaining Work

- Lease-backed recurring scheduler.
- Scheduler ownership runbook for multi-shard/multi-process rollout.
- Recurring work completion payout scheduling. The one-shot `mhcat-work-payout` operator command exists separately.
- Optional stronger dry-run estimate for work-energy rows that are below max but will exceed max after increment.
