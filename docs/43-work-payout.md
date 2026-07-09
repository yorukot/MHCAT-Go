# Work Payout One-Shot Tool

Status: implemented as a dry-run-first operational command, not a recurring bot-startup scheduler.

## Legacy Behavior

Legacy work payout is in `MHCAT/handler/gift.js`.

- Runs every minute.
- Schedules only when `client.shard && client.shard.ids[0] === 0`.
- Reads `work_user` rows where `state != "待業中"` and `end_time <= round(now_seconds)`, then the loop effectively pays only rows where `end_time < round(now_seconds)`.
- Adds `work_user.get_coin` to the matching `coin` document by reading the old balance and writing `coin + get_coin`.
- Creates a `coin` document when no balance exists.
- Sets `coin.today` for new balance documents from `gift_change.time`:
  - no `gift_change` row: `today = 1`;
  - non-zero `gift_change.time`: `today = now_seconds`;
  - `gift_change.time == 0`: legacy JavaScript truthiness bug sets `today = now_seconds`.
- Resets `work_user.state` to `"待業中"`.
- Leaves `work_user.end_time`, `work_user.energi`, and `work_user.get_coin` unchanged.
- Sends no Discord message.

## Go Implementation

The Go refactor adds:

- `cmd/mhcat-work-payout`
- `internal/adapters/mongo/repositories.WorkPayoutRepository`
- `internal/core/ports.WorkPayoutRepository`
- `internal/core/domain.WorkPayoutResult`

Default command behavior is dry-run:

```bash
go run ./cmd/mhcat-work-payout --dry-run
```

Apply requires all of:

```bash
MHCAT_JOBS_WORK_PAYOUT_ENABLED=true
MHCAT_SCHEDULER_LEASE_ENABLED=true
MHCAT_SCHEDULER_LEASE_OWNER=staging-worker
go run ./cmd/mhcat-work-payout --apply
```

`--dry-run=false` is rejected. The write path can only be entered through `--apply`.

## Intentional Compatibility Fixes

The Go payout does not copy the legacy read-modify-write coin update. It uses Mongo `$inc` with `$setOnInsert` so concurrent balance changes are not overwritten by a stale read.

The Go payout also fixes the legacy `gift_change.time == 0` new-balance bug. A normalized reset marker of `0` means daily-reset mode, so a newly created coin document gets `today = 1`. Non-zero reset markers still use `today = now_seconds` like the legacy rolling-cooldown path.

Due payout rows must have non-empty `guild`, non-empty `user`, non-idle `state`, and positive `end_time`. `get_coin` is applied as stored, including zero or negative values, because the legacy code did not block those values. Operators should audit impossible rewards before apply.

## Lease Safety

Apply mode acquires `mhcat_scheduler_locks` with the configured lease name before any payout writes. If another owner holds the lease, the command prints a skip report and exits with code `2`.

Dry-run does not acquire a lease and performs no writes.

The lease prevents two Go operator processes from intentionally running the payout at the same time. It is not a complete job idempotency system.

## Known Limitation

The current write sequence increments coins before resetting the `work_users.state`. If the process crashes after coin increment and before state reset, rerunning may pay that row again. This matches the legacy ordering risk in spirit and is documented instead of hidden.

If the state reset matches no document after a coin increment, the command fails with a state-conflict error instead of reporting success. That can indicate Node.js still owns the loop, another operator is mutating the same rows, or legacy duplicate/stale rows need audit. The coin increment may already have happened, so do not retry blindly without checking the affected row.

Future recurring scheduler work should consider either a transaction, a per-job payout marker, or a state transition before payout with a recovery path. That would be a schema/behavior change requiring ADR, audit, rollback guidance, and tests.

The command also acquires a lease once and does not renew it. Apply validation requires the lease TTL to be greater than the command timeout, but very large backlogs should still be handled in bounded operational batches before this becomes a recurring scheduler.

## Safety

- The tool does not run from `cmd/mhcat-bot`.
- The tool does not sync Discord commands.
- The tool does not create Mongo indexes.
- The tool does not repair or backfill documents.
- The tool sends no Discord messages.
- Dry-run counts eligible jobs and performs no writes.
- Apply requires the work-payout gate, scheduler-lease gate, scheduler owner, and `--apply`.

## Production Preconditions

Before production apply:

- Run `cmd/mhcat-mongo-audit`.
- Confirm duplicate risks for `coins.guild/member`, `work_users.guild/user`, and `gift_changes.guild`.
- Audit due rows for mixed/non-numeric `coin`, `end_time`, and `get_coin`, and for zero/negative rewards.
- Confirm Node.js bot and Go tool are not both trying to own the payout loop.
- Review the crash-idempotency limitation above.
- Capture backup/restore point or at least an operational rollback plan.
- Run against staging data first and compare sample balances/state changes manually.

## Remaining Work

- Recurring scheduler or worker process.
- Job-level idempotency beyond lease ownership.
- Production runbook for Node-to-Go job ownership handoff.
- Optional unique indexes after duplicate audits.
