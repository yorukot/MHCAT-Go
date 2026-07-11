# Scheduled Economy Jobs Parity Contract

Status: parity-audited against the active economy blocks in legacy `handler/cron.js` and `handler/gift.js`, and current Go one-shot tools, recurring workers, repositories, scheduler lease store, app lifecycle, configuration gates, fake-time tests, guarded Mongo harnesses, and race coverage. Live two-replica timing smoke remains required before production rollout.

## Scope

This contract covers the `00:00 Asia/Taipei` coin daily reset and work-energy refill, plus the every-minute completed-work payout. Detailed behavior and rollback contracts remain in [41-economy-daily-reset.md](41-economy-daily-reset.md), [42-scheduler-lease-foundation.md](42-scheduler-lease-foundation.md), [43-work-payout.md](43-work-payout.md), and [105-work-system.md](105-work-system.md).

## Daily Reset

The recurring worker and `mhcat-economy-reset --apply` share lease `daily-reset`. The worker preserves the legacy midnight schedule, normalized `gift_changes.time` exclusion, `coins.today=0` reset, natural-order iteration of every `work_sets` row, repeated increments caused by duplicate work settings, and final max-energy clamp. It does not catch up after downtime or send Discord messages.

Dry-run performs no writes or lease acquisition. Apply and recurring execution require independent write, scheduler, owner, timeout, and lease gates. A contending owner skips without writes.

## Work Payout

The recurring worker and `mhcat-work-payout --apply` share the configured payout lease and preserve the legacy minute schedule, rounded Unix-time comparison, due-row selection, signed and mixed scalar payout arithmetic, balance creation, `today` selection, one-row duplicate coin selection, and reset to `待業中`. They leave work end time, energy, and reward fields unchanged and send no Discord messages.

Go intentionally replaces the legacy stale read-modify-write with an atomic balance increment and deterministic per-work-row marker. This prevents concurrent balance overwrite and duplicate credit after a crash between coin mutation and state reset. The marker is additive and ignored by legacy Mongoose readers; it must remain during rollback while a credited row could still be due.

## Ownership And Lifecycle

Every Go process schedules the same ticks, but only the current lease holder writes. Owners must be unique because same-owner reacquisition is allowed. Workers skip overlapping callbacks in one process, release leases after success/failure and graceful shutdown, and use timeouts shorter than the lease TTL.

Mongo leases cannot coordinate with Node. Stop all Node processes loading the corresponding legacy handlers before enabling Go or running one-shot apply. No path syncs commands, creates indexes, repairs documents, or performs automatic compensation for a partially completed daily reset.

## Migration And Rollback

Before production ownership transfer, audit and back up `coins`, `gift_changes`, `work_sets`, and `work_users`, including duplicate logical keys and malformed/non-finite scalars. Rehearse dry-run, one-shot apply, held-lease contention, failure/retry, and two-replica ticks against a disposable database.

Rollback disables and stops every Go owner, confirms lease release or expiry, then restores one Node owner. Do not automatically reverse successful reset values, remove payout markers, merge duplicates, or infer work state from balances; restore reviewed rows from the captured backup when data rollback is required.

## Verification

```bash
go test ./internal/core/services/economy ./internal/adapters/mongo/repositories \
  ./internal/adapters/mongo ./internal/config ./internal/app \
  ./cmd/mhcat-economy-reset ./cmd/mhcat-work-payout ./cmd/mhcat-scheduler-lease
go test -race ./internal/core/services/economy ./internal/adapters/mongo/repositories \
  ./internal/adapters/mongo ./internal/app
go vet ./...
```

Guarded Mongo integration tests require a generated disposable database and explicit integration flags. Never point them at production.
