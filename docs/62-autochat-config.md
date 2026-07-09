# Auto-Chat Config Slice

Status: implemented behind explicit runtime and command-sync gates.

## Legacy References

- Set command: `MHCAT/slashCommands/實用工具/chat.js`
- Delete command: `MHCAT/slashCommands/實用工具/chat_delete.js`
- Config model: `MHCAT/models/chat.js`
- Runtime message handler: `MHCAT/events/Chatbot.js`
- External handoff models: `MHCAT/models/chatgpt.js`, `MHCAT/models/chatgpt_get.js`

## Implemented Surface

This slice implements only the config commands:

- `/自動聊天頻道`
- `/自動聊天頻道刪除`

The `messageCreate` chatbot runtime is not implemented in this slice.

## UI/UX Parity

The implemented command paths preserve:

- command names and descriptions
- required `頻道` channel option for `/自動聊天頻道`
- text/news channel type restriction (`0`, `5`)
- runtime Manage Messages permission check
- public defer/edit response flow
- red permission and missing-config error embeds
- green `自動聊天系統` success embeds
- delete missing-config text `你沒有設定過，我不知道要刪除甚麼!`

## Mongo Compatibility

Collection: `chats`

Fields:

- `guild`
- `channel`

The Go repository writes the same fields as legacy. It updates all duplicate `{guild}` rows before falling back to an upsert, deletes all duplicate rows for a guild on delete, and does not create indexes during bot startup.

The candidate `{guild:1}` singleton index remains duplicate-audit gated.

## Gates

Runtime:

```bash
MHCAT_FEATURE_AUTOCHAT_CONFIG_ENABLED=true
```

Command sync:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_AUTOCHAT_CONFIG=true
```

Both flags must be paired in staging. `mhcat-staging-preflight` and the staging command-sync scripts reject auto-chat command sync when the runtime flag is not enabled.

## Not Implemented

This slice does not implement:

- `events/Chatbot.js`
- Message Content intent enablement
- Guild Messages gateway runtime
- local keyword auto-replies
- ChatGPT handoff writes/reads in `chatgpts`
- `chatgpt_gets.price` accounting
- the inferred external ChatGPT worker contract
- chatbot report-error button behavior

The config commands can be staged independently, but production chatbot runtime still requires a separate review for privacy, worker ownership, rate limits, mention safety, and duplicate `chats` data.

## Staging Checklist

1. Use an isolated staging guild and staging database.
2. Run command sync dry-run with `MHCAT_COMMAND_SYNC_INCLUDE_AUTOCHAT_CONFIG=true`.
3. Enable `MHCAT_FEATURE_AUTOCHAT_CONFIG_ENABLED=true` before applying command definitions.
4. Run `mhcat-staging-preflight`.
5. Apply guild-scoped command sync only after paired gate checks pass.
6. Verify `/自動聊天頻道` writes `chats.guild` and `chats.channel`.
7. Verify `/自動聊天頻道刪除` deletes the staging guild config and preserves the legacy missing-config error.
8. Confirm no chatbot replies are emitted by Go; message runtime remains intentionally disabled.
