# Operational Runbook

Status: Wave 5.3 staging smoke guardrails plus gated feature slices. `cmd/mhcat-bot` can optionally open the Discord Gateway and route `InteractionCreate` events for implemented commands whose runtime flags are enabled. Gateway remains disabled by default. `cmd/mhcat-command-sync` exists for dry-run command diff planning and staging-guild-only apply operations. `cmd/mhcat-mongo-audit` is a read-only Mongo audit CLI. `cmd/mhcat-mongo-index` is a dry-run/default index diff CLI with explicit apply guardrails.

## How to Configure

Primary Go env vars:

- `MHCAT_DISCORD_TOKEN`
- `MHCAT_MONGODB_URI`
- `MHCAT_MONGODB_DATABASE`
- `MHCAT_ENV`
- `MHCAT_LOG_LEVEL`
- `MHCAT_LOG_FORMAT`
- `MHCAT_DISCORD_ENABLE_GATEWAY`
- `MHCAT_DISCORD_GATEWAY_CONNECT_TIMEOUT`
- `MHCAT_DISCORD_INTERACTION_TIMEOUT`
- `MHCAT_DISCORD_GATEWAY_SMOKE_TEST`
- `MHCAT_DISCORD_GATEWAY_SMOKE_TEST_TIMEOUT`
- `MHCAT_DISCORD_MESSAGE_CONTENT_INTENT`
- `MHCAT_DISCORD_GUILD_MEMBERS_INTENT`
- `MHCAT_DISCORD_GUILD_MESSAGE_REACTIONS_INTENT`
- `MHCAT_FEATURE_TICKETS_ENABLED`
- `MHCAT_FEATURE_POLLS_ENABLED`
- `MHCAT_FEATURE_USAGE_TRACKING_ENABLED`
- `MHCAT_FEATURE_ECONOMY_QUERY_ENABLED`
- `MHCAT_FEATURE_ECONOMY_SIGNIN_ENABLED`
- `MHCAT_FEATURE_ECONOMY_SETTINGS_ENABLED`
- `MHCAT_FEATURE_ECONOMY_COIN_ADMIN_ENABLED`
- `MHCAT_FEATURE_ECONOMY_COIN_RANK_ENABLED`
- `MHCAT_FEATURE_ECONOMY_COIN_RESET_ENABLED`
- `MHCAT_FEATURE_ECONOMY_RPS_ENABLED`
- `MHCAT_FEATURE_ECONOMY_GAME_ENABLED`
- `MHCAT_FEATURE_ECONOMY_SHOP_ENABLED`
- `MHCAT_FEATURE_ECONOMY_PROFILE_ENABLED`
- `MHCAT_FEATURE_WORK_ENABLED`
- `MHCAT_FEATURE_WARNINGS_ENABLED`
- `MHCAT_FEATURE_WARNING_SETTINGS_ENABLED`
- `MHCAT_FEATURE_WARNING_REMOVAL_ENABLED`
- `MHCAT_FEATURE_WARNING_ISSUE_ENABLED`
- `MHCAT_FEATURE_MESSAGE_CLEANUP_ENABLED`
- `MHCAT_FEATURE_DELETE_DATA_ENABLED`
- `MHCAT_FEATURE_TRANSLATE_ENABLED`
- `MHCAT_FEATURE_REDEEM_ENABLED`
- `MHCAT_FEATURE_AUTOCHAT_CONFIG_ENABLED`
- `MHCAT_FEATURE_AUTOCHAT_FALLBACK_ENABLED`
- `MHCAT_FEATURE_AUTOCHAT_PAID_HANDOFF_ENABLED`
- `MHCAT_AUTOCHAT_PAID_OWNERSHIP_CONFIRMED`
- `MHCAT_FEATURE_AUTO_NOTIFICATION_CONFIG_ENABLED`
- `MHCAT_FEATURE_AUTO_NOTIFICATION_DELIVERY_ENABLED`
- `MHCAT_FEATURE_DAILY_RESET_SCHEDULER_ENABLED`
- `MHCAT_FEATURE_WORK_PAYOUT_SCHEDULER_ENABLED`
- `MHCAT_FEATURE_LOGGING_CONFIG_ENABLED`
- `MHCAT_FEATURE_LOGGING_MESSAGE_EVENTS_ENABLED`
- `MHCAT_FEATURE_LOGGING_CHANNEL_EVENTS_ENABLED`
- `MHCAT_FEATURE_LOGGING_VOICE_EVENTS_ENABLED`
- `MHCAT_FEATURE_GACHA_PRIZE_LIST_ENABLED`
- `MHCAT_FEATURE_GACHA_DRAW_ENABLED`
- `MHCAT_FEATURE_GACHA_PRIZE_CREATE_ENABLED`
- `MHCAT_FEATURE_GACHA_PRIZE_EDIT_ENABLED`
- `MHCAT_FEATURE_GACHA_PRIZE_DELETE_ENABLED`
- `MHCAT_FEATURE_LOTTERY_DISABLED_COMMAND_ENABLED`
- `MHCAT_FEATURE_STATS_QUERY_ENABLED`
- `MHCAT_FEATURE_STATS_CREATE_ENABLED`
- `MHCAT_FEATURE_STATS_ROLE_COUNT_ENABLED`
- `MHCAT_FEATURE_STATS_DELETE_ENABLED`
- `MHCAT_FEATURE_STATS_RENAME_WORKER_ENABLED`
- `MHCAT_FEATURE_ANNOUNCEMENT_CONFIG_ENABLED`
- `MHCAT_FEATURE_ANNOUNCEMENT_SEND_ENABLED`
- `MHCAT_FEATURE_ANNOUNCEMENT_RELAY_ENABLED`
- `MHCAT_FEATURE_ANTI_SCAM_CONFIG_ENABLED`
- `MHCAT_FEATURE_ANTI_SCAM_REPORT_ENABLED`
- `MHCAT_FEATURE_ANTI_SCAM_MESSAGE_DELETE_ENABLED`
- `MHCAT_FEATURE_BIRTHDAY_CONFIG_ENABLED`
- `MHCAT_FEATURE_TEXT_XP_CONFIG_ENABLED`
- `MHCAT_FEATURE_TEXT_XP_ACCRUAL_ENABLED`
- `MHCAT_FEATURE_VOICE_XP_CONFIG_ENABLED`
- `MHCAT_FEATURE_VOICE_XP_SESSIONS_ENABLED`
- `MHCAT_FEATURE_XP_ROLE_CONFIG_ENABLED`
- `MHCAT_FEATURE_VOICE_ROOM_CONFIG_ENABLED`
- `MHCAT_FEATURE_VOICE_ROOM_LOCK_ENABLED`
- `MHCAT_FEATURE_XP_PROFILE_DISABLED_COMMANDS_ENABLED`
- `MHCAT_FEATURE_XP_ADMIN_ENABLED`
- `MHCAT_FEATURE_XP_RESET_ENABLED`
- `MHCAT_FEATURE_XP_RANK_ENABLED`
- `MHCAT_FEATURE_JOIN_ROLE_CONFIG_ENABLED`
- `MHCAT_FEATURE_JOIN_ROLE_ASSIGNMENT_ENABLED`
- `MHCAT_FEATURE_WELCOME_MESSAGE_CONFIG_ENABLED`
- `MHCAT_FEATURE_WELCOME_MESSAGE_DELIVERY_ENABLED`
- `MHCAT_FEATURE_LEAVE_MESSAGE_DELIVERY_ENABLED`
- `MHCAT_FEATURE_VERIFICATION_CONFIG_ENABLED`
- `MHCAT_FEATURE_VERIFICATION_FLOW_ENABLED`
- `MHCAT_FEATURE_ACCOUNT_AGE_CONFIG_ENABLED`
- `MHCAT_FEATURE_ACCOUNT_AGE_POLICY_ENABLED`
- `MHCAT_FEATURE_ROLE_SELECTION_ENABLED`
- `MHCAT_STAGING_MODE`
- `MHCAT_STAGING_GUILD_ID`
- `MHCAT_STAGING_ALLOWED_APPLICATION_ID`
- `MHCAT_STAGING_REQUIRE_GUILD_SCOPE`
- `MHCAT_STAGING_ALLOW_COMMAND_APPLY`
- `MHCAT_STAGING_ALLOW_GATEWAY_SMOKE`
- `MHCAT_STAGING_SMOKE_TIMEOUT`
- `MHCAT_STAGING_EXPECTED_COMMANDS`

Command registration is not a bot startup mode in production. Use `mhcat-command-sync` as a separate operational step; it defaults to dry-run and requires explicit apply flags before any Discord command write.

Legacy aliases to support:

- `TOKEN` -> `DISCORD_TOKEN`
- `MONGOOSE_CONNECTION_STRING` -> `MONGODB_URI`
- `JOIN_WEBHOOK` -> `JOIN_WEBHOOK_URL`
- `LEAVE_WEBHOOK` -> `LEAVE_WEBHOOK_URL`
- `READY_WEBHOOK` -> `READY_WEBHOOK_URL`
- `REPORT_WEBHOOK` -> `REPORT_WEBHOOK_URL`

### Slash Usage Tracking

`MHCAT_FEATURE_USAGE_TRACKING_ENABLED=false` is the default and uses the no-op tracker. There is no command-sync companion flag. When enabled, the global interaction middleware writes one best-effort atomic increment to `all_use_counts` for every slash attempt before route lookup, permission checks, and command handling. Components, modals, and autocomplete interactions do not increment the counter. Tracking errors are ignored by the command path, and each write is bounded to 500 ms.

Before enabling the gate, stop the Node `events/SlashCommands.js` usage-counter owner for the same application/guilds and confirm the target database is disposable staging data. Run `go run ./cmd/mhcat-staging-preflight --format text`; the expected `usage-tracking-runtime-readiness` warning records that ownership check. Audit duplicate and null/blank `slashcommand_name` rows before any unique index apply. The runtime does not create the candidate index.

Statements in feature-specific sections that a command does not write usage counters describe that feature handler and assume this global gate is disabled. With the gate enabled, invoking any slash command can still update `all_use_counts` independently of the feature's own Mongo behavior.

## How to Run Locally

```bash
go run ./cmd/mhcat-bot
```

Missing required env should fail safely without printing secrets.

Gateway is disabled by default. With valid env and gateway disabled, startup connects/pings Mongo, creates the Discord session object, registers runtime feature modules in memory, and exits cleanly without opening the Discord Gateway.

To run with gateway enabled:

```bash
MHCAT_MONGODB_URI='<uri>' \
MHCAT_MONGODB_DATABASE=mhcat \
MHCAT_DISCORD_TOKEN='<token>' \
MHCAT_DISCORD_ENABLE_GATEWAY=true \
go run ./cmd/mhcat-bot
```

This registers the internal `InteractionCreate` event handler and gateway event handlers for explicitly enabled feature slices. It does not register Discord application commands. Feature Mongo writes happen only for explicit config commands whose runtime flags are enabled.

One-shot smoke mode:

```bash
MHCAT_MONGODB_URI='<uri>' \
MHCAT_MONGODB_DATABASE=mhcat \
MHCAT_DISCORD_TOKEN='<token>' \
MHCAT_DISCORD_ENABLE_GATEWAY=true \
MHCAT_DISCORD_GATEWAY_SMOKE_TEST=true \
go run ./cmd/mhcat-bot
```

Smoke mode waits for gateway ready or timeout and then shuts down without sending messages, registering commands, or writing Mongo.

Wave 5.3 requires staging flags for smoke:

```bash
MHCAT_STAGING_MODE=true \
MHCAT_STAGING_ALLOW_GATEWAY_SMOKE=true \
scripts/staging/gateway-smoke.sh
```

## How to Run Staging

- Use a separate Discord application/token.
- Use a staging guild.
- Use an isolated or sanitized staging MongoDB.
- Keep command sync scope guild-only.
- Pair optional feature command-sync flags with matching runtime flags, for example `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_QUERY=true` requires `MHCAT_FEATURE_ECONOMY_QUERY_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_RANK=true` requires `MHCAT_FEATURE_ECONOMY_COIN_RANK_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_RPS=true` requires `MHCAT_FEATURE_ECONOMY_RPS_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_GAME=true` requires `MHCAT_FEATURE_ECONOMY_GAME_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SHOP=true` requires `MHCAT_FEATURE_ECONOMY_SHOP_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_PROFILE=true` requires `MHCAT_FEATURE_ECONOMY_PROFILE_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SIGNIN=true` requires `MHCAT_FEATURE_ECONOMY_SIGNIN_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SETTINGS=true` requires `MHCAT_FEATURE_ECONOMY_SETTINGS_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_ADMIN=true` requires `MHCAT_FEATURE_ECONOMY_COIN_ADMIN_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_WORK=true` requires `MHCAT_FEATURE_WORK_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_WARNINGS=true` requires `MHCAT_FEATURE_WARNINGS_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_WARNING_SETTINGS=true` requires `MHCAT_FEATURE_WARNING_SETTINGS_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_WARNING_REMOVAL=true` requires `MHCAT_FEATURE_WARNING_REMOVAL_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_WARNING_ISSUE=true` requires `MHCAT_FEATURE_WARNING_ISSUE_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_MESSAGE_CLEANUP=true` requires `MHCAT_FEATURE_MESSAGE_CLEANUP_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_DELETE_DATA=true` requires `MHCAT_FEATURE_DELETE_DATA_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_TRANSLATE=true` requires `MHCAT_FEATURE_TRANSLATE_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_BALANCE_QUERY=true` requires `MHCAT_FEATURE_BALANCE_QUERY_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_REDEEM=true` requires `MHCAT_FEATURE_REDEEM_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_AUTO_NOTIFICATION_CONFIG=true` requires `MHCAT_FEATURE_AUTO_NOTIFICATION_CONFIG_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_LOGGING_CONFIG=true` requires `MHCAT_FEATURE_LOGGING_CONFIG_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_LIST=true` requires `MHCAT_FEATURE_GACHA_PRIZE_LIST_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_DRAW=true` requires `MHCAT_FEATURE_GACHA_DRAW_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_CREATE=true` requires `MHCAT_FEATURE_GACHA_PRIZE_CREATE_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_EDIT=true` requires `MHCAT_FEATURE_GACHA_PRIZE_EDIT_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_DELETE=true` requires `MHCAT_FEATURE_GACHA_PRIZE_DELETE_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_LOTTERY_DISABLED_COMMAND=true` requires `MHCAT_FEATURE_LOTTERY_DISABLED_COMMAND_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_STATS_QUERY=true` requires `MHCAT_FEATURE_STATS_QUERY_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_STATS_CREATE=true` requires `MHCAT_FEATURE_STATS_CREATE_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_STATS_ROLE_COUNT=true` requires `MHCAT_FEATURE_STATS_ROLE_COUNT_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_STATS_DELETE=true` requires `MHCAT_FEATURE_STATS_DELETE_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_CONFIG=true` requires `MHCAT_FEATURE_ANNOUNCEMENT_CONFIG_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_TEXT_XP_CONFIG=true` requires `MHCAT_FEATURE_TEXT_XP_CONFIG_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_VOICE_XP_CONFIG=true` requires `MHCAT_FEATURE_VOICE_XP_CONFIG_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_XP_ROLE_CONFIG=true` requires `MHCAT_FEATURE_XP_ROLE_CONFIG_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_XP_PROFILE_DISABLED_COMMANDS=true` requires `MHCAT_FEATURE_XP_PROFILE_DISABLED_COMMANDS_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_XP_ADMIN=true` requires `MHCAT_FEATURE_XP_ADMIN_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_XP_RESET=true` requires `MHCAT_FEATURE_XP_RESET_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_CONFIG=true` requires `MHCAT_FEATURE_VOICE_ROOM_CONFIG_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_LOCK=true` requires `MHCAT_FEATURE_VOICE_ROOM_LOCK_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_JOIN_ROLE_CONFIG=true` requires `MHCAT_FEATURE_JOIN_ROLE_CONFIG_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_CONFIG=true` requires `MHCAT_FEATURE_VERIFICATION_CONFIG_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_FLOW=true` requires `MHCAT_FEATURE_VERIFICATION_FLOW_ENABLED=true`, and `MHCAT_COMMAND_SYNC_INCLUDE_ROLE_SELECTION=true` requires `MHCAT_FEATURE_ROLE_SELECTION_ENABLED=true`.
- Tickets follow the same pairing rule: `MHCAT_COMMAND_SYNC_INCLUDE_TICKETS=true` requires `MHCAT_FEATURE_TICKETS_ENABLED=true`; command inclusion remains staging-only, and any runtime rollout requires exclusive Node/Go ownership.
- Announcement config and send commands are independently paired: `MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_CONFIG=true` requires `MHCAT_FEATURE_ANNOUNCEMENT_CONFIG_ENABLED=true`, and `MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_SEND=true` requires `MHCAT_FEATURE_ANNOUNCEMENT_SEND_ENABLED=true`. Bound relay is event-only and separately requires `MHCAT_FEATURE_ANNOUNCEMENT_RELAY_ENABLED=true`, Gateway, Guild Messages, and Message Content.
- Before enabling an announcement family, stop or gate its Node command/modal/event owner for the same bot/guild. Audit duplicate and scalar-drift keys in `guilds`/`ann_all_sets`, confirm dashboard/shared writers preserve unrelated `guilds` fields, and do not create a startup index or automatic repair. Follow the [announcement parity contract](76-announcement.md).
- Anti-scam toggle/report commands are independently paired with `MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_CONFIG` and `MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_REPORT`; report also requires a safe `MHCAT_REPORT_WEBHOOK_URL` or legacy `REPORT_WEBHOOK`. Message deletion is event-only and requires its feature gate plus Gateway, Guild Messages, and Message Content.
- Before enabling an anti-scam family, stop/gate the matching Node command or `events/safe_server.js`, audit both collections and external catalog writers, and do not normalize URLs or create indexes. Follow the [anti-scam parity contract](77-anti-scam.md).
- Economy coin reset has an additional message confirmation requirement: `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_RESET=true` requires `MHCAT_FEATURE_ECONOMY_COIN_RESET_ENABLED=true`, `MHCAT_DISCORD_ENABLE_GATEWAY=true`, `MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true`, and `MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true`.
- Role selection has an additional reaction-event requirement: `MHCAT_COMMAND_SYNC_INCLUDE_ROLE_SELECTION=true` requires `MHCAT_FEATURE_ROLE_SELECTION_ENABLED=true`, `MHCAT_DISCORD_ENABLE_GATEWAY=true`, and `MHCAT_DISCORD_GUILD_MESSAGE_REACTIONS_INTENT=true`.
- XP rank follows the same staging pairing rule: `MHCAT_COMMAND_SYNC_INCLUDE_XP_RANK=true` requires `MHCAT_FEATURE_XP_RANK_ENABLED=true`.
- Run `go run ./cmd/mhcat-staging-preflight --format text`.
- Run `scripts/staging/command-sync-dry-run.sh`.
- Review the plan.
- Optionally run `scripts/staging/command-sync-apply-guild.sh` only with `MHCAT_STAGING_MODE=true` and `MHCAT_STAGING_ALLOW_COMMAND_APPLY=true`.
- Run `scripts/staging/gateway-smoke.sh` only with `MHCAT_STAGING_ALLOW_GATEWAY_SMOKE=true`.
- Keep scheduler and high-risk feature groups disabled.

