# Birthday Config and Profile Slice

Status: implemented behind explicit runtime and command-sync gates.

## Legacy References

- Command: `MHCAT/slashCommands/生日系統/birthday.js`
- Guild config model: `MHCAT/models/birthday_set.js`
- User birthday model: `MHCAT/models/birthday.js`
- Commented scheduler: `MHCAT/handler/gift.js`

## Implemented Surface

This slice implements:

- `/生日系統 祝福語設定`
- `/生日系統 增加`
- `/生日系統 是否允許管理員設定`
- `/生日系統 刪除`
- `/生日系統 生日列表`

The command definition preserves all five legacy subcommands:

- `祝福語設定`
- `增加`
- `刪除`
- `是否允許管理員設定`
- `生日列表`

`增加` implements the legacy two-step hour/minute select flow and writes the selected date/time into `birthdays` after the minute selection is completed.

## UI/UX Parity

The implemented config path preserves:

- command name `生日系統`
- command description `讓你的伺服器可以為生日的人慶生!`
- subcommand and option names/descriptions
- legacy UTC choices `UTC+0` through `UTC+23`, with values `+00:00` through `+23:00`
- runtime Manage Messages permission check
- Discord.js Administrator override for every runtime permission check
- public defer/edit response flow
- success title `<:cake:1065654305983570041> 生日系統祝福語設定`
- success description fields for message, UTC, self-set permission, notification channel, and role
- legacy `client.emoji.channel` value `<:Channel:994524759289233438>`
- exact message whitespace, including a message made entirely of whitespace

The implemented profile paths preserve:

- public defer/edit response flow
- `/生日系統 增加` legacy config-required, Manage Messages, self-only, admin-opt-out, year, month, and day validation errors
- `/生日系統 增加` hour select title, description, placeholder, footer text, 24 hour options, and 5-minute-step minute options
- legacy date quirks: explicit year `0` is accepted, February allows day 29 regardless of year, and no other leap-year check runs
- nearest-second rounding for the visible five-minute deadline and reuse of the original slash-command avatar in both select prompts
- `/生日系統 增加` final success title and visible date/time description
- `刪除` runtime Manage Messages permission check
- delete success title `<:trashbin:995991389043163257> 刪除生日日期資料`
- allow-admin success title `<a:green_tick:994529015652163614> 成功變更資料`
- allow-admin footer `本人還是可以設定喔!`
- list title `🎂 生日列表`
- random list embed color
- list attachment name `discord.txt`
- cached attachment names in the exact `username#discriminator(id)` form, including `username#0`; uncached members use `找不到使用者!(id)`
- inline member rows only below 100 profiles; 100 or more profiles use the legacy attachment-only notice
- legacy missing-data errors for delete and list

The command definition intentionally does not set Discord-side default member permissions. Legacy registered the command without a default permission gate and enforced Manage Messages inside the command body for the config/admin paths.

The legacy `hour_menu` and `min_menu` custom IDs are not registered by the Go router. They are generic IDs reused by unrelated legacy flows, so registering them broadly would risk handling the wrong live message. Go uses versioned IDs:

- `mhcat:v1:birthday:hour:state=<id>`
- `mhcat:v1:birthday:minute:state=<id>`

The state ID is a collision-checked random 96-bit value pointing to process-local pending state. State expires exactly at the same five-minute boundary as the legacy Discord collector, and expired entries are pruned as new flows begin. Random IDs prevent an old menu from binding to a newly created flow after a process restart. If the process restarts or state expires, the user must rerun `/生日系統 增加`.

The Go add flow writes only after the minute selection succeeds. Legacy deletes an existing `birthday` row before showing the hour select, which can lose data if the user times out. Go keeps the existing row unchanged until the replacement date/time is complete; this is an intentional data-safety fix.

## Mongo Compatibility

Collection: `birthday_sets`

Fields:

- `guild`
- `msg`
- `utc`
- `channel`
- `everyone_can_set_birthday_date`
- `role`

