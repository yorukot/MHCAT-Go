# Discord Intents Plan

Status: Phase 1.5 Gate B input. Legacy evidence: `MHCAT/index.js` enables Guilds, GuildMembers, GuildMessages, GuildMessageReactions, GuildVoiceStates, and MessageContent.

## Intent Matrix

| Intent | Legacy feature requiring it | Can remove? | Replacement | Risk if removed | Go default |
| --- | --- | --- | --- | --- | --- |
| `Guilds` | Slash commands, guild/channel/role metadata, Ready/GuildCreate, most interaction features | No | None | Bot cannot operate core guild interactions | Enabled in dev and prod |
| `GuildMembers` | Welcome/leave messages, join roles, account-age checks, member remove logging, some role/member fetch paths; slash `/驗證` uses interaction member data plus REST role/nickname side effects and does not require member-event handling by itself | Partially | Use interaction member data and REST fetch-on-demand for command paths; keep intent for join/leave/account-age event features | Join/leave automation, account-age policy, and member remove logging stop working | Disabled by default; enable with `ENABLE_GUILD_MEMBERS_INTENT=true` for parity event features |
| `GuildMessages` | Text XP, chatbot, anti-scam, announcement message listeners, XP reset `^確認^` confirmation, legacy prefix/restart message, message logging | Partially | Convert operator commands to slash; use interactions/components/modals for new flows; keep only for message-event features | Text XP, chatbot, anti-scam, XP reset confirmation, and message logging stop working | Disabled by default; enable when message-event features are enabled |
| `GuildMessageReactions` | Reaction roles and reaction logging | Partially | Prefer buttons/selects for new role assignment; keep for legacy reaction-role parity | Reaction role add/remove and reaction logs stop working | Disabled by default; enable with reaction-role/logging feature flags |
| `GuildVoiceStates` | Voice XP, dynamic voice rooms, voice lock, voice join/leave state | No for those features | None; REST cannot replace voice state events | Voice XP and dynamic voice channels stop working | Disabled by default; enable with `ENABLE_VOICE_STATE_INTENT=true` for voice features |
| `MessageContent` | Legacy restart text, prefix handler, text XP content checks, chatbot prompts, anti-scam URL scanning, announcement/chat listeners, XP reset `^確認^` confirmation, some message logging | Yes for restart/prefix; no for content-driven features unless redesigned | Replace restart with owner-only slash command or out-of-band process restart; convert admin flows to slash/modal/component; content features require explicit opt-in or redesign | Content-based XP/chatbot/anti-scam/logging and XP reset confirmation stop working or lose detail | Disabled in dev and prod by default; enable only with `ENABLE_MESSAGE_CONTENT_INTENT=true` and documented feature flags |

## Feature Decisions

### Message Content

Message Content is not enabled by default in Go.

Features that truly require Message Content if preserved as-is:

- Text XP from arbitrary messages.
- Chatbot/autochat prompt capture.
- Anti-scam URL scanning.
- Message create/update/delete logging with content.
- Legacy announcement/chat message handlers. The Go bound announcement relay is implemented only when `MHCAT_FEATURE_ANNOUNCEMENT_RELAY_ENABLED=true` and requires `MHCAT_DISCORD_ENABLE_GATEWAY=true`, `MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true`, and `MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true`.
- XP reset full-server confirmation. The Go `/經驗值重製` slice is implemented only when `MHCAT_FEATURE_XP_RESET_ENABLED=true` and requires `MHCAT_DISCORD_ENABLE_GATEWAY=true`, `MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true`, and `MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true`.

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
