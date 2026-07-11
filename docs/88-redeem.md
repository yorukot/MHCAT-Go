# Redeem Parity Contract

Status: parity-audited against the active legacy slash command, global slash dispatcher, Mongoose 6.4.6 number hydration and document deletion, Mongo first-match behavior, current Go document/repository/service/handler/runtime wiring, command sync, and staging preflight. Runtime and command sync remain disabled by default. Live staging smoke is still required before production ownership.

## Scope

This contract covers:

- `/兌換` metadata, public availability, usage, ephemeral UI, and failures;
- exact `codes` lookup, expiry, consumption, and duplicate behavior;
- `chatgpt_gets` balance replacement, mixed BSON, partial progress, staging, and rollback.

Legacy sources:

- `slashCommands/管理系統/get_something.js`
- `models/code.js`
- `models/chatgpt_get.js`
- `events/SlashCommands.js`
- Mongoose 6.4.6
- discord.js 14.25.1

`/查看餘額`, local auto-chat, and paid ChatGPT handoff are separate runtime contracts. They share `chatgpt_gets` but are not enabled by this command.

## Gates And Ownership

Enable only with paired staging flags:

```bash
MHCAT_FEATURE_REDEEM_ENABLED=true
MHCAT_STAGING_MODE=true
MHCAT_COMMAND_SYNC_INCLUDE_REDEEM=true
```

Both feature flags default to false. Command sync is guild-scoped and staging-only. Preflight rejects sync without runtime and warns when runtime is enabled without sync. Runtime startup opens the existing `codes` and `chatgpt_gets` collections and registers the route only when the repository is available.

Stop the Node `/兌換` owner before enabling Go for the same bot/guild. There is no shared lease. Concurrent owners can compete to acknowledge and consume the same code.

The command requires normal Gateway interaction delivery and Mongo read/write access. It requires no Message Content, Guild Members, reaction, or voice intent. Enabling it does not enable either auto-chat runtime.

## Definition And Usage

The command remains publicly discoverable with:

- name `兌換`;
- description `兌換代碼`;
- chat-input type;
- one required string option `代碼`, description `輸入您的代碼`;
- no Discord default-member permission.

There is no runtime permission check. Legacy cooldown metadata is `60`, but the global dispatcher does not enforce cooldowns. Go adds no cooldown.

Usage belongs only to global slash middleware. With tracking enabled, every routed success, missing, expired, validation, or backend-failure attempt records exactly one best-effort `兌換` increment before repository work. Production wiring removes the handler-level tracker to prevent a second success-only event.

## Exact UI

The command defers ephemerally before Mongo lookup and edits the deferred original response. Visibility is fixed by the defer. Mentions are explicitly suppressed.

Success contains one embed with:

- author name `成功兌換代碼!`;
- author icon `https://media.discordapp.net/attachments/991337796960784424/1078883215462383697/success.gif`;
- footer `你可以使用/查看餘額進行查詢剩餘餘額`;
- discord.js named `Green`, exact `0x57F287`;
- no title, description, fields, components, or files.

Missing code contains one embed with:

- title `<a:Discord_AnimatedNo:1015989839809757295> | 找不到這個代碼!`;
- discord.js named `Red`, exact `0xED4245`.

Expired code uses the same red shape and exact title:

```text
<a:Discord_AnimatedNo:1015989839809757295> | 這個代碼為防止遭人惡意使用，已過期，如果你是代碼擁有者，請前往支援伺服器開啟客服頻道!
```

Controlled validation/backend failures use:

```text
<a:Discord_AnimatedNo:1015989839809757295> | 很抱歉，出現了未知的錯誤，請重試!
```

Raw Mongo errors, credentials, query details, and code values are never shown.

## Code Read And Expiry

Both implementations query one row by exact `{code: <raw option>}` with no trimming and no sort. Padded and all-space codes therefore remain distinct and usable when stored exactly. If duplicates exist, Mongo may return any matching row.

Mongoose schema fields `price` and `time` are `Number`. Go mirrors Mongoose number hydration before validation and expiry. Relevant stored-value behavior includes:

| Stored value | Hydrated number behavior |
| --- | --- |
| integers/doubles/Decimal128/numeric strings | numeric value |
| hexadecimal string `0x10` | `16` |
| exponent string `1e3` | `1000` |
| BSON null / exact empty string | `0` |
| missing / BSON undefined | `NaN` |
| malformed string/document/array | `NaN` |
| booleans and dates | JavaScript/Mongoose numeric conversion |
| values above JavaScript safe integer range | rounded as a JavaScript number |

A code expires only when `Date.now() - time > 604800000`. The exact seven-day boundary remains valid. Null time hydrates to zero and is expired. Missing or malformed time hydrates to `NaN`; the comparison is false, so the code remains usable.

Negative, zero, infinite, and ordinary numeric prices retain JavaScript arithmetic behavior. Missing or malformed price hydrates to `NaN`. Go intentionally rejects `NaN` before deleting the code; see Intentional Differences.

## Consumption And Balance Shape

After validation and expiry, Go deletes the exact fetched code with `{_id, code}`. This preserves legacy document-instance `data_code.delete()` behavior: with duplicate codes, only the fetched row is consumed and its price is used. The identity-qualified delete also means concurrent attempts using the same fetched row can succeed only once.

After code deletion, both implementations perform one unsorted `chatgpt_gets.findOne({guild})`:

- with no row, insert fresh `{guild, price: codePrice}`;
- with a row, calculate `existing.price + codePrice` using Mongoose-number/JavaScript arithmetic;
- delete only that selected row by identity;
- insert fresh `{guild, price: next}` with a new `_id`;
- leave all other duplicate guild rows unchanged;
- drop extra fields from the selected row because the replacement contains only `guild` and `price`.

Go awaits each operation. Legacy starts document deletes and saves without awaiting them, so its success edit can race persistence. Go sends success only after the insert completes.

There is no transaction, upsert, deduplication, retry, repair, or index creation. Different codes redeemed concurrently for one guild can still race over the same selected balance row; one code may be consumed even when its balance replacement later fails.

## Failure And Partial Progress

The write order is intentionally non-transactional:

1. delete the fetched code;
2. read one balance row;
3. delete that selected balance row when present;
4. insert the replacement balance row;
5. edit the deferred response to success.

Observed and preserved outcomes:

| Failure point | Go state and UI |
| --- | --- |
| code lookup fails | no feature write; generic red error |
| code identity delete fails | no confirmed consume; generic red error |
| code was already consumed | no balance write; missing-code red error |
| balance read fails after code delete | code lost, balance unchanged; generic red error |
| malformed selected balance price | code lost, balance row preserved; generic red error |
| selected balance delete fails/loses race | code lost, no replacement; generic red error |
| replacement insert fails | code and selected balance row lost; generic red error |
| Discord edit fails after insert | code consumed and balance credited; error returned to dispatcher |

Operations are not retried automatically because replay could double-credit or consume another duplicate. Recovery requires an operator-reviewed code/balance repair based on backups and logs.

Legacy ignores callback errors and does not await deletes/saves. It can show success before persistence and can destroy the code and selected balance row when a malformed price leads the replacement save to reject. Go intentionally fails visibly and avoids some malformed-state destruction.

## Data And Migration

No automatic database migration is required or performed. Enabling the feature uses existing collections and creates no index.

Before staging, audit and snapshot:

- duplicate and mixed-type `codes.code`, `codes.price`, and `codes.time` rows;
- duplicate and mixed-type `chatgpt_gets.guild` and `chatgpt_gets.price` rows;
- existing indexes on both collections;
- every external writer to `codes` and `chatgpt_gets`;
- whether staging codes and balances are disposable.

Do not normalize values, deduplicate rows, or create unique `{code:1}` or `{guild:1}` indexes solely to enable redeem. Both duplicate first-match behaviors are part of the observed legacy contract. Paid auto-chat has stricter singleton requirements and must be assessed under its separate contract.

