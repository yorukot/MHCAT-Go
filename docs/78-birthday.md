# Birthday Parity Contract

Status: parity-audited against the active legacy slash command, Mongoose schemas, error renderer, global slash dispatcher, and discord.js behavior. Runtime and command sync remain disabled by default. Live staging smoke is still required before production ownership.

## Scope

This contract covers all active `/生日系統` behavior:

- `祝福語設定`;
- `增加` and its hour/minute selectors;
- `刪除`;
- `是否允許管理員設定`;
- `生日列表`;
- `birthday_sets` and `birthdays` compatibility;
- usage ownership, migration, staging, and rollback.

Legacy sources:

- `slashCommands/生日系統/birthday.js`
- `models/birthday_set.js`
- `models/birthday.js`
- `functions/errors_edit.js`
- `events/SlashCommands.js`

The birthday sender and temporary-role block in `handler/gift.js` is entirely commented out. It is inactive legacy behavior and is not restored by this contract.

## Gates And Ownership

Birthday routes require:

```bash
MHCAT_FEATURE_BIRTHDAY_CONFIG_ENABLED=true
MHCAT_STAGING_MODE=true
MHCAT_COMMAND_SYNC_INCLUDE_BIRTHDAY_CONFIG=true
```

Command sync is guild-scoped and staging-only. Preflight and staging scripts reject command sync when the runtime flag is absent.

Stop or gate the Node `/生日系統` owner for the target bot and guild before Go owns the command. Do not run Node and Go writers concurrently against the same birthday collections. No scheduler, Gateway intent, recurring worker, notification send, or role mutation is enabled by these flags.

## Command And Usage Contract

The definition preserves the exact public command name, description, five subcommands, option order/types/descriptions, required flags, and all 24 UTC choices from `UTC+0`/`+00:00` through `UTC+23`/`+23:00`. It intentionally has no Discord default-member-permission gate.

Legacy metadata declares cooldown `5`, but the global dispatcher does not enforce cooldowns. Go therefore adds no birthday cooldown.

The command publicly defers. `祝福語設定` requires Manage Messages at runtime. `增加` requires Manage Messages only when the config disallows self-service or when setting another user. Because the legacy delete comparison is broken, `刪除` is effectively manager-only and Go preserves that effective behavior. `是否允許管理員設定` and `生日列表` are public.

Discord Administrator satisfies every Manage Messages check, matching discord.js permission behavior.

Usage belongs only to the global slash middleware. With `MHCAT_FEATURE_USAGE_TRACKING_ENABLED=true`, one best-effort attempt is recorded before route validation. Birthday handlers and selector routes do not write usage directly.

## Config Contract

`祝福語設定` preserves exact message whitespace, UTC value, selected channel, Boolean self-service setting, optional role, green success embed, and legacy error/docs UI. A missing role is displayed as `null` and stored as BSON null.

Legacy deletes the first config row and starts an unawaited replacement insert. Go applies typed field patches to every duplicate `{guild}` match and upserts only when none match. This removes the delete/insert gap while preserving unrelated fields and rollback readability.

## Add Selector Contract

`增加` preserves validation order and messages:

1. birthday config must exist;
2. self-service/Manage Messages policy;
3. non-manager users may target only themselves;
4. managers cannot target a profile with `allow:false`;
5. optional year, month, and day validation.

Year `0` and an omitted year are accepted. Nonzero years must be from 1900 through the current year. February permits day 29 regardless of leap year.

The hour selector preserves all 24 legacy options in order `1..23,0`; the minute selector preserves `0..55` in five-minute steps. Labels, descriptions, emoji IDs, placeholders, titles, footer text/avatar, and random colors remain exact. The visible deadline uses JavaScript-equivalent nearest-second rounding, while state expires exactly five minutes after command handling begins.

Raw legacy IDs `hour_menu` and `min_menu` collide with cron. Go emits owner-scoped versioned IDs:

- `mhcat:v1:birthday:hour:state=<random-id>`
- `mhcat:v1:birthday:minute:state=<random-id>`

State is process-local, random, and valid for five minutes. Cross-user, expired, malformed, or post-restart clicks receive an ephemeral retry/error and do not alter the public prompt.

Legacy deletes an existing profile before selector completion. Go intentionally keeps it unchanged until minute confirmation succeeds, preventing timeout/restart data loss. Successful self and manager saves write the selected date/time with `allow:true` and preserve the exact completion text. Visible user mentions are suppressed as an intentional safety policy.

## Delete, Preference, And List Contract

`刪除` preserves effective manager-only authorization, missing-profile UI, green delete embed, and visible target mention. Go deletes every duplicate `{guild,user}` row; the legacy command deletes only the first arbitrary match.

Legacy `是否允許管理員設定` deletes the current row and constructs a replacement but never calls `save()`. Go intentionally persists the preference promised by the success embed. Existing date/time fields are preserved; a new preference creates a profile with null date/time fields.

`生日列表` preserves repository order, random color, title, count text, date rendering, and `discord.txt` bytes. Attachment names use cached `username#discriminator`; missing members use `找不到使用者!`. The embed uses visible mentions below 100 profiles and switches to the legacy file-only notice at 100. Go does not fetch uncached users and suppresses mention notifications.

