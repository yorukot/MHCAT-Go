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

The Go payout does not copy the legacy read-modify-write coin update. It uses one Mongo aggregation-pipeline update that conditionally increments `coin` and writes the job marker atomically, so concurrent balance changes are not overwritten by a stale read and a crash retry cannot credit the same job twice.

The Go payout also fixes the legacy `gift_change.time == 0` new-balance bug. A normalized reset marker of `0` means daily-reset mode, so a newly created coin document gets `today = 1`. Non-zero reset markers still use `today = now_seconds` like the legacy rolling-cooldown path.

Due payout rows must have Mongo's guaranteed `_id`, non-empty `guild`, non-empty `user`, non-idle `state`, and positive `end_time`. `get_coin` is applied as stored, including zero or negative values, because the legacy code did not block those values. Operators should audit impossible rewards before apply.

## Lease Safety

Apply mode acquires `mhcat_scheduler_locks` with the configured lease name before any payout writes. If another owner holds the lease, the command prints a skip report and exits with code `2`.

Dry-run does not acquire a lease and performs no writes.

The lease prevents two Go operator processes from intentionally running the payout at the same time. Job idempotency is independent of the lease and is enforced by the coin marker described below.

## Crash Idempotency

Each valid due row gets two deterministic values:

- a marker key derived from `work_users._id`;
- a job token derived from `_id`, `guild`, `user`, `state`, `end_time`, and `get_coin`.

The coin update stores the latest token and `end_time` under `coins.mhcat_work_payouts.<marker-key>` in the same atomic update that increments `coin`. A retry with the same token preserves the balance and then completes the state reset. A delayed attempt with an older `end_time`, or a different token at the same `end_time`, is rejected instead of overwriting a newer marker.

Only the latest marker for each distinct `work_users._id` is retained, so repeated jobs on the normal single work row do not append unbounded history. Deleting and recreating work rows can leave historical marker keys; no automatic cleanup is performed.

Existing coin rows are targeted by their stable `_id`. A missing balance is created with a deterministic BSON ObjectID so a crash retry resolves the same row. If more than one `coins` row exists for `{guild,member}`, the command returns `ErrWorkPayoutCoinConflict` before crediting that job. Duplicate `work_users` rows remain independently payable because each has a distinct `_id` and marker key.

The final state update matches the exact `_id`, guild, user, state, end time, and reward snapshot. If that snapshot changed after the credit, the command returns a state-conflict error and does not reset a newer job. The already-written marker makes retrying the original snapshot balance-safe.

This protection applies to Go payout attempts. Legacy Node does not read or write the marker, so Node and Go ownership must still be exclusive.

The command also acquires a lease once and does not renew it. Apply validation requires the lease TTL to be greater than the command timeout, but very large backlogs should still be handled in bounded operational batches before this becomes a recurring scheduler.

## Safety

- The tool does not run from `cmd/mhcat-bot`.
- The tool does not sync Discord commands.
- The tool does not create Mongo indexes.
- The tool does not repair or backfill documents.
- The marker field is added lazily only to coin rows that receive a Go work payout.
- The tool sends no Discord messages.
- Dry-run counts eligible jobs and performs no writes.
- Apply requires the work-payout gate, scheduler-lease gate, scheduler owner, and `--apply`.

## Production Preconditions

Before production apply:

- Run `cmd/mhcat-mongo-audit`.
- Confirm duplicate risks for `coins.guild/member`, `work_users.guild/user`, and `gift_changes.guild`.
- Audit due rows for mixed/non-numeric `coin`, `end_time`, and `get_coin`, and for zero/negative rewards.
- Confirm Node.js bot and Go tool are not both trying to own the payout loop.
- Confirm the additive marker field and rollback behavior below are acceptable to all `coins` consumers.
- Capture backup/restore point or at least an operational rollback plan.
- Run against staging data first and compare sample balances/state changes manually.

## Remaining Work

- Recurring scheduler or worker process.
- Production runbook for Node-to-Go job ownership handoff.
- Optional unique indexes after duplicate audits.

## Rollback

1. Stop or disable every Go work-payout owner and confirm the configured lease is released or expired.
2. Restore the Node minute-loop owner only after Go can no longer write payouts.
3. Leave `coins.mhcat_work_payouts` in place. Legacy Mongoose and dashboard readers ignore the additive field, while preserving it allows a later Go rollout to recognize prior jobs.
4. Do not unset markers while a due work row may already have been credited but not reset. Removing that marker can make the next Go retry credit it again.

No backfill is required. Marker removal is optional only after all Go payout paths are retired and an audit confirms there are no in-flight due rows.

## Verification

The opt-in Mongo integration tests cover:

- a crash after atomic coin commit but before state reset, followed by a no-double-credit retry;
- concurrent same-token contenders crediting the balance exactly once;
- a newer job replacing the retained marker and rejecting an older delayed attempt;
- deterministic missing-balance creation;
- duplicate coin rejection before writes;
- independent payout markers for duplicate work rows.
