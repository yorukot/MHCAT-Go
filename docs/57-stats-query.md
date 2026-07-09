# Stats Query Slice

Implemented slice: `/統計系統查詢`.

Legacy evidence:
- `MHCAT/slashCommands/統計系統/number.js` defines command name `統計系統查詢`, description `查詢統計消息`, cooldown metadata `10`, and a public `interaction.reply` with a single static embed.
- The static embed title is `統計系統查詢` and the description explains the 10-minute update cadence plus available user/channel counters.

Go behavior:
- Runtime route is available only when `MHCAT_FEATURE_STATS_QUERY_ENABLED=true`.
- Command sync includes the command only when `MHCAT_COMMAND_SYNC_INCLUDE_STATS_QUERY=true`.
- Staging preflight and staging scripts reject unpaired command-sync/runtime flags.
- The handler replies once through the responder and disables allowed mentions.
- The handler preserves the legacy random-color embed behavior, with deterministic color injection in tests.
- Usage tracking remains the no-op production tracker unless a future usage repository is approved.

Intentionally not implemented:
- `統計系統創建`.
- `統計身分組人數`.
- `統計系統刪除`.
- `Number` / `role_number` Mongo writes.
- Channel create/delete/rename side effects.
- `handler/channel_status.js` 20-minute rename worker.
- Any Mongo index creation.

Known legacy issues for future slice:
- Legacy count logic uses old channel type strings in a discord.js v14 codebase.
- The legacy rename worker can run per shard/process without a distributed lease.
- Role-count creation can orphan old channels when rerun.
- Future implementation needs explicit Discord rate-limit/debounce and scheduler ownership design before enabling channel mutations.
