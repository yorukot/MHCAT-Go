# Local Auto-Chat Fallback Parity Contract

Status: parity-audited against the active legacy MessageCreate listener, Mongoose 6.4.6 Number/String hydration, JavaScript comparison/string/object-key/UTF-16 behavior, the exact legacy response corpus, current Go service/runtime/app wiring, configuration validation, and staging preflight. The event runtime remains disabled by default. Live staging smoke is still required before production ownership.

## Scope

This contract covers:

- the event-only local response path for human guild messages;
- `chatgpt_gets` balance eligibility and `chats` channel routing;
- exact `說出` and corpus reply behavior, typing/delay, failures, privacy, staging, and rollback.

Legacy sources:

- `events/Chatbot.js`, second `messageCreate` listener;
- `chat.json`;
- `models/chat.js`;
- `models/chatgpt_get.js`;
- Mongoose 6.4.6;
- Node/JavaScript and discord.js 14.25.1 behavior.

The config slash commands are canonical in [89-autochat-config.md](89-autochat-config.md). Paid positive-balance ChatGPT handoff is canonical in [91-autochat-paid.md](91-autochat-paid.md); `/查看餘額` and `/兌換` are also separate contracts. This local runtime does not call an AI provider, debit balance, or read/write `chatgpts`.

## Gate And Ownership

Enable only with all event prerequisites:

```bash
MHCAT_FEATURE_AUTOCHAT_FALLBACK_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true
MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true
```

The feature gate defaults to false. Configuration validation and staging preflight reject the runtime unless Gateway, Guild Messages, and Message Content are all enabled. It has no slash command and no command-sync include flag.

Stop the Node `events/Chatbot.js` owner before enabling Go for the same bot/guild. There is no lease or shared event identity. Concurrent owners can both reply to the same fallback-eligible message.

Message Content exposes ordinary guild message text to the bot. Review Discord privileged-intent approval, privacy policy, retention/logging, target channels, and operator access before enabling it. The runtime requires Mongo read access to `chatgpt_gets` and `chats`, plus Discord typing/send access.

## Event Eligibility And Read Order

Both implementations ignore:

- direct messages and events without a guild;
- bot-authored messages, including bot member payloads;
- messages without usable channel/message identity;
- messages outside the selected configured channel;
- balances that are nonnegative under legacy coercion.

For an otherwise eligible event, legacy reads in this order and Go preserves the same normal short circuit:

1. one unsorted `chatgpt_gets.findOne({guild})`;
2. return silently when that row's hydrated `price >= 0`;
3. one unsorted `chats.findOne({guild})`;
4. return silently when missing or when hydrated `channel !== message.channel.id`;
5. produce a local reply.

Duplicate balance/config rows retain arbitrary first-match behavior. The configured channel comparison preserves stored whitespace and Mongoose String coercion under [89-autochat-config.md](89-autochat-config.md).

## Balance Split

The balance model field `price` is Mongoose `Number`. The exact local eligibility matrix is:

| Selected `chatgpt_gets.price` state | Local behavior |
| --- | --- |
| no row | reply locally |
| missing / BSON undefined | reply locally |
| malformed / NaN | reply locally |
| negative finite number | reply locally |
| negative infinity | reply locally |
| BSON null / exact empty string | silent because hydrated `null >= 0` is true |
| whitespace string / negative zero / zero | silent |
| positive finite number | silent; paid runtime owns this state when separately enabled |
| positive infinity | silent in legacy comparison |
| boolean/date/numeric scalar | use Mongoose Number hydration, then the same `>= 0` rule |

Go consumes the canonical balance string produced by the shared Mongoose-number adapter. The literal canonical `null` is treated as nonnegative; `undefined`, malformed, and `NaN` remain fallback eligible.

## Exact `說出` Replies

Legacy's broken guards effectively execute the special branch only when content contains exact `說出`. Go preserves the resulting behavior:

1. remove every `說` and `出` character from the complete message;
2. if nothing remains, reply `說出甚麼?`;
3. if original content contains `幹`, reply `很抱歉，讀取到你說出了一些不好的字元，因此拒絕說出w\n字元:`;
4. otherwise, if stripped content contains `我`, replace only the first `我` with `你` and wrap the result in double quotes;
5. otherwise reply with stripped content unchanged.

Because legacy evaluates `"幹" || "操" || "bitch"` as `"幹"`, only `幹` activates the rejection. `操` and `bitch` are repeated normally. These replies are immediate: no typing indicator and no artificial delay.

## Corpus Search

The embedded `legacy_chat.json` is byte-for-byte identical to legacy `chat.json`:

- SHA-256 `9a04b0212a7b32f1cfbe1cb5579a0ac35ec3e1ec8f72843291ede61cc9ee667a`;
- 22,431 bytes;
- 380 unique keys after JSON parsing.

Go reproduces JavaScript `Object.keys` ordering: canonical array-index keys first in ascending numeric order, then other keys in source insertion order. The first keys are `0`, `881`, and `v`.

Similarity preserves legacy behavior:

- lowercase both values with JavaScript-compatible Unicode casing;
- compute Levenshtein edit distance over UTF-16 code units;
- divide by the original longer UTF-16 length;
- replace the selected result only on strictly greater probability, so ties keep the first key;
- if every probability is zero, reply `我看不懂你的意思，在講一次好不好w` immediately.

The legacy code constructs a `回報錯誤` button but never includes it in the reply payload. Go correctly sends no component.

## Discord Lifecycle

