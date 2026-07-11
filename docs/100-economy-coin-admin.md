# Economy Coin Admin Parity Contract

Status: parity-audited against `slashCommands/代幣系統/addcoin.js`, `models/coin.js`, and current Go definition, handler, service, scalar adapter, Mongo repository, app wiring, staging guards, integration harness, and race coverage. Live Discord and operator-gated real-Mongo smoke remain required before production rollout.

## Scope

This contract covers `/代幣增加`. Despite its name, the command can add or reduce a selected user's balance. It does not authorize sign-in, games, reset, XP reward, duplicate cleanup, or index creation.

## Definition And Ownership

The command preserves exact name/description, required user/string/integer options, malformed choice description/localizations, `add`/`reduce` choices, Manage Messages default permission, cooldown metadata `10`, and malformed documentation URL `https://docsmhcat.yorukot.meocs/coin_increase`. Legacy did not centrally enforce the cooldown; Go adds no local throttle.

Runtime is disabled by default behind `MHCAT_FEATURE_ECONOMY_COIN_ADMIN_ENABLED=true`. Staging command sync separately requires `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_ADMIN=true`; config, preflight, and scripts reject sync without runtime. Startup never syncs commands or creates Mongo objects.

Global slash middleware owns usage. With tracking enabled, success, permission denial, and arithmetic error each record exactly one best-effort event. App wiring prevents module double counting.

## Lifecycle And UI

The command publicly defers, then checks runtime Manage Messages before mutation. Errors edit to one red embed with exact animated-error prefix and one of:

- `你需要有\`訊息管理\`才能使用此指令`;
- `不可減到負數!`;
- `不可以加超過\`999999999\`!!`;
- the controlled unknown-error fallback for malformed/internal command states.

Success uses one random-color embed with exact title `<:money:997374193026994236>已為<resolved username>\`增加|減少\`:\`<raw amount>\`個代幣!`. The target name comes from resolved Discord user-option metadata before any guild lookup. Footer text is the selected Chinese operation concatenated with the raw signed amount, and its icon is the moderator's guild display avatar with user-avatar fallback.

The selected operation controls the visible label even when a negative amount reverses the effective arithmetic. No user mention is emitted, and allowed mentions are empty.

## Signed Arithmetic

Discord integer `數量` accepts positive, zero, and negative values. Existing rows use JavaScript-equivalent arithmetic:

- add computes `coin + Number(amount)` and rejects only when the result is above `999999999`;
- reduce computes `coin - Number(amount)` and rejects only when the result is below zero.

Therefore negative add can create a negative balance, negative reduce can increase above `999999999`, and zero succeeds. These are legacy behaviors, not recommended administrative policy.

For a missing row, reduce always returns `不可減到負數!`. Add creates the row with raw amount and `today:1`; creation is intentionally uncapped, including negative and over-limit amounts, matching Mongoose casting of legacy `today:true` to Number `1`.

## Scalars And Duplicates

Existing `coin` uses preserved Mongoose-visible number text. Decimal, null-as-zero, and positive/negative infinity retain JavaScript arithmetic and operation-specific comparisons. The resulting BSON number and response metadata preserve decimals/infinities. Missing, malformed, or NaN coin values fail before Go mutation instead of returning legacy success while a Mongoose update cast likely fails.

The repository reads one arbitrary natural `{guild,member}` row, computes from it, and updates one arbitrary matching row with `$set`. Extra duplicates remain unchanged and visible to rollback readers/ranks. No unique index, transaction, or automatic deduplication is created.

Concurrent admins can still overwrite each other because compatibility uses read/compute/set rather than an atomic arithmetic pipeline. Production requires exclusive Node/Go command ownership, duplicate audit, and an explicit concurrency decision.

## Verification

```bash
go test ./internal/core/domain \
  ./internal/core/services/economy \
  ./internal/adapters/mongo/documents \
  ./internal/adapters/mongo/repositories \
  ./internal/discord/features/economy \
  ./internal/adapters/discordgo ./internal/discord/interactions \
  ./internal/app ./internal/config ./cmd/mhcat-staging-preflight

go test -race ./internal/core/services/economy \
  ./internal/adapters/mongo/repositories \
  ./internal/discord/features/economy ./internal/app

go vet ./...
go run ./tools/parity-audit --legacy-root ../MHCAT --format markdown
```

The static report must remain `74/74` with zero definition drift. Tests lock definition quirks, permission/error UI, resolved username, operation/raw amount text, random color, guild avatar, signed/zero/uncapped creation arithmetic, scalar formatting, one-row updates, route isolation, flags, and usage ownership.

Disposable real-Mongo verification is operator-gated:

```bash
MHCAT_RUN_MONGO_INTEGRATION_TESTS=true \
MHCAT_MONGODB_URI='<disposable-uri>' \
go test ./internal/adapters/mongo/repositories \
  -run '^TestEconomyCoinAdminMongoIntegration'
```

It verifies decimal/null arithmetic and one-row duplicate mutation. The shared harness drops the generated database. Never use production.

## Staging Smoke

1. Stop Node `/代幣增加` ownership. Use a staging guild and disposable restored database.
2. Audit duplicate `{guild,member}` rows and BSON types/values for `coin` and `today`. Back up `coins`.
3. Enable both coin-admin flags, run preflight and guarded command-sync dry-run, and verify only the managed command change.
4. Test permission denial and missing-row reduce; verify exact public red embeds and no write.
5. Test missing-row add with ordinary, zero, negative, and over-limit amounts. Verify exact `coin`, numeric `today:1`, target username, random color, raw amount, and guild avatar.
6. Test existing positive/negative add/reduce around zero and `999999999`, plus decimal, null, and infinite fixtures. Verify operation-specific guards.
7. Seed duplicates, adjust once, and verify only one row changed. Run concurrent disposable adjustments and record overwrite behavior.
8. With usage enabled, verify one event for every success/error attempt. Confirm no index/collection creation and smoke enabled downstream balance/rank/profile consumers.

## Rollback

Disable coin-admin runtime and remove the managed staging command through guarded sync. Stop Go ownership before restoring Node. Audit all affected `{guild,member}` rows; duplicate natural-read order means the row shown before/after rollback may differ. Restore reviewed backup values only. Do not automatically clamp signed balances, merge duplicates, or create/drop indexes.

Production rollout remains gated on live Discord smoke, disposable real-Mongo execution, duplicate and scalar audits, concurrency/repair decisions, backup, and exclusive ownership.
