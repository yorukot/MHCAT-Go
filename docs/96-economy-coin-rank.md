# Economy Coin Rank Parity Contract

Status: parity-audited against `slashCommands/代幣系統/coin_rank.js`, the `coin_rank` branch in `events/rank.js`, legacy assets/fonts and Mongoose behavior, plus current Go definition, service, renderer, Mongo adapter, component router, app wiring, staging guards, and race coverage. Live Discord and operator-gated real-Mongo smoke remain required before production rollout.

## Scope

This contract covers `/代幣排行榜`, its five pagination controls, and viewer-target control. Other economy and XP rank surfaces are separate.

## Definition And Ownership

The chat-input command is `代幣排行榜`, description `查詢代幣的排行榜`, with no options or default member permission. Legacy cooldown metadata is `10`; neither legacy central dispatch nor Go adds a local throttle.

Runtime is disabled by default behind `MHCAT_FEATURE_ECONOMY_COIN_RANK_ENABLED=true`. Staging guild sync separately requires `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_RANK=true`; config, preflight, and scripts reject sync without runtime. Startup never syncs Discord commands.

Global slash middleware is the only usage owner. With tracking enabled, one best-effort `all_use_counts` event is recorded before route/handler work. Components add no usage event, and app wiring removes the module tracker to prevent a second success-only increment.

## Slash Lifecycle

The public slash path immediately replies with the exact orange loading embed:

- author `正在努力為您尋找資料!` and legacy loading GIF;
- footer `MHCAT 帶給你最好的discord體驗!`;
- caller avatar forced to Discord CDN PNG form, with legacy fallback;
- color `#FF5809`.

After reads, lookups, icon fetch, and rendering, it edits the original response to remove the embed and attach `user-info.png` plus components. Allowed mentions are suppressed.

## Data And Ranking

The command reads all guild `coins` rows in natural Mongo order, reverses them, then applies a stable JavaScript-compatible descending `b.coin-a.coin` sort. Equal numeric values retain reverse-document order. Duplicate `{guild,member}` rows remain separate entries.

Mongoose-visible scalar behavior is preserved:

- integers, decimals, null, `Infinity`, and `-Infinity` retain visible text;
- missing, malformed, and hydrated `NaN` values display `undefined`;
- comparison uses JavaScript coercion, including null as zero and nonnumeric comparisons as false;
- abbreviations use `K`, `M`, and `G`, one decimal, `.0` removal, and JavaScript half-up `toFixed(1)` behavior (`1250` is `1.3K`);
- positive infinity follows legacy formatting as `InfinityG`.

Viewer rank uses an unsorted `findOne`-equivalent balance and counts every row for which `row.coin < viewer.coin` is false. Ties use the legacy last-place rank: two users tied at the top both display `#2`. Missing viewer balance displays `沒有資料` and omits the target row.

## Canvas UI

Every render is a 1000x500 PNG over `asset/coin_rank_background.png`. It preserves:

- guild icon requested as PNG, rounded at 128x128 radius 40, then scaled to 70x70 at `(33,10)`; `asset/blue_discord.png` fallback;
- guild name at `(115,50)`, legacy UTF-16 width truncation, and TC/SC/JP/HK/Noto/Bengali/Arabic/emoji/Taipei fallback fonts at 37px;
- title `代幣排行榜` at `(118,74)` initially and `(118,70)` on update, 20px;
- centered viewer rank at `(710,70)` and creation date `YYYY/MM/DD` at `(790,70)`, 30px Comic Sans plus language fallbacks;
- ten numbered slots on every page, including empty pages, split five left/five right with 484px offset and 74px rows;
- legacy page-based rank font quirk: 40px through page 99, 30px on pages 100-999, 25px after page 999;
- user tags at 25px, including `username#discriminator` for nonzero legacy discriminators;
- missing users as `找不到該名使用者` and coin values at 15px.

Icon HTTP fetches honor context, time out after two seconds, and cap data at 2 MiB. Failed lookups/icons use visible fallbacks instead of leaving rejected image promises unresolved.

## Pagination

