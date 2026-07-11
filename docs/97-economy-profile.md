# Economy Profile Parity Contract

Status: parity-audited against `slashCommands/代幣系統/user-info.js`, the `my-profile` branch in `events/btn.js`, legacy assets/fonts/Mongoose behavior, and current Go definition, service, renderer, Mongo adapter, component router, app wiring, staging guards, and race coverage. Live Discord and operator-gated real-Mongo smoke remain required before production rollout.

## Scope

This contract covers `my-profile` / localized `我的檔案`, optional user selection, and `<user>my-profile` refresh. It does not enable the separately disabled `/聊天經驗` or `/語音經驗` profile commands.

## Definition And Ownership

The command retains exact English base name/description, zh-TW/zh-CN/en-US/en-GB name and description localizations, optional user option `user` with all legacy localizations, no default permission, cooldown metadata `10`, and documentation URL `https://docsmhcat.yorukot.me/docs/snig`. Legacy did not centrally enforce the cooldown; Go adds no local throttle.

Runtime is disabled by default behind `MHCAT_FEATURE_ECONOMY_PROFILE_ENABLED=true`. Staging guild sync separately requires `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_PROFILE=true`; config, preflight, and scripts reject sync without runtime. Startup never syncs commands.

Global slash middleware owns usage. With tracking enabled, one best-effort `all_use_counts` event is recorded per slash attempt before handler work. Refresh components add none, and app wiring removes the module tracker to prevent double counting.

## Slash Lifecycle

The public slash path immediately replies with the orange loading embed: exact author/loading GIF, footer text, caller avatar forced to static Discord CDN PNG with fallback, and `#FF5809`. It resolves the optional member (or caller), reads and renders, then edits the original response to `user-info.png` plus one green `更新` button with emoji `<:update:1020532095212335235>` and ID `<target>my-profile`.

Member nickname takes precedence over username. The header always appends ` #<discriminator>` when Discord supplies it, including migrated `#0`, matching the legacy template. The display string uses legacy UTF-16 width truncation; guild name is not truncated by application code.

## Refresh Lifecycle

Only exact 17-20 digit `<user>my-profile` IDs route. A missing member returns the exact ephemeral red `很抱歉，這位使用者已退出該伺服器!!` response before changing the message.

For an existing member, refresh first updates the existing message to the loading embed using the clicker's static PNG avatar, explicitly clears the old attachment and components, then edits the original response with the refreshed PNG/button. Any member may refresh a public card; the embedded target remains the queried user. Exact parsing intentionally replaces legacy substring matching to prevent ID collisions.

## Canvas UI

The card is exactly 1500x750 over `asset/background_profile.png` and attached as `user-info.png`. It preserves:

- target avatar requested as PNG, drawn to a rounded 128x128 radius-40 intermediate, then scaled to 98x98 at `(42,30)`; `asset/yellow_discord.png` fallback;
- 45px localized member heading, 25px full guild name, and 40px Comic Sans creation/join dates;
- TC/SC/JP/HK/Noto/Bengali/Arabic/emoji fallback fonts and Oswald 30px XP progress text;
- text/voice progress bars at legacy coordinates, colors, 35px height, rounded radius, and percentage width formula;
- centered 40px text/voice/coin ranks, XP, levels, coin/sign, work, economy settings, and XP multiplier values at legacy coordinates;
- exact missing-data labels: `沒有資料`, `沒有資料!`, `無資料`, zero, `待業中`, and their legacy `#` prefixes.

Failed avatar HTTP fetches use fallback art. Fetches honor context, time out after two seconds, and are bounded, an intentional reliability improvement over rejected legacy image promises.

## Mongo Reads And Scalars

The card reads, in legacy nesting order semantics, first matching rows plus guild lists from:

1. `work_sets` by guild;
2. `work_users` by guild/user;
3. `gift_changes` by guild;
4. `voice_xps` and `text_xps` by guild/member plus guild lists;
5. `coins` by guild/member plus guild list.

Duplicate singleton/member rows remain allowed. `findOne`-equivalent reads select an arbitrary natural match; list rows remain natural `_id` order and duplicates participate separately in ranks.

Read-only display metadata preserves Mongoose-visible null, missing/malformed, decimal, `Infinity`, and `-Infinity` states for coin/today, economy settings/time, work settings/energy/end time, and XP/level. Operational integer/float fields remain separate for write consumers.

