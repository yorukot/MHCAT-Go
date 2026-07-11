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

Optional global slash usage-tracking smoke flag:

```bash
export MHCAT_FEATURE_USAGE_TRACKING_ENABLED=true
```

Enable this only with disposable staging `all_use_counts` rows after stopping the Node `events/SlashCommands.js` counter owner. There is no command-sync companion flag. The preflight should emit `usage-tracking-runtime-readiness status=warn`. All later statements that expect no usage-counter write assume this gate is unset or `false`; when it is `true`, expect one global increment per slash attempt and no increment for components, modals, or autocomplete interactions.

Optional ticket smoke flags:

```bash
export MHCAT_FEATURE_TICKETS_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_TICKETS=true
```

Set both together when testing ticket commands. Do not include ticket commands in command sync if the runtime feature flag is disabled, and stop every Node/extra-Go ticket owner first. Follow the canonical [ticket staging checklist](74-ticket.md#staging-smoke).

Optional poll smoke flags:

```bash
export MHCAT_FEATURE_POLLS_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_POLLS=true
export MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true
```

Set the runtime and command-sync flags together only after stopping every Node/extra-Go poll owner. Enable the Guild Members privileged intent on the staging application as well as the env flag for exact member totals and export names. Follow the canonical [poll staging checklist](75-poll.md#staging-smoke), including its duplicate/type audit, legacy-message migration, failure compensation, concurrency, and rollback cases.

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

Optional daily-reset scheduler smoke flags:

```bash
export MHCAT_FEATURE_DAILY_RESET_SCHEDULER_ENABLED=true
export MHCAT_JOBS_DAILY_RESET_ENABLED=true
export MHCAT_DISCORD_ENABLE_GATEWAY=true
export MHCAT_SCHEDULER_LEASE_ENABLED=true
export MHCAT_SCHEDULER_LEASE_OWNER=staging-reset-bot-a
export MHCAT_SCHEDULER_LEASE_TTL=2m
export MHCAT_SCHEDULER_LEASE_TIMEOUT=10s
export MHCAT_JOBS_DAILY_RESET_TIMEOUT=60s
```

This path has no command-sync flag. It writes `coins.today`, `work_users.energi`, and `mhcat_scheduler_locks` across the staging database at `00:00 Asia/Taipei`. Stop Node `handler/cron.js`, use only disposable economy/work fixtures, and give each Go replica a unique owner. The one-shot `mhcat-economy-reset --apply` uses the same `daily-reset` lease; its dry-run remains lease-free.

Optional recurring work-payout smoke flags:

```bash
export MHCAT_FEATURE_WORK_PAYOUT_SCHEDULER_ENABLED=true
export MHCAT_JOBS_WORK_PAYOUT_ENABLED=true
export MHCAT_JOBS_WORK_PAYOUT_LEASE_NAME=work-payout-staging
export MHCAT_JOBS_WORK_PAYOUT_TIMEOUT=60s
export MHCAT_DISCORD_ENABLE_GATEWAY=true
export MHCAT_SCHEDULER_LEASE_ENABLED=true
export MHCAT_SCHEDULER_LEASE_OWNER=staging-payout-bot-a
export MHCAT_SCHEDULER_LEASE_TTL=2m
export MHCAT_SCHEDULER_LEASE_TIMEOUT=10s
```

This path has no command-sync flag. It writes due `coins`/`work_users` rows and `mhcat_scheduler_locks` every minute. `MHCAT_JOBS_WORK_PAYOUT_DRY_RUN` does not apply to the worker. Stop Node `handler/gift.js`, use disposable fixtures, audit duplicate balances and marker shapes, and give every Go replica a unique owner. CLI apply must use the same lease name.

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

Optional economy coin-reset smoke flags:

```bash
export MHCAT_FEATURE_ECONOMY_COIN_RESET_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_RESET=true
export MHCAT_DISCORD_ENABLE_GATEWAY=true
export MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true
export MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true
```

Set all five together only in an isolated staging guild/database when testing `/代幣重製`. This owner-only destructive path deletes or divides all guild `coins` rows only after the same-channel `^確認^` message, so use disposable balances only.

Optional economy RPS smoke flags:

```bash
export MHCAT_FEATURE_ECONOMY_RPS_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_RPS=true
```

Set both together only in an isolated staging database when testing `/剪刀石頭布`. This path writes existing `coins` rows, so use disposable balances only.

Optional economy game smoke flags:

```bash
export MHCAT_FEATURE_ECONOMY_GAME_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_GAME=true
```

Set both together only in an isolated staging database when testing `/代幣遊戲`. This path writes two-player `coins` wagers transactionally and uses process-local component/transition session state, so use disposable balances on a transaction-capable replica set or sharded cluster only. Verify the knowledge acceptance/start/reveal timing, carried countdown, higher/lower draw delay, knowledge/blackjack timeout payouts, removed components, and graceful shutdown as described in `docs/67-economy-game-lifecycle.md`. A failed settlement must not leave one player's balance changed; do not manually retry an unknown commit result.

Optional economy shop smoke flags:

```bash
export MHCAT_FEATURE_ECONOMY_SHOP_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SHOP=true
```

Set both together only in an isolated staging database when testing `/代幣商店`. This path writes `ghps`, subtracts `coins`, can add roles, and can DM prize codes, so use disposable shop and balance rows only.

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

Set both together only in an isolated staging guild/database when testing the currently implemented `打工系統` dashboard redirect, interface/detail/start flow, setup/delete, and energy grant paths. One-shot payout has atomic retry markers, but do not enable production work command sync until recurring ownership, duplicate/marker audits, and isolated payout smoke are complete.

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

Set both together only in an isolated staging database when testing `/兌換`. Seed a safe staging `codes` row with `code`, numeric `price`, and numeric millisecond `time`; the command deletes that row and credits `chatgpt_gets.price`. It does not itself enable either auto-chat runtime or require Message Content intent.

Optional auto-chat local fallback smoke flags:

```bash
export MHCAT_FEATURE_AUTOCHAT_FALLBACK_ENABLED=true
export MHCAT_DISCORD_ENABLE_GATEWAY=true
export MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true
export MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true
```

This event-only path registers no commands and writes no Mongo data. Configure a disposable staging `chats` channel first, seed no `chatgpt_gets` row or a negative/malformed `price`, and verify a human `你好` message receives the legacy local reply. Also verify `price: 0` and positive prices remain silent. Do not run the Node and Go Chatbot handlers against the same guild during this smoke test.

Optional paid auto-chat handoff smoke flags:

```bash
export MHCAT_FEATURE_AUTOCHAT_PAID_HANDOFF_ENABLED=true
export MHCAT_AUTOCHAT_PAID_OWNERSHIP_CONFIRMED=true
export MHCAT_DISCORD_ENABLE_GATEWAY=true
export MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true
export MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true
```

Set the ownership acknowledgment only after confirming a transaction-capable replica-set/sharded staging Mongo deployment, exactly one row per guild in `chats`/`chatgpts`/`chatgpt_gets`, a compatible external worker, and no concurrent Node `events/Chatbot.js` owner. Seed a disposable positive numeric balance, verify one prompt produces one debit and one response after ten seconds, verify an immediate second prompt is not charged, and inspect conversation-ID preservation/reset plus mention warnings. The local fallback can be enabled at the same time to exercise negative/missing balances; zero remains silent.

Optional auto-notification config smoke flags:

```bash
export MHCAT_FEATURE_AUTO_NOTIFICATION_CONFIG_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_AUTO_NOTIFICATION_CONFIG=true
```

Set both together only when testing `automatic-notification`, `/自動通知列表`, and `/自動通知刪除`. This path writes pending `cron_sets` setup rows, completes direct cron or owner-scoped weekday/hour/minute wizard submissions, sends setup previews, reads/deletes `cron_sets` rows, and may clean abandoned setup drafts whose `cron` is null or missing. It does not enable Message Content intent or recurring delivery by itself.

Optional auto-notification delivery smoke flags:

```bash
export MHCAT_FEATURE_AUTO_NOTIFICATION_DELIVERY_ENABLED=true
export MHCAT_DISCORD_ENABLE_GATEWAY=true
export MHCAT_SCHEDULER_LEASE_ENABLED=true
export MHCAT_SCHEDULER_LEASE_OWNER=staging-auto-notification
export MHCAT_SCHEDULER_LEASE_TTL=2m
export MHCAT_SCHEDULER_LEASE_TIMEOUT=10s
```

Delivery may be tested without syncing the config commands. It reads active `cron_sets`, sends their persisted messages to Discord, and writes the `auto-notification-delivery` row in `mhcat_scheduler_locks`. Disable the Node `handler/cron.js` process before setting this gate, and use only a disposable schedule and channel.

Optional logging-config smoke flags:

```bash
export MHCAT_FEATURE_LOGGING_CONFIG_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_LOGGING_CONFIG=true
```

Set both together only when testing `/set-log-channel`. The command remains publicly discoverable, enforces Manage Messages at runtime, and uses the exact owner-scoped ten-minute legacy UI. It writes typed legacy-compatible `loggings` values and does not enable logging events. Use the [logging parity contract](48-logging-config.md) for exact payload, scalar, ownership, and rollback checks.

Optional logging message-event smoke flags:

```bash
export MHCAT_FEATURE_LOGGING_MESSAGE_EVENTS_ENABLED=true
export MHCAT_DISCORD_ENABLE_GATEWAY=true
export MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true
export MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true
```

Set these only when the staging guild has a disposable `loggings` row pointing to a staging-only log channel with `message_update` and/or `message_delete` enabled. This emits message edit/delete embeds from cached gateway data and applies the legacy exact-target-plus-channel delete-audit rule; it does not enable channel or voice logging.

Optional logging channel-event smoke flags:

```bash
export MHCAT_FEATURE_LOGGING_CHANNEL_EVENTS_ENABLED=true
export MHCAT_DISCORD_ENABLE_GATEWAY=true
```

Set these only when the staging guild has a disposable `loggings` row pointing to a staging-only log channel with `channel_update` enabled. This emits topic/permission embeds from cached old channel data, including literal null topics and first-entry action-11 attribution; it does not enable message or voice logging.

Optional logging voice-event smoke flags:

```bash
export MHCAT_FEATURE_LOGGING_VOICE_EVENTS_ENABLED=true
export MHCAT_DISCORD_ENABLE_GATEWAY=true
export MHCAT_DISCORD_VOICE_STATE_INTENT=true
```

Set these only when the staging guild has a disposable `loggings` row pointing to a staging-only log channel with `member_voice_update` enabled. This emits legacy-style voice join/leave embeds from cached gateway member/channel data, including bot members; it does not enable message/channel logging and intentionally ignores moves and mute/deafen-only updates.

Before enabling any logging event flag, stop all Node owners that load `events/LoggingSystem.js` for the same bot/guilds. Go and Node have no logging lease and must never consume these events concurrently.

Optional gacha prize-list smoke flags:

```bash
export MHCAT_FEATURE_GACHA_PRIZE_LIST_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_LIST=true
```

Set both together only when testing read-only `/扭蛋獎池查詢`. This path reads `gifts` and `gift_changes`; it does not draw prizes, write coins, decrement inventory, send DMs, or enable shop behavior.

Optional gacha draw smoke flags:

```bash
export MHCAT_FEATURE_GACHA_DRAW_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_GACHA_DRAW=true
```

Set both together only when testing `/扭蛋` against isolated staging `coins`, `gifts`, and `gift_changes` fixtures. This path charges coins, decrements or deletes auto-delete prize inventory, may send prize-code DMs, and may send notification-channel winner messages.

Optional gacha prize-create smoke flags:

```bash
export MHCAT_FEATURE_GACHA_PRIZE_CREATE_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_CREATE=true
```

Set both together only when testing `/扭蛋獎池增加` against disposable `gifts` rows. This path inserts one prize row after the legacy duplicate and pool-size checks; it does not draw prizes, write user coin balances, decrement inventory counts, send DMs, or enable shop behavior.

Optional gacha prize-edit smoke flags:

```bash
export MHCAT_FEATURE_GACHA_PRIZE_EDIT_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_EDIT=true
```

Set both together only when testing `/扭蛋獎品編輯` against disposable `gifts` rows. This path replaces one prize row by `{guild,gift_name}` using the legacy delete-plus-insert shape; it does not draw prizes, write user coin balances, decrement inventory counts, send DMs, or enable shop behavior.

Optional gacha prize-delete smoke flags:

```bash
export MHCAT_FEATURE_GACHA_PRIZE_DELETE_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_DELETE=true
```

Set both together only when testing `/扭蛋獎池刪除` against disposable `gifts` rows. This path deletes one prize row by `{guild,gift_name}`; it does not draw prizes, write coins, decrement inventory counts, send DMs, or enable shop behavior.

Optional lottery disabled-command smoke flags:

```bash
export MHCAT_FEATURE_LOTTERY_DISABLED_COMMAND_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_LOTTERY_DISABLED_COMMAND=true
```

Set both together only when testing `/抽獎設置` unavailable-response parity. This path performs no lottery creation, no `lotters` write, no public lottery panel send, and no `lotter*` component behavior.

Optional existing-lottery component smoke flags:

```bash
export MHCAT_DISCORD_ENABLE_GATEWAY=true
export MHCAT_FEATURE_LOTTERY_COMPONENTS_ENABLED=true
```

This separate path operates only on existing numeric `<digits>lotter` message buttons and `lotters` rows. It can append `member`, set `end:true`, attach `discord.txt`, and send a winner message to the stored channel. It does not enable or sync `/抽獎設置` and must use disposable copied rows/channels with Node button ownership disabled for the staging guild.

Optional stats query smoke flags:

```bash
export MHCAT_FEATURE_STATS_QUERY_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_STATS_QUERY=true
```

Set both together only when testing `/統計系統查詢`. This path sends the legacy static help embed only; it does not write `Number`/`role_number`, create channels, rename channels, or enable `channel_status`.

Optional stats create smoke flags:

```bash
export MHCAT_FEATURE_STATS_CREATE_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_STATS_CREATE=true
export MHCAT_DISCORD_ENABLE_GATEWAY=true
export MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true
```

Set these together only when testing `/統計系統創建` in an isolated staging guild. This path creates the legacy stats category/base channels, can add channel-count/text-count/voice-count channels, and writes `numbers`; it does not write `role_numbers` rows or enable `channel_status`.

Optional stats role-count smoke flags:

```bash
export MHCAT_FEATURE_STATS_ROLE_COUNT_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_STATS_ROLE_COUNT=true
export MHCAT_DISCORD_ENABLE_GATEWAY=true
export MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true
```

Set these together only when testing `/統計身分組人數` in an isolated staging guild with an existing stats base config. This path creates a text or voice role-count channel and replaces one `role_numbers` row by `{guild,role}`; it does not delete old channels or enable `channel_status`.

Optional stats delete smoke flags:

```bash
export MHCAT_FEATURE_STATS_DELETE_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_STATS_DELETE=true
```

Set both together only when testing `/統計系統刪除` against disposable staging `numbers` rows. This path deletes guild-scoped legacy stats config rows and does not delete Discord channels or enable `channel_status`.

Optional stats rename worker smoke flags:

```bash
export MHCAT_FEATURE_STATS_RENAME_WORKER_ENABLED=true
export MHCAT_DISCORD_ENABLE_GATEWAY=true
export MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true
```

Set these only when testing the event-only `channel_status` parity worker in an isolated staging guild. Seed disposable `numbers`/`role_numbers` rows and matching stat channels, leave the legacy worker stopped for that guild, wait one 20-minute interval, and verify channel names plus stored `*_name`/`channel_name` counters update. Missing channels and Discord/API failures should be skipped/logged, not fatal.

Optional announcement config smoke flags:

```bash
export MHCAT_FEATURE_ANNOUNCEMENT_CONFIG_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_CONFIG=true
```

Set both together only in an isolated staging database when testing `/公告頻道設置`. This path writes legacy-compatible `guilds.announcement_id` and `ann_all_sets` config rows.

For one-time `/公告發送` staging tests, pair `MHCAT_FEATURE_ANNOUNCEMENT_SEND_ENABLED=true` with `MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_SEND=true`. This path preserves the legacy modal/preview/confirm/send UI and sends to `guilds.announcement_id`, but uses versioned confirmation IDs and suppresses tag mentions as an intentional safety fix. It still does not enable Message Content relay or user-message deletion.

Before either command test, stop the matching Node command/modal owner, audit duplicate and scalar-drift keys in `guilds` and `ann_all_sets`, and confirm shared dashboard writers preserve unrelated `guilds` fields. The canonical config/send/relay smoke and rollback sequence is in the [announcement parity contract](76-announcement.md).

Optional anti-scam smoke flags:

```bash
export MHCAT_FEATURE_ANTI_SCAM_CONFIG_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_CONFIG=true
export MHCAT_FEATURE_ANTI_SCAM_REPORT_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_REPORT=true
export MHCAT_REPORT_WEBHOOK_URL=https://example.test/webhook
```

Enable only the families under test. Message deletion is event-only and separately requires `MHCAT_FEATURE_ANTI_SCAM_MESSAGE_DELETE_ENABLED=true`, Gateway, Guild Messages, and Message Content. Stop matching Node ownership, audit `good_webs`/`not_a_good_webs`, and use the canonical [anti-scam staging smoke](77-anti-scam.md#staging-smoke).

Optional birthday command smoke flags:

```bash
export MHCAT_FEATURE_BIRTHDAY_CONFIG_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_BIRTHDAY_CONFIG=true
```

Set both together only with an isolated staging guild/database. Stop Node `/生日系統`, audit `birthday_sets`/`birthdays`, keep the commented delivery block inactive, and run the canonical [birthday staging smoke](78-birthday.md#staging-smoke).

Optional text-XP config smoke flags:

```bash
export MHCAT_FEATURE_TEXT_XP_CONFIG_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_TEXT_XP_CONFIG=true
```

Set both together only in an isolated staging database when testing `/聊天經驗設定` and `/聊天經驗刪除`. This path writes `text_xp_channels`; it does not enable Message Content intent, text XP accrual, rank cards, voice XP, or XP rewards.

Optional text-XP accrual smoke flags:

```bash
export MHCAT_FEATURE_TEXT_XP_ACCRUAL_ENABLED=true
export MHCAT_DISCORD_ENABLE_GATEWAY=true
export MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true
export MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true
```

Use only an isolated staging database with disposable `text_xps`, `coins`, `gift_changes`, `text_xp_channels`, and `chat_roles` rows. This event-only path has no command-sync flag and updates text XP/level on guild messages, sends configured/default level-up announcements or legacy fallback errors when a `text_xp_channels` row exists, applies configured `chat_roles` reward-role changes, and grants legacy XP coin rewards from `gift_changes.xp_multiple` only after the configured announcement path succeeds.

Optional voice-XP config smoke flags:

```bash
export MHCAT_FEATURE_VOICE_XP_CONFIG_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_VOICE_XP_CONFIG=true
```

Set both together only in an isolated staging database when testing `/語音經驗設定` and `/語音經驗刪除`. This path writes `voice_xp_channels`; it does not enable Voice State intent, voice XP accrual, rank cards, or XP rewards. The legacy `背景` option is visible for command UI parity, but the legacy command did not save it.

Optional voice-XP session smoke flags:

```bash
export MHCAT_FEATURE_VOICE_XP_SESSIONS_ENABLED=true
export MHCAT_DISCORD_ENABLE_GATEWAY=true
export MHCAT_DISCORD_VOICE_STATE_INTENT=true
```

Use only an isolated staging database with disposable `voice_xps`, `voice_xp_channels`, `voice_roles`, `coins`, and `gift_changes` rows. This event-only path has no command-sync flag, marks `leavejoin` as users join or leave voice channels, starts the legacy 30-second XP loop for users who join after the process is running, and reconciles existing `leavejoin:"join"` rows on startup.

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

Optional XP admin smoke flags:

```bash
export MHCAT_FEATURE_XP_ADMIN_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_XP_ADMIN=true
```

Set both together only in an isolated staging database when testing `/經驗值改變`. This path writes `text_xps` or `voice_xps` profile rows with legacy string `xp` and `leavel` fields and sets `voice_xps.leavejoin` on voice-profile insert. It does not enable XP accrual, rank cards, automatic role assignment/removal, coin rewards, or gateway intents.

Optional XP reset smoke flags:

```bash
export MHCAT_FEATURE_XP_RESET_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_XP_RESET=true
export MHCAT_DISCORD_ENABLE_GATEWAY=true
export MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true
export MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true
```

Set all five together only in an isolated staging guild/database when testing `/經驗值重製`. Use a guild-owner staging account and only disposable `text_xps`/`voice_xps` rows. Individual reset deletes the selected member row immediately; full-server reset waits for the legacy `^確認^` message before deleting guild XP rows.

Optional XP rank smoke flags:

```bash
export MHCAT_FEATURE_XP_RANK_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_XP_RANK=true
```

Set both together only in an isolated staging database when testing `/聊天排行榜` and `/語音排行榜`. This path reads `text_xps`/`voice_xps` and renders legacy-style `user-info.png` leaderboard pages with legacy rank buttons. It does not enable XP accrual, `/聊天經驗` profile cards, automatic reward roles, coin rewards, gateway intents, indexes, or Mongo writes.

Optional voice-room config smoke flags:

```bash
export MHCAT_FEATURE_VOICE_ROOM_CONFIG_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_CONFIG=true
```

Set both together only in an isolated staging guild/database when testing `/語音包廂設置`, `/語音包廂刪除`, and dynamic voice-room creation. The slash commands write/delete `voice_channels` config rows. When gateway Voice State events are enabled, trigger joins create/move users into dynamic voice rooms, persist `voice_channel_ids`, seed nullable `lock_channels` rows when the config is lockable, and delete empty tracked rooms. Use the [voice-room parity contract](72-voice-room-config.md) for exact values, ownership, and rollback checks.

Optional voice-room lock smoke flags:

```bash
export MHCAT_FEATURE_VOICE_ROOM_LOCK_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_LOCK=true
export MHCAT_DISCORD_ENABLE_GATEWAY=true
export MHCAT_DISCORD_VOICE_STATE_INTENT=true
```

Set all four together only in an isolated staging guild/database when testing `/上鎖頻道` and existing passworded `lock_channels` rows. This path replaces an existing `lock_channels` row for the caller's current voice channel, sends the locked-room password prompt when an unauthorized user joins that channel, disconnects that user, DMs the legacy instructions, and routes the generated prompt button plus `<channel>anser` modal submit to append `ok_people` after a correct password.

Optional join-role config smoke flags:

```bash
export MHCAT_FEATURE_JOIN_ROLE_CONFIG_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_JOIN_ROLE_CONFIG=true
```

Set both together only in an isolated staging guild/database. Stop Node command ownership and audit `join_roles` first. The commands are public, enforce Manage Messages at runtime, write typed config, and do not enable assignment. Follow [81-join-role.md](81-join-role.md).

Optional join-role assignment smoke flags:

```bash
export MHCAT_FEATURE_JOIN_ROLE_ASSIGNMENT_ENABLED=true
export MHCAT_DISCORD_ENABLE_GATEWAY=true
export MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true
```

Set these only for disposable member/bot assignment smoke after confirming cached roles and bot member state. This path registers no slash commands and has no command-sync include flag. Run the complete [join-role contract](81-join-role.md) smoke.

Optional welcome-message config smoke flags:

```bash
export MHCAT_FEATURE_WELCOME_MESSAGE_CONFIG_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_WELCOME_MESSAGE_CONFIG=true
```

Set both together only in an isolated staging guild/database when testing `/加入訊息設置` and `/退出訊息設置`. `/加入訊息設置` is a dashboard redirect only. `/退出訊息設置` writes `leave_messages`; it does not enable Guild Members intent, welcome/leave event sending, join-message modal writes, verification, or account-age kick behavior.

Run the complete config/data portion of [82-welcome-leave.md](82-welcome-leave.md), including exact UI, scalar rows, duplicate alignment, usage, and rollback.

Optional welcome-message delivery smoke flags:

```bash
export MHCAT_FEATURE_WELCOME_MESSAGE_DELIVERY_ENABLED=true
export MHCAT_DISCORD_ENABLE_GATEWAY=true
export MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true
```

Set these only after the staging dashboard or legacy data has a safe `join_messages` row for the staging guild. This event-only path registers no slash commands, has no command-sync include flag, reads `join_messages`, sends a legacy-style welcome embed on `guildMemberAdd`, and performs no Mongo writes.

If testing the legacy MHCAT-server special welcome embed, set all empty-by-default `MHCAT_LEGACY_WELCOME_SPECIAL_*` values together for the staging target. They include the special guild ID, bot ID, send channel ID, and the four visible channel mentions in the legacy description. Do not commit private guild/channel IDs.

Run the complete generic/special welcome portion of [82-welcome-leave.md](82-welcome-leave.md), including cached/missing channels, placeholders, colors, identities, account-age ordering, and rollback.

Optional verification config smoke flags:

```bash
export MHCAT_FEATURE_VERIFICATION_CONFIG_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_CONFIG=true
```

Set both together only in an isolated staging guild/database when testing `/驗證設置`. Stop the Node owner and audit `verifications` first. This public command writes typed duplicate-safe config, enforces Manage Messages at runtime, and does not enable `/驗證` or account-age policy. Follow [80-verification.md](80-verification.md).

Optional verification flow smoke flags:

```bash
export MHCAT_FEATURE_VERIFICATION_FLOW_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_FLOW=true
```

Set both together only after a safe staging `verifications` row exists. Keep smoke single-process: state is guild/user-bound, atomic, five-minute, and not shared across processes. Use disposable roles/members and run the complete [verification contract](80-verification.md) smoke. Account-age remains separate.

Optional account-age config smoke flags:

```bash
export MHCAT_FEATURE_ACCOUNT_AGE_CONFIG_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_ACCOUNT_AGE_CONFIG=true
```

Set both together only in an isolated staging guild/database when testing `/帳號需創建時數`. This path writes `create_hours` with legacy-compatible fields: `guild`, string `hours` in seconds, and nullable `channel`. It preserves the legacy public defer/edit UI and permission text, but it does not by itself kick members. Follow the canonical [account-age staging smoke](79-account-age.md#staging-smoke).

Optional account-age member gate smoke flags:

```bash
export MHCAT_FEATURE_ACCOUNT_AGE_POLICY_ENABLED=true
export MHCAT_DISCORD_ENABLE_GATEWAY=true
export MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true
```

Set these only after `/帳號需創建時數 小時數` has configured a staging `create_hours` row. This path reads `create_hours` during `guildMemberAdd`, sends the legacy bilingual DM, kicks too-new members with the legacy reason, optionally logs to a cached configured channel, and stops later member-add handlers so join roles/welcome messages do not run after a kick. Use only a disposable staging member/account and run [79-account-age.md](79-account-age.md).

Optional role-selection smoke flags:

```bash
export MHCAT_FEATURE_ROLE_SELECTION_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_ROLE_SELECTION=true
export MHCAT_DISCORD_ENABLE_GATEWAY=true
export MHCAT_DISCORD_GUILD_MESSAGE_REACTIONS_INTENT=true
```

Set all four together only in an isolated staging guild/database when testing `/選取身分組-表情符號`, `/選取身分組刪除-表情符號`, and `/選取身分組-按鈕`. The single role-selection runtime flag owns setup commands, modal/buttons, and reaction events together; do not split ownership or run the corresponding Node paths concurrently. This path writes `message_reactions` and `btns`, adds reactions to staging messages, and changes roles on reaction/button use. Use roles below the bot's highest role and follow the canonical [role-selection staging checklist](73-role-selection.md#staging-smoke).

Optional leave-message delivery smoke flags:

```bash
export MHCAT_FEATURE_LEAVE_MESSAGE_DELIVERY_ENABLED=true
export MHCAT_DISCORD_ENABLE_GATEWAY=true
export MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true
```

Set these only after `/退出訊息設置` has configured a staging `leave_messages` row. This path registers no slash commands and performs no Mongo writes while handling member-remove events.

Run the complete leave delivery portion of [82-welcome-leave.md](82-welcome-leave.md), including cached/missing channels, placeholders, colors, raw title, identities, and rollback.

Do not paste real values into committed docs.

## 1. Verify Staging Target

- Confirm token belongs to a staging Discord application.
- Confirm staging guild is not production.
- Confirm `MHCAT_COMMAND_SYNC_SCOPE` is unset or set to `guild`.
- Confirm `MHCAT_DISCORD_MESSAGE_CONTENT_INTENT` is unset or `false`, unless testing bound announcement relay, local/paid auto-chat, economy coin reset confirmation, or XP reset confirmation behind the corresponding explicit feature gate.
- Confirm `MHCAT_COMMAND_SYNC_ALLOW_DELETE=false`.
- Confirm `MHCAT_COMMAND_SYNC_ALLOW_BULK_OVERWRITE=false`.
- If `MHCAT_FEATURE_USAGE_TRACKING_ENABLED=true`, confirm the Node `events/SlashCommands.js` counter owner is stopped, the target `all_use_counts` rows are disposable, and duplicate/null/blank command-name rows have been reviewed.
- If `MHCAT_COMMAND_SYNC_INCLUDE_TICKETS=true`, confirm `MHCAT_FEATURE_TICKETS_ENABLED=true`.
- If `MHCAT_COMMAND_SYNC_INCLUDE_POLLS=true`, confirm `MHCAT_FEATURE_POLLS_ENABLED=true`, `MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true`, the application Guild Members privileged intent is enabled, every Node/extra-Go poll owner is stopped, and duplicate/malformed `polls` rows have been reviewed without applying an index.
- If `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_QUERY=true`, confirm `MHCAT_FEATURE_ECONOMY_QUERY_ENABLED=true`.
- If `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SIGNIN=true`, confirm `MHCAT_FEATURE_ECONOMY_SIGNIN_ENABLED=true` and the database is isolated staging data.
- If `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SETTINGS=true`, confirm `MHCAT_FEATURE_ECONOMY_SETTINGS_ENABLED=true` and the database is isolated staging data.
- If `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_ADMIN=true`, confirm `MHCAT_FEATURE_ECONOMY_COIN_ADMIN_ENABLED=true`, the database is isolated staging data, and all target `coins` rows are disposable.
- If `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_RANK=true`, confirm `MHCAT_FEATURE_ECONOMY_COIN_RANK_ENABLED=true` and the database is isolated staging data.
- If `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_RESET=true`, confirm `MHCAT_FEATURE_ECONOMY_COIN_RESET_ENABLED=true`, `MHCAT_DISCORD_ENABLE_GATEWAY=true`, `MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true`, `MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true`, the tester is the staging guild owner, and the staging database has only disposable `coins` rows for the reset target.
- If `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_RPS=true`, confirm `MHCAT_FEATURE_ECONOMY_RPS_ENABLED=true`, the database is isolated staging data, and all target `coins` rows are disposable.
- If `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_GAME=true`, confirm `MHCAT_FEATURE_ECONOMY_GAME_ENABLED=true`, Mongo supports transactions, Node game ownership is stopped, the database is isolated staging data, and all target `coins` rows are disposable; include knowledge and blackjack timeout cases.
- If `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SHOP=true`, confirm `MHCAT_FEATURE_ECONOMY_SHOP_ENABLED=true`, the database is isolated staging data, all target `ghps`/`coins` rows are disposable, and test roles are below the bot's highest role.
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
- If `MHCAT_FEATURE_AUTOCHAT_FALLBACK_ENABLED=true`, confirm gateway, Guild Messages, and Message Content are enabled; `chats.channel` targets a disposable staging channel; `chatgpt_gets` fixtures are safe; and the Node Chatbot handler is not concurrently active for that guild.
- If `MHCAT_FEATURE_AUTOCHAT_PAID_HANDOFF_ENABLED=true`, confirm `MHCAT_AUTOCHAT_PAID_OWNERSHIP_CONFIRMED=true`, transaction-capable Mongo, clean singleton duplicate audits, a compatible external worker, disposable `chatgpts`/`chatgpt_gets` rows, and exclusive Go MessageCreate ownership.
- If `MHCAT_COMMAND_SYNC_INCLUDE_AUTO_NOTIFICATION_CONFIG=true`, confirm `MHCAT_FEATURE_AUTO_NOTIFICATION_CONFIG_ENABLED=true` and the staging database/channel targets are disposable for auto-notification setup/list/delete.
- If `MHCAT_FEATURE_AUTO_NOTIFICATION_DELIVERY_ENABLED=true`, confirm Gateway and scheduler leases are enabled, the lease owner is unique, the Node `handler/cron.js` owner is stopped, and every active staging `cron_sets` target/payload is safe to send.
- If `MHCAT_FEATURE_DAILY_RESET_SCHEDULER_ENABLED=true`, confirm the daily-reset write/Gateway/lease gates are enabled, lease TTL exceeds reset plus lease timeout, Node `handler/cron.js` is stopped, owners are unique, and every staging `coins`/`gift_changes`/`work_sets`/`work_users` row is disposable.
- If `MHCAT_FEATURE_WORK_PAYOUT_SCHEDULER_ENABLED=true`, confirm the payout write/Gateway/lease gates are enabled, CLI and worker lease names match, lease TTL exceeds payout plus lease timeout, Node `handler/gift.js` is stopped, owners are unique, duplicate/marker audits are clean, and every due staging `coins`/`work_users` row is disposable.
- If `MHCAT_COMMAND_SYNC_INCLUDE_LOGGING_CONFIG=true`, confirm `MHCAT_FEATURE_LOGGING_CONFIG_ENABLED=true`, the selected channel/database are staging-only, and testing includes public discoverability plus runtime Manage Messages denial.
- If `MHCAT_FEATURE_LOGGING_MESSAGE_EVENTS_ENABLED=true`, confirm Gateway, Guild Messages, and Message Content are enabled; `loggings.channel_id` resolves to a staging-only target under Mongoose scalar coercion; cached source messages are disposable; and Node `LoggingSystem.js` ownership is stopped.
- If `MHCAT_FEATURE_LOGGING_CHANNEL_EVENTS_ENABLED=true`, confirm Gateway is enabled; `loggings.channel_id` resolves to a staging-only target; topic/overwrite/audit fixtures are disposable; and Node `LoggingSystem.js` ownership is stopped.
- If `MHCAT_FEATURE_LOGGING_VOICE_EVENTS_ENABLED=true`, confirm Gateway and Voice State intent are enabled; `loggings.channel_id` resolves to a staging-only target; human/bot test members and voice channels are safe; and Node `LoggingSystem.js` ownership is stopped.
- If `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_LIST=true`, confirm `MHCAT_FEATURE_GACHA_PRIZE_LIST_ENABLED=true` and the staging database has safe gacha fixtures.
- If `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_DRAW=true`, confirm `MHCAT_FEATURE_GACHA_DRAW_ENABLED=true`, the staging database has isolated `coins`/`gifts`/`gift_changes` fixtures, and DMs/notification-channel sends are acceptable for the test account.
- If `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_CREATE=true`, confirm `MHCAT_FEATURE_GACHA_PRIZE_CREATE_ENABLED=true` and the staging database has only disposable `gifts` fixtures for the target prize name.
- If `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_EDIT=true`, confirm `MHCAT_FEATURE_GACHA_PRIZE_EDIT_ENABLED=true` and the staging database has only disposable `gifts` fixtures for the target prize name.
- If `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_DELETE=true`, confirm `MHCAT_FEATURE_GACHA_PRIZE_DELETE_ENABLED=true` and the staging database has disposable `gifts` fixtures for the target prize name.
- If `MHCAT_COMMAND_SYNC_INCLUDE_LOTTERY_DISABLED_COMMAND=true`, confirm `MHCAT_FEATURE_LOTTERY_DISABLED_COMMAND_ENABLED=true` and that the expected result is only the unavailable embed.
- If `MHCAT_COMMAND_SYNC_INCLUDE_STATS_QUERY=true`, confirm `MHCAT_FEATURE_STATS_QUERY_ENABLED=true` and that the expected result is only the static stats help embed.
- If `MHCAT_COMMAND_SYNC_INCLUDE_STATS_CREATE=true`, confirm `MHCAT_FEATURE_STATS_CREATE_ENABLED=true`, the staging guild can safely receive new stat channels, and the staging database has disposable `numbers` rows.
- If `MHCAT_COMMAND_SYNC_INCLUDE_STATS_ROLE_COUNT=true`, confirm `MHCAT_FEATURE_STATS_ROLE_COUNT_ENABLED=true`, the staging guild can safely receive a role-count stat channel, and the staging database has disposable `role_numbers` rows plus an existing stats base `numbers` row.
- If `MHCAT_COMMAND_SYNC_INCLUDE_STATS_DELETE=true`, confirm `MHCAT_FEATURE_STATS_DELETE_ENABLED=true` and the staging database has disposable `numbers` rows.
- If `MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_CONFIG=true`, confirm `MHCAT_FEATURE_ANNOUNCEMENT_CONFIG_ENABLED=true`, matching Node command ownership is stopped, duplicate/type/shared-writer audits are reviewed, and the staging database can safely patch `guilds` and `ann_all_sets`.
- If `MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_SEND=true`, confirm `MHCAT_FEATURE_ANNOUNCEMENT_SEND_ENABLED=true`, matching Node command/modal ownership is stopped, and the six-second confirmation flow can send to a disposable channel.
- If `MHCAT_FEATURE_ANNOUNCEMENT_RELAY_ENABLED=true`, confirm Gateway/Guild Messages/Message Content flags, stop Node `events/ann_message.js` ownership for the same bot/guild, review duplicate/type findings, and use a bound channel safe for send-before-delete tests. Follow [76-announcement.md](76-announcement.md).
- If either anti-scam command-sync flag is enabled, confirm its matching runtime gate; report also requires a safe webhook. Stop matching Node commands and review duplicate/type/raw-URL/external-writer findings.
- If `MHCAT_FEATURE_ANTI_SCAM_MESSAGE_DELETE_ENABLED=true`, confirm Gateway/Guild Messages/Message Content flags, stop Node `events/safe_server.js`, and use a disposable channel/catalog safe for deletion and bot-message tests. Follow [77-anti-scam.md](77-anti-scam.md).
- If `MHCAT_COMMAND_SYNC_INCLUDE_BIRTHDAY_CONFIG=true`, confirm `MHCAT_FEATURE_BIRTHDAY_CONFIG_ENABLED=true`, stop Node `/生日系統`, review duplicate/type/value/external-writer findings, keep delivery inactive, and use disposable rows. Follow [78-birthday.md](78-birthday.md).
- If `MHCAT_COMMAND_SYNC_INCLUDE_TEXT_XP_CONFIG=true`, confirm `MHCAT_FEATURE_TEXT_XP_CONFIG_ENABLED=true` and the staging database can safely write `text_xp_channels`.
- If `MHCAT_COMMAND_SYNC_INCLUDE_VOICE_XP_CONFIG=true`, confirm `MHCAT_FEATURE_VOICE_XP_CONFIG_ENABLED=true` and the staging database can safely write `voice_xp_channels`.
- If `MHCAT_COMMAND_SYNC_INCLUDE_XP_ROLE_CONFIG=true`, confirm `MHCAT_FEATURE_XP_ROLE_CONFIG_ENABLED=true`, the staging database can safely write `chat_roles`/`voice_roles`, and the test roles are below the bot's highest role.
- If `MHCAT_COMMAND_SYNC_INCLUDE_XP_PROFILE_DISABLED_COMMANDS=true`, confirm `MHCAT_FEATURE_XP_PROFILE_DISABLED_COMMANDS_ENABLED=true` and that the expected result is only the replacement embed.
- If `MHCAT_COMMAND_SYNC_INCLUDE_XP_ADMIN=true`, confirm `MHCAT_FEATURE_XP_ADMIN_ENABLED=true` and the staging database has only disposable `text_xps`/`voice_xps` rows for the test member.
- If `MHCAT_COMMAND_SYNC_INCLUDE_XP_RESET=true`, confirm `MHCAT_FEATURE_XP_RESET_ENABLED=true`, `MHCAT_DISCORD_ENABLE_GATEWAY=true`, `MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true`, `MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true`, the tester is the staging guild owner, and the staging database has only disposable `text_xps`/`voice_xps` rows for the reset target.
- If `MHCAT_COMMAND_SYNC_INCLUDE_XP_RANK=true`, confirm `MHCAT_FEATURE_XP_RANK_ENABLED=true` and the staging database has disposable `text_xps`/`voice_xps` rows for leaderboard rendering.
- If `MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_CONFIG=true`, confirm `MHCAT_FEATURE_VOICE_ROOM_CONFIG_ENABLED=true` and the staging database can safely write/delete `voice_channels`.
- If `MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_LOCK=true`, confirm `MHCAT_FEATURE_VOICE_ROOM_LOCK_ENABLED=true`, gateway and Voice State intent are enabled, and the staging database can safely replace disposable `lock_channels` rows.
- If `MHCAT_COMMAND_SYNC_INCLUDE_JOIN_ROLE_CONFIG=true`, confirm `MHCAT_FEATURE_JOIN_ROLE_CONFIG_ENABLED=true`, the staging database can safely write `join_roles`, and the test role is below the bot's highest role.
- If `MHCAT_FEATURE_JOIN_ROLE_ASSIGNMENT_ENABLED=true`, confirm `MHCAT_DISCORD_ENABLE_GATEWAY=true`, `MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true`, the staging database has safe `join_roles` rows, and the target roles are below the bot's highest role.
- If `MHCAT_COMMAND_SYNC_INCLUDE_WELCOME_MESSAGE_CONFIG=true`, confirm `MHCAT_FEATURE_WELCOME_MESSAGE_CONFIG_ENABLED=true` and the staging database can safely write `leave_messages`.
- If `MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_CONFIG=true`, confirm `MHCAT_FEATURE_VERIFICATION_CONFIG_ENABLED=true`, the staging database can safely write `verifications`, and the target role is below the bot's highest role.
- If `MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_FLOW=true`, confirm `MHCAT_FEATURE_VERIFICATION_FLOW_ENABLED=true`, a staging `verifications` row exists or `/驗證設置` will be tested first, the target role is below the bot's highest role, and the test accepts process-local challenge state.
- If `MHCAT_COMMAND_SYNC_INCLUDE_ACCOUNT_AGE_CONFIG=true`, confirm `MHCAT_FEATURE_ACCOUNT_AGE_CONFIG_ENABLED=true`, matching Node command ownership is stopped, duplicate/type/value audits are reviewed, and the staging database can safely write `create_hours`.
- If `MHCAT_FEATURE_ACCOUNT_AGE_POLICY_ENABLED=true`, confirm Gateway/Guild Members, stop the leading Node member-add branch, review `create_hours`, verify handler order, and use a disposable too-new member. Follow [79-account-age.md](79-account-age.md).
- If `MHCAT_COMMAND_SYNC_INCLUDE_ROLE_SELECTION=true`, confirm `MHCAT_FEATURE_ROLE_SELECTION_ENABLED=true`, `MHCAT_DISCORD_ENABLE_GATEWAY=true`, `MHCAT_DISCORD_GUILD_MESSAGE_REACTIONS_INTENT=true`, the staging database can safely write `message_reactions` and `btns`, and the test role is below the bot's highest role.
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

For poll staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_POLLS=true`;
- `MHCAT_FEATURE_POLLS_ENABLED=true`;
- plan includes managed `投票創建`;
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

For economy coin-reset staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_RESET=true`;
- `MHCAT_FEATURE_ECONOMY_COIN_RESET_ENABLED=true`;
- `MHCAT_DISCORD_ENABLE_GATEWAY=true`;
- `MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true`;
- `MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true`;
- plan includes managed `代幣重製`;
- plan still performs no create/update/delete during dry-run.

For economy RPS staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_RPS=true`;
- `MHCAT_FEATURE_ECONOMY_RPS_ENABLED=true`;
- plan includes managed `剪刀石頭布`;
- plan still performs no create/update/delete during dry-run.

For economy game staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_GAME=true`;
- `MHCAT_FEATURE_ECONOMY_GAME_ENABLED=true`;
- preflight reports `economy-game-runtime-safety status=warn` for the manual topology/ownership/restart review;
- plan includes managed `代幣遊戲`;
- plan still performs no create/update/delete during dry-run;
- live smoke uses transaction-capable Mongo, verifies the 500-millisecond knowledge start and five-second reveal/draw states, and verifies both knowledge and blackjack timeout forfeits remove components and settle once.

For economy shop staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SHOP=true`;
- `MHCAT_FEATURE_ECONOMY_SHOP_ENABLED=true`;
- plan includes managed `代幣商店`;
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
- plan includes managed `automatic-notification`, `自動通知列表`, and `自動通知刪除`;
- plan still performs no create/update/delete during dry-run.

Auto-notification delivery has no command-sync include flag and should not change the dry-run plan. With delivery enabled, staging preflight must warn that the runtime sends persisted `cron_sets` payloads, writes `mhcat_scheduler_locks`, and requires the Node `handler/cron.js` owner to be disabled.

Daily-reset scheduling also has no command-sync include flag. With it enabled, staging preflight must warn about the `00:00 Asia/Taipei` run, `coins.today`/`work_users.energi` writes, `daily-reset` lease, and exclusive Node ownership.

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

For gacha draw staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_DRAW=true`;
- `MHCAT_FEATURE_GACHA_DRAW_ENABLED=true`;
- plan includes managed `扭蛋`;
- plan still performs no create/update/delete during dry-run.

For gacha prize-delete staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_DELETE=true`;
- `MHCAT_FEATURE_GACHA_PRIZE_DELETE_ENABLED=true`;
- plan includes managed `扭蛋獎池刪除`;
- plan still performs no create/update/delete during dry-run.

For gacha prize-edit staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_EDIT=true`;
- `MHCAT_FEATURE_GACHA_PRIZE_EDIT_ENABLED=true`;
- plan includes managed `扭蛋獎品編輯`;
- plan still performs no create/update/delete during dry-run.

For gacha prize-create staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_CREATE=true`;
- `MHCAT_FEATURE_GACHA_PRIZE_CREATE_ENABLED=true`;
- plan includes managed `扭蛋獎池增加`;
- plan still performs no create/update/delete during dry-run.

For lottery disabled-command staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_LOTTERY_DISABLED_COMMAND=true`;
- `MHCAT_FEATURE_LOTTERY_DISABLED_COMMAND_ENABLED=true`;
- plan includes managed `抽獎設置`;
- `/抽獎設置` returns the legacy unavailable embed and performs no lottery creation/write.

For existing-lottery component staging smoke, expected additionally:

- seed a disposable future-dated `lotters` row whose numeric ID ends in `lotter`, and post disposable enter/search buttons using that ID;
- enter with an eligible member and verify one `{id,time}` participant append;
- verify duplicate, ended, full, required-role, and forbidden-role responses do not mutate the row;
- search and verify the participant count, self-entry status, `discord.txt`, and owner/guild-owner controls;
- verify another user cannot invoke `<id>restart` or `<id>stop` by custom ID;
- reroll as the owner, verify one winner message in `message_channel`, explicit winner mentions only, and `end:true`;
- verify `/抽獎設置` remains unavailable or unrouted according to its independent command flag.

For stats query staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_STATS_QUERY=true`;
- `MHCAT_FEATURE_STATS_QUERY_ENABLED=true`;
- plan includes managed `統計系統查詢`;
- `/統計系統查詢` returns the legacy static stats help embed and performs no Mongo or channel writes.

For stats create staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_STATS_CREATE=true`;
- `MHCAT_FEATURE_STATS_CREATE_ENABLED=true`;
- plan includes managed `統計系統創建`;
- run `/統計系統創建 統計頻道類型:文字頻道`, verify the stats category plus member/user/bot channels are created, and verify one `numbers` row is written for the staging guild;
- rerun with `統計選項:頻道數量` and verify one channel-count stat channel is added under the saved category.

For stats role-count staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_STATS_ROLE_COUNT=true`;
- `MHCAT_FEATURE_STATS_ROLE_COUNT_ENABLED=true`;
- plan includes managed `統計身分組人數`;
- run `/統計身分組人數` after the base stats row exists, verify the role-count channel is created, and verify one `role_numbers` row is replaced for the staging guild/role.

For stats delete staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_STATS_DELETE=true`;
- `MHCAT_FEATURE_STATS_DELETE_ENABLED=true`;
- plan includes managed `統計系統刪除`;
- seed a disposable `numbers` row for the staging guild, run `/統計系統刪除`, and verify the row is deleted while Discord channels are untouched.

For stats rename worker staging smoke, expected additionally:

- `MHCAT_FEATURE_STATS_RENAME_WORKER_ENABLED=true`;
- `MHCAT_DISCORD_ENABLE_GATEWAY=true`;
- `MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true`;
- no command-sync flag is needed;
- seed disposable `numbers`/`role_numbers` rows and matching stat channels, leave the legacy worker stopped, wait one 20-minute interval, and verify renamed channels and updated old-number fields.

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

For text-XP accrual staging smoke, expected additionally:

- `MHCAT_FEATURE_TEXT_XP_ACCRUAL_ENABLED=true`;
- `MHCAT_DISCORD_ENABLE_GATEWAY=true`;
- `MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true`;
- `MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true`;
- no command-sync plan changes are expected; verify `text_xps.xp`/`leavel` changes and any `coins.coin` XP rewards only against disposable rows.

For voice-XP session staging smoke, expected additionally:

- `MHCAT_FEATURE_VOICE_XP_SESSIONS_ENABLED=true`;
- `MHCAT_DISCORD_ENABLE_GATEWAY=true`;
- `MHCAT_DISCORD_VOICE_STATE_INTENT=true`;
- no command-sync plan changes are expected; verify `voice_xps.leavejoin` changes only against disposable rows.

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

For XP admin staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_XP_ADMIN=true`;
- `MHCAT_FEATURE_XP_ADMIN_ENABLED=true`;
- plan includes managed `經驗值改變`;
- `/經驗值改變` adjusts only disposable staging `text_xps`/`voice_xps` rows and performs no XP accrual or rank side effects.

For XP reset staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_XP_RESET=true`;
- `MHCAT_FEATURE_XP_RESET_ENABLED=true`;
- `MHCAT_DISCORD_ENABLE_GATEWAY=true`;
- `MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true`;
- `MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true`;
- plan includes managed `經驗值重製`;
- plan still performs no create/update/delete during dry-run.

For XP rank staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_XP_RANK=true`;
- `MHCAT_FEATURE_XP_RANK_ENABLED=true`;
- plan includes managed `聊天排行榜` and `語音排行榜`;
- `/聊天排行榜` and `/語音排行榜` render `user-info.png` and perform no Mongo writes.

For voice-room config staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_CONFIG=true`;
- `MHCAT_FEATURE_VOICE_ROOM_CONFIG_ENABLED=true`;
- plan includes managed `語音包廂設置` and `語音包廂刪除`;

For voice-room lock staging smoke, expected additionally:

- `MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_LOCK=true`;
- `MHCAT_FEATURE_VOICE_ROOM_LOCK_ENABLED=true`;
- `MHCAT_DISCORD_ENABLE_GATEWAY=true`;
- `MHCAT_DISCORD_VOICE_STATE_INTENT=true`;
- plan includes managed `上鎖頻道`;
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

If poll inclusion is enabled, expected:

- create/update managed `投票創建` only in addition to the utility commands;
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

If economy coin-reset inclusion is enabled, expected:

- create/update managed `代幣重製` only in addition to the utility commands;
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

- create/update managed `automatic-notification`, `自動通知列表`, and `自動通知刪除` only in addition to the utility commands;
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

If XP admin inclusion is enabled, expected:

- create/update managed `經驗值改變` only in addition to the utility commands and other explicitly included features;
- no command deletion;
- no bulk overwrite;
- no global command mutation.

If XP reset inclusion is enabled, expected:

- create/update managed `經驗值重製` only in addition to the utility commands and other explicitly included features;
- no command deletion;
- no bulk overwrite;
- no global command mutation.

If XP rank inclusion is enabled, expected:

- create/update managed `聊天排行榜` and `語音排行榜` only in addition to the utility commands and other explicitly included features;
- no command deletion;
- no bulk overwrite;
- no global command mutation.

If voice-room config inclusion is enabled, expected:

- create/update managed `語音包廂設置` and `語音包廂刪除` only in addition to the utility commands;

If voice-room lock inclusion is enabled, expected:

- create/update managed `上鎖頻道` only in addition to the utility commands and other explicitly included features;
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
- no Mongo feature write unless `MHCAT_FEATURE_USAGE_TRACKING_ENABLED=true`, in which case gateway startup alone still writes nothing;
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

If global usage tracking was enabled:

- record the staging `all_use_counts` values for `ping` and `info` before invoking either command;
- run `/ping` once and verify exactly one `ping` row increases by one with a numeric `count`;
- run `/info bot` once and verify exactly one `info` row increases by one;
- click the `botinfoupdate` refresh component and verify neither usage row changes;
- confirm the preflight ownership warning was reviewed and no second Node owner produced a double increment.

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

If economy coin-reset flags were enabled and command sync apply was reviewed:

- seed disposable staging `coins` rows for the guild, including at least one row for division and at least one row for deletion;
- run `/代幣重製` as a non-owner and verify the legacy red permission embed appears with no balance mutation;
- run `/代幣重製` as the staging guild owner, type a wrong confirmation once in the same channel, and verify the reset is cancelled without mutating `coins`;
- rerun `/代幣重製` without `除以多少`, type `^確認^` in the same channel as the same owner within 60 seconds, and verify all disposable guild `coins` rows are deleted;
- reseed disposable `coins` rows, run `/代幣重製 除以多少:<divisor>` as the staging guild owner, type `^確認^`, and verify every guild balance is updated with legacy rounded division;
- verify no economy sign-in, gacha, work payout, XP reward, index, or usage-counter side effect happened.

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

- run `/automatic-notification channel:<test channel>` and submit the legacy modal with a safe direct cron such as `*/30 * * * *`;
- verify the completion message includes the generated id and a setup preview message appears in the interaction channel;
- verify the staging `cron_sets` row contains `guild`, `channel`, `id`, `cron`, and rollback-compatible `message`;
- repeat setup with `cancel` or `取消`, verify the legacy-style weekday/hour/minute selects and five-minute deadline, then complete selections and verify the expected `<minutes> <hours> * * <weekdays>` cron value;
- verify a different user cannot advance that wizard and an expired wizard returns the safe ephemeral retry error;
- create or confirm a safe staging `cron_sets` active row with `guild`, `id`, `cron`, and `channel`;
- optionally create a disposable staging draft row with null or missing `cron`;
- run `/自動通知列表` and verify the legacy list embed includes active rows and does not render pending drafts;
- verify the pending draft row was cleaned from staging data;
- run `/自動通知刪除 id:<active id>` and verify the legacy green delete embed appears;
- run `/自動通知刪除 id:<missing id>` and verify the legacy red missing-id tutorial embed appears;
- verify no recurring channel send occurred unless the independent delivery gate was explicitly enabled, and verify no index creation or Message Content intent was involved.

If auto-notification delivery was explicitly enabled:

- stop the legacy Node process that loads `handler/cron.js` before starting the Go bot;
- use an isolated database and audit all active `cron_sets` rows so only the intended disposable row can send;
- seed one active row with a unique `{guild,id}`, the disposable channel, and a five-field cron for the next `Asia/Taipei` minute; use a payload containing visible `content` plus `embeds:[{data:{title,description,color:10944422}}]` to exercise legacy numeric color decoding;
- run `go run ./cmd/mhcat-staging-preflight --format text` and require the auto-notification delivery warning, then start exactly one Go lease owner;
- inspect `mhcat_scheduler_locks` or run `go run ./cmd/mhcat-scheduler-lease --name auto-notification-delivery --action status` and verify that owner holds an unexpired lease;
- at the scheduled UTC+8 minute, verify exactly one channel message has the expected content, embed title/description, and color `#A6FFA6`;
- start no second sender; if testing contention with another Go process, give it a different owner and verify only the lease holder schedules/sends;
- change the cron and allow up to 30 seconds for reconciliation, then verify only the changed schedule is active;
- delete the `{guild,id}` row before its next occurrence and verify no future send occurs, including if an already-registered callback reaches its tick;
- stop the Go bot gracefully and verify `auto-notification-delivery` is no longer held before any Node rollback;
- keep the delivery gate off after smoke and remove the disposable row; no index should have been created and Message Content intent is not required.

If daily-reset scheduling was explicitly enabled:

- stop every Node process that loads `handler/cron.js`;
- seed only disposable `coins`, `gift_changes`, `work_sets`, and `work_users` rows, including one daily-mode guild and one nonzero rolling-cooldown guild;
- record `coins.today` and `work_users.energi`, then run `go run ./cmd/mhcat-economy-reset --dry-run` and verify expected counts without a lease write;
- run one-shot apply with the write/lease gates and verify `lease_acquired=true`, expected reset/refill/clamp values, and a released `daily-reset` lease;
- hold `daily-reset` with a different owner, rerun apply, and verify exit code `2` with no data changes;
- restore fresh before-values, start two Go replicas with distinct owners, and keep them running across `00:00 Asia/Taipei`;
- verify one replica logs reset completion, the contender logs a held-lease skip, daily-mode `coins.today` becomes `0`, rolling-cooldown guild rows remain unchanged, and each work row receives at most one increment before clamp;
- stop both Go replicas gracefully, verify `daily-reset` is not held, disable the scheduler flag, and remove fixtures;
- if the run partially fails, inspect per-guild after-values before any retry because another apply can increment already-processed work energy again.

If announcement relay was explicitly enabled:

- execute the bound-relay cases in the canonical [announcement staging smoke](76-announcement.md#staging-smoke), including exact/legacy random colors, unsupported colors, send-before-delete failures, visible non-pinging tags, empty/attachment-only retention, ignored bot/DM/unconfigured messages, and exclusive Node/Go ownership;
- preserve `guilds` and `ann_all_sets`, create no index or automatic repair, then disable the relay and intents according to the canonical rollback sequence.

If anti-scam runtime was explicitly enabled:

- execute the config/report/deletion cases in the canonical [anti-scam staging smoke](77-anti-scam.md#staging-smoke), including exact UI, pinned URL cases, webhook failures, scalar reads, raw catalog values, bot scanning, and delete/warning failures;
- preserve `good_webs` and `not_a_good_webs`, create no index/normalization/repair, then disable commands, event gates, and intents according to the canonical rollback sequence.

If join-role assignment was explicitly enabled:

- run every UI/color, scalar/duplicate, usage, audience, cache/hierarchy, owner-DM, role-failure, continuation, account-age-ordering, and rollback case in [81-join-role.md](81-join-role.md);
- verify welcome, verification, account-age, and leave behavior occurs only under its separate gate.

If welcome-message delivery was explicitly enabled:

- run every generic/special welcome case in [82-welcome-leave.md](82-welcome-leave.md), not only the happy path;
- create or confirm a staging `join_messages` row with safe `channel`, `message_content`, `color`, and optional `img`;
- join the staging guild with a disposable test member;
- verify one legacy-style welcome embed appears in the configured channel;
- verify the author text is `🪂 歡迎加入 <guild name>`;
- verify `(MEMBERNAME)`, `{MEMBERNAME}`, `{membername}`, `(TAG)`, `{TAG}`, and `{tag}` replacements match the saved template;
- verify only the joining member mention is allowed and everyone/role mentions do not ping;
- verify no command registration, Mongo write, verification, leave send, or account-age kick happened unless those separate feature flags are explicitly enabled.

If leave-message delivery was explicitly enabled:

- run every leave setup/delivery/data/rollback case in [82-welcome-leave.md](82-welcome-leave.md), not only the happy path;
- configure `/退出訊息設置` in a staging-only channel;
- remove a test member from the staging guild;
- verify one legacy-style leave embed appears in the configured channel;
- verify `(MEMBERNAME)`, `{MEMBERNAME}`, `(ID)`, and `{ID}` replacements match the saved template;
- verify no command registration, Mongo write, welcome send, verification, or account-age kick happened unless those separate feature flags are explicitly enabled.

If verification flow was explicitly enabled:

- run every metadata/UI, scalar/duplicate, usage, captcha/state, wrong/correct answer, role/nickname ordering, legacy-ID, failure, and rollback case in [80-verification.md](80-verification.md);
- verify no account-age behavior, startup Mongo mutation, or handler-local usage write occurs.

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

If role-selection flags were enabled and command sync apply was reviewed:

The canonical checklist, including duplicate/type audit, custom-emoji deletion, stale buttons, REST failures, and usage verification, is in the [role-selection parity contract](73-role-selection.md#staging-smoke). The minimum shared smoke remains:

- run `/選取身分組-表情符號` against a disposable staging message with a staging role below the bot's highest role;
- verify the bot adds the configured reaction to the target message;
- verify `message_reactions` stores `guild`, `message`, `react`, and `role`;
- react and unreact as a test member, then verify the configured role is added and removed;
- react and unreact as a bot in the same Go process, then verify neither event changes the configured role;
- record that a reaction-remove event first observed after a Go process restart has no member payload and bot detection is best-effort when the member is absent from Discord state;
- run `/選取身分組刪除-表情符號` and verify the matching `message_reactions` row is deleted;
- run `/選取身分組-按鈕`, submit the legacy `領取身分系統!` modal, verify the public `選取身分組` panel appears, click add/delete, and verify `btns` stores the add/delete IDs with the configured role;
- verify unassignable roles and missing reaction config return legacy red errors.

If ticket flags were enabled and command sync apply was reviewed:

The canonical duplicate/type audit, stale-modal race, failure compensation, overwrite, owner-close, usage, and rollback checks are in the [ticket parity contract](74-ticket.md#staging-smoke). The minimum shared smoke remains:

- `/私人頻道設置`
- submit the setup modal with a safe test title/content/color;
- click the `tic` ticket open button;
- verify a private text channel is created under the selected category;
- verify the welcome embed and `del` button appear;
- click `del` from the ticket channel;
- verify the channel is deleted or the legacy denial embed appears when expected.

If poll flags were enabled and command sync apply was reviewed:

Run the complete canonical [poll staging checklist](75-poll.md#staging-smoke). Do not reduce it to a create-and-vote happy path: it covers exact validation/UI, old and versioned IDs, `poll_menu`/colon choices, atomic concurrency, exports, source refreshes, insertion rollback, ownership, and Node rollback constraints.

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

- verify a member without Manage Messages can discover `/set-log-channel` but receives the exact public red runtime denial;
- as a manager, run `/set-log-channel channel:<staging log channel>` and verify the exact yellow title/text/footer, four option labels/descriptions/values, placeholder, and one-to-four bounds;
- have another member press the menu and verify the owner error, then verify the invoking user cannot save at or after the original ten-minute deadline;
- choose multiple values twice and verify selected text updates while every menu option remains visually non-default;
- verify typed `loggings.channel_id`, `message_update`, `message_delete`, `channel_update`, and `member_voice_update` values replace all duplicate guild rows and no index appears;
- on a disposable copied row, seed numeric `channel_id` plus string/numeric toggle scalars and verify Mongoose-compatible reads, then restore typed values with the command;
- verify an orphaned static `loggin_create` component asks for a rerun and cannot save an unknown channel;
- verify config/select activity emits no event and adds no usage write beyond the one initial slash attempt when usage tracking is enabled.

If logging message-event flags were enabled:

- ensure the staging `loggings` row has `channel_id` pointing to the staging log channel and enables `message_update` and/or `message_delete`;
- send then edit a cached human message and verify `訊息編輯` uses the pre-edit author, exact color/field/code-block spacing, new attachment values in raw order, PNG/default author avatar, WebP/GIF bot footer, and no mention ping;
- verify unchanged content, a bot-authored message, and an update without cached old content emit nothing;
- delete a cached message with an action-72 audit entry whose target and channel both match; verify the matching executor is shown;
- repeat with either audit target or channel mismatched and verify the displayed deleter falls back to the message author;
- verify no channel or voice log event is emitted.

If logging channel-event flags were enabled:

- ensure the staging `loggings` row has `channel_id` pointing to the staging log channel and enables `channel_update`;
- edit a disposable topic to/from API null and verify `頻道主題更新` renders literal `null`, uses the first action-11 audit executor without target filtering, and sends no ping;
- change topic and permissions in one event where possible and verify the topic branch takes precedence;
- edit disposable role and user overwrites and verify `頻道權限更新` uses overwrite-type mention syntax plus the exact legacy permission/default/allow/deny ordering;
- verify no message or voice log event is emitted.

If logging voice-event flags were enabled:

- ensure the staging `loggings` row has `channel_id` pointing to the staging log channel and enables `member_voice_update`;
- join as a human and bot and verify `使用者加入語音頻道` uses the new-state member/channel, cached name, exact avatar/footer, and no mention ping;
- leave and verify `使用者退出語音頻道` uses the old-state member/channel;
- move directly between two staging voice channels and change mute/deafen state, verifying neither action emits a voice logging embed, matching legacy behavior;
- verify no message or channel log event and no usage-counter write is emitted.

If XP reward-role config flags were enabled and command sync apply was reviewed:

- run `/聊天經驗身分組設定 增加` with a staging role below the bot's highest role, a safe level, and both `到達等級後自動刪除身分組` choices across test rows;
- verify the legacy green add/modify embed appears and the staging `chat_roles` row stores `guild`, string `leavel`, `role`, and `delete_when_not`;
- run `/聊天經驗身分組設定 設定查詢` and verify the legacy query embed plus `上一頁`/`下一頁` buttons when there are enough seeded rows;
- run `/聊天經驗身分組設定 刪除` for the staged level/role and verify the legacy green delete embed and row removal;
- repeat the same add/query/delete path for `/語音經驗身分組設定` and `voice_roles`;
- verify unassignable roles, missing deletes, and over-limit seeded data return legacy red errors;
- verify no XP accrual, rank rendering, automatic role assignment/removal, coin reward, gateway intent, index, or usage-counter write happened.

If XP admin flags were enabled and command sync apply was reviewed:

- seed or choose a disposable staging member profile in `text_xps` and `voice_xps`;
- run `/經驗值改變 聊天經驗改變` with positive and negative `經驗值` values and verify the legacy success embed plus string `xp`/`leavel` profile updates;
- run `/經驗值改變 語音經驗改變` for a member without a voice profile and verify the new `voice_xps` row stores `guild`, `member`, string `xp`, string `leavel`, and `leavejoin: leave`;
- verify a member without Kick Members receives the legacy red permission embed;
- verify no XP accrual, rank rendering, automatic role assignment/removal, coin reward, gateway intent, index, or usage-counter write happened.

If XP reset flags were enabled and command sync apply was reviewed:

- seed disposable staging `text_xps` and `voice_xps` rows for one test member plus at least one additional guild row if testing full-server reset;
- run `/經驗值重製 重製個人聊天經驗 使用者:<staging member>` as the staging guild owner and verify the selected member's `text_xps` row is deleted;
- run `/經驗值重製 重製個人語音經驗 使用者:<staging member>` and verify the selected member's `voice_xps` row is deleted;
- run the same individual reset against a missing profile and verify the legacy red `這位使用者還沒有任何的經驗值喔!` embed;
- run `/經驗值重製 聊天經驗重製`, verify the legacy warning prompt appears, type a wrong confirmation once, and verify the reset is cancelled without deleting rows;
- rerun `/經驗值重製 聊天經驗重製`, type `^確認^` in the same channel as the same owner within 60 seconds, and verify all staging guild `text_xps` rows are deleted;
- repeat the full-server confirmation path for `/經驗值重製 語音經驗重製` against disposable `voice_xps` rows;
- verify non-owner attempts return the legacy red permission embed and no rows are deleted;
- verify no XP accrual, rank rendering, automatic role assignment/removal, coin reward, index, or usage-counter write happened.

If XP rank flags were enabled and command sync apply was reviewed:

- seed disposable staging `text_xps` and `voice_xps` rows for several members, including the tester;
- run `/聊天排行榜` and verify the loading embed is replaced by a `user-info.png` leaderboard with legacy pagination buttons;
- run `/語音排行榜` and verify the same PNG/button behavior, including the legacy page label and viewer-target button;
- click previous/next and viewer-target rank buttons where enabled and verify the message updates rather than sending a new public message;
- verify no XP accrual, profile-card render, automatic role assignment/removal, coin reward, gateway intent, index, or Mongo write happened.

If voice-room config flags were enabled and command sync apply was reviewed:

- verify both config commands are discoverable to a non-manager but return the legacy runtime permission denial;
- run `/語音包廂設置` as a manager with a staging voice or stage trigger, raw template ` {name}-{name} `, `是否予許房主上鎖`, and explicit `設定人數上限:0`;
- verify the legacy green setup embed appears;
- verify the staging `voice_channels` row contains `guild`, `ticket_channel`, numeric `limit:0`, the unchanged raw `name`, the setup-time `parent`, and Boolean `lock`;
- move the trigger to another category, join it as a disposable member, and verify the room uses that current category/overwrites, replaces only the first `{name}` token, moves the member, persists `voice_channel_ids`, and seeds a nullable `lock_channels` row plus owner DM;
- leave the created room and verify the empty room is deleted, the matching `voice_channel_ids` row is removed, and the matching lock seed row is removed;
- run `/語音包廂刪除` for the same trigger channel and verify the legacy delete embed appears and the matching staging rows are removed;
- optionally configure a stage trigger under a disposable staging category, verify deleting by the stage trigger follows the legacy category branch and returns the missing-category response without removing its trigger row, then delete by category and verify only rows with that `parent` are removed;
- with usage tracking disabled verify no counter write; when separately testing usage tracking, verify exactly one `all_use_counts` increment per config slash attempt and none for voice-state events.

If voice-room lock flags were enabled and command sync apply was reviewed:

- seed a disposable `lock_channels` row for the staging guild/current voice channel with the invoking user as `owner`;
- join that staging voice channel and run `/上鎖頻道` with a password;
- verify the legacy ephemeral success embed appears and the row is replaced with raw `lock_anser`, `owner`, `text_channel`, and BSON-array `ok_people:[]`;
- from another test user, join the locked voice channel and verify the prompt message appears in `text_channel`, the user is disconnected, and the DM instruction is sent;
- verify another user cannot open the prompt and the intended user cannot open it at or after exactly 60 seconds;
- before expiry, click the generated prompt button, submit the exact password, and verify the user ID is added once to `ok_people`;
- run `/上鎖頻道` without `密碼` and verify the success embed shows `null` and Mongo stores BSON null; separately seed an empty BSON string and verify it remains distinct on reads;
- verify not-in-voice, missing lock row, and non-owner cases return the legacy red ephemeral errors;
- with usage tracking disabled verify no counter write; when separately testing usage tracking, verify exactly one `all_use_counts` increment per `/上鎖頻道` slash attempt and none for prompt buttons or modal submissions.

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
- no Mongo feature write happened unless an explicitly enabled feature below owns one.
  - Exception: ticket smoke writes the legacy-compatible `tickets` config only after successful modal submit.
  - Exception: poll smoke inserts and atomically updates disposable staging `polls` rows; a forced failed insert must delete the exact public message it sent.
  - Exception: economy coin-admin smoke writes disposable staging `coins` rows only.
  - Exception: economy coin-reset smoke deletes or divides disposable staging `coins` rows only after the owner `^確認^` confirmation.
  - Exception: role-selection smoke writes staging `message_reactions` and `btns` rows only after setup commands.
  - Exception: logging-config smoke writes the legacy-compatible `loggings` config only after the setup select is submitted.
  - Exception: logging message-event smoke sends edit/delete embeds to the configured staging log channel only.
  - Exception: logging channel-event smoke sends topic/permission embeds to the configured staging log channel only.
  - Exception: logging voice-event smoke sends join/leave embeds to the configured staging log channel only.
  - Exception: delete-data smoke deletes selected disposable staging config rows only.
  - Exception: auto-notification config smoke writes/completes disposable setup `cron_sets` rows, sends a setup preview, and deletes selected rows and abandoned pending drafts only.
  - Exception: auto-notification delivery smoke sends persisted disposable messages and acquires/renews/releases `mhcat_scheduler_locks`; it does not mutate `cron_sets`.
  - Exception: daily-reset smoke mutates disposable `coins.today`/`work_users.energi` and acquires/releases `daily-reset` in `mhcat_scheduler_locks`.
  - Exception: voice-room config smoke writes/deletes legacy-compatible `voice_channels` rows and, with gateway Voice State events enabled, creates/moves/deletes disposable dynamic rooms plus `voice_channel_ids`/lock seed rows.
  - Exception: voice-room lock smoke replaces one disposable `lock_channels` row and can add the authorized user to `ok_people`.
  - Exception: XP reset smoke deletes disposable staging `text_xps`/`voice_xps` rows only after an individual reset command or full-reset `^確認^` confirmation.
  - Exception: global usage tracking smoke increments `all_use_counts` once per slash attempt when `MHCAT_FEATURE_USAGE_TRACKING_ENABLED=true`.

## 6. Record Result

Record sanitized result locally under `.smoke/` if useful. Do not store tokens, interaction tokens, private user content, private channel IDs, or raw Mongo URIs.
