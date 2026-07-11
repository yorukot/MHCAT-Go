# Economy Coin Reset Parity Contract

Status: parity-audited against `slashCommands/代幣系統/coin_rest.js`, `models/coin.js`, and current Go definition, slash/event handlers, confirmation store, scalar adapter, Mongo repository, app/event wiring, staging guards, integration harness, and race coverage. Live Discord and operator-gated real-Mongo smoke remain required before production rollout.

## Scope And Ownership

This contract covers destructive `/代幣重製` and its same-channel `^確認^` message. It preserves exact definition metadata, optional integer `除以多少`, guild-owner policy, cooldown metadata `100`, emoji, and documentation URL. Legacy did not centrally enforce the cooldown; Go adds no local throttle.

Runtime is disabled by default behind `MHCAT_FEATURE_ECONOMY_COIN_RESET_ENABLED=true`. Staging sync separately requires `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_RESET=true`. Confirmation additionally requires gateway, Guild Messages intent, and Message Content intent. Preflight/scripts enforce these prerequisites; startup creates no command, collection, or index.

Global slash middleware owns usage. Owner and non-owner slash attempts each record exactly one best-effort event. Confirmation messages add none.

## Slash And Confirmation UI

A non-owner receives the exact ephemeral red `只有服主可以使用這個指令!` embed. Owner lookup failure receives a controlled ephemeral unknown-error embed. An owner receives the exact public warning:

```text
:warning: | 一但重製，___**將無法復原**___，如確定要還原請於60秒內輸入`^確認^`(只有一次機會)!!!
```

One pending confirmation is stored for the guild/channel/user for 60 seconds. The first matching nonbot message consumes it. Exact `^確認^` executes; other content replies with the exact red cancellation embed. Expiry is silent.

Go intentionally keeps one pending entry per owner/channel, replacing an older request, instead of allowing multiple legacy collectors to consume one message and launch multiple resets. It validates guild/channel/user identity and ignores bot messages.

## Reset Behavior

Omitted or numeric-zero divisor deletes every guild `coins` document, including duplicates. Any nonzero integer, including negative, divides each cursor row's Mongoose-visible number and applies JavaScript `Math.round`.

Success replies with green `<:trashbin:995991389043163257>成功重製伺服器內所有代幣`. Wrong confirmation and repository errors use the exact animated-no red prefix. Go reports `這伺服器沒有任何的代幣喔!` for an empty guild; legacy's malformed empty-array test produced no reply, an intentional reliability fix.

Division preserves decimal, null-as-zero, positive/negative infinity, and negative-half rounding toward positive infinity. Missing, malformed, or NaN coin values fail when encountered instead of scheduling a likely failed Mongoose cast.

Legacy iterates every fetched row but divides with `updateOne({guild,member})`. Go preserves that shape: each cursor row computes from its own original value and updates one arbitrary matching duplicate. One duplicate may be rewritten repeatedly while others remain untouched. Affected count means cursor rows processed.

Delete mode removes all guild rows. Divide mode is sequential and nontransactional; later failures can leave earlier changes. No unique index, transaction, deduplication, or rollback snapshot is created.

## Verification

```bash
go test ./internal/core/domain ./internal/core/services/economy \
  ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories \
  ./internal/discord/features/economy ./internal/discord/events \
  ./internal/app ./internal/config ./cmd/mhcat-staging-preflight

go test -race ./internal/core/services/economy \
  ./internal/adapters/mongo/repositories \
  ./internal/discord/features/economy ./internal/app

go vet ./...
go run ./tools/parity-audit --legacy-root ../MHCAT --format markdown
```

Tests lock owner denial, warning/wrong/success/error UI, same-channel/owner/one-shot/expiry confirmation, delete/divide/negative rounding, scalar behavior, duplicate iteration, dependency gates, and slash-only usage ownership. The static report must remain `74/74` with zero drift.

Disposable real-Mongo verification is operator-gated:

```bash
MHCAT_RUN_MONGO_INTEGRATION_TESTS=true \
MHCAT_MONGODB_URI='<disposable-uri>' \
go test ./internal/adapters/mongo/repositories \
  -run '^TestEconomyCoinResetMongoIntegration'
```

It verifies decimal/null/infinite division and duplicate cursor/update-one behavior. The generated database is dropped. Never use production.

## Staging And Rollback

1. Stop Node command/confirmation ownership. Audit and back up every guild `coins` row, including duplicates/scalars, in disposable staging.
2. Enable runtime/sync/gateway/intents, run preflight and guarded sync dry-run, then test non-owner, wrong/cross-user/cross-channel/bot/expired confirmations.
3. Confirm delete mode preserves another guild. Reseed and test positive/negative divisors with ordinary, decimal, null, infinite, malformed, and duplicate fixtures.
4. Verify exact UI, one usage event per slash and none per message, partial failures, no Mongo object creation, and downstream reads.

Rollback by disabling runtime/sync/event ownership before restoring Node, then restoring `coins` from reviewed backup. Divide/delete cannot be inverted reliably because of rounding and duplicates. Never automatically multiply, merge, normalize, or create/drop indexes.

Production remains gated on live Discord smoke, disposable Mongo execution, backup/restore rehearsal, duplicate/scalar audit, event ownership, and explicit destructive-operation approval.
