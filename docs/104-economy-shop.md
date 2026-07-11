# Economy Shop Parity Contract

Status: parity-audited against `slashCommands/代幣系統/ghp_shop.js`, `models/ghp.js`, `models/coin.js`, and current Go definition, handlers, session store, scalar adapters, repository, role/DM ports, app wiring, staging guards, guarded Mongo harness, and race coverage. Live Discord and operator-gated real-Mongo smoke remain required before production rollout.

## Scope And Ownership

This contract covers `/代幣商店 商品增加`, `商品刪除`, and `商品查詢`, plus item detail, quantity keypad, purchase, role assignment, prize-code DM, `ghps` inventory, and `coins` debit. Exact definition metadata, option order/text/requirements, advertised permission text, cooldown metadata `5`, emoji, and docs paths are preserved. Legacy did not centrally enforce cooldowns; Go adds no local throttle.

Runtime is disabled by default behind `MHCAT_FEATURE_ECONOMY_SHOP_ENABLED=true`. Staging guild sync separately requires `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SHOP=true`; preflight/scripts reject unpaired or non-staging sync. Global slash middleware records exactly one best-effort usage event for every slash attempt. Buttons record none.

## Admin Commands

Add/delete defer publicly and enforce Manage Messages at runtime, matching the advertised mixed policy where query/purchase remain public. Add preserves JavaScript UTF-16 name length (`<=15`), positive integer price, explicit-zero/missing count defaulting to `1`, negative count rejection, role hierarchy checks, `Date.now()` millisecond IDs, 25-item cap, duplicate-ID check, nullable code/role, and exact success/error embeds and docs links.

Delete finds and removes one arbitrary matching `{guild,commodity_id}` row and emits the exact store success/error UI. Neither path creates an index or normalizes existing rows.

## Browse And Purchase UI

Query preserves Mongo cursor order, up to 25 inline fields, five rows of five blue item buttons, random embed color, guild name, requester tag/global-avatar footer, typo `按扭`, exact emojis/text, and ten-minute requester/message ownership. The first valid item click consumes the browse collector and opens one random-color detail with the green purchase button. Other users receive the exact private denial.

Purchase opens the four-row legacy digit keypad (`1`-`9`, backspace, `0`, confirm), preserves empty backspace rendering and random colors, rejects quantities above current stock, and limits a cached existing role item to one. A missing cached role allows multiple purchases as legacy does. The nested collector shares the original ten-minute deadline and process-local state; restart leaves old controls inert.

Success removes components and preserves the exact item/quantity embed. Existing roles are added best-effort. Non-null prize codes produce the exact best-effort private DM. Missing item, zero/empty quantity, insufficient balance, stock overflow, requester mismatch, and backend failures preserve their legacy response class and visible text.

## Scalar And Duplicate Contract

`need_coin` is a legacy Mongoose String and `commodity_count`, `commodity_id`, and `coin` are Mongoose Numbers. Reads preserve decimal, null-as-zero, positive/negative infinity, numeric strings, and JavaScript number formatting in list/detail and arithmetic. Malformed/NaN price, stock, or balance fails closed instead of allowing likely asynchronous cast/write failure.

Slash-created prices/counts remain positive integers. Existing null or negative prices retain legacy arithmetic, including free or balance-increasing purchases. Decimal stock decrements fractionally; null stock rejects every positive quantity; infinite stock stays infinite.

Generated commodity IDs are positive integer milliseconds. An anomalous non-integer ID remains visible exactly in list fields/buttons but its click fails closed rather than truncating into a different item. This is an intentional safety difference; audit/repair such rows before rollout.

Balance lookup reads one arbitrary duplicate and debit uses one independently arbitrary `UpdateOne`; other duplicates remain unchanged. Item lookup is also arbitrary. For auto-delete stock exactly `1`, the selected `_id` is deleted and the legacy logical `UpdateOne` still runs, potentially decrementing another duplicate. Other auto-delete stock performs one logical update. `auto_delete=false` leaves stock unchanged.

## Failure And Ordering Differences

Legacy launched role, DM, inventory, and coin operations without awaiting them. Go checks inventory and coin writes sequentially, then performs role/DM best-effort. The write pair remains nontransactional: inventory can change before a coin failure, and no automatic retry or compensation occurs. This preserves the operational risk without exposing success before checked database writes.

Sessions are process-local and isolated per requester/message, matching the effective ownership of each legacy command-local collector. No distributed purchase lock exists; concurrent purchases can still race stale stock/balance reads.

## Verification

```bash
go test ./internal/core/domain ./internal/adapters/mongo/documents \
  ./internal/core/services/economy ./internal/testutil/fakemongo \
  ./internal/adapters/mongo/repositories ./internal/discord/features/economy ./internal/app
go test -race ./internal/core/domain ./internal/core/services/economy \
  ./internal/adapters/mongo/repositories ./internal/discord/features/economy ./internal/app
go vet ./...
go run ./tools/parity-audit --legacy-root ../MHCAT --format markdown
```

Operator-gated scalar and duplicate evidence uses a generated disposable database:

```bash
MHCAT_RUN_MONGO_INTEGRATION_TESTS=true \
MHCAT_MONGODB_URI='<disposable-uri>' \
go test ./internal/adapters/mongo/repositories \
  -run '^TestEconomyShopMongoIntegration'
```

The harness drops its database. Never use production.

## Staging And Rollback

1. Stop Node shop ownership. Back up/audit every `ghps` and affected `coins` row by `_id`, including duplicates, scalar types, role IDs, and code sensitivity.
2. Enable paired staging flags only, run preflight/sync dry-run, and use disposable channels, roles, codes, inventory, and balances.
3. Verify exact add/delete/list/detail/keypad/error/success UI, requester isolation, expiry/restart, role hierarchy/missing-role behavior, code DM, and one usage per slash.
4. Exercise decimal/null/infinity/malformed prices, stock, balances, duplicate item/balance rows, concurrent purchase, inventory-success/coin-failure, and side-effect failures.

Rollback by disabling sync/runtime before restoring Node. Restore `ghps` and `coins` together from reviewed `_id` backups; inspect role assignments and code DMs separately. Never infer stock from coin debit, merge duplicates, reissue codes automatically, or create an index during rollback.

Production remains gated on live Discord smoke, disposable Mongo execution, backup/restore rehearsal, scalar/duplicate/code audit, concurrency policy, exclusive ownership, and explicit economy-write approval.
