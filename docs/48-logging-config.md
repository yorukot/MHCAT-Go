# Logging Config and Event Parity Contract

Status: `/set-log-channel` configuration and the active message, channel, and voice logging paths are parity-audited behind disabled-by-default, independent gates.

## Legacy References

- Config command: `MHCAT/slashCommands/管理系統/create_logging.js`
- Event runtime: `MHCAT/events/LoggingSystem.js`
- Model: `MHCAT/models/logging.js`
- Command: `set-log-channel`, localized as `設置日誌` and `设置日志`
- Legacy component: `loggin_create`
- Collection: `loggings`

The active legacy event file ends at the unterminated `/*` on `LoggingSystem.js:364`. Everything after that boundary is commented out and is not part of the parity target.

## Independent Gates

Config runtime and staging command sync:

```bash
MHCAT_FEATURE_LOGGING_CONFIG_ENABLED=true
MHCAT_COMMAND_SYNC_INCLUDE_LOGGING_CONFIG=true
```

The runtime flag exposes the handler; the command-sync flag includes the managed guild command only in staging mode. Command sync requires the runtime flag, but config maintenance needs neither Gateway nor Message Content intent and does not activate any logging event.

Message update/delete logging:

```bash
MHCAT_FEATURE_LOGGING_MESSAGE_EVENTS_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true
MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true
```

Channel topic/permission logging:

```bash
MHCAT_FEATURE_LOGGING_CHANNEL_EVENTS_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
```

Voice join/leave logging:

```bash
MHCAT_FEATURE_LOGGING_VOICE_EVENTS_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_VOICE_STATE_INTENT=true
```

All four feature flags default to `false`. The three event gates can read an existing `loggings` row without enabling or syncing `/set-log-channel`; enabling one event family does not enable either of the others.

## Command Contract

The command definition preserves the legacy public surface:

- name `set-log-channel`, description `Set where log messages should send`, and all English/Traditional Chinese/Simplified Chinese localizations;
- one required channel option named `channel` / `頻道`, limited to text and news channels (Discord types `0` and `5`);
- no default member-permission restriction, so the command remains discoverable to ordinary guild members;
- legacy cooldown metadata `10`; neither implementation applies a feature-local cooldown in this path;
- a public defer followed by an edit of the original response.

Manage Messages (`8192`) is checked at runtime. Denial renders the legacy title ``<a:Discord_AnimatedNo:1015989839809757295> | 你需要有`訊息管理`才能使用此指令`` with discord.js `Red` (`#ED4245`).

An authorized invocation renders the exact yellow (`#FFDC35`) `<:logfile:985948561625710663> 日誌系統` embed, prompt text, footer text, and select placeholder. The bot footer avatar follows discord.js defaults: WebP for a static custom avatar and GIF for an animated avatar. The select accepts one through four values:

| Label | Description | Stored value |
| --- | --- | --- |
| `訊息更新` | `當訊息編輯時發送日誌` | `訊息更新` |
| `訊息刪除` | `當訊息刪除時發送日誌` | `訊息刪除` |
| `頻道更新` | `當頻道更新時發送日誌` | `頻道更新` |
| `用戶語音狀態更新` | `當用戶離開或加入或是靜音之類的時發送這個通知` | `用戶語音更新` |

The generated component ID carries the selected channel, invoking user, and one absolute ten-minute deadline. Only that user can save. A selection received at or after the deadline is rejected, even if the message was updated by an earlier selection. Successful selections replace all four toggle values, display the selected values as `` `value`,`value` ``, retain the menu, and leave every option visually non-default like legacy. Cross-user and expired interactions receive safe ephemeral red errors.

The old static `loggin_create` ID cannot recover the channel held by the Node collector closure. Go recognizes an orphaned legacy component and updates it with a safe instruction to rerun `/set-log-channel` instead of guessing a target channel.

## Mongo Compatibility

`loggings` retains these fields:

- `guild`
- `channel_id`
- `message_update`
- `message_delete`
- `channel_update`
- `member_voice_update`