## How to Run Production

Pending Wave 8. Production must not start until:

- command registration mode is explicitly chosen
- Mongo indexes/audits have been reviewed
- rollback to Node.js has been rehearsed
- feature ownership is exclusive between Node and Go

## How to Shard

- Use DiscordGo shard ID/count identify settings.
- Guild-scoped work must run on the owning shard.
- Global background jobs must run in a single scheduler process or under a Mongo-backed lease.
- Do not copy discord.js `ShardingManager` assumptions directly.

## How to Check Health

Planned checks:

- process started
- config loaded and redacted
- Mongo ping succeeds
- Discord gateway connected when explicitly enabled
- command registration state known
- scheduler leadership state known

## How to Check MongoDB Connectivity

Available through bot startup ping or the audit/index tools:

```bash
MHCAT_MONGODB_URI='<uri>' MHCAT_MONGODB_DATABASE=mhcat go run ./cmd/mhcat-mongo-audit --sample-limit 0
```

## How to Bootstrap Indexes

Dry-run index comparison:

```bash
MHCAT_MONGODB_URI='<uri>' MHCAT_MONGODB_DATABASE=mhcat go run ./cmd/mhcat-mongo-index --dry-run --format json
```

Rules:

- Dry-run is default.
- Unique/high-risk indexes require duplicate audit first.
- Production startup does not create high-risk indexes.
- Wave 3 never drops or modifies existing indexes.
- Apply mode is not part of `make check` and must not be run against production without an approved dry-run review.

## How to Run Data Audit

Available now:

```bash
MHCAT_MONGODB_URI='<uri>' MHCAT_MONGODB_DATABASE=mhcat go run ./cmd/mhcat-mongo-audit --format json
```

Temporary legacy helper retained:

```bash
MONGODB_URI='<uri>' MONGODB_DATABASE='<db>' node MHCAT-REFACTOR/tools/mongo-audit-readonly.mjs
```

or with legacy alias:

```bash
MONGOOSE_CONNECTION_STRING='<uri>' MONGODB_DATABASE='<db>' node MHCAT-REFACTOR/tools/mongo-audit-readonly.mjs
```

The helper is read-only and prints a JSON report with redacted URI credentials. Do not paste raw sensitive user content or secrets from audit output into docs.

Planned future repair/backfill tooling:

```bash
mhcat-tools mongo audit
mhcat-tools data validate --collection <name>
mhcat-tools data audit-types --collection <name>
mhcat-tools data audit-duplicates --collection <name> --keys <keys>
```

## How to Run Data Repair / Backfill

Planned:

```bash
mhcat-tools data repair --name <repair> --dry-run
mhcat-tools data repair --name <repair> --apply
mhcat-tools data backfill --name <backfill> --dry-run
mhcat-tools data backfill --name <backfill> --apply
```

Never run apply without backup/restore point and reviewed dry-run output.

## How to Sync Commands

Available now as Wave 2 infrastructure:

```bash
MHCAT_DISCORD_TOKEN='<token>' \
MHCAT_DISCORD_APPLICATION_ID='<application-id>' \
MHCAT_COMMAND_SYNC_SCOPE=guild \
MHCAT_COMMAND_SYNC_GUILD_ID='<guild-id>' \
go run ./cmd/mhcat-command-sync --dry-run
```

Apply mode:

```bash
go run ./cmd/mhcat-command-sync --apply
```

Deletion requires both `--apply` and `--allow-delete`. Bulk overwrite is not a default path and requires both `--apply` and `--allow-bulk-overwrite`.

Bot startup must not mutate global or guild commands.

Production command registration must be diff-based and must not run independently on every shard.

Wave 5.2 ships local definitions and runtime handlers for `help`, `ping`, and `info` with the `bot` subcommand. Dry-run output may show create/update operations for only those commands and skipped unknown remote commands for the remaining legacy set. Full legacy command definitions are later feature-parity waves.

The current built-in `help`, `ping`, and all four `info` subcommands are parity-audited in [94-utility-builtins.md](94-utility-builtins.md). They require no utility Mongo collection or migration. Linux `/info bot` samples host CPU for one second; `/info shard` initially has no fields and its refresh reads local process metrics immediately. App wiring currently supports one Discord session/shard. Do not raise shard count without explicit cross-process count/metric aggregation and staging.

Parity-audited `/私人頻道設置` and `/私人頻道刪除` are available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_TICKETS=true
MHCAT_FEATURE_TICKETS_ENABLED=true
```

This one runtime gate owns both commands, ticket-shaped `nal` submissions, and exact `tic`/`del` buttons. Stop corresponding Node handlers and extra Go owners, audit `tickets` duplicates/malformed IDs, and do not create `tickets_guild` during enablement. Follow the [ticket parity contract](74-ticket.md) for exact smoke and rollback steps.

Parity-audited `/投票創建` and all poll components are available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_POLLS=true
MHCAT_FEATURE_POLLS_ENABLED=true
```

This one runtime gate owns the slash command, every `mhcat:v1:poll:*` component, and all legacy `poll_<choice>`, `see_result`, `poll_menu`, and `menu_choose` routes. Stop `slashCommands/管理系統/poll.js`, `events/poll.js`, and every extra Go poll owner for the same bot/guild before enablement; pairing the flags does not make shared ownership safe. Exact non-bot participation totals and export member names also require the application Guild Members privileged intent plus `MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true`; poll routing itself requires no Message Content intent. Preserve `polls`, create no startup index, and follow the [poll parity contract](75-poll.md) for audit, staging smoke, versioned-message rollback, and ownership transfer.

Read-only `/代幣查詢` is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_QUERY=true
MHCAT_FEATURE_ECONOMY_QUERY_ENABLED=true
```

This command reads `coins` and `gift_changes` only. It does not write economy data, usage counters, or indexes.

Read-only `/代幣排行榜` is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_RANK=true
MHCAT_FEATURE_ECONOMY_COIN_RANK_ENABLED=true
```

This command reads `coins` only and renders a PNG leaderboard with legacy pagination buttons. It does not write economy data, usage counters, or indexes.

Read-only `my-profile` is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_PROFILE=true
MHCAT_FEATURE_ECONOMY_PROFILE_ENABLED=true
```

This command reads `coins`, `gift_changes`, `text_xps`, `voice_xps`, `work_sets`, and `work_users`, then renders the legacy-style `user-info.png` profile card with a refresh button. It does not write economy data, usage counters, or indexes.

`/coin-related-settings` is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SETTINGS=true
MHCAT_FEATURE_ECONOMY_SETTINGS_ENABLED=true
```