For a matched corpus response:

1. send typing best-effort to the source channel;
2. wait an integer random delay from 1000 milliseconds inclusive through 5000 milliseconds exclusive;
3. send a Discord message reply referencing the source message with the exact corpus text.

The minimum is exactly `1s`; the maximum possible delay is `4.999s`. Special `說出` and unknown replies are immediate and do not send typing. The source message is never deleted or edited by this local path.

Legacy message replies can use Discord's default replied-user mention behavior and do not constrain mentions inside corpus/user-derived text. Go intentionally suppresses user, role, everyone, and replied-user mentions. No corpus text or `說出` output can ping anyone under Go.

## Failures

Legacy ignores Mongo callback errors. A balance lookup failure commonly looks like a missing balance and can continue to local config/reply; a config lookup failure commonly looks missing and stays silent.

Go intentionally returns repository errors to structured event logging and sends no reply. A nonnegative balance short-circuits before config lookup, so config failure cannot affect that silent state. Discord typing failures are best-effort; reply-send failures and delay context cancellation are returned and logged. No automatic reply retry occurs because retry can duplicate a message.

On shutdown/cancellation, Go cancels an in-flight delay and does not send the delayed reply. Legacy timers can continue until process termination. This is an intentional bounded-shutdown difference.

## Data And Migration

The local runtime performs only one optional `chatgpt_gets` read and one optional `chats` read per eligible human message. It performs no Mongo write, debit, insert, update, delete, transaction, index creation, normalization, backfill, or startup migration.

Before staging, audit without repairing:

- duplicate and mixed-type `chatgpt_gets.guild`/`price` rows under [87-balance-query.md](87-balance-query.md);
- duplicate and mixed-type `chats.guild`/`channel` rows under [89-autochat-config.md](89-autochat-config.md);
- current first-match winners and all external writers;
- whether configured channels and test messages are disposable;
- Message Content/privacy approval and exclusive Node/Go ownership.

Do not deduplicate, normalize, or create unique indexes solely to enable this read-only runtime. The paid runtime has stricter transaction/singleton requirements and must be evaluated separately.

This event path does not invoke slash usage middleware and writes no `all_use_counts` row.

## Intentional Differences

Intentional differences are limited to:

- all reply and replied-user mentions are suppressed;
- Mongo failures stop and log instead of being treated as missing data;
- context cancellation stops delayed replies;
- typing/send failures are handled explicitly;
- cryptographically secure randomness replaces `Math.random` while preserving the exact delay range.

Event filters, balance-first read order, Mongoose balance/channel coercion, arbitrary duplicates, missing/negative/malformed split, null/zero silence, exact corpus bytes/order/search, `說出` bugs, immediate versus delayed lifecycle, typing, source-message preservation, and no-Mongo-write behavior are preserved.

## Parity Tests

Run focused coverage:

```bash
go test ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/core/services/autochat ./internal/discord/features/autochat ./internal/discord/events ./internal/app ./internal/config ./cmd/mhcat-staging-preflight
go test -race ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/core/services/autochat ./internal/discord/features/autochat ./internal/discord/events ./internal/app
go vet ./internal/core/services/autochat ./internal/discord/features/autochat ./internal/discord/events ./internal/app
```

Existing opt-in real-Mongo tests for `chatgpt_gets` and `chats` prove Mongoose number/string hydration, arbitrary first match, no read-path mutation, and guild isolation. The local runtime tests use the exact embedded corpus and deterministic injected timing; they send no live Discord messages.

## Staging Smoke

1. Use an isolated staging guild/database/channel, stop Node `events/Chatbot.js`, and snapshot/audit `chats` and `chatgpt_gets`.
2. Confirm Gateway, Guild Messages, Message Content, privacy approval, bot channel access, and only the fallback feature gate are enabled.
3. Seed one exact configured channel and no balance row; send `你好` as a human and verify typing plus exact corpus reply after `1.000s` through `4.999s`.
4. Verify bot, DM, wrong-channel, missing-identity, and nonconfigured-guild messages remain silent.
5. Test `說出`, `說出操`, `說出bitch`, `說出幹`, and repeated `我`; verify exact immediate outputs with no typing.
6. Test an unrelated input with zero similarity and verify the exact immediate unknown reply.
7. Exercise missing, negative, negative-infinity, malformed, NaN, undefined, null/empty, whitespace, zero, positive, and positive-infinity balance fixtures against the canonical matrix.
8. Seed duplicate config/balance rows and verify one arbitrary first match without mutation or index creation.
9. Force Mongo, typing, reply-send, and cancellation failures; verify no duplicate retry and no raw data sent to Discord.
10. Verify source messages remain, all mentions are suppressed, no balance/config/usage row changes, and paid handoff stays disabled.

## Rollback

1. Disable `MHCAT_FEATURE_AUTOCHAT_FALLBACK_ENABLED` and stop the Go event owner.
2. No command sync removal is needed because this runtime registers no command.
3. Wait for or cancel in-flight event contexts; verify no delayed Go replies remain pending.
4. Restore only fixtures intentionally changed during smoke; the runtime itself writes no Mongo state.
5. Restore Node only after confirming no Go fallback handler remains and ownership is exclusive.
6. Smoke one missing/negative balance reply and one zero-balance silent case under the restored owner.

Production ownership remains blocked on live staging smoke, exclusive event ownership, Message Content/privacy approval, reviewed shared data/writers, and acceptance of the mention/error/cancellation differences. No database migration or index is required.
