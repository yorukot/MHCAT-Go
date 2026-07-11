# Balance Query Parity Contract

Status: parity-audited against the active legacy slash command, global slash dispatcher, Mongoose 6.4.6 `Number` hydration, Mongo `findOne` behavior, current Go document/repository/service/handler/runtime wiring, command sync, and staging preflight. Runtime and command sync remain disabled by default. Live staging smoke is still required before production ownership.

## Scope

This contract covers:

- `/查看餘額` metadata, public availability, usage, ephemeral UI, and failures;
- read-only `chatgpt_gets` lookup, missing/duplicate/mixed-BSON behavior, migration, staging, and rollback.

Legacy sources:

- `slashCommands/管理系統/check_price.js`
- `models/chatgpt_get.js`
- `events/SlashCommands.js`
- Mongoose 6.4.6
- discord.js 14.25.1

`/兌換`, local auto-chat, and the paid ChatGPT handoff in [91-autochat-paid.md](91-autochat-paid.md) are separate writer/runtime contracts. They share `chatgpt_gets` but are not enabled by this command.

## Gates And Ownership

Enable only with paired staging flags:

```bash
MHCAT_FEATURE_BALANCE_QUERY_ENABLED=true
MHCAT_STAGING_MODE=true
MHCAT_COMMAND_SYNC_INCLUDE_BALANCE_QUERY=true
```

Both feature flags default to false. Command sync is guild-scoped and staging-only. Preflight rejects sync without runtime and warns when runtime is enabled without sync. Runtime startup opens the existing `chatgpt_gets` collection and registers the route only when the repository is available.

Stop the Node `/查看餘額` owner before enabling Go for the same bot/guild. There is no shared lease. Concurrent owners can compete to acknowledge the same interaction even though both feature paths are read-only.

The command requires normal Gateway interaction delivery but no Message Content, Guild Members, reaction, or voice intent. It needs Mongo read access only.

## Definition And Usage

The command remains publicly discoverable with:

- name `查看餘額`;
- description `查看剩餘餘額`;
- chat-input type;
- no options;
- no Discord default-member permission.

There is no runtime permission check. Legacy cooldown metadata is `10`, but the global dispatcher does not enforce cooldowns. Go adds no cooldown.

Usage belongs only to global slash middleware. With tracking enabled, every routed success or backend-failure attempt records exactly one best-effort `查看餘額` increment before repository work. Production wiring removes the handler-level tracker to prevent a second success-only event. The optional global usage write is independent of balance data.

## Exact UI

The command defers ephemerally before Mongo lookup and edits the deferred original response. The successful response contains one embed with:

- author name `伺服器目前剩於餘額: <price>`;
- author icon `https://media.discordapp.net/attachments/991337796960784424/1078883215462383697/success.gif`;
- discord.js named `Green`, exact `0x57F287`;
- no title, description, footer, fields, components, or files.

The misspelling `剩於` is preserved. The response remains ephemeral because visibility is fixed by the initial defer. Mentions are explicitly suppressed.

## Mongo Read Contract

The legacy Mongoose model name `chatgpt_get` resolves to collection `chatgpt_gets` with schema fields `guild: String` and `price: Number`. Both implementations query one row by exact `{guild: <interaction guild ID>}` with no sort.

If no row exists, the displayed amount is `0`. If duplicate rows exist, Mongo may return any matching row; in a normal insertion-order collection the first row is observed until it is removed. Neither implementation merges, sums, rejects, repairs, or deletes duplicates in this read path.

The Go decoder mirrors Mongoose `Number` hydration and JavaScript template-string formatting for mixed legacy BSON. Audited examples include:

| Stored `price` | Displayed value |
| --- | --- |
| field missing / BSON undefined | `undefined` |
| BSON null / exact empty string | `null` |
| whitespace string | `0` |
| numeric string `12.5` | `12.5` |
| hexadecimal string `0x10` | `16` |
| exponent string `1e3` | `1000` |
| malformed string/document/array/NaN | `undefined` |
| booleans | `1` / `0` |
| BSON date | millisecond number |
| numeric binary data | parsed number |
| positive/negative infinity | `Infinity` / `-Infinity` |
| integer `9007199254740993` | `9007199254740992` after JavaScript rounding |
| `1e20` / `1e21` | `100000000000000000000` / `1e+21` |

Decimal128, integer, double, Symbol, timestamp, ObjectID, and binary categories use the same Mongoose-number conversion where valid. JavaScript scientific notation thresholds and exponent formatting are preserved.

## Failures

Legacy ignores the `findOne` callback error. Because `data` is absent on a normal backend failure, it commonly edits the response to the same green `0` balance, masking the outage.

