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
- parity-audited ticket setup/delete/open/close flow with Mongo compatibility, exact legacy routes/UI, failure compensation, and a disabled-by-default ownership gate; see the [ticket parity contract](docs/74-ticket.md).
- parity-audited poll create/vote/owner-menu/result/export flow with exact legacy UI, Mongoose-compatible reads, atomic writes, failure compensation, and a disabled-by-default ownership gate; see the [poll parity contract](docs/75-poll.md).
- economy `coins`/`gift_changes`/XP/work BSON compatibility, read-only query repository, gated `/代幣查詢` handler, gated `/代幣排行榜` PNG leaderboard, gated `my-profile` profile PNG, gated `/簽到` sign-in write slice, gated `coin-related-settings` config write slice, gated `/代幣增加` admin coin write slice, gated `/剪刀石頭布` game write slice, and gated `/代幣重製` owner-only reset slice.
- gated `打工系統` command schema, legacy dashboard-redirect UI for `新增打工事項`, legacy-style `打工介面` list/detail/start UI, and admin setup/delete/energy flows with explicit work repository writes.
- gated read-only `/警告紀錄` warning-history lookup with legacy embed text and safer permission/member-cache handling.
- gated `/警告設定` warning escalation config command with rollback-compatible `errors_sets` writes.
- gated `/警告清除` and `/警告全部清除` warning-removal commands with legacy embeds, best-effort DMs, and `warndbs` mutations.
- gated `/警告` warning-issue command with legacy embeds/DMs, `warndbs` appends, role hierarchy checks, and configured kick/ban threshold actions.
- gated `/刪除訊息` message cleanup command with legacy permission gates, ephemeral embeds, 1000-message cap, and 14-day Discord delete cutoff handling.
- gated `/刪除資料` destructive config cleanup command with the legacy warning select UI and guild-scoped deletes for the selected legacy config collection.
- gated `/翻譯` utility command with legacy loading/final embed shape and safe external-provider error handling.
- gated `/兌換` redeem-code command with legacy ephemeral success/error embeds and rollback-compatible `codes`/`chatgpt_gets` writes.
- gated auto-chat local fallback MessageCreate runtime with the legacy response corpus, `說出` replies, typing delay, and read-only `chats`/`chatgpt_gets` gating.
- gated paid auto-chat MessageCreate handoff with transactional `chatgpt_gets` debit plus rollback-compatible `chatgpts` publication, legacy cooldown/reset timing, and mention-safe worker replies.
- gated `automatic-notification`, `/自動通知列表`, and `/自動通知刪除` setup/list/delete commands plus a separately gated lease-backed recurring delivery worker for legacy `cron_sets` rows.
- parity-audited `/set-log-channel` config plus independently gated message update/delete, channel topic/permission, and voice join/leave emitters; exact UI, Mongo compatibility, ownership, and rollback are in the [logging parity contract](docs/48-logging-config.md).
- gated read-only `/扭蛋獎池查詢` prize-pool query with legacy embed text and rollback-compatible `gifts`/`gift_changes` reads.
- gated `/扭蛋獎池增加` prize add command with legacy Manage Messages permission, ephemeral success/error embeds, and rollback-compatible `gifts` inserts.
- gated `/扭蛋獎品編輯` prize edit command with legacy Manage Messages permission, ephemeral success/error embeds, and one-row `gifts` replacement by `{guild,gift_name}`.
- gated `/扭蛋獎池刪除` prize delete command with legacy Manage Messages permission, success/error embeds, and one-row `gifts` deletion by `{guild,gift_name}`.
- gated config-only `/公告頻道設置` command with legacy subcommands, embeds, and rollback-compatible `guilds`/`ann_all_sets` writes.
- gated bound announcement message relay from legacy `ann_message.js`, disabled by default and requiring explicit gateway/message-content flags.
- gated config-only `/聊天經驗設定` and `/聊天經驗刪除` commands with legacy embed/preview UI and rollback-compatible `text_xp_channels` writes.
- gated text XP message accrual from legacy `events/text_xp.js`, disabled by default and requiring explicit gateway/Guild Messages/Message Content flags; it updates `text_xps` profile XP/levels, sends configured level-up announcements and fallback errors, applies configured text reward roles, and grants legacy XP coin rewards from `gift_changes.xp_multiple` after the configured announcement path succeeds.
- gated voice XP runtime from legacy `events/voice_xp.js`, disabled by default and requiring explicit gateway/Voice State intent flags; it maintains `voice_xps.leavejoin`, starts the legacy 30-second voice XP loop on join, stops it on leave/shutdown, and preserves legacy voice announcements, owner fallbacks, roles, and XP coin rewards.
- gated `/聊天經驗身分組設定` and `/語音經驗身分組設定` reward-role config commands with legacy pagination UI and rollback-compatible `chat_roles`/`voice_roles` writes.
- gated disabled-response `/聊天經驗` and `/語音經驗` commands that preserve the legacy replacement message pointing users to `/我的檔案`.
- gated `/經驗值改變` XP admin command with legacy Kick Members permission, success/error embeds, and rollback-compatible `text_xps`/`voice_xps` writes.
- parity-audited, gated `/語音包廂設置` and `/語音包廂刪除` commands with legacy embed UI, rollback-compatible `voice_channels` writes/deletes, and dynamic room events when gateway voice states are enabled; see the [voice-room parity contract](docs/72-voice-room-config.md).
- parity-audited, gated `/上鎖頻道` password command and voice lock prompt/modal path with legacy embed UI and rollback-compatible `lock_channels.lock_anser`/`ok_people` writes; dynamic voice-room create/move/delete and `voice_channel_ids` cleanup are restored for configured triggers when gateway voice-state events are enabled.
- gated `guildMemberAdd` join-role assignment from legacy `join_roles`, disabled by default and requiring explicit Gateway + Guild Members intent.
- gated `guildMemberRemove` leave-message delivery from legacy `leave_messages`, disabled by default and requiring explicit Gateway + Guild Members intent.
- dry-run-first, lease-gated `mhcat-economy-reset` tool plus a separately gated `00:00 Asia/Taipei` recurring worker for the legacy `coins.today` reset and `work_users.energi` refill/clamp path.
- dry-run-first, lease-gated, crash-idempotent `mhcat-work-payout` one-shot tool plus a separately gated every-minute recurring worker for the legacy `handler/gift.js` completed-work payout path.
- Mongo-backed scheduler lease primitive, read-only-by-default diagnostic CLI, and lease-backed recurring consumers for automatic-notification delivery, economy daily reset, and work payout.
- staging guild command sync apply completed for managed `help`, `info`, and `ping`.
- gateway smoke completed with the gateway explicitly enabled for the smoke invocation.

