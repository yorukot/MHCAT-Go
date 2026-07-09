# Staging Smoke Runbook

Status: Wave 5.3 staging-only process. Do not use production token, production application ID, production guild, or production MongoDB.

## Required Env

```bash
export MHCAT_DISCORD_TOKEN='<staging bot token>'
export MHCAT_DISCORD_APPLICATION_ID='<staging application id>'
export MHCAT_STAGING_GUILD_ID='<staging guild id>'
export MHCAT_MONGODB_URI='<staging mongo uri>'
export MHCAT_MONGODB_DATABASE='<staging database>'
```

Optional safety pin:

```bash
export MHCAT_STAGING_ALLOWED_APPLICATION_ID="$MHCAT_DISCORD_APPLICATION_ID"
```

Optional ticket smoke flags:

```bash
export MHCAT_FEATURE_TICKETS_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_TICKETS=true
```

Set both together when testing ticket commands. Do not include ticket commands in command sync if the runtime feature flag is disabled.

Optional economy query smoke flags:

```bash
export MHCAT_FEATURE_ECONOMY_QUERY_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_QUERY=true
```

Set both together when testing `/代幣查詢`. This smoke path is read-only and must not enable sign-in, shop, gacha, work, XP, or economy writes.

Optional economy sign-in smoke flags:

```bash
export MHCAT_FEATURE_ECONOMY_SIGNIN_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SIGNIN=true
```

Set both together only in an isolated staging database when testing `/簽到` and `/簽到列表`. Do not enable production sign-in until duplicate audits for `coins`/`sign_lists`, the unique-key/index plan, and the daily reset job ownership are complete.

Optional economy settings smoke flags:

```bash
export MHCAT_FEATURE_ECONOMY_SETTINGS_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SETTINGS=true
```

Set both together only in an isolated staging database when testing `/coin-related-settings`. This path writes `gift_changes`, which is shared by gacha, sign-in, daily reset, and XP reward behavior.

Optional economy coin-admin smoke flags:

```bash
export MHCAT_FEATURE_ECONOMY_COIN_ADMIN_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_ADMIN=true
```

Set both together only in an isolated staging database when testing `/代幣增加`. This path writes `coins`, requires Manage Messages, and should use disposable balance rows only.

Optional economy coin-rank smoke flags:

```bash
export MHCAT_FEATURE_ECONOMY_COIN_RANK_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_RANK=true
```

Set both together only in an isolated staging database when testing `/代幣排行榜`. This path reads `coins` and renders a PNG leaderboard without writing economy data.

Optional economy profile smoke flags:

```bash
export MHCAT_FEATURE_ECONOMY_PROFILE_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_PROFILE=true
```

Set both together only in an isolated staging database when testing `my-profile`. This path reads `coins`, `gift_changes`, `text_xps`, `voice_xps`, `work_sets`, and `work_users`, and renders a PNG profile card without writing economy data.

Optional work command smoke flags:

```bash
export MHCAT_FEATURE_WORK_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_WORK=true
```

Set both together only in an isolated staging guild/database when testing the currently implemented `打工系統` dashboard redirect, interface/detail/start flow, setup/delete, and energy grant paths. Do not enable production work command sync until scheduler ownership, duplicate audits, and payout idempotency are reviewed.

Optional warning-history smoke flags:

```bash
export MHCAT_FEATURE_WARNINGS_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_WARNINGS=true
```

Set both together only when testing the read-only `/警告紀錄` lookup. This path reads `warndbs` and does not create, remove, or escalate warnings.

Optional warning-settings smoke flags:

```bash
export MHCAT_FEATURE_WARNING_SETTINGS_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_WARNING_SETTINGS=true
```

Set both together only when testing the config-only `/警告設定` path against isolated staging data. This path writes `errors_sets` only and does not create warnings, delete messages, kick, ban, or run escalation.

Optional warning-removal smoke flags:

```bash
export MHCAT_FEATURE_WARNING_REMOVAL_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_WARNING_REMOVAL=true
```

Set both together only when testing `/警告清除` and `/警告全部清除` against isolated staging warning fixtures. This path mutates `warndbs` and sends best-effort DMs, but does not create warnings, delete messages, kick, ban, or run escalation.

Optional warning-issue smoke flags:

```bash
export MHCAT_FEATURE_WARNING_ISSUE_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_WARNING_ISSUE=true
```

Set both together only when testing `/警告` against isolated staging warning fixtures and disposable test members. This path appends `warndbs`, sends best-effort DMs, and can kick or ban when `errors_sets` thresholds are met.

Optional message-cleanup smoke flags:

```bash
export MHCAT_FEATURE_MESSAGE_CLEANUP_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_MESSAGE_CLEANUP=true
```

Set both together only when testing `/刪除訊息` in disposable staging channels. This path deletes recent Discord messages, requires Manage Messages, requires Administrator above 200 requested messages, refuses more than 1000, and writes no Mongo data.

Optional delete-data smoke flags:

```bash
export MHCAT_FEATURE_DELETE_DATA_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_DELETE_DATA=true
```

Set both together only when testing `/刪除資料` against disposable staging config rows. This path deletes selected guild-scoped legacy config rows for join/leave messages, logging, stats, autochat, verification, text/voice XP, or ticket settings.

Optional translate smoke flags:

```bash
export MHCAT_FEATURE_TRANSLATE_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_TRANSLATE=true
```

Set both together only when testing `/翻譯`. This path calls the external translate provider, does not require Message Content intent, and does not write Mongo feature data.

Optional balance-query smoke flags:

```bash
export MHCAT_FEATURE_BALANCE_QUERY_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_BALANCE_QUERY=true
```

Set both together only when testing `/查看餘額`. This path reads `chatgpt_gets.price`, does not require Message Content intent, and does not write Mongo feature data.

Optional redeem smoke flags:

```bash
export MHCAT_FEATURE_REDEEM_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_REDEEM=true
```

Set both together only in an isolated staging database when testing `/兌換`. Seed a safe staging `codes` row with `code`, numeric `price`, and numeric millisecond `time`; the command deletes that row and credits `chatgpt_gets.price`. It does not enable ChatGPT/autochat message runtime or require Message Content intent.

Optional auto-notification config smoke flags:

```bash
export MHCAT_FEATURE_AUTO_NOTIFICATION_CONFIG_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_AUTO_NOTIFICATION_CONFIG=true
```

Set both together only when testing `/自動通知列表` and `/自動通知刪除`. This path reads/deletes `cron_sets` rows and may clean abandoned setup drafts whose `cron` is null or missing. It does not enable `automatic-notification`, the cron modal/select flow, Message Content intent, recurring scheduler ownership, or notification sends.