The Go repository writes the legacy fields and updates all duplicate `{guild}` rows before falling back to an upsert when no row exists. It does not create indexes during bot startup. The candidate `{guild:1}` singleton index remains duplicate-audit gated.

The role field is stored as `null` when the optional role is not provided, matching the legacy command's `role ? role.id : null` write behavior.

The config decoder accepts the legacy BSON boolean and numeric boolean shapes for `everyone_can_set_birthday_date`; a null role remains unset.

Collection: `birthdays`

Fields:

- `guild`
- `user`
- `birthday_year`
- `birthday_month`
- `birthday_day`
- `send_msg_hour`
- `send_msg_min`
- `allow`

The Go repository updates all duplicate `{guild,user}` rows before falling back to an upsert when no row exists. `刪除` deletes duplicate `{guild,user}` rows to keep rollback-compatible data cleanup. It does not create indexes during bot startup; `{guild:1,user:1}` remains duplicate-audit gated.

`增加` writes:

- `birthday_year`
- `birthday_month`
- `birthday_day`
- `send_msg_hour`
- `send_msg_min`
- `allow: true`

The profile decoder accepts the scalar shapes produced or cast by the Mongoose schema: null optional numbers, integral BSON doubles, `int32`, `int64`, and boolean values for numeric fields, plus boolean or numeric values for `allow`. This lets Go read existing Mongoose `Number` rows without changing their meaning.

`是否允許管理員設定` intentionally fixes a clear legacy bug: the Node command deletes or creates the replacement `birthday` document but never calls `save()`, even though it replies that the setting changed. Go persists the stated `allow` value and preserves existing birthday date/time fields when present. If no profile exists, Go writes a legacy-compatible placeholder row with null date/time fields.

Birthday-list attachment names use Discord's guild-member cache only and never fall back to REST, matching `interaction.guild.members.cache.get(...)` in Node. The embed continues to show visible `<@user>` rows in repository order.

The list and add-success responses suppress allowed mentions while preserving visible `<@user>` text. This differs from legacy ping behavior as a safety fix to avoid notifying listed or configured members from management flows.

## Gates

Runtime:

```bash
MHCAT_FEATURE_BIRTHDAY_CONFIG_ENABLED=true
```

Command sync:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_BIRTHDAY_CONFIG=true
```

Both flags must be paired in staging. `mhcat-staging-preflight` and the staging command-sync scripts reject birthday command sync when the runtime flag is not enabled.

## Not Implemented

This slice does not implement:

- birthday notification sends
- temporary birthday role add/remove
- recurring birthday scheduler ownership

The scheduler block in `handler/gift.js` is commented out in legacy, so this slice intentionally does not restore birthday delivery or role scheduling.

## Staging Checklist

1. Use an isolated staging guild and staging database.
2. Run command sync dry-run with `MHCAT_COMMAND_SYNC_INCLUDE_BIRTHDAY_CONFIG=true`.
3. Enable `MHCAT_FEATURE_BIRTHDAY_CONFIG_ENABLED=true` before applying the command definition.
4. Run `mhcat-staging-preflight`.
5. Apply guild-scoped command sync only after the paired gate checks pass.
6. Verify `/生日系統 祝福語設定` writes `birthday_sets`, preserves message whitespace, and renders the legacy success embed for both Manage Messages and Administrator actors.
7. Verify `/生日系統 是否允許管理員設定` writes `birthdays.allow` and preserves/nulls date fields as expected.
8. Verify `/生日系統 增加` renders the hour select, then the minute select, then writes `birthdays` date/time fields with `allow: true`; include explicit year `0` and February 29 cases.
9. Verify stale or cross-user birthday add component clicks return an ephemeral retry/error message and do not update the public prompt.
10. Verify `/生日系統 刪除` deletes the target staging user's `birthdays` row and preserves the legacy missing-data error.
11. Verify `/生日系統 生日列表` renders cached legacy and `#0` user tags, missing-user fallback text, the 100-row cutoff, and `discord.txt` without firing mention pings.
