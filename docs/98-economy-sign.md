# Economy Sign Parity Contract

Status: parity-audited against `slashCommands/代幣系統/sing.js`, `sing_list.js`, the `sing` branch in `events/btn.js`, legacy assets/fonts/Mongoose behavior, and current Go definitions, services, renderer, Mongo adapter, component router, app wiring, staging guards, and race coverage. Live Discord and operator-gated real-Mongo smoke remain required before production rollout.

## Scope

This contract covers `/簽到`, `/簽到列表`, and sign calendar navigation. It does not authorize other economy writes, unique indexes, duplicate cleanup, or daily-reset ownership.

## Definition And Ownership

Both guild commands preserve their exact names, descriptions, documentation URL, and no default permission. Legacy cooldown metadata is 60 seconds for `/簽到` and 10 seconds for `/簽到列表`; neither runtime centrally enforces those metadata cooldowns.

Runtime is disabled by default behind `MHCAT_FEATURE_ECONOMY_SIGNIN_ENABLED=true`. Staging command sync separately requires `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SIGNIN=true`; config, preflight, and scripts reject sync without runtime. Startup never syncs commands or creates Mongo objects.

Global slash middleware owns usage. With tracking enabled, each slash attempt adds exactly one best-effort event. Calendar components add none; app wiring and tests prevent module-level double counting.

## Sign Lifecycle

`/簽到` publicly replies with the exact orange loading embed: legacy author, loading GIF, footer, caller avatar forced to static PNG with fallback, and color `#FF5809`. It then edits the original response to only `sign.png` and four secondary emoji buttons.

The image status is exactly one of:

- `🎉 | 今天有準時簽到很棒喔! 明天也要記得來簽到.w.`;
- `⚠ | 你今天已經簽到過了!請於隔天(0:00)後再來簽到!` in daily mode;
- `⚠ | 你今天已經簽到過了喔!` in rolling mode;
- `⚠ | 不可以加超過\`999999999\`!!` for an existing-balance cap failure.

Duplicate and cap outcomes still render the calendar card. Other repository/render/Discord errors propagate through controlled runtime handling instead of leaving a silently raced callback tree.

## Calendar UI

`sign.png` is exactly 1000x707. It uses legacy `asset/background.png`, the radius-5/sigma-5 separable blur, 50% black overlay, `asset/mhcat_white.png` at `(20,35)`, and target avatar at `(900,35)` after the legacy 128x128 radius-40 mask and 80x80 scale. Failed avatar fetches use `asset/yellow_discord.png`.

The card preserves the Comic Sans month, weekday, and day sizes; localized username font/size and right alignment; cyan month, yellow weekday, green weekday/red weekend numbers; exact grid geometry; `asset/verify_icon.png` positions; centered 30px status; and six-week month rows. Go intentionally keeps the correct `Fri.` label on redraw rather than the legacy `Fir.` typo.

The four buttons navigate previous year, previous month, next month, and next year with the exact legacy emojis. New cards emit bounded versioned IDs under `mhcat:v1:economy:sign_page`; exact legacy `/<17-20 digit user>_sing{YYYY}-[M]` IDs remain routable during rollback migration. Navigation first replaces the old image/buttons with the loading embed and clears attachments, then renders the embedded target user's name/avatar and requested month. It records no usage event.

## Sign List UI

`/簽到列表` publicly defers, then edits to one random-color embed titled `簽到人數資訊` plus UTF-8 `discord.txt`. The description preserves exact count/self-status emoji text, inline `┃ username#discriminator ┃` names below 100 rows, and the exact file-only notice at 100. Migrated accounts retain `#0`; missing members render `使用者已消失!`.

The export preserves repository order. Daily rows are `tag(id:user)`; rolling rows append `簽到時間:YYYY/MM/DD HH:mm:ss [台北標準時間]`. Empty lists still attach an empty file and render the legacy empty inline delimiter.

## Mode And Scalar Semantics

A missing `gift_changes` row selects daily mode. An existing row selects daily mode only when hydrated numeric `time` is exactly zero. Missing/null/NaN `time` on an existing row selects rolling mode with 86400 seconds; fractional, negative, and infinite numbers retain JavaScript truthiness and comparison behavior.

Current epoch seconds use `Math.round(Date.now()/1000)` compatibility. `/簽到列表` daily membership uses numeric strict `coin.today === 1`. Rolling membership preserves Mongoose-visible null, missing, decimal, negative, and infinite values and requires both `now - today < cooldown` and `> 0`.

