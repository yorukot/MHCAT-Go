# Scheduler Lease Foundation

Status: implemented as shared infrastructure. The `mhcat-work-payout` one-shot apply path and the recurring automatic-notification delivery worker use it.

## Purpose

Legacy MHCAT runs scheduler work inside the bot process with shard-0 checks. That is not enough for a Go rollout with multiple processes or deployment replicas. The lease foundation gives recurring jobs a single-owner primitive before they mutate MongoDB or send Discord messages.

## Implemented

- `internal/core/domain.SchedulerLease`
- `internal/core/ports.SchedulerLeaseStore`
- `internal/adapters/mongo.SchedulerLeaseStore`
- `internal/testutil/fakemongo.SchedulerLeaseStore`
- `cmd/mhcat-scheduler-lease`
- `internal/core/services/notifications.DeliveryWorker`, using lease name `auto-notification-delivery`

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

- `cmd/mhcat-bot` uses the lease store only for automatic-notification delivery when its delivery, Gateway, and lease gates are all enabled.
- `cmd/mhcat-work-payout --apply` uses the lease to prevent multiple Go operators from owning the payout run at the same time.
- The automatic-notification worker schedules and sends only while it holds `auto-notification-delivery`, reconciles at most every 30 seconds, and releases the lease during graceful shutdown.
- The lease does not coordinate with legacy Node, so `handler/cron.js` must be disabled before Go delivery starts.
- No command sync, Mongo index creation, or feature repair is performed by lease infrastructure.
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

## Current And Next Consumers

Current consumer:

- persisted automatic notifications from `MHCAT/handler/cron.js`, restored in Go behind `MHCAT_FEATURE_AUTO_NOTIFICATION_DELIVERY_ENABLED=true`.

Future scheduler/job slices should use this lease before writes:

- recurring economy daily reset if it moves from one-shot CLI to background job;
- recurring work payout if it moves from one-shot CLI to a background job;
- birthday/lottery scheduler decisions if those inactive legacy paths are restored by ADR.

## Open Decisions

- Lease owner naming for production processes.
- Whether future recurring jobs run in `cmd/mhcat-bot` or a dedicated worker binary.
- Job-specific lease TTL and renewal cadence beyond automatic-notification delivery.
- Metrics/logging for lease contention and missed ticks.
