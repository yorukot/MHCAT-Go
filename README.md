# MHCAT Refactor

Platform Wave A shared Discord runtime infrastructure, Wave 5.11 utility UI parity, and gated feature-slice parity for the Go + DiscordGo + MongoDB refactor.

This module currently provides:

- config loading and validation;
- secret redaction helpers;
- structured `log/slog` logging;
- app lifecycle and shutdown wiring;
- MongoDB client creation, connect, ping, and disconnect only;
- MongoDB collection catalog for all 47 legacy Mongoose models;
- MongoDB audit report models and read-only audit CLI;
- MongoDB index plan/diff models and dry-run index CLI;
- MongoDB atomic update helper builders;
- MongoDB transaction runner shell and repository base contracts;
- DiscordGo session creation and opt-in intent builder;
- opt-in gateway event dispatch for ready/resume, messages, reactions, guild members, and voice state events;
- opt-in Discord Gateway open/close lifecycle;
- DiscordGo `InteractionCreate` adapter into the internal router;
- command registry metadata types with option choices and channel type constraints;
- command diff planning and dry-run command sync CLI;
- interaction responder state machine;
- response attachment, component disabled/default state, and allowed-mentions mapping;
- interaction router and middleware interfaces;
- driver-agnostic Discord side-effect ports for messages, channels, roles, members, and audit-log reads;
- versioned component/modal custom ID encoder and parser;
- legacy component/modal custom ID compatibility decoders;
- collision and ambiguity tests for known legacy custom IDs;
- feature module registry and utility feature route binding;
- low-risk utility command handlers for `help`, `ping`, `info bot`, `info shard`, `info user`, and `info guild`.
- legacy-style `/help` embeds, category select menu, link buttons, category pages, and command detail pages.
- legacy-style `/info bot` system embed and `botinfoupdate` refresh button.
- legacy-style `/info shard` embed and `shardinfoupdate` refresh button.
- legacy-style `/info user` and `/info guild` read-only embeds.
- legacy-style `/代幣查詢` read-only embed when `MHCAT_FEATURE_ECONOMY_QUERY_ENABLED=true`.
- runtime routing for `/help`, `/ping`, `/info bot`, `/info shard`, `/info user`, and `/info guild` when the gateway is explicitly enabled.
- staging-only command sync and gateway smoke guardrails.
- local-only staging preflight reporting before live staging attempts.
- local MongoDB Compose service for host-side smoke/audit runs.
- sanitized production Mongo read-only audit notes.
- Platform Wave B collection-name contract tests for legacy Mongoose compatibility.
- ticket config BSON compatibility and repository contract foundation.
- poll BSON compatibility, repository contract foundation, and gated create/vote/owner-menu/result/chart/export handlers.
- economy `coins`/`gift_changes`/XP/work BSON compatibility, read-only query repository, gated `/代幣查詢` handler, gated `/代幣排行榜` PNG leaderboard, gated `my-profile` profile PNG, gated `/簽到` sign-in write slice, gated `coin-related-settings` config write slice, and gated `/代幣增加` admin coin write slice.
- gated `打工系統` command schema, legacy dashboard-redirect UI for `新增打工事項`, legacy-style `打工介面` list/detail/start UI, and admin setup/delete/energy flows with explicit work repository writes.
- gated read-only `/警告紀錄` warning-history lookup with legacy embed text and safer permission/member-cache handling.
- gated `/警告設定` warning escalation config command with rollback-compatible `errors_sets` writes.
- gated `/警告清除` and `/警告全部清除` warning-removal commands with legacy embeds, best-effort DMs, and `warndbs` mutations.
- gated `/警告` warning-issue command with legacy embeds/DMs, `warndbs` appends, role hierarchy checks, and configured kick/ban threshold actions.
- gated `/刪除訊息` message cleanup command with legacy permission gates, ephemeral embeds, 1000-message cap, and 14-day Discord delete cutoff handling.
- gated `/刪除資料` destructive config cleanup command with the legacy warning select UI and guild-scoped deletes for the selected legacy config collection.
- gated `/翻譯` utility command with legacy loading/final embed shape and safe external-provider error handling.
- gated `/兌換` redeem-code command with legacy ephemeral success/error embeds and rollback-compatible `codes`/`chatgpt_gets` writes.
- gated `/自動通知列表` and `/自動通知刪除` config-maintenance commands with rollback-compatible `cron_sets` reads/deletes.
- gated `/set-log-channel` logging-configuration command and `loggin_create`-compatible select route; event log emitters remain disabled.
- gated read-only `/扭蛋獎池查詢` prize-pool query with legacy embed text and rollback-compatible `gifts`/`gift_changes` reads.
- gated `/扭蛋獎池增加` prize add command with legacy Manage Messages permission, ephemeral success/error embeds, and rollback-compatible `gifts` inserts.
- gated `/扭蛋獎品編輯` prize edit command with legacy Manage Messages permission, ephemeral success/error embeds, and one-row `gifts` replacement by `{guild,gift_name}`.
- gated `/扭蛋獎池刪除` prize delete command with legacy Manage Messages permission, success/error embeds, and one-row `gifts` deletion by `{guild,gift_name}`.
- gated config-only `/公告頻道設置` command with legacy subcommands, embeds, and rollback-compatible `guilds`/`ann_all_sets` writes.
- gated bound announcement message relay from legacy `ann_message.js`, disabled by default and requiring explicit gateway/message-content flags.
- gated config-only `/聊天經驗設定` and `/聊天經驗刪除` commands with legacy embed/preview UI and rollback-compatible `text_xp_channels` writes.
- gated `/聊天經驗身分組設定` and `/語音經驗身分組設定` reward-role config commands with legacy pagination UI and rollback-compatible `chat_roles`/`voice_roles` writes.
- gated disabled-response `/聊天經驗` and `/語音經驗` commands that preserve the legacy replacement message pointing users to `/我的檔案`.
- gated `/經驗值改變` XP admin command with legacy Kick Members permission, success/error embeds, and rollback-compatible `text_xps`/`voice_xps` writes.
- gated config-only `/語音包廂設置` and `/語音包廂刪除` commands with legacy embed UI and rollback-compatible `voice_channels` writes/deletes.
- gated command-only `/上鎖頻道` password command with legacy ephemeral embed UI and rollback-compatible `lock_channels.lock_anser` writes; dynamic voice-room creation, lock components/modals, member moves, and permission overwrites remain disabled.
- gated `guildMemberAdd` join-role assignment from legacy `join_roles`, disabled by default and requiring explicit Gateway + Guild Members intent.
- gated `guildMemberRemove` leave-message delivery from legacy `leave_messages`, disabled by default and requiring explicit Gateway + Guild Members intent.
- dry-run-first `mhcat-economy-reset` one-shot operational tool for the legacy Asia/Taipei daily `coins.today` reset and `work_users.energi` refill/clamp path.
- dry-run-first, lease-gated `mhcat-work-payout` one-shot operational tool for the legacy `handler/gift.js` completed-work payout path.
- Mongo-backed scheduler lease primitive and read-only-by-default diagnostic CLI for future single-owner recurring jobs.
- staging guild command sync apply completed for managed `help`, `info`, and `ping`.
- gateway smoke completed with the gateway explicitly enabled for the smoke invocation.

Not implemented yet:

- prefix commands;
- high-risk slash command feature groups;
- command registration from `cmd/mhcat-bot`;
- most Mongo feature repositories;
- default Mongo index creation;
- production feature writes.
- ticket and poll runtime commands/components unless their explicit feature flags are enabled.
- economy query/rank/sign-in/settings/coin-admin runtime and command sync unless their explicit feature flags are enabled.
- recurring scheduler loops for daily reset, work payout, and automatic notifications; the lease primitive exists but is not wired into bot startup.
- reaction-role moderation commands.
- message/channel/voice logging event emitters.
- announcement relay tag pings; relay messages suppress mentions by default.

