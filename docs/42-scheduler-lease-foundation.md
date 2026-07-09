# Scheduler Lease Foundation

Status: implemented as shared infrastructure. The `mhcat-work-payout` one-shot apply path uses it; no recurring job is wired into `cmd/mhcat-bot`.

## Purpose

Legacy MHCAT runs scheduler work inside the bot process with shard-0 checks. That is not enough for a Go rollout with multiple processes or deployment replicas. The lease foundation gives future recurring jobs a single-owner primitive before they mutate MongoDB or send Discord messages.

## Implemented

- `internal/core/domain.SchedulerLease`
- `internal/core/ports.SchedulerLeaseStore`
- `internal/adapters/mongo.SchedulerLeaseStore`
- `internal/testutil/fakemongo.SchedulerLeaseStore`
- `cmd/mhcat-scheduler-lease`

The Mongo collection is:

```txt
mhcat_scheduler_locks
```

The lock name is stored as `_id`, so the default Mongo `_id` uniqueness is the lock identity. No new index is required for the first implementation and no index is created automatically.

`expires_at`, `created_at`, and `updated_at` are UTC instants. Job schedule timezones, such as legacy `Asia/Taipei`, must be handled by the job scheduler layer rather than the lease store.

## Lease Semantics

Acquire:

- Acquires a missing lease by inserting a new document.
- Re-acquires an expired lease.
- Lets the same owner re-acquire and increments the fencing token.
- Returns `Acquired=false` when another owner still holds an unexpired lease.

Renew:

- Requires name, owner, fence token, and an unexpired lease.
- Extends `expires_at`.
- Fails with `ErrSchedulerLeaseNotHeld` if ownership/fence/expiry no longer match.

Release:

- Requires name, owner, and fence token.
- Marks only the matching held lease as expired and clears its owner.
- Preserves the document so the next acquire can increment `fence` monotonically.
- Fails with `ErrSchedulerLeaseNotHeld` if the caller no longer owns it.

## Safety Boundaries

- The lease store is not used by `cmd/mhcat-bot` yet.
- `cmd/mhcat-work-payout --apply` uses the lease to prevent multiple Go operators from owning the payout run at the same time.
- No recurring scheduler is started.
- No command sync, Mongo index creation, feature repair, or feature writes are introduced by this slice.
- The diagnostic CLI defaults to read-only `status`.
- CLI write actions require `MHCAT_SCHEDULER_LEASE_ENABLED=true` and `--apply`.
- `internal/core/**` remains driver-agnostic.

## Diagnostic CLI

Read-only status:

```bash
go run ./cmd/mhcat-scheduler-lease --name daily-reset --action status
```

Explicit write example:

```bash
MHCAT_SCHEDULER_LEASE_ENABLED=true \
MHCAT_SCHEDULER_LEASE_OWNER=staging-worker \
go run ./cmd/mhcat-scheduler-lease --name daily-reset --action acquire --apply
```

Renew/release require the current `fence` token:

```bash
go run ./cmd/mhcat-scheduler-lease --name daily-reset --action renew --fence 3 --apply
go run ./cmd/mhcat-scheduler-lease --name daily-reset --action release --fence 3 --apply
```

## Next Consumers

Future scheduler/job slices should use this lease before writes:

- recurring economy daily reset if it moves from one-shot CLI to background job;
- persisted automatic notifications from `MHCAT/handler/cron.js`;
- birthday/lottery scheduler decisions if those inactive legacy paths are restored by ADR.

## Open Decisions

- Lease owner naming for production processes.
- Whether recurring jobs run in `cmd/mhcat-bot` or a dedicated worker binary.
- Lease TTL and renewal cadence per job.
- Metrics/logging for lease contention and missed ticks.