Optional global usage tracking may update `all_use_counts`. That generic collection is not redeem state and requires no redeem migration.

## Intentional Differences

Intentional differences are limited to:

- Go rejects malformed/`NaN` code prices before deletion instead of allowing legacy's destructive unawaited flow and misleading success UI;
- malformed selected balance prices preserve that row, while the already-consumed code still requires repair;
- backend/write failures receive controlled red UI instead of being ignored or reported as success;
- operations are awaited, and success is sent only after balance insertion;
- exact fetched identity and atomic code deletion prevent two Go attempts from consuming one row twice;
- invalid synthetic/no-guild interactions receive controlled error handling;
- mentions are explicitly suppressed.

Raw code matching, arbitrary duplicate selection, exact seven-day check, Mongoose number coercion, negative prices, one-row delete-plus-insert balance replacement, duplicate preservation, non-transactional write order, public definition, and ephemeral UI are preserved.

## Parity Tests

Run focused coverage:

```bash
go test ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/core/domain ./internal/core/ports ./internal/core/services/redeem ./internal/discord/features/redeem ./internal/app ./internal/config ./internal/parity ./cmd/mhcat-command-sync ./cmd/mhcat-staging-preflight
go test -race ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/core/domain ./internal/core/ports ./internal/core/services/redeem ./internal/discord/features/redeem ./internal/app
go vet ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/core/domain ./internal/core/ports ./internal/core/services/redeem ./internal/discord/features/redeem ./internal/app
go run ./tools/parity-audit --legacy-root ../MHCAT --format markdown
```

Run the opt-in integration tests only against disposable Mongo:

```bash
MHCAT_RUN_MONGO_INTEGRATION_TESTS=true \
MHCAT_MONGODB_URI='<disposable-uri>' \
go test ./internal/adapters/mongo/repositories \
  -run '^TestRedeemMongoIntegration' \
  -count=1
```

The integration tests create and drop only generated test databases. They lock exact fetched-code deletion, duplicate preservation, one-row balance replacement, new `_id`, dropped extra fields, negative price, malformed balance failure state, guild isolation, and one successful concurrent consume.

## Staging Smoke

1. Use an isolated staging guild/database, stop the Node command owner, and snapshot `codes` and `chatgpt_gets`.
2. Audit types, duplicates, indexes, and external writers without repairing them. Use only disposable fixtures.
3. Enable paired flags, run preflight and command-sync dry-run, review guild apply, and start one Go owner.
4. Confirm public discoverability, exact required string option, no default permission, and ephemeral defer/edit behavior.
5. Redeem a fresh numeric code; verify exact success UI, exact code `_id` deletion, one selected balance replacement with a new `_id`, and no change to another guild.
6. Verify padded/all-space lookup, negative price, exact seven-day boundary, expired retention, null-time expiry, and missing/malformed-time usability.
7. Seed duplicate codes and balances; verify only the fetched code and one selected balance row are replaced, while other duplicates remain.
8. Verify missing and expired exact red UI, then force read/write failures and verify generic red UI with no raw detail.
9. Exercise malformed code and balance prices only with disposable fixtures; verify the documented partial-progress matrix and repair procedure.
10. With usage tracking enabled separately, verify exactly one increment for success, missing, expired, and backend failure.

## Rollback

1. Disable command-sync inclusion and remove only the managed staging `兌換` command.
2. Disable the runtime gate and stop the Go owner before restoring Node.
3. Reconcile every staging attempt before restoring fixtures: a displayed error does not prove the code or balance was unchanged.
4. Restore consumed codes or lost balance rows only from reviewed snapshots/logs; do not blindly replay requests.
5. Do not create/drop indexes, normalize values, or deduplicate as part of emergency rollback.
6. Restore Node only after confirming no Go redeem route remains, then smoke one new disposable code under the restored owner.

Production ownership remains blocked on live staging smoke, exclusive command ownership, reviewed shared writers/types/duplicates, backups and repair procedures for non-transactional partial progress, and acceptance of the documented intentional differences. No automatic migration or index is required.
