# Birthday Config Slice

Status: implemented behind explicit runtime and command-sync gates.

## Legacy References

- Command: `MHCAT/slashCommands/生日系統/birthday.js`
- Guild config model: `MHCAT/models/birthday_set.js`
- User birthday model: `MHCAT/models/birthday.js`
- Commented scheduler: `MHCAT/handler/gift.js`

## Implemented Surface

This slice implements only:

- `/生日系統 祝福語設定`

The command definition preserves all five legacy subcommands:

- `祝福語設定`
- `增加`
- `刪除`
- `是否允許管理員設定`
- `生日列表`

Only `祝福語設定` writes data in this slice. The other subcommands return a staged unavailable embed so the command shape can be reviewed in staging without claiming the birthday date/profile flows are complete.

## UI/UX Parity

The implemented config path preserves:

- command name `生日系統`
- command description `讓你的伺服器可以為生日的人慶生!`
- subcommand and option names/descriptions
- legacy UTC choices `UTC+0` through `UTC+23`, with values `+00:00` through `+23:00`
- runtime Manage Messages permission check
- public defer/edit response flow
- success title `<:cake:1065654305983570041> 生日系統祝福語設定`
- success description fields for message, UTC, self-set permission, notification channel, and role
- legacy `client.emoji.channel` value `<:Channel:994524759289233438>`

The command definition intentionally does not set Discord-side default member permissions. Legacy registered the command without a default permission gate and enforced Manage Messages inside the command body for the config/admin paths.

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

- birthday date add/delete/profile writes in `birthdays`
- `是否允許管理員設定`
- `生日列表`
- `hour_menu` and `min_menu` component flows
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
6. Verify `/生日系統 祝福語設定` writes `birthday_sets` and renders the legacy success embed.
7. Verify the other birthday subcommands return the staged unavailable embed.