Deliberately unavailable by default or still outstanding:

- prefix-command runtime; the legacy checkout has no active `commands/` tree to dispatch;
- command registration from `cmd/mhcat-bot`; command changes remain owned by the explicit sync CLI;
- default Mongo index creation and automatic data repair;
- production feature writes and recurring workers; implemented write paths remain behind disabled-by-default runtime, sync, ownership, and lease gates;
- production ticket/poll/economy rollout and live staging smoke, despite their gated runtime handlers and repositories being implemented;
- announcement relay attachment forwarding and tag pings; relay messages intentionally suppress mentions;
- external ChatGPT worker confirmation/deployment and dashboard-owned workflows.

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
- `/代幣重製` when explicitly enabled with `MHCAT_FEATURE_ECONOMY_COIN_RESET_ENABLED=true`, gateway enabled, Guild Messages intent enabled, and Message Content intent enabled
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
- `automatic-notification`, `/自動通知列表`, and `/自動通知刪除` when explicitly enabled with `MHCAT_FEATURE_AUTO_NOTIFICATION_CONFIG_ENABLED=true`
- `/set-log-channel` when explicitly enabled with `MHCAT_FEATURE_LOGGING_CONFIG_ENABLED=true`
- `/扭蛋獎池查詢` when explicitly enabled with `MHCAT_FEATURE_GACHA_PRIZE_LIST_ENABLED=true`
- `/扭蛋獎池增加` when explicitly enabled with `MHCAT_FEATURE_GACHA_PRIZE_CREATE_ENABLED=true`
- `/扭蛋獎品編輯` when explicitly enabled with `MHCAT_FEATURE_GACHA_PRIZE_EDIT_ENABLED=true`
- `/扭蛋獎池刪除` when explicitly enabled with `MHCAT_FEATURE_GACHA_PRIZE_DELETE_ENABLED=true`
- `/抽獎設置` disabled-command parity response when explicitly enabled with `MHCAT_FEATURE_LOTTERY_DISABLED_COMMAND_ENABLED=true`
- existing numeric `lotter*` buttons when explicitly enabled with `MHCAT_FEATURE_LOTTERY_COMPONENTS_ENABLED=true`
- `/統計系統查詢` when explicitly enabled with `MHCAT_FEATURE_STATS_QUERY_ENABLED=true`
- `/統計系統創建` when explicitly enabled with `MHCAT_FEATURE_STATS_CREATE_ENABLED=true`
- `/統計身分組人數` when explicitly enabled with `MHCAT_FEATURE_STATS_ROLE_COUNT_ENABLED=true`
- `/統計系統刪除` when explicitly enabled with `MHCAT_FEATURE_STATS_DELETE_ENABLED=true`
- `/公告頻道設置` when explicitly enabled with `MHCAT_FEATURE_ANNOUNCEMENT_CONFIG_ENABLED=true`
- `/公告發送` modal preview/confirm/send flow when explicitly enabled with `MHCAT_FEATURE_ANNOUNCEMENT_SEND_ENABLED=true`
- `/聊天經驗設定` and `/聊天經驗刪除` when explicitly enabled with `MHCAT_FEATURE_TEXT_XP_CONFIG_ENABLED=true`
- text XP message accrual, configured level-up announcements/fallbacks, configured text reward-role changes, and XP coin rewards when explicitly enabled with `MHCAT_FEATURE_TEXT_XP_ACCRUAL_ENABLED=true`, gateway enabled, Guild Messages intent enabled, and Message Content intent enabled
- `/語音經驗設定` and `/語音經驗刪除` when explicitly enabled with `MHCAT_FEATURE_VOICE_XP_CONFIG_ENABLED=true`
- voice XP session tracking and 30-second runtime accrual when explicitly enabled with `MHCAT_FEATURE_VOICE_XP_SESSIONS_ENABLED=true`, gateway enabled, and Voice State intent enabled
- `/聊天經驗身分組設定` and `/語音經驗身分組設定` when explicitly enabled with `MHCAT_FEATURE_XP_ROLE_CONFIG_ENABLED=true`
- `/聊天經驗` and `/語音經驗` disabled-command replacement responses when explicitly enabled with `MHCAT_FEATURE_XP_PROFILE_DISABLED_COMMANDS_ENABLED=true`
- `/經驗值改變` when explicitly enabled with `MHCAT_FEATURE_XP_ADMIN_ENABLED=true`
- `/經驗值重製` when explicitly enabled with `MHCAT_FEATURE_XP_RESET_ENABLED=true`, gateway enabled, Guild Messages intent enabled, and Message Content intent enabled
- `/聊天排行榜` and `/語音排行榜` when explicitly enabled with `MHCAT_FEATURE_XP_RANK_ENABLED=true`
- `/語音包廂設置` and `/語音包廂刪除` when explicitly enabled with `MHCAT_FEATURE_VOICE_ROOM_CONFIG_ENABLED=true`
- `/上鎖頻道` when explicitly enabled with `MHCAT_FEATURE_VOICE_ROOM_LOCK_ENABLED=true`, gateway enabled, and Voice State intent enabled
- `/加入身份組設置` and `/加入身份組刪除` when explicitly enabled with `MHCAT_FEATURE_JOIN_ROLE_CONFIG_ENABLED=true`
- `/加入訊息設置` and `/退出訊息設置` when explicitly enabled with `MHCAT_FEATURE_WELCOME_MESSAGE_CONFIG_ENABLED=true`
- `/驗證設置` when explicitly enabled with `MHCAT_FEATURE_VERIFICATION_CONFIG_ENABLED=true`
- `/驗證` captcha flow when explicitly enabled with `MHCAT_FEATURE_VERIFICATION_FLOW_ENABLED=true`
- `/帳號需創建時數` when explicitly enabled with `MHCAT_FEATURE_ACCOUNT_AGE_CONFIG_ENABLED=true`
- `/選取身分組-按鈕`, `/選取身分組-表情符號`, and `/選取身分組刪除-表情符號` when explicitly enabled with `MHCAT_FEATURE_ROLE_SELECTION_ENABLED=true`, gateway enabled, and Guild Message Reactions intent enabled