Optional logging-config smoke flags:

```bash
export MHCAT_FEATURE_LOGGING_CONFIG_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_LOGGING_CONFIG=true
```

Set both together only when testing `/set-log-channel`. This path writes the legacy-compatible `loggings` config row for the staging guild. It does not enable message/channel/voice logging event emitters.

Optional gacha prize-list smoke flags:

```bash
export MHCAT_FEATURE_GACHA_PRIZE_LIST_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_LIST=true
```

Set both together only when testing read-only `/扭蛋獎池查詢`. This path reads `gifts` and `gift_changes`; it does not draw prizes, write coins, decrement inventory, send DMs, or enable shop behavior.

Optional lottery disabled-command smoke flags:

```bash
export MHCAT_FEATURE_LOTTERY_DISABLED_COMMAND_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_LOTTERY_DISABLED_COMMAND=true
```

Set both together only when testing `/抽獎設置` unavailable-response parity. This path performs no lottery creation, no `lotters` write, no public lottery panel send, and no `lotter*` component behavior.

Optional stats query smoke flags:

```bash
export MHCAT_FEATURE_STATS_QUERY_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_STATS_QUERY=true
```

Set both together only when testing `/統計系統查詢`. This path sends the legacy static help embed only; it does not write `Number`/`role_number`, create channels, rename channels, or enable `channel_status`.

Optional stats delete smoke flags:

```bash
export MHCAT_FEATURE_STATS_DELETE_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_STATS_DELETE=true
```

Set both together only when testing `/統計系統刪除` against disposable staging `numbers` rows. This path deletes guild-scoped legacy stats config rows and does not delete Discord channels or enable `channel_status`.

Optional announcement config smoke flags:

```bash
export MHCAT_FEATURE_ANNOUNCEMENT_CONFIG_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_CONFIG=true
```

Set both together only in an isolated staging database when testing `/公告頻道設置`. This path writes legacy-compatible `guilds.announcement_id` and `ann_all_sets` config rows.

For one-time `/公告發送` staging tests, pair `MHCAT_FEATURE_ANNOUNCEMENT_SEND_ENABLED=true` with `MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_SEND=true`. This path preserves the legacy modal/preview/confirm/send UI and sends to `guilds.announcement_id`, but uses versioned confirmation IDs and suppresses tag mentions as an intentional safety fix. It still does not enable Message Content relay or user-message deletion.

Optional text-XP config smoke flags:

```bash
export MHCAT_FEATURE_TEXT_XP_CONFIG_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_TEXT_XP_CONFIG=true
```

Set both together only in an isolated staging database when testing `/聊天經驗設定` and `/聊天經驗刪除`. This path writes `text_xp_channels`; it does not enable Message Content intent, text XP accrual, rank cards, voice XP, or XP rewards.

Optional voice-XP config smoke flags:

```bash
export MHCAT_FEATURE_VOICE_XP_CONFIG_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_VOICE_XP_CONFIG=true
```

Set both together only in an isolated staging database when testing `/語音經驗設定` and `/語音經驗刪除`. This path writes `voice_xp_channels`; it does not enable Voice State intent, voice XP accrual, rank cards, or XP rewards. The legacy `背景` option is visible for command UI parity, but the legacy command did not save it.

Optional XP reward-role config smoke flags:

```bash
export MHCAT_FEATURE_XP_ROLE_CONFIG_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_XP_ROLE_CONFIG=true
```

Set both together only in an isolated staging guild/database when testing `/聊天經驗身分組設定` and `/語音經驗身分組設定`. This path writes `chat_roles` and `voice_roles` config rows with legacy `leavel` strings; it does not enable Message Content intent, Guild Messages intent, Voice State intent, XP accrual, rank cards, automatic role assignment/removal, or coin rewards. Use staging roles below the bot's highest role.

Optional disabled XP profile smoke flags:

```bash
export MHCAT_FEATURE_XP_PROFILE_DISABLED_COMMANDS_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_XP_PROFILE_DISABLED_COMMANDS=true
```

Set both together when testing `/聊天經驗` and `/語音經驗`. This path only returns the legacy replacement embed pointing users to `/我的檔案`; it does not read XP collections, render rank cards, write Mongo, enable accrual, or require gateway intents.

Optional voice-room config smoke flags:

```bash
export MHCAT_FEATURE_VOICE_ROOM_CONFIG_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_CONFIG=true
```

Set both together only in an isolated staging guild/database when testing `/語音包廂設置` and `/語音包廂刪除`. This path writes/deletes `voice_channels` config rows only; it does not enable Voice State intent, dynamic room creation/deletion, `voice_channel_ids`, `/上鎖頻道`, or lock passwords.

Optional join-role config smoke flags:

```bash
export MHCAT_FEATURE_JOIN_ROLE_CONFIG_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_JOIN_ROLE_CONFIG=true
```

Set both together only in an isolated staging guild/database when testing `/加入身份組設置` and `/加入身份組刪除`. This path writes `join_roles`; it does not enable Guild Members intent or automatic member-add role assignment. Use a role below the bot's highest role to match the legacy hierarchy check.

Optional join-role assignment smoke flags:

```bash
export MHCAT_FEATURE_JOIN_ROLE_ASSIGNMENT_ENABLED=true
export MHCAT_DISCORD_ENABLE_GATEWAY=true
export MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true
```

Set these only when testing automatic role assignment in an isolated staging guild. This path registers no slash commands and has no command-sync include flag.

Optional welcome-message config smoke flags:

```bash
export MHCAT_FEATURE_WELCOME_MESSAGE_CONFIG_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_WELCOME_MESSAGE_CONFIG=true
```

Set both together only in an isolated staging guild/database when testing `/加入訊息設置` and `/退出訊息設置`. `/加入訊息設置` is a dashboard redirect only. `/退出訊息設置` writes `leave_messages`; it does not enable Guild Members intent, welcome/leave event sending, join-message modal writes, verification, or account-age kick behavior.

Optional welcome-message delivery smoke flags:

```bash
export MHCAT_FEATURE_WELCOME_MESSAGE_DELIVERY_ENABLED=true
export MHCAT_DISCORD_ENABLE_GATEWAY=true
export MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true
```

Set these only after the staging dashboard or legacy data has a safe `join_messages` row for the staging guild. This event-only path registers no slash commands, has no command-sync include flag, reads `join_messages`, sends a legacy-style welcome embed on `guildMemberAdd`, and performs no Mongo writes.

