# Stats Commands Slice

Implemented slices: `/統計系統查詢`, `/統計系統創建`, `/統計身分組人數`, and `/統計系統刪除`.

Legacy evidence:
- `MHCAT/slashCommands/統計系統/number.js` defines command name `統計系統查詢`, description `查詢統計消息`, cooldown metadata `10`, and a public `interaction.reply` with a single static embed.
- The static embed title is `統計系統查詢` and the description explains the 10-minute update cadence plus available user/channel counters.
- `MHCAT/slashCommands/統計系統/number_delete.js` defines command name `統計系統刪除`, description `刪除統計消息`, runtime Manage Messages check, and a deferred public edit that deletes the guild `Number` row.
- The delete success title is `<a:greentick:980496858445135893> | 成功刪除，該類別以下的頻道我已經管不了囉!(類別id:<parent>)`; missing config returns `你還沒有創建過統計數據，是要刪除甚麼啦!`.
- `MHCAT/slashCommands/統計系統/role_create.js` defines command name `統計身分組人數`, description `統計某個特定的身分組的人數`, runtime Manage Messages check, required channel-type string choice, and required role option.
- The role-count path requires an existing `Number` row, creates a text or voice channel named `<role name>: <member count>`, and writes `role_numbers` fields `guild`, `channel`, `channel_name`, and `role`.

Go behavior:
- Runtime route is available only when `MHCAT_FEATURE_STATS_QUERY_ENABLED=true`.
- Command sync includes the command only when `MHCAT_COMMAND_SYNC_INCLUDE_STATS_QUERY=true`.
- `/統計系統創建` is available only when `MHCAT_FEATURE_STATS_CREATE_ENABLED=true`.
- Command sync includes the create command only when `MHCAT_COMMAND_SYNC_INCLUDE_STATS_CREATE=true`.
- `/統計身分組人數` is available only when `MHCAT_FEATURE_STATS_ROLE_COUNT_ENABLED=true`.
- Command sync includes the role-count command only when `MHCAT_COMMAND_SYNC_INCLUDE_STATS_ROLE_COUNT=true`.
- `/統計系統刪除` is available only when `MHCAT_FEATURE_STATS_DELETE_ENABLED=true`.
- Command sync includes the delete command only when `MHCAT_COMMAND_SYNC_INCLUDE_STATS_DELETE=true`.
- Staging preflight and staging scripts reject unpaired command-sync/runtime flags.
- The handler replies once through the responder and disables allowed mentions.
- The handler preserves the legacy random-color embed behavior, with deterministic color injection in tests.
- The role-count handler creates the legacy role stat channel, stores `channel_name` as the current member count string, and replaces duplicate `{guild,role}` `role_numbers` rows.
- The delete handler deletes legacy `numbers` rows for the guild, returning the first row's `parent` in the legacy success title. It does not delete Discord channels, matching the legacy command's actual behavior.
- Usage tracking remains the no-op production tracker unless a future usage repository is approved.

Implemented separately in the create slice:
- `統計系統創建` creates the legacy category/base stat channels, writes `numbers`, and can add channel-count/text-count/voice-count stat channels.

Intentionally not implemented:
- Channel rename side effects.
- `handler/channel_status.js` 20-minute rename worker.
- Any Mongo index creation.

Known legacy issues for future slice:
- Legacy count logic uses old channel type strings in a discord.js v14 codebase.
- The legacy rename worker can run per shard/process without a distributed lease.
- Role-count creation can orphan old channels when rerun; Go preserves the replacement row behavior and does not delete old Discord channels.
- Future rename implementation needs explicit Discord rate-limit/debounce and scheduler ownership design before enabling recurring channel mutations.
