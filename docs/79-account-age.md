# Account-Age Parity Contract

Status: parity-audited against the active legacy config command, `guildMemberAdd` policy branch, Mongoose schema, global slash dispatcher, and discord.js behavior. Both runtimes and command sync remain disabled by default. Live staging smoke is still required before production ownership.

## Scope

This contract covers:

- `/帳號需創建時數 小時數`;
- `/帳號需創建時數 被踢出資訊頻道`;
- `/帳號需創建時數 創建時數刪除`;
- `/帳號需創建時數 被踢出資訊頻道刪除`;
- the leading account-age branch of `events/welcome.js` on `guildMemberAdd`;
- `create_hours` compatibility, usage, intent, ownership, staging, and rollback.

Legacy sources:

- `slashCommands/群組防護/create_hours.js`
- `events/welcome.js`
- `models/create_hours.js`
- `events/SlashCommands.js`
- `config.json`

Join-role and welcome-message behavior after the account-age branch are separate contracts. Anti-scam URL behavior from the same command category is covered by [77-anti-scam.md](77-anti-scam.md).

## Gates And Ownership

Config command routes require:

```bash
MHCAT_FEATURE_ACCOUNT_AGE_CONFIG_ENABLED=true
MHCAT_STAGING_MODE=true
MHCAT_COMMAND_SYNC_INCLUDE_ACCOUNT_AGE_CONFIG=true
```

The member-add policy is event-only and independently requires:

```bash
MHCAT_FEATURE_ACCOUNT_AGE_POLICY_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true
```

Command sync is guild-scoped and staging-only. Config validation and preflight reject missing Gateway/Guild Members readiness for the policy. Stop the equivalent Node command or member-add branch before Go owns that family. Config can migrate separately from the policy, but concurrent Node/Go writers against `create_hours` must not overlap.

## Command And Usage Contract

The definition preserves the exact public name, description, four subcommands, option order/types/descriptions, required flags, and text/news channel restriction (`0`, `5`). It intentionally sets no Discord default member permissions.

Legacy metadata declares cooldown `10`, but the global dispatcher does not enforce it. Go adds no account-age cooldown.

The command publicly defers and then requires Kick Members for every subcommand. Discord Administrator satisfies that check. Exact error titles, green success/delete embeds, emoji IDs, descriptions, and typo `發送使用者資運` remain unchanged. Visible channel text does not ping because Go explicitly suppresses mentions.

`小時數` rejects zero and negative input with `不可為負數或0!!!`, multiplies the full Discord-safe integer range by 3600 using JavaScript-compatible numeric semantics, stores seconds, and shows one-decimal day rounding. Setting hours preserves the exact existing Mongoose-cast channel scalar. Setting/deleting a log channel requires an existing config. Full config deletion removes every duplicate guild row as an intentional cleanup fix.

Usage belongs only to global slash middleware. With `MHCAT_FEATURE_USAGE_TRACKING_ENABLED=true`, one best-effort attempt is recorded before route lookup and permission/validation checks. Account-age handlers and member events do not write usage directly.

## Member-Add Policy

The policy reads the first `create_hours` row for the guild before join-role or welcome handlers. Missing or unusable config allows later handlers. A member matches only when:

```txt
(current milliseconds - account-created milliseconds) / 1000 < Number(hours)
```

The exact threshold is allowed. Positive finite fractional/exponent forms retain legacy Number behavior. Null, empty, zero, negative, nonnumeric, compound, NaN, and Infinity-like values are treated as inactive rather than risking an unbounded kick policy.

For a match, Go preserves:

- the exact bilingual red DM title/body;
- configured-hours footer and global user avatar;
- exact bilingual kick reason;
- optional cache-only log-channel lookup;
- exact log title, field, nearest-second Discord timestamp, thumbnail/footer/tag, and random color;
- stop-propagation before join-role and welcome behavior.

Legacy starts DM, kick, and log promises without awaiting them. Go intentionally attempts DM, awaits kick, and only then sends a BAN-style log. Non-context DM failure does not bypass the kick. Kick failure sends no misleading BAN log. Log failure is reported after the member is already kicked. A configured channel absent from the guild cache is silently skipped, matching `guild.channels.cache.get()`.

## Mongo Compatibility

Collection and fields remain exact:

- `create_hours.guild`: Mongoose String guild ID;
- `create_hours.hours`: Mongoose String seconds threshold;
- `create_hours.channel`: nullable Mongoose String channel ID.

Separate read and write DTOs are used. Reads apply Mongoose-compatible scalar-to-string conversion, then JavaScript Number parsing for `hours`. Positive finite values, including fractions and exponent strings, are preserved. Null or compound values remain unusable without crashing the event runtime. Raw usable channel whitespace is preserved.