If testing the legacy MHCAT-server special welcome embed, set all empty-by-default `MHCAT_LEGACY_WELCOME_SPECIAL_*` values together for the staging target. They include the special guild ID, bot ID, send channel ID, and the four visible channel mentions in the legacy description. Do not commit private guild/channel IDs.

Optional verification config smoke flags:

```bash
export MHCAT_FEATURE_VERIFICATION_CONFIG_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_CONFIG=true
```

Set both together only in an isolated staging guild/database when testing `/驗證設置`. This path writes `verifications`; it does not enable `/驗證`, captcha generation, verification buttons/modals, member role assignment, nickname changes, or account-age kick behavior. Use a role below the bot's highest role to match the legacy hierarchy check.

Optional verification flow smoke flags:

```bash
export MHCAT_FEATURE_VERIFICATION_FLOW_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_FLOW=true
```

Set both together only after `/驗證設置` has configured a staging `verifications` row. This path reads `verifications`, sends `captcha.jpeg`, opens the legacy verification modal, adds the configured role, and optionally applies the nickname template. New Go-generated IDs use state IDs and do not embed captcha answers. The in-memory challenge store has a 5-minute TTL and is not shared across processes. Account-age kick remains a separate gated member-add policy.

Optional account-age config smoke flags:

```bash
export MHCAT_FEATURE_ACCOUNT_AGE_CONFIG_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_ACCOUNT_AGE_CONFIG=true
```

Set both together only in an isolated staging guild/database when testing `/帳號需創建時數`. This path writes `create_hours` with legacy-compatible fields: `guild`, string `hours` in seconds, and nullable `channel`. It preserves the legacy public defer/edit UI and permission text, but it does not by itself kick members.

Optional account-age member gate smoke flags:

```bash
export MHCAT_FEATURE_ACCOUNT_AGE_POLICY_ENABLED=true
export MHCAT_DISCORD_ENABLE_GATEWAY=true
export MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true
```

Set these only after `/帳號需創建時數 小時數` has configured a staging `create_hours` row. This path reads `create_hours` during `guildMemberAdd`, sends the legacy bilingual DM, kicks too-new members with the legacy reason, optionally logs to the configured channel, and stops later member-add handlers so join roles/welcome messages do not run after a kick. Use only a disposable staging member/account for smoke testing.

Optional leave-message delivery smoke flags:

```bash
export MHCAT_FEATURE_LEAVE_MESSAGE_DELIVERY_ENABLED=true
export MHCAT_DISCORD_ENABLE_GATEWAY=true
export MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true
```

Set these only after `/退出訊息設置` has configured a staging `leave_messages` row. This path registers no slash commands and performs no Mongo writes while handling member-remove events.

Do not paste real values into committed docs.

## 1. Verify Staging Target

