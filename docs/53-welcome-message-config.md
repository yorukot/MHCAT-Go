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

Duplicate-row policy:

- reads and modal defaults retain legacy `findOne` first-row behavior;
- an explicit channel or modal save updates every existing `leave_messages` row for the guild to the same values;
- this intentionally fixes the legacy sequence of first-match, non-awaited `updateOne` calls, which could leave duplicate rows with conflicting channel/content/title/color combinations;
- rows are not deleted, merged, or backfilled, so rollback to Node remains compatible after an explicit Go save;
- no startup repair or index creation runs. The candidate unique `{guild:1}` index remains blocked until a live duplicate, missing/null/blank key, scalar-type, and external-writer audit is clean.

No database migration is required to enable this slice. Before command ownership moves to Go, run the read-only Mongo audit and inspect `leave_messages` duplicate/type findings. Keep Node and Go setup ownership exclusive during rollout; do not deduplicate or apply the candidate unique index merely to enable the feature.

## Leave Delivery

When `MHCAT_FEATURE_LEAVE_MESSAGE_DELIVERY_ENABLED=true`, the Go bot listens for `guildMemberRemove`, reads the existing `leave_messages` row, and sends the legacy-style leave embed to the configured channel.

Preserved legacy behavior:

- embed title comes from `title` without placeholder replacement;
- embed description comes from `message_content`;
- placeholders `(MEMBERNAME)`, `(ID)`, `{ID}`, and `{MEMBERNAME}` are replaced once in the description, matching the legacy `String.replace` order;
- raw usernames, including leading/trailing spaces, are retained, and replacement values preserve JavaScript handling for `$$`, `$&`, ``$` ``, and `$'`;
- nonempty all-space titles and message bodies remain configured values, matching JavaScript truthiness;
- embed color uses the stored color or `Random`;
- the leaving member's guild-specific display avatar is used as the thumbnail when available;
- the embed timestamp is set.

Safety behavior:

- missing or incomplete `leave_messages` config is a no-op;
- allowed mentions are suppressed;
- the event path performs no Mongo writes and registers no slash commands.
- delivery failures continue to later independent event handlers while still being returned for gateway logging.

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
- raw usernames and JavaScript replacement-string tokens are preserved;
- nonempty all-space message bodies are delivered;
- `{TAG}` and `(TAG)` render as `<@userID>`;
- thumbnail is the joining member's guild-specific display avatar when available;
- color uses the pinned Discord.js 14.25.1 color table or six-digit hex; exact `Random` and `RANDOM` values randomize;
- `img` is used as the embed image when present;
- timestamp is set.

Safety behavior:

- the event path performs no Mongo writes and registers no slash commands;
- only the joining user mention is allowed for legacy tag placeholders; everyone and role mentions remain suppressed;
- account-age policy registers before welcome delivery and stops propagation after a kick;
- welcome delivery is registered before join-role assignment so role-assignment failures do not suppress the welcome embed, matching legacy's non-awaited role callback.
- welcome read/send failures continue to join-role assignment and remain visible to gateway error logging.

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

All seven values must be set together. This preserves the special legacy welcome UI without hardcoding private guild/channel IDs in Go. The special author icon uses the bot's guild-specific display avatar when cached, and the member tag intentionally keeps the legacy manual `username#discriminator` form, including `username#0` for migrated usernames.

## Intentionally Not Implemented

- join-message modal/config write flow, because the current legacy slash command redirects to dashboard.
- verification/captcha and account-age policy remain separately gated from welcome delivery.
- production index creation.

## Known Legacy Bugs Preserved

- `/加入訊息設置` dashboard URL is malformed: `https://mhcat.yorukot.meguilds/<guild>/welcome`.
- join command docs URL is malformed: `https://docsmhcat.yorukot.meocs/join_message`.
- leave-message option/label wording says “加入訊息”.
- colors accepted by `validate-color` but rejected by Discord.js were historically written before preview construction failed; Go rejects those unusable values before writing.

## Next Step

Implement one of:

- staging smoke for leave-message delivery in an isolated guild; or
- next feature parity slice from the checklist.