This command writes `gift_changes` with legacy field names and requires Manage Messages. It updates all duplicate rows for the guild so old `findOne` consumers remain rollback-compatible until duplicate audit and unique-index work are complete. Do not enable it in production until shared gacha/sign-in/XP consumers and duplicate audits are reviewed.

`/代幣增加` is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_ADMIN=true
MHCAT_FEATURE_ECONOMY_COIN_ADMIN_ENABLED=true
```

This command requires Manage Messages, writes `coins`, creates missing balances only for add operations, preserves signed amounts and Mongoose-visible numeric scalars with add-only upper and reduce-only lower guards, and updates one arbitrary duplicate `{guild,member}` row. Test only against disposable staging data until duplicate audits and production ownership are reviewed.

`/剪刀石頭布` is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_RPS=true
MHCAT_FEATURE_ECONOMY_RPS_ENABLED=true
```

This command reads and updates one arbitrary existing `coins` row, preserves decimal/null/infinite Mongoose number behavior, rejects missing or insufficient balances, subtracts half the wager on ties using legacy integer flooring, and does not cap post-win balances at `999999999`. Test only against disposable staging balances until duplicate/scalar audits and economy ownership are reviewed.

`/代幣遊戲` is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_GAME=true
MHCAT_FEATURE_ECONOMY_GAME_ENABLED=true
```

This command writes existing `coins` rows for two-player wagers in `21點`, `知識王`, and `比大小`, and uses process-local component/transition session state. Reserve and settlement update both players in one Mongo transaction, so the deployment must use a replica set or sharded cluster; standalone Mongo rejects game writes. The runtime preserves the 500-millisecond knowledge start, component-free five-second knowledge reveals, and five-second higher/lower draw screen; balances remain reserved until delayed settlement completes. Knowledge and blackjack settle legacy forfeits on their strict 21/31-second timeout ticks, remove components, and cancel all pending callbacks during graceful shutdown. Test only against disposable staging balances, do not blindly retry an unknown transaction commit result, and do not production-sync until duplicate audits, restart reconciliation, and economy ownership are reviewed. See `docs/67-economy-game-lifecycle.md`.

`/代幣商店` is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SHOP=true
MHCAT_FEATURE_ECONOMY_SHOP_ENABLED=true
```

This command writes `ghps` shop rows, subtracts `coins`, can add roles, and can DM prize codes. Test only against disposable staging `ghps` and `coins` rows with roles below the bot's highest role; purchase inventory and coin updates are not transactional.

`/代幣重製` is available only when staging command sync, runtime, gateway, Guild Messages, and Message Content flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_RESET=true
MHCAT_FEATURE_ECONOMY_COIN_RESET_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true
MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true
```

This command requires the Discord guild owner. It sends the legacy destructive warning and requires the same owner to type `^確認^` in the same channel within 60 seconds before mutating every staging guild `coins` row. Omit `除以多少` or pass `0` to delete rows; pass a nonzero divisor to divide balances with legacy rounding. Test only against disposable staging balances; command sync, config validation, and staging scripts reject unpaired sync/runtime/gateway flags.

Bot startup still does not sync commands. Run command sync manually and review dry-run output before any `--apply`.

The warning command families are available only when their staging command-sync and runtime flags are paired:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_WARNINGS=true
MHCAT_FEATURE_WARNINGS_ENABLED=true
MHCAT_COMMAND_SYNC_INCLUDE_WARNING_SETTINGS=true
MHCAT_FEATURE_WARNING_SETTINGS_ENABLED=true
MHCAT_COMMAND_SYNC_INCLUDE_WARNING_REMOVAL=true
MHCAT_FEATURE_WARNING_REMOVAL_ENABLED=true
MHCAT_COMMAND_SYNC_INCLUDE_WARNING_ISSUE=true
MHCAT_FEATURE_WARNING_ISSUE_ENABLED=true
```

Enable only the families under test. `/警告紀錄` reads `warndbs` and intentionally does not enforce its advertised Manage Messages metadata. Settings writes `errors_sets`; removal mutates `warndbs`; issue appends warnings, sends best-effort DMs, and may kick or ban disposable members. When global usage tracking is separately enabled, each slash attempt records exactly one event. No warning gate creates indexes or runs startup repair/migration. Back up and audit duplicate/mixed rows, unknown actions, and malformed thresholds before smoke. Exact UI, permission, scalar, threshold, duplicate, ownership, smoke, and rollback requirements are in [84-warning-system.md](84-warning-system.md).

Destructive `/刪除訊息` is available only with paired staging command-sync and runtime flags:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_MESSAGE_CLEANUP=true
MHCAT_FEATURE_MESSAGE_CLEANUP_ENABLED=true
```

Use only a disposable staging channel and one active Node/Go owner. The command is publicly discoverable, defers ephemerally, requires Manage Messages, adds Administrator above 200, and refuses more than 1000. Go waits for sequential Discord batches, reports actual confirmed deletes, scans target-user pages with a cursor, and retains every message older than 14 days. Earlier batches cannot roll back if a later API call fails. It requires no Message Content intent, Mongo repository, index, or database migration. Follow [85-message-cleanup.md](85-message-cleanup.md) before enabling or retrying a partial failure.

Destructive `/刪除資料` is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_DELETE_DATA=true
MHCAT_FEATURE_DELETE_DATA_ENABLED=true
```

This command requires Manage Messages to create the legacy owner-scoped one-hour prompt and deletes all duplicate guild rows from only the selected legacy config target. Back up and use disposable staging rows only. Follow the exact ownership, isolation, smoke, and rollback contract in [83-delete-data.md](83-delete-data.md).

`/翻譯` is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_TRANSLATE=true
MHCAT_FEATURE_TRANSLATE_ENABLED=true
```

This command calls the external translate provider through a driver-agnostic port. It does not require Message Content intent or Mongo feature data. Go preserves the public loading follow-up, uses a 10-second provider budget inside a 15-second interaction floor, and edits the same follow-up to a safe red error instead of leaving it stuck. Follow [86-translate.md](86-translate.md) for exact UI, live provider smoke, ownership, and rollback.

`/查看餘額` is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_BALANCE_QUERY=true
MHCAT_FEATURE_BALANCE_QUERY_ENABLED=true
```

This command reads one arbitrary matching `chatgpt_gets.price`, preserves Mongoose number display and exact ephemeral UI, and returns a controlled red error on backend failure. It does not enable auto-chat, require Message Content, mutate balances, or create indexes. Audit shared writers and follow [87-balance-query.md](87-balance-query.md) for exact staging and rollback.

`/兌換` is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_REDEEM=true
MHCAT_FEATURE_REDEEM_ENABLED=true
```

This command reads an exact raw `codes.code`, deletes only the fetched row, then replaces one arbitrary matching `chatgpt_gets` row by delete+insert. It preserves Mongoose number coercion, the legacy 7-day `>` expiry check, duplicate rows, and exact ephemeral UI. The flow is non-transactional: snapshot both collections and use disposable staging fixtures because an error can occur after code or balance deletion. It does not enable auto-chat, require Message Content intent, or create indexes. Follow [88-redeem.md](88-redeem.md).

Auto-chat config commands are available only with paired staging flags:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_AUTOCHAT_CONFIG=true
MHCAT_FEATURE_AUTOCHAT_CONFIG_ENABLED=true
```

`/自動聊天頻道` and `/自動聊天頻道刪除` remain publicly discoverable but require Manage Messages at runtime. Set replaces one arbitrary fetched `chats` row by delete+insert; delete removes one fetched row. Both preserve duplicate rows and Mongoose String channel coercion, create no index, and are non-transactional. Snapshot/audit `chats`, stop Node command ownership, and use disposable channels. Follow [89-autochat-config.md](89-autochat-config.md). These flags do not enable either MessageCreate runtime.

The event-only local auto-chat fallback is available only with all message runtime prerequisites enabled:

```bash
MHCAT_FEATURE_AUTOCHAT_FALLBACK_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true
MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true
```

This gate registers no slash commands and performs no Mongo writes. It reads balance before config, using one arbitrary `chatgpt_gets` and `chats` first match. Missing, negative, malformed, NaN, undefined, and negative-infinity balances use the exact legacy corpus; null/empty, zero, positive, and positive-infinity balances stay silent. Replies preserve `說出`, UTF-16 search, typing, and `[1s,5s)` timing while suppressing all mentions. Stop Node event ownership and follow [90-autochat-fallback.md](90-autochat-fallback.md).

The event-only paid auto-chat handoff requires an additional ownership acknowledgment:

```bash
MHCAT_FEATURE_AUTOCHAT_PAID_HANDOFF_ENABLED=true
MHCAT_AUTOCHAT_PAID_OWNERSHIP_CONFIRMED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true
MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true
```