Implemented utility commands:

- `/help`
- `/ping`
- `/info bot`
- `/info shard`
- `/info user`
- `/info guild`
- `/代幣查詢` when explicitly enabled with `MHCAT_FEATURE_ECONOMY_QUERY_ENABLED=true`
- `/簽到` when explicitly enabled with `MHCAT_FEATURE_ECONOMY_SIGNIN_ENABLED=true`
- `/coin-related-settings` when explicitly enabled with `MHCAT_FEATURE_ECONOMY_SETTINGS_ENABLED=true`
- `/代幣增加` when explicitly enabled with `MHCAT_FEATURE_ECONOMY_COIN_ADMIN_ENABLED=true`
- `/代幣排行榜` when explicitly enabled with `MHCAT_FEATURE_ECONOMY_COIN_RANK_ENABLED=true`
- `/剪刀石頭布` when explicitly enabled with `MHCAT_FEATURE_ECONOMY_RPS_ENABLED=true`
- `/my-profile` when explicitly enabled with `MHCAT_FEATURE_ECONOMY_PROFILE_ENABLED=true`
- `/打工系統 新增打工事項` dashboard redirect, `/打工系統 打工介面` list/detail/start flow, and work admin setup/delete/energy flows when explicitly enabled with `MHCAT_FEATURE_WORK_ENABLED=true`
- `/警告紀錄` when explicitly enabled with `MHCAT_FEATURE_WARNINGS_ENABLED=true`
- `/警告設定` when explicitly enabled with `MHCAT_FEATURE_WARNING_SETTINGS_ENABLED=true`
- `/警告清除` and `/警告全部清除` when explicitly enabled with `MHCAT_FEATURE_WARNING_REMOVAL_ENABLED=true`
- `/警告` when explicitly enabled with `MHCAT_FEATURE_WARNING_ISSUE_ENABLED=true`
- `/刪除訊息` when explicitly enabled with `MHCAT_FEATURE_MESSAGE_CLEANUP_ENABLED=true`
- `/刪除資料` when explicitly enabled with `MHCAT_FEATURE_DELETE_DATA_ENABLED=true`
- `/翻譯` when explicitly enabled with `MHCAT_FEATURE_TRANSLATE_ENABLED=true`
- `/兌換` when explicitly enabled with `MHCAT_FEATURE_REDEEM_ENABLED=true`
- `/自動通知列表` and `/自動通知刪除` when explicitly enabled with `MHCAT_FEATURE_AUTO_NOTIFICATION_CONFIG_ENABLED=true`
- `/set-log-channel` when explicitly enabled with `MHCAT_FEATURE_LOGGING_CONFIG_ENABLED=true`
- `/扭蛋獎池查詢` when explicitly enabled with `MHCAT_FEATURE_GACHA_PRIZE_LIST_ENABLED=true`
- `/扭蛋獎池增加` when explicitly enabled with `MHCAT_FEATURE_GACHA_PRIZE_CREATE_ENABLED=true`
- `/扭蛋獎品編輯` when explicitly enabled with `MHCAT_FEATURE_GACHA_PRIZE_EDIT_ENABLED=true`
- `/扭蛋獎池刪除` when explicitly enabled with `MHCAT_FEATURE_GACHA_PRIZE_DELETE_ENABLED=true`
- `/抽獎設置` disabled-command parity response when explicitly enabled with `MHCAT_FEATURE_LOTTERY_DISABLED_COMMAND_ENABLED=true`
- `/統計系統查詢` when explicitly enabled with `MHCAT_FEATURE_STATS_QUERY_ENABLED=true`
- `/統計系統創建` when explicitly enabled with `MHCAT_FEATURE_STATS_CREATE_ENABLED=true`
- `/統計身分組人數` when explicitly enabled with `MHCAT_FEATURE_STATS_ROLE_COUNT_ENABLED=true`
- `/統計系統刪除` when explicitly enabled with `MHCAT_FEATURE_STATS_DELETE_ENABLED=true`
- `/公告頻道設置` when explicitly enabled with `MHCAT_FEATURE_ANNOUNCEMENT_CONFIG_ENABLED=true`
- `/公告發送` modal preview/confirm/send flow when explicitly enabled with `MHCAT_FEATURE_ANNOUNCEMENT_SEND_ENABLED=true`
- `/聊天經驗設定` and `/聊天經驗刪除` when explicitly enabled with `MHCAT_FEATURE_TEXT_XP_CONFIG_ENABLED=true`
- `/語音經驗設定` and `/語音經驗刪除` when explicitly enabled with `MHCAT_FEATURE_VOICE_XP_CONFIG_ENABLED=true`
- `/聊天經驗身分組設定` and `/語音經驗身分組設定` when explicitly enabled with `MHCAT_FEATURE_XP_ROLE_CONFIG_ENABLED=true`
- `/聊天經驗` and `/語音經驗` disabled-command replacement responses when explicitly enabled with `MHCAT_FEATURE_XP_PROFILE_DISABLED_COMMANDS_ENABLED=true`
- `/經驗值改變` when explicitly enabled with `MHCAT_FEATURE_XP_ADMIN_ENABLED=true`
- `/經驗值重製` when explicitly enabled with `MHCAT_FEATURE_XP_RESET_ENABLED=true`, gateway enabled, Guild Messages intent enabled, and Message Content intent enabled
- `/語音包廂設置` and `/語音包廂刪除` when explicitly enabled with `MHCAT_FEATURE_VOICE_ROOM_CONFIG_ENABLED=true`
- `/上鎖頻道` when explicitly enabled with `MHCAT_FEATURE_VOICE_ROOM_LOCK_ENABLED=true`, gateway enabled, and Voice State intent enabled
- `/加入身份組設置` and `/加入身份組刪除` when explicitly enabled with `MHCAT_FEATURE_JOIN_ROLE_CONFIG_ENABLED=true`
- `/加入訊息設置` and `/退出訊息設置` when explicitly enabled with `MHCAT_FEATURE_WELCOME_MESSAGE_CONFIG_ENABLED=true`
- `/驗證設置` when explicitly enabled with `MHCAT_FEATURE_VERIFICATION_CONFIG_ENABLED=true`
- `/驗證` captcha flow when explicitly enabled with `MHCAT_FEATURE_VERIFICATION_FLOW_ENABLED=true`
- `/帳號需創建時數` when explicitly enabled with `MHCAT_FEATURE_ACCOUNT_AGE_CONFIG_ENABLED=true`

Implemented event features:

- Bound announcement relay when explicitly enabled with `MHCAT_FEATURE_ANNOUNCEMENT_RELAY_ENABLED=true`, `MHCAT_DISCORD_ENABLE_GATEWAY=true`, `MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true`, and `MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true`.
- Join-role assignment on member join when explicitly enabled with `MHCAT_FEATURE_JOIN_ROLE_ASSIGNMENT_ENABLED=true`, `MHCAT_DISCORD_ENABLE_GATEWAY=true`, and `MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true`.
- Welcome-message delivery on member join when explicitly enabled with `MHCAT_FEATURE_WELCOME_MESSAGE_DELIVERY_ENABLED=true`, `MHCAT_DISCORD_ENABLE_GATEWAY=true`, and `MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true`.
- Leave-message delivery on member leave when explicitly enabled with `MHCAT_FEATURE_LEAVE_MESSAGE_DELIVERY_ENABLED=true`, `MHCAT_DISCORD_ENABLE_GATEWAY=true`, and `MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true`.
- Account-age member join gate when explicitly enabled with `MHCAT_FEATURE_ACCOUNT_AGE_POLICY_ENABLED=true`, `MHCAT_DISCORD_ENABLE_GATEWAY=true`, and `MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true`.

