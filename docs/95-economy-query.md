# Economy Query Parity Contract

Status: parity-audited against `slashCommands/代幣系統/coin.js`, both legacy Mongoose models, and the current Go definition, service, Mongo adapter, Discord handler, app wiring, staging guards, and race coverage. Live Discord staging and operator-gated real-Mongo execution remain required before production rollout.

## Scope

This contract covers `/代幣查詢` with no option and with optional user option `使用者`. Other economy, gacha, work, and XP behavior is outside this contract.

## Definition And Ownership

The chat-input definition has name `代幣查詢`, description `查詢你有多少代幣`, and one optional user option `使用者` described as `要查詢的使用者`. It has no default member permission. Legacy cooldown metadata is `10`, but legacy did not centrally enforce it and Go adds no local throttle. Compatibility metadata retains the malformed legacy URL `https://docsmhcat.yorukot.meocs/coin`.

Runtime is disabled by default and requires `MHCAT_FEATURE_ECONOMY_QUERY_ENABLED=true`. Staging guild command sync separately requires `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_QUERY=true`; config, preflight, and staging scripts reject sync without runtime. Bot startup never syncs Discord commands.

Global slash middleware is the only production usage owner. With usage tracking enabled, every attempt records exactly one best-effort `all_use_counts` event before route lookup and handler work. App wiring removes the module tracker to prevent a second success-only event. With usage tracking disabled, this command writes nothing.

## Interaction And User Selection

The handler immediately defers ephemerally and edits the original response. With no option it queries the caller and labels the subject `你`. With `使用者`, it queries the resolved user ID and uses the resolved Discord username. Guild-member lookup is only a fallback for adapter/test input without resolved option metadata; failed fallback displays the user ID.

The footer icon always remains the caller's display avatar, including when another user is selected.

## Success UI

The single ephemeral embed uses a random 24-bit color and no timestamp. Its title is:

```text
<:money:997374193026994236><subject>目前有:`<coin>`個代幣!
```

Its description preserves the exact question emoji, sign-in/chat text, cat-jump emoji, configured gacha threshold, and shop sentence. Allowed mentions are suppressed.

With a `gift_changes` row, the footer performs JavaScript-style `coin_number - coin` coercion. A positive result displays `<subject>還差:你還差<difference>就可以扭蛋了，加油!!`; zero, negative, or `NaN` displays `<subject>還差:你可以扭蛋了!!使用\`/扭蛋\`進行扭蛋`.

Without a `gift_changes` row, the description threshold defaults to `500`, while the footer intentionally remains `<subject>還差:500` regardless of balance. This preserves the legacy ternary-placement quirk.

## Legacy BSON Scalars

Mongoose exposed raw scalar values through template interpolation and subtraction. The read-only Go path retains display metadata separately from operational integer fields:

| BSON state | Visible text | Subtraction behavior |
| --- | --- | --- |
| missing, malformed, or BSON `NaN` | `undefined` | `NaN`; says gacha is available |
| null or empty raw value | `null` | coerces to zero |
| integer or decimal | JavaScript-style number text | numeric comparison |
| positive infinity | `Infinity` | numeric comparison |
| negative infinity | `-Infinity` | numeric comparison |

This applies independently to `coins.coin` and an existing `gift_changes.coin_number`. Typed integer fields remain available to mutating consumers; compatibility text is read-only presentation state.

## Missing Rows And Errors

No matching `coins` row produces one fixed red embed titled:

```text
<a:Discord_AnimatedNo:1015989839809757295> | 你還沒有任何代幣欸使用`/簽到`或是多講話，都可以獲得代幣喔!
```

Legacy callbacks ignored Mongo `err`, so backend failure was treated as absent data or left the deferred interaction unresolved. Go propagates non-not-found repository failures to the controlled runtime error path. This intentional reliability difference prevents internal leakage; successful and true no-row UI remains exact.

## Mongo And Migration

The command performs unsorted `findOne`-equivalent reads from `coins` by `{guild, member}`, then `gift_changes` by `{guild}` only after a balance is found. Duplicate matches are deliberately allowed and selection remains arbitrary, as in Mongoose.

Repository construction obtains collection handles only. It creates no database, collection, index, migration, backfill, or row. Execution performs no economy write. No migration is required for this command. Before any future unique key, operators must separately reconcile duplicates; that operation is outside this contract and can change which legacy row displays.

## Verification

Run:

```bash
go test ./internal/adapters/mongo/documents \
  ./internal/adapters/mongo/repositories \
  ./internal/core/services/economy \
  ./internal/discord/features/economy \
  ./internal/app ./internal/config ./cmd/mhcat-staging-preflight

go test -race ./internal/adapters/mongo/documents \
  ./internal/adapters/mongo/repositories \
  ./internal/core/services/economy \
  ./internal/discord/features/economy ./internal/app

go vet ./internal/adapters/mongo/documents \
  ./internal/adapters/mongo/repositories \
  ./internal/core/services/economy \
  ./internal/discord/features/economy ./internal/app

go run ./tools/parity-audit --legacy-root ../MHCAT --format markdown
```

The static report must remain `74/74` with zero drift, missing definitions, extras, or parser errors.

Real Mongo verification uses a disposable generated database:

```bash
MHCAT_RUN_MONGO_INTEGRATION_TESTS=true \
MHCAT_MONGODB_URI='<disposable-uri>' \
go test ./internal/adapters/mongo/repositories \
  -run '^TestEconomyQueryMongoIntegration'
```

It proves startup has no collection side effect, scalar hydration survives the driver, and duplicates remain readable. Cleanup drops the generated database. Never point this test at production.

## Staging Smoke

1. Stop Node interaction ownership and use a staging token, guild, and reviewed read-only Mongo data.
2. Enable both query flags and run preflight plus guarded guild command-sync dry-run; confirm only the managed addition/update is planned.
3. Query self and another resolved user. Verify ephemeral defer/edit, exact username, caller footer avatar, random color, and mention suppression.
4. Exercise no balance, no config, configured below/equal/above balance, and disposable missing/null/decimal/infinite scalar fixtures.
5. Seed duplicate disposable rows and verify one existing match renders without writes; do not claim deterministic ordering.
6. With usage disabled, compare collections/indexes before and after and verify no writes. With usage enabled separately, verify one global increment per attempt.
7. Force Mongo and Discord edit failures; verify no internal error text, duplicate response, collection creation, or economy mutation.

## Rollback

Disable query runtime and remove the managed staging guild command through guarded command sync. Stop Go ownership before restoring Node. No economy rollback or migration reversal exists because this command writes no data or indexes. Usage rows, when separately enabled, follow the global usage rollback procedure.

Production rollout remains gated on live Discord smoke, disposable real-Mongo execution, and exclusive Node/Go interaction ownership.