This path writes `chatgpts`, debits `chatgpt_gets.price`, and uses a Mongo transaction so request publication and charging commit together. Before setting the ownership acknowledgment, confirm a replica-set/sharded Mongo deployment, clean singleton duplicate audits, the compatible external worker, and that Node `events/Chatbot.js` is stopped for the target guilds. The bot-side handoff preserves legacy pricing/overdraw, the ten-second guard/read, 40-second conversation reset, exact handoff fields, and input/output mention warnings. Follow [91-autochat-paid.md](91-autochat-paid.md) for the exact worker scalar matrix, transaction failure handling, staging smoke, reconciliation, and rollback.

Auto-notification setup/list/delete commands are available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_AUTO_NOTIFICATION_CONFIG=true
MHCAT_FEATURE_AUTO_NOTIFICATION_CONFIG_ENABLED=true
```

These commands read/write/delete legacy `cron_sets` rows. `automatic-notification` creates a pending setup row, opens the legacy modal, accepts direct cron expressions or the five-minute owner-scoped weekday/hour/minute wizard, writes the rollback-compatible message payload, and sends a best-effort preview message. `/自動通知列表` filters abandoned setup drafts from the response and cleans rows whose `cron` is null or missing. `/自動通知刪除` deletes one `{guild,id}` row. Config maintenance does not require Message Content intent and does not enable recurring delivery by itself.

Recurring automatic-notification delivery is an independent runtime:

```bash
MHCAT_FEATURE_AUTO_NOTIFICATION_DELIVERY_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_SCHEDULER_LEASE_ENABLED=true
MHCAT_SCHEDULER_LEASE_OWNER=staging-auto-notification
MHCAT_SCHEDULER_LEASE_TTL=2m
MHCAT_SCHEDULER_LEASE_TIMEOUT=10s
```

The worker acquires `auto-notification-delivery`, interprets five-field cron in fixed UTC+8 named `Asia/Taipei`, and reconciles active `cron_sets` rows every 30 seconds or one-third of the lease TTL, whichever is shorter. It reloads `{guild,id}` immediately before each channel send, requires a cached channel belonging to the row's guild, allows the same user/role/everyone mentions as legacy `channel.send`, removes schedules after row deletion or lease loss, and releases the lease during graceful shutdown. It writes only `mhcat_scheduler_locks`; it does not mutate active `cron_sets` rows or require Message Content intent. Disable the Node `handler/cron.js` owner before enabling this runtime. Follow [92-auto-notification.md](92-auto-notification.md) for exact compatibility, isolated staging, reconciliation, and rollback.

`/set-log-channel` is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_LOGGING_CONFIG=true
MHCAT_FEATURE_LOGGING_CONFIG_ENABLED=true
```

The command remains publicly discoverable and checks Manage Messages at runtime. It preserves the exact public red/yellow embeds, one-to-four select options, bot footer avatar format, invoking-user ownership, and absolute ten-minute deadline. New versioned IDs retain channel/user/deadline state; orphaned `loggin_create` components return a safe rerun error. Reads accept Mongoose-compatible String/Boolean scalar values, while successful selection writes a typed `channel_id` and four Booleans to every duplicate guild row, upserting only when none exists. Config maintenance does not create indexes, emit logs, require Message Content intent, or activate any event family. See the [logging parity contract](48-logging-config.md).

Logging message update/delete events are available only when the event runtime flag and gateway message intents are explicitly enabled:

```bash
MHCAT_FEATURE_LOGGING_MESSAGE_EVENTS_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true
MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true
```

This independent runtime reads existing `loggings` rows and emits only message edit/delete embeds when `message_update` or `message_delete` is selected. It requires cached old/deleted payloads, treats the cached pre-edit author as authoritative, preserves exact code-block/attachment/avatar/footer formatting, and suppresses mention parsing. Delete attribution queries action `72` with limit `1` and uses the executor only on an exact author-target and source-channel match; otherwise it displays the original author. It does not enable `/set-log-channel`, create indexes, or emit channel/voice logs.

Logging channel topic/permission update events are available only when the event runtime flag and Gateway are explicitly enabled:

```bash
MHCAT_FEATURE_LOGGING_CHANNEL_EVENTS_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
```

This independent runtime reads existing `loggings` rows and emits only channel topic and permission-overwrite updates when `channel_update` is selected. It requires the cached old channel, renders a null topic as literal `null`, gives a topic change precedence over permission changes, and uses the first action-`11` audit entry without target filtering. Permission diffs preserve the legacy 41-label and default/allow/deny order while matching overwrites by ID and formatting role/user mentions from overwrite type. Sent embeds suppress mention parsing. It does not enable `/set-log-channel`, create indexes, or emit message/voice logs.

Logging voice join/leave events are available only when the event runtime flag, Gateway, and Voice State intent are explicitly enabled:

```bash
MHCAT_FEATURE_LOGGING_VOICE_EVENTS_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_VOICE_STATE_INTENT=true
```

This independent runtime reads existing `loggings` rows and emits only voice join/leave embeds when `member_voice_update` is selected. Join uses the new-state member/channel; leave uses the old-state member/channel. Human and bot joins/leaves are logged, while direct channel moves and mute/deafen-only updates emit nothing. It uses cached member/channel names and suppresses mention parsing. It does not enable `/set-log-channel`, create indexes, or emit message/channel logs.

Before enabling any Go logging event family, stop every Node process that loads `events/LoggingSystem.js` for the same bot/guilds. There is no event-owner lease, so overlap produces duplicate logs and competing audit attribution. Everything after the unterminated comment at legacy `LoggingSystem.js:364` is inactive and outside the parity surface. Slash usage is written once by global middleware only when usage tracking is enabled; selects and gateway events never write usage counters. Use the [logging parity contract](48-logging-config.md) for exact smoke and rollback steps.

Read-only `/扭蛋獎池查詢` is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_LIST=true
MHCAT_FEATURE_GACHA_PRIZE_LIST_ENABLED=true
```

This command reads `gifts` and `gift_changes` only. It does not draw prizes, decrement inventory, send DMs, mutate coins, write usage counters, create indexes, or enable gacha shop behavior. Pools with more than 25 prizes are split across multiple embeds to avoid the legacy Discord API field-limit failure.

`/扭蛋` is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_GACHA_DRAW=true
MHCAT_FEATURE_GACHA_DRAW_ENABLED=true
```

This command reads `coins`, `gifts`, and `gift_changes`, updates matching `coins` rows for the member, decrements or deletes drawn auto-delete `gifts` rows by `{guild,gift_name}`, and may send prize-code DMs plus one notification-channel winner message per non-air draw. It preserves the legacy draw-count choices, loading GIF follow-up, 8.5-second reveal, final result embed, error follow-ups, and per-draw prize-pool reload. It intentionally applies one inventory decrement/delete per drawn prize instead of the legacy duplicate async decrement loops. Use only against isolated staging balances and disposable prize rows until duplicate audits, transaction policy, and DM failure policy are reviewed.

`/扭蛋獎池增加` is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_CREATE=true
MHCAT_FEATURE_GACHA_PRIZE_CREATE_ENABLED=true
```

This command requires Manage Messages and inserts one `gifts` row using the legacy fields `guild`, `gift_name`, `gift_code`, `gift_chence`, `auto_delete`, `gift_count`, and `give_coin`. It preserves the legacy ephemeral defer/edit success and red error embeds, exact untrimmed name/code text, JavaScript numeric name guard, duplicate-name check, optional defaults, zero-chance BSON `null`, and 25-row pool guard. It does not draw prizes, decrement inventory counts, send DMs, mutate user coin balances, write usage counters, create indexes, or enable gacha shop behavior. Use only against disposable staging gacha prize rows until backups and duplicate-name policy are reviewed.

`/扭蛋獎品編輯` is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_EDIT=true
MHCAT_FEATURE_GACHA_PRIZE_EDIT_ENABLED=true
```

This command requires Manage Messages and replaces one exact-name `gifts` row by `{guild,gift_name}` using the legacy fields `guild`, `gift_name`, `gift_code`, `gift_chence`, `auto_delete`, `gift_count`, and `give_coin`. It preserves untrimmed name/code text, the JavaScript numeric name guard, the legacy ephemeral defer/edit success and red error embeds, and legacy merge quirks: omitted or zero chance/coin keep the old value, false `自動刪除` does not override an existing true value, and omitted or zero count saves as `1`. The write path deletes the old row before inserting the merged replacement and has no transaction rollback. It does not draw prizes, decrement inventory counts, send DMs, mutate user coin balances, write usage counters, create indexes, or enable gacha shop behavior. Use only against disposable staging gacha prize rows.

