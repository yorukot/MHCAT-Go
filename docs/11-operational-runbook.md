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
- `MHCAT_FEATURE_TICKETS_ENABLED`
- `MHCAT_FEATURE_POLLS_ENABLED`
- `MHCAT_FEATURE_ECONOMY_QUERY_ENABLED`
- `MHCAT_FEATURE_ECONOMY_SIGNIN_ENABLED`
- `MHCAT_FEATURE_ECONOMY_SETTINGS_ENABLED`
- `MHCAT_FEATURE_WORK_ENABLED`
- `MHCAT_FEATURE_WARNINGS_ENABLED`
- `MHCAT_FEATURE_TRANSLATE_ENABLED`
- `MHCAT_FEATURE_AUTO_NOTIFICATION_CONFIG_ENABLED`
- `MHCAT_FEATURE_LOGGING_CONFIG_ENABLED`
- `MHCAT_FEATURE_GACHA_PRIZE_LIST_ENABLED`
- `MHCAT_FEATURE_LOTTERY_DISABLED_COMMAND_ENABLED`
- `MHCAT_FEATURE_STATS_QUERY_ENABLED`
- `MHCAT_FEATURE_ANNOUNCEMENT_CONFIG_ENABLED`
- `MHCAT_FEATURE_TEXT_XP_CONFIG_ENABLED`
- `MHCAT_FEATURE_VOICE_XP_CONFIG_ENABLED`
- `MHCAT_FEATURE_XP_PROFILE_DISABLED_COMMANDS_ENABLED`
- `MHCAT_FEATURE_JOIN_ROLE_CONFIG_ENABLED`
- `MHCAT_FEATURE_JOIN_ROLE_ASSIGNMENT_ENABLED`
- `MHCAT_FEATURE_WELCOME_MESSAGE_CONFIG_ENABLED`
- `MHCAT_FEATURE_WELCOME_MESSAGE_DELIVERY_ENABLED`
- `MHCAT_FEATURE_LEAVE_MESSAGE_DELIVERY_ENABLED`
- `MHCAT_FEATURE_VERIFICATION_CONFIG_ENABLED`
- `MHCAT_FEATURE_VERIFICATION_FLOW_ENABLED`
- `MHCAT_FEATURE_ACCOUNT_AGE_CONFIG_ENABLED`
- `MHCAT_FEATURE_ACCOUNT_AGE_POLICY_ENABLED`
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
- Pair optional feature command-sync flags with matching runtime flags, for example `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_QUERY=true` requires `MHCAT_FEATURE_ECONOMY_QUERY_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SIGNIN=true` requires `MHCAT_FEATURE_ECONOMY_SIGNIN_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SETTINGS=true` requires `MHCAT_FEATURE_ECONOMY_SETTINGS_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_WORK=true` requires `MHCAT_FEATURE_WORK_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_WARNINGS=true` requires `MHCAT_FEATURE_WARNINGS_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_TRANSLATE=true` requires `MHCAT_FEATURE_TRANSLATE_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_BALANCE_QUERY=true` requires `MHCAT_FEATURE_BALANCE_QUERY_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_AUTO_NOTIFICATION_CONFIG=true` requires `MHCAT_FEATURE_AUTO_NOTIFICATION_CONFIG_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_LOGGING_CONFIG=true` requires `MHCAT_FEATURE_LOGGING_CONFIG_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_LIST=true` requires `MHCAT_FEATURE_GACHA_PRIZE_LIST_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_LOTTERY_DISABLED_COMMAND=true` requires `MHCAT_FEATURE_LOTTERY_DISABLED_COMMAND_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_STATS_QUERY=true` requires `MHCAT_FEATURE_STATS_QUERY_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_CONFIG=true` requires `MHCAT_FEATURE_ANNOUNCEMENT_CONFIG_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_TEXT_XP_CONFIG=true` requires `MHCAT_FEATURE_TEXT_XP_CONFIG_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_VOICE_XP_CONFIG=true` requires `MHCAT_FEATURE_VOICE_XP_CONFIG_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_XP_PROFILE_DISABLED_COMMANDS=true` requires `MHCAT_FEATURE_XP_PROFILE_DISABLED_COMMANDS_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_JOIN_ROLE_CONFIG=true` requires `MHCAT_FEATURE_JOIN_ROLE_CONFIG_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_CONFIG=true` requires `MHCAT_FEATURE_VERIFICATION_CONFIG_ENABLED=true`, and `MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_FLOW=true` requires `MHCAT_FEATURE_VERIFICATION_FLOW_ENABLED=true`.
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

