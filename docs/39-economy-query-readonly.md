# Economy Query Read-only Slice

## Scope

Implemented the low-risk read-only `/代幣查詢` slice:

- command definition `代幣查詢` with optional user option `使用者`;
- read-only Mongo repository for `coins` and `gift_changes`;
- permissive BSON decoding for legacy numeric drift;
- runtime feature gate `MHCAT_FEATURE_ECONOMY_QUERY_ENABLED=false`;
- command-sync include gate `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_QUERY=false`;
- staging preflight/script pairing checks for sync/runtime flags;
- legacy-style ephemeral embed responses.

## Legacy UI Preserved

The handler follows `slashCommands/代幣系統/coin.js`:

- initial response is deferred ephemeral;
- no-balance title is `<a:Discord_AnimatedNo:1015989839809757295> | 你還沒有任何代幣欸使用\`/簽到\`或是多講話，都可以獲得代幣喔!`;
- success title format is `<:money:997374193026994236><user>目前有:\`<coin>\`個代幣!`;
- description keeps the legacy “我該如何獲取代幣?” copy;
- missing `gift_changes` config displays the legacy default gacha cost `500`;
- the legacy footer quirk is preserved for missing config: `<user>還差:500`;
- selected-user lookups use the Discord info port when available and fall back to the user ID if lookup fails.

## Intentional Internal Changes

- Success embed color is stable instead of Discord.js `Random`; the old random color does not affect behavior and makes tests brittle.
- Mongo callback errors are mapped through typed Go errors instead of being ignored or exposed.
- The economy handler does not own usage writes. The separate global middleware increments `all_use_counts` only when `MHCAT_FEATURE_USAGE_TRACKING_ENABLED=true`.

## Outside This Slice

- `/簽到` now has its own disabled-by-default staging write slice documented in `docs/40-economy-signin.md`.
- `/代幣排行榜` is implemented separately as a disabled-by-default read-only PNG leaderboard slice.
- `/代幣增加`, `/剪刀石頭布`, and `/代幣重製` are implemented separately as disabled-by-default staging coin write slices; `my-profile` is implemented separately as a disabled-by-default read-only PNG profile slice; remaining economy games/shop paths, production write ownership, XP rewards, and daily reset ownership are outside this read-only slice.
- Production command sync for `/代幣查詢`.
- Mongo index creation on `coins` or `gift_changes`.
- Production-ready economy write repositories.

## Tests Added

- command definition and staging ownership validation;
- BSON decode tests for `coins` and `gift_changes`;
- read-only query service tests;
- handler tests for self query, selected user query, no-balance error, no-config default, and usage tracking;
- app wiring tests proving the route is disabled without an explicit repository;
- config/command-sync/preflight tests for economy-query gates.

## Next Step

Run staging guild dry-run for `/代幣查詢` with both `MHCAT_FEATURE_ECONOMY_QUERY_ENABLED=true` and `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_QUERY=true`, then smoke `/代幣查詢` against staging Mongo before enabling additional economy commands.
