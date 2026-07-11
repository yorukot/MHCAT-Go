# Delete Data Parity Contract

Status: parity-audited against the active legacy slash command, message component collector, all nine Mongoose models, global slash dispatcher, discord.js response behavior, Mongo collection naming, runtime wiring, command sync, and staging preflight. Runtime and command sync remain disabled by default. Live staging smoke is still required before production ownership.

## Scope

This contract covers:

- `/刪除資料` command metadata, permission, usage, prompt, and select handling;
- owner binding and the one-hour component lifetime;
- deletion of join, leave, logging, stats, autochat, verification, text XP, voice XP, and ticket configuration;
- duplicate cleanup, guild/collection isolation, migration, staging, and rollback.

Legacy sources:

- `slashCommands/管理系統/delete_data.js`
- `models/join_message.js`
- `models/leave_message.js`
- `models/logging.js`
- `models/Number.js`
- `models/chat.js`
- `models/verification.js`
- `models/text_xp_channel.js`
- `models/voice_xp_channel.js`
- `models/ticket.js`
- `events/SlashCommands.js`

## Gates And Ownership

The destructive command requires paired staging-only flags:

```bash
MHCAT_FEATURE_DELETE_DATA_ENABLED=true
MHCAT_STAGING_MODE=true
MHCAT_COMMAND_SYNC_INCLUDE_DELETE_DATA=true
```

Both flags default to false. Command sync is guild-scoped and staging-only. Preflight rejects sync without runtime, warns on runtime without sync, and passes the paired state. Runtime startup requires the default Mongo adapter and constructs all nine target collections before registering the slash and component routes.

Stop the Node `/刪除資料` owner before enabling Go. Never run both owners for the same bot/guilds: either process could accept a destructive selection, and there is no shared lease.

## Command And Usage Contract

The definition remains publicly discoverable with exact name `刪除資料`, description `刪除之前設置過的資料`, no options, and no Discord default-member permission. Legacy `UserPerms` is help metadata. The handler checks Manage Messages (`8192`) at runtime.

The slash interaction defers publicly. Permission denial edits the deferred response with exact red title:

`<a:Discord_AnimatedNo:1015989839809757295> | 你需要有\`訊息管理\`才能使用此指令`

Legacy cooldown metadata is `10`, but the global dispatcher does not enforce it. Go adds no cooldown. Usage belongs only to global slash middleware: each slash attempt is recorded once when tracking is enabled, including denied attempts. Select interactions do not write usage.

## Prompt UI

An allowed slash leaves the deferred original response unchanged and sends one public follow-up. The embed preserves:

- title `<:trashbin:995991389043163257> 刪除資料`;
- exact three-line description, including duplicated warning, `一但`, and missing space after the explosion emoji;
- a random color in the discord.js `Random` range;
- footer `請三思!!!`;
- footer icon and thumbnail `https://media.discordapp.net/attachments/991337796960784424/996749656161779853/6lnjr0.gif`.

The string select keeps custom ID `delete-data`, placeholder `🗑 選擇你要刪除的資料!`, one required value, and this exact order:

| Label/value | Description | Emoji | Collection |
| --- | --- | --- | --- |
| `加入訊息` | `🗑 加入訊息 刪除!` | `<:joines:953970547849592884>` | `join_messages` |
| `離開訊息` | `🗑 離開訊息 刪除!` | `<:leaves:956444050792280084>` | `leave_messages` |
| `審核日誌` | `🗑 審核日誌 刪除!` | `<:logfile:985948561625710663>` | `loggings` |
| `統計系統` | `🗑 統計系統 刪除!` | `<:statistics:986108146747600928>` | `numbers` |
| `自動聊天` | `🗑 自動聊天 刪除!` | `<:ChatBot:956863473910947850>` | `chats` |
| `驗證設置` | `🗑 驗證設置 刪除!` | `<:tickmark:985949769224556614>` | `verifications` |
| `聊天經驗設置` | `🗑 聊天經驗設置 刪除!` | `<:xp:990254386792005663>` | `text_xp_channels` |
| `語音經驗設置` | `🗑 語音經驗設置 刪除!` | `<:Voice:994844272790610011>` | `voice_xp_channels` |
| `私人頻道設置` | `🗑 私人頻道設置 刪除!` | `<:ticket:985945491093205073>` | `tickets` |

All mentions are suppressed.

## Select Ownership And Lifetime

Legacy attaches a collector to the follow-up for exactly one hour and filters to the invoking user. It does not recheck Manage Messages after prompt creation. Go recovers the original interaction user and snowflake from Discord message metadata:

- only that user may select;
- the owner remains authorized for the valid prompt even if their permission changes later;
- a selection just before one hour is accepted;
- a selection at or after one hour is rejected before Mongo access;
- foreign and expired selections use the exact missing-option content.

The static legacy custom ID carries no owner or deadline. For old or synthetic messages without original-user metadata, Go requires current Manage Messages as a compatibility fallback. A missing or malformed original snowflake cannot establish expiry, so no deadline rejection is applied; original-user ownership is still enforced when that independent metadata exists, otherwise the permission fallback applies. Real Discord-authored follow-ups provide valid owner and snowflake metadata.

Node collectors disappear on process restart. Go can continue routing a metadata-backed panel after restart until its original one-hour deadline. This is an intentional availability improvement; it does not extend the deadline or broaden ownership.

The component calls `deferReply()` publicly. Legacy later passes `ephemeral: true` to `editReply`, but reply visibility cannot be changed by an edit, so success and missing responses are public. Go preserves the effective public behavior:

- success: `<a:green_tick:994529015652163614> **| 成功刪除該設定!**`;
- missing/foreign/expired/backend failure: `<a:Discord_AnimatedNo:1015989839809757295> **| 你沒有設定過這個選項!**`.

