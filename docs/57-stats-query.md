# Stats Commands

Status: historical implementation overview. Exact compatibility, ownership, staging, and rollback are canonical in [93-stats.md](93-stats.md).

Covered legacy paths:
- `slashCommands/統計系統/number.js`
- `slashCommands/統計系統/number_create.js`
- `slashCommands/統計系統/role_create.js`
- `slashCommands/統計系統/number_delete.js`
- `handler/channel_status.js`

Preserved behavior:
- `/統計系統查詢` sends the exact public static embed, including whitespace, wording, and random color.
- `/統計系統創建` preserves permission checks, text/voice names, category name, overwrite order and bitfields, error/success embeds, first-create loading follow-up, optional-create direct success follow-up, and cache-only parent selection.
- Base counts preserve `guild.memberCount`, cached bot count, and derived non-bot count. Channel counts preserve the discord.js v14 stale-string bug: total includes cached categories, while text and voice counts are both zero.
- Initial `numbers` writes preserve all legacy field names and 38 explicit BSON `null` fields. Optional voice creation displays the voice count but stores the legacy `textnumber`-derived value in `voicenumber_name`.
- `/統計身分組人數` sends and edits the legacy loading follow-up, uses cached role-member counts (including `@everyone`), creates the exact channel names/overwrites, and writes `role_numbers` fields `guild`, `channel`, `channel_name`, and `role`.
- `/統計系統刪除` preserves the public defer/edit payloads, returns the first matched row's parent ID, and does not delete Discord channels.
- The rename worker waits 20 minutes before its first run, repeats every 20 minutes, uses cached guild/member/channel state, applies the legacy first-match string replacement, and renames deleted-role counters to zero.
- Global slash usage middleware remains the only `all_use_counts` writer when `MHCAT_FEATURE_USAGE_TRACKING_ENABLED=true`.

Intentional reliability differences:
- Go awaits channel and Mongo writes and returns/logs safe failures; legacy frequently launched unawaited writes and could report success before persistence.
- Initial stats creation uses an upsert, optional counter fields update together, role replacement removes duplicate `{guild,role}` rows, and delete removes all duplicate `{guild}` rows.
- The rename worker serializes runs, batches each base row's counter update, skips missing/API-failed channels without advancing stored counters, avoids redundant same-name PATCH requests, and continues after malformed or missing role rows. Legacy could overlap runs, partially advance counters, or stop the remaining role scan early.
- Go suppresses allowed mentions explicitly and requires the interaction application ID for the bot member overwrite.
- No Mongo index is created, and no distributed worker lease is added. Do not run Node and Go rename ownership for the same guilds.

Runtime gates:
- Query: `MHCAT_FEATURE_STATS_QUERY_ENABLED=true` plus `MHCAT_COMMAND_SYNC_INCLUDE_STATS_QUERY=true` for sync.
- Create: `MHCAT_FEATURE_STATS_CREATE_ENABLED=true`, gateway, Guild Members intent, plus `MHCAT_COMMAND_SYNC_INCLUDE_STATS_CREATE=true` for sync.
- Role count: `MHCAT_FEATURE_STATS_ROLE_COUNT_ENABLED=true`, gateway, Guild Members intent, plus `MHCAT_COMMAND_SYNC_INCLUDE_STATS_ROLE_COUNT=true` for sync.
- Delete: `MHCAT_FEATURE_STATS_DELETE_ENABLED=true` plus `MHCAT_COMMAND_SYNC_INCLUDE_STATS_DELETE=true` for sync.
- Rename worker: `MHCAT_FEATURE_STATS_RENAME_WORKER_ENABLED=true`, gateway, and Guild Members intent; it has no command-sync flag.
