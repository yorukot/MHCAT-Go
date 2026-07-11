# Paid Auto-Chat Handoff Parity Contract

Status: parity-audited against the active legacy paid `MessageCreate` listener, Mongoose 6.4.6 Number/String/Boolean hydration, JavaScript UTF-16 and comparison behavior, current Go service/runtime/app/config/preflight wiring, and replica-set Mongo transaction tests. The event runtime remains disabled by default. The external worker and live staging behavior must still be confirmed before production ownership.

## Scope

This contract covers:

- positive-balance auto-chat eligibility and configured-channel routing;
- input and busy warnings, pricing, debit, worker request publication, and response UI;
- `chatgpt_gets` and `chatgpts` mixed BSON, timing, duplicates, transactions, migration, staging, and rollback.

Legacy sources:

- `events/Chatbot.js`, first `messageCreate` listener;
- `models/chat.js`;
- `models/chatgpt.js`;
- `models/chatgpt_get.js`;
- Mongoose 6.4.6;
- Node/JavaScript and discord.js 14.25.1 behavior.

The config commands are canonical in [89-autochat-config.md](89-autochat-config.md), and the local corpus path is canonical in [90-autochat-fallback.md](90-autochat-fallback.md). `/查看餘額` and `/兌換` are separate contracts. The Go bot preserves the Mongo handoff; it does not call OpenAI or another AI provider directly.

## Gates And Ownership

Enable only with every paid event gate:

```bash
MHCAT_FEATURE_AUTOCHAT_PAID_HANDOFF_ENABLED=true
MHCAT_AUTOCHAT_PAID_OWNERSHIP_CONFIRMED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true
MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true
```

The gates default to false. Configuration validation and staging preflight reject missing Gateway, Guild Messages, Message Content, or ownership confirmation. The feature has no slash command or command-sync include flag.

Ownership confirmation is an operator assertion, not a distributed lease. Before setting it:

1. stop the Node `events/Chatbot.js` owner for the target bot/guilds;
2. confirm a replica-set or sharded Mongo deployment that supports transactions;
3. confirm exactly one applicable `chats`, `chatgpts`, and `chatgpt_gets` row per guild;
4. confirm the external worker consumes the six legacy fields and preserves request `time`;
5. use disposable staging balances, handoffs, channels, and messages.

Overlapping Node and Go listeners can charge or publish twice. Multiple Go processes rely on the Mongo transaction and optimistic writes, not the ownership flag, to serialize a guild request.

Message Content exposes ordinary guild messages. Review privileged-intent approval, privacy policy, retention/logging, target channels, worker access, and Mongo credentials before enabling the runtime.

## Eligibility And Read Order

Both paths ignore direct messages, events without a guild, and bot-authored messages. Go also requires usable guild, channel, and source-message IDs before repository work.

Legacy reads in this order, and Go preserves the normal short circuit:

1. one `chatgpt_gets` balance lookup by exact guild;
2. return silently unless the hydrated balance is eligible;
3. one `chats` config lookup by exact guild;
4. return silently unless the hydrated channel exactly equals the event channel;
5. reject unsafe input or queue the paid handoff.

An ineligible balance therefore stays silent even if config Mongo is unavailable. Stored channel whitespace is not trimmed during comparison. The shared balance/config adapters preserve their canonical Mongoose scalar behavior.

Legacy proceeds when `data_chatgpt.price > 0`. Go intentionally narrows this to a positive finite value:

| Selected `price` | Paid behavior |
| --- | --- |
| missing row, missing/undefined field, null, exact empty/whitespace string, zero, negative, malformed, NaN | silent |
| positive finite Mongoose-compatible scalar | continue |
| positive infinity | silent in Go; legacy would continue |
| duplicate guild rows | Go fails closed; legacy selects one arbitrary row |

The transaction re-reads and revalidates the balance after the service precheck. A balance that becomes unavailable before queueing produces no warning, charge, or worker request.

## Unsafe Input And Busy UI

Any input containing the exact ASCII character `@` is unsafe. Before any debit or handoff write, the runtime:

1. sends a normal channel message with exact content `為防止伺服器招到tag攻擊，請勿在與機器人聊天時含有@`;
2. deletes the source message best-effort;
3. waits exactly four seconds;
4. deletes the warning best-effort.

When an existing handoff is younger than ten seconds, it sends `一次只能傳送一個消息，請等待機器人回復完成後在繼續!`, deletes the source best-effort, waits exactly two seconds, and deletes the warning best-effort. A busy request is not charged and does not modify handoff state.

