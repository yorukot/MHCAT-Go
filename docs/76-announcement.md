# Announcement Parity Contract

Status: parity-audited against the legacy announcement command, generic modal branch, bound-message relay, Mongoose schemas, pinned `validate-color` package, and discord.js color resolution. Runtime and command sync remain disabled by default. Live staging smoke is still required before production ownership.

## Scope

This contract covers:

- `/公告頻道設置 一次性公告頻道`;
- `/公告頻道設置 綁定公告頻道`;
- `/公告頻道設置 綁定公告頻道刪除`;
- `/公告發送`, its modal, preview, confirmation, cancellation, and final send;
- bound-channel `messageCreate` relay behavior from `events/ann_message.js`;
- `guilds.announcement_id` and `ann_all_sets` compatibility;
- command/component ownership, usage accounting, staging, and rollback.

Legacy sources:

- `slashCommands/公告系統/announcement_set_channel.js`
- `slashCommands/公告系統/announcement.js`
- `events/modal.js`
- `events/ann_message.js`
- `events/SlashCommands.js`
- `models/guild.js`
- `models/ann_all_set.js`
- `config.json`

## Gates And Ownership

Config command routes require:

```bash
MHCAT_FEATURE_ANNOUNCEMENT_CONFIG_ENABLED=true
MHCAT_STAGING_MODE=true
MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_CONFIG=true
```

One-time send routes require:

```bash
MHCAT_FEATURE_ANNOUNCEMENT_SEND_ENABLED=true
MHCAT_STAGING_MODE=true
MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_SEND=true
```

Bound relay events require:

```bash
MHCAT_FEATURE_ANNOUNCEMENT_RELAY_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true
MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true
```

The three gates are independent but share the same Mongo fields. Config ownership covers `/公告頻道設置`; send ownership covers `/公告發送`, announcement modal submits, and its transient confirmation components; relay ownership covers bound-channel message events and user-message deletion.

Do not leave the equivalent Node command, `events/modal.js` announcement branch, or `events/ann_message.js` active for the same bot/guild while Go owns that route family. The legacy generic modal listener cannot be disabled by announcement custom ID alone without a Node-side gate, so rollout must use a reviewed Node branch gate or exclusive bot/guild ownership.

Command sync is guild-scoped and staging-only. Preflight and staging scripts reject config/send inclusion without the matching runtime flag and warn when a runtime is enabled without its managed command.

## Command And Usage Contract

Both command definitions preserve exact names, descriptions, option order, option descriptions, required flags, and channel types `0` and `5`. They also preserve the legacy documentation URLs, including the malformed config URL `https://docsmhcat.yorukot.meocs/ann_set`.

Neither command sets Discord default member permissions, so both remain publicly discoverable like legacy. Runtime checks require Manage Messages.

The config source omitted `UserPerms`, so its exact denial title remains:

```txt
<a:Discord_AnimatedNo:1015989839809757295> | 你需要有`undefined`才能使用此指令
```

The send source declares `UserPerms: '訊息管理'`, so its exact ephemeral denial title is:

```txt
<a:Discord_AnimatedNo:1015989839809757295> | 你需要有`訊息管理`才能使用此指令
```

Both use discord.js named `Red`, `0xED4245`.

Usage belongs only to the global slash middleware. It records one best-effort attempt before route lookup and permission checks when `MHCAT_FEATURE_USAGE_TRACKING_ENABLED=true`. Announcement handlers do not write usage directly; modal, confirm, cancel, and relay interactions add no usage event.

## Config Command UI

`/公告頻道設置` publicly defers before its runtime permission check, matching legacy.

One-time channel creation/update preserves:

- title `<:megaphone:985943890148327454> 公告系統`;
- create text `<:Channel:994524759289233438> **您的公告頻道成功__創建__!!**`;
- update text `<:Channel:994524759289233438> **您的公告頻道成功__更新__!!**`;
- second line `**您目前的公告頻道為**:<#channel-id>`;
- `client.color.greate`, `0x53FF53`.

Bound-channel creation/update preserves:

- title `<:megaphone:985943890148327454> 綁定型公告系統`;
- exact `創建`/`更新` text;
- legacy second-line wording `**新增綁定型公告頻道為**:<#channel-id>` for both branches;
- `0x53FF53`.

Bound deletion preserves the trash emoji, exact `刪除` text, channel mention, and green color. A missing row returns `你沒有對這個頻道設定過綁定型公告!` in the animated-no red error title.

The command preserves nonempty whitespace in `標註` and `標題`. Color is not trimmed into validity.

Bound setup uses the pinned `validate-color` 2.2.4 default validator and additionally accepts exact `Random`. It therefore accepts CSS names and syntaxes such as `AliceBlue`, `#fff`, and `rgb(0 0 0)`, even though discord.js cannot render every accepted value as an embed color later. Go preserves those config values; an unsupported relay color safely leaves the original message unchanged instead of reproducing the legacy thrown error.

## One-Time Send UI

The slash command shows a modal with:

- legacy title `公告系統`;
- fields `anntag`, `anncolor`, `anntitle`, and `anncontent` in that order;
- exact legacy labels and short/paragraph styles;
- all fields required.

New modals use a bounded `mhcat:v1:announcement:submit:*` ID. Legacy `nal` modal submissions still route by the documented first field `anntag`. Unmodified Node also routes a Go modal by its first field rather than modal ID, so an already-open Go modal remains submit-compatible during rollback.

The modal validates color first. Legacy modal validation accepted HTML color names/syntaxes before discord.js resolved the embed color. Go accepts the same successful intersection, including six-digit `#RRGGBB` and supported exact Discord color names. Values such as unprefixed `53FF53`, `Random`, `#fff`, lowercase `red`, and `AliceBlue` return the controlled legacy color error rather than hanging after a discord.js exception.

Raw nonempty whitespace in tag, title, and content is preserved. The preview keeps:

- raw tag content;
- raw title and body;
- selected color;
- footer `來自<user.tag>的公告`;
- invoking-user avatar.

The confirmation keeps:

- title `是否將此訊息送往公告?(請於六秒內點擊:P)`;
- color `#00ff19`;
- primary `✅` / `是` button;
- danger `❎` / `否` button;
- a six-second deadline and follow-up deletion at that deadline.

New buttons use owner/guild-scoped state IDs:

```txt
mhcat:v1:announcement:confirm:state=<id>
mhcat:v1:announcement:cancel:state=<id>
```

The state store is bounded to 512 drafts, expires at the exact six-second deadline, and does not let an unauthorized click consume the owner's state. Raw `announcement_yes` and `announcement_no` remain rejected as ambiguous because legacy work flows reused the same IDs and their collector closure does not survive process transfer.

Go presents the confirmation immediately after the preview edit instead of preserving the legacy 500-millisecond `setTimeout`. This removes a nonfunctional delay while keeping the same visible prompt and deadline.

Confirmation reads `guilds.announcement_id`, sends the final embed, and returns exact success content:

```txt
<a:green_tick:994529015652163614> | 成功發送!
```

A missing row, blank channel, or exact string `0` returns the legacy multiline setup instruction. Cancellation returns `已取消`.

## Bound Relay Contract

The relay ignores DMs, bot authors, unconfigured channels, and empty/whitespace-only content. For a configured text message it preserves:

- stored `tag` as visible message content;
- stored title;
- user message content as embed description;
- footer `來自<author.tag>的公告` and author avatar;
- exact stored supported color;
- random colors in the inclusive `0x000000` through `0xFFFFFF` range for exact `Random` or legacy `RANDOM`.

Lowercase or whitespace-padded random values are not normalized. Unsupported stored colors produce no replacement and do not delete the original.

The Go relay sends the replacement before deleting the original. If send fails, the original remains. If delete fails after a successful send, the replacement remains and the failure is logged by event dispatch.

Attachment-only and empty-content messages are left unchanged. Legacy did not forward attachments and could throw while building an empty embed; Go does not reproduce that unresolved failure.

## Mention Safety

Legacy preview, final send, and bound relay used raw `tag` content with default mention parsing. Go preserves the visible text but sends empty allowed mentions, so `@everyone`, role mentions, and user mentions do not ping.

This is a deliberate security boundary. Restoring real tag pings requires a separate ADR, an explicit operator-controlled allowlist, and staging tests. It must not be changed as a UI-only parity fix.

## Mongo Compatibility

Collections and fields remain exact:

- `guilds`: `guild`, `announcement_id`, shared `voice_detection`;
- `ann_all_sets`: `guild`, `announcement_id`, `tag`, `color`, `title`.

Writes remain typed BSON strings readable by Mongoose. Existing `guilds` rows are patch-updated so dashboard/shared fields remain intact. Go updates all duplicate matches before upserting and deletes all exact duplicate bound rows, instead of legacy first-row update/delete behavior.

Reads use separate permissive DTOs. Mongoose-compatible String coercion accepts string, Boolean, numeric, decimal, ObjectID, symbol, and JavaScript scalar forms supported by the shared decoder. Compound object/array values remain unusable. Read values are no longer trimmed or case-normalized.

The application creates no startup index. Candidate unique indexes on `guilds.guild` and `ann_all_sets.{guild,announcement_id}` must not be applied until duplicate keys, null/missing keys, scalar drift, shared dashboard writes, and malformed color/tag/title values are audited. No TTL index or draft collection is used.

## Intentional Safety And Reliability Differences

- Tag text is visible but never pings.
- Versioned state IDs replace collision-prone raw confirmation IDs.
- Confirmation is owner/guild scoped and survives unrelated button activity, but process restart invalidates its six-second in-memory draft.
- The prompt is immediate rather than delayed 500 milliseconds.
- Unsupported colors and stale/malformed components return controlled outcomes instead of unresolved defers or thrown errors.
- Final one-time send is awaited before success; legacy could report success before an unawaited channel send failed.
- Bound relay sends before deleting; legacy launched deletion first without awaiting either operation.
- Empty/attachment-only content is retained rather than entering the legacy invalid-empty-embed path.
- Config writes are awaited, patch existing rows, update duplicates together, and preserve unrelated `guilds` fields.