- Confirm token belongs to a staging Discord application.
- Confirm staging guild is not production.
- Confirm `MHCAT_COMMAND_SYNC_SCOPE` is unset or set to `guild`.
- Confirm `MHCAT_DISCORD_MESSAGE_CONTENT_INTENT` is unset or `false`, unless testing bound announcement relay with `MHCAT_FEATURE_ANNOUNCEMENT_RELAY_ENABLED=true`.
- Confirm `MHCAT_COMMAND_SYNC_ALLOW_DELETE=false`.
- Confirm `MHCAT_COMMAND_SYNC_ALLOW_BULK_OVERWRITE=false`.
- If `MHCAT_COMMAND_SYNC_INCLUDE_TICKETS=true`, confirm `MHCAT_FEATURE_TICKETS_ENABLED=true`.
- If `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_QUERY=true`, confirm `MHCAT_FEATURE_ECONOMY_QUERY_ENABLED=true`.
- If `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SIGNIN=true`, confirm `MHCAT_FEATURE_ECONOMY_SIGNIN_ENABLED=true` and the database is isolated staging data.
- If `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SETTINGS=true`, confirm `MHCAT_FEATURE_ECONOMY_SETTINGS_ENABLED=true` and the database is isolated staging data.
- If `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_ADMIN=true`, confirm `MHCAT_FEATURE_ECONOMY_COIN_ADMIN_ENABLED=true`, the database is isolated staging data, and all target `coins` rows are disposable.
- If `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_RANK=true`, confirm `MHCAT_FEATURE_ECONOMY_COIN_RANK_ENABLED=true` and the database is isolated staging data.
- If `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_PROFILE=true`, confirm `MHCAT_FEATURE_ECONOMY_PROFILE_ENABLED=true` and the database is isolated staging data.
- If `MHCAT_COMMAND_SYNC_INCLUDE_WORK=true`, confirm `MHCAT_FEATURE_WORK_ENABLED=true` and that the test accepts the partial work command surface.
- If `MHCAT_COMMAND_SYNC_INCLUDE_WARNINGS=true`, confirm `MHCAT_FEATURE_WARNINGS_ENABLED=true` and the staging guild has safe warning-history fixtures.
- If `MHCAT_COMMAND_SYNC_INCLUDE_WARNING_SETTINGS=true`, confirm `MHCAT_FEATURE_WARNING_SETTINGS_ENABLED=true` and the staging database is isolated because `/警告設定` writes `errors_sets`.
- If `MHCAT_COMMAND_SYNC_INCLUDE_WARNING_REMOVAL=true`, confirm `MHCAT_FEATURE_WARNING_REMOVAL_ENABLED=true` and the staging database has disposable `warndbs` fixtures for warning-removal commands.
- If `MHCAT_COMMAND_SYNC_INCLUDE_WARNING_ISSUE=true`, confirm `MHCAT_FEATURE_WARNING_ISSUE_ENABLED=true`, the staging database has disposable `warndbs` fixtures, and target test members can safely receive warning DMs/kick/ban actions.
- If `MHCAT_COMMAND_SYNC_INCLUDE_MESSAGE_CLEANUP=true`, confirm `MHCAT_FEATURE_MESSAGE_CLEANUP_ENABLED=true`, the target channel is disposable, and test messages can be safely deleted.
- If `MHCAT_COMMAND_SYNC_INCLUDE_DELETE_DATA=true`, confirm `MHCAT_FEATURE_DELETE_DATA_ENABLED=true` and the selected staging config rows are disposable.
- If `MHCAT_COMMAND_SYNC_INCLUDE_TRANSLATE=true`, confirm `MHCAT_FEATURE_TRANSLATE_ENABLED=true` and external translate calls are allowed for the staging bot.
- If `MHCAT_COMMAND_SYNC_INCLUDE_BALANCE_QUERY=true`, confirm `MHCAT_FEATURE_BALANCE_QUERY_ENABLED=true` and the staging database has safe `chatgpt_gets` fixtures or no row.
- If `MHCAT_COMMAND_SYNC_INCLUDE_REDEEM=true`, confirm `MHCAT_FEATURE_REDEEM_ENABLED=true` and the staging database has only disposable `codes` fixtures for `/兌換`.
- If `MHCAT_COMMAND_SYNC_INCLUDE_AUTO_NOTIFICATION_CONFIG=true`, confirm `MHCAT_FEATURE_AUTO_NOTIFICATION_CONFIG_ENABLED=true` and the staging database has safe `cron_sets` fixtures for list/delete.
- If `MHCAT_COMMAND_SYNC_INCLUDE_LOGGING_CONFIG=true`, confirm `MHCAT_FEATURE_LOGGING_CONFIG_ENABLED=true` and the selected log channel is staging-only.
- If `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_LIST=true`, confirm `MHCAT_FEATURE_GACHA_PRIZE_LIST_ENABLED=true` and the staging database has safe gacha fixtures.
- If `MHCAT_COMMAND_SYNC_INCLUDE_LOTTERY_DISABLED_COMMAND=true`, confirm `MHCAT_FEATURE_LOTTERY_DISABLED_COMMAND_ENABLED=true` and that the expected result is only the unavailable embed.
- If `MHCAT_COMMAND_SYNC_INCLUDE_STATS_QUERY=true`, confirm `MHCAT_FEATURE_STATS_QUERY_ENABLED=true` and that the expected result is only the static stats help embed.
- If `MHCAT_COMMAND_SYNC_INCLUDE_STATS_DELETE=true`, confirm `MHCAT_FEATURE_STATS_DELETE_ENABLED=true` and the staging database has disposable `numbers` rows.
- If `MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_CONFIG=true`, confirm `MHCAT_FEATURE_ANNOUNCEMENT_CONFIG_ENABLED=true`, the staging database can safely write `guilds` and `ann_all_sets`, and the expected result is only config changes.
- If `MHCAT_FEATURE_ANNOUNCEMENT_RELAY_ENABLED=true`, confirm `MHCAT_DISCORD_ENABLE_GATEWAY=true`, `MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true`, `MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true`, the staging bound channel is safe for message deletion tests, and tag pings are expected to be suppressed.
- If `MHCAT_COMMAND_SYNC_INCLUDE_TEXT_XP_CONFIG=true`, confirm `MHCAT_FEATURE_TEXT_XP_CONFIG_ENABLED=true` and the staging database can safely write `text_xp_channels`.
- If `MHCAT_COMMAND_SYNC_INCLUDE_VOICE_XP_CONFIG=true`, confirm `MHCAT_FEATURE_VOICE_XP_CONFIG_ENABLED=true` and the staging database can safely write `voice_xp_channels`.
- If `MHCAT_COMMAND_SYNC_INCLUDE_XP_ROLE_CONFIG=true`, confirm `MHCAT_FEATURE_XP_ROLE_CONFIG_ENABLED=true`, the staging database can safely write `chat_roles`/`voice_roles`, and the test roles are below the bot's highest role.
- If `MHCAT_COMMAND_SYNC_INCLUDE_XP_PROFILE_DISABLED_COMMANDS=true`, confirm `MHCAT_FEATURE_XP_PROFILE_DISABLED_COMMANDS_ENABLED=true` and that the expected result is only the replacement embed.
- If `MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_CONFIG=true`, confirm `MHCAT_FEATURE_VOICE_ROOM_CONFIG_ENABLED=true` and the staging database can safely write/delete `voice_channels`.
- If `MHCAT_COMMAND_SYNC_INCLUDE_JOIN_ROLE_CONFIG=true`, confirm `MHCAT_FEATURE_JOIN_ROLE_CONFIG_ENABLED=true`, the staging database can safely write `join_roles`, and the test role is below the bot's highest role.
- If `MHCAT_FEATURE_JOIN_ROLE_ASSIGNMENT_ENABLED=true`, confirm `MHCAT_DISCORD_ENABLE_GATEWAY=true`, `MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true`, the staging database has safe `join_roles` rows, and the target roles are below the bot's highest role.
- If `MHCAT_COMMAND_SYNC_INCLUDE_WELCOME_MESSAGE_CONFIG=true`, confirm `MHCAT_FEATURE_WELCOME_MESSAGE_CONFIG_ENABLED=true` and the staging database can safely write `leave_messages`.
- If `MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_CONFIG=true`, confirm `MHCAT_FEATURE_VERIFICATION_CONFIG_ENABLED=true`, the staging database can safely write `verifications`, and the target role is below the bot's highest role.
- If `MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_FLOW=true`, confirm `MHCAT_FEATURE_VERIFICATION_FLOW_ENABLED=true`, a staging `verifications` row exists or `/驗證設置` will be tested first, the target role is below the bot's highest role, and the test accepts process-local challenge state.
- If `MHCAT_COMMAND_SYNC_INCLUDE_ACCOUNT_AGE_CONFIG=true`, confirm `MHCAT_FEATURE_ACCOUNT_AGE_CONFIG_ENABLED=true` and the staging database can safely write `create_hours`.
- If `MHCAT_FEATURE_ACCOUNT_AGE_POLICY_ENABLED=true`, confirm `MHCAT_DISCORD_ENABLE_GATEWAY=true`, `MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true`, a staging `create_hours` row exists or `/帳號需創建時數 小時數` will be tested first, and a disposable too-new test account/member is available.
- If `MHCAT_FEATURE_LEAVE_MESSAGE_DELIVERY_ENABLED=true`, confirm `MHCAT_DISCORD_ENABLE_GATEWAY=true`, `MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true`, the staging database has a safe `leave_messages` row, and the target channel is staging-only.

Run local-only preflight before any Discord or Mongo action:

```bash
go run ./cmd/mhcat-staging-preflight --format text
```

Expected:

- exits `0` only when all required staging env is present and safe flags are set;
- exits non-zero with a deterministic missing/unsafe check list otherwise;
- does not contact Discord or MongoDB;
- does not print raw token or Mongo URI password.

## 2. Command Sync Dry-Run

```bash
scripts/staging/command-sync-dry-run.sh
```

Expected:

- guild scope only;
- plan includes only managed `help`, `ping`, `info`, plus explicitly included managed feature commands;
- unknown remote commands are skipped;
- no create/update/delete is performed;
- no token is printed.