`/扭蛋獎池刪除` is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_DELETE=true
MHCAT_FEATURE_GACHA_PRIZE_DELETE_ENABLED=true
```

This command requires Manage Messages and deletes one exact-name `gifts` row by `{guild,gift_name}` without trimming the submitted or displayed name. It preserves the legacy public defer/edit success and red error embeds, but it does not draw prizes, decrement inventory counts, send DMs, mutate coins, write usage counters, create indexes, or enable gacha shop behavior. Use only against disposable staging gacha prize rows.

Disabled `/抽獎設置` parity is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_LOTTERY_DISABLED_COMMAND=true
MHCAT_FEATURE_LOTTERY_DISABLED_COMMAND_ENABLED=true
```

This command preserves the current legacy unavailable embed. It does not create `lotters` documents, send public lottery panels, write usage counters to Mongo, create indexes, or enable old `lotter*` component behavior.

Existing lottery-message buttons are a separate runtime:

```bash
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_FEATURE_LOTTERY_COMPONENTS_ENABLED=true
```

This gate reads existing `lotters`, atomically appends eligible participants with legacy duplicate/cap/date/role error precedence, returns the legacy participant embed and complete `discord.txt`, sets `end:true` on stop or positive-count reroll, and sends one aggregate winner message on positive-count reroll. Search preserves the 99/100-name cutoff, legacy discriminator and missing-user labels, and Node 20 Taipei timestamp formatting. Reroll preserves stored prize whitespace, bot guild display color, replacement draws, and the legacy nonpositive-count deferred no-op; it intentionally rechecks destructive authorization, allowlists winner mentions, and caps oversized winner counts at 50. It does not enable `/抽獎設置` or create new lottery panels. Test only with disposable copied rows and channels, and do not run Node and Go button ownership for the same guild simultaneously.

Static `/統計系統查詢` parity is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_STATS_QUERY=true
MHCAT_FEATURE_STATS_QUERY_ENABLED=true
```

This command preserves the legacy static stats help embed. It does not create `numbers`/`role_numbers` feature rows, create/delete channels, rename channels, create indexes, or enable the `channel_status` scheduler. When global usage tracking is enabled, the slash middleware still writes exactly one `all_use_counts` event.

Channel-create `/統計系統創建` parity is available only when all required command, runtime, gateway, and intent flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_STATS_CREATE=true
MHCAT_FEATURE_STATS_CREATE_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true
```

This command creates the legacy stats category and base member/user/bot counter channels, can add channel-count/text-count/voice-count stat channels, and writes `numbers` rows. It does not write `role_numbers`, create indexes, or enable the `channel_status` scheduler. Use an isolated staging guild and disposable `numbers` rows.

Role-count `/統計身分組人數` parity is available only when all required command, runtime, gateway, and intent flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_STATS_ROLE_COUNT=true
MHCAT_FEATURE_STATS_ROLE_COUNT_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true
```

This command requires Manage Messages and an existing stats base config, creates a text or voice channel named `<role name>: <member count>`, and replaces the legacy `role_numbers` row for `{guild,role}`. It does not delete old stat channels, create indexes, or enable the `channel_status` scheduler. Use an isolated staging guild and disposable `role_numbers` rows.

Config-row `/統計系統刪除` parity is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_STATS_DELETE=true
MHCAT_FEATURE_STATS_DELETE_ENABLED=true
```

This command requires Manage Messages, deletes legacy `numbers` rows for the guild, and preserves the legacy success/error embed text. It does not delete Discord channels, create indexes, or enable the `channel_status` scheduler. Test only against disposable staging stats config rows.

Stats channel rename parity is event-only and has no command-sync flag. Test it only against an isolated staging guild and disposable `numbers`/`role_numbers` rows:

```bash
MHCAT_FEATURE_STATS_RENAME_WORKER_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true
```

The worker starts with the gateway runtime, then runs on the legacy 20-minute interval. It renames configured member/user/bot/channel/text/voice stat channels and role-count channels using the legacy replace-old-number-or-use-new-number rule, and updates only the corresponding stored old-number fields after a successful rename/no-op decision. It skips missing channels, logs Discord/API failures, writes no indexes, and deletes no Discord channels. Do not run it beside the legacy `channel_status.js` owner for the same guilds. Follow [93-stats.md](93-stats.md) for exact BSON, duplicate, staging, reconciliation, and rollback behavior.

Config-only `/聊天經驗設定` and `/聊天經驗刪除` are available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_TEXT_XP_CONFIG=true
MHCAT_FEATURE_TEXT_XP_CONFIG_ENABLED=true
```

These commands write only legacy-compatible `text_xp_channels` config fields and require Manage Messages. They update duplicate rows for a guild and only upsert when no row exists. They do not enable Message Content intent, Guild Messages intent, text XP accrual, rank rendering, voice XP, automatic reward-role assignment/removal, or usage-counter writes.

Text XP message accrual is event-only and has no command-sync flag. Test it only against disposable staging `text_xps`, `text_xp_channels`, `chat_roles`, `coins`, and `gift_changes` rows with `MHCAT_FEATURE_TEXT_XP_ACCRUAL_ENABLED=true`, `MHCAT_DISCORD_ENABLE_GATEWAY=true`, `MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true`, and `MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true`; it updates XP and level fields, sends configured/default level-up announcements and legacy fallbacks, applies configured `chat_roles`, and grants XP coin rewards after the configured announcement path succeeds.

Config-only `/語音經驗設定` and `/語音經驗刪除` are available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_VOICE_XP_CONFIG=true
MHCAT_FEATURE_VOICE_XP_CONFIG_ENABLED=true
```

These commands write only legacy-compatible `voice_xp_channels` config fields and require Manage Messages. They update duplicate rows for a guild, only upsert when no row exists, and clear `background` because the legacy command showed `背景` but did not save it. They do not enable Voice State intent, voice XP accrual, rank rendering, automatic reward-role assignment/removal, or usage-counter writes.

Voice XP runtime is event-only and has no command-sync flag. Test it only against disposable staging `voice_xps`, `voice_xp_channels`, `voice_roles`, `coins`, and `gift_changes` rows with `MHCAT_FEATURE_VOICE_XP_SESSIONS_ENABLED=true`, `MHCAT_DISCORD_ENABLE_GATEWAY=true`, and `MHCAT_DISCORD_VOICE_STATE_INTENT=true`; it marks `leavejoin` join/leave state, starts one 30-second XP loop per joined user, reconciles existing `leavejoin:"join"` rows on startup, sends configured/default level announcements with owner DM fallbacks, applies `voice_roles`, and grants XP coin rewards after the configured announcement path succeeds.

Config-only `/聊天經驗身分組設定` and `/語音經驗身分組設定` are available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_XP_ROLE_CONFIG=true
MHCAT_FEATURE_XP_ROLE_CONFIG_ENABLED=true
```

These commands add, delete, query, and paginate legacy-compatible `chat_roles` and `voice_roles` reward-role config rows. They require Manage Messages, preserve the legacy misspelled `leavel` string field, check that the selected role is assignable by the bot before saving, and create no indexes. They do not enable text/voice XP accrual, rank rendering, automatic reward-role assignment/removal, Message Content intent, Guild Messages intent, Voice State intent, or usage-counter writes. Follow the payload, preserved-quirk, and staging audit in `docs/68-xp-reward-role-config.md`.

Disabled-response `/聊天經驗` and `/語音經驗` are available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_XP_PROFILE_DISABLED_COMMANDS=true
MHCAT_FEATURE_XP_PROFILE_DISABLED_COMMANDS_ENABLED=true
```

These commands preserve the current legacy replacement embed that tells users to use `/我的檔案`. They do not read `text_xps` or `voice_xps`, render rank cards, award XP, require gateway intents, write Mongo, or enable level-role behavior.

XP admin `/經驗值改變` is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_XP_ADMIN=true
MHCAT_FEATURE_XP_ADMIN_ENABLED=true
```

This command requires Kick Members and adjusts one selected member's text or voice XP profile. It writes legacy-compatible `text_xps`/`voice_xps` rows using string `xp` and misspelled string `leavel`, and voice-profile inserts set `leavejoin` to `leave`. Test only against disposable staging XP rows. It does not enable text/voice XP accrual, rank rendering, automatic reward-role assignment/removal, Message Content intent, Guild Messages intent, Voice State intent, or usage-counter writes. Follow the adjustment, payload, compatibility, and staging audit in `docs/69-xp-admin.md`.

XP reset `/經驗值重製` is available only when staging command sync, runtime, gateway, Guild Messages, and Message Content flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_XP_RESET=true
MHCAT_FEATURE_XP_RESET_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true
MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true
```

This command requires the Discord guild owner. Individual text/voice reset subcommands delete only the selected member's `text_xps` or `voice_xps` rows. Full-server text/voice reset subcommands send the legacy destructive warning and require the same owner to type `^確認^` in the same channel within 60 seconds before deleting all staging guild XP rows for that collection. Test only against disposable staging XP rows; command sync and staging scripts reject unpaired sync/runtime flags. Follow the payload, confirmation-state, compatibility, and staging audit in `docs/70-xp-reset.md`.