JavaScript behavior retained includes `Number`, `parseInt` truncation, `nFormatter` K/M/G thresholds and half-up one-decimal rounding, `InfinityG`, `NaN`, tie-last rank counting, malformed-comparison inclusion, strict numeric `today === 1/0`, truthy negative/infinite cooldowns, and `Math.round(Date.now()/1000)` status time. Infinite XP levels would make the legacy decrement loop nonterminating; Go fails that rank calculation closed as `NaN` rather than hanging the interaction.

## Mongo Writes And Migration

Profile slash and refresh paths write no economy/profile collection. Construction creates no collection, index, schema, migration, or backfill. No migration is required.

Before unique indexes or duplicate cleanup on any six collections, operators must audit which natural first row is currently visible and how duplicate list rows affect ranks. Such cleanup can visibly change cards and is outside this contract. Bot startup never applies the catalog's offline index plan.

## Errors

Legacy callbacks ignored Mongo errors and asynchronous image failures could strand the loading response. Go propagates backend/render/Discord failures through controlled runtime handling, bounds media fetches, and can resolve uncached members through the Discord provider. Successful and true missing-member UI remains legacy-compatible without exposing internal errors.

## Verification

```bash
go test ./internal/adapters/mongo/documents \
  ./internal/adapters/mongo/repositories \
  ./internal/core/services/economy \
  ./internal/discord/features/economy \
  ./internal/discord/responses ./internal/adapters/discordgo \
  ./internal/discord/customid ./internal/discord/interactions \
  ./internal/app ./internal/config ./cmd/mhcat-staging-preflight

go test -race ./internal/adapters/mongo/documents \
  ./internal/adapters/mongo/repositories \
  ./internal/core/services/economy \
  ./internal/discord/features/economy \
  ./internal/discord/responses ./internal/adapters/discordgo ./internal/app

go vet ./internal/adapters/mongo/documents \
  ./internal/adapters/mongo/repositories \
  ./internal/core/services/economy \
  ./internal/discord/features/economy \
  ./internal/discord/responses ./internal/adapters/discordgo ./internal/app

go run ./tools/parity-audit --legacy-root ../MHCAT --format markdown
```

The static report must remain `74/74` with zero drift. Tests lock dimensions, avatar pixels/mask, fonts, text width, full guild name, rounded bars, coordinates, scalar display/ranks/status, button grammar, attachment clearing, two-stage refresh, strict routing, and usage ownership.

Disposable real-Mongo verification is operator-gated:

```bash
MHCAT_RUN_MONGO_INTEGRATION_TESTS=true \
MHCAT_MONGODB_URI='<disposable-uri>' \
go test ./internal/adapters/mongo/repositories \
  -run '^TestEconomyProfileMongoIntegration'
```

It verifies no startup mutation, six-collection hydration, natural duplicate order, raw scalars, ranks, and status. Cleanup drops the generated database. Never use production.

## Staging Smoke

1. Stop Node interaction/component ownership. Use a staging token/guild and reviewed read-only Mongo data.
2. Enable both profile flags, run preflight and guarded guild-sync dry-run, and confirm only the managed addition/update.
3. Run self and optional-user profiles. Verify public loading reply, static caller PNG, exact 1500x750 card, target avatar/name/dates, fonts, all values, and refresh button.
4. Test missing optional rows, duplicate rows, ties, migrated/legacy discriminators, long CJK/emoji names, and null/missing/malformed/decimal/infinite fixtures in every source.
5. Refresh as the target and another member. Verify immediate loading update clears PNG/buttons, then final edit restores them. Remove the target and verify the ephemeral error leaves the old card unchanged.
6. With usage disabled, compare all collections/indexes before/after. With usage enabled separately, verify one slash increment and none from refresh.
7. Force Mongo, media, renderer, and Discord failures; verify controlled behavior, no secret leakage, duplicate response, collection/index creation, or data mutation.

## Rollback

Disable profile runtime and remove the managed staging-guild command through guarded sync. Stop Go ownership before restoring Node so both do not handle legacy `my-profile` buttons. No data rollback or migration reversal exists because the feature writes no profile/economy data or indexes. Usage rows follow the separate global rollback procedure.

Production rollout remains gated on live Discord card/refresh smoke, disposable real-Mongo execution, and exclusive ownership.