## Mongo And Deletion Semantics

Every operation filters by exact normalized `{guild}` in exactly one selected collection. It cannot delete another guild or target. No Discord channels/messages/roles are deleted. No usage, repair, marker, audit, or index collections are written by the feature handler.

Legacy uses `findOne({guild})` and `data.delete()`, removing one arbitrary matching row. Go intentionally uses `DeleteMany({guild})` to remove every duplicate singleton row. This prevents a duplicate row from leaving the supposedly deleted feature active. A zero deletion count maps to the legacy missing response. Mongo operational errors intentionally collapse to the same controlled missing response because legacy callbacks ignore `err` and treat absent data as missing.

Mongoose model names pluralize to the nine collection names in the table. The stats model is tracked as `models/Number.js`, but the command requires lowercase `models/number.js`. That branch fails on case-sensitive Node deployments. Go intentionally repairs the branch and deletes from the actual `numbers` collection.

The repository performs no startup read, repair, deduplication, index creation, or schema migration. Existing data is untouched until an authorized select. No migration is required to enable Go. Candidate singleton indexes remain subject to each owning feature's duplicate/type/external-writer review and must not be created merely for delete-data.

## Intentional Differences

Intentional differences are limited to:

- all duplicate selected-guild rows are deleted instead of one arbitrary row;
- the broken case-sensitive stats branch targets the real `numbers` collection;
- metadata-backed panels survive a Go restart only until the original one-hour deadline;
- foreign/expired interactions receive a controlled missing response instead of collector non-delivery;
- metadata-less panels use a Manage Messages fallback;
- missing/malformed synthetic snowflake metadata cannot enforce expiry;
- Mongo failures return controlled UI instead of callback/unhandled-rejection behavior;
- mentions are suppressed.

Exact command metadata, prompt payload, target order, owner binding, deadline, effective public visibility, collection names, guild/target isolation, and response text are preserved.

## Migration And Staging

1. Use an isolated staging guild and disposable database rows only.
2. Stop the matching Node command owner before enabling Go.
3. Back up every target collection that may be selected. Deletion is intentionally irreversible in the UI.
4. Audit all nine collections by `{guild}` for duplicates, missing/null/blank/scalar-drift guild keys, current indexes, and dashboard/external writers.
5. Preserve rows as-is. Do not normalize, deduplicate, backfill, or index before smoke.
6. Pair runtime/sync flags; run preflight and command-sync dry-run before reviewed guild apply.
7. Seed at least two disposable duplicates in one target plus same-guild rows in another target and another-guild rows in the selected target.
8. Confirm one active owner and retain the backup until rollback review is complete.

## Parity Tests

Focused coverage locks metadata, UI, public visibility, permission/owner/deadline behavior, usage ownership, target order, service validation, all collection mappings, duplicate cleanup, guild/target isolation, app wiring, config defaults, command sync, and preflight. Run:

```bash
go test ./internal/core/domain ./internal/core/services/moderation ./internal/adapters/mongo/repositories ./internal/adapters/discordgo ./internal/discord/features/moderation ./internal/app ./internal/config ./internal/parity ./cmd/mhcat-command-sync ./cmd/mhcat-staging-preflight
go test -race ./internal/core/services/moderation ./internal/adapters/mongo/repositories ./internal/adapters/discordgo ./internal/discord/features/moderation ./internal/app
go vet ./internal/core/services/moderation ./internal/adapters/mongo/repositories ./internal/adapters/discordgo ./internal/discord/features/moderation ./internal/app
go run ./tools/parity-audit
```

The real-Mongo duplicate/isolation test is opt-in and must use a disposable database:

```bash
MHCAT_RUN_MONGO_INTEGRATION_TESTS=true \
MHCAT_MONGODB_URI='<disposable-uri>' \
go test ./internal/adapters/mongo/repositories -run '^TestDeleteDataMongoIntegrationTargetsAndDuplicateCleanup$' -count=1
```

## Staging Smoke

1. Complete backup, audit, owner shutdown, preflight, command-sync dry-run, reviewed guild apply, and runtime startup.
2. Confirm the command is publicly discoverable and a user without Manage Messages receives the exact public red denial.
3. Confirm the exact public random-color warning follow-up, duplicated text, footer/image, target order/descriptions/emojis, and one slash usage event.
4. Select a seeded duplicate target as the owner; confirm public success, every selected-guild duplicate is gone, another guild and another target remain, and no index was created.
5. Repeat the missing target and confirm exact public missing content.
6. Have another moderator try the owner's panel and confirm no deletion; then verify the owner can still use a fresh panel if their permission is removed after creation.
7. Verify just-before and at/after one-hour behavior with disposable panels if staging timing permits.
8. Select stats on a case-sensitive deployment and confirm `numbers` deletion repairs the legacy filename defect.
9. Disable gates, remove only the managed staging command, preserve remaining rows, and perform rollback checks.

## Rollback

1. Disable command-sync inclusion and remove only the managed staging `刪除資料` command.
2. Disable the runtime gate and stop the Go owner before restoring Node.
3. Preserve all remaining target collections and indexes. Do not repair or deduplicate during emergency rollback.
4. Restore deleted disposable rows from the pre-smoke backup if continued testing requires them. Disabling Go cannot recreate deleted settings.
5. Restore the Node owner only after confirming no Go route remains.
6. Recheck command denial/prompt plus one disposable non-stats target. Remember that Node's stats branch remains broken on case-sensitive filesystems.
7. Review any overlap interval and Mongo audit logs for unintended destructive selections.

Production ownership remains blocked on live staging smoke, verified backups, exclusive ownership, and reviewed audits for every target collection. No unique index or data rewrite is required merely to enable Go.