XP rank `/聊天排行榜` and `/語音排行榜` are available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_XP_RANK=true
MHCAT_FEATURE_XP_RANK_ENABLED=true
```

Set both together only in an isolated staging database. This path reads `text_xps`/`voice_xps`, renders legacy-style `user-info.png` leaderboard pages with legacy rank buttons, and writes no Mongo data. It does not enable XP accrual, `/聊天經驗` profile cards, automatic reward roles, coin rewards, gateway intents, or usage-counter writes. Ensure the legacy rank background, fallback icon, and font files are available from the bot working directory, then follow the payload, ranking-math, asset, intentional-difference, and staging audit in `docs/71-xp-rank.md`.

Voice-room `/語音包廂設置` and `/語音包廂刪除` are available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_CONFIG=true
MHCAT_FEATURE_VOICE_ROOM_CONFIG_ENABLED=true
```

The command definitions remain publicly discoverable like legacy, while both handlers require Manage Messages. Setup writes legacy-compatible `voice_channels`, preserving raw name whitespace and explicit `limit:0`; only the first `{name}` is replaced. Slash delete removes config rows only and preserves the legacy branch where only a type-2 voice channel deletes by trigger while stage/category selections delete by parent. With Gateway and Voice State intent enabled, human or bot trigger joins create rooms under the trigger's current parent, copy current overwrites plus owner management permissions, persist `voice_channel_ids`, seed nullable `lock_channels`, move the member, and delete empty tracked rooms. Follow the compatibility, ownership, smoke, and rollback audit in the [voice-room parity contract](72-voice-room-config.md).

Command-only `/上鎖頻道` is available only when staging command sync, runtime, gateway, and Voice State flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_LOCK=true
MHCAT_FEATURE_VOICE_ROOM_LOCK_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_VOICE_STATE_INTENT=true
```

This publicly discoverable feature reads the actor's current voice channel from DiscordGo state, verifies ownership of the existing `lock_channels` row, and replaces it with a raw nullable `lock_anser` password and empty BSON-array `ok_people`. Existing passworded rows send the exact legacy prompt/DM colors, disconnect unauthorized humans, and bind the prompt button to the joining user and a 60-second deadline. Modal submits compare raw passwords exactly and add correct users idempotently; old orphaned `lock_start` buttons return a retry error. When global usage tracking is enabled, slash attempts increment `all_use_counts` once through middleware; voice-state, button, and modal handling add no usage writes. See the [voice-room parity contract](72-voice-room-config.md).

Config-only `/加入身份組設置` and `/加入身份組刪除` are available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_JOIN_ROLE_CONFIG=true
MHCAT_FEATURE_JOIN_ROLE_CONFIG_ENABLED=true
```

These publicly discoverable commands write typed legacy-compatible `join_roles`, require Manage Messages at runtime, and preserve exact UI. Assignment is independently gated by Gateway/Guild Members, uses cached hierarchy checks, and continues after bad rows. It does not enable welcome, verification, or account-age. Follow [81-join-role.md](81-join-role.md).

Config-only `/驗證設置` is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_CONFIG=true
MHCAT_FEATURE_VERIFICATION_CONFIG_ENABLED=true
```

This publicly discoverable command writes legacy-compatible `verifications` fields and requires Manage Messages at runtime. It performs the legacy bot-role hierarchy check before saving. It does not enable `/驗證`, member side effects, or account-age policy. Audit and rollback requirements are in [80-verification.md](80-verification.md).

The full `/驗證` captcha flow is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_FLOW=true
MHCAT_FEATURE_VERIFICATION_FLOW_ENABLED=true
```

This flow reads `verifications`, sends the exact ephemeral captcha/button UI, opens the legacy modal, assigns the role, and optionally renames. New IDs use guild/user-bound five-minute process-local state with atomic completion; strict legacy IDs remain supported. It creates no indexes or handler-local usage writes. Multi-process rollout requires shared state or verified sticky routing; follow [80-verification.md](80-verification.md).

Config-only `/帳號需創建時數` is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_ACCOUNT_AGE_CONFIG=true
MHCAT_FEATURE_ACCOUNT_AGE_CONFIG_ENABLED=true
```

This command writes only the legacy-compatible `create_hours` config fields: `guild`, string `hours` in seconds, and nullable `channel`. It requires Kick Members, preserves the legacy public defer/edit reply UI, success/error embeds, and the legacy typo `發送使用者資運`. It does not by itself enable member kicking.

The account-age member-add policy is a separate event path. Enable it only in an isolated staging guild/database after the config command has created a `create_hours` row:

```bash
MHCAT_FEATURE_ACCOUNT_AGE_POLICY_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true
```

The policy reads `create_hours` on `guildMemberAdd`. If the account is younger than the configured threshold, it sends the legacy bilingual DM embed, kicks with the legacy reason, optionally logs the legacy embed to a cached `channel`, and stops later member-add handlers so join-role and welcome behavior do not run after a kick. Unlike legacy unhandled promises, Go awaits kick/log errors and ignores only non-context DM failures so closed DMs do not bypass the protection.

See the canonical [account-age parity contract](79-account-age.md) for exact UI/data/event behavior, migration, staging, and rollback.

Role-selection commands and reaction role events are available only when staging command sync, runtime, Gateway, and reaction intent flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_ROLE_SELECTION=true
MHCAT_FEATURE_ROLE_SELECTION_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_GUILD_MESSAGE_REACTIONS_INTENT=true
```

This slice writes legacy-compatible `message_reactions` rows for reaction-role mappings and `btns` rows for role-button add/delete mappings. It preserves the legacy `nal` modal and legacy `<id>add`/`<id>delete` button IDs, checks the bot can assign the configured role, adds the configured reaction to the target message, and handles reaction add/remove events by adding or removing the configured role. `MHCAT_FEATURE_ROLE_SELECTION_ENABLED` is deliberately one ownership boundary for setup commands, modal/buttons, and reaction events. Do not split these paths into separate Go gates or leave the corresponding Node interaction/reaction handlers active while Go owns the feature.

Reaction-add gateway payloads include the member and are retained in Discord state so a later remove from the same process can still identify and ignore bots. Discord reaction-remove payloads do not include the member. A remove first observed after a process restart is therefore best-effort: when the member is absent from state, bot identity is unknown and the event is handled as a non-bot. Test only with staging roles below the bot's highest role and staging messages; this slice does not create indexes or route-level usage-counter writes. The exact migration, duplicate-audit, smoke, and rollback procedure is in the [role-selection parity contract](73-role-selection.md).

Welcome-message member-add delivery is a separate event path. Enable it only when existing dashboard/legacy `join_messages` rows are safe for the staging guild:

```bash
MHCAT_FEATURE_WELCOME_MESSAGE_DELIVERY_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true
```

The delivery path reads `join_messages`, sends the legacy welcome embed on `guildMemberAdd`, allows only the joining user mention for `(TAG)`/`{TAG}` placeholders, and performs no Mongo writes. The legacy MHCAT-server special welcome embed is available without hardcoded IDs by setting `MHCAT_LEGACY_WELCOME_SPECIAL_GUILD_ID`, `MHCAT_LEGACY_WELCOME_SPECIAL_BOT_ID`, `MHCAT_LEGACY_WELCOME_SPECIAL_CHANNEL_ID`, `MHCAT_LEGACY_WELCOME_SPECIAL_CHAT_CHANNEL_ID`, `MHCAT_LEGACY_WELCOME_SPECIAL_HELP_CHANNEL_ID`, `MHCAT_LEGACY_WELCOME_SPECIAL_BUG_CHANNEL_ID`, and `MHCAT_LEGACY_WELCOME_SPECIAL_SUPPORT_CHANNEL_ID` together.

Generic, special, and leave sends require cached guild channels and never REST-bypass a missing cache entry. Exact UI/data behavior, ownership, migration, staging, and rollback are in [82-welcome-leave.md](82-welcome-leave.md).

## Economy Daily Reset

Go provides a dry-run/apply command and a separately gated recurring worker for the legacy `00:00 Asia/Taipei` economy reset. Both Go write paths coordinate through lease name `daily-reset`.

Preview only:

```bash
go run ./cmd/mhcat-economy-reset --dry-run
```

Apply requires the write gate, scheduler lease gate, a unique owner, and the explicit CLI flag:

```bash
MHCAT_JOBS_DAILY_RESET_ENABLED=true \
MHCAT_SCHEDULER_LEASE_ENABLED=true \
MHCAT_SCHEDULER_LEASE_OWNER=staging-reset-cli \
go run ./cmd/mhcat-economy-reset --apply
```