Reads use a dedicated compatibility document. `channel_id` accepts Mongoose String-compatible scalar values, including strings and numeric/Boolean/ObjectID scalars. Toggle fields accept Mongoose Boolean-compatible scalars: native Booleans, recognized `true`/`1`/`yes` and `false`/`0`/`no` strings, and numeric `1`; unsupported values decode false. This lets event runtimes consume mixed legacy rows without a migration rewrite.

New config writes remain typed: `channel_id` is a BSON string and all toggles are BSON Booleans. Save updates every existing `{guild}` duplicate with `$set`; only a guild with no match receives an upserted row. Event reads use a 30-second positive/negative in-process cache with immediate local refresh after writes; cross-process config changes can take up to one cache window to appear. No startup index is created. The duplicate-safe non-unique `loggings_guild_lookup` index may be explicitly applied before the unique `{guild:1}` candidate passes duplicate audit, and must be removed before promoting that same-key unique index.

## Active Event Contract

Every event family reads `loggings` by guild, requires its own toggle, and sends to `channel_id`. A missing row, false toggle, missing target ID, bot-authored message, unchanged payload, or unavailable required cached snapshot is silent as applicable. All emitted messages disable user, role, and everyone mention parsing.

User and audit-actor custom avatars are forced to PNG. A missing custom/default user avatar uses `https://i.imgur.com/B91C90T.png`. Bot footer avatars use discord.js default WebP/GIF behavior.

### Message Update

- Requires the cached pre-edit message and uses its author, username, avatar, and bot flag as authoritative.
- Ignores bot authors and content that did not change.
- Emits the legacy `訊息編輯` embed with color `#46A3FF` and old message, new message, and new attachment fields in the original order.
- Preserves the legacy code-block spacing exactly as ```` ```<content> ``` ```` and renders empty content as ```` ``` ``` ````.
- Preserves raw non-empty attachment URL values and order, joined with commas; no attachments render `**沒有附件**`.

### Message Delete

- Requires cached deleted-message author/content data and ignores bot authors.
- Fetches message-delete audit action `72` with limit `1`.
- Uses the audit executor only when that entry's target exactly equals the deleted message author and its channel exactly equals the source channel. Otherwise the displayed deleter falls back to the message author. The stale legacy comment about a 20-second age check has no corresponding check.
- Emits the legacy `訊息刪除` embed with color `#84C1FF`, exact content code-block spacing, and raw ordered attachment text.

### Channel Update

- Requires the cached old channel snapshot and the `channel_update` toggle.
- A topic change takes precedence over permission changes in the same event. A Discord API null topic renders the literal legacy text `null`; an actual string `"null"` remains visually identical.
- Both topic and permission paths fetch channel-update audit action `11` with limit `1` and use the first returned executor without target/channel filtering, matching `entries.first()`.
- Topic changes emit the legacy `頻道主題更新` embed with color `#FF8040`.
- Permission changes compare current and old overwrites by ID, emit one `頻道權限更新` embed (`#FF5809`) for each changed/new overwrite still present in the new snapshot, and preserve the legacy 41-label order from `Create Instant Invite` through `Moderate Members`.
- Permission field lines remain default transitions first, then allows, then denies. Role/user mention syntax is selected from Discord's overwrite type.

### Voice State

- A join is only empty old channel to non-empty new channel and uses the new-state member/channel.
- A leave is only non-empty old channel to empty new channel and uses the old-state member/channel.
- Human and bot members can produce join/leave logs, matching the active legacy handler.
- Join and leave emit the exact `使用者加入語音頻道` (`#F235FA`) and `使用者退出語音頻道` (`#FA359A`) embeds with cached channel names.
- Direct channel-to-channel moves and mute/deafen-only changes emit nothing.

No boost, role, nickname, member join/remove, invite, reaction, emoji, sticker, thread, stage, guild, or other logging path after the line-364 comment boundary is active or implemented by this slice.

## Intentional Safety and Reliability Differences

- Go updates duplicate config rows together instead of racing an unawaited legacy delete against a replacement insert.
- Config writes and interaction responses are awaited.
- Versioned component IDs replace process-local collector state; stale, malformed, cross-user, and orphaned interactions return controlled errors.
- Missing audit entries, actors, or audit API responses degrade safely instead of dereferencing null.
- Permission overwrites are matched by ID instead of using the legacy mismatched-array index bug.
- Role/user formatting uses overwrite type instead of a potentially stale role-cache lookup.
- Emitted log embeds suppress mention parsing even though their visible mention syntax is preserved.
- Go can send through Discord REST to a configured channel absent from local cache; legacy required `client.channels.cache` to contain it.

