# Economy Settings Parity Contract

Status: parity-audited against `slashCommands/代幣系統/coin_cange.js`, `models/gift_change.js`, and current Go definition, handler, service, BSON adapter, Mongo repository, app wiring, staging guards, integration harness, and race coverage. Live Discord and operator-gated real-Mongo smoke remain required before production rollout.

## Scope

This contract covers `/coin-related-settings` / localized `代幣相關設定`. It writes shared `gift_changes` configuration but does not authorize gacha, sign-in, XP reward, notification delivery, index creation, or duplicate cleanup.

## Definition And Ownership

The command preserves exact base/localized names and descriptions, five required options and localizations, text/announcement channel restriction, Manage Messages default permission, cooldown metadata `10`, and the legacy malformed documentation URL `https://docsmhcat.yorukot.meocs/required_coins`. Legacy did not centrally enforce the cooldown; Go adds no local throttle.

Runtime is disabled by default behind `MHCAT_FEATURE_ECONOMY_SETTINGS_ENABLED=true`. Staging command sync separately requires `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SETTINGS=true`; config, preflight, and scripts reject sync without runtime. Startup never syncs commands or creates Mongo objects.

Global slash middleware owns usage. With tracking enabled, success, permission denial, and validation failure each record exactly one best-effort event. App wiring removes module tracking to prevent double counting.

## Lifecycle And UI

The command publicly defers before validation, matching legacy. Validation order is maximum gacha cost, maximum sign reward, negative cooldown, then runtime Manage Messages. Discord cannot change a public defer to ephemeral later, so visible error embeds remain public rather than reproducing legacy's invalid `ephemeral:true` edit attempt.

Errors use one red embed with exact title prefix `<a:Discord_AnimatedNo:1015989839809757295> | ` and exact legacy text for the maximum, cooldown, and permission cases.

Success edits the original response to one random-color embed with exact multiline title values, description `通知頻道:<#channel>`, and footer `MHCAT`. The footer uses the member's guild display avatar, including animated guild avatars, with user-avatar fallback. Allowed mentions are empty; the visible channel mention remains intact.

## Values And Scalars

`coin-raffle-takes`, `check-in-cooldown-time`, and `check-in-give-coins` are Discord integers. `level-up-multiply-amount` is a Discord number and retains fractional values. Legacy rejects only gacha/sign values above `999999999` and cooldown below zero. Negative gacha cost, sign reward, and XP multiplier remain accepted, stored, and displayed.

Cooldown hours are multiplied by 3600 using JavaScript-compatible double arithmetic. Values whose seconds exceed `int64` retain the double in `ResetMarkerText` and BSON; the operational integer fallback saturates positive overflow so daily-reset/work-payout mode checks still identify rolling mode. Ordinary values retain their exact integer meaning, and zero remains daily mode.

## Mongo Replacement

The new row uses exact legacy fields `guild`, `coin_number`, `sign_coin`, `channel`, `xp_multiple`, and `time`. Save finds and deletes one arbitrary natural matching `gift_changes` row, then inserts one new row. With no match, it only inserts. With duplicates, one old row remains for each extra duplicate, so count and natural-read ambiguity match legacy final-state behavior.

Go sequences and checks delete/insert errors instead of launching non-awaited Mongoose operations. A delete success followed by insert failure is still a possible partial replacement and requires backup/audit recovery. No unique index or transaction is created automatically.

Because `gift_changes` is shared, replacement immediately affects sign mode/reward, gacha price/notification, XP reward multiplier, daily reset exclusion, profile display, and work-payout marker behavior wherever those consumers are enabled. Production requires a duplicate/scalar audit and a reviewed shared-consumer ownership plan.

## Verification

```bash
go test ./internal/adapters/mongo/documents \
  ./internal/adapters/mongo/repositories \
  ./internal/core/services/economy \
  ./internal/discord/features/economy \
  ./internal/adapters/discordgo ./internal/discord/interactions \
  ./internal/app ./internal/config ./cmd/mhcat-staging-preflight

go test -race ./internal/adapters/mongo/documents \
  ./internal/adapters/mongo/repositories \
  ./internal/core/services/economy \
  ./internal/discord/features/economy \
  ./internal/adapters/discordgo ./internal/app

go vet ./...
go run ./tools/parity-audit --legacy-root ../MHCAT --format markdown
```

The static report must remain `74/74` with zero definition drift. Tests lock definition metadata, validation/permission order, negative and oversized values, random success color, exact embed text/channel/footer, guild avatar mapping, BSON scalar writes, replacement behavior, route isolation, flags, and usage ownership.

Disposable real-Mongo verification is operator-gated:

```bash
MHCAT_RUN_MONGO_INTEGRATION_TESTS=true \
MHCAT_MONGODB_URI='<disposable-uri>' \
go test ./internal/adapters/mongo/repositories \
  -run '^TestEconomySettingsMongoIntegration'
```

It verifies one-row duplicate replacement and exact negative/fractional/time fields. The shared harness drops the generated database. Never use production.

## Staging Smoke

1. Stop Node settings ownership. Use a staging guild and disposable restored database.
2. Audit duplicate `gift_changes.guild` rows and BSON types/values for all six fields. Back up the collection.
3. Enable both settings flags, run preflight and guarded command-sync dry-run, and confirm only the managed command change.
4. Test maximum and negative-cooldown errors before permission denial. Verify public defer/edit, exact red title, and no write.
5. Test zero, ordinary, negative gacha/sign/XP, fractional XP, and oversized positive cooldown inputs. Verify exact random-color success text, channel mention, guild avatar, and stored BSON.
6. Seed two duplicate rows, save once, and verify one old plus one new row remain. Confirm no collection/index creation.
7. With usage enabled, verify one event for each success/error attempt. Smoke every enabled shared consumer against the new row.
8. Force delete, insert, and Discord edit failures separately; record and repair any delete-success/insert-failure state from backup.

## Rollback

Disable settings runtime and remove the managed staging command through guarded sync. Stop Go ownership before restoring Node. Audit `gift_changes` first: do not assume the previously visible row survived replacement when duplicates existed. Restore only a reviewed backup row; do not automatically deduplicate, normalize scalars, or create/drop indexes.

Production rollout remains gated on live Discord smoke, disposable real-Mongo execution, backup/repair procedure, duplicate review, shared-consumer validation, and exclusive command ownership.
