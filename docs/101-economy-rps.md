# Economy Rock Paper Scissors Parity Contract

Status: parity-audited against `slashCommands/代幣系統/rock_paper_scissors.js`, `models/coin.js`, and current Go definition, handler, service, scalar adapter, Mongo repository, app wiring, staging guards, integration harness, and race coverage. Live Discord and operator-gated real-Mongo smoke remain required before production rollout.

## Scope

This contract covers `/剪刀石頭布`, the single-player wager game against a random computer choice. It does not authorize two-player economy games, coin administration, reset, duplicate cleanup, or index creation.

## Definition And Ownership

The command preserves exact name/description, required integer wager, required string choice with `剪刀`/`石頭`/`布`, option descriptions, cooldown metadata `10`, emoji metadata, and documentation URL. It has no permission requirement. Legacy did not centrally enforce the cooldown; Go adds no local throttle.

Runtime is disabled by default behind `MHCAT_FEATURE_ECONOMY_RPS_ENABLED=true`. Staging command sync separately requires `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_RPS=true`; config, preflight, and scripts reject sync without runtime. Startup never syncs commands or creates Mongo objects.

Global slash middleware owns usage. With tracking enabled, invalid wager, missing/insufficient balance, and success each record exactly one best-effort event. App wiring prevents module double counting.

## Lifecycle And Errors

The command publicly defers, validates the wager before reading Mongo, selects one uniformly random computer choice, mutates the balance, and edits the original response.

Validation/error embeds are red with exact animated-no prefix. Visible text remains:

- above `999999999`: `最高代幣設定數只能是999999999`;
- zero/negative: `至少要大於1!!`;
- missing balance: `你沒有足夠的代幣進行此次遊玩!`;
- insufficient balance: `你沒有足夠的代幣進行此次遊玩` without the final exclamation mark.

Legacy attempted `ephemeral:true` while editing an already public defer, which Discord cannot apply. Go intentionally keeps these edits public.

## Result UI

Success is one random-color embed titled `<a:girl:983775481100914788> __**剪刀石頭布!**__`. The exact three-line description shows player emoji/choice, computer emoji/choice, and `你獲得了` for wins or `你失去了` for losses and ties, followed by the absolute wager loss/gain.

Emoji mapping is `✂️剪刀`, `🪨石頭`, and `🖐布`. Footer text is `剪刀石頭布! | MHCAT`; its icon is the caller's guild display avatar, including animated guild avatars, with user-avatar fallback. Allowed mentions are empty.

## Wager And Outcome

Wager must be `1..999999999`. Before the draw, legacy affordability is `coin - wager < 0`; a tie therefore remains affordable even though it loses only half.

Computer choice is uniform over the three values. A tie subtracts `parseInt(wager / 2)`, equivalent to floor for positive integer wagers. A loss subtracts the full wager. A win adds the full wager and intentionally has no post-win `999999999` cap.

## Scalars And Duplicates

The repository reads one arbitrary natural `{guild,member}` row and uses preserved Mongoose-visible `coin` number text. Decimal, null-as-zero, and positive/negative infinity retain JavaScript affordability and arithmetic. Updated BSON/result metadata preserve decimals and infinities. Missing, malformed, or NaN values fail before Go mutation rather than returning legacy success while a Mongoose cast likely fails.

One arbitrary matching row is updated with `$set`; extra duplicates remain unchanged. No unique index, transaction, or automatic deduplication is created. Read/compute/set remains susceptible to concurrent overwrite, matching the legacy shape more closely than an arithmetic transaction.

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

The static report must remain `74/74` with zero definition drift. Tests lock definition/options, wager errors, punctuation, all outcomes and tie floor, random injection, exact emoji/embed/footer/color, guild avatar, no-cap wins, scalar arithmetic, one-row duplicates, route isolation, flags, and usage ownership.

Disposable real-Mongo verification is operator-gated:

```bash
MHCAT_RUN_MONGO_INTEGRATION_TESTS=true \
MHCAT_MONGODB_URI='<disposable-uri>' \
go test ./internal/adapters/mongo/repositories \
  -run '^TestEconomyRPSMongoIntegration'
```

It verifies decimal/infinite arithmetic and one-row duplicate mutation. The shared harness drops the generated database. Never use production.

## Staging Smoke

1. Stop Node `/剪刀石頭布` ownership. Use a staging guild and disposable restored database.
2. Audit duplicate `{guild,member}` rows and BSON types/values for `coin`. Back up `coins`.
3. Enable both RPS flags, run preflight and guarded command-sync dry-run, and verify only the managed command change.
4. Test zero, negative, and over-limit wagers; verify exact public error titles and no mutation.
5. Test missing and insufficient balances; verify the intentional punctuation difference.
6. Force each player/computer outcome. Verify emojis, random color, guild avatar, full loss, floored tie loss, and no-cap win.
7. Test decimal, null, and infinite fixtures. Seed duplicates and verify one row changes. Run concurrent disposable plays and record overwrite behavior.
8. With usage enabled, verify one event per success/error attempt. Confirm no index/collection creation and smoke balance/rank/profile consumers.

## Rollback

Disable RPS runtime and remove the managed staging command through guarded sync. Stop Go ownership before restoring Node. Audit every affected `{guild,member}` row; duplicate natural-read order can expose a different balance after rollback. Restore reviewed backup values only. Do not automatically clamp wins, merge duplicates, normalize scalars, or create/drop indexes.

Production rollout remains gated on live Discord smoke, disposable real-Mongo execution, duplicate/scalar/concurrency audits, backup/repair procedure, and exclusive ownership.
