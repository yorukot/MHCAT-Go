# Announcement Slice

Status: historical implementation note, superseded by the parity-audited [announcement contract](76-announcement.md). Runtime and command sync remain behind explicit gates.

## Scope

This slice implements legacy `/公告頻道設置`:

- `一次性公告頻道`
- `綁定公告頻道`
- `綁定公告頻道刪除`

It preserves the legacy command name, option names, descriptions, Manage Messages permission requirement, public defer/edit response flow, success/error embed text, `client.color.greate` green, and legacy collection/field names.

This slice also implements explicitly gated legacy `/公告發送`:

- legacy `公告系統` modal title
- fields `anntag`, `anncolor`, `anntitle`, `anncontent`
- legacy field labels
- preview content and embed shape
- confirmation title, button labels, emojis, and colors
- missing announcement-channel text
- success text `<a:green_tick:994529015652163614> | 成功發送!`

This slice also implements the explicitly gated legacy bound relay from `events/ann_message.js`:

- watches `messageCreate` only when gateway, Guild Messages intent, Message Content intent, and `MHCAT_FEATURE_ANNOUNCEMENT_RELAY_ENABLED=true` are all enabled
- looks up `ann_all_sets` by legacy `{guild, announcement_id}`
- relays the user's message content into an embed with legacy title, color, and footer text `來自<author tag>的公告`
- sends `tag` as the message content and then deletes the original user message

## Data

- One-time announcement channel writes `guilds` field `announcement_id` by legacy `guild`.
- Bound announcement config writes `ann_all_sets` fields `guild`, `announcement_id`, `tag`, `color`, and `title`.
- Deletes remove matching `ann_all_sets` rows by `{guild, announcement_id}`.
- No index is created by startup.
- No Mongo repair/backfill is run.

The repository updates all matching duplicate rows before upsert where possible, matching the project's duplicate-tolerant Mongo rollout strategy.

## Gates

- Runtime: `MHCAT_FEATURE_ANNOUNCEMENT_CONFIG_ENABLED=true`
- Command sync: `MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_CONFIG=true`
- One-time send runtime: `MHCAT_FEATURE_ANNOUNCEMENT_SEND_ENABLED=true`
- One-time send command sync: `MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_SEND=true`
- Bound relay runtime: `MHCAT_FEATURE_ANNOUNCEMENT_RELAY_ENABLED=true`
- Bound relay required gateway flags: `MHCAT_DISCORD_ENABLE_GATEWAY=true`, `MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true`, `MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true`

Both must be paired in staging. `mhcat-staging-preflight` and the staging scripts reject unpaired command-sync/runtime flags.

## One-Time Send Safety

The legacy raw confirmation IDs `announcement_yes` and `announcement_no` are not used for new Go-generated buttons because legacy work flows reuse those IDs. New buttons use versioned custom IDs:

```txt
mhcat:v1:announcement:confirm:state=<id>
mhcat:v1:announcement:cancel:state=<id>
```

The preview and final send use empty allowed mentions. This intentionally fixes the legacy behavior that allowed user-entered `tag` text to ping `@everyone`, roles, or users before and after confirmation.

Announcement draft state is held in bounded in-memory storage for the short preview/confirm flow. No draft data is written to Mongo.

## Bound Relay Safety

The relay preserves the user-visible replacement-message UI but intentionally fixes two legacy failure modes:

- It sends the bot relay before deleting the original user message, so a Discord send failure does not lose the original content.
- It uses empty allowed mentions, so stored `tag` values such as `@everyone`, role mentions, or user mentions do not ping by default.

The relay accepts both `Random` and legacy `RANDOM` color values for read compatibility.

## Non-goals

This slice does not implement:

- tag mentions that actually ping users/roles/everyone
- attachment-only relay; empty-content messages are ignored instead of deleted

## Safety Notes

Go bot response messages and relay messages use empty allowed mentions by default. Any future change that allows real `tag` pings must add an allowlist and staging tests first.