Implemented event features:

- Bound announcement relay when explicitly enabled with `MHCAT_FEATURE_ANNOUNCEMENT_RELAY_ENABLED=true`, `MHCAT_DISCORD_ENABLE_GATEWAY=true`, `MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true`, and `MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true`.
- Join-role assignment on member join when explicitly enabled with `MHCAT_FEATURE_JOIN_ROLE_ASSIGNMENT_ENABLED=true`, `MHCAT_DISCORD_ENABLE_GATEWAY=true`, and `MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true`.
- Welcome-message delivery on member join when explicitly enabled with `MHCAT_FEATURE_WELCOME_MESSAGE_DELIVERY_ENABLED=true`, `MHCAT_DISCORD_ENABLE_GATEWAY=true`, and `MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true`.
- Leave-message delivery on member leave when explicitly enabled with `MHCAT_FEATURE_LEAVE_MESSAGE_DELIVERY_ENABLED=true`, `MHCAT_DISCORD_ENABLE_GATEWAY=true`, and `MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true`.
- Account-age member join gate when explicitly enabled with `MHCAT_FEATURE_ACCOUNT_AGE_POLICY_ENABLED=true`, `MHCAT_DISCORD_ENABLE_GATEWAY=true`, and `MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true`.
- Reaction role add/remove handling when explicitly enabled with `MHCAT_FEATURE_ROLE_SELECTION_ENABLED=true`, `MHCAT_DISCORD_ENABLE_GATEWAY=true`, and `MHCAT_DISCORD_GUILD_MESSAGE_REACTIONS_INTENT=true`.

`/help` now uses the legacy embed/select-menu interface instead of the temporary text placeholder. The help menu can display legacy categories and command documentation entries before every linked feature handler is implemented; that mirrors the Node.js bot's help interface and does not mean those feature groups are runtime-complete.

`/info bot` now uses the legacy embed/button interface instead of the temporary text status output. The refresh button uses the legacy `botinfoupdate` custom ID through the typed parser/router path.

`/info shard` uses the legacy embed/button style and `shardinfoupdate` route, but it intentionally shows shard fields immediately instead of copying the old empty initial embed.

`/info user` and `/info guild` use the legacy embed layouts with read-only Discord snapshots. Lookup failures return a red safe error embed and do not expose internal errors.

Known external, intentionally inactive, or rollout gaps include lottery creation/panel generation, the disabled legacy XP profile-card branches, commented-out birthday delivery, announcement relay attachment forwarding/tag pings, production scheduler ownership, external ChatGPT worker confirmation/deployment, and dashboard-owned workflows. Economy game/shop writes, economy/profile/XP leaderboard images, and gacha prize delivery are implemented behind their explicit gates.

`/簽到` is a staging-gated write slice, not a production-ready economy rollout. Do not enable it against production until duplicate audits and unique-key/index plans for `coins`/`sign_lists` are complete and daily-reset ownership is assigned exclusively to the lease-backed one-shot or recurring Go path after Node cron is stopped.

`/代幣增加` is a disabled-by-default staging admin write slice. It requires Manage Messages, writes legacy-compatible `coins` rows, rejects negative balances and balances above `999999999`, and must be paired with `MHCAT_FEATURE_ECONOMY_COIN_ADMIN_ENABLED=true` plus `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_ADMIN=true` only against disposable staging data until duplicate audits and production ownership are reviewed.

`/剪刀石頭布` is a disabled-by-default staging game write slice. It writes existing `coins` rows only, rejects missing or insufficient balances, preserves legacy tie/win/loss wager behavior, and must be paired with `MHCAT_FEATURE_ECONOMY_RPS_ENABLED=true` plus `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_RPS=true` only against disposable staging data until duplicate audits and production ownership are reviewed.

`/代幣重製` is a disabled-by-default staging owner-only destructive write slice. It writes all `coins` rows for a guild by either deleting them when `除以多少` is omitted or `0`, or dividing balances with legacy rounding when a divisor is provided. It must be paired with `MHCAT_FEATURE_ECONOMY_COIN_RESET_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_RESET=true`, gateway, Guild Messages, and Message Content flags, and tested only against disposable staging balances.

Role selection is parity-audited as a disabled-by-default staging slice. `/選取身分組-表情符號` writes legacy-compatible `message_reactions` rows and adds the configured reaction to the target message; `/選取身分組刪除-表情符號` deletes the matching row; `/選取身分組-按鈕` writes legacy-compatible `btns` rows and opens the legacy `nal` modal to publish an add/delete button panel. The runtime must be paired with `MHCAT_FEATURE_ROLE_SELECTION_ENABLED=true`, `MHCAT_COMMAND_SYNC_INCLUDE_ROLE_SELECTION=true`, `MHCAT_DISCORD_ENABLE_GATEWAY=true`, and `MHCAT_DISCORD_GUILD_MESSAGE_REACTIONS_INTENT=true` for staging smoke. `MHCAT_FEATURE_ROLE_SELECTION_ENABLED` intentionally owns the setup commands, modal/buttons, and reaction events as one boundary; do not split them across separate Go ownership gates. Reaction-add member payloads are retained for later remove events, but a remove first observed after process restart has no member payload and remains best-effort when that member is absent from Discord state, including bot detection. Exact UI, data compatibility, reliability differences, ownership, smoke, and rollback rules are in the [role-selection parity contract](docs/73-role-selection.md).

Auto-notification setup/list/delete and delivery are independent, disabled-by-default paths. `MHCAT_FEATURE_AUTO_NOTIFICATION_CONFIG_ENABLED=true` plus `MHCAT_COMMAND_SYNC_INCLUDE_AUTO_NOTIFICATION_CONFIG=true` enables `automatic-notification`, `/自動通知列表`, and `/自動通知刪除`. Recurring sends require `MHCAT_FEATURE_AUTO_NOTIFICATION_DELIVERY_ENABLED=true`, Gateway, scheduler lease enablement, and a non-empty lease owner; the worker uses the `auto-notification-delivery` lease and `Asia/Taipei` schedule time. Disable the Node `handler/cron.js` owner before enabling Go delivery. See `docs/66-auto-notification-config.md`.

