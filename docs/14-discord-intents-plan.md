# Discord Intents Plan

Status: Phase 1.5 Gate B input. Legacy evidence: `MHCAT/index.js` enables Guilds, GuildMembers, GuildMessages, GuildMessageReactions, GuildVoiceStates, and MessageContent.

## Intent Matrix

| Intent | Legacy feature requiring it | Can remove? | Replacement | Risk if removed | Go default |
| --- | --- | --- | --- | --- | --- |
| `Guilds` | Slash commands, guild/channel/role metadata, Ready/GuildCreate, most interaction features | No | None | Bot cannot operate core guild interactions | Enabled in dev and prod |
| `GuildMembers` | Welcome/leave messages, join roles, account-age checks, member remove logging, poll non-bot totals/export names, and some role/member fetch paths; slash `/驗證` uses interaction member data plus REST role/nickname side effects and does not require member-event handling by itself | Partially | Use interaction member data and REST fetch-on-demand for command paths; keep intent for join/leave/account-age event features and exact poll member parity | Join/leave automation, account-age policy, and member remove logging stop; polls fall back to zero/current voter totals and missing-member export labels | Disabled by default; enable with `MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true` for parity features |
| `GuildMessages` | Text XP, chatbot, anti-scam, announcement message listeners, XP reset `^確認^` confirmation, legacy prefix/restart message, message logging | Partially | Convert operator commands to slash; use interactions/components/modals for new flows; keep only for message-event features | Text XP, chatbot, anti-scam, XP reset confirmation, and message logging stop working | Disabled by default; enable when message-event features are enabled |
| `GuildMessageReactions` | Legacy reaction roles | Partially | Prefer buttons/selects for new features; keep the intent only for the parity-audited role-selection ownership slice | Reaction role add/remove stops; setup/button commands are intentionally not split onto a separate gate | Disabled by default; `MHCAT_FEATURE_ROLE_SELECTION_ENABLED=true` requires `MHCAT_DISCORD_ENABLE_GATEWAY=true` and `MHCAT_DISCORD_GUILD_MESSAGE_REACTIONS_INTENT=true` |
| `GuildVoiceStates` | Voice XP, dynamic voice rooms, voice lock, voice join/leave state | No for those features | None; REST cannot replace voice state events | Voice XP and dynamic voice channels stop working | Disabled by default; enable with `ENABLE_VOICE_STATE_INTENT=true` for voice features |
| `MessageContent` | Legacy restart text, prefix handler, text XP content checks, chatbot prompts, anti-scam URL scanning, announcement/chat listeners, XP reset `^確認^` confirmation, some message logging | Yes for restart/prefix; no for content-driven features unless redesigned | Replace restart with owner-only slash command or out-of-band process restart; convert admin flows to slash/modal/component; content features require explicit opt-in or redesign | Content-based XP/chatbot/anti-scam/logging and XP reset confirmation stop working or lose detail | Disabled in dev and prod by default; enable only with `ENABLE_MESSAGE_CONTENT_INTENT=true` and documented feature flags |

## Feature Decisions

### Message Content

Message Content is not enabled by default in Go.

Features that truly require Message Content if preserved as-is:

- Text XP from arbitrary messages.
- Chatbot/autochat prompt capture.
- Anti-scam URL scanning/deletion requires the explicit event gate, Gateway, Guild Messages, Message Content, and exclusive Node/Go ownership; see the [anti-scam parity contract](77-anti-scam.md).
- Message create/update/delete logging with content.
- Legacy announcement/chat message handlers. The Go bound announcement relay requires `MHCAT_FEATURE_ANNOUNCEMENT_RELAY_ENABLED=true`, Gateway, Guild Messages, Message Content, and exclusive Node/Go event ownership; see the [announcement parity contract](76-announcement.md).
- XP reset full-server confirmation. The Go `/經驗值重製` slice is implemented only when `MHCAT_FEATURE_XP_RESET_ENABLED=true` and requires `MHCAT_DISCORD_ENABLE_GATEWAY=true`, `MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true`, and `MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true`.

Logging config itself requires no gateway intent. Its message, channel, and voice event families are independently gated; message logging needs Guild Messages and Message Content, channel logging needs Gateway channel updates, and voice logging needs Voice State intent. Do not overlap any enabled Go event family with Node `events/LoggingSystem.js`; see the [logging parity contract](48-logging-config.md).

Features that should not keep Message Content:

- Restart-by-message: replace with owner-only slash command or out-of-band deployment restart.
- Prefix commands: do not implement unless a documented legacy command remains impossible as slash/context/modal/component.
- Admin setup flows: use slash commands, modals, selects, and buttons.

If Message Content is disabled:

- The Go bot must not register or start content-dependent handlers.
- It should log a redacted startup summary listing disabled features.
- User-facing command/help output must not claim disabled features are active.

### Prefix Commands

The legacy `MessageCreate.js` prefix handler exists, but no active `commands/` directory was found in this checkout. Go should not implement prefix commands by default.

Decision:

- No prefix command router in Wave 1.
- No restart-by-message.
- If future research proves a live prefix command is required, it must go through ADR, parity tests, and explicit Message Content config.

### Guild Member Cache

Legacy sets `GuildMemberManager.maxSize: Infinity`, which is not acceptable as a default in Go.

Decision:

- Do not require an unbounded member cache.
- Use event payloads for join/leave where available.
- Use REST fetch-on-demand with context timeouts and rate-limit handling for command paths.
- Add a bounded cache only where measurements justify it.
- Poll creation/rerenders use guild-member listing for exact non-bot participation totals, and result/export paths use bulk member lookup for exact tags. With no usable member access, the runtime remains responsive but degrades to zero/current unique-voter totals and `使用者已退出伺服器!` labels.

