# Auto-Notification Config And Delivery

Status: setup/list/delete and lease-backed recurring delivery are implemented behind independent disabled-by-default gates.

## Scope

The config path exposes the legacy maintenance commands:

- `automatic-notification`
- `/自動通知列表`
- `/自動通知刪除`

The setup command creates a pending `cron_sets` row, opens the legacy `cron_set*` modal, accepts a direct cron expression or the legacy-style weekday/hour/minute wizard, writes the rollback-compatible message payload, and sends a best-effort preview to the interaction channel.

The delivery path starts no command. It acquires one Mongo lease, reconciles persisted active schedules, and sends their messages at five-field cron times interpreted as fixed UTC+8 named `Asia/Taipei`.

## Flags

Config runtime and command sync:

```bash
MHCAT_FEATURE_AUTO_NOTIFICATION_CONFIG_ENABLED=true
MHCAT_COMMAND_SYNC_INCLUDE_AUTO_NOTIFICATION_CONFIG=true
```

Set those together only when syncing the three config commands. Staging preflight and command-sync scripts reject an unpaired include flag.

Recurring delivery:

```bash
MHCAT_FEATURE_AUTO_NOTIFICATION_DELIVERY_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_SCHEDULER_LEASE_ENABLED=true
MHCAT_SCHEDULER_LEASE_OWNER=staging-auto-notification
MHCAT_SCHEDULER_LEASE_TTL=2m
MHCAT_SCHEDULER_LEASE_TIMEOUT=10s
```

Config and delivery are independent. Delivery does not require command sync, Guild Messages intent, or Message Content intent. Config-only smoke should leave delivery disabled.

## Mongo Behavior

Schedule collection: `cron_sets`

Legacy fields used:

- `guild`
- `id`
- `cron`
- `channel`
- `message`

Lease collection: `mhcat_scheduler_locks`, with `_id` and `lock_name` equal to `auto-notification-delivery`.

`automatic-notification` inserts a pending row with `guild`, `channel`, `id`, `cron:null`, and `message:null`. A valid direct submit or final wizard selection sets `cron` and `message`. The simplified flow writes `<minutes> <hours> * * <weekdays>`. Plain payloads use `{content: ...}`; embed payloads use `{content: <string-or-null>, embeds:[{data:{...}}]}`.

New setup writes Discord embed colors as legacy numeric values. `Random` is resolved once during setup and that selected color is persisted. Existing rows with numeric colors, string colors, or a literal `Random` value remain readable. A color without content, title, or description is still rejected; a color with plain content but no title/description does not create an embed, matching legacy `content || title` behavior.

Direct setup cron is parsed with `robfig/cron`; its next two occurrences must be at least 15 minutes apart, and weekday `7` is accepted as Sunday. Invalid input, including `cancel` or `取消`, opens the simplified wizard. The delivery worker accepts any valid persisted five-field cron, including rows created outside the setup command.

The wizard preserves the legacy options, multi-select limits, completion text, and five-minute lifetime. Go uses owner-scoped versioned IDs (`mhcat:v1:cron:<week|hour|minute>:state=<id>`) because the legacy generic IDs collide with birthday controls. Wizard state is process-local; a restart or expiry requires rerunning setup. `/自動通知列表` filters and removes abandoned rows whose `cron` is null or missing.

`/自動通知刪除` removes one row by `{guild,id}`. The delivery callback reloads that identity before every send, so deleting the row prevents future delivery even before reconciliation removes the in-memory cron entry.

Delivery writes no `cron_sets` fields and creates no indexes.

## Delivery Ownership

The worker:

- acquires and renews only `auto-notification-delivery`;
- schedules nothing while another Go owner holds the lease;
- reconciles at most every 30 seconds, or every one-third of the lease TTL when shorter;
- adds new rows, replaces changed cron expressions, and removes deleted rows;
- removes all cron entries after lease expiry or renewal loss;
- reloads `{guild,id}` immediately before each send;
- skips invalid rows and logs reconciliation/send failures without stopping other valid schedules;
- allows user, role, and everyone mentions like legacy `channel.send`;
- releases the held lease during graceful application shutdown.

The Mongo lease cannot coordinate with the legacy Node process. Stop every Node process that loads `handler/cron.js` before enabling Go delivery, or notifications can be duplicated.

## Config Smoke

1. Use an isolated staging guild, database, and channel.
2. Enable only the config runtime and command-sync flags.
3. Run `go run ./cmd/mhcat-staging-preflight --format text`.
4. Review `scripts/staging/command-sync-dry-run.sh`, then apply only to the staging guild.
5. Run `/automatic-notification channel:<test channel>` with a safe direct cron such as `*/30 * * * *`; verify the generated ID, persisted row, numeric color, and one setup preview.
6. Repeat with `cancel` or `取消`; complete the weekday/hour/minute controls within five minutes and verify the generated cron.
7. Verify cross-user and expired wizard interactions return safe ephemeral errors.
8. Create a disposable pending row, run `/自動通知列表`, and verify the draft is omitted and removed.
9. Verify `/自動通知刪除` success and missing-ID responses.
10. Confirm no recurring send or scheduler lease write occurred.

## Delivery Smoke

1. Stop the Node process that loads `handler/cron.js`.
2. Audit the isolated `cron_sets` collection and remove every active row except the intended disposable fixture.
3. Seed a unique `{guild,id}` row targeting a disposable channel. Set `cron` to the next `Asia/Taipei` minute and use this message shape:

```javascript
{
  content: "staging auto-notification",
  embeds: [{
    data: {
      title: "Staging delivery",
      description: "numeric color compatibility",
      color: 10944422
    }
  }]
}
```

4. Enable the delivery, Gateway, and lease flags with a unique owner.
5. Run staging preflight and require its warning about `cron_sets` sends, `mhcat_scheduler_locks` writes, and Node ownership.
6. Start one Go bot. Inspect the lease with `go run ./cmd/mhcat-scheduler-lease --name auto-notification-delivery --action status` and verify the expected owner holds it.
7. At the scheduled UTC+8 minute, verify exactly one message with the expected content/embed and color `#A6FFA6`.
8. If testing two Go replicas, use distinct owner names and verify only the lease holder sends.
9. Change the cron and allow up to 30 seconds for reconciliation. Verify the old expression no longer fires.
10. Delete the fixture before its next occurrence and verify no future message is sent.
11. Gracefully stop Go and verify the lease is released. Remove the fixture and disable delivery.

## Rollback

1. Stop the Go bot and verify `auto-notification-delivery` is released or expired.
2. Set `MHCAT_FEATURE_AUTO_NOTIFICATION_DELIVERY_ENABLED=false`.
3. Restore Node only after Go can no longer send.
4. Keep `cron_sets` unchanged unless removing disposable staging fixtures; both implementations read the legacy payload shape.

Setup handlers do not own usage writes. The separate global middleware increments `all_use_counts` when `MHCAT_FEATURE_USAGE_TRACKING_ENABLED=true`.
