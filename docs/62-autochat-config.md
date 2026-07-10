# Auto-Chat Runtime

Status: config commands, local fallback, and the bot-side paid handoff are implemented behind independent disabled-by-default gates. The paid path still requires the separately deployed external worker.

## Legacy References

- Set command: `MHCAT/slashCommands/實用工具/chat.js`
- Delete command: `MHCAT/slashCommands/實用工具/chat_delete.js`
- Runtime handler: `MHCAT/events/Chatbot.js`
- Local response corpus: `MHCAT/chat.json`
- Handoff models: `MHCAT/models/chatgpt.js`, `MHCAT/models/chatgpt_get.js`

## Implemented Surface

Config commands:

- `/自動聊天頻道`
- `/自動聊天頻道刪除`

Local fallback:

- reads `chats.channel` and `chatgpt_gets.price`
- uses the legacy local corpus for a missing, negative, or malformed balance
- preserves zero-balance silence
- preserves `說出`, fuzzy matching, typing, and the randomized 1-5 second delay
- suppresses all mentions in replies

Paid handoff:

- accepts only human guild messages in the configured channel with a positive finite balance
- rejects input containing `@`, deletes the source, and deletes the legacy warning after four seconds
- preserves the legacy 10-second in-flight guard and two-second busy warning cleanup
- debits `chatgpt_gets.price` by the JavaScript UTF-16 length rule times `0.00003`
- writes the exact worker contract in `chatgpts`: `guild`, `resid_c`, `resid_p`, `reply`, `message`, and `time`
- preserves `resid_c`/`resid_p` from 10 through 40 seconds and resets both after 40 seconds
- sends typing, waits the legacy fixed ten seconds, then reads the response for the exact request timestamp
- replies to the source message and substitutes the legacy safety warning when worker output contains `@`

The external worker code was not found in the workspace. Go publishes and consumes the legacy Mongo handoff; it does not call an AI provider directly.

## Data Safety

The balance debit and handoff publication run in one Mongo transaction. A rejected or failed `chatgpts` write therefore cannot leave a charge without a request. This intentionally improves the legacy non-transactional order and requires a replica-set or sharded Mongo deployment.

The repository:

- patches legacy fields instead of replacing worker-owned state
- targets existing rows by `_id`
- uses a deterministic ObjectID for a missing `chatgpts` singleton
- rejects duplicate `{guild}` rows in either `chatgpts` or `chatgpt_gets`
- accepts legacy numeric `price` and numeric millisecond `time` BSON types
- creates no indexes at startup

Run the duplicate audit before staging or production. Candidate `{guild:1}` unique indexes remain explicit, audit-gated operations.

## Gates

Config commands:

```bash
MHCAT_FEATURE_AUTOCHAT_CONFIG_ENABLED=true
MHCAT_COMMAND_SYNC_INCLUDE_AUTOCHAT_CONFIG=true
```

Local fallback:

```bash
MHCAT_FEATURE_AUTOCHAT_FALLBACK_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true
MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true
```

Paid handoff:

```bash
MHCAT_FEATURE_AUTOCHAT_PAID_HANDOFF_ENABLED=true
MHCAT_AUTOCHAT_PAID_OWNERSHIP_CONFIRMED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true
MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true
```

Set `MHCAT_AUTOCHAT_PAID_OWNERSHIP_CONFIRMED=true` only after all of these are true:

1. The external worker is confirmed active and compatible with the six legacy `chatgpts` fields.
2. Mongo supports transactions.
3. Duplicate audits for `chats`, `chatgpts`, and `chatgpt_gets` are clean.
4. The Node `events/Chatbot.js` MessageCreate owner is stopped for the target guilds.
5. The staging rows and channel are disposable.

The local and paid gates may be enabled together to restore the full legacy balance split. Paid handles positive balances; local handles missing, negative, or malformed balances; zero remains silent.

## Staging Checklist

1. Run `mhcat-staging-preflight` and review its paid-handoff warning.
2. Use a replica-set staging Mongo database and an isolated Discord guild/channel.
3. Seed exactly one `chats` row and one positive numeric `chatgpt_gets` row.
4. Confirm the worker changes `chatgpts.message`, preserves `time`, and sets its normal conversation fields within ten seconds.
5. Verify a normal message is charged once and receives the worker response.
6. Verify a second message inside ten seconds gets the transient busy warning and is not charged.
7. Verify conversation IDs are retained through 40 seconds and reset after 40 seconds.
8. Verify input and worker output containing `@` cannot ping users, roles, or everyone.
9. Enable the local fallback as well and verify negative/missing balance uses the corpus while zero remains silent.
10. Keep Node and Go MessageCreate ownership exclusive throughout the smoke test.

## Rollback

Disable the Go paid gate first, wait for in-flight ten-second reads to finish, and then restore the Node handler if needed. No schema rollback is required because Go writes only the legacy fields and deterministic `_id` values are valid Mongoose documents.