Apply exits with code `2` without writes when another owner holds `daily-reset`, and releases the lease after success or failure. The command can reset `coins.today` and refill/clamp `work_users.energi`. It does not create indexes, repair documents, sync commands, or write any other feature data.

Recurring runtime:

```bash
MHCAT_FEATURE_DAILY_RESET_SCHEDULER_ENABLED=true
MHCAT_JOBS_DAILY_RESET_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_SCHEDULER_LEASE_ENABLED=true
MHCAT_SCHEDULER_LEASE_OWNER=staging-reset-bot-a
MHCAT_SCHEDULER_LEASE_TTL=2m
MHCAT_SCHEDULER_LEASE_TIMEOUT=10s
MHCAT_JOBS_DAILY_RESET_TIMEOUT=60s
```

Every Go replica schedules `0 0 * * *` at fixed UTC+8 named `Asia/Taipei`, but only the per-tick lease holder writes. The worker makes no Discord API call and needs no privileged intent; Gateway is currently required for its app lifecycle. The lease TTL must exceed reset timeout plus lease-operation timeout. Stop Node `handler/cron.js` before Go apply or recurring ownership. Do not blindly retry a partial failure because already-processed work guilds can receive another energy increment. See `docs/41-economy-daily-reset.md`.

## Work Payout

The Go refactor provides a one-shot payout command and a separately gated recurring worker for completed legacy work jobs in `MHCAT/handler/gift.js`.

Preview only:

```bash
go run ./cmd/mhcat-work-payout --dry-run
```

Apply requires the work-payout gate, scheduler lease gate, scheduler owner, and explicit CLI flag:

```bash
MHCAT_JOBS_WORK_PAYOUT_ENABLED=true \
MHCAT_SCHEDULER_LEASE_ENABLED=true \
MHCAT_SCHEDULER_LEASE_OWNER=staging-worker \
go run ./cmd/mhcat-work-payout --apply
```

The command conditionally increments `coins.coin` and writes the latest per-work-row token under `coins.mhcat_work_payouts` in one atomic update. A retry after the coin commit reports `idempotent_replays` and does not increment again. Missing balances use a deterministic ObjectID; duplicate `{guild,member}` coin rows fail before credit. The command resets only the exact due `work_users` snapshot to `待業中`. It does not create indexes, repair documents, sync commands, or send Discord messages.

If another process holds the configured scheduler lease, apply mode skips the payout and exits with code `2`. Do not run production apply until duplicate and marker-shape audit results are reviewed and Node.js is no longer owning the same minute payout loop. On rollback, stop Go, confirm lease release/expiry, then restore Node; leave marker fields in place.

Recurring staging ownership requires:

```bash
MHCAT_FEATURE_WORK_PAYOUT_SCHEDULER_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_JOBS_WORK_PAYOUT_ENABLED=true
MHCAT_JOBS_WORK_PAYOUT_LEASE_NAME=work-payout
MHCAT_JOBS_WORK_PAYOUT_TIMEOUT=60s
MHCAT_SCHEDULER_LEASE_ENABLED=true
MHCAT_SCHEDULER_LEASE_OWNER=staging-worker
MHCAT_SCHEDULER_LEASE_TTL=2m
MHCAT_SCHEDULER_LEASE_TIMEOUT=10s
```

Every replica schedules `* * * * *` at fixed `Asia/Taipei`; only the per-tick lease holder writes. The worker skips a local overlapping callback, releases after success/failure/shutdown, and shares the configured lease with CLI apply. It performs no Discord API call, but Gateway provides its current process lifecycle. `MHCAT_JOBS_WORK_PAYOUT_DRY_RUN` is CLI-only and does not make the recurring worker read-only. Stop Node `handler/gift.js`, use unique owner names, audit duplicates/marker shapes, and run an isolated two-replica minute-tick smoke before production.

## Work Command Runtime

The `打工系統` command is disabled by default. The current Go runtime preserves the legacy `新增打工事項` dashboard redirect UI and a legacy-style `打工介面` flow that can list jobs, show the captcha modal, render role-filtered job buttons, show job detail, start a job, show the busy override prompt, and cancel the prompt. It also implements legacy-style `打工系統設定`, `打工事項刪除`, `增加個人精力`, and `增加全體精力` behind explicit admin repository wiring and Manage Messages checks. Completed-work payout is independently gated from this command runtime.

To smoke this partial command in staging only:

```bash
export MHCAT_FEATURE_WORK_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_WORK=true
```

Run `mhcat-staging-preflight` before command sync. It rejects `MHCAT_COMMAND_SYNC_INCLUDE_WORK=true` unless the runtime flag is also enabled, and command sync still requires staging guild scope. Do not sync `打工系統` to production until recurring payout ownership, duplicate audits, and isolated payout smoke are complete or a documented partial rollout is accepted. The start path can create a missing `work_users` row, deduct `energi`, and set `state`/`end_time`/`get_coin` through an atomic update. Admin paths can upsert/update `work_sets`, delete `work_somethings`, and clamp `work_users.energi`. They do not write payout state, coins, indexes, or scheduler leases.

## Birthday Command Runtime

Birthday command ownership is disabled by default. In an isolated staging guild/database, pair:

```bash
export MHCAT_FEATURE_BIRTHDAY_CONFIG_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_BIRTHDAY_CONFIG=true
```

Stop the Node `/生日系統` owner, audit `birthday_sets`/`birthdays`, run preflight and command-sync dry-run, and follow [78-birthday.md](78-birthday.md). These flags authorize only the five active slash subcommands and their selectors. They do not authorize the fully commented `handler/gift.js` delivery/temporary-role block, a scheduler, repairs, or index creation. During rollback, disable Go routing and wait five minutes for versioned pending selector IDs to expire before restoring Node command ownership.

## Scheduler Lease

The scheduler lease foundation is wired into bot startup for separately gated automatic-notification delivery, daily reset, and work payout workers.

- Collection: `mhcat_scheduler_locks`.
- Identity: `_id` equals `lock_name`.
- Ownership: `owner` plus monotonic `fence`.
- Expiry: UTC `expires_at`.
- Release behavior: marks the lease expired and clears owner; it does not delete the document.

The automatic-notification worker continuously owns `auto-notification-delivery`, renews before expiry, removes cron entries after lease loss, and releases on graceful shutdown. Daily-reset CLI apply and each midnight worker tick acquire/release `daily-reset`. Work-payout CLI apply and each minute tick acquire/release the configured payout lease. Contenders perform no writes. Leases do not coordinate with Node, so each corresponding Node owner must be disabled before Go ownership.

Read-only status:

```bash
go run ./cmd/mhcat-scheduler-lease --name daily-reset --action status
```

Write diagnostics require both the env gate and `--apply`:

```bash
MHCAT_SCHEDULER_LEASE_ENABLED=true \
MHCAT_SCHEDULER_LEASE_OWNER=staging-worker \
go run ./cmd/mhcat-scheduler-lease --name daily-reset --action acquire --apply
```

## Canary Rollout

1. Staging bot/token in staging guild.
2. Shadow read-only validation with Discord/Mongo writes disabled.
3. Canary guild with read-only slash commands.
4. Canary low-risk writes.
5. Canary event handlers.
6. Canary schedulers/background jobs.
7. Global rollout after parity and rollback checks.

Node and Go must not both own the same guild/feature at the same time.

## How to Roll Back to Node.js Bot

1. Stop Go process.
2. Disable Go feature flags and command registration.
3. Restart Node.js bot with existing env.
4. Confirm Node can read documents written during canary.
5. If an index caused issues, follow index-specific rollback notes.

Rollback must not require Mongo data mutation unless a prior ADR explicitly accepted that risk.

## How to Troubleshoot Discord Rate Limits

- Check command registration mode.
- Check REST queue depth and 429 metrics.
- Check channel rename loops and bulk operations.
- Disable high-churn feature flags if needed.
- Avoid repeated global command registration.

## How to Troubleshoot Mongo Timeouts

- Run `mhcat-tools mongo ping`.
- Check slow query logs and planned indexes.
- Check context timeout metrics.
- Disable hot event features if unindexed scans are causing incident.
- Use dry-run index plan before apply.

## How to Rotate Secrets

- Rotate Discord token and webhooks outside the repo.
- Update env/secret store.
- Restart bot.
- Confirm redacted logs do not expose values.
- Run hardcoded secret scan.
- The legacy hardcoded webhook should be revoked if still active.

## Incident Checklist

- Identify active feature flags and owning process.
- Confirm Node and Go are not double-handling a guild/feature.
- Check Mongo connectivity and scheduler leadership.
- Check Discord REST 429s and gateway reconnects.
- Decide rollback vs feature-disable.
- Record any data repair need as audit/dry-run first.
