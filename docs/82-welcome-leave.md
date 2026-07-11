# Welcome And Leave Message Parity Contract

Status: parity-audited against the active legacy dashboard redirect, leave setup/modal flow, `guildMemberAdd` and `guildMemberRemove` handlers, Mongoose schemas, slash dispatcher, discord.js payload behavior, and gateway cache semantics. Config and both delivery paths remain disabled by default. Live staging smoke is still required before production ownership.

## Scope

This contract covers:

- `/加入訊息設置` dashboard redirect;
- `/退出訊息設置` channel selection, modal, persistence, and preview;
- generic `guildMemberAdd` welcome delivery from `join_messages`;
- optional special MHCAT welcome delivery;
- `guildMemberRemove` leave delivery from `leave_messages`;
- usage, intents, ownership, migration, staging, and rollback.

Legacy sources:

- `slashCommands/加入設置/join_messag.js`
- `slashCommands/加入設置/leave_message.js`
- `events/modal.js`
- `events/welcome.js`
- `models/join_message.js`
- `models/leave_message.js`
- `handler/slash_commands.js`
- `events/SlashCommands.js`

Join-role, account-age, and verification are separate ownership families.

## Gates And Ownership

The config commands require paired staging-only flags:

```bash
MHCAT_FEATURE_WELCOME_MESSAGE_CONFIG_ENABLED=true
MHCAT_STAGING_MODE=true
MHCAT_COMMAND_SYNC_INCLUDE_WELCOME_MESSAGE_CONFIG=true
```

Generic and special welcome delivery are event-only and independently require:

```bash
MHCAT_FEATURE_WELCOME_MESSAGE_DELIVERY_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true
```

Leave delivery is separately gated:

```bash
MHCAT_FEATURE_LEAVE_MESSAGE_DELIVERY_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true
```

Command sync is guild-scoped and staging-only. Stop matching Node slash/modal/member-event owners before enabling Go. Config, welcome delivery, and leave delivery may migrate separately, but no family may have concurrent Node and Go ownership.

The legacy special welcome is empty by default. All seven `MHCAT_LEGACY_WELCOME_SPECIAL_*` IDs must be configured together; Go does not hardcode private guild/channel IDs.

## Command And Usage Contract

Both definitions preserve exact names, descriptions, required text/news channel options, channel type order, and malformed legacy wording. `/加入訊息設置` also preserves its malformed documentation URL. Legacy cooldown metadata is `10`, but the global dispatcher does not enforce it. Go adds no cooldown.

Both commands remain publicly discoverable because legacy `UserPerms` is help metadata, not a Discord default permission. Runtime behavior differs by command exactly as legacy:

- `/加入訊息設置` has no permission check, ignores the required channel, replies publicly without deferring, and writes nothing;
- `/退出訊息設置` requires Manage Messages at runtime; denial is an immediate ephemeral red animated-no reply, while success opens the modal without deferring.

The dashboard reply preserves:

- title `<a:announcement:1005035747197337650> | 該指令已經移往控制面板，請前往控制面板進行設定`;
- color `#df1f2f`;
- URL `https://mhcat.yorukot.meguilds/<guild>/welcome` including the missing slash;
- link label `點我前往儀錶板設定!` and arrow emoji.

Usage belongs only to global slash middleware. With tracking enabled, exactly one best-effort event is recorded before route/permission/validation checks for every slash attempt. Handlers, modal submits, and member events do not write usage directly.

## Leave Setup UI

The modal keeps custom ID `nal`, title `退出訊息設置!`, and field order color, title, content. Exact IDs and labels remain:

- `leave_msgcolor`: `請輸入你的加入訊息要甚麼顏色(要隨機顏色可輸入:Random)`;
- `leave_msgtitle`: `請輸入訊息標題`;
- `leave_msgcontent`: `請輸入訊息內文(如要顯示用戶名可輸入: {MEMBERNAME} )`.

All fields are required. Existing raw whitespace is retained in modal defaults. The legacy router chose a modal route from the first field, so Go accepts both actual color-first `leave_msgcolor` and older `leave_msgcontent` route evidence.

Modal submit defers publicly. It accepts exact `Random`, six-digit hex with or without `#`, and exact pinned discord.js color names. Padded values, lowercase aliases, short hex, and `RANDOM` are rejected with `你傳送的並不是顏色(色碼)`. Unknown/missing config returns `很抱歉，出現了未知的錯誤!`.

