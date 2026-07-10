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

The command preserves the legacy public defer/edit behavior, runtime `KickMembers` permission requirement, success embed titles, color choices, delete icon, one-decimal day rounding, and the legacy typo `發送使用者資運`.

The command definition intentionally does not set Discord-side `DefaultMemberPermissions`. Legacy registered the command without a default permission gate and returned its own public Kick Members error embed at runtime, so the Go definition keeps the command visible and enforces permission in the handler.

The member gate preserves the legacy DM title, bilingual DM body, configured-hours footer, kick reason, optional log embed title/field/footer, rounded Discord timestamp, random log color, and event ordering. An account is kicked only while its age is strictly less than the configured threshold; an account exactly at the threshold is allowed. The gate runs before join-role assignment so a too-new member does not receive join-role or welcome side effects after being kicked.

## Mongo Compatibility

Collection: `create_hours`

Fields:

- `guild`: Discord guild ID string.
- `hours`: string number of seconds.
- `channel`: nullable Discord channel ID string.

Missing and BSON `null` channel values decode as no log channel. String channel IDs are preserved. A missing, BSON `null`, malformed, zero, or negative `hours` value is treated as an inactive legacy gate for member joins, matching the legacy `Number(data.hours)` comparison: the member is not kicked and later join-role/welcome handlers continue. Database transport errors are still surfaced rather than being mistaken for malformed data, but use the event dispatcher's continue-and-report path so independent welcome/join-role handlers still run.

No index is created by app startup. The candidate `{guild:1}` index remains duplicate-audit gated.

Intentional compatibility fix: deleting the config removes all duplicate rows for the guild. This is safer than preserving arbitrary duplicate singleton rows and is documented as a cleanup behavior. Updating hours preserves the existing channel even if the prior `hours` value is malformed.

## Reliability Notes

The legacy implementation did not await/catch DM and kick promises. The Go implementation awaits those operations in order. It intentionally ignores non-context DM delivery failures so closed DMs do not bypass the kick, but it returns kick/log failures to the event dispatcher for logging.

Intentional bug fix: if the kick fails, the Go implementation does not send the legacy BAN-style log embed. Legacy could still attempt the log because `member.kick()` was not awaited, but that could mislead admins into thinking a member was kicked when Discord rejected the kick.

Guild-name lookup is best effort; if it fails after the gateway event already provided account creation time, the DM falls back to the guild ID rather than bypassing the gate.

Allowed mentions are suppressed in command, DM, and log responses. This prevents a crafted guild name or migrated data from producing an unintended mention.

Member-event avatars use guild-specific display avatars when Discord provides them, matching `member.displayAvatarURL()` in the legacy runtime.

## Verification Coverage

Automated coverage locks:

- all four command definitions, option order, descriptions, required flags, and text/news channel restrictions;
- public defer behavior, runtime permission rejection, success/delete/error embeds, colors, icons, rounded day text, and the legacy typo;
- positive-hour conversion to seconds and channel preservation during hour updates;
- missing/null/string channel decoding and invalid legacy `hours` shapes;
- strict pre-threshold kick and exact-threshold allow behavior;
- malformed-threshold fail-open behavior through the event dispatcher;
- DM, kick reason, log embed, timestamp rounding, avatar, footer, and allowed-mention payloads;
- kick failure, missing log channel, gateway account timestamp, guild-name fallback, and downstream handler ordering;
- runtime/config/command-sync/preflight gates.

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