`/help` now uses the legacy embed/select-menu interface instead of the temporary text placeholder. The help menu can display legacy categories and command documentation entries before every linked feature handler is implemented; that mirrors the Node.js bot's help interface and does not mean those feature groups are runtime-complete.

`/info bot` now uses the legacy embed/button interface instead of the temporary text status output. The refresh button uses the legacy `botinfoupdate` custom ID through the typed parser/router path.

`/info shard` uses the legacy embed/button style and `shardinfoupdate` route, but it intentionally shows shard fields immediately instead of copying the old empty initial embed.

`/info user` and `/info guild` use the legacy embed layouts with read-only Discord snapshots. Lookup failures return a red safe error embed and do not expose internal errors.

Not implemented yet:

- role button, remaining economy game/shop/reset writes, text/voice XP accrual, rank cards, gacha draw/shop, gift delivery, lottery creation/join/reroll/stop, announcement relay attachment handling/tag pings, stats rename worker, recurring work scheduler ownership, cron, ChatGPT/chat worker, dashboard, auto-chat features, and logging event emitters.

`/簽到` is a staging-gated write slice, not a production-ready economy rollout. Do not enable it against production until duplicate audits and unique-key/index plans for `coins`/`sign_lists` are complete, and the daily reset is either run by the explicit one-shot tool under an operator process or owned by a future lease-backed scheduler.

`/代幣增加` is a disabled-by-default staging admin write slice. It requires Manage Messages, writes legacy-compatible `coins` rows, rejects negative balances and balances above `999999999`, and must be paired with `MHCAT_FEATURE_ECONOMY_COIN_ADMIN_ENABLED=true` plus `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_ADMIN=true` only against disposable staging data until duplicate audits and production ownership are reviewed.

`/剪刀石頭布` is a disabled-by-default staging game write slice. It writes existing `coins` rows only, rejects missing or insufficient balances, preserves legacy tie/win/loss wager behavior, and must be paired with `MHCAT_FEATURE_ECONOMY_RPS_ENABLED=true` plus `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_RPS=true` only against disposable staging data until duplicate audits and production ownership are reviewed.

`/自動通知列表` and `/自動通知刪除` are config-maintenance commands only. They read/delete legacy `cron_sets` rows behind `MHCAT_FEATURE_AUTO_NOTIFICATION_CONFIG_ENABLED=true` and `MHCAT_COMMAND_SYNC_INCLUDE_AUTO_NOTIFICATION_CONFIG=true`; they do not implement `automatic-notification`, cron setup modals/selects, scheduler ownership, or notification sends.

`/兌換` is disabled by default and is available only when paired with `MHCAT_FEATURE_REDEEM_ENABLED=true` and `MHCAT_COMMAND_SYNC_INCLUDE_REDEEM=true` in staging command sync. It consumes legacy `codes` rows, credits `chatgpt_gets.price`, enforces the legacy 7-day expiry check, and does not enable ChatGPT/autochat message runtime.

Because `/info user` and `/info guild` changed the local command definition shape, staging or production Discord command definitions must be updated through `mhcat-command-sync`; bot startup still does not sync commands.

## Environment

Required for a successful run:

- `MHCAT_DISCORD_TOKEN`
- `MHCAT_MONGODB_URI`
- `MHCAT_MONGODB_DATABASE`

Safe defaults:

- `MHCAT_ENV=development`
- `MHCAT_LOG_LEVEL=info`
- `MHCAT_LOG_FORMAT=text`
- `MHCAT_DISCORD_ENABLE_GATEWAY=false`
- `MHCAT_DISCORD_GATEWAY_CONNECT_TIMEOUT=15s`
- `MHCAT_DISCORD_INTERACTION_TIMEOUT=2500ms`
- `MHCAT_DISCORD_GATEWAY_SMOKE_TEST=false`
- `MHCAT_DISCORD_GATEWAY_SMOKE_TEST_TIMEOUT=30s`
- `MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=false`
- `MHCAT_DISCORD_GUILD_MEMBERS_INTENT=false`
- `MHCAT_DISCORD_GUILD_MESSAGES_INTENT=false`
- `MHCAT_DISCORD_GUILD_MESSAGE_REACTIONS_INTENT=false`
- `MHCAT_DISCORD_VOICE_STATE_INTENT=false`
- `MHCAT_FEATURE_TICKETS_ENABLED=false`
- `MHCAT_FEATURE_POLLS_ENABLED=false`
- `MHCAT_FEATURE_ECONOMY_QUERY_ENABLED=false`
- `MHCAT_FEATURE_ECONOMY_SIGNIN_ENABLED=false`
- `MHCAT_FEATURE_ECONOMY_SETTINGS_ENABLED=false`
- `MHCAT_FEATURE_ECONOMY_COIN_ADMIN_ENABLED=false`
- `MHCAT_FEATURE_ECONOMY_COIN_RANK_ENABLED=false`
- `MHCAT_FEATURE_ECONOMY_RPS_ENABLED=false`
- `MHCAT_FEATURE_ECONOMY_PROFILE_ENABLED=false`
- `MHCAT_FEATURE_WORK_ENABLED=false`
- `MHCAT_FEATURE_WARNINGS_ENABLED=false`
- `MHCAT_FEATURE_WARNING_SETTINGS_ENABLED=false`
- `MHCAT_FEATURE_WARNING_REMOVAL_ENABLED=false`
- `MHCAT_FEATURE_WARNING_ISSUE_ENABLED=false`
- `MHCAT_FEATURE_MESSAGE_CLEANUP_ENABLED=false`
- `MHCAT_FEATURE_DELETE_DATA_ENABLED=false`
- `MHCAT_FEATURE_TRANSLATE_ENABLED=false`
- `MHCAT_FEATURE_BALANCE_QUERY_ENABLED=false`
- `MHCAT_FEATURE_REDEEM_ENABLED=false`
- `MHCAT_FEATURE_AUTO_NOTIFICATION_CONFIG_ENABLED=false`
- `MHCAT_FEATURE_LOGGING_CONFIG_ENABLED=false`
- `MHCAT_FEATURE_GACHA_PRIZE_LIST_ENABLED=false`
- `MHCAT_FEATURE_GACHA_PRIZE_CREATE_ENABLED=false`
- `MHCAT_FEATURE_GACHA_PRIZE_EDIT_ENABLED=false`
- `MHCAT_FEATURE_GACHA_PRIZE_DELETE_ENABLED=false`
- `MHCAT_FEATURE_LOTTERY_DISABLED_COMMAND_ENABLED=false`
- `MHCAT_FEATURE_STATS_QUERY_ENABLED=false`
- `MHCAT_FEATURE_STATS_CREATE_ENABLED=false`
- `MHCAT_FEATURE_STATS_ROLE_COUNT_ENABLED=false`
- `MHCAT_FEATURE_STATS_DELETE_ENABLED=false`
- `MHCAT_FEATURE_ANNOUNCEMENT_CONFIG_ENABLED=false`
- `MHCAT_FEATURE_ANNOUNCEMENT_SEND_ENABLED=false`
- `MHCAT_FEATURE_ANNOUNCEMENT_RELAY_ENABLED=false`
- `MHCAT_FEATURE_TEXT_XP_CONFIG_ENABLED=false`
- `MHCAT_FEATURE_VOICE_XP_CONFIG_ENABLED=false`
- `MHCAT_FEATURE_XP_ROLE_CONFIG_ENABLED=false`
- `MHCAT_FEATURE_XP_PROFILE_DISABLED_COMMANDS_ENABLED=false`
- `MHCAT_FEATURE_XP_ADMIN_ENABLED=false`
- `MHCAT_FEATURE_XP_RESET_ENABLED=false`
- `MHCAT_FEATURE_VOICE_ROOM_CONFIG_ENABLED=false`
- `MHCAT_FEATURE_VOICE_ROOM_LOCK_ENABLED=false`
- `MHCAT_FEATURE_JOIN_ROLE_CONFIG_ENABLED=false`
- `MHCAT_FEATURE_JOIN_ROLE_ASSIGNMENT_ENABLED=false`
- `MHCAT_FEATURE_WELCOME_MESSAGE_CONFIG_ENABLED=false`
- `MHCAT_FEATURE_WELCOME_MESSAGE_DELIVERY_ENABLED=false`
- `MHCAT_FEATURE_LEAVE_MESSAGE_DELIVERY_ENABLED=false`
- `MHCAT_FEATURE_VERIFICATION_CONFIG_ENABLED=false`
- `MHCAT_FEATURE_VERIFICATION_FLOW_ENABLED=false`
- `MHCAT_FEATURE_ACCOUNT_AGE_CONFIG_ENABLED=false`
- `MHCAT_FEATURE_ACCOUNT_AGE_POLICY_ENABLED=false`
- `MHCAT_JOBS_DAILY_RESET_ENABLED=false`
- `MHCAT_JOBS_DAILY_RESET_DRY_RUN=true`
- `MHCAT_JOBS_DAILY_RESET_TIMEOUT=60s`
- `MHCAT_JOBS_WORK_PAYOUT_ENABLED=false`
- `MHCAT_JOBS_WORK_PAYOUT_DRY_RUN=true`
- `MHCAT_JOBS_WORK_PAYOUT_TIMEOUT=60s`
- `MHCAT_JOBS_WORK_PAYOUT_LEASE_NAME=work-payout`
- `MHCAT_SCHEDULER_LEASE_ENABLED=false`
- `MHCAT_SCHEDULER_LEASE_TTL=2m`
- `MHCAT_SCHEDULER_LEASE_TIMEOUT=10s`
- `MHCAT_STAGING_MODE=false`
- `MHCAT_STAGING_REQUIRE_GUILD_SCOPE=true`
- `MHCAT_STAGING_ALLOW_COMMAND_APPLY=false`
- `MHCAT_STAGING_ALLOW_GATEWAY_SMOKE=false`
- `MHCAT_STAGING_SMOKE_TIMEOUT=60s`
- `MHCAT_STAGING_EXPECTED_COMMANDS=help,ping,info`
- `MHCAT_MONGO_CONNECT_TIMEOUT=10s`
- `MHCAT_MONGO_PING_TIMEOUT=5s`
- `MHCAT_SHUTDOWN_TIMEOUT=10s`
- `MHCAT_MONGO_AUDIT_SAMPLE_LIMIT=20`
- `MHCAT_MONGO_AUDIT_LARGE_DOC_BYTES=1048576`
- `MHCAT_MONGO_AUDIT_TIMEOUT=30s`
- `MHCAT_MONGO_INDEX_DRY_RUN=true`
- `MHCAT_MONGO_INDEX_APPLY=false`
- `MHCAT_MONGO_INDEX_ALLOW_UNIQUE=false`
- `MHCAT_MONGO_INDEX_ALLOW_TTL=false`
- `MHCAT_MONGO_INDEX_TIMEOUT=60s`