The preview preserves raw title/content, member avatar thumbnail, timestamp, selected color, and exact content typo:

`下面為預覽，想修改嗎?再次輸入指令即可修改((MEMBERNAME)在到時候會變正常喔)`

## Mongo Compatibility

Collections and fields remain exact:

- `join_messages`: `guild`, `enable`, `message_content`, `color`, `channel`, optional `img`;
- `leave_messages`: `guild`, nullable `message_content`, nullable `title`, nullable `color`, `channel`.

Separate permissive read and typed write DTOs preserve legacy catalogs. Mongoose String scalars, including booleans, numbers, decimals, and ObjectIDs, are usable without trimming. Missing/null/compound message values become unusable empty values rather than BSON decode failures.

For `join_messages.enable`, only Mongoose-cast false values disable delivery: Boolean `false`, numeric zero, and exact strings `false`, `0`, or `no`. Missing, null, true values, and uncastable/compound values remain enabled because legacy checks `enable === false`.

`/加入訊息設置` never writes `join_messages`; dashboard remains its owner. Leave setup creates an absent row with string `guild`/`channel` and null content/title/color. Modal save writes typed strings in one atomic `$set`.

Reads retain legacy first-row behavior. Explicit leave setup updates every duplicate guild row together, intentionally fixing legacy first-match and three non-awaited update calls that could split fields across duplicates. Rows are not deleted, merged, or backfilled, so rollback remains Mongoose-compatible.

No startup repair or index creation runs. Candidate unique `{guild:1}` indexes for both collections remain blocked on duplicate, missing/null/blank/scalar-drift keys, malformed content/colors, dashboard/external writers, and exclusive ownership review. No migration is required merely to enable Go.

## Generic Welcome Delivery

Missing config, explicit disabled state, missing channel/content/color, or an uncached target channel is a no-op. Channel lookup is cache-only before payload construction, matching `guild.channels.cache.get`; Go never REST-sends around a stale/missing cache entry.

The embed preserves:

- author `🪂 歡迎加入 <guild name>`;
- guild icon, falling back to the bot's guild-specific avatar;
- joining member's guild-specific avatar thumbnail;
- optional `img`, pinned discord.js color, and timestamp;
- first-occurrence replacement order `(MEMBERNAME)`, `{MEMBERNAME}`, `{membername}`, `(TAG)`, `{TAG}`, `{tag}`;
- JavaScript replacement-string behavior for `$$`, `$&`, ``$` ``, and `$'`.

`RANDOM` is converted to discord.js `Random`; `Random` is already valid. Invalid nonempty stored colors reproduce the legacy no-send outcome as a controlled event error rather than sending a red substitute. Raw all-space content remains truthy and is delivered.

Only the joining user ID is allowed for tag placeholders. Everyone, role, and unrelated user pings are suppressed intentionally.

## Special Welcome Delivery

When all special IDs match the joining guild and current bot, special delivery replaces generic delivery. It preserves the exact MHCAT author, `https://dsc.gg/MHCAT` URL, channel references/text, random color, member thumbnail, `https://i.imgur.com/cLCPRNq.png` image, timestamp, and manual `username#discriminator` form, including `username#0`.

The author icon uses the bot's cached guild-specific avatar. The special target is also cache-only. Missing configuration leaves special mode disabled; no IDs are compiled into Go.

## Leave Delivery

Missing/incomplete config, missing content, or an uncached target channel is a no-op. Title is sent raw and receives no placeholder replacement. Description replacement order is `(MEMBERNAME)`, `(ID)`, `{ID}`, `{MEMBERNAME}`, with first-occurrence and JavaScript replacement-string semantics.

The embed uses the departing member's guild-specific avatar thumbnail, stored pinned color or `RANDOM`/`Random`, and timestamp. It has no image. Invalid nonempty stored colors produce a controlled no-send event error. All mentions are suppressed.

## Event Ordering And Intentional Differences

Account-age policy registers first and stops welcome and join-role behavior after a matched kick. Welcome registers before join-role. Welcome failures continue to role assignment; later role failures cannot suppress an already-sent welcome. Leave delivery is independent.

Intentional differences are limited to:

- duplicate leave rows are aligned on explicit setup;
- modal content fields save atomically and are awaited;
- malformed colors/config return controlled errors rather than unhandled callback exceptions;
- malformed rows missing required title/color values fail closed as no-ops instead of attempting a colorless or literal-null embed;
- missing cached special channel is a no-op instead of a null-channel exception;
- special IDs are explicit environment configuration, not hardcoded;
- mentions are allowlisted/suppressed;
- global middleware is the only usage owner;
- event failures continue to independent handlers while remaining visible to gateway logging.

Exact metadata/UI, visibility, runtime permission, modal field order/text, dashboard typo, collection/field names, cache-only routing, placeholders, valid colors, payloads, and account-age stop behavior are preserved.

## Migration And Staging

1. Use an isolated staging guild/database, staging-only channels, and disposable members.
2. Stop matching Node slash/modal/member-event owners for each enabled family.
3. Audit both collections for duplicates, scalar types, missing/null/blank guild/channel values, malformed colors/content/images, stale channels, indexes, and dashboard/external writers.
4. Preserve data as-is. Do not normalize, deduplicate, backfill, or index merely to enable Go.
5. Pair config runtime/sync flags; run preflight and command-sync dry-run before reviewed guild apply.
6. Enable each delivery path only with Gateway/Guild Members and verified channel cache readiness.
7. Configure all special IDs only for the intended guild/bot and verify every referenced channel.
8. Confirm account-age and join-role ownership/order when member-add families are staged together.

## Parity Tests

Focused coverage locks metadata/public visibility, runtime permission, dashboard/modal/preview UI, usage ownership, scalar/Boolean reads, typed duplicate-safe writes, generic/special/leave embeds, placeholder and replacement quirks, color behavior, cache-only channels, guild-specific identities, event ordering, app wiring, intents, ownership, command sync, and preflight. Run:

```bash
go test ./internal/core/domain ./internal/core/services/onboarding ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/adapters/discordgo ./internal/discord/features/onboarding ./internal/discord/events ./internal/app ./internal/config ./internal/parity ./cmd/mhcat-command-sync ./cmd/mhcat-staging-preflight
go test -race ./internal/core/services/onboarding ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/adapters/discordgo ./internal/discord/features/onboarding ./internal/discord/events ./internal/app
go vet ./internal/core/services/onboarding ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/adapters/discordgo ./internal/discord/features/onboarding ./internal/discord/events ./internal/app
go run ./tools/parity-audit
```

## Staging Smoke

1. Review read-only collection audits and confirm one owner per config/welcome/leave family.
2. Run preflight, command-sync dry-run, reviewed guild apply, and runtime startup.
3. Confirm both commands are publicly discoverable; verify unrestricted public dashboard reply and ephemeral leave permission denial.
4. Verify exact malformed dashboard URL/title/color/button and that the selected join channel is ignored.
5. Verify leave modal order/labels/defaults, exact accepted/rejected colors, public submit/preview, typo, raw whitespace, typed writes, duplicate alignment, and one usage event per slash attempt.
6. Seed scalar/null/compound rows in disposable data and verify controlled reads without repair or index creation.
7. Enable generic welcome and test disabled/missing/incomplete/uncached rows, all placeholders, replacement tokens, guild/bot icon fallback, member avatar, image, colors, and invalid-color no-send.
8. Configure all special IDs and verify the complete special payload; then remove one value and confirm generic behavior resumes.
9. Enable leave delivery and test raw title, description placeholders, replacement tokens, avatar, colors, uncached channel, and invalid-color no-send.
10. Enable account-age and join-role separately; verify matched accounts receive neither welcome nor roles and independent failures continue correctly.
11. Disable gates, remove only managed staging commands, preserve both collections, and perform rollback checks.

## Rollback

1. Disable command-sync inclusion and remove only the two managed staging commands.
2. Disable welcome and leave delivery gates, then config runtime; stop every Go owner.
3. Preserve `join_messages` and `leave_messages`. Typed writes remain Mongoose-readable; do not repair data or indexes during emergency rollback.
4. Restore Node slash/modal/member-event owners only after confirming no matching Go owner remains.
5. Recheck dashboard redirect, one leave setup/preview, one generic join, optional special join, and one leave event in staging.
6. Review any overlap interval for duplicate messages or conflicting leave config writes.

Production ownership remains blocked on live staging smoke and a reviewed duplicate/type/channel/dashboard-writer audit. Unique indexes are optional and must never be created merely to enable Go.
