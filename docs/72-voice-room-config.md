# Voice Room Config, Runtime, and Locking

Status: config commands, dynamic room lifecycle, and password locking are parity-audited behind disabled-by-default gates.

## Legacy References

- Config command: `MHCAT/slashCommands/語音包廂/voice_channel.js`
- Config delete command: `MHCAT/slashCommands/語音包廂/voice_channel_delete.js`
- Password command: `MHCAT/slashCommands/語音包廂/lock_channel.js`
- Voice-state runtime: `MHCAT/events/voice_create.js`
- Password modal: `MHCAT/events/modal.js`
- Models: `MHCAT/models/voice_channel.js`, `voice_channel_id.js`, and `lock_channel.js`

## Gates

Config runtime and command sync:

```bash
MHCAT_FEATURE_VOICE_ROOM_CONFIG_ENABLED=true
MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_CONFIG=true
```

Those flags expose `/語音包廂設置` and `/語音包廂刪除`. The command definitions remain discoverable to guild members because legacy did not register effective default member permissions. Both handlers still require Manage Messages at runtime, including the Discord Administrator override.

Dynamic create/move/cleanup additionally needs events to reach the process:

```bash
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_VOICE_STATE_INTENT=true
```

The config flags can run without those gateway flags for command-only maintenance. No dynamic room side effect occurs until Gateway and Voice State events are active.

Password command, prompt, and modal runtime:

```bash
MHCAT_FEATURE_VOICE_ROOM_LOCK_ENABLED=true
MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_LOCK=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_VOICE_STATE_INTENT=true
```

Config and lock gates are independent. Lockable room creation needs the config gate to seed `lock_channels`; the lock gate can operate on existing rows without exposing the config commands. Config validation and staging scripts require Gateway and Voice State intent whenever the lock runtime is enabled.

## Command Contract

`/語音包廂設置` preserves:

- required voice/stage channel `語音頻道`;
- required raw string `設定頻道名稱`;
- required Boolean `是否予許房主上鎖`, including the legacy spelling;
- optional integer `設定人數上限`;
- public defer/edit flow and exact green success/error embeds;
- runtime Manage Messages denial text;
- `0` for omitted or explicitly supplied unlimited user count;
- exact room-name whitespace and only the first JavaScript-style `{name}` replacement.

Negative limits and values above `99` return `必須為1-99的整數!`. The visible text says `1-99`, but explicit `0` remains accepted because legacy treated it like omission.

`/語音包廂刪除` preserves the selected-channel type branch. A type-2 voice channel deletes by `{guild,ticket_channel}`. Every other channel type deletes by `{guild,parent}`. That means a stage channel follows the category branch and normally returns `你沒有對這個類別沒有設定喔!`; this legacy bug is intentionally covered by parity tests. Slash deletion removes config rows only and does not delete already active dynamic rooms.

`/上鎖頻道` remains publicly discoverable, defers ephemerally, requires the caller to own an existing lock row for their current voice channel, and replaces that row. Password input is not trimmed. Leading/trailing whitespace is preserved in Mongo, the success embed, and exact modal comparison. Omitting the option writes BSON `null` and displays `null`; an explicitly stored empty BSON string remains distinguishable from null for legacy reads.

The password prompt and modal preserve the legacy labels, emojis, text, channel link, fixed prompt/success/error colors (`#53FF53` and `#EA0000`), and per-message random DM colors. A deleted room returns the legacy plain red title `很抱歉，該包廂可能已被刪除!`.

## Dynamic Runtime

On a voice channel change, the config runtime:

- reads `voice_channels` by `{guild,ticket_channel}`;
- creates a type-2 voice room for human or bot members, matching legacy;
- resolves the trigger channel's current category rather than relying on its setup-time parent;
- preserves the raw template and replaces only its first `{name}` token with the username;
- copies current category permission overwrites;
- grants the joining owner Manage Channels and Manage Roles in the initial create request;
- persists `{guild,channel_id}` in `voice_channel_ids`;
- seeds a nullable `lock_channels` row and sends the random-color owner DM when the config is lockable;
- moves the member only after state writes succeed;
- deletes the lock row, Discord channel, and state rows after a tracked room becomes empty.

Bots still create rooms from configured triggers, but password enforcement skips bot joins like legacy.

For a passworded existing `lock_channels` row, an unauthorized human join sends the mention/prompt to `text_channel`, disconnects the member, and sends the random-color DM. The generated prompt ID carries the channel, target user, and exact 60-second deadline within Discord's 100-character limit. It opens the legacy `<channel>anser` modal only for that user before the deadline. Correct input adds the user to `ok_people`; subsequent joins are allowed.

## Mongo Compatibility

Collection `voice_channels` retains:

- `guild`
- `ticket_channel`
- `limit`
- `name`
- nullable `parent`
- `lock`

Reads apply Mongoose-compatible scalar coercion to `limit`, `name`, `parent`, and `lock`. Writes remain typed, preserve raw `name`, write unlimited as numeric `0`, and write a missing parent as null. Save updates every duplicate trigger row and upserts only when none match. Trigger deletion and category deletion remove all matching duplicate rows.

