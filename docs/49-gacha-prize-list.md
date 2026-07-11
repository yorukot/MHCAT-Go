# Gacha Draw and Prize Maintenance Parity Contract

Status: parity-audited against the legacy draw, prize-list, prize-create, prize-edit, and prize-delete commands, Mongoose models, and current Go definitions, services, handlers, scalar adapters, repository, app wiring, staging guards, and focused race coverage. Live Discord and operator-gated real-Mongo smoke remain required before production rollout.

## Legacy References

- `MHCAT/slashCommands/扭蛋系統/giftlist.js`
- `MHCAT/slashCommands/扭蛋系統/gashapon.js`
- `MHCAT/slashCommands/扭蛋系統/giftadd.js`
- `MHCAT/slashCommands/扭蛋系統/giftadd copy.js`
- `MHCAT/slashCommands/扭蛋系統/gift_delete.js`
- `MHCAT/models/gift.js`
- `MHCAT/models/gift_change.js`
- `MHCAT/models/coin.js`

## Draw Scope

- Slash command: `扭蛋`
- Runtime flag: `MHCAT_FEATURE_GACHA_DRAW_ENABLED=true`
- Command sync flag: `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_DRAW=true`
- Optional option: string `連抽` with legacy choices `5`, `11`, `17`, and `23`
- Mongo reads: `coins`, `gifts`, `gift_changes`
- Mongo writes: updates matching `coins` rows for the member; decrements or deletes auto-delete `gifts` rows by `{guild,gift_name}`
- Discord behavior: public defer, loading GIF follow-up, 8.5-second reveal delay, edit of that follow-up to the final legacy result embed, best-effort prize-code DM, and one best-effort notification-channel winner embed per non-air draw

The command preserves the legacy visible draw UI, draw-count mapping, missing-balance/empty-pool/insufficient-coin errors, default gacha cost `500`, weighted air result, prize-code DM fields, and notification-channel embed text. Paid draws remain `1`, `5`, `10`, `15`, or `20`; actual result lines remain `1`, `5`, `11`, `17`, or `23`. Multi-draws reload the `gifts` pool before every actual draw and apply that draw's inventory mutation immediately, so a depleted auto-delete prize is unavailable to later draws in the same command.

## Draw Safety Fix

The legacy draw code can touch auto-delete prize inventory twice through separate async callback loops. The Go implementation preserves the per-draw pool reload and depletion behavior but applies one intended decrement/delete per drawn auto-delete prize. Coin updates are also applied once as `starting coin - cost + give_coin total`, and duplicate `{guild,member}` coin rows are updated together for rollback compatibility with the current duplicate-audit posture. Prize inventory writes remain non-transactional and target one legacy prize name row, so keep this path limited to isolated staging data until production duplicate and transaction policy is reviewed.

## Prize-list Scope

- Slash command: `扭蛋獎池查詢`
- Runtime flag: `MHCAT_FEATURE_GACHA_PRIZE_LIST_ENABLED=true`
- Command sync flag: `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_LIST=true`
- Mongo reads only: `gifts` by `guild`, `gift_changes` by `guild`
- Mongo writes: none
- Discord behavior: public defer, edit original with legacy-style embeds

The command preserves legacy visible text, emoji IDs, public response behavior, and the misspelled legacy `gift_chence` field. Missing `gift_changes` config falls back to the legacy defaults: gacha cost `500`, sign coins `25`, and XP multiple `0`.

## Intentional Fix

The legacy command adds one embed field per prize and can fail Discord validation when a guild has more than 25 prizes. The Go handler preserves the exact one-embed UI for pools up to 25 prizes and splits larger pools into multiple embeds. It does not change Mongo data.

## Prize-add Scope

- Slash command: `扭蛋獎池增加`
- Runtime flag: `MHCAT_FEATURE_GACHA_PRIZE_CREATE_ENABLED=true`
- Command sync flag: `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_CREATE=true`
- Required options: string `獎品名稱`, number `機率`
- Optional options: string `獎品代碼`, boolean `自動刪除`, integer `獎品數量`, integer `給予硬幣`
- Permission: Manage Messages (`8192`) at command definition and runtime levels
- Mongo write: inserts one `gifts` row with legacy-compatible fields
- Discord behavior: ephemeral defer, edit original with legacy-style success/error embeds