Exact command UI, raw text, successful color behavior, fixed embeds/messages, modal fields, footer text, confirmation controls, random range, config fields, and supported legacy modal migration are preserved.

## Runtime Ownership And Migration

Before enabling any announcement gate:

1. Stop or gate the corresponding Node command/modal/event route for the target bot and guild.
2. Audit `guilds` duplicates by `guild` and `ann_all_sets` duplicates by `{guild,announcement_id}`.
3. Audit missing/non-string keys, String scalar drift, null/compound tag/title/color fields, unsupported colors, and shared dashboard writers.
4. Preserve both collections and all fields. Do not backfill, deduplicate, or create indexes during the ownership switch.
5. Enable only the required config, send, or relay gates in an isolated staging guild.
6. For relay, enable Gateway, Guild Messages, and privileged Message Content intent only after privacy review.
7. Run staging preflight, command-sync dry-run, and the canonical smoke below before apply.

The config and send commands can move independently, but operators must account for their shared `guilds.announcement_id`. Relay can remain on Node while Go config writes because typed Go fields are Mongoose-readable, but concurrent config writers can still race and should not overlap.

## Parity Tests

Focused tests lock command metadata, permission presentation, fixed messages, colors, raw whitespace, both color validators, modal/preview/confirmation shape, six-second expiry, owner/guild state, webhook follow-up deletion, legacy/versioned routes, ambiguous raw IDs, Mongoose scalar reads, typed writes, relay order/safety, usage ownership, app wiring, and gate pairing. Run:

```bash
go test ./internal/core/domain ./internal/core/services/announcements ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/discord/customid ./internal/discord/responses ./internal/adapters/discordgo ./internal/discord/features/announcements ./internal/discord/events ./internal/discord/interactions ./internal/discord/runtime ./internal/app ./internal/config ./internal/parity ./cmd/mhcat-command-sync ./cmd/mhcat-staging-preflight
go test -race ./internal/core/domain ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/discord/responses ./internal/adapters/discordgo ./internal/discord/features/announcements ./internal/discord/runtime ./internal/app
go vet ./internal/core/domain ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/discord/responses ./internal/adapters/discordgo ./internal/discord/features/announcements ./internal/app
```

## Staging Smoke

1. Use an isolated guild/database with a manager, a non-manager, a second ordinary user, disposable text/announcement channels, and no active Node announcement owner.
2. Audit duplicate/type/color findings in `guilds` and `ann_all_sets`; confirm no index apply is planned.
3. Enable paired config/send flags as needed. Run preflight and command-sync dry-run before staging apply.
4. Verify both commands remain discoverable to a non-manager. Confirm config denial contains literal `undefined`, send denial contains `訊息管理`, and only send denial is ephemeral.
5. Exercise one-time create/update plus bound create/update/delete/missing branches and compare exact titles, descriptions, colors, channel mentions, and Mongo fields.
6. Test exact `Random`, `RANDOM`, lowercase `random`, `#RRGGBB`, unprefixed hex, short hex, CSS names/functions, and leading whitespace according to each flow's validator contract.
7. Submit `/公告發送`; verify exact modal fields, raw whitespace preservation, preview footer/avatar, no tag ping, confirmation controls, and deletion at six seconds.
8. Have a second user click confirm/cancel and verify denial does not consume owner state. Confirm/cancel as owner and verify exact responses.
9. Verify missing/`0` one-time channel text, a Discord send failure with no false success, and successful final output with no tag ping.
10. Seed numeric/Boolean scalar values in disposable copied fields and verify Mongoose-compatible reads; restore typed values afterward. Confirm compound values remain unusable.
11. If relay is enabled, verify Gateway/Guild Messages/Message Content gates, text relay UI, modern/legacy author tags, exact random range behavior, no ping, and send-before-delete failure ordering.
12. Verify empty/attachment-only messages remain untouched, bot/DM/unconfigured messages are ignored, and Node plus Go never process the same message.
13. With usage tracking enabled separately, verify one increment per slash attempt and none for modal, confirm, cancel, or relay events.
14. Disable gates, remove managed staging commands, preserve data, and execute rollback checks.

## Rollback

1. Disable `MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_CONFIG` and `MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_SEND`; remove only those managed staging commands through command sync.
2. Disable the matching Go runtime gates. Keep Node handlers stopped until no Go process can route the same announcement interaction/event family.
3. Preserve `guilds` and `ann_all_sets`; typed Go writes remain Mongoose-readable. Do not mutate or drop indexes during emergency rollback.
4. Wait at least six seconds for Go confirmation prompts to expire, or accept that unmodified Node cannot route their versioned confirm/cancel IDs. Already-open Go modals remain Node-compatible because legacy routes by first field `anntag`.
5. Confirm shared dashboard writers still read/write `guilds` without losing `voice_detection` or unrelated fields.
6. Restore only the intended Node command/modal/relay branches and verify a config read, one-time send, and bound relay in staging.
7. Re-enable production ownership only after confirming no Go process can receive the same routes or message events.
