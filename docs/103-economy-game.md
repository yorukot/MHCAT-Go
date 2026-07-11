# Economy Coin Game Parity Contract

Status: parity-audited against `slashCommands/代幣系統/game.js`, `topic.json`, `models/coin.js`, and current Go definition, handlers, session/timer store, scalar adapter, transactional repository, app wiring, staging guards, integration harness, and race coverage. Live Discord and operator-gated real-Mongo smoke remain required before production rollout.

## Scope And Ownership

This contract covers `/代幣遊戲` subcommands `21點`, `知識王`, and `比大小`, their invites/tutorials/buttons, five-round knowledge flow, blackjack turns/cards, higher/lower draw, timeout/forfeit settlement, and two-player `coins` writes. Definition metadata, option text/order/requirements, cooldown metadata `10`, emoji, and legacy documentation URL are preserved. Legacy did not centrally enforce the cooldown; Go adds no local throttle.

Runtime is disabled by default behind `MHCAT_FEATURE_ECONOMY_GAME_ENABLED=true`. Staging guild sync separately requires `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_GAME=true`; preflight and scripts reject unpaired or non-staging sync. Global slash middleware records exactly one best-effort usage event for every slash attempt. Buttons and scheduled transitions record none.

## Invite And UI Contract

All branches defer publicly, check both balances, then send the exact random-color public invite with `yesssss`/`nooooo`. Blackjack adds `teach21point`; higher/lower adds `thansize`; both private tutorials remain available after rejection while accept/reject become disabled. Knowledge and blackjack preserve private acceptance feedback. Higher/lower immediately replaces the invite with the public draw GIF and sends no extra acceptance reply.

Only the invited user can accept. The challenger receives the exact private denial. Self-invites still show the normal public invite, but the sole user's accept click hits that challenger denial and reserves nothing. Either participant may reject. Invites expire silently after 30 seconds.

- `比大小` draws each value inclusively from `0` through `100`, shows the legacy text/GIF for five seconds with no components, then displays both values and settlement.
- `知識王` uses the bundled legacy topic bank and answer custom IDs. Its first question appears after 500 milliseconds. Both answers trigger a component-free five-second reveal with choices, scores, correct answer, marker quirks, and remaining count. The countdown runs during the reveal; later questions expose about 15 seconds and at most about 750 points. The strict `time < 0` timeout occurs on the 21st tick.
- `21點` preserves tutorial text, 40-card weighted deck shape, card emoji/formatting, private card view, player-specific action IDs, random colors, first-turn text, pink-arrow later-turn history, and private action feedback. The strict `time > 30` timeout occurs on the 31st tick.

Knowledge timeout pays `2 * wager` only to players who answered the current round. Neither answer burns the reserved pot. Blackjack timeout pays `2 * wager` to the player whose turn is not active. Terminal messages remove components.

## Wager And Mongo Contract

Knowledge rejects every negative wager. The legacy `< -1` guards for blackjack and higher/lower allow exactly `-1`; reserve adds one coin to both players, ties net to zero, and a winner return of `-2` leaves the winner down one and loser up one. Values below `-1` return the exact wager error. Zero remains valid.

Affordability and arithmetic preserve Mongoose-visible decimal, null-as-zero, and positive/negative infinity values. Missing balances use the player-specific legacy error. Missing/malformed/NaN scalar values fail closed rather than allowing a likely asynchronous Mongoose cast failure.

Each balance read is one arbitrary `{guild,member}` document and each write uses `UpdateOne` on that same logical filter. With duplicates, the row read and row written are independently arbitrary; one duplicate changes and the others remain. No index, deduplication, or normalization is created.

Go intentionally rechecks balances at acceptance and wraps both reserve writes or both settlement writes in one Mongo transaction. Legacy used stale invite-time documents and unawaited independent writes. A replica set or sharded deployment is therefore required. Transaction failure rolls back the pair; unknown commit results are not blindly retried.

## Session Safety Differences

Sessions, phases, generations, and timers are process-local. Serialized claims prevent button/timer/transition double settlement, stale callbacks cannot replace a newer phase, terminal settlement removes the session before Discord editing, and graceful shutdown cancels callbacks. These are reliability fixes over overlapping legacy collectors.

Restart after reserve loses the session and requires manual pair reconciliation. Multiple unresolved invites between the same users retain ambiguous legacy custom IDs until a component binds a message. Node and Go must never own this command concurrently.

## Verification

```bash
go test ./internal/core/domain ./internal/core/services/economy \
  ./internal/testutil/fakemongo ./internal/adapters/mongo/repositories \
  ./internal/discord/features/economy ./internal/app
go test -race ./internal/core/services/economy \
  ./internal/adapters/mongo/repositories ./internal/discord/features/economy ./internal/app
go vet ./...
go run ./tools/parity-audit --legacy-root ../MHCAT --format markdown
```

Operator-gated scalar, duplicate, rollback, and concurrency evidence requires a disposable transaction-capable Mongo URI:

```bash
MHCAT_RUN_MONGO_TRANSACTION_INTEGRATION_TESTS=true \
MHCAT_MONGODB_URI='<disposable-uri>' \
go test ./internal/adapters/mongo/repositories \
  -run '^TestEconomyCoinGameMongoTransactionIntegration'
```

The generated database is dropped. Never use production.

## Staging And Rollback

1. Stop Node ownership. Back up and audit every player row, scalar type, and duplicate `_id`; use only disposable balances.
2. Use replica-set/sharded Mongo, enable paired runtime/sync flags, run preflight and sync dry-run, then verify exact invite/reject/tutorial/self-invite UI.
3. Exercise positive, zero, `-1`, decimal, null, infinity, malformed, and duplicate fixtures across all games; verify one usage per slash and none per component.
4. Verify draw/reveal/turn timing, both timeout directions, terminal races, transaction rollback, unknown outcomes, and graceful shutdown.

Rollback by disabling sync/runtime before restoring Node. Inspect every accepted unfinished session and reconcile both players from the reviewed backup as one pair. Do not retry one side, multiply rounded values, merge duplicates, or create an index during rollback.

Production remains gated on live Discord smoke, disposable Mongo execution, backup/restore rehearsal, duplicate/scalar audit, restart reconciliation, exclusive ownership, and explicit economy-write approval.