`/兌換` is disabled by default and is available only when paired with `MHCAT_FEATURE_REDEEM_ENABLED=true` and `MHCAT_COMMAND_SYNC_INCLUDE_REDEEM=true` in staging command sync. It consumes legacy `codes` rows, credits `chatgpt_gets.price`, enforces the legacy 7-day expiry check, and does not itself enable either auto-chat runtime.

The local auto-chat fallback is independently disabled by default. `MHCAT_FEATURE_AUTOCHAT_FALLBACK_ENABLED=true` requires gateway, Guild Messages, and Message Content flags; it reads the configured `chats` channel and `chatgpt_gets.price`, then replies from the bundled legacy corpus only for missing, negative, or malformed balances. Zero/nonnegative balances remain silent.

The paid auto-chat handoff is separately disabled by default. It requires `MHCAT_FEATURE_AUTOCHAT_PAID_HANDOFF_ENABLED=true`, `MHCAT_AUTOCHAT_PAID_OWNERSHIP_CONFIRMED=true`, gateway, Guild Messages, Message Content, a transaction-capable Mongo deployment, and the compatible external worker. It atomically debits `chatgpt_gets.price` and publishes the legacy `chatgpts` request, then reads the worker response after ten seconds. Do not set the ownership acknowledgment until the Node MessageCreate handler is stopped and worker/schema/duplicate audits are complete; see `docs/62-autochat-config.md`.

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
- `MHCAT_FEATURE_ECONOMY_COIN_RESET_ENABLED=false`
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
- `MHCAT_FEATURE_AUTOCHAT_CONFIG_ENABLED=false`
- `MHCAT_FEATURE_AUTOCHAT_FALLBACK_ENABLED=false`
- `MHCAT_FEATURE_AUTOCHAT_PAID_HANDOFF_ENABLED=false`
- `MHCAT_AUTOCHAT_PAID_OWNERSHIP_CONFIRMED=false`
- `MHCAT_FEATURE_AUTO_NOTIFICATION_CONFIG_ENABLED=false`
- `MHCAT_FEATURE_AUTO_NOTIFICATION_DELIVERY_ENABLED=false`
- `MHCAT_FEATURE_LOGGING_CONFIG_ENABLED=false`
- `MHCAT_FEATURE_GACHA_PRIZE_LIST_ENABLED=false`
- `MHCAT_FEATURE_GACHA_PRIZE_CREATE_ENABLED=false`
- `MHCAT_FEATURE_GACHA_PRIZE_EDIT_ENABLED=false`
- `MHCAT_FEATURE_GACHA_PRIZE_DELETE_ENABLED=false`
- `MHCAT_FEATURE_LOTTERY_DISABLED_COMMAND_ENABLED=false`
- `MHCAT_FEATURE_LOTTERY_COMPONENTS_ENABLED=false`
- `MHCAT_FEATURE_STATS_QUERY_ENABLED=false`
- `MHCAT_FEATURE_STATS_CREATE_ENABLED=false`
- `MHCAT_FEATURE_STATS_ROLE_COUNT_ENABLED=false`
- `MHCAT_FEATURE_STATS_DELETE_ENABLED=false`
- `MHCAT_FEATURE_ANNOUNCEMENT_CONFIG_ENABLED=false`
- `MHCAT_FEATURE_ANNOUNCEMENT_SEND_ENABLED=false`
- `MHCAT_FEATURE_ANNOUNCEMENT_RELAY_ENABLED=false`
- `MHCAT_FEATURE_TEXT_XP_CONFIG_ENABLED=false`
- `MHCAT_FEATURE_TEXT_XP_ACCRUAL_ENABLED=false`
- `MHCAT_FEATURE_VOICE_XP_CONFIG_ENABLED=false`
- `MHCAT_FEATURE_VOICE_XP_SESSIONS_ENABLED=false`
- `MHCAT_FEATURE_XP_ROLE_CONFIG_ENABLED=false`
- `MHCAT_FEATURE_XP_PROFILE_DISABLED_COMMANDS_ENABLED=false`
- `MHCAT_FEATURE_XP_ADMIN_ENABLED=false`
- `MHCAT_FEATURE_XP_RESET_ENABLED=false`
- `MHCAT_FEATURE_XP_RANK_ENABLED=false`
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
- `MHCAT_FEATURE_ROLE_SELECTION_ENABLED=false`
- `MHCAT_FEATURE_DAILY_RESET_SCHEDULER_ENABLED=false`
- `MHCAT_FEATURE_WORK_PAYOUT_SCHEDULER_ENABLED=false`
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
- `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_RESET=false`
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
- `MHCAT_COMMAND_SYNC_INCLUDE_XP_RANK=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_CONFIG=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_LOCK=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_JOIN_ROLE_CONFIG=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_WELCOME_MESSAGE_CONFIG=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_CONFIG=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_FLOW=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_ACCOUNT_AGE_CONFIG=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_ROLE_SELECTION=false`

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

The legacy Node bot runs a daily economy cron at `00:00 Asia/Taipei` on shard 0. Go provides both a dry-run/apply command and a separately gated recurring bot worker. Every Go write path uses lease name `daily-reset`.

Dry-run is the default and performs no writes:

```bash
go run ./cmd/mhcat-economy-reset --dry-run
```

Apply mode requires an explicit flag, the write gate, and lease ownership:

```bash
MHCAT_JOBS_DAILY_RESET_ENABLED=true \
MHCAT_SCHEDULER_LEASE_ENABLED=true \
MHCAT_SCHEDULER_LEASE_OWNER=staging-reset-cli \
go run ./cmd/mhcat-economy-reset --apply
```

The command:

- previews or resets `coins.today` for guilds not using rolling sign-in cooldowns;
- previews or refills/clamps `work_users.energi` from `work_sets.get_energy` and `work_sets.max_energy`;
- uses normalized `gift_changes.time` decoding instead of copying the legacy raw `$ne: 0` edge case;
- exits with code `2` without writes when another owner holds `daily-reset`;
- does not create indexes, repair data, or sync commands.

