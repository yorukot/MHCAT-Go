# Auto-Notification Config Slice

Status: implemented for `automatic-notification`, `/自動通知列表`, and `/自動通知刪除`. Disabled by default.

## Scope

This slice exposes the legacy setup/list/delete maintenance commands for automatic notification schedules:

- `automatic-notification`
- `/自動通知列表`
- `/自動通知刪除`

The setup path creates a pending `cron_sets` row, opens the legacy `cron_set*` modal, accepts direct cron expressions, writes the rollback-compatible message payload, and sends a best-effort preview message to the interaction channel.

It does not implement:

- `week_menu`, `hour_menu`, or `min_menu`
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

`automatic-notification` inserts a pending row with `guild`, `channel`, `id`, `cron:null`, and `message:null`, then completes that row after a valid direct cron modal submit by setting `cron` and `message`. The message BSON keeps the legacy scheduler shape: plain messages are `{content: ...}`, and embeds are `{content: <string-or-null>, embeds:[{data:{...}}]}`.

The direct-cron validator is intentionally conservative for this slice: it accepts five-field cron strings containing numbers, `*`, `/`, `,`, and `-`, rejects an every-minute `*` minute field, and rejects minute steps below 15. The legacy simplified `cancel`/select-menu flow is still pending.

`/自動通知列表` reads rows by `guild`, renders active rows with non-null/non-missing `cron`, and cleans abandoned setup drafts whose `cron` is null or missing.

`/自動通知刪除` deletes one row by `{guild,id}` to preserve the legacy `findOne(...).delete()` behavior. It does not create indexes or modify scheduler state.

## Staging Smoke

1. Use an isolated staging guild and database.
2. Enable both flags.
3. Run `go run ./cmd/mhcat-staging-preflight --format text`.
4. Run `scripts/staging/command-sync-dry-run.sh` and confirm the plan includes `automatic-notification`, `自動通知列表`, and `自動通知刪除` in addition to previously approved commands.
5. Apply only after dry-run review.
6. Run `/automatic-notification channel:<test channel>`, submit the modal with a safe direct cron such as `*/30 * * * *`, and confirm the setup completion message includes the generated id.
7. Confirm the staging `cron_sets` row changed from pending to active and contains the expected `cron` and rollback-compatible `message` shape.
8. Confirm a preview message was sent to the interaction channel.
9. Optionally create a disposable pending draft with null or missing `cron`.
10. Run `/自動通知列表` and confirm active rows render and pending drafts do not.
11. Run `/自動通知刪除 id:<active id>` and confirm the legacy green delete embed appears.
12. Run `/自動通知刪除 id:<missing id>` and confirm the legacy red missing-id tutorial embed appears.

No staging step should start a recurring scheduler, depend on Message Content intent, or write usage counters. Preview sends are part of setup smoke; recurring notification sends remain disabled.