Command sync variables:

- `MHCAT_DISCORD_APPLICATION_ID`
- `MHCAT_COMMAND_SYNC_SCOPE=guild`
- `MHCAT_COMMAND_SYNC_GUILD_ID`
- `MHCAT_COMMAND_SYNC_DRY_RUN=true`
- `MHCAT_COMMAND_SYNC_ALLOW_DELETE=false`
- `MHCAT_COMMAND_SYNC_ALLOW_BULK_OVERWRITE=false`
- `MHCAT_COMMAND_SYNC_STRICT=true`
- `MHCAT_COMMAND_SYNC_INCLUDE_TICKETS=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_POLLS=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_QUERY=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SIGNIN=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SETTINGS=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_ADMIN=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_RANK=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_RPS=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_PROFILE=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_WORK=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_WARNINGS=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_WARNING_SETTINGS=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_WARNING_REMOVAL=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_WARNING_ISSUE=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_MESSAGE_CLEANUP=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_DELETE_DATA=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_TRANSLATE=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_BALANCE_QUERY=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_REDEEM=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_AUTO_NOTIFICATION_CONFIG=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_LOGGING_CONFIG=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_LIST=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_CREATE=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_EDIT=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_DELETE=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_LOTTERY_DISABLED_COMMAND=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_STATS_QUERY=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_STATS_CREATE=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_STATS_ROLE_COUNT=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_STATS_DELETE=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_CONFIG=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_SEND=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_TEXT_XP_CONFIG=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_VOICE_XP_CONFIG=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_XP_ROLE_CONFIG=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_XP_PROFILE_DISABLED_COMMANDS=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_XP_ADMIN=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_XP_RESET=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_CONFIG=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_LOCK=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_JOIN_ROLE_CONFIG=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_WELCOME_MESSAGE_CONFIG=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_CONFIG=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_FLOW=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_ACCOUNT_AGE_CONFIG=false`

Legacy aliases:

- `TOKEN` -> `MHCAT_DISCORD_TOKEN`
- `MONGOOSE_CONNECTION_STRING` -> `MHCAT_MONGODB_URI`

New `MHCAT_*` variables take precedence over legacy aliases. If both are set and differ, only redacted values may be logged.

## Commands

```bash
make fmt
make test
make vet
make build
make run
make bot-run-no-gateway
make bot-smoke-gateway
make command-sync-dry-run
make economy-reset-dry-run
make work-payout-dry-run
make scheduler-lease-status
make staging-preflight
make staging-command-sync-dry-run
make staging-command-sync-apply-guild-confirmed
make staging-gateway-smoke-confirmed
make mongo-compose-up
make mongo-compose-ps
make mongo-compose-down
make mongo-local-audit
make mongo-audit
make mongo-index-dry-run
make check
```

`make check` runs:

```bash
go fmt ./...
go test ./...
go vet ./...
go build ./cmd/mhcat-bot
go build ./cmd/mhcat-command-sync
go build ./cmd/mhcat-mongo-audit
go build ./cmd/mhcat-mongo-index
go build ./cmd/mhcat-staging-preflight
go build ./cmd/mhcat-economy-reset
go build ./cmd/mhcat-scheduler-lease
go build ./cmd/mhcat-work-payout
```

## Running

With no env, the app exits non-zero with a clear missing config error. It does not connect to Discord, register commands, create Mongo indexes, or write Mongo feature data.

With env configured:

```bash
MHCAT_MONGODB_URI='<uri>' \
MHCAT_MONGODB_DATABASE=mhcat \
MHCAT_DISCORD_TOKEN='<token>' \
MHCAT_DISCORD_ENABLE_GATEWAY=false \
go run ./cmd/mhcat-bot
```

When gateway is disabled, the app creates a Discord session object but does not open the Discord Gateway. Mongo is limited to connect and ping.

To opt into gateway mode:

```bash
MHCAT_MONGODB_URI='<uri>' \
MHCAT_MONGODB_DATABASE=mhcat \
MHCAT_DISCORD_TOKEN='<token>' \
MHCAT_DISCORD_ENABLE_GATEWAY=true \
go run ./cmd/mhcat-bot
```

Gateway mode registers the internal `InteractionCreate` event handler and an event dispatcher shell. Message, reaction, guild-member, and voice-state event handlers are registered only when their explicit intent flags are enabled. It does not register or sync application commands, does not create Mongo indexes, and does not write feature data. Message Content remains disabled unless `MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true` is explicitly set.

One-shot gateway smoke mode is also opt-in:

```bash
MHCAT_MONGODB_URI='<uri>' \
MHCAT_MONGODB_DATABASE=mhcat \
MHCAT_DISCORD_TOKEN='<token>' \
MHCAT_DISCORD_ENABLE_GATEWAY=true \
MHCAT_DISCORD_GATEWAY_SMOKE_TEST=true \
go run ./cmd/mhcat-bot
```

Smoke mode waits for gateway ready or `MHCAT_DISCORD_GATEWAY_SMOKE_TEST_TIMEOUT`, then shuts down cleanly. It does not send messages, register commands, or write Mongo data.

Wave 5.3 additionally requires staging flags for gateway smoke:

```bash
MHCAT_STAGING_MODE=true \
MHCAT_STAGING_ALLOW_GATEWAY_SMOKE=true \
scripts/staging/gateway-smoke.sh
```

## Economy Daily Reset

The legacy Node bot runs a daily economy cron at `00:00 Asia/Taipei` on shard 0. The Go refactor currently provides a one-shot operational command instead of a recurring bot-startup scheduler, because recurring jobs need a lease/owner mechanism before multi-process rollout.

Dry-run is the default and performs no writes:

```bash
go run ./cmd/mhcat-economy-reset --dry-run
```

Apply mode requires both an explicit flag and an env gate:

```bash
MHCAT_JOBS_DAILY_RESET_ENABLED=true \
go run ./cmd/mhcat-economy-reset --apply
```

The command:

- previews or resets `coins.today` for guilds not using rolling sign-in cooldowns;
- previews or refills/clamps `work_users.energi` from `work_sets.get_energy` and `work_sets.max_energy`;
- uses normalized `gift_changes.time` decoding instead of copying the legacy raw `$ne: 0` edge case;
- does not create indexes, repair data, sync commands, or run from `cmd/mhcat-bot`.

Do not run apply against production until duplicate audits and rollback notes are reviewed.

## Work Payout

The legacy Node bot checks completed work jobs every minute in `handler/gift.js` on shard 0. The Go refactor currently provides a one-shot operational command instead of a recurring bot-startup scheduler.

Dry-run is the default and performs no writes:

```bash
go run ./cmd/mhcat-work-payout --dry-run
```

Apply mode requires an explicit flag, the work-payout gate, and scheduler lease ownership:

```bash
MHCAT_JOBS_WORK_PAYOUT_ENABLED=true \
MHCAT_SCHEDULER_LEASE_ENABLED=true \
MHCAT_SCHEDULER_LEASE_OWNER=staging-worker \
go run ./cmd/mhcat-work-payout --apply
```

The command:

- scans due `work_users` rows with the effective legacy `end_time < round(now_seconds)` guard;
- increments `coins.coin` with atomic `$inc` and creates missing balance documents with `$setOnInsert`;
- resets paid `work_users.state` to `待業中`;
- fixes the legacy `gift_change.time == 0` new-balance bug by using `today=1` for daily-reset mode;
- requires the scheduler lease before apply writes;
- does not create indexes, repair data, sync commands, send Discord messages, or run from `cmd/mhcat-bot`.

Do not run apply against production until duplicate audits for `coins`, `work_users`, and `gift_changes` are reviewed and the Node.js bot is no longer owning the same payout loop.

## Scheduler Lease

`internal/adapters/mongo.SchedulerLeaseStore` provides acquire/renew/release primitives on the `mhcat_scheduler_locks` collection. It uses `_id == lock_name` and preserves a monotonic `fence` token across releases.

Status is read-only:

```bash
go run ./cmd/mhcat-scheduler-lease --name daily-reset --action status
```

Write diagnostics require an env gate and explicit apply flag:

```bash
MHCAT_SCHEDULER_LEASE_ENABLED=true \
MHCAT_SCHEDULER_LEASE_OWNER=staging-worker \
go run ./cmd/mhcat-scheduler-lease --name daily-reset --action acquire --apply
```

This is still infrastructure only:

- no recurring scheduler starts from `cmd/mhcat-bot`;
- `mhcat-work-payout --apply` uses the lease, but no recurring bot-startup job uses it yet;
- no lease indexes are created beyond MongoDB's default `_id` index;
- no lease write action runs without `MHCAT_SCHEDULER_LEASE_ENABLED=true --apply`.

## Command Sync

`mhcat-command-sync` is a separate CLI. It is not called by `cmd/mhcat-bot` and does not run on shard ready.

Default usage is dry-run:

```bash
MHCAT_DISCORD_TOKEN='<token>' \
MHCAT_DISCORD_APPLICATION_ID='<application-id>' \
MHCAT_COMMAND_SYNC_SCOPE=guild \
MHCAT_COMMAND_SYNC_GUILD_ID='<guild-id>' \
go run ./cmd/mhcat-command-sync --dry-run
```

Apply mode can modify Discord application commands and requires an explicit flag:

```bash
go run ./cmd/mhcat-command-sync --apply
```

Wave 5.3 apply mode is restricted to staging guild sync. It requires `MHCAT_STAGING_MODE=true`, `MHCAT_STAGING_ALLOW_COMMAND_APPLY=true`, guild scope, and no delete/bulk-overwrite flags. Deletion and bulk overwrite are disabled for staging smoke.

Wave 5.3 ships local command definitions, runtime handlers, and staging guild sync guardrails for `help`, `info` with the `bot` subcommand, and `ping`. These definitions appear in command-sync dry-run output. Bot startup still does not run command sync or mutate Discord application commands.

Wave 5.7 applied the managed `help`, `info`, and `ping` commands to the staging guild only. Bot startup still does not run command sync or mutate Discord application commands. Wave 5.8 and Wave 5.9 change only runtime response UI, so the staging slash command definitions do not need to be re-applied for these fixes.

The read-only `/代幣查詢` command is available only when `MHCAT_FEATURE_ECONOMY_QUERY_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_QUERY=true`; staging preflight and scripts reject unpaired sync/runtime flags.

The `/簽到` command is available only when `MHCAT_FEATURE_ECONOMY_SIGNIN_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SIGNIN=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command writes `coins` and `sign_lists`, so use only isolated staging data until the production duplicate/index/reset blockers in `docs/40-economy-signin.md` are closed.

The `/coin-related-settings` command is available only when `MHCAT_FEATURE_ECONOMY_SETTINGS_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SETTINGS=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command writes `gift_changes` using legacy field names and an atomic patch/update path instead of the legacy delete-then-insert flow. It requires Manage Messages at the command definition and runtime levels.

The `/剪刀石頭布` command is available only when `MHCAT_FEATURE_ECONOMY_RPS_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_RPS=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command writes existing `coins` rows only, rejects missing or insufficient balances, and preserves the legacy game behavior where winning can move a balance above `999999999`; use only disposable staging balances until economy ownership and duplicate audits are reviewed.

The `打工系統` command is available only when `MHCAT_FEATURE_WORK_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_WORK=true`; staging preflight rejects unpaired sync/runtime flags. The current work slices preserve the legacy dashboard redirect for `新增打工事項`, implement legacy-style `打工介面` list/captcha/detail/start UI, and implement `打工系統設定`, `打工事項刪除`, `增加個人精力`, and `增加全體精力` behind explicit admin repository wiring and Manage Messages checks. The start and energy paths can create/update `work_users` through atomic repository methods and do not write coins or payout state. Recurring scheduler ownership and payout idempotency remain pending.

The `/警告紀錄` command is available only when `MHCAT_FEATURE_WARNINGS_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_WARNINGS=true`; staging preflight rejects unpaired sync/runtime flags. This command reads `warndbs` only and does not create, remove, or escalate warnings.

The `/警告設定` command is available only when `MHCAT_FEATURE_WARNING_SETTINGS_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_WARNING_SETTINGS=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command writes only `errors_sets.guild`, `ban_count`, and `move` using duplicate-friendly update/upsert behavior. It does not create warnings, delete messages, kick, ban, or run escalation.

