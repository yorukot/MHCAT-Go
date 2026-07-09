# Account-Age Protection Slice

Status: implemented behind explicit staging/runtime gates.

## Legacy References

- Command: `MHCAT/slashCommands/群組防護/create_hours.js`
- Member-add event: `MHCAT/events/welcome.js`
- Model: `MHCAT/models/create_hours.js`

## Implemented Surface

- `/帳號需創建時數 小時數`
- `/帳號需創建時數 被踢出資訊頻道`
- `/帳號需創建時數 創建時數刪除`
- `/帳號需創建時數 被踢出資訊頻道刪除`
- Optional `guildMemberAdd` account-age gate.

The command is included in command sync only when `MHCAT_COMMAND_SYNC_INCLUDE_ACCOUNT_AGE_CONFIG=true`, and its runtime handler is enabled only when `MHCAT_FEATURE_ACCOUNT_AGE_CONFIG_ENABLED=true`.

The member-add gate is separate and requires all of:

```bash
MHCAT_FEATURE_ACCOUNT_AGE_POLICY_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true
```

## UI/UX Parity

The command preserves the legacy public defer/edit behavior, runtime `KickMembers` permission requirement, success embed titles, color choices, delete icon, and the legacy typo `發送使用者資運`.

The command definition intentionally does not set Discord-side `DefaultMemberPermissions`. Legacy registered the command without a default permission gate and returned its own public Kick Members error embed at runtime, so the Go definition keeps the command visible and enforces permission in the handler.

The member gate preserves the legacy DM title, bilingual DM body, kick reason, optional log embed title/field/footer, and event ordering. The gate runs before join-role assignment so a too-new member does not receive join-role or welcome side effects after being kicked.

## Mongo Compatibility

Collection: `create_hours`

Fields:

- `guild`: Discord guild ID string.
- `hours`: string number of seconds.
- `channel`: nullable Discord channel ID string.

No index is created by app startup. The candidate `{guild:1}` index remains duplicate-audit gated.

Intentional compatibility fix: deleting the config removes all duplicate rows for the guild. This is safer than preserving arbitrary duplicate singleton rows and is documented as a cleanup behavior. Updating hours preserves the existing channel even if the prior `hours` value is malformed.

## Reliability Notes

The legacy implementation did not await/catch DM and kick promises. The Go implementation intentionally ignores non-context DM delivery failures so closed DMs do not bypass the kick, but it does return kick/log failures to the event dispatcher for logging.

Intentional bug fix: if the kick fails, the Go implementation does not send the legacy BAN-style log embed. Legacy could still attempt the log because `member.kick()` was not awaited, but that could mislead admins into thinking a member was kicked when Discord rejected the kick.

Guild-name lookup is best effort; if it fails after the gateway event already provided account creation time, the DM falls back to the guild ID rather than bypassing the gate.

## Staging Checklist

1. Use an isolated staging guild and staging database.
2. Run command sync dry-run with `MHCAT_COMMAND_SYNC_INCLUDE_ACCOUNT_AGE_CONFIG=true`.
3. Enable `MHCAT_FEATURE_ACCOUNT_AGE_CONFIG_ENABLED=true` before applying the command definition.
4. Configure `/帳號需創建時數 小時數` with a safe staging threshold.
5. Optionally configure `被踢出資訊頻道` to a staging-only channel.
6. Enable the member gate only with a disposable too-new account.
7. Verify DM, kick reason, optional log embed, and that later member-add handlers stop.

## Not Implemented

- Production rollout.
- Unique index creation.
- Automated data repair for duplicate `create_hours` rows.
- Any anti-scam URL protection behavior from the same legacy command category.
