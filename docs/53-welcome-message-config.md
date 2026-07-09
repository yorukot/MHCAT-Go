# Welcome Message Config Slice

Status: implemented behind explicit flags.

## Scope

Implemented:

- `/加入訊息設置`
- `/退出訊息設置`
- legacy `nal` leave-message modal submit route
- legacy-compatible `leave_messages` document writes
- optional `guildMemberAdd` welcome-message delivery behind a separate event flag
- optional `guildMemberRemove` leave-message delivery behind a separate event flag

Flags:

- `MHCAT_FEATURE_WELCOME_MESSAGE_CONFIG_ENABLED=false`
- `MHCAT_COMMAND_SYNC_INCLUDE_WELCOME_MESSAGE_CONFIG=false`
- `MHCAT_FEATURE_WELCOME_MESSAGE_DELIVERY_ENABLED=false`
- `MHCAT_FEATURE_LEAVE_MESSAGE_DELIVERY_ENABLED=false`

Both flags must be paired for staging command sync. The bot still does not sync/register commands from startup.

The delivery flag is event-only and has no command-sync include flag. It additionally requires:

- `MHCAT_DISCORD_ENABLE_GATEWAY=true`
- `MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true`

## Legacy UI Preserved

`/加入訊息設置` preserves the current legacy behavior: the required `頻道` option exists, but the handler replies with the dashboard redirect embed and link button only. It does not write `join_messages`.

`/退出訊息設置` preserves:

- Manage Messages runtime permission error text;
- modal custom ID `nal`;
- modal title `退出訊息設置!`;
- field IDs `leave_msgcolor`, `leave_msgtitle`, `leave_msgcontent`;
- legacy field labels, including the old “加入訊息” wording;
- preview content typo: `((MEMBERNAME)在到時候會變正常喔)`.

## Compatibility Fix

Legacy `events/modal.js` routes modal submits by the first text input custom ID. The actual leave-message modal first field is `leave_msgcolor`, not `leave_msgcontent`. The Go legacy modal parser now accepts both `leave_msgcolor` and `leave_msgcontent` for `welcome/leave_submit`.

## Mongo Behavior

Collection: `leave_messages`

Prepare command:

- if no document exists for guild, upserts `{ guild, channel, message_content: null, title: null, color: null }`;
- if documents exist, updates `channel` for all matching guild documents and uses the first loaded document for modal defaults.

Modal submit:

- validates legacy color or exact `Random`;
- updates `message_content`, `title`, and `color` in one atomic `$set`;
- does not create indexes;
- does not write `join_messages`.

## Leave Delivery

When `MHCAT_FEATURE_LEAVE_MESSAGE_DELIVERY_ENABLED=true`, the Go bot listens for `guildMemberRemove`, reads the existing `leave_messages` row, and sends the legacy-style leave embed to the configured channel.

Preserved legacy behavior:

- embed title comes from `title` without placeholder replacement;
- embed description comes from `message_content`;
- placeholders `(MEMBERNAME)`, `(ID)`, `{ID}`, and `{MEMBERNAME}` are replaced once in the description, matching the legacy `String.replace` order;
- embed color uses the stored color or `Random`;
- the leaving member avatar is used as the thumbnail;
- the embed timestamp is set.

Safety behavior:

- missing or incomplete `leave_messages` config is a no-op;
- allowed mentions are suppressed;
- the event path performs no Mongo writes and registers no slash commands.

## Welcome Delivery

When `MHCAT_FEATURE_WELCOME_MESSAGE_DELIVERY_ENABLED=true`, the Go bot listens for `guildMemberAdd`, reads the existing dashboard/legacy `join_messages` row, and sends the legacy-style welcome embed to the configured channel.

Required runtime flags:

- `MHCAT_DISCORD_ENABLE_GATEWAY=true`
- `MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true`

Preserved legacy behavior:

- missing `join_messages` config is a no-op;
- `enable === false` disables delivery, while missing `enable` is treated as enabled for legacy compatibility;
- embed author name is `🪂 歡迎加入 <guild name>`;
- author icon uses guild icon, with bot avatar fallback;
- description replaces one occurrence each of `(MEMBERNAME)`, `{MEMBERNAME}`, `{membername}`, `(TAG)`, `{TAG}`, and `{tag}` in legacy order;
- `{TAG}` and `(TAG)` render as `<@userID>`;
- thumbnail is the joining member avatar;
- color uses stored hex/CSS color, or random only when stored color is exactly `RANDOM`;
- `img` is used as the embed image when present;
- timestamp is set.

Safety behavior:

- the event path performs no Mongo writes and registers no slash commands;
- only the joining user mention is allowed for legacy tag placeholders; everyone and role mentions remain suppressed;
- account-age policy registers before welcome delivery and stops propagation after a kick;
- welcome delivery is registered before join-role assignment so role-assignment failures do not suppress the welcome embed, matching legacy's non-awaited role callback.

Special MHCAT welcome:

Legacy had hardcoded guild, bot, and channel IDs for the MHCAT server welcome embed. The Go refactor does not hardcode those IDs. Set all three values together only in the intended staging/production environment:

```bash
MHCAT_LEGACY_WELCOME_SPECIAL_GUILD_ID=<guild-id>
MHCAT_LEGACY_WELCOME_SPECIAL_BOT_ID=<bot-id>
MHCAT_LEGACY_WELCOME_SPECIAL_CHANNEL_ID=<channel-id>
MHCAT_LEGACY_WELCOME_SPECIAL_CHAT_CHANNEL_ID=<chat-channel-id>
MHCAT_LEGACY_WELCOME_SPECIAL_HELP_CHANNEL_ID=<help-channel-id>
MHCAT_LEGACY_WELCOME_SPECIAL_BUG_CHANNEL_ID=<bug-report-channel-id>
MHCAT_LEGACY_WELCOME_SPECIAL_SUPPORT_CHANNEL_ID=<support-channel-id>
```

All seven values must be set together. This preserves the special legacy welcome UI without hardcoding private guild/channel IDs in Go.

## Intentionally Not Implemented

- join-message modal/config write flow, because the current legacy slash command redirects to dashboard.
- verification/captcha is not enabled by this welcome-message slice; `/驗證` is controlled by the separate verification-flow gate. Account-age kick remains separate.
- production index creation.

## Known Legacy Bugs Preserved

- `/加入訊息設置` dashboard URL is malformed: `https://mhcat.yorukot.meguilds/<guild>/welcome`.
- join command docs URL is malformed: `https://docsmhcat.yorukot.meocs/join_message`.
- leave-message option/label wording says “加入訊息”.

## Next Step

Implement one of:

- staging smoke for leave-message delivery in an isolated guild; or
- next feature parity slice from the checklist.