The recurring worker additionally requires `MHCAT_FEATURE_DAILY_RESET_SCHEDULER_ENABLED=true` and `MHCAT_DISCORD_ENABLE_GATEWAY=true`. It schedules `0 0 * * *` at fixed UTC+8 named `Asia/Taipei`, acquires/releases `daily-reset` for each tick, and performs no Discord API calls. Stop Node `handler/cron.js` before either Go apply or recurring ownership. See `docs/41-economy-daily-reset.md` for staging and rollback.

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
- conditionally increments `coins.coin` and writes `mhcat_work_payouts` in the same atomic Mongo pipeline update;
- skips the increment on a same-token crash retry and reports `idempotent_replays`;
- rejects duplicate `{guild,member}` coin rows before crediting an affected job;
- resets only the exact paid `work_users` job snapshot to `待業中`;
- fixes the legacy `gift_change.time == 0` new-balance bug by using `today=1` for daily-reset mode;
- requires the scheduler lease before apply writes;
- does not create indexes, repair data, sync commands, or send Discord messages.

Do not run apply against production until duplicate audits for `coins`, `work_users`, and `gift_changes` are reviewed, all `coins` consumers accept the additive marker field, and the Node.js bot is no longer owning the same payout loop. Leave markers in place during rollback; Node ignores them, while removing an in-flight marker can make a Go retry pay twice.

The separately gated recurring worker requires `MHCAT_FEATURE_WORK_PAYOUT_SCHEDULER_ENABLED=true`, Gateway, the existing work-payout job gate, and scheduler lease ownership. It schedules `* * * * *` at fixed `Asia/Taipei`, shares `MHCAT_JOBS_WORK_PAYOUT_LEASE_NAME` with CLI apply, skips local overlap, and releases the lease after every tick. `MHCAT_JOBS_WORK_PAYOUT_DRY_RUN` is CLI-only; enabling the recurring flag authorizes writes. The lease TTL must exceed payout timeout plus lease timeout. Stop legacy `handler/gift.js` before Go ownership. See `docs/43-work-payout.md`.

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

Current consumers and boundaries:

- automatic-notification delivery starts from `cmd/mhcat-bot` only when its delivery, Gateway, and lease gates are enabled;
- the recurring worker owns the `auto-notification-delivery` lease, renews it while active, and releases it during graceful shutdown;
- the one-shot and recurring daily-reset paths acquire `daily-reset` only around each write run; contenders skip and the holder releases after completion or failure;
- `mhcat-work-payout --apply` and the recurring every-minute worker share the configured payout lease; each holder releases after completion or failure;
- no lease indexes are created beyond MongoDB's default `_id` index;
- diagnostic CLI lease writes require `MHCAT_SCHEDULER_LEASE_ENABLED=true --apply`; recurring workers write leases only when all job-specific prerequisites pass config validation.

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

The `/代幣重製` command is available only when `MHCAT_FEATURE_ECONOMY_COIN_RESET_ENABLED=true`, `MHCAT_DISCORD_ENABLE_GATEWAY=true`, `MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true`, and `MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_RESET=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command requires the Discord guild owner and a same-channel `^確認^` message within 60 seconds before deleting or dividing all guild `coins` balances. Test only against disposable staging rows.

The role-selection commands are available only when `MHCAT_FEATURE_ROLE_SELECTION_ENABLED=true`, `MHCAT_DISCORD_ENABLE_GATEWAY=true`, and `MHCAT_DISCORD_GUILD_MESSAGE_REACTIONS_INTENT=true`. The feature flag is one runtime ownership gate for commands, modal/buttons, and reaction events rather than three independently deployable switches. To include the commands in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_ROLE_SELECTION=true`; staging preflight and scripts reject unpaired sync/runtime flags. See the [role-selection parity contract](docs/73-role-selection.md).

The `打工系統` command is available only when `MHCAT_FEATURE_WORK_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_WORK=true`; staging preflight rejects unpaired sync/runtime flags. The current work slices preserve the legacy dashboard redirect for `新增打工事項`, implement legacy-style `打工介面` list/captcha/detail/start UI, and implement `打工系統設定`, `打工事項刪除`, `增加個人精力`, and `增加全體精力` behind explicit admin repository wiring and Manage Messages checks. The start and energy paths can create/update `work_users` through atomic repository methods and do not write coins or payout state. Completed-work payout has crash-idempotent one-shot and separately gated recurring paths; production ownership remains disabled pending staging smoke and duplicate audits.

The `/警告紀錄` command is available only when `MHCAT_FEATURE_WARNINGS_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_WARNINGS=true`; staging preflight rejects unpaired sync/runtime flags. This command reads `warndbs` only and does not create, remove, or escalate warnings.

The `/警告設定` command is available only when `MHCAT_FEATURE_WARNING_SETTINGS_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_WARNING_SETTINGS=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command writes only `errors_sets.guild`, `ban_count`, and `move` using duplicate-friendly update/upsert behavior. It does not create warnings, delete messages, kick, ban, or run escalation.

The `/警告清除` and `/警告全部清除` commands are available only when `MHCAT_FEATURE_WARNING_REMOVAL_ENABLED=true`. To include them in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_WARNING_REMOVAL=true`; staging preflight and scripts reject unpaired sync/runtime flags. These commands mutate only `warndbs`, preserve the legacy public success/error embeds, and send legacy-style best-effort DMs. They do not create warnings, delete messages, kick, ban, or run escalation.

The `/警告` command is available only when `MHCAT_FEATURE_WARNING_ISSUE_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_WARNING_ISSUE=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command appends legacy `warndbs.content` entries with Asia/Taipei timestamps, preserves the legacy public success/error embeds and target DM, enforces Manage Messages and moderator-vs-target role hierarchy, and reads `errors_sets` to run configured `停權`/`踢出` threshold actions for existing warning records. Test only against disposable staging warning data because it can kick or ban members when thresholds are met.

The `/刪除訊息` command is available only when `MHCAT_FEATURE_MESSAGE_CLEANUP_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_MESSAGE_CLEANUP=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command deletes recent Discord messages only, requires Manage Messages, requires Administrator above 200 requested messages, refuses more than 1000, uses ephemeral legacy completion/error embeds, and does not write Mongo data. Test only in disposable staging channels.