For ticket staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_TICKETS=true`;
- `MHCAT_FEATURE_TICKETS_ENABLED=true`;
- plan includes managed `私人頻道設置` and `私人頻道刪除`;
- plan still performs no create/update/delete during dry-run.

For economy query staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_QUERY=true`;
- `MHCAT_FEATURE_ECONOMY_QUERY_ENABLED=true`;
- plan includes managed `代幣查詢`;
- plan still performs no create/update/delete during dry-run.

For economy sign-in staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SIGNIN=true`;
- `MHCAT_FEATURE_ECONOMY_SIGNIN_ENABLED=true`;
- plan includes managed `簽到` and `簽到列表`;
- plan still performs no create/update/delete during dry-run.

For economy settings staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SETTINGS=true`;
- `MHCAT_FEATURE_ECONOMY_SETTINGS_ENABLED=true`;
- plan includes managed `coin-related-settings`;
- plan still performs no create/update/delete during dry-run.

For economy coin-admin staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_ADMIN=true`;
- `MHCAT_FEATURE_ECONOMY_COIN_ADMIN_ENABLED=true`;
- plan includes managed `代幣增加`;
- plan still performs no create/update/delete during dry-run.

For economy coin-rank staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_RANK=true`;
- `MHCAT_FEATURE_ECONOMY_COIN_RANK_ENABLED=true`;
- plan includes managed `代幣排行榜`;
- plan still performs no create/update/delete during dry-run.

For economy profile staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_PROFILE=true`;
- `MHCAT_FEATURE_ECONOMY_PROFILE_ENABLED=true`;
- plan includes managed `my-profile`;
- plan still performs no create/update/delete during dry-run.

For work command staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_WORK=true`;
- `MHCAT_FEATURE_WORK_ENABLED=true`;
- plan includes managed `打工系統`;
- plan still performs no create/update/delete during dry-run.

For warning-history staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_WARNINGS=true`;
- `MHCAT_FEATURE_WARNINGS_ENABLED=true`;
- plan includes managed `警告紀錄`;
- plan still performs no create/update/delete during dry-run.

For warning-settings staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_WARNING_SETTINGS=true`;
- `MHCAT_FEATURE_WARNING_SETTINGS_ENABLED=true`;
- plan includes managed `警告設定`;
- plan still performs no create/update/delete during dry-run.

For warning-removal staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_WARNING_REMOVAL=true`;
- `MHCAT_FEATURE_WARNING_REMOVAL_ENABLED=true`;
- plan includes managed `警告清除` and `警告全部清除`;
- plan still performs no create/update/delete during dry-run.

For warning-issue staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_WARNING_ISSUE=true`;
- `MHCAT_FEATURE_WARNING_ISSUE_ENABLED=true`;
- plan includes managed `警告`;
- plan still performs no create/update/delete during dry-run.

For message-cleanup staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_MESSAGE_CLEANUP=true`;
- `MHCAT_FEATURE_MESSAGE_CLEANUP_ENABLED=true`;
- plan includes managed `刪除訊息`;
- plan still performs no create/update/delete during dry-run.

For delete-data staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_DELETE_DATA=true`;
- `MHCAT_FEATURE_DELETE_DATA_ENABLED=true`;
- plan includes managed `刪除資料`;
- plan still performs no create/update/delete during dry-run.

For translate staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_TRANSLATE=true`;
- `MHCAT_FEATURE_TRANSLATE_ENABLED=true`;
- plan includes managed `翻譯`;
- plan still performs no create/update/delete during dry-run.

For balance-query staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_BALANCE_QUERY=true`;
- `MHCAT_FEATURE_BALANCE_QUERY_ENABLED=true`;
- plan includes managed `查看餘額`;
- plan still performs no create/update/delete during dry-run.

For redeem staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_REDEEM=true`;
- `MHCAT_FEATURE_REDEEM_ENABLED=true`;
- plan includes managed `兌換`;
- plan still performs no create/update/delete during dry-run.

For auto-notification config staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_AUTO_NOTIFICATION_CONFIG=true`;
- `MHCAT_FEATURE_AUTO_NOTIFICATION_CONFIG_ENABLED=true`;
- plan includes managed `自動通知列表` and `自動通知刪除`;
- plan still performs no create/update/delete during dry-run.

For logging-config staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_LOGGING_CONFIG=true`;
- `MHCAT_FEATURE_LOGGING_CONFIG_ENABLED=true`;
- plan includes managed `set-log-channel`;
- plan still performs no create/update/delete during dry-run.

For gacha prize-list staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_LIST=true`;
- `MHCAT_FEATURE_GACHA_PRIZE_LIST_ENABLED=true`;
- plan includes managed `扭蛋獎池查詢`;
- plan still performs no create/update/delete during dry-run.

For lottery disabled-command staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_LOTTERY_DISABLED_COMMAND=true`;
- `MHCAT_FEATURE_LOTTERY_DISABLED_COMMAND_ENABLED=true`;
- plan includes managed `抽獎設置`;
- `/抽獎設置` returns the legacy unavailable embed and performs no lottery creation/write.

For stats query staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_STATS_QUERY=true`;
- `MHCAT_FEATURE_STATS_QUERY_ENABLED=true`;
- plan includes managed `統計系統查詢`;
- `/統計系統查詢` returns the legacy static stats help embed and performs no Mongo or channel writes.

For stats delete staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_STATS_DELETE=true`;
- `MHCAT_FEATURE_STATS_DELETE_ENABLED=true`;
- plan includes managed `統計系統刪除`;
- seed a disposable `numbers` row for the staging guild, run `/統計系統刪除`, and verify the row is deleted while Discord channels are untouched.

For text-XP config staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_TEXT_XP_CONFIG=true`;
- `MHCAT_FEATURE_TEXT_XP_CONFIG_ENABLED=true`;
- plan includes managed `聊天經驗設定` and `聊天經驗刪除`;
- plan still performs no create/update/delete during dry-run.

For voice-XP config staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_VOICE_XP_CONFIG=true`;
- `MHCAT_FEATURE_VOICE_XP_CONFIG_ENABLED=true`;
- plan includes managed `語音經驗設定` and `語音經驗刪除`;
- plan still performs no create/update/delete during dry-run.

For XP reward-role config staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_XP_ROLE_CONFIG=true`;
- `MHCAT_FEATURE_XP_ROLE_CONFIG_ENABLED=true`;
- plan includes managed `聊天經驗身分組設定` and `語音經驗身分組設定`;
- plan still performs no create/update/delete during dry-run.