Read-only `/代幣查詢` is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_QUERY=true
MHCAT_FEATURE_ECONOMY_QUERY_ENABLED=true
```

This command reads `coins` and `gift_changes` only. It does not write economy data, usage counters, or indexes.

`/coin-related-settings` is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SETTINGS=true
MHCAT_FEATURE_ECONOMY_SETTINGS_ENABLED=true
```

This command writes `gift_changes` with legacy field names and requires Manage Messages. It updates all duplicate rows for the guild so old `findOne` consumers remain rollback-compatible until duplicate audit and unique-index work are complete. Do not enable it in production until shared gacha/sign-in/XP consumers and duplicate audits are reviewed.

Bot startup still does not sync commands. Run command sync manually and review dry-run output before any `--apply`.

Read-only `/警告紀錄` is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_WARNINGS=true
MHCAT_FEATURE_WARNINGS_ENABLED=true
```

This command reads `warndbs` only. It does not create/remove warnings, run escalation rules, delete messages, kick/ban members, write usage counters, or create indexes. It intentionally enforces Manage Messages and falls back to moderator IDs when old warning rows reference uncached members.

`/翻譯` is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_TRANSLATE=true
MHCAT_FEATURE_TRANSLATE_ENABLED=true
```

This command calls the external translate provider through a driver-agnostic port. It does not require Message Content intent, does not read or write Mongo feature data, and returns a safe red error embed instead of leaving the legacy loading embed stuck when the provider fails.

`/查看餘額` is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_BALANCE_QUERY=true
MHCAT_FEATURE_BALANCE_QUERY_ENABLED=true
```

This command reads `chatgpt_gets.price` only. It does not enable ChatGPT/autochat message runtime, does not require Message Content intent, and writes no Mongo feature data.

Auto-notification list/delete commands are available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_AUTO_NOTIFICATION_CONFIG=true
MHCAT_FEATURE_AUTO_NOTIFICATION_CONFIG_ENABLED=true
```

These commands read/delete legacy `cron_sets` rows only. `/自動通知列表` filters abandoned setup drafts from the response and cleans rows whose `cron` is null or missing. `/自動通知刪除` deletes one `{guild,id}` row. This does not enable `automatic-notification`, the cron modal/select flow, Message Content intent, recurring scheduler ownership, or notification sends. See `docs/66-auto-notification-config.md`.

`/set-log-channel` is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_LOGGING_CONFIG=true
MHCAT_FEATURE_LOGGING_CONFIG_ENABLED=true
```

This command writes the legacy-compatible `loggings` config fields and requires Manage Messages. It updates all duplicate rows for the guild and only upserts when no row exists. It does not create indexes, emit message/channel/voice logs, require Message Content intent, or enable audit-log event processing.

Read-only `/扭蛋獎池查詢` is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_LIST=true
MHCAT_FEATURE_GACHA_PRIZE_LIST_ENABLED=true
```

This command reads `gifts` and `gift_changes` only. It does not draw prizes, decrement inventory, send DMs, mutate coins, write usage counters, create indexes, or enable gacha shop behavior. Pools with more than 25 prizes are split across multiple embeds to avoid the legacy Discord API field-limit failure.

Disabled `/抽獎設置` parity is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_LOTTERY_DISABLED_COMMAND=true
MHCAT_FEATURE_LOTTERY_DISABLED_COMMAND_ENABLED=true
```

This command preserves the current legacy unavailable embed. It does not create `lotters` documents, send public lottery panels, register lottery buttons, write usage counters to Mongo, create indexes, or enable old `lotter*` component behavior.

Static `/統計系統查詢` parity is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_STATS_QUERY=true
MHCAT_FEATURE_STATS_QUERY_ENABLED=true
```

