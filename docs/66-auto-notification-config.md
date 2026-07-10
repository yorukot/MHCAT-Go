# Auto-Notification Config Slice

Status: implemented for `automatic-notification`, `/自動通知列表`, and `/自動通知刪除`. Disabled by default.

## Scope

This slice exposes the legacy setup/list/delete maintenance commands for automatic notification schedules:

- `automatic-notification`
- `/自動通知列表`
- `/自動通知刪除`

The setup path creates a pending `cron_sets` row, opens the legacy `cron_set*` modal, accepts direct cron expressions or the legacy-style weekday/hour/minute wizard, writes the rollback-compatible message payload, and sends a best-effort preview message to the interaction channel.

It does not implement recurring scheduler ownership or automatic notification sends.

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

`automatic-notification` inserts a pending row with `guild`, `channel`, `id`, `cron:null`, and `message:null`, then completes that row after a valid direct cron submit or the final simplified minute selection by setting `cron` and `message`. The simplified flow writes `<minutes> <hours> * * <weekdays>`, exactly matching the legacy shape. The message BSON keeps the legacy scheduler shape: plain messages are `{content: ...}`, and embeds are `{content: <string-or-null>, embeds:[{data:{...}}]}`.

Direct cron expressions are parsed with `robfig/cron`. The next two occurrences must be at least 15 minutes apart, matching the legacy setup check, and weekday `7` is accepted as Sunday. Any syntactically invalid value, including `cancel` or `取消`, opens the simplified flow.

The simplified flow preserves the legacy visible weekday, 24-hour, and five-minute options, multi-select limits, random-color embeds, completion text, and five-minute lifetime. Go-generated controls use owner-scoped versioned IDs (`mhcat:v1:cron:<week|hour|minute>:state=<id>`) because the legacy generic IDs collide with the birthday flow. Wizard message content remains process-local and is written to Mongo only after the final minute selection. A restart or expiry requires rerunning the command; abandoned pending rows are removed by `/自動通知列表` cleanup.

`/自動通知列表` reads rows by `guild`, renders active rows with non-null/non-missing `cron`, and cleans abandoned setup drafts whose `cron` is null or missing.

`/自動通知刪除` deletes one row by `{guild,id}` to preserve the legacy `findOne(...).delete()` behavior. It does not create indexes or modify scheduler state.

## Staging Smoke

1. Use an isolated staging guild and database.
2. Enable both flags.
3. Run `go run ./cmd/mhcat-staging-preflight --format text`.
4. Run `scripts/staging/command-sync-dry-run.sh` and confirm the plan includes `automatic-notification`, `自動通知列表`, and `自動通知刪除` in addition to previously approved commands.
5. Apply only after dry-run review.
6. Run `/automatic-notification channel:<test channel>`, submit the modal with a safe direct cron such as `*/30 * * * *`, and confirm the setup completion message includes the generated id.
7. Repeat with `cancel` or `取消`; verify the weekday, hour, and minute selects match the legacy labels/options, complete within five minutes, and produce the expected five-field cron string.
8. Confirm both staging `cron_sets` rows changed from pending to active and contain the expected `cron` and rollback-compatible `message` shape.
9. Confirm each setup sent one preview message to the interaction channel.
10. Verify another user cannot advance the wizard and an expired wizard returns the safe ephemeral retry error.
11. Optionally create a disposable pending draft with null or missing `cron`.
12. Run `/自動通知列表` and confirm active rows render and pending drafts do not.
13. Run `/自動通知刪除 id:<active id>` and confirm the legacy green delete embed appears.
14. Run `/自動通知刪除 id:<missing id>` and confirm the legacy red missing-id tutorial embed appears.

No staging step should start a recurring scheduler, depend on Message Content intent, or write usage counters. Preview sends are part of setup smoke; recurring notification sends remain disabled.