Collection `voice_channel_ids` retains `guild` and `channel_id`. State creation uses `$setOnInsert`; cleanup removes duplicate matching rows. No startup reconciliation or index creation occurs.

Collection `lock_channels` retains:

- `guild`
- `channel_id`
- nullable `lock_anser`
- `owner`
- nullable `text_channel`
- `ok_people`

Reads preserve Mongoose String coercion for password, owner, and text-channel scalars. Mixed `ok_people` arrays remain readable; only exact string entries can authorize a Discord user, matching legacy strict equality. Scalar string array values are read as one-element arrays. New rows always encode empty `ok_people` as BSON `[]`, not null, so the first `$addToSet` succeeds. Correct answers use `$addToSet` to avoid duplicate authorization entries.

All candidate indexes remain duplicate-audit gated. The app creates no index at startup.

## Intentional Safety Differences

- Parentless triggers create an unparented room; legacy dereferenced a null parent and failed.
- Owner permissions are included atomically during channel creation instead of being applied by a delayed timer.
- Config writes update duplicate rows instead of racing an unawaited delete against a replacement insert.
- Dynamic state and lock writes are awaited before moving the member; failed tracking or move operations clean up created state/channel data best-effort.
- Empty-room cleanup waits for lock, channel, and state deletion in a deterministic order.
- Prompt IDs replace process-local `lock_start` collector state with channel/user/deadline data. Old orphaned `lock_start` buttons return a safe retry error.
- Prompt delivery and disconnect are awaited; DM delivery is best-effort.
- Missing rows and repository failures return safe responses instead of continuing into legacy null dereferences.
- Correct password authorization is idempotent through `$addToSet` instead of appending duplicate IDs.
- User/role/everyone mentions are suppressed except for the intended locked-user prompt mention.

The stage-channel delete branch and bot trigger behavior are not safety fixes; both intentionally preserve legacy behavior.

## Runtime Ownership

Do not run the Node process that loads `events/voice_create.js` alongside the Go voice-room event runtime. The two processes cannot coordinate and can create duplicate rooms, race member moves, send duplicate lock prompts, or delete shared state. Stop Node ownership before enabling Go Gateway handling, and disable Go before rollback to Node.

## Parity Contracts

Focused tests lock public command registration, runtime permissions, option definitions, exact response payloads/colors, raw names/passwords, zero limits, stage deletion, bot creation, current-category resolution, parentless safety, permission overwrites, state cleanup, 60-second user-scoped prompts, Mongo scalar/null/array compatibility, and collection/filter names. Run:

```bash
go test ./internal/core/domain ./internal/core/services/voice ./internal/discord/features/voice ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/adapters/discordgo ./internal/testutil/fakemongo ./internal/app ./internal/parity ./internal/config
```

## Staging Smoke

1. Use an isolated guild, database, category, trigger channel, text channel, and two test users.
2. Stop every Node process that loads `events/voice_create.js`.
3. Enable config/runtime sync flags and run `mhcat-staging-preflight` plus command-sync dry-run.
4. Verify both config commands are discoverable to a non-manager, then verify runtime permission denial.
5. As a manager, set a trigger with name ` {name}-{name} `, explicit limit `0`, and locking enabled. Confirm the raw name, `limit:0`, current parent, and Boolean lock value in `voice_channels`.
6. Apply command sync, enable Gateway and Voice State intent, join the trigger, and verify one room is created under the trigger's current category with only the first token replaced.
7. Verify the owner has Manage Channels/Roles, the member is moved, `voice_channel_ids` contains the room, `lock_channels.lock_anser` and `text_channel` are null, and `ok_people` is an empty BSON array.
8. Run `/上鎖頻道 密碼: secret ` with intentional surrounding spaces and verify the stored and displayed value is unchanged.
9. Join with the second user. Verify one prompt mention, disconnect, DM, and no password modal for another user or after exactly 60 seconds.
10. Enter the exact password, verify the legacy success/link response and one `ok_people` entry, then rejoin successfully.
11. Leave the room empty and verify the Discord room, lock row, and tracked state are removed.
12. Move the trigger to another category and repeat once; verify the new room uses the current category and its overwrites.
13. Delete by the trigger voice channel and by a disposable category, checking the exact legacy success/missing responses.

## Rollback

1. Disable command-sync include flags and remove the managed staging commands with the reviewed sync plan.
2. Disable `MHCAT_FEATURE_VOICE_ROOM_LOCK_ENABLED` and `MHCAT_FEATURE_VOICE_ROOM_CONFIG_ENABLED`.
3. Stop Go Gateway ownership and confirm no process can create, move, prompt, or clean up voice rooms.
4. Preserve `voice_channels`, `voice_channel_ids`, and `lock_channels`; Node can read Go's legacy-compatible scalar/null/array layout.
5. Restore Node only after Go is stopped. Audit stale active room/state rows manually rather than creating indexes or deleting production data during rollback.

Slash-command usage belongs to the global interaction middleware in the assembled runtime. It increments `all_use_counts` exactly once per `/語音包廂設置`, `/語音包廂刪除`, or `/上鎖頻道` attempt only when `MHCAT_FEATURE_USAGE_TRACKING_ENABLED=true`; voice-state events, prompt buttons, and modal submissions do not add usage writes.