This command preserves the legacy static stats help embed. It does not create `Number`/`role_number` rows, create/delete channels, rename channels, write usage counters to Mongo, create indexes, or enable the `channel_status` scheduler.

Config-only `/聊天經驗設定` and `/聊天經驗刪除` are available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_TEXT_XP_CONFIG=true
MHCAT_FEATURE_TEXT_XP_CONFIG_ENABLED=true
```

These commands write only legacy-compatible `text_xp_channels` config fields and require Manage Messages. They update duplicate rows for a guild and only upsert when no row exists. They do not enable Message Content intent, Guild Messages intent, text XP accrual, rank rendering, voice XP, XP reward roles, or usage-counter writes.

Config-only `/語音經驗設定` and `/語音經驗刪除` are available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_VOICE_XP_CONFIG=true
MHCAT_FEATURE_VOICE_XP_CONFIG_ENABLED=true
```

These commands write only legacy-compatible `voice_xp_channels` config fields and require Manage Messages. They update duplicate rows for a guild, only upsert when no row exists, and clear `background` because the legacy command showed `背景` but did not save it. They do not enable Voice State intent, voice XP accrual, rank rendering, XP reward roles, or usage-counter writes.

Disabled-response `/聊天經驗` and `/語音經驗` are available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_XP_PROFILE_DISABLED_COMMANDS=true
MHCAT_FEATURE_XP_PROFILE_DISABLED_COMMANDS_ENABLED=true
```

These commands preserve the current legacy replacement embed that tells users to use `/我的檔案`. They do not read `text_xps` or `voice_xps`, render rank cards, award XP, require gateway intents, write Mongo, or enable level-role behavior.

Config-only `/加入身份組設置` and `/加入身份組刪除` are available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_JOIN_ROLE_CONFIG=true
MHCAT_FEATURE_JOIN_ROLE_CONFIG_ENABLED=true
```

These commands write only legacy-compatible `join_roles` config fields and require Manage Messages. The setup command performs the legacy bot-role hierarchy check through the Discord adapter before saving. This slice does not enable Guild Members intent, `guildMemberAdd` role assignment, join/leave message emitters, verification, account-age kick, or usage-counter writes.

Config-only `/驗證設置` is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_CONFIG=true
MHCAT_FEATURE_VERIFICATION_CONFIG_ENABLED=true
```

This command writes only legacy-compatible `verifications` config fields and requires Manage Messages. The setup command performs the legacy bot-role hierarchy check through the Discord adapter before saving. It does not enable `/驗證`, captcha generation, verification buttons/modals, member role assignment, nickname changes, account-age kick, or usage-counter writes.

The full `/驗證` captcha flow is available only when both staging command sync and runtime flags are explicitly enabled:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_FLOW=true
MHCAT_FEATURE_VERIFICATION_FLOW_ENABLED=true
```

This flow reads `verifications`, sends a legacy-style ephemeral `captcha.jpeg` prompt with the green `點我進行驗證!` button, shows the legacy `請輸入驗證碼!` modal, assigns the configured role, and optionally applies the legacy `{name}` nickname template. New Go-generated component/modal IDs use a bounded state ID rather than embedding the captcha answer; old `<captcha>verification` and `<captcha>ver` IDs remain supported for live-message compatibility. It does not create Mongo indexes or usage-counter writes.

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

The policy reads `create_hours` on `guildMemberAdd`. If the account is younger than the configured threshold, it sends the legacy bilingual DM embed, kicks with the legacy reason, optionally logs the legacy embed to `channel`, and stops later member-add handlers so join-role and welcome behavior do not run after a kick. Unlike legacy unhandled promises, Go awaits kick/log errors and ignores only non-context DM failures so closed DMs do not bypass the protection.

See `docs/60-account-age-protection.md` for the exact legacy references, Mongo compatibility notes, and staging checklist.

Welcome-message member-add delivery is a separate event path. Enable it only when existing dashboard/legacy `join_messages` rows are safe for the staging guild:

```bash
MHCAT_FEATURE_WELCOME_MESSAGE_DELIVERY_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true
```