The `/警告清除` and `/警告全部清除` commands are available only when `MHCAT_FEATURE_WARNING_REMOVAL_ENABLED=true`. To include them in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_WARNING_REMOVAL=true`; staging preflight and scripts reject unpaired sync/runtime flags. These commands mutate only `warndbs`, preserve the legacy public success/error embeds, and send legacy-style best-effort DMs. They do not create warnings, delete messages, kick, ban, or run escalation.

The `/警告` command is available only when `MHCAT_FEATURE_WARNING_ISSUE_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_WARNING_ISSUE=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command appends legacy `warndbs.content` entries with Asia/Taipei timestamps, preserves the legacy public success/error embeds and target DM, enforces Manage Messages and moderator-vs-target role hierarchy, and reads `errors_sets` to run configured `停權`/`踢出` threshold actions for existing warning records. Test only against disposable staging warning data because it can kick or ban members when thresholds are met.

The `/刪除訊息` command is available only when `MHCAT_FEATURE_MESSAGE_CLEANUP_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_MESSAGE_CLEANUP=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command deletes recent Discord messages only, requires Manage Messages, requires Administrator above 200 requested messages, refuses more than 1000, uses ephemeral legacy completion/error embeds, and does not write Mongo data. Test only in disposable staging channels.

The `/刪除資料` command is available only when `MHCAT_FEATURE_DELETE_DATA_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_DELETE_DATA=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command requires Manage Messages, presents the legacy destructive select UI, and deletes guild-scoped rows for the selected legacy config target: join messages, leave messages, audit logs, stats config, autochat config, verification config, text/voice XP config, or ticket config. Test only against disposable staging config data.

The `/翻譯` command is available only when `MHCAT_FEATURE_TRANSLATE_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_TRANSLATE=true`; staging preflight rejects unpaired sync/runtime flags. This command calls an external Google Translate-compatible endpoint through a provider port, does not require Message Content intent, and does not touch Mongo feature data.

The `/set-log-channel` command is available only when `MHCAT_FEATURE_LOGGING_CONFIG_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_LOGGING_CONFIG=true`; staging preflight rejects unpaired sync/runtime flags. This command writes only the legacy-compatible `loggings` config and does not enable logging event emitters.

The `/扭蛋獎池查詢` command is available only when `MHCAT_FEATURE_GACHA_PRIZE_LIST_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_LIST=true`; staging preflight rejects unpaired sync/runtime flags. This command reads `gifts` and `gift_changes` only. It does not draw prizes, decrement inventory, mutate coins, send DMs, create indexes, or enable shop behavior.

The `/扭蛋獎池增加` command is available only when `MHCAT_FEATURE_GACHA_PRIZE_CREATE_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_CREATE=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command requires Manage Messages and inserts one legacy `gifts` row after the legacy duplicate-name and prize-count checks. It does not draw prizes, decrement inventory counts, mutate coins outside the prize config row, send DMs, create indexes, or enable shop behavior. Test only against disposable staging gacha data.

The `/扭蛋獎品編輯` command is available only when `MHCAT_FEATURE_GACHA_PRIZE_EDIT_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_EDIT=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command requires Manage Messages and replaces one legacy `gifts` row matching `{guild,gift_name}` with merged submitted/default values. It follows the legacy delete-plus-insert shape and has no transaction/rollback path, so test only against disposable staging gacha data. It does not draw prizes, decrement inventory counts, mutate user coin balances, send DMs, create indexes, or enable shop behavior.

The `/扭蛋獎池刪除` command is available only when `MHCAT_FEATURE_GACHA_PRIZE_DELETE_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_DELETE=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command requires Manage Messages and deletes one legacy `gifts` row matching `{guild,gift_name}`, returning the legacy success/error embeds. It does not draw prizes, decrement inventory counts, mutate coins, send DMs, create indexes, or enable shop behavior. Test only against disposable staging gacha data.

The `/抽獎設置` disabled-command parity response is available only when `MHCAT_FEATURE_LOTTERY_DISABLED_COMMAND_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_LOTTERY_DISABLED_COMMAND=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command preserves the current legacy unavailable embed and does not create lottery rows, send lottery panels, register lottery buttons, write Mongo, or enable old `lotter*` component behavior.

The `/統計系統查詢` command is available only when `MHCAT_FEATURE_STATS_QUERY_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_STATS_QUERY=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command preserves the legacy static stats help embed and does not read/write Mongo, create/delete channels, rename channels, create indexes, or enable `channel_status` scheduler behavior.

The `/統計系統創建` command is available only when `MHCAT_FEATURE_STATS_CREATE_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_STATS_CREATE=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command requires Manage Messages, creates the legacy stats category plus base member/user/bot channels, can add the legacy channel-count/text-count/voice-count stat channels after the base row exists, and writes rollback-compatible `numbers` rows. It does not delete Discord channels, create indexes, or enable the `channel_status` rename scheduler. Test only in an isolated staging guild/database because it creates Discord channels and writes `numbers`.

The `/統計身分組人數` command is available only when `MHCAT_FEATURE_STATS_ROLE_COUNT_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_STATS_ROLE_COUNT=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command requires Manage Messages, requires an existing `/統計系統創建` base config, creates a text or voice stat channel named `<role name>: <member count>`, and replaces the legacy `role_numbers` row for `{guild,role}`. It does not delete old stat channels, create indexes, or enable the `channel_status` rename scheduler. Test only in an isolated staging guild/database because it creates Discord channels and writes `role_numbers`.

The `/統計系統刪除` command is available only when `MHCAT_FEATURE_STATS_DELETE_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_STATS_DELETE=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command requires Manage Messages, deletes legacy `numbers` rows for the guild, and preserves the legacy success/error embeds. It does not delete Discord channels, create indexes, or enable `channel_status`.

The `/公告頻道設置` command is available only when `MHCAT_FEATURE_ANNOUNCEMENT_CONFIG_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_CONFIG=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command writes only the legacy-compatible `guilds.announcement_id` and `ann_all_sets` config rows and does not enable Message Content relay, user-message deletion, or bound announcement sends.

The `/公告發送` command is available only when `MHCAT_FEATURE_ANNOUNCEMENT_SEND_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_SEND=true`; staging preflight and scripts reject unpaired sync/runtime flags. It preserves the legacy modal labels, preview embed, confirmation title, button labels/emojis, missing-config text, and success text, but uses versioned custom IDs and suppresses mentions in the preview and final send as an intentional safety fix for legacy tag-ping behavior.

The bound announcement relay is available only when `MHCAT_FEATURE_ANNOUNCEMENT_RELAY_ENABLED=true` with gateway, Guild Messages intent, and Message Content intent explicitly enabled. It reads existing `ann_all_sets` rows, sends a legacy-style embed in the bound channel, then deletes the original message after the send succeeds. It suppresses mentions in the stored `tag` value and ignores empty-content messages as intentional safety fixes.

The `/聊天經驗設定` and `/聊天經驗刪除` commands are available only when `MHCAT_FEATURE_TEXT_XP_CONFIG_ENABLED=true`. To include them in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_TEXT_XP_CONFIG=true`; staging preflight rejects unpaired sync/runtime flags. These commands write only the legacy-compatible `text_xp_channels` config and do not enable text XP accrual, rank cards, voice XP, Message Content intent, or XP reward behavior.