For disabled XP profile staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_XP_PROFILE_DISABLED_COMMANDS=true`;
- `MHCAT_FEATURE_XP_PROFILE_DISABLED_COMMANDS_ENABLED=true`;
- plan includes managed `聊天經驗` and `語音經驗`;
- `/聊天經驗` and `/語音經驗` return the red replacement embed and perform no Mongo writes.

For voice-room config staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_CONFIG=true`;
- `MHCAT_FEATURE_VOICE_ROOM_CONFIG_ENABLED=true`;
- plan includes managed `語音包廂設置` and `語音包廂刪除`;
- plan still performs no create/update/delete during dry-run.

For welcome-message config staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_WELCOME_MESSAGE_CONFIG=true`;
- `MHCAT_FEATURE_WELCOME_MESSAGE_CONFIG_ENABLED=true`;
- plan includes managed `加入訊息設置` and `退出訊息設置`;
- plan still performs no create/update/delete during dry-run.

Welcome-message delivery has no command-sync include flag and should not change the dry-run plan. It is tested only through gateway member-add smoke after a staging `join_messages` row exists.

For verification config staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_CONFIG=true`;
- `MHCAT_FEATURE_VERIFICATION_CONFIG_ENABLED=true`;
- plan includes managed `驗證設置`;
- plan still performs no create/update/delete during dry-run.

For verification flow staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_FLOW=true`;
- `MHCAT_FEATURE_VERIFICATION_FLOW_ENABLED=true`;
- plan includes managed `驗證`;
- plan still performs no create/update/delete during dry-run.

For account-age config staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_ACCOUNT_AGE_CONFIG=true`;
- `MHCAT_FEATURE_ACCOUNT_AGE_CONFIG_ENABLED=true`;
- plan includes managed `帳號需創建時數`;
- plan still performs no create/update/delete during dry-run.

## 3. Optional Staging Guild Apply

Only after dry-run review:

```bash
MHCAT_STAGING_MODE=true \
MHCAT_STAGING_ALLOW_COMMAND_APPLY=true \
scripts/staging/command-sync-apply-guild.sh
```

Expected:

- create/update managed `help`, `ping`, `info`, and explicitly included managed feature commands only;
- no command deletion;
- no bulk overwrite;
- no global command mutation.

If ticket inclusion is enabled, expected:

- create/update managed `help`, `ping`, `info`, `私人頻道設置`, and `私人頻道刪除` only;
- no command deletion;
- no bulk overwrite;
- no global command mutation.

If economy query inclusion is enabled, expected:

- create/update managed `代幣查詢` only in addition to the utility commands;
- no command deletion;
- no bulk overwrite;
- no global command mutation.

If economy sign-in inclusion is enabled, expected:

- create/update managed `簽到` and `簽到列表` only in addition to the utility commands;
- no command deletion;
- no bulk overwrite;
- no global command mutation.

If economy coin-admin inclusion is enabled, expected:

- create/update managed `代幣增加` only in addition to the utility commands;
- no command deletion;
- no bulk overwrite;
- no global command mutation.

If economy coin-rank inclusion is enabled, expected:

- create/update managed `代幣排行榜` only in addition to the utility commands;
- no command deletion;
- no bulk overwrite;
- no global command mutation.

If economy profile inclusion is enabled, expected:

- create/update managed `my-profile` only in addition to the utility commands;
- no command deletion;
- no bulk overwrite;
- no global command mutation.

If work command inclusion is enabled, expected:

- create/update managed `打工系統` only in addition to the utility commands;
- no command deletion;
- no bulk overwrite;
- no global command mutation.

If warning-history inclusion is enabled, expected:

- create/update managed `警告紀錄` only in addition to the utility commands;
- no command deletion;
- no bulk overwrite;
- no global command mutation.

If warning-settings inclusion is enabled, expected:

- create/update managed `警告設定` only in addition to the utility commands;
- no command deletion;
- no bulk overwrite;
- no global command mutation.

If warning-removal inclusion is enabled, expected:

- create/update managed `警告清除` and `警告全部清除` only in addition to the utility commands;
- no command deletion;
- no bulk overwrite;
- no global command mutation.

If warning-issue inclusion is enabled, expected:

- create/update managed `警告` only in addition to the utility commands;
- no command deletion;
- no bulk overwrite;
- no global command mutation.

If message-cleanup inclusion is enabled, expected:

- create/update managed `刪除訊息` only in addition to the utility commands;
- no command deletion;
- no bulk overwrite;
- no global command mutation.

If delete-data inclusion is enabled, expected:

- create/update managed `刪除資料` only in addition to the utility commands;
- no command deletion;
- no bulk overwrite;
- no global command mutation.

If translate inclusion is enabled, expected:

- create/update managed `翻譯` only in addition to the utility commands;
- no command deletion;
- no bulk overwrite;
- no global command mutation.

If balance-query inclusion is enabled, expected:

- create/update managed `查看餘額` only in addition to the utility commands;
- no command deletion;
- no bulk overwrite;
- no global command mutation.

If redeem inclusion is enabled, expected:

- create/update managed `兌換` only in addition to the utility commands;
- no command deletion;
- no bulk overwrite;
- no global command mutation.

If logging-config inclusion is enabled, expected:

- create/update managed `set-log-channel` only in addition to the utility commands;
- no command deletion;
- no bulk overwrite;
- no global command mutation.

If auto-notification config inclusion is enabled, expected:

- create/update managed `自動通知列表` and `自動通知刪除` only in addition to the utility commands;
- no command deletion;
- no bulk overwrite;
- no global command mutation.

If verification-config inclusion is enabled, expected:

- create/update managed `驗證設置` only in addition to the utility commands;
- no command deletion;
- no bulk overwrite;
- no global command mutation.

If verification-flow inclusion is enabled, expected:

- create/update managed `驗證` only in addition to the utility commands;
- no command deletion;
- no bulk overwrite;
- no global command mutation.

If account-age config inclusion is enabled, expected:

- create/update managed `帳號需創建時數` only in addition to the utility commands;
- no command deletion;
- no bulk overwrite;
- no global command mutation.

If XP reward-role config inclusion is enabled, expected:

- create/update managed `聊天經驗身分組設定` and `語音經驗身分組設定` only in addition to the utility commands;
- no command deletion;
- no bulk overwrite;
- no global command mutation.

If voice-room config inclusion is enabled, expected:

- create/update managed `語音包廂設置` and `語音包廂刪除` only in addition to the utility commands;
- no command deletion;
- no bulk overwrite;
- no global command mutation.

## 4. Gateway Smoke