Legacy includes `ephemeral: true` in the channel-send object, but ephemeral has no effect for ordinary channel messages. Go sends the same visible channel warning. Go intentionally suppresses all mentions in both warnings.

## Pricing

The exact legacy price is:

```text
JavaScript len(message.content) * 0.00003
```

`len` replaces each JavaScript UTF-16 code unit outside `0x00` through `0xff` with two ASCII characters, then takes the resulting length. Therefore:

- ASCII and Latin-1 code units cost one unit;
- other BMP code units cost two units;
- a supplementary character consists of two surrogate code units and costs four units.

Go reproduces this over UTF-16 code units. Empty messages have zero cost and can still publish. There is no sufficient-balance check: a positive balance may be overdrawn below zero, matching legacy behavior.

## Transaction And Handoff

Go intentionally performs the debit and publication in one Mongo transaction. Within it:

1. resolve exactly one `chatgpt_gets` row and preserve its raw `price` value and `_id`;
2. resolve zero or one `chatgpts` row;
3. reject duplicate balance or handoff rows before writing;
4. debit the exact balance row only if `_id`, guild, and raw `price` still match;
5. patch or upsert the handoff request;
6. commit both writes together.

The debit writes a BSON double result after Mongoose-compatible numeric coercion. No balance type normalization occurs before a successful request. A failed handoff write, write conflict, cancellation, or transaction abort leaves the balance unchanged.

The request patch always sets exactly these worker-visible values:

| Field | Request value |
| --- | --- |
| `guild` | Discord guild ID string |
| `reply` | `false` |
| `message` | exact source content |
| `time` | request Unix milliseconds |
| `resid_c` | null only for a missing/reset conversation; otherwise untouched |
| `resid_p` | null only for a missing/reset conversation; otherwise untouched |

Unknown extra fields on an existing handoff remain. A missing handoff uses a deterministic guild-scoped ObjectID so an ambiguous retried upsert cannot create unbounded duplicates. This `_id` remains valid to Mongoose and the dashboard.

## Timing And Conversation State

Legacy evaluates `Date.now() - data.time` after Mongoose Number hydration:

| Age/state | Result |
| --- | --- |
| age below `10000ms` | busy |
| exactly `10000ms` through exactly `40000ms` | preserve `resid_c` and `resid_p` |
| age above `40000ms` | set both IDs to null |
| missing/undefined/malformed/NaN time | preserve conversation |
| null time | reset because Mongoose retains null and subtraction coerces it to zero |
| positive infinity | busy |
| negative infinity | reset |

Fractional and scalar Mongoose Number values use the same strict boundaries. Go evaluates one captured request timestamp, while legacy calls `Date.now()` independently around its branches and write. The captured timestamp prevents branch/write drift and is the worker request identity.

## Worker Response And Discord Lifecycle

After a committed request:

1. send typing best-effort in the source channel;
2. wait exactly ten seconds;
3. read exactly one `chatgpts` row for the guild;
4. require its Mongoose-coerced `time` to equal the exact request timestamp;
5. reply to the original source message.

The worker's `reply` field is ignored; legacy always calls `message.reply`. Normal output is the Mongoose String-hydrated `message`. Supported scalar forms include strings/symbols, booleans, JavaScript-formatted numbers, Decimal128, ObjectID hex, unsigned timestamp decimal text, Buffer/binary UTF-8, JavaScript Date text, and regular-expression text. Missing, null, BSON JavaScript/Code, arrays, documents, MinKey, MaxKey, and other uncastable values produce no response.

If output contains `@`, the bot replaces the complete output with exact text `由於chatGPT回傳回來的訊息含有@，為防止遭到tag攻擊，已自動迴避該則消息!`. Go suppresses all user, role, everyone, and replied-user mentions for both normal and replacement replies.

The queued source message is not deleted or edited. The bot does not poll, retry, or wait beyond the fixed ten seconds. A worker that responds late, changes `time`, or leaves the request message unchanged can yield stale/no useful output; staging must verify the real worker contract.

## Failures And Intentional Safety Differences

Legacy ignores Mongo callback errors, does not await debit/publication writes, and can charge without publishing. It can also cross-wire overlapping requests and accepts arbitrary duplicate rows.

Go intentionally differs by:

- atomically committing debit and publication;
- rejecting duplicate `chatgpt_gets` or `chatgpts` rows;
- accepting only positive finite balances;
- comparing the original raw balance value during debit;
- binding responses to the exact request `time`;
- using a deterministic missing-handoff ObjectID;
- suppressing all mentions;
- stopping local fallback propagation after unsafe, busy, or queued paid handling;
- returning/logging repository and Discord send failures;
- cancelling warning/response delays with context shutdown.