The `/語音經驗設定` and `/語音經驗刪除` commands are available only when `MHCAT_FEATURE_VOICE_XP_CONFIG_ENABLED=true`. To include them in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_VOICE_XP_CONFIG=true`; staging preflight rejects unpaired sync/runtime flags. These commands write only the legacy-compatible `voice_xp_channels` config, preserve the old visible `背景` option without saving it, and do not enable Voice State intent, voice XP accrual, rank cards, or XP reward behavior.

The `/聊天經驗身分組設定` and `/語音經驗身分組設定` commands are available only when `MHCAT_FEATURE_XP_ROLE_CONFIG_ENABLED=true`. To include them in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_XP_ROLE_CONFIG=true`; staging preflight and scripts reject unpaired sync/runtime flags. These commands add, delete, query, and paginate legacy-compatible `chat_roles` and `voice_roles` reward-role config rows with the legacy misspelled `leavel` field. They require Manage Messages and verify the selected role is assignable by the bot before saving. They do not enable XP accrual, rank cards, automatic role assignment/removal, coin rewards, Message Content intent, Guild Messages intent, Voice State intent, or usage-counter writes.

The `/聊天經驗` and `/語音經驗` disabled-command replacement responses are available only when `MHCAT_FEATURE_XP_PROFILE_DISABLED_COMMANDS_ENABLED=true`. To include them in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_XP_PROFILE_DISABLED_COMMANDS=true`; staging preflight and scripts reject unpaired sync/runtime flags. These commands only preserve the legacy red embed telling users to use `/我的檔案`; they do not read XP collections, render rank cards, award XP, or write Mongo data.

The `/經驗值改變` command is available only when `MHCAT_FEATURE_XP_ADMIN_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_XP_ADMIN=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command requires Kick Members, adjusts one member's `text_xps` or `voice_xps` row with legacy `xp`/`leavel` strings, and sets `voice_xps.leavejoin=leave` when creating a voice profile. Test only against disposable staging XP rows until duplicate audits and XP ownership are reviewed. It does not enable XP accrual, rank cards, automatic role assignment/removal, coin rewards, gateway intents, or usage-counter writes.

The `/經驗值重製` command is available only when `MHCAT_FEATURE_XP_RESET_ENABLED=true`, `MHCAT_DISCORD_ENABLE_GATEWAY=true`, `MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true`, and `MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_XP_RESET=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command preserves the legacy owner-only check, immediate individual text/voice XP deletes, and the full-server `^確認^` message confirmation before deleting all `text_xps` or `voice_xps` rows for a guild. Test only against disposable staging XP rows.

The `/語音包廂設置` and `/語音包廂刪除` commands are available only when `MHCAT_FEATURE_VOICE_ROOM_CONFIG_ENABLED=true`. To include them in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_CONFIG=true`; staging preflight and scripts reject unpaired sync/runtime flags. These commands only write/delete legacy-compatible `voice_channels` config rows and preserve the legacy visible success/error embeds. They do not enable `voiceStateUpdate`, create/move/delete dynamic voice channels, write `voice_channel_ids`, or run lock/password side effects.

The `/上鎖頻道` command is available only when `MHCAT_FEATURE_VOICE_ROOM_LOCK_ENABLED=true`, `MHCAT_DISCORD_ENABLE_GATEWAY=true`, and `MHCAT_DISCORD_VOICE_STATE_INTENT=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_LOCK=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command reads the invoking member's current voice state, verifies the existing `lock_channels` row owner, and replaces the row with legacy-compatible `lock_anser`, `owner`, `text_channel`, and empty `ok_people`. It does not create dynamic rooms, write `voice_channel_ids`, move members, edit permission overwrites, or enable legacy lock buttons/modals.

The `/加入身份組設置` and `/加入身份組刪除` commands are available only when `MHCAT_FEATURE_JOIN_ROLE_CONFIG_ENABLED=true`. To include them in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_JOIN_ROLE_CONFIG=true`; staging preflight rejects unpaired sync/runtime flags. These commands write only the legacy-compatible `join_roles` config and preserve the legacy visible embeds. Automatic member-add role assignment is a separate event path and requires `MHCAT_FEATURE_JOIN_ROLE_ASSIGNMENT_ENABLED=true`, gateway enabled, and Guild Members intent enabled. It does not enable welcome messages, leave messages, verification, or account-age kick behavior.

The `/加入訊息設置` and `/退出訊息設置` commands are available only when `MHCAT_FEATURE_WELCOME_MESSAGE_CONFIG_ENABLED=true`. To include them in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_WELCOME_MESSAGE_CONFIG=true`; staging preflight and scripts reject unpaired sync/runtime flags. `/加入訊息設置` preserves the current legacy dashboard redirect UI and performs no Mongo write. `/退出訊息設置` writes only the legacy-compatible `leave_messages` config and preserves the legacy modal/preview UI. Welcome-message delivery is a separate member-add event path and requires `MHCAT_FEATURE_WELCOME_MESSAGE_DELIVERY_ENABLED=true`, gateway enabled, and Guild Members intent enabled; it reads existing dashboard/legacy `join_messages` rows. Leave-message event delivery is a separate member-remove path and requires `MHCAT_FEATURE_LEAVE_MESSAGE_DELIVERY_ENABLED=true`, gateway enabled, and Guild Members intent enabled. These commands do not enable join-message modal writes, verification, or account-age kick behavior.

Legacy MHCAT-server special welcome output is opt-in through empty-by-default config values: `MHCAT_LEGACY_WELCOME_SPECIAL_GUILD_ID`, `MHCAT_LEGACY_WELCOME_SPECIAL_BOT_ID`, `MHCAT_LEGACY_WELCOME_SPECIAL_CHANNEL_ID`, `MHCAT_LEGACY_WELCOME_SPECIAL_CHAT_CHANNEL_ID`, `MHCAT_LEGACY_WELCOME_SPECIAL_HELP_CHANNEL_ID`, `MHCAT_LEGACY_WELCOME_SPECIAL_BUG_CHANNEL_ID`, and `MHCAT_LEGACY_WELCOME_SPECIAL_SUPPORT_CHANNEL_ID`. Set all seven together only for the actual staging/production target that should use the legacy MHCAT special welcome embed; no special guild/channel IDs are hardcoded in Go.

The `/驗證設置` command is available only when `MHCAT_FEATURE_VERIFICATION_CONFIG_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_CONFIG=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command writes only the legacy-compatible `verifications` config row and preserves the legacy setup response UI.

The `/驗證` captcha flow is available only when `MHCAT_FEATURE_VERIFICATION_FLOW_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_FLOW=true`; staging preflight and scripts reject unpaired sync/runtime flags. The visible flow preserves the legacy `/驗證` defer, `captcha.jpeg` attachment, green `點我進行驗證!` button, `請輸入驗證碼!` modal, success/error embed text, role assignment, and optional nickname template behavior. New Go-generated custom IDs intentionally use a bounded state ID instead of embedding the captcha answer; legacy `<captcha>verification` and `<captcha>ver` IDs are still decoded for live-message compatibility.

The `/帳號需創建時數` command is available only when `MHCAT_FEATURE_ACCOUNT_AGE_CONFIG_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_ACCOUNT_AGE_CONFIG=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command writes only the legacy-compatible `create_hours` config (`guild`, string `hours` seconds, nullable `channel`) and preserves the legacy public defer/edit reply UI, permission text, success embeds, and the legacy typo `發送使用者資運`.

The account-age join gate is separate from the config command and runs only when `MHCAT_FEATURE_ACCOUNT_AGE_POLICY_ENABLED=true`, gateway is enabled, and Guild Members intent is enabled. It reads `create_hours`, DMs too-new members with the legacy bilingual embed, kicks with the legacy reason, optionally logs to the configured channel, and stops later member-add handlers so join-role/welcome behavior does not run after a kick. Unlike legacy unhandled promises, Go awaits kick/log errors and intentionally ignores only non-context DM delivery failures so a closed DM does not bypass the protection.

Staging dry-run:

```bash
go run ./cmd/mhcat-staging-preflight --format text