```bash
MHCAT_STAGING_MODE=true \
MHCAT_STAGING_ALLOW_GATEWAY_SMOKE=true \
scripts/staging/gateway-smoke.sh
```

Expected:

- config loads;
- Mongo connects and pings;
- Discord session opens gateway;
- `InteractionCreate` handler is registered;
- ready or timeout is reported;
- no command sync or command registration;
- no Mongo feature write;
- clean shutdown.

## 5. Manual Interactions

With the bot running in staging gateway mode, run:

- `/ping`
- `/help`
- `/help 指令名稱:ping`
- `/info bot`
- click the `/info bot` `更新` button
- `/info shard`
- click the `/info shard` `更新` button
- `/info user`
- `/info guild`

If economy query flags were enabled and command sync apply was reviewed:

- `/代幣查詢`
- `/代幣查詢 使用者:<staging member>`
- verify no-balance users receive the legacy red embed;
- verify users with `coins` data receive the legacy balance embed.

If economy coin-admin flags were enabled and command sync apply was reviewed:

- seed or choose only disposable staging `coins` rows for the target member;
- run `/代幣增加 使用者:<staging member> 增加或減少:增加 數量:1` with a Manage Messages test account;
- verify the legacy success embed title shows the target username, operation, and amount;
- verify a missing target row is created only for add operations, with `coin` incremented and `today` stored compatibly;
- run `/代幣增加 使用者:<staging member> 增加或減少:減少 數量:1` and verify the balance decreases without going negative;
- run a reduce larger than the disposable balance and verify the legacy red `不可減到負數!` embed appears with no balance mutation;
- seed a disposable balance near `999999999`, add past the max, and verify the legacy red max-balance embed appears with no balance mutation;
- verify command access is denied without Manage Messages.

If redeem flags were enabled and command sync apply was reviewed:

- seed one disposable staging `codes` row with a fresh `time` millisecond value and known `price`;
- run `/兌換 代碼:<seeded code>`;
- verify the ephemeral green `成功兌換代碼!` embed appears;
- verify the `codes` row is deleted and `chatgpt_gets.price` for the staging guild increased by the code price;
- run `/兌換 代碼:<missing code>` and verify the legacy red missing-code embed;
- seed an expired code older than 7 days, run `/兌換`, and verify the legacy expired-code embed while the code remains unconsumed.

If translate flags were enabled and command sync apply was reviewed:

- `/翻譯 要的翻譯:你好 目標語言:en`;
- verify the loading embed appears before the final translated embed;
- verify provider failures return a safe red error embed and do not expose raw provider details.

If message-cleanup flags were enabled and command sync apply was reviewed:

- use a disposable staging text channel containing only test messages;
- run `/刪除訊息 刪除數量:1` and verify the ephemeral legacy `清理完成!` embed appears;
- run `/刪除訊息 刪除數量:1001` and verify the legacy maximum-count error appears without deleting messages;
- with a non-Administrator test moderator, run `/刪除訊息 刪除數量:201` and verify the legacy permission error appears;
- optionally run `/刪除訊息 刪除數量:10 使用者:<test user>` and verify only that user's recent test messages are targeted;
- verify no Mongo feature data or indexes changed.

If delete-data flags were enabled and command sync apply was reviewed:

- seed one disposable staging config row for the selected target collection;
- run `/刪除資料` and verify the legacy destructive warning embed and `delete-data` select appear;
- select the seeded target and verify the ephemeral legacy success content appears;
- repeat with a missing target and verify the legacy missing-config content appears;
- verify only the selected guild-scoped disposable rows were deleted and no indexes were created.

If auto-notification config flags were enabled and command sync apply was reviewed:

- create or confirm a safe staging `cron_sets` active row with `guild`, `id`, `cron`, and `channel`;
- optionally create a disposable staging draft row with null or missing `cron`;
- run `/自動通知列表` and verify the legacy list embed includes active rows and does not render pending drafts;
- verify the pending draft row was cleaned from staging data;
- run `/自動通知刪除 id:<active id>` and verify the legacy green delete embed appears;
- run `/自動通知刪除 id:<missing id>` and verify the legacy red missing-id tutorial embed appears;
- verify no `automatic-notification` setup modal, scheduler job, channel send, index creation, or Message Content intent was involved.

If announcement relay was explicitly enabled:

- create or confirm a staging `ann_all_sets` row for a non-production channel;
- send a text-only test message in that bound channel;
- verify the bot sends a legacy-style announcement embed with footer `來自<author tag>的公告`;
- verify the original user message is deleted only after the bot relay appears;
- verify stored `tag` text such as `@everyone` is visible but does not ping;
- verify attachment-only or empty-content messages are not deleted by this Go slice.

If join-role assignment was explicitly enabled:

- create or confirm staging `join_roles` rows for `all_user`, `all_member`, and/or `all_bot`;
- join the staging guild with a test member and, if applicable, a test bot;
- verify only matching roles are assigned;
- verify welcome messages, verification, account-age kick, and leave messages are not expected unless their separate feature flags are explicitly enabled.

If welcome-message delivery was explicitly enabled:

- create or confirm a staging `join_messages` row with safe `channel`, `message_content`, `color`, and optional `img`;
- join the staging guild with a disposable test member;
- verify one legacy-style welcome embed appears in the configured channel;
- verify the author text is `🪂 歡迎加入 <guild name>`;
- verify `(MEMBERNAME)`, `{MEMBERNAME}`, `{membername}`, `(TAG)`, `{TAG}`, and `{tag}` replacements match the saved template;
- verify only the joining member mention is allowed and everyone/role mentions do not ping;
- verify no command registration, Mongo write, verification, leave send, or account-age kick happened unless those separate feature flags are explicitly enabled.

If leave-message delivery was explicitly enabled:

- configure `/退出訊息設置` in a staging-only channel;
- remove a test member from the staging guild;
- verify one legacy-style leave embed appears in the configured channel;
- verify `(MEMBERNAME)`, `{MEMBERNAME}`, `(ID)`, and `{ID}` replacements match the saved template;
- verify no command registration, Mongo write, welcome send, verification, or account-age kick happened unless those separate feature flags are explicitly enabled.

If verification flow was explicitly enabled:

