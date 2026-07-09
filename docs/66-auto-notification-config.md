# Auto-Notification Config Slice

Status: implemented for `/自動通知列表` and `/自動通知刪除` only. Disabled by default.

## Scope

This slice exposes the legacy list/delete maintenance commands for existing automatic notification schedules:

- `/自動通知列表`
- `/自動通知刪除`

It does not implement:

- `automatic-notification`
- the `cron_set*` setup modal
- `week_menu`, `hour_menu`, or `min_menu`
- message payload creation
- recurring scheduler ownership
- automatic notification sends

## Flags

Runtime:

```bash
MHCAT_FEATURE_AUTO_NOTIFICATION_CONFIG_ENABLED=true
```

Command sync:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_AUTO_NOTIFICATION_CONFIG=true
```

Set both together in staging. `mhcat-staging-preflight` and the staging command-sync scripts reject command sync when the runtime flag is not enabled.

## Mongo Behavior

Collection: `cron_sets`

Legacy fields used:

- `guild`
- `id`
- `cron`
- `channel`
- `message`

`/自動通知列表` reads rows by `guild`, renders active rows with non-null/non-missing `cron`, and cleans abandoned setup drafts whose `cron` is null or missing.

`/自動通知刪除` deletes one row by `{guild,id}` to preserve the legacy `findOne(...).delete()` behavior. It does not create indexes or modify scheduler state.

## Staging Smoke

1. Use an isolated staging guild and database.
2. Enable both flags.
3. Run `go run ./cmd/mhcat-staging-preflight --format text`.
4. Run `scripts/staging/command-sync-dry-run.sh` and confirm the plan includes only `自動通知列表` and `自動通知刪除` in addition to previously approved commands.
5. Apply only after dry-run review.
6. Create or confirm a safe staging `cron_sets` active row.
7. Optionally create a disposable pending draft with null or missing `cron`.
8. Run `/自動通知列表` and confirm active rows render and pending drafts do not.
9. Run `/自動通知刪除 id:<active id>` and confirm the legacy green delete embed appears.
10. Run `/自動通知刪除 id:<missing id>` and confirm the legacy red missing-id tutorial embed appears.

No staging step should open the setup modal, start a scheduler, send notification messages, enable Message Content intent, or write usage counters.