Go intentionally returns a controlled ephemeral error instead of leaving the defer unresolved or reporting a false balance:

- title `<a:Discord_AnimatedNo:1015989839809757295> | 很抱歉，出現了未知的錯誤，請重試!`;
- discord.js named `Red`, exact `0xED4245`;
- no raw Mongo error, URI, credential, query, or internal detail.

Missing rows are not errors and still display green `0`. A Discord defer/edit failure is returned to the dispatcher and structured logging; writes are not retried because this feature performs none.

## Data And Migration

The feature executes one `findOne` and performs no balance write, upsert, delete, transaction, index creation, schema migration, repair, or startup backfill. Enabling or disabling it leaves `chatgpt_gets` unchanged. No database migration is required.

Before staging, audit `chatgpt_gets` because the same collection is written by redeem and paid auto-chat:

- count duplicate `{guild}` rows;
- inspect null, missing, string, non-finite, and malformed `price` values;
- record existing indexes and external writers;
- preserve rows as-is for query parity.

Do not deduplicate, normalize, or create a unique `{guild:1}` index solely to enable this read command. Paid auto-chat intentionally fails closed on duplicate state and has stricter rollout requirements; this query intentionally preserves legacy first-match behavior.

Optional global usage tracking may update `all_use_counts`, as it does for every slash command. That generic collection is not balance state and requires no balance migration.

## Intentional Differences

Intentional differences are limited to:

- Mongo/backend failure returns a controlled red error instead of legacy's misleading green zero;
- invalid synthetic/no-guild interactions receive controlled error handling instead of dereferencing `interaction.guild.id`;
- mentions are explicitly suppressed.

Exact public definition, no runtime permission, unenforced cooldown, ephemeral defer/edit lifecycle, author text/icon, green color, missing-row zero, Mongoose number display, exact guild lookup, arbitrary duplicate first match, and read-only behavior are preserved.

## Parity Tests

Run focused coverage:

```bash
go test ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/core/services/utility ./internal/discord/features/balance ./internal/app ./internal/config ./internal/parity ./cmd/mhcat-command-sync ./cmd/mhcat-staging-preflight
go test -race ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/core/services/utility ./internal/discord/features/balance ./internal/app
go vet ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/core/services/utility ./internal/discord/features/balance ./internal/app
go run ./tools/parity-audit --legacy-root ../MHCAT --format markdown
```

Run the opt-in integration test only against disposable Mongo:

```bash
MHCAT_RUN_MONGO_INTEGRATION_TESTS=true \
MHCAT_MONGODB_URI='<disposable-uri>' \
go test ./internal/adapters/mongo/repositories \
  -run '^TestBalanceMongoIntegrationPreservesMongooseValuesAndFirstMatch$' \
  -count=1
```

The integration test locks mixed values, first-match duplicates, missing rows, unchanged row counts, and guild isolation through the real driver. It drops only its generated test database.

## Staging Smoke

1. Use an isolated staging guild/database, stop the Node command owner, and back up or snapshot `chatgpt_gets` before seeding fixtures.
2. Audit duplicates/types/indexes/external writers; do not repair them for this command.
3. Enable paired flags, run preflight and command-sync dry-run, review guild apply, and start one Go owner.
4. Confirm the command is publicly discoverable with no options/default permission; run it as a normal member.
5. With no guild row, verify one ephemeral exact green author embed displays `0`.
6. Seed a numeric row and verify exact amount, typo, GIF icon, color, and no Mongo mutation.
7. Seed disposable null, empty, hex/exponent string, malformed, NaN, infinity, and large-integer rows; verify the canonical display matrix.
8. Seed two duplicate rows, verify one arbitrary first match, remove only that fixture manually, and verify the remaining value; do not infer production duplicate safety from this query.
9. Force a Mongo read failure and verify exact ephemeral `0xED4245` generic UI with no raw detail.
10. With usage tracking enabled separately, verify one increment for success and backend failure, no duplicate handler event, no feature/index writes, and another guild remains isolated.

## Rollback

1. Disable command-sync inclusion and remove only the managed staging `查看餘額` command.
2. Disable the runtime gate and stop the Go owner before restoring Node.
3. Restore only fixtures intentionally changed during smoke; the command itself writes no balance data.
4. Do not create/drop indexes or deduplicate as part of emergency rollback.
5. Restore Node only after confirming no Go balance route remains.
6. Recheck missing-row zero and one numeric row under the restored owner; remember Node masks backend failures as zero.

Production ownership remains blocked on live staging smoke, exclusive command ownership, reviewed `chatgpt_gets` writers/types/duplicates, and acceptance of the controlled backend-error difference. No database migration or index is required.