- run `/驗證設置` first if the staging `verifications` row does not already exist;
- run `/驗證`;
- verify the response attaches `captcha.jpeg`;
- verify the green `點我進行驗證!` button appears with the arrow emoji;
- click the button and verify the modal title is `請輸入驗證碼!`;
- submit a wrong answer and verify the legacy wrong-answer error appears;
- run `/驗證` again, click the button, submit the correct answer from the generated staging image, and verify the legacy success embed appears;
- verify the configured staging role was added;
- if a rename template was configured, verify `{name}` is replaced with the member username;
- verify owner nickname changes return the legacy owner error if tested with a guild owner account;
- verify no account-age kick, command registration, Mongo index creation, or usage-counter write happened unless the account-age policy is explicitly enabled for a separate member-add smoke.

If account-age config smoke was explicitly enabled:

- run `/帳號需創建時數 小時數` with a safe threshold such as `24`;
- verify the command publicly defers/edits and returns the legacy success embed title `<a:green_tick:994529015652163614>群組防護系統`;
- optionally run `/帳號需創建時數 被踢出資訊頻道` with a staging-only log channel;
- verify the success embed preserves the legacy wording `已為您設定當未達創建時數時會在:` and typo `發送使用者資運`;
- optionally run `/帳號需創建時數 被踢出資訊頻道刪除` and `/帳號需創建時數 創建時數刪除` only against staging data;
- verify no member kick happens unless the separate account-age member gate is explicitly enabled.

If account-age member gate smoke was explicitly enabled:

- confirm `/帳號需創建時數 小時數` has created a staging `create_hours` row;
- join the staging guild with a disposable account younger than the configured threshold, or lower the threshold only in a throwaway staging guild where the outcome is safe;
- verify the disposable member receives the legacy bilingual DM when DMs are open;
- verify the member is kicked with the legacy reason;
- if a log channel was configured, verify the log embed title is `低於管理員所設定的時數` and contains the account creation timestamp field;
- verify join-role/welcome side effects do not run after the account-age kick.

If ticket flags were enabled and command sync apply was reviewed:

- `/私人頻道設置`
- submit the setup modal with a safe test title/content/color;
- click the `tic` ticket open button;
- verify a private text channel is created under the selected category;
- verify the welcome embed and `del` button appear;
- click `del` from the ticket channel;
- verify the channel is deleted or the legacy denial embed appears when expected.

If work command flags were enabled and command sync apply was reviewed:

- `/打工系統 新增打工事項`;
- verify the legacy red dashboard-redirect embed appears;
- verify the link button points to the staging guild dashboard path;
- `/打工系統 打工介面`;
- if captcha is enabled, verify the legacy captcha modal appears and wrong answers return the legacy captcha error content;
- verify the legacy-style work-list embed, role-filtered job buttons, and job-detail embed;
- press the detail view's `確認打工` button in an isolated staging guild and verify the legacy success embed or busy/energy error appears;
- with an admin test account, run `/打工系統 打工系統設定` and verify the legacy green setup embed and `work_sets` staging document change;
- create/remove a staging-only work item through the dashboard or fixture, then run `/打工系統 打工事項刪除` and verify the legacy delete embed and `work_somethings` staging row removal;
- run `/打工系統 增加個人精力` for a staging test user and verify `work_users.energi` is clamped to the configured maximum;
- run `/打工系統 增加全體精力` only in a small isolated staging guild and verify existing `work_users` rows are clamped;
- verify no coin payout, payout state, index, scheduler lease, or production collection changes occurred.

If logging-config flags were enabled and command sync apply was reviewed:

- `/set-log-channel channel:<staging log channel>`;
- verify the legacy yellow setup embed and log-type select appear;
- choose one or more staging log types;
- verify `loggings.channel_id`, `message_update`, `message_delete`, `channel_update`, and `member_voice_update` reflect the selection in the staging database;
- verify no message/channel/voice log event is emitted by the Go bot from this config slice.

If XP reward-role config flags were enabled and command sync apply was reviewed:

- run `/聊天經驗身分組設定 增加` with a staging role below the bot's highest role, a safe level, and both `到達等級後自動刪除身分組` choices across test rows;
- verify the legacy green add/modify embed appears and the staging `chat_roles` row stores `guild`, string `leavel`, `role`, and `delete_when_not`;
- run `/聊天經驗身分組設定 設定查詢` and verify the legacy query embed plus `上一頁`/`下一頁` buttons when there are enough seeded rows;
- run `/聊天經驗身分組設定 刪除` for the staged level/role and verify the legacy green delete embed and row removal;
- repeat the same add/query/delete path for `/語音經驗身分組設定` and `voice_roles`;
- verify unassignable roles, missing deletes, and over-limit seeded data return legacy red errors;
- verify no XP accrual, rank rendering, automatic role assignment/removal, coin reward, gateway intent, index, or usage-counter write happened.

If voice-room config flags were enabled and command sync apply was reviewed:

- run `/語音包廂設置` with a staging voice or stage channel, a `設定頻道名稱` template containing `{name}`, `是否予許房主上鎖`, and optional `設定人數上限`;
- verify the legacy green setup embed appears;
- verify the staging `voice_channels` row contains `guild`, `ticket_channel`, `limit`, `name`, `parent`, and `lock`;
- run `/語音包廂刪除` for the same trigger channel and verify the legacy delete embed appears and the matching staging rows are removed;
- optionally configure a second trigger under a disposable staging category, run `/語音包廂刪除` for the category, and verify only rows with that `parent` are removed;
- verify no Voice State intent, dynamic channel creation/deletion, `voice_channel_ids`, `lock_channel`, `/上鎖頻道`, or usage-counter write happened.

Verify:

- response arrives through the interaction response path;
- `/info bot` uses the legacy system embed and the refresh button updates through `botinfoupdate`;
- `/info shard` shows shard fields immediately and the refresh button updates through `shardinfoupdate`;
- `/info user` uses the legacy user-information embed without leaking lookup errors;
- `/info guild` uses the legacy server-information embed without leaking lookup errors;
- no duplicate initial response;
- no raw internal error displayed;
- no Message Content intent required;
- no command deletion happened;
- no Mongo feature write happened.
  - Exception: ticket smoke writes the legacy-compatible `tickets` config only after successful modal submit.
  - Exception: economy coin-admin smoke writes disposable staging `coins` rows only.
  - Exception: logging-config smoke writes the legacy-compatible `loggings` config only after the setup select is submitted.
  - Exception: delete-data smoke deletes selected disposable staging config rows only.
  - Exception: auto-notification config smoke deletes selected `cron_sets` rows and abandoned pending drafts only.
  - Exception: voice-room config smoke writes/deletes legacy-compatible `voice_channels` rows only.

## 6. Record Result

Record sanitized result locally under `.smoke/` if useful. Do not store tokens, interaction tokens, private user content, private channel IDs, or raw Mongo URIs.