Writes remain typed BSON strings and nulls. Hours/channel updates preserve unrelated fields and legacy one-row update behavior. Explicit full deletion removes duplicate guild rows. No startup index, repair, deduplication, or backfill runs.

The candidate unique `{guild:1}` index remains blocked on duplicate, missing/null/blank/scalar-drift keys, malformed thresholds/channels, external writers, and exclusive ownership review. No data operation is approved merely to enable Go.

## Intentional Differences

- Global middleware is the only slash usage owner.
- Full config deletion removes duplicate guild rows.
- Invalid and non-finite thresholds fail open instead of risking mass kicks.
- DM/kick/log operations are ordered and awaited; failed kicks do not emit BAN logs.
- Guild-name lookup can fall back to the guild ID when the gateway event lacks a name.
- Explicit empty allowed mentions prevent crafted names/config from pinging.
- Backend failures use controlled handling rather than raw-error Discord output.

Exact command metadata/UI, permission behavior, day/timestamp rounding, threshold boundary, DM/kick/log payloads, global user identity, cache-only log lookup, event ordering, collection/field names, and typed Node rollback compatibility are preserved.

## Migration And Staging

Before enabling either family:

1. Use an isolated staging guild/database and disposable member account.
2. Stop the matching Node command/event owner.
3. Audit duplicate keys, missing/null/blank/scalar-drift guilds, raw threshold/channel values, stale channel IDs, and external/dashboard writers.
4. Preserve `create_hours` exactly. Do not normalize, deduplicate, backfill, or create an index merely to enable Go.
5. Pair config runtime/sync flags and run preflight plus command-sync dry-run before apply.
6. Enable the policy only with Gateway/Guild Members and a reviewed safe threshold.
7. Confirm account-age registers before welcome and join-role handlers.

## Parity Tests

Focused tests lock exact command metadata/UI, permissions, rounding, full integer range, scalar/fraction reads, raw channels, typed writes, usage ownership, DM/kick/log payloads and failures, cache/global-avatar behavior, threshold boundaries, event ordering, app wiring, gates, command sync, and preflight. Run:

```bash
go test ./internal/core/domain ./internal/core/services/onboarding ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/adapters/discordgo ./internal/discord/features/onboarding ./internal/app ./internal/config ./internal/parity ./cmd/mhcat-command-sync ./cmd/mhcat-staging-preflight
go test -race ./internal/core/services/onboarding ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/adapters/discordgo ./internal/discord/features/onboarding ./internal/app
go vet ./internal/core/domain ./internal/core/services/onboarding ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/adapters/discordgo ./internal/discord/features/onboarding ./internal/app ./cmd/mhcat-staging-preflight
```

## Staging Smoke

1. Confirm no matching Node config or member-add owner and review the `create_hours` audit.
2. Run preflight, command-sync dry-run, reviewed guild apply, and config runtime startup.
3. Confirm public discoverability and exact Kick Members/non-manager/Administrator UI.
4. Test ordinary, one-hour, fractional-day, and large safe hour inputs; compare exact UI and stored string seconds.
5. Test channel set/delete, missing-config errors, raw/scalar-seeded channel preservation, duplicate behavior, and full delete.
6. With policy disabled, confirm config changes do not kick members.
7. Enable policy readiness and test too-new, exact-threshold, and older disposable accounts.
8. Verify exact DM, kick reason, global avatar/tag, cache-hit log, cache-miss skip, timestamp rounding, and mention suppression.
9. Deny DMs and confirm kick continues; deny kick and confirm no BAN log; deny log send and confirm kick remains.
10. Enable welcome/join-role separately and confirm too-new matches stop both while allowed/invalid-config/read-error cases continue.
11. With usage tracking enabled separately, verify one increment per slash attempt and none for member events.
12. Disable gates, remove only the managed staging command, preserve data, and execute rollback checks.

## Rollback

1. Disable the command-sync include and remove only the managed staging command.
2. Disable policy and config runtime gates; stop all Go processes that can receive the command/event families.
3. Preserve `create_hours`. Typed Go writes remain Mongoose-readable; do not repair or change indexes during emergency rollback.
4. Restore only the intended Node command and/or leading member-add branch after confirming no Go owner remains.
5. Verify one config update, one exact-threshold allow, one too-new DM/kick/log, and downstream stop behavior in staging.
6. Review any overlap interval because duplicate DMs/kicks cannot be transactionally undone.
7. Re-enable production ownership only after confirming the alternate runtime is stopped.