`sign_coin` is 25 only when the config row is absent. Numeric configured values retain decimals and infinities; null is numeric zero. Undefined/NaN rewards fail before Go writes. Existing balances enforce the legacy `999999999` cap, while first-time creation remains uncapped. Explicit `coin:null` adds from zero; missing/nonnumeric coin fails closed rather than reproducing a partial legacy callback write.

Legacy first-user marker behavior is asymmetric: no config stores `today:1`; any config row stores rounded epoch seconds, even when `time:0`. Existing daily-mode users store `1`; existing rolling users store rounded epoch seconds.

## Mongo Writes And Intentional Safety

Sign-in touches only `coins` and `sign_lists`. Existing coin mutation is one conditional aggregation update, so eligibility, cap, null handling, reward addition, and marker replacement are atomic for the selected row. First-user creation is separately guarded. Calendar mutation runs only after successful coin mutation and uses `$addToSet` at `date.<year>.<month>`.

This intentionally does not reproduce legacy non-awaited behavior where coin/today/calendar writes could succeed independently, duplicate days could accumulate, or a cap/duplicate/error attempt could still alter calendar data. Go also does not auto-deduplicate existing rows: natural first-match update behavior remains, and duplicate `coins` or `sign_lists` rows stay observable.

No transaction spans coin and calendar writes. A coin success followed by calendar failure remains possible and requires audit/repair. Startup creates no collection or index. Production requires reviewed duplicate handling for `{guild,member}` in both collections before any unique index is applied through the explicit offline process.

## Daily Reset

Daily mode requires exactly one Asia/Taipei midnight reset owner. The one-shot and recurring Go paths share lease `daily-reset`; Node `handler/cron.js` and Go reset ownership must never overlap. Rolling mode does not use the midnight marker reset.

## Verification

```bash
go test ./internal/adapters/mongo/documents \
  ./internal/adapters/mongo/repositories \
  ./internal/core/services/economy \
  ./internal/discord/features/economy \
  ./internal/discord/responses ./internal/adapters/discordgo \
  ./internal/discord/customid ./internal/discord/interactions \
  ./internal/app ./internal/config ./cmd/mhcat-staging-preflight

go test -race ./internal/adapters/mongo/repositories \
  ./internal/core/services/economy \
  ./internal/discord/features/economy ./internal/app

go vet ./...
go run ./tools/parity-audit --legacy-root ../MHCAT --format markdown
```

The static report must remain `74/74` with zero definition drift. Tests lock loading/final lifecycles, canvas pixels/assets/fonts/grid/avatar, statuses, old/new IDs, attachment clearing, list embed/export, mode and scalar boundaries, filters/update pipeline, and usage ownership.

Disposable real-Mongo verification is operator-gated:

```bash
MHCAT_RUN_MONGO_INTEGRATION_TESTS=true \
MHCAT_MONGODB_URI='<disposable-uri>' \
go test ./internal/adapters/mongo/repositories \
  -run '^TestEconomySignInMongoIntegration'
```

It verifies no startup mutation, null/decimal writes, first-user markers, calendar writes, and duplicate rows. Cleanup drops the generated database. Never use production.

## Staging Smoke

1. Stop Node command/component and reset ownership. Use a staging guild and disposable restored database.
2. Run the read-only Mongo audit. Review duplicate `coins`/`sign_lists`, duplicate `gift_changes`, and scalar types for `coin`, `today`, `sign_coin`, and `time`.
3. Enable both sign flags, run preflight and guarded command-sync dry-run, and review only the two managed commands.
4. Test first/existing/duplicate/cap outcomes in missing-config daily, `time:0` daily, and positive rolling modes. Verify exact card, status, target avatar, icons, and buttons.
5. Navigate year/month boundaries using newly emitted and retained legacy IDs. Verify immediate loading replacement clears the old attachment/components and redraws the embedded target.
6. Test `/簽到列表` at 0, 1, 99, and 100 rows, self present/absent, missing members, `#0`, and rolling timestamps. Compare exact `discord.txt` bytes.
7. With usage enabled, verify one event per slash and none per component. Compare collections/indexes before/after and confirm only expected row fields changed.
8. Run one exclusive daily reset in daily mode. Force coin/calendar/Discord failures and record any partial coin-success/calendar-failure row for repair review.

## Rollback

Disable sign runtime and remove the two managed staging commands through guarded sync. Stop Go component ownership before restoring Node so both do not handle retained legacy IDs. Stop the Go daily-reset owner before restoring Node cron. No automatic data rollback is safe: audit `coins` and `sign_lists`, restore only reviewed backup rows, and do not drop indexes or normalize scalars without a separate approved plan.

Production rollout remains gated on live Discord smoke, disposable real-Mongo execution, duplicate/index decisions, calendar repair procedure, and exclusive command/component/reset ownership.