The `/刪除資料` command is available only when `MHCAT_FEATURE_DELETE_DATA_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_DELETE_DATA=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command requires Manage Messages, presents the legacy destructive select UI, and deletes guild-scoped rows for the selected legacy config target: join messages, leave messages, audit logs, stats config, autochat config, verification config, text/voice XP config, or ticket config. Test only against disposable staging config data.

The `/翻譯` command is available only when `MHCAT_FEATURE_TRANSLATE_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_TRANSLATE=true`; staging preflight rejects unpaired sync/runtime flags. This command calls an external Google Translate-compatible endpoint through a provider port, does not require Message Content intent, and does not touch Mongo feature data.

The `/set-log-channel` command is available only when `MHCAT_FEATURE_LOGGING_CONFIG_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_LOGGING_CONFIG=true`; staging preflight rejects unpaired sync/runtime flags. The command remains publicly discoverable, checks Manage Messages at runtime, preserves the exact yellow/red UI and owner-scoped ten-minute select, and writes typed legacy-compatible `loggings` values while event reads accept mixed Mongoose-compatible scalars. Message, channel, and voice emitters have separate disabled-by-default gates and can consume existing config without enabling this command. Stop Node `events/LoggingSystem.js` ownership before enabling any Go logging event gate. See the [logging parity contract](docs/48-logging-config.md).

The `/扭蛋獎池查詢` command is available only when `MHCAT_FEATURE_GACHA_PRIZE_LIST_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_LIST=true`; staging preflight rejects unpaired sync/runtime flags. This command reads `gifts` and `gift_changes` only. It does not draw prizes, decrement inventory, mutate coins, send DMs, create indexes, or enable shop behavior.

The `/扭蛋獎池增加` command is available only when `MHCAT_FEATURE_GACHA_PRIZE_CREATE_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_CREATE=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command requires Manage Messages and inserts one legacy `gifts` row after the legacy duplicate-name and prize-count checks. It does not draw prizes, decrement inventory counts, mutate coins outside the prize config row, send DMs, create indexes, or enable shop behavior. Test only against disposable staging gacha data.

The `/扭蛋獎品編輯` command is available only when `MHCAT_FEATURE_GACHA_PRIZE_EDIT_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_EDIT=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command requires Manage Messages and replaces one legacy `gifts` row matching `{guild,gift_name}` with merged submitted/default values. It follows the legacy delete-plus-insert shape and has no transaction/rollback path, so test only against disposable staging gacha data. It does not draw prizes, decrement inventory counts, mutate user coin balances, send DMs, create indexes, or enable shop behavior.

The `/扭蛋獎池刪除` command is available only when `MHCAT_FEATURE_GACHA_PRIZE_DELETE_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_DELETE=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command requires Manage Messages and deletes one legacy `gifts` row matching `{guild,gift_name}`, returning the legacy success/error embeds. It does not draw prizes, decrement inventory counts, mutate coins, send DMs, create indexes, or enable shop behavior. Test only against disposable staging gacha data.

The `/抽獎設置` disabled-command parity response is available only when `MHCAT_FEATURE_LOTTERY_DISABLED_COMMAND_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_LOTTERY_DISABLED_COMMAND=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command preserves the current legacy unavailable embed and does not create lottery rows, send lottery panels, or enable old buttons.

Existing numeric `lotter*` buttons are independently available only when `MHCAT_FEATURE_LOTTERY_COMPONENTS_ENABLED=true` with Gateway enabled. This runtime reads existing rollback-compatible `lotters` rows, atomically appends eligible participants, returns participant search/export UI, sets `end:true` for stop/reroll, and sends one winner message on reroll. It does not enable lottery creation or command sync; test only against disposable staging rows and channels.

The `/統計系統查詢` command is available only when `MHCAT_FEATURE_STATS_QUERY_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_STATS_QUERY=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command preserves the legacy static stats help embed and does not read/write Mongo, create/delete channels, rename channels, create indexes, or enable `channel_status` scheduler behavior.

The `/統計系統創建` command is available only when `MHCAT_FEATURE_STATS_CREATE_ENABLED=true`, `MHCAT_DISCORD_ENABLE_GATEWAY=true`, and `MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_STATS_CREATE=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command requires Manage Messages, creates the legacy stats category plus base member/user/bot channels, can add the legacy channel-count/text-count/voice-count stat channels after the base row exists, and writes rollback-compatible `numbers` rows. It does not delete Discord channels, create indexes, or enable the `channel_status` rename scheduler. Test only in an isolated staging guild/database because it creates Discord channels and writes `numbers`.

The `/統計身分組人數` command is available only when `MHCAT_FEATURE_STATS_ROLE_COUNT_ENABLED=true`, `MHCAT_DISCORD_ENABLE_GATEWAY=true`, and `MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_STATS_ROLE_COUNT=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command requires Manage Messages, requires an existing `/統計系統創建` base config, creates a text or voice stat channel named `<role name>: <member count>`, and replaces the legacy `role_numbers` row for `{guild,role}`. It does not delete old stat channels, create indexes, or enable the `channel_status` rename scheduler. Test only in an isolated staging guild/database because it creates Discord channels and writes `role_numbers`.

The `/統計系統刪除` command is available only when `MHCAT_FEATURE_STATS_DELETE_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_STATS_DELETE=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command requires Manage Messages, deletes legacy `numbers` rows for the guild, and preserves the legacy success/error embeds. It does not delete Discord channels, create indexes, or enable `channel_status`.

The `/公告頻道設置` command is available only when `MHCAT_FEATURE_ANNOUNCEMENT_CONFIG_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_CONFIG=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command writes only the legacy-compatible `guilds.announcement_id` and `ann_all_sets` config rows and does not enable Message Content relay, user-message deletion, or bound announcement sends.

The `/公告發送` command is available only when `MHCAT_FEATURE_ANNOUNCEMENT_SEND_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_SEND=true`; staging preflight and scripts reject unpaired sync/runtime flags. It preserves the legacy modal labels, preview embed, confirmation title, button labels/emojis, missing-config text, and success text, but uses versioned custom IDs and suppresses mentions in the preview and final send as an intentional safety fix for legacy tag-ping behavior.

The bound announcement relay is available only when `MHCAT_FEATURE_ANNOUNCEMENT_RELAY_ENABLED=true` with gateway, Guild Messages intent, and Message Content intent explicitly enabled. It reads existing `ann_all_sets` rows, sends a legacy-style embed in the bound channel, then deletes the original message after the send succeeds. It suppresses mentions in the stored `tag` value and ignores empty-content messages as intentional safety fixes.