The command preserves the legacy duplicate-prize error, overfull-pool error, optional defaults (`自動刪除=true`, `獎品數量=1`, `給予硬幣=0`), and the success embed fields from `giftadd.js`. Prize names and codes are matched, stored, and displayed without trimming. The apparent name-length guard is also preserved as the legacy JavaScript numeric comparison: numeric-looking names greater than `200` are rejected, while long nonnumeric names are accepted. A submitted zero chance is stored as BSON `null`, matching the legacy falsey constructor expression. The Go path intentionally keeps this as prize-config maintenance only: it does not draw prizes, decrement inventory, mutate user coin balances, send prize-code DMs, or enable shop behavior.

## Prize-edit Scope

- Slash command: `扭蛋獎品編輯`
- Runtime flag: `MHCAT_FEATURE_GACHA_PRIZE_EDIT_ENABLED=true`
- Command sync flag: `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_EDIT=true`
- Required option: string `獎品名稱`
- Optional options: number `機率`, string `獎品代碼`, boolean `自動刪除`, integer `獎品數量`, integer `給予硬幣`
- Permission: Manage Messages (`8192`) at command definition and runtime levels
- Mongo write: deletes one `gifts` row by `{guild,gift_name}` and inserts one merged replacement row
- Discord behavior: ephemeral defer, edit original with legacy-style success/error embeds

The command preserves the legacy success embed title `<a:green_tick:994529015652163614>編輯成功成功`, exact untrimmed prize names/codes, the JavaScript numeric name guard, and the visible submitted/default field text. It also preserves legacy merge quirks: a new non-empty prize code replaces the old one; omitted or zero chance and give-coin keep the old value; false `自動刪除` does not override an existing true value; and omitted or zero count saves as `1`. Missing prize uses the same red `找不到這個獎品!` error as delete instead of allowing a legacy panic path.

The write path intentionally follows the legacy delete-plus-insert shape and does not run in a transaction. If insertion fails after deletion, the old row is not restored. Keep this path limited to disposable staging prize rows until a production backup and rollback policy exists.

## Prize-delete Scope

- Slash command: `扭蛋獎池刪除`
- Runtime flag: `MHCAT_FEATURE_GACHA_PRIZE_DELETE_ENABLED=true`
- Command sync flag: `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_DELETE=true`
- Required option: string `獎品名稱`
- Permission: Manage Messages (`8192`) at command definition and runtime levels
- Mongo write: deletes one `gifts` row by `{guild,gift_name}`
- Discord behavior: public defer, edit original with legacy-style success/error embeds

The command preserves the legacy missing-prize error `找不到這個獎品!` and success embed title `<a:green_tick:994529015652163614>成功刪除!` with description `獎品名:<gift_name>`. Prize names are matched and displayed exactly without trimming. It intentionally uses one-row delete semantics to match legacy `findOne(...); data.delete()` behavior when duplicate prize names exist.

## Operational Gates

Indexes and automatic data repair are intentionally outside the runtime parity scope. Production rollout requires an explicit audit and operator-reviewed repair policy for duplicate or malformed `coins`, `gifts`, and `gift_changes` rows; the refactor must not silently normalize legacy data.

Shop behavior is separately parity-audited in [104-economy-shop.md](104-economy-shop.md), including its UI, scalar coercion, duplicate-row behavior, role and DM side effects, staging gates, and rollback requirements.

## Verification

```bash
go test ./internal/core/services/gacha ./internal/adapters/mongo/documents \
  ./internal/adapters/mongo/repositories ./internal/discord/features/gacha ./internal/app
go test -race ./internal/core/services/gacha ./internal/adapters/mongo/repositories \
  ./internal/discord/features/gacha ./internal/app
go vet ./...
```

The guarded Mongo integration harness requires a generated disposable database and skips unless both `MHCAT_RUN_MONGO_INTEGRATION_TESTS=true` and `MHCAT_MONGODB_URI` are set. Never point it at production.

## Rollout Notes

Do not sync the command unless the runtime flag is enabled for the same staging bot. `mhcat-staging-preflight` and the staging scripts reject unpaired command-sync/runtime flags.

For `/扭蛋獎池增加`, use only disposable staging prize rows until the production duplicate-name policy and backup process are reviewed.

For `/扭蛋獎品編輯`, use only disposable staging prize rows. The command replaces data using delete-plus-insert and has no rollback path.

For `/扭蛋獎池刪除`, use only disposable staging prize rows. The command removes data and has no undo path.

For `/扭蛋`, use only isolated staging balances and prize rows. The command mutates `coins` and `gifts`, may send DMs, and may send notification-channel messages.

Production rollout still requires a live audit of `coins`, `gifts`, and `gift_changes`, especially duplicate `coins.{guild,member}` rows, duplicate `gift_changes.guild` rows, duplicate prize names, impossible prize counts/chances, and guilds with more than 25 prize rows.