MHCAT_DISCORD_TOKEN='<staging-token>' \
MHCAT_DISCORD_APPLICATION_ID='<staging-application-id>' \
MHCAT_STAGING_GUILD_ID='<staging-guild-id>' \
scripts/staging/command-sync-dry-run.sh
```

Staging guild apply, only after reviewing dry-run output:

```bash
MHCAT_STAGING_MODE=true \
MHCAT_STAGING_ALLOW_COMMAND_APPLY=true \
MHCAT_DISCORD_TOKEN='<staging-token>' \
MHCAT_DISCORD_APPLICATION_ID='<staging-application-id>' \
MHCAT_STAGING_GUILD_ID='<staging-guild-id>' \
scripts/staging/command-sync-apply-guild.sh
```

This path may create/update the managed base utility commands and any explicitly included managed feature commands in the staging guild only. For account-age config smoke, `/帳號需創建時數` is included only when `MHCAT_COMMAND_SYNC_INCLUDE_ACCOUNT_AGE_CONFIG=true` is paired with `MHCAT_FEATURE_ACCOUNT_AGE_CONFIG_ENABLED=true`. Staging apply must not touch global commands, delete commands, or use bulk overwrite.

`mhcat-staging-preflight` is local-only. It does not contact Discord or MongoDB and exits non-zero until all required staging env and safety flags are set.

## Local Mongo Compose

This repo includes a local MongoDB Compose service for staging smoke and audit tooling when a production or staging database URI should not be used from the host.

```bash
make mongo-compose-up
make mongo-compose-ps
```

Use this local URI for host-side Go commands:

```bash
MHCAT_MONGODB_URI='mongodb://127.0.0.1:27018/mhcat-database?directConnection=true'
MHCAT_MONGODB_DATABASE=mhcat-database
```

The local Compose database is empty by default. It is not a production snapshot and does not create indexes or feature data. Stop it with:

```bash
make mongo-compose-down
```

If Docker is not running, start Docker Desktop or OrbStack before `make mongo-compose-up`.

## Mongo Audit

`mhcat-mongo-audit` is read-only. It lists collections, counts documents, reads current indexes, samples document field/type shapes, and reports missing/unknown catalog collections. It does not create indexes or write documents.
It also reports duplicate logical-key risks for catalogued unique candidates, including `coins_guild_member` and `sign_lists_guild_member`, before any unique index is considered.

The first production read-only inventory is summarized in `docs/26-production-mongo-readonly-audit.md`. It does not include raw document values or credentials.

```bash
MHCAT_MONGODB_URI='<uri>' \
MHCAT_MONGODB_DATABASE=mhcat \
go run ./cmd/mhcat-mongo-audit --format json
```

Useful flags:

- `--sample-limit <n>`
- `--large-doc-bytes <n>`
- `--timeout <duration>`
- `--format text|json`
- `--output <path>`

## Mongo Index Dry-Run

`mhcat-mongo-index` defaults to dry-run and compares the local partial index plan with live indexes. It never drops indexes in Wave 3.

```bash
MHCAT_MONGODB_URI='<uri>' \
MHCAT_MONGODB_DATABASE=mhcat \
go run ./cmd/mhcat-mongo-index --dry-run --format json
```

Index apply is not default. Missing safe indexes can only be created with explicit `--apply`. Unique indexes additionally require `--allow-unique` and a clean duplicate audit. TTL indexes require `--allow-ttl` and a retention ADR/note in the index plan.

Wave 3 still has no feature repositories, no production Mongo feature writes by default, no command registration from bot startup, and no feature parity implementation.

## Custom ID Parsing

Wave 4 adds the shared parser layer for components and modals. New Go-generated IDs use:

```txt
mhcat:v1:<feature>:<action>:<payload>
```

Encoded custom IDs are length-checked against Discord's 100-character `custom_id` limit. Payloads are bounded and must not contain secrets or raw untrusted text. If future feature data is too large or sensitive for a custom ID, the feature should store state separately and encode only a state reference.

Legacy live-message compatibility is handled through explicit decoders for documented high-confidence IDs, including ticket buttons, polls, verification prompts, rank pagination, sign/profile buttons, role buttons, voice-lock prompts, shop/game controls, and setup modals. Ambiguous legacy IDs return typed parse errors instead of being routed through broad substring matching.

The parser is now used by the legacy help menu, ticket setup modal path, poll components, sign-in page buttons, and verification prompt/modal routes. Role-button, most economy, game, and most modal feature behavior are still not implemented.

## Ticket Slice

The current ticket slice implements the legacy private-channel setup/open/close flow behind explicit wiring:

- `私人頻道設置` command definition and handler.
- `私人頻道刪除` command definition and handler.
- Legacy setup modal title and field labels.
- Ticket panel embed with legacy `tic` open button.
- `tickets` config is saved only after valid modal submit, fixing the legacy premature-write bug.
- `tic` creates a private text channel with legacy permission overwrites and welcome UI.
- `del` deletes the ticket channel when the actor is allowed to close it.

Ticket runtime routes can be enabled with:

```txt
MHCAT_FEATURE_TICKETS_ENABLED=true
```

When enabled in the default app runtime, the bot constructs the Mongo `tickets` repository and Discord side-effect ports after Mongo connect. It still does not sync slash commands from bot startup.

Ticket command sync remains separately gated. To include ticket commands in a staging guild dry-run/apply, set:

```txt
MHCAT_STAGING_MODE=true
MHCAT_COMMAND_SYNC_INCLUDE_TICKETS=true
```

The command-sync CLI rejects ticket inclusion outside staging guild scope. Deletion and bulk overwrite remain disabled unless their separate explicit unsafe flags are used, and staging still rejects them.

Still not implemented:

- Production ticket command sync.
- Staging smoke for the full ticket setup/open/close flow.

## Poll Slice

Poll Wave A/B implements the low-risk poll runtime foundation behind explicit flags:

- `投票創建` command definition and handler.
- Legacy-style public poll embed, choice buttons, result button, and owner select menu.
- Versioned Go-generated custom IDs for new poll components, while legacy `poll_<choice>`, `see_result`, and `poll_menu` still decode for live old messages.
- `polls` BSON compatibility with legacy fields, including `join_member[].choise`.
- Repository-level vote add/remove semantics to avoid the old full-array overwrite race.
- Owner toggles for public result, change choice, anonymous, end/reopen, and max-choice selection.
- Result embeds with `file.jpg` chart and `discord.txt` export.
- Owner-menu Excel export as `poll_info.xlsx`, with anonymous export still blocked like legacy.

Poll runtime routes can be enabled with:

```txt
MHCAT_FEATURE_POLLS_ENABLED=true
```

Poll command sync remains separately gated. To include poll commands in a staging guild dry-run/apply, set:

```txt
MHCAT_STAGING_MODE=true
MHCAT_FEATURE_POLLS_ENABLED=true
MHCAT_COMMAND_SYNC_INCLUDE_POLLS=true
```

The command-sync CLI rejects poll inclusion outside staging guild scope. Deletion and bulk overwrite remain disabled unless their separate explicit unsafe flags are used, and staging still rejects them.

Still not implemented:

- Production poll command sync.
- Staging smoke for the full poll create/vote/result/export/menu flow.

## Utility Feature Tests

Run the utility feature pipeline tests with:

```bash
go test ./internal/core/features ./internal/core/services/utility ./internal/discord/features/utility ./internal/discord/commands ./internal/discord/interactions
```

Usage tracking is currently wired through runtime middleware and defaults to no-op behavior. It does not write to MongoDB or increment legacy `all_use_count` yet.

## Mongo Integration Test

Mongo integration tests are skipped by default. To run them:

```bash
MHCAT_RUN_MONGO_INTEGRATION_TESTS=true \
MHCAT_MONGODB_URI='<uri>' \
MHCAT_MONGODB_DATABASE=mhcat \
go test ./internal/adapters/mongo
```

Do not use production credentials for local tests unless the environment is explicitly approved for read-only connectivity checks.