The event paths still depend on discordgo cached old message/channel/member snapshots. Events that arrive without those snapshots cannot reproduce legacy content and are intentionally silent. No event queue or feature-specific rate-limit policy is added beyond the Discord client.

## Runtime Ownership

Do not run the Node process that loads `events/LoggingSystem.js` alongside any Go logging event gate for the same bot/guilds. The runtimes have no shared lease and will emit duplicate logs with competing audit attribution. Stop Node event ownership before enabling Go message, channel, or voice logging; disable all Go logging event flags before restoring Node.

## Parity Tests

Focused contracts cover the public definition, runtime permission, exact prompt/error/select payloads, owner and exact deadline checks, selected-state rendering, footer/avatar formats, scalar BSON reads, duplicate-safe writes, cached event conversion, exact embeds/colors/field spacing, raw attachments, audit rules, topic nulls, the 41-entry permission order, old/new voice member selection, ignored moves, gates, app wiring, command sync, and preflight. Run:

```bash
go test ./internal/core/domain ./internal/core/services/moderation ./internal/discord/features/logging ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/adapters/discordgo ./internal/discord/events ./internal/testutil/fakemongo ./internal/app ./internal/parity ./internal/config ./cmd/mhcat-command-sync ./cmd/mhcat-staging-preflight
```

## Staging Smoke

1. Use an isolated guild, database, text/news log channel, disposable source channels, voice channel, manager, ordinary user, and bot test member.
2. Stop every Node process that loads `events/LoggingSystem.js` before enabling a Go logging event gate.
3. Enable config/runtime sync flags, run staging preflight and command-sync dry-run, then apply only the reviewed managed `set-log-channel` change.
4. Confirm an ordinary member can discover the command but receives the exact runtime Manage Messages denial.
5. As a manager, verify the exact yellow prompt, footer, four options, and one-to-four selection bounds. Confirm another user cannot operate it and that the original user is rejected at/after ten minutes.
6. Select multiple values twice. Confirm the embed text changes, menu options remain non-default, and typed `channel_id`/Boolean values replace all duplicate guild rows without an index appearing.
7. On a disposable copied row, exercise numeric `channel_id` and string/numeric toggle values to confirm Mongoose-compatible reads, then restore typed values with the command.
8. Enable only message events. Edit and delete cached human messages, checking exact colors, field order/spacing, raw attachment order, old-author authority, PNG/default avatars, WebP/GIF footer, and no mention ping. Verify bots and uncached edits are silent.
9. Delete once with an exact action-72 audit target/channel match and once with either field mismatched; confirm executor attribution only for the exact match and author fallback otherwise.
10. Enable only channel events. Change a topic to/from null, then edit role and user overwrites. Confirm literal `null`, first action-11 audit-entry attribution without target filtering, legacy permission line order, correct role/user syntax, and no mention ping.
11. Enable only voice events. Join and leave as a human and bot, checking new-state join and old-state leave identity/channel data. Verify a direct move and mute/deafen-only update emit nothing.
12. Confirm each event gate emits only its own family and that config/select/event activity does not change slash usage beyond the one initial command attempt.

## Rollback

1. Disable command-sync inclusion and remove the managed staging command through a reviewed sync plan.
2. Disable all three Go logging event flags and stop Go event ownership before restarting the Node event runtime.
3. Disable `MHCAT_FEATURE_LOGGING_CONFIG_ENABLED` if config maintenance is also returning to Node.
4. Preserve `loggings`; typed Go writes are directly readable by the legacy Mongoose model. Do not add/drop indexes or delete duplicate production rows during rollback.
5. Restore Node only after confirming no Go process can receive the corresponding logging events.

Slash usage belongs to the assembled runtime's global interaction middleware. It increments `all_use_counts` exactly once per `/set-log-channel` attempt only when `MHCAT_FEATURE_USAGE_TRACKING_ENABLED=true`, including permission-denied attempts. Select components and all logging gateway events do not write usage counters.