The first row preserves page minus 10, page minus 1, disabled `<current>/<total>`, page plus 1, and page plus 10. IDs remain `[viewer]{page}coin_rank`, with exact emojis, styles, and disable boundaries. Empty data intentionally shows `1/0` and slots 1-10.

When the viewer has a balance, the second row has two disabled spacers, target ID `[viewer]coin_rank {Math.trunc(rank/10)}`, and two spacers. The legacy off-by-one is retained, so rank 10 targets page index 1.

Components directly update the existing message. Any member may press public controls; the embedded viewer remains the rank target. A missing viewer returns exact ephemeral red `找不到資料!請於幾分鐘後重試!`. Go accepts only exact legacy ID grammar instead of substring matches, an intentional collision-prevention difference.

## Mongo And Migration

The feature reads `coins` only. Construction and execution create no collection, index, migration, backfill, or economy row. No migration is required.

Natural order and duplicates are rollback compatibility behavior. Before adding a unique `{guild,member}` index or changing query sort/projection, operators must audit duplicates and approve the visible ordering change. The catalog's nonunique rank index is an offline migration plan only and is not created at startup.

## Errors And Reliability Differences

Legacy callbacks ignored Mongo errors and image promises could reject after loading. Go propagates backend/render/edit failures to controlled runtime handling, bounds icon downloads, strictly parses IDs, and may resolve an uncached member through Discord instead of failing the cache-only legacy check. These prevent hangs, collisions, and internal leakage without changing successful UI.

## Verification

```bash
go test ./internal/adapters/mongo/documents \
  ./internal/adapters/mongo/repositories \
  ./internal/core/services/economy \
  ./internal/discord/features/economy \
  ./internal/discord/customid ./internal/discord/interactions \
  ./internal/app ./internal/config ./cmd/mhcat-staging-preflight

go test -race ./internal/adapters/mongo/repositories \
  ./internal/core/services/economy \
  ./internal/discord/features/economy ./internal/app

go vet ./internal/adapters/mongo/repositories \
  ./internal/core/services/economy \
  ./internal/discord/features/economy ./internal/app

go run ./tools/parity-audit --legacy-root ../MHCAT --format markdown
```

The static report must remain `74/74` with zero drift. Tests lock dimensions, icon pixels, font stacks, text width, geometry, empty slots, lifecycle, controls, routing, scalar ordering, and usage ownership.

Disposable real-Mongo verification is operator-gated:

```bash
MHCAT_RUN_MONGO_INTEGRATION_TESTS=true \
MHCAT_MONGODB_URI='<disposable-uri>' \
go test ./internal/adapters/mongo/repositories \
  -run '^TestEconomyCoinRankMongoIntegration'
```

It verifies natural order, duplicates, malformed/decimal hydration, service ordering, and viewer rank. Cleanup drops the generated database. Never use production.

## Staging Smoke

1. Stop Node interaction/component ownership. Use a staging token/guild and reviewed read-only Mongo data.
2. Enable both rank flags, run preflight and guarded guild-sync dry-run, and confirm only the managed addition/update.
3. Run empty data and verify loading embed, `1/0`, slots 1-10, one component row, background, fallback icon, and no writes.
4. Seed 12 or more disposable rows including ties, duplicates, a legacy discriminator, missing user, decimal, null, malformed, and infinite values. Verify order, labels, viewer rank, geometry, fonts, and both pages.
5. Exercise all page controls, target-viewer, another member clicking, and missing embedded viewer. Verify direct updates and disabled boundaries.
6. With usage disabled, compare collections/indexes before/after. With usage enabled separately, verify one slash increment and none from components.
7. Force Mongo, icon HTTP, render asset, and Discord failures; verify controlled behavior, no secret text, duplicate response, or Mongo mutation.

## Rollback

Disable rank runtime and remove the managed staging-guild command through guarded sync. Stop Go ownership before restoring Node so both do not handle legacy `coin_rank` components. No economy rollback or migration reversal exists because this feature writes no economy data or indexes. Usage rows follow the separate global rollback procedure.

Production rollout remains gated on live Discord image/component smoke, disposable real-Mongo execution, and exclusive ownership.