### Dev and Production Defaults

Wave 1 skeleton defaults:

```txt
Guilds: enabled
GuildMembers: disabled
GuildMessages: disabled
GuildMessageReactions: disabled
GuildVoiceStates: disabled
MessageContent: disabled
```

Parity profile for staging/canary, only when the matching features are implemented:

```txt
Guilds: enabled
GuildMembers: enabled if welcome/verification/member logging is enabled
GuildMessages: enabled if text XP/chatbot/anti-scam/message logging is enabled
GuildMessageReactions: enabled if reaction roles/logging are enabled
GuildVoiceStates: enabled if voice XP/dynamic voice rooms are enabled
MessageContent: enabled only if content-dependent features are explicitly enabled
```

Current approved Message Content exception:

```txt
Announcement bound relay: MHCAT_FEATURE_ANNOUNCEMENT_RELAY_ENABLED=true
Required flags: MHCAT_DISCORD_ENABLE_GATEWAY=true, MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true, MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true
Default: disabled
Ownership: stop/gate Node events/ann_message.js for the same bot/guild; follow docs/76-announcement.md

Anti-scam deletion: MHCAT_FEATURE_ANTI_SCAM_MESSAGE_DELETE_ENABLED=true
Required flags: MHCAT_DISCORD_ENABLE_GATEWAY=true, MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true, MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true
Default: disabled
Ownership: stop/gate Node events/safe_server.js for the same bot/guild; follow docs/77-anti-scam.md

XP reset confirmation: MHCAT_FEATURE_XP_RESET_ENABLED=true
Required flags: MHCAT_DISCORD_ENABLE_GATEWAY=true, MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true, MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true
Default: disabled
```

Current approved Guild Members exception:

```txt
Join-role assignment: MHCAT_FEATURE_JOIN_ROLE_ASSIGNMENT_ENABLED=true
Leave-message delivery: MHCAT_FEATURE_LEAVE_MESSAGE_DELIVERY_ENABLED=true
Required additional flag: MHCAT_DISCORD_ENABLE_GATEWAY=true
Default: disabled
```

Current verification command flow:

```txt
/驗證 slash flow: MHCAT_FEATURE_VERIFICATION_FLOW_ENABLED=true
Intent dependency: no Guild Members gateway intent by itself; uses interaction member data and Discord role/nickname REST side effects.
Account-age kick: `MHCAT_FEATURE_ACCOUNT_AGE_POLICY_ENABLED=true` is a separate member-join policy and requires `MHCAT_DISCORD_ENABLE_GATEWAY=true` plus `MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true`. It remains disabled by default and must be staged in an isolated guild because it can DM/kick members and stop later member-add handlers.
Default: disabled
```

Current ticket interaction flow:

```txt
Ticket runtime: MHCAT_FEATURE_TICKETS_ENABLED=true
Privileged intents: none; slash/modal/button payloads plus Discord REST provide the required data
Ownership: setup/delete commands, ticket-shaped nal submits, tic, and del stay under this one gate
Default: disabled
```

See the [ticket parity contract](74-ticket.md) before staging or production ownership changes.

Current poll interaction flow:

```txt
Poll runtime: MHCAT_FEATURE_POLLS_ENABLED=true
Message Content intent: not required
Guild Members intent: required for exact non-bot totals and bulk export names; enable the application privileged intent and MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true
Ownership: /投票創建, all mhcat:v1:poll:* components, and legacy poll_<choice>/see_result/poll_menu/menu_choose routes stay under this one gate
Default: disabled
```

Do not overlap Node and Go poll owners. See the [poll parity contract](75-poll.md) before staging, migration, or rollback.

Current approved reaction-role exception:

```txt
Role-selection runtime: MHCAT_FEATURE_ROLE_SELECTION_ENABLED=true
Required flags: MHCAT_DISCORD_ENABLE_GATEWAY=true, MHCAT_DISCORD_GUILD_MESSAGE_REACTIONS_INTENT=true
Guild Members intent: not required; buttons use interaction member roles and REST writes
Known limitation: a reaction remove first observed after restart has best-effort bot identity when the member is absent from state
Ownership: setup commands, modal/buttons, and reaction events stay under this one gate
Default: disabled
```

See the [role-selection parity contract](73-role-selection.md) before staging or production ownership changes.

## Config Requirements

- `ENABLE_MESSAGE_CONTENT_INTENT`
- `ENABLE_GUILD_MEMBERS_INTENT`
- `ENABLE_GUILD_MESSAGES_INTENT`
- `ENABLE_GUILD_MESSAGE_REACTIONS_INTENT`
- `ENABLE_VOICE_STATE_INTENT`

Startup validation must reject impossible combinations, for example:

- `MHCAT_FEATURE_TEXT_XP_ACCRUAL_ENABLED=true` while `MHCAT_DISCORD_GUILD_MESSAGES_INTENT=false` or `MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=false`.
- `CHATBOT_ENABLED=true` while `ENABLE_MESSAGE_CONTENT_INTENT=false`.
- `MHCAT_FEATURE_VOICE_XP_SESSIONS_ENABLED=true` while `MHCAT_DISCORD_VOICE_STATE_INTENT=false`.

## Tests Required

- Intent builder table tests.
- Disabled-intent startup validation tests.
- Feature flag to intent dependency tests.
- Message Content disabled behavior tests.
- Owner-only restart command permission tests.
- Fetch-on-demand timeout and rate-limit tests.
- Poll member-list success/fallback and export-name tests.