## Mongo Compatibility

Collections and fields remain exact:

- `birthday_sets`: `guild`, `msg`, `utc`, `channel`, `everyone_can_set_birthday_date`, `role`;
- `birthdays`: `guild`, `user`, `birthday_year`, `birthday_month`, `birthday_day`, `send_msg_hour`, `send_msg_min`, `allow`.

Separate read DTOs apply Mongoose-compatible scalar conversion. String fields accept usable BSON scalar strings/numbers/Booleans/ObjectIDs; Boolean fields accept legacy Boolean, numeric, and recognized string forms; numeric profile fields accept exact integral Mongoose Number forms. Null/missing optional numbers remain nil. Compound, nonnumeric, fractional, or overflowing values remain unusable without aborting a broad list scan.

Writes remain typed BSON strings, integers, Booleans, and nulls. Save operations patch all duplicate logical-key rows before upserting. Deletes remove duplicate profile matches. The application creates no startup index and runs no repair, deduplication, or backfill.

Candidate indexes on `birthday_sets.guild` and `{birthdays.guild,birthdays.user}` remain blocked on duplicate, missing/null/blank key, scalar-drift, malformed date/time, external writer, and ownership audits. The production read-only inventory observed 9,683 `birthdays` rows with only `_id_`; no index apply is authorized by this contract.

## Intentional Differences

- Existing profiles are retained while add selectors are pending or stale.
- The allow-admin preference is actually persisted.
- Duplicate config/profile rows are aligned or deleted rather than leaving arbitrary survivors.
- Database failures return controlled UI rather than leaking raw errors.
- Visible management-flow user/channel/role text does not trigger mentions.
- Generic legacy component IDs are replaced by versioned owner-scoped state IDs.

Exact metadata, validation order, effective permissions, selector payloads, five-minute deadline, response text, attachment bytes, collection/field names, and typed Node rollback compatibility are preserved.

## Migration And Staging

Before enabling birthday ownership:

1. Stop the Node birthday command owner for the target bot/guild.
2. Use an isolated staging guild/database and disposable birthday rows.
3. Audit duplicate logical keys, missing/null/blank keys, scalar types, malformed date/time values, role/channel/user IDs, and dashboard/external writers.
4. Preserve both collections and every field. Do not deduplicate, normalize, backfill, or create indexes merely to enable Go.
5. Pair runtime and command-sync flags, then run preflight and command-sync dry-run before apply.
6. Keep birthday delivery and temporary-role scheduling absent; enabling command ownership does not authorize them.

## Parity Tests

Focused tests lock definitions, permissions, validation order, exact selectors/UI/attachments, random color bounds, expiry, stale-state safety, scalar reads, typed duplicate-safe writes, global usage ownership, app wiring, gates, command sync, and preflight. Run:

```bash
go test ./internal/core/domain ./internal/core/services/birthday ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/discord/features/birthday ./internal/app ./internal/config ./internal/parity ./cmd/mhcat-command-sync ./cmd/mhcat-staging-preflight
go test -race ./internal/core/services/birthday ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/discord/features/birthday ./internal/app
go vet ./internal/core/domain ./internal/core/services/birthday ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/discord/features/birthday ./internal/app ./cmd/mhcat-staging-preflight
```

## Staging Smoke

1. Confirm no Node birthday owner and no birthday scheduler is active for the staging guild.
2. Audit both collections and confirm no repair/index apply is planned.
3. Run preflight, command-sync dry-run, reviewed guild apply, and runtime startup with paired flags.
4. Confirm public discoverability and exact manager/non-manager/Administrator authorization for all five subcommands.
5. Test config with/without role, true/false self-service, whitespace-only message, duplicate rows, and scalar-seeded rows; compare exact embed and BSON.
6. Test self add, manager add, non-manager cross-user denial, target opt-out, omitted/zero/invalid/future year, all month bounds, and February 29.
7. Compare all hour/minute options and exact completion text. Test wrong-user, malformed, expired, and restart-stale selectors; confirm an old profile remains unchanged until completion.
8. Test preference create/update and confirm fields are preserved/null as specified.
9. Test manager/non-manager delete, missing row, and duplicate deletion.
10. Test list cache hit/miss, `#0`, null date fields, 99/100 cutoff, exact `discord.txt`, repository order, and no mention notifications.
11. With usage tracking enabled separately, verify one increment per slash attempt and none for component interactions.
12. Disable gates, remove only the managed staging command, preserve data, and execute rollback checks.

## Rollback

1. Disable the birthday command-sync include and remove only the managed staging command.
2. Disable `MHCAT_FEATURE_BIRTHDAY_CONFIG_ENABLED` and stop all Go processes that can route birthday interactions.
3. Wait five minutes for any visible Go selector IDs to expire; they cannot be resumed by Node.
4. Preserve `birthday_sets` and `birthdays`. Typed Go writes remain Mongoose-readable; do not repair or mutate indexes during emergency rollback.
5. Restore only the active Node `/生日系統` command path and verify config, completed add, preference, delete, and list in staging.
6. Do not uncomment or deploy the legacy `handler/gift.js` birthday block as part of rollback.
7. Re-enable production ownership only after confirming the alternate command runtime is stopped.