Go preserves the normal pricing, overdraw, warning text/timing, busy/reset boundaries, request fields, conversation-ID behavior, typing, fixed response delay, source reply, ignored worker `reply`, and output safety substitution.

The feature writes no slash usage event because it is an event-only path.

## Data And Migration

No schema migration, backfill, normalization, collection rename, startup write, or startup index is required. Constructing the repositories and transaction runner against an empty database creates no collections.

Before staging, read-only audit:

- exact collection names `chats`, `chatgpts`, and `chatgpt_gets`;
- duplicate, missing, null, and mixed-type guild/key/value fields;
- current handoff `resid_c`, `resid_p`, `reply`, `message`, and `time` shapes;
- all balance writers, especially redeem and the external worker/operator tools;
- all handoff writers and whether the worker preserves `time`;
- existing indexes and transaction deployment topology.

Do not deduplicate or normalize production data merely to enable the feature. A unique `{guild:1}` index for each singleton remains an explicit owner-wide migration after clean audits and worker review. Paid queueing fails closed until duplicates are resolved; balance query, redeem, config, and local fallback retain their separately documented duplicate contracts.

## Parity Tests

Run focused coverage:

```bash
go test ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/core/services/autochat ./internal/discord/features/autochat ./internal/discord/events ./internal/app ./internal/config ./cmd/mhcat-staging-preflight
go test -race ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/core/services/autochat ./internal/discord/features/autochat ./internal/discord/events ./internal/app
go vet ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/core/services/autochat ./internal/discord/features/autochat ./internal/app
```

Replica-set transaction coverage is opt-in:

```bash
MHCAT_RUN_MONGO_TRANSACTION_INTEGRATION_TESTS=true \
MHCAT_MONGODB_URI='mongodb://127.0.0.1:27017/?replicaSet=rs0&directConnection=true' \
go test -race ./internal/adapters/mongo/repositories \
  -run '^TestAutoChatPaidMongoTransactionIntegration' -count=1
```

The integration suite proves startup no-mutation, lifecycle timing, rollback after handoff rejection, legacy overdraw, duplicate fail-closed behavior without partial writes, and concurrent single-charge publication.

## Staging Smoke

1. Use an isolated guild/database/channel, stop Node `events/Chatbot.js`, and snapshot `chats`, `chatgpts`, and `chatgpt_gets`.
2. Confirm replica-set/sharded transactions, worker compatibility, Message Content/privacy approval, exact singleton audits, and every paid gate.
3. Seed one configured channel, one small positive numeric balance, and no handoff; send one human prompt and verify one debit, one six-field request, typing, and one source reply after ten seconds.
4. Verify the worker preserves request `time`, updates `message`, and writes expected conversation IDs before the read.
5. Send a prompt costing more than the positive balance and verify one negative resulting balance plus one response.
6. Send an immediate second prompt; verify exact two-second warning cleanup, source deletion, no debit, and no request mutation.
7. Verify exactly 10 seconds preserves conversation IDs, exactly 40 seconds preserves them, and greater than 40 seconds nulls both.
8. Exercise missing, null, malformed, NaN, fractional, infinity, date, boolean, and numeric-string balance/time values against this contract.
9. Exercise every accepted/rejected worker scalar, stale/changed time, late response, `reply` true/false, and output containing `@`.
10. Seed duplicate balance and handoff fixtures separately; verify fail-closed behavior and no partial write. Remove only disposable fixtures after inspection.
11. Force handoff validation failure, write conflict, Mongo outage, typing/send failure, and cancellation; verify transaction rollback and no duplicate retry.
12. Enable local fallback concurrently and verify the canonical positive/local/silent balance split with only one user-visible response.
13. Confirm no startup collections/indexes, no usage rows, no unintended mentions, and no writes outside the target guild rows.

## Rollback

1. Disable `MHCAT_FEATURE_AUTOCHAT_PAID_HANDOFF_ENABLED` and stop the Go event owner.
2. Allow or cancel in-flight ten-second waits and inspect any request whose transaction committed but response was not sent.
3. Do not refund automatically. Reconcile each committed handoff timestamp and balance debit against worker/Discord evidence.
4. Restore only disposable smoke fixtures; no schema or index rollback is normally required.
5. Restore Node only after confirming no Go paid handler remains and ownership is exclusive.
6. Smoke one normal paid response, one busy request, and one local fallback state under the restored owner.

Production ownership remains blocked on live worker smoke, exclusive event ownership, transaction-capable Mongo, clean singleton audits, Message Content/privacy approval, and explicit acceptance of the safety differences above. No automatic database migration or index is required.