The `/聊天經驗設定` and `/聊天經驗刪除` commands are available only when `MHCAT_FEATURE_TEXT_XP_CONFIG_ENABLED=true`. To include them in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_TEXT_XP_CONFIG=true`; staging preflight rejects unpaired sync/runtime flags. These commands write only the legacy-compatible `text_xp_channels` config and do not enable text XP accrual, rank cards, voice XP, Message Content intent, or XP reward behavior. Text XP message accrual is a separate event gate: `MHCAT_FEATURE_TEXT_XP_ACCRUAL_ENABLED=true` requires `MHCAT_DISCORD_ENABLE_GATEWAY=true`, `MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true`, and `MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true`; it updates rollback-compatible `text_xps` rows using the legacy message XP formula and level reset behavior, sends the configured/default level-up announcement when a `text_xp_channels` row exists, sends legacy fallback errors when the configured announcement channel is missing or unusable, applies configured `chat_roles` reward-role changes on level-up, and grants legacy XP coin rewards from `gift_changes.xp_multiple` after the configured announcement path succeeds. Profile cards remain disabled.

The `/語音經驗設定` and `/語音經驗刪除` commands are available only when `MHCAT_FEATURE_VOICE_XP_CONFIG_ENABLED=true`. To include them in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_VOICE_XP_CONFIG=true`; staging preflight rejects unpaired sync/runtime flags. These commands write only the legacy-compatible `voice_xp_channels` config, preserve the old visible `背景` option without saving it, and do not enable Voice State intent, rank cards, or runtime XP reward behavior. Voice XP runtime is a separate event gate: `MHCAT_FEATURE_VOICE_XP_SESSIONS_ENABLED=true` requires `MHCAT_DISCORD_ENABLE_GATEWAY=true` and `MHCAT_DISCORD_VOICE_STATE_INTENT=true`, creates missing `voice_xps` rows with legacy string `xp`/`leavel`, marks `leavejoin` as `join` or `leave`, and starts one legacy 30-second XP loop per active joined user. The runtime preserves legacy `+5 XP` ticks, configured/default level-up announcements, owner DM fallbacks, `voice_roles` changes, and XP coin rewards from `gift_changes.xp_multiple`. On startup it also reconciles persisted `leavejoin:"join"` rows and starts one deduplicated worker per active profile before registering new Voice State events.

The `/聊天經驗身分組設定` and `/語音經驗身分組設定` commands are available only when `MHCAT_FEATURE_XP_ROLE_CONFIG_ENABLED=true`. To include them in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_XP_ROLE_CONFIG=true`; staging preflight and scripts reject unpaired sync/runtime flags. These commands add, delete, query, and paginate legacy-compatible `chat_roles` and `voice_roles` reward-role config rows with the legacy misspelled `leavel` field. They require Manage Messages and verify the selected role is assignable by the bot before saving. These commands alone do not enable XP accrual, rank cards, coin rewards, Message Content intent, Guild Messages intent, Voice State intent, or usage-counter writes; text reward-role assignment/removal runs only through the separate text accrual event gate. Exact UI behavior, preserved query quirks, intentional Go differences, and staging checks are recorded in `docs/68-xp-reward-role-config.md`.

The `/聊天經驗` and `/語音經驗` disabled-command replacement responses are available only when `MHCAT_FEATURE_XP_PROFILE_DISABLED_COMMANDS_ENABLED=true`. To include them in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_XP_PROFILE_DISABLED_COMMANDS=true`; staging preflight and scripts reject unpaired sync/runtime flags. These commands only preserve the legacy red embed telling users to use `/我的檔案`; they do not read XP collections, render rank cards, award XP, or write Mongo data.

The `/經驗值改變` command is available only when `MHCAT_FEATURE_XP_ADMIN_ENABLED=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_XP_ADMIN=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command requires Kick Members, adjusts one member's `text_xps` or `voice_xps` row with legacy `xp`/`leavel` strings, and sets `voice_xps.leavejoin=leave` when creating a voice profile. Test only against disposable staging XP rows until duplicate audits and XP ownership are reviewed. It does not enable XP accrual, rank cards, automatic role assignment/removal, coin rewards, gateway intents, or usage-counter writes. Exact adjustment quirks, payloads, compatibility behavior, and staging checks are recorded in `docs/69-xp-admin.md`.

The `/經驗值重製` command is available only when `MHCAT_FEATURE_XP_RESET_ENABLED=true`, `MHCAT_DISCORD_ENABLE_GATEWAY=true`, `MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true`, and `MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_XP_RESET=true`; staging preflight and scripts reject unpaired sync/runtime flags. This command preserves the legacy owner-only check, immediate individual text/voice XP deletes, and the full-server `^確認^` message confirmation before deleting all `text_xps` or `voice_xps` rows for a guild. Test only against disposable staging XP rows. Exact payloads, empty-collection quirks, confirmation state, intentional differences, and staging checks are recorded in `docs/70-xp-reset.md`.

The `/聊天排行榜` and `/語音排行榜` commands are available only when `MHCAT_FEATURE_XP_RANK_ENABLED=true`. To include them in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_XP_RANK=true`; staging preflight and scripts reject unpaired sync/runtime flags. These commands read `text_xps`/`voice_xps`, render legacy-style `user-info.png` leaderboard pages with legacy rank buttons, and write no Mongo data. They do not enable XP accrual, `/聊天經驗` profile cards, automatic reward roles, coin rewards, gateway intents, or usage-counter writes. Exact ranking math, payloads, canvas behavior, malformed-component hardening, asset requirements, intentional differences, and staging checks are recorded in `docs/71-xp-rank.md`.

