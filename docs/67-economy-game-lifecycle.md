# Economy Game Lifecycle

Status: transactional wagers and active-game timeout settlement are implemented behind the existing disabled-by-default economy game gates. Production ownership, duplicate cleanup, and restart recovery remain open.

## Legacy Reference

- Command and collectors: `MHCAT/slashCommands/代幣系統/game.js`
- Knowledge question bank: `MHCAT/topic.json`
- Balance model: `MHCAT/models/coin.js`

## Implemented Behavior

All three subcommands preserve the legacy invite, component IDs, and two-player wager model:

- `比大小` settles immediately after acceptance.
- `知識王` starts a per-question countdown at `20`; the legacy interval times out on the strict `time < 0` tick, 21 seconds after question start.
- `21點` starts each turn at `0`; the legacy interval times out on the strict `time > 30` tick, 31 seconds after turn start.
- Knowledge question UI retains the legacy relative timestamp 15 seconds after question start.
- Blackjack turn UI retains the legacy relative timestamp 30 seconds after turn start.

Knowledge timeout settlement matches the current round's response state. Each player who answered receives `2 * wager`; each player who did not answer receives `0`. If neither answered, the reserved pot is not returned. The message names every non-responder and removes all components.

Blackjack timeout settlement gives `2 * wager` to the player whose turn is not active and `0` to the timed-out player. The message names the timed-out player and removes all components.

## Concurrency And Failure Safety

State-changing components and timer callbacks claim one process-local session operation at a time. A turn generation prevents an older timer or busy-operation retry from replacing the next turn's timer. Successful terminal settlement deletes the session before the Discord edit, so an edit failure cannot settle the wager again.

Reserve and settlement update both `coins` balances in one Mongo transaction. A settlement error fails the in-memory session closed and is logged; application code does not blindly retry because an unknown transaction commit result could otherwise pay twice. Operators must inspect Mongo state before manual recovery.

The interaction dispatcher owns timer shutdown. Graceful app shutdown cancels pending timers and waits for callbacks before Discord and Mongo connections close.

## Remaining Limits

- Sessions and timer generations are process-local and do not survive restart.
- A restart after wager reserve but before terminal settlement still requires operator reconciliation.
- Multiple unresolved invites between the same two users in one channel retain ambiguous legacy button IDs until the first component binds a message.
- Duplicate `{guild,member}` rows remain audit debt even though matching duplicates are updated together.
- Node and Go do not share game ownership; they must not handle the same command rollout concurrently.

## Gates

```bash
MHCAT_FEATURE_ECONOMY_GAME_ENABLED=true
MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_GAME=true
```

Use a transaction-capable replica set or sharded Mongo deployment. Keep both flags disabled in production until exclusive ownership and reconciliation procedures are approved.

## Staging Checklist

1. Run `mhcat-staging-preflight` and review the economy-game safety warning.
2. Use an isolated guild and disposable `coins` rows on replica-set or sharded Mongo.
3. Run the duplicate audit for both player keys.
4. Stop Node ownership for the staging command.
5. Verify `比大小` reserves and settles both balances once.
6. In `知識王`, let neither player answer and confirm both wagers remain deducted and components disappear.
7. Repeat with one player answering and confirm only that player receives `2 * wager`.
8. In `21點`, let each player's turn expire separately and confirm the other player receives `2 * wager`.
9. Press a terminal action close to the deadline and confirm only one settlement occurs.
10. Gracefully stop the bot during an active game and confirm shutdown completes without a late Discord edit.

## Rollback

Disable command sync and the Go runtime before restoring Node ownership. Active process-local sessions cannot be transferred. Inspect every accepted but unfinished game and reconcile both balances as one pair; do not retry one side independently. No schema rollback is required.
