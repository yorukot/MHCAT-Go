# Gacha Prize-list Slice

Status: implemented behind explicit staging/runtime gates.

## Legacy References

- `MHCAT/slashCommands/扭蛋系統/giftlist.js`
- `MHCAT/models/gift.js`
- `MHCAT/models/gift_change.js`

## Implemented Scope

- Slash command: `扭蛋獎池查詢`
- Runtime flag: `MHCAT_FEATURE_GACHA_PRIZE_LIST_ENABLED=true`
- Command sync flag: `MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_LIST=true`
- Mongo reads only: `gifts` by `guild`, `gift_changes` by `guild`
- Mongo writes: none
- Discord behavior: public defer, edit original with legacy-style embeds

The command preserves legacy visible text, emoji IDs, public response behavior, and the misspelled legacy `gift_chence` field. Missing `gift_changes` config falls back to the legacy defaults: gacha cost `500`, sign coins `25`, and XP multiple `0`.

## Intentional Fix

The legacy command adds one embed field per prize and can fail Discord validation when a guild has more than 25 prizes. The Go handler preserves the exact one-embed UI for pools up to 25 prizes and splits larger pools into multiple embeds. It does not change Mongo data.

## Not Implemented

- `/扭蛋` draw flow
- prize add/edit/delete commands
- gacha/shop purchase paths
- coin balance mutation
- prize inventory decrement
- prize code DMs
- notification channel sends
- indexes or data repair

## Rollout Notes

Do not sync the command unless the runtime flag is enabled for the same staging bot. `mhcat-staging-preflight` and the staging scripts reject unpaired command-sync/runtime flags.

Production rollout still requires a live audit of `gifts` and `gift_changes`, especially duplicate `gift_changes.guild` rows, impossible prize counts/chances, and guilds with more than 25 prize rows.
