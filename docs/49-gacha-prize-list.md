# Gacha Prize-list and Prize-delete Slices

Status: implemented behind explicit staging/runtime gates.

## Legacy References

- `MHCAT/slashCommands/扭蛋系統/giftlist.js`
- `MHCAT/slashCommands/扭蛋系統/gift_delete.js`
- `MHCAT/models/gift.js`
- `MHCAT/models/gift_change.js`

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

The command preserves the legacy duplicate-prize error, overfull-pool error, optional defaults (`自動刪除=true`, `獎品數量=1`, `給予硬幣=0`), and the success embed fields from `giftadd.js`. The Go path intentionally keeps this as prize-config maintenance only: it does not draw prizes, decrement inventory, mutate user coin balances, send prize-code DMs, or enable shop behavior.

## Prize-delete Scope

- Slash command: `扭蛋獎池刪除`
- Runtime flag: `MHCAT_FEATURE_GACHA_PRIZE_DELETE_ENABLED=true`
- Command sync flag: `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_DELETE=true`
- Required option: string `獎品名稱`
- Permission: Manage Messages (`8192`) at command definition and runtime levels
- Mongo write: deletes one `gifts` row by `{guild,gift_name}`
- Discord behavior: public defer, edit original with legacy-style success/error embeds

The command preserves the legacy missing-prize error `找不到這個獎品!` and success embed title `<a:green_tick:994529015652163614>成功刪除!` with description `獎品名:<gift_name>`. It intentionally uses one-row delete semantics to match legacy `findOne(...); data.delete()` behavior when duplicate prize names exist.

## Not Implemented

- `/扭蛋` draw flow
- prize edit command
- gacha/shop purchase paths
- coin balance mutation
- prize inventory decrement
- prize code DMs
- notification channel sends
- indexes or data repair

## Rollout Notes

Do not sync the command unless the runtime flag is enabled for the same staging bot. `mhcat-staging-preflight` and the staging scripts reject unpaired command-sync/runtime flags.

For `/扭蛋獎池刪除`, use only disposable staging prize rows. The command removes data and has no undo path.

For `/扭蛋獎池增加`, use only disposable staging prize rows until the production duplicate-name policy and backup process are reviewed.

Production rollout still requires a live audit of `gifts` and `gift_changes`, especially duplicate `gift_changes.guild` rows, impossible prize counts/chances, and guilds with more than 25 prize rows.