The delivery path reads `join_messages`, sends the legacy welcome embed on `guildMemberAdd`, allows only the joining user mention for `(TAG)`/`{TAG}` placeholders, and performs no Mongo writes. The legacy MHCAT-server special welcome embed is available without hardcoded IDs by setting `MHCAT_LEGACY_WELCOME_SPECIAL_GUILD_ID`, `MHCAT_LEGACY_WELCOME_SPECIAL_BOT_ID`, `MHCAT_LEGACY_WELCOME_SPECIAL_CHANNEL_ID`, `MHCAT_LEGACY_WELCOME_SPECIAL_CHAT_CHANNEL_ID`, `MHCAT_LEGACY_WELCOME_SPECIAL_HELP_CHANNEL_ID`, `MHCAT_LEGACY_WELCOME_SPECIAL_BUG_CHANNEL_ID`, and `MHCAT_LEGACY_WELCOME_SPECIAL_SUPPORT_CHANNEL_ID` together.

## Economy Daily Reset

The Go refactor provides a one-shot reset command for the legacy `00:00 Asia/Taipei` economy reset. It is not wired into `cmd/mhcat-bot` and it is not a recurring scheduler.

Preview only:

```bash
go run ./cmd/mhcat-economy-reset --dry-run
```

Apply requires both an explicit env gate and CLI flag:

```bash
MHCAT_JOBS_DAILY_RESET_ENABLED=true \
go run ./cmd/mhcat-economy-reset --apply
```

The command can reset `coins.today` and refill/clamp `work_users.energi`. It does not create indexes, repair documents, sync commands, or write any other feature data. Do not run production apply until the dry-run output and duplicate audit results are reviewed.

## Work Payout

The Go refactor provides a one-shot payout command for completed legacy work jobs in `MHCAT/handler/gift.js`. It is not wired into `cmd/mhcat-bot` and it is not a recurring scheduler.

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

The command can increment `coins.coin`, create missing `coins` documents, and reset due `work_users.state` to `待業中`. It does not create indexes, repair documents, sync commands, send Discord messages, or write from bot startup.

If another process holds the configured scheduler lease, apply mode skips the payout and exits with code `2`. Do not run production apply until duplicate audit results are reviewed and Node.js is no longer owning the same minute payout loop.

## Work Command Runtime

The `打工系統` command is disabled by default. The current Go runtime preserves the legacy `新增打工事項` dashboard redirect UI and a legacy-style `打工介面` flow that can list jobs, show the captcha modal, render role-filtered job buttons, show job detail, start a job, show the busy override prompt, and cancel the prompt. It also implements legacy-style `打工系統設定`, `打工事項刪除`, `增加個人精力`, and `增加全體精力` behind explicit admin repository wiring and Manage Messages checks. Recurring jobs remain intentionally unimplemented.

To smoke this partial command in staging only:

```bash
export MHCAT_FEATURE_WORK_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_WORK=true
```

Run `mhcat-staging-preflight` before command sync. It rejects `MHCAT_COMMAND_SYNC_INCLUDE_WORK=true` unless the runtime flag is also enabled, and command sync still requires staging guild scope. Do not sync `打工系統` to production until scheduler ownership, duplicate audits, and payout idempotency are reviewed or a documented partial rollout is accepted. The start path can create a missing `work_users` row, deduct `energi`, and set `state`/`end_time`/`get_coin` through an atomic update. Admin paths can upsert/update `work_sets`, delete `work_somethings`, and clamp `work_users.energi`. They do not write payout state, coins, indexes, or scheduler leases.

## Scheduler Lease

The scheduler lease foundation is implemented in code but not wired into bot startup.

- Collection: `mhcat_scheduler_locks`.
- Identity: `_id` equals `lock_name`.
- Ownership: `owner` plus monotonic `fence`.
- Expiry: UTC `expires_at`.
- Release behavior: marks the lease expired and clears owner; it does not delete the document.

No recurring job should be enabled until it uses the lease and has job-specific idempotency tests. This applies to work payout, automatic notifications, and any future background daily reset loop.

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