The `/語音包廂設置` and `/語音包廂刪除` commands are available only when `MHCAT_FEATURE_VOICE_ROOM_CONFIG_ENABLED=true`. To include them in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_CONFIG=true`; staging preflight and scripts reject unpaired sync/runtime flags. Their definitions are publicly discoverable like legacy, while handlers require Manage Messages. Setup preserves raw names and explicit `0` limits; deletion preserves the legacy type-2 voice versus category/stage branch. When the app is also running the gateway with `MHCAT_DISCORD_VOICE_STATE_INTENT=true`, configured human or bot joins create/move users into dynamic voice rooms under the trigger's current parent, persist `voice_channel_ids`, seed lockable `lock_channels` rows, and delete empty tracked rooms. See the [voice-room parity contract](docs/72-voice-room-config.md).

The `/上鎖頻道` command is available only when `MHCAT_FEATURE_VOICE_ROOM_LOCK_ENABLED=true`, `MHCAT_DISCORD_ENABLE_GATEWAY=true`, and `MHCAT_DISCORD_VOICE_STATE_INTENT=true`. To include it in staging command-sync dry-run/apply, also set `MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_LOCK=true`; staging preflight and scripts reject unpaired sync/runtime flags. This publicly discoverable command reads the invoking member's current voice state, verifies the existing `lock_channels` row owner, and replaces the row with a raw nullable `lock_anser`, `owner`, nullable `text_channel`, and empty BSON-array `ok_people`. Existing passworded rows prompt and disconnect unauthorized users, bind the generated button to that user for exactly 60 seconds, preserve exact password matching, and add correct users to `ok_people` through the legacy answer modal. Scalar/null compatibility, exact UI, ownership, staging, and rollback details are in the [voice-room parity contract](docs/72-voice-room-config.md).

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

`mhcat-mongo-index` defaults to dry-run and compares the local partial index plan with live indexes. It never drops indexes.

```bash
MHCAT_MONGODB_URI='<uri>' \
MHCAT_MONGODB_DATABASE=mhcat \
go run ./cmd/mhcat-mongo-index --dry-run --format json
```

Index apply is not default. Missing safe indexes can only be created with explicit `--apply`. Unique indexes additionally require `--allow-unique` and a clean duplicate audit. TTL indexes require `--allow-ttl` and a retention ADR/note in the index plan.

Feature repositories and gated Mongo write paths now exist, but production writes remain disabled by default. The index tool is still operator-invoked only, and bot startup still performs no command registration or automatic index mutation.

## Custom ID Parsing

The shared parser layer handles components and modals. New Go-generated IDs use:

```txt
mhcat:v1:<feature>:<action>:<payload>
```

Encoded custom IDs are length-checked against Discord's 100-character `custom_id` limit. Payloads are bounded and must not contain secrets or raw untrusted text. If future feature data is too large or sensitive for a custom ID, the feature should store state separately and encode only a state reference.

Legacy live-message compatibility is handled through explicit decoders for documented high-confidence IDs, including ticket buttons, polls, verification prompts, rank pagination, sign/profile buttons, role buttons, voice-lock prompts, shop/game controls, and setup modals. Ambiguous legacy IDs return typed parse errors instead of being routed through broad substring matching.

The typed parser/router is used by help/info refresh, ticket, poll, sign/profile/rank pagination, role selection, economy shop/game, verification, voice lock, logging, lottery, work, birthday, notification, announcement, and XP component/modal paths. Feature gates still control whether each parsed route is registered; parser coverage and 74/74 command-definition parity are not substitutes for runtime behavior and UI verification.

## Ticket Slice

The ticket slice is parity-audited as a disabled-by-default private-channel setup/open/close flow behind explicit wiring:

- `私人頻道設置` command definition and handler.
- `私人頻道刪除` command definition and handler.
- Legacy setup modal title and field labels.
- Ticket panel embed with legacy `tic` open button.
- `tickets` config is saved only after valid modal submit, with stale-modal rejection and identity-scoped rollback when panel delivery fails.
- `tic` creates a private text channel with legacy permission overwrites and welcome UI.
- Failed welcome delivery removes the newly created channel so retry is not blocked.
- `del` deletes the ticket channel for its user-ID owner or an actor with Manage Messages.

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

Exact UI, color handling, Mongoose scalar compatibility, duplicate/index policy, Node/Go ownership, staging smoke, and rollback are recorded in the [ticket parity contract](docs/74-ticket.md).

Remaining rollout work:

- Production ticket command sync.
- Live staging smoke using the canonical ticket checklist.

## Poll Slice

The poll flow is parity-audited behind explicit flags:

- `投票創建` command definition and handler.
- Legacy-style public poll embed, choice buttons, result button, and owner select menu.
- Versioned Go-generated custom IDs for new poll components, while legacy `poll_<choice>`, `see_result`, and `poll_menu` still decode for live old messages.
- Mongoose-compatible `polls` reads and rollback-compatible typed writes, including `join_member[].choise`.
- Atomic vote add/remove and owner-toggle semantics, cancellation-safe creation rollback, and centralized slash-only usage tracking.
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

Exact UTF-16 validation, raw whitespace, initial/dynamic labels, colors, percentage rounding, legacy component migration, Mongo/index policy, exclusive ownership, smoke, and rollback constraints are recorded in the [poll parity contract](docs/75-poll.md).

Remaining rollout work:

- Production poll command sync.
- Live staging smoke using the canonical poll checklist.

## Utility Feature Tests

Run the utility feature pipeline tests with:

```bash
go test ./internal/core/features ./internal/core/services/utility ./internal/discord/features/utility ./internal/discord/commands ./internal/discord/interactions
```

Usage tracking is wired through the global slash middleware and remains disabled by default. Enable the Mongo-backed tracker only with:

```bash
MHCAT_FEATURE_USAGE_TRACKING_ENABLED=true
```

When enabled, each slash attempt that reaches the interaction router atomically increments `all_use_counts.count` by `slashcommand_name` before route lookup, permission checks, and handler work. Unknown, denied, and handler-failing slash attempts are counted; components, modals, and autocomplete interactions are not. The write has a 500 ms timeout, normalizes missing or malformed legacy counts to zero, and cannot fail the command path.

Stop the Node `events/SlashCommands.js` usage-counter owner before enabling this gate or every slash attempt can be counted twice. Use disposable staging rows first, and audit duplicate plus null/blank `slashcommand_name` rows before applying any unique index. Elsewhere in the documentation, a feature statement that it does not write usage counters refers to the feature handler itself and assumes this separate global gate is disabled.

## Mongo Integration Test

Mongo integration tests are skipped by default. To run them:

```bash
MHCAT_RUN_MONGO_INTEGRATION_TESTS=true \
MHCAT_MONGODB_URI='<uri>' \
MHCAT_MONGODB_DATABASE=mhcat \
go test ./internal/adapters/mongo
```

Do not use production credentials for local tests unless the environment is explicitly approved for read-only connectivity checks.
