# Auto-Chat Config and Local Fallback

Status: config commands and the local fallback runtime are implemented behind separate explicit gates. The paid ChatGPT handoff remains disabled.

## Legacy References

- Set command: `MHCAT/slashCommands/實用工具/chat.js`
- Delete command: `MHCAT/slashCommands/實用工具/chat_delete.js`
- Runtime message handler and local matcher: `MHCAT/events/Chatbot.js`
- Local response corpus: `MHCAT/chat.json`
- External handoff models: `MHCAT/models/chatgpt.js`, `MHCAT/models/chatgpt_get.js`

## Implemented Surface

Config commands:

- `/自動聊天頻道`
- `/自動聊天頻道刪除`

The separately gated MessageCreate runtime restores the legacy local fallback path. It:

- ignores DMs and bot-authored messages
- reads the configured `chats.channel`
- reads `chatgpt_gets.price` without writing it
- uses the bundled legacy response corpus when the balance row is missing, negative, or not numeric
- preserves the legacy no-response state when the balance is zero
- leaves positive-balance messages for the still-disabled paid handoff path
- preserves the legacy `說出` echo behavior and fuzzy response matching
- sends a typing indicator and waits a legacy-randomized 1 to 5 seconds for corpus matches
- replies to the source message with all mentions suppressed

The legacy report-error button was constructed but never attached to a sent response, so there is no active report component to restore.

## Config UI Parity

The command paths preserve the legacy names/options, text/news channel restriction, Manage Messages check, public defer/edit flow, green `自動聊天系統` success embeds, and the missing-config text `你沒有設定過，我不知道要刪除甚麼!`.

## Mongo Compatibility

Collections read by the local runtime:

- `chats`: `guild`, `channel`
- `chatgpt_gets`: `guild`, `price`

The runtime performs no Mongo writes. The config commands continue to update all duplicate `{guild}` `chats` rows before falling back to an upsert, delete all duplicate rows for a guild, and create no indexes during startup.

The candidate `{guild:1}` singleton indexes remain duplicate-audit gated.

## Gates

Config command runtime and staging command sync:

```bash
MHCAT_FEATURE_AUTOCHAT_CONFIG_ENABLED=true
MHCAT_COMMAND_SYNC_INCLUDE_AUTOCHAT_CONFIG=true
```

Local fallback runtime:

```bash
MHCAT_FEATURE_AUTOCHAT_FALLBACK_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true
MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true
```

The fallback gate registers no commands and can be staged independently after a `chats` row exists. Configuration validation rejects the fallback gate unless all three gateway/message prerequisites are enabled.

## Not Implemented

The following paid path remains unavailable:

- request writes and response polling through `chatgpts`
- per-message debit writes to `chatgpt_gets.price`
- the 10-second in-flight guard and 40-second conversation reset
- the inferred external ChatGPT worker and its ownership/completion protocol

Guilds with positive balance therefore receive no Go auto-chat reply. Do not enable the local fallback expecting paid ChatGPT behavior.

## Staging Checklist

1. Use an isolated staging guild and database.
2. Configure a staging channel with `/自動聊天頻道` or seed one `chats` row.
3. Seed no `chatgpt_gets` row or a negative/malformed `price` to test local replies.
4. Enable the fallback gate and all required gateway/message intents.
5. Verify bot and DM messages are ignored.
6. Verify messages outside the configured channel are ignored.
7. Verify `你好` produces the legacy corpus response as a Discord reply after typing.
8. Verify `說出我是誰` responds immediately and does not ping mentions.
9. Verify `price: 0` and positive prices produce no local reply.
10. Keep Node and Go MessageCreate ownership exclusive during smoke testing.
