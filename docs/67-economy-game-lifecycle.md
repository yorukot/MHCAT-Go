# Economy Game Lifecycle

Status: transactional wagers, legacy transition UI/timing, and active-game timeout settlement are implemented behind the existing disabled-by-default economy game gates. The canonical parity and rollout contract is [103-economy-game.md](103-economy-game.md); this file remains the detailed lifecycle checklist.

## Legacy Reference

- Command and collectors: `MHCAT/slashCommands/代幣系統/game.js`
- Knowledge question bank: `MHCAT/topic.json`
- Balance model: `MHCAT/models/coin.js`

## Implemented Behavior

All three subcommands preserve the legacy random-color invite, component IDs, and two-player wager model. Knowledge and blackjack preserve their private acceptance feedback; higher/lower immediately replaces the invite with its public draw message and sends no extra acceptance reply. A rejected invite disables only accept/reject; the blackjack and higher/lower tutorial buttons remain available and render their full legacy private embeds.

Wager guards remain branch-specific. `知識王` rejects every negative wager, while the legacy `< -1` guards in `21點` and `比大小` allow exactly `-1`. Reserving that wager adds one coin to each player; a tie returns `-1` to each and nets to zero, while a `2 * wager` winner return removes two coins, leaving the winner down one and the loser up one. Values below `-1` still receive the exact `賭注必須大於-1` error.

Self-invites preserve the legacy lifecycle: the normal public invite mentions and names the same user as challenger and opponent, but that user's accept click receives the private `你不是被邀請者，無法選擇接受!` denial because the challenger guard runs first. No wager is reserved, and the invite expires silently.

- `比大小` removes the invite components, displays the legacy random-draw text/GIF for five seconds, and only then settles and displays both numbers.
- `知識王` sends the legacy ephemeral acceptance embed and displays the first question after 500 milliseconds.
- When both knowledge players answer, the game removes all components and displays both choices, both scores, the correct answer, marked answer options, and the remaining-question count for five seconds before the next question or final result.
- `知識王` resets its per-question countdown to `20` when the reveal starts. The countdown continues during the five-second reveal, so a newly displayed later-round question begins at about `750` available points and has about 15 visible seconds remaining. The legacy interval still times out on the strict `time < 0` tick, 21 seconds after that reveal-time reset.
- `21點` starts each turn at `0`; the legacy interval times out on the strict `time > 30` tick, 31 seconds after turn start.
- `21點` sends the legacy ephemeral `成功接受!!` feedback while updating the accepted invite to the first turn.
- Later blackjack turns show the pink-arrow turn banner, previous player's `抽牌`/`略過` choice, legacy private action feedback colors, and the correct player-specific button row. Private card views retain the legacy comma-space formatting.
- Knowledge question UI retains the legacy relative timestamp 15 seconds after the question is displayed.
- Blackjack turn UI retains the legacy relative timestamp 30 seconds after turn start.

Knowledge timeout settlement matches the current round's response state. Each player who answered receives `2 * wager`; each player who did not answer receives `0`. If neither answered, the reserved pot is not returned. The message names every non-responder and removes all components.

Blackjack timeout settlement gives `2 * wager` to the player whose turn is not active and `0` to the timed-out player. The message names the timed-out player and removes all components.

## Concurrency And Failure Safety

State-changing components and scheduled callbacks claim one process-local session operation at a time. Explicit phases distinguish higher/lower drawing, knowledge startup, knowledge question, knowledge reveal, and blackjack turn state. A generation prevents an older transition, timeout, or busy-operation retry from replacing the next scheduled event. Successful terminal settlement deletes the session before the Discord edit, so an edit failure cannot settle the wager again.

Reserve and settlement update both `coins` balances in one Mongo transaction. A settlement error fails the in-memory session closed and is logged; application code does not blindly retry because an unknown transaction commit result could otherwise pay twice. Operators must inspect Mongo state before manual recovery.

The interaction dispatcher owns timer shutdown. Graceful app shutdown cancels pending timers and waits for callbacks before Discord and Mongo connections close.

## Remaining Limits

- Sessions and timer generations are process-local and do not survive restart.
- A restart after wager reserve but before terminal settlement still requires operator reconciliation.
- Multiple unresolved invites between the same two users in one channel retain ambiguous legacy button IDs until the first component binds a message.
- Duplicate `{guild,member}` rows remain audit debt. One arbitrary row is read and one independently arbitrary matching row is updated, preserving legacy ambiguity.
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
5. Verify `比大小` shows the draw text/GIF for five seconds, keeps both balances reserved during that delay, and then settles both balances once.
6. Verify `知識王` sends the private acceptance embed, waits 500 milliseconds before the first question, and shows a component-free five-second reveal after both answers.
7. Confirm the next knowledge question appears after the reveal with about 15 seconds and `750` maximum points remaining.
8. Let neither knowledge player answer and confirm both wagers remain deducted and components disappear.
9. Repeat with one player answering and confirm only that player receives `2 * wager`.
10. Reject blackjack and higher/lower invites and confirm accept/reject are disabled while the tutorial button still opens the full private legacy guide.
11. In `21點`, confirm the private acceptance response and pink-arrow action history on both turns, then let each player's turn expire separately and confirm the other player receives `2 * wager`.
12. In disposable fixtures, submit `-1` to `21點` and `比大小`, verify the invite and signed reserve/settlement result, then verify `知識王 -1` and every game at `-2` return the wager error.
13. Invite the challenger themselves, verify the normal public invite, then verify their accept click is denied without mutating the balance.
14. Press a terminal action close to the deadline and confirm only one settlement occurs.
15. Gracefully stop the bot during a pending transition or active game and confirm shutdown completes without a late Discord edit.

## Rollback

Disable command sync and the Go runtime before restoring Node ownership. Active process-local sessions cannot be transferred. Inspect every accepted but unfinished game and reconcile both balances as one pair; do not retry one side independently. No schema rollback is required.
