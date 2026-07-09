# Text XP Config Slice

Status: implemented behind explicit runtime and command-sync gates.

## Legacy References

- `MHCAT/slashCommands/經驗系統/text_set.js`
- `MHCAT/slashCommands/經驗系統/text_set_delete.js`
- `MHCAT/models/text_xp_channel.js`

## Implemented Scope

- Slash command: `聊天經驗設定`
- Slash command: `聊天經驗刪除`
- Runtime flag: `MHCAT_FEATURE_TEXT_XP_CONFIG_ENABLED=true`
- Command sync flag: `MHCAT_COMMAND_SYNC_INCLUDE_TEXT_XP_CONFIG=true`
- Mongo collection: `text_xp_channels`
- Mongo fields preserved: `guild`, `channel`, `color`, `message`, `background`
- Permission: Manage Messages (`8192`) at command definition and runtime check
- Discord behavior: public defer, legacy-style green/red embeds, optional preview message

This slice is config-only. It does not enable message-create XP accrual, rank cards, level-role rewards, voice XP, coin rewards, or Message Content intent.

## Legacy UI/UX Preserved

`聊天經驗設定` keeps:

- required channel option `頻道`, channel types text/news (`0`, `5`);
- optional string options `訊息` and `顏色`;
- red permission error title with the legacy animated-no emoji;
- invalid color error `你傳送的並不是顏色(色碼)`;
- success embed title `聊天經驗系統`;
- success description `您的聊天經驗升等頻道成功創建\n您目前的升等通知頻道為 <#channel>`;
- optional preview content beginning `以下為你的訊息預覽:` and the legacy `<:line:992363971803881493>` separator.

`聊天經驗刪除` keeps:

- no options;
- Manage Messages runtime check;
- success embed title `聊天經驗系統` and description `成功刪除!`;
- missing config error `你本來就沒有對聊天經驗設定喔!`.

## Intentional Safety Fixes

- Legacy deletes one found config document and inserts a new one. Go updates every duplicate `{guild}` row and only upserts when no row exists, avoiding a temporary missing-config window and keeping duplicate legacy rows consistent until a duplicate audit and unique-index plan are approved.
- Legacy preview sends raw message content with default mentions. Go preserves the visible preview text but uses empty allowed mentions to avoid accidental `@everyone`, role, or user pings during configuration.

## Compatibility Notes

- Saving a new text-XP config clears the legacy optional `background` field to mirror the delete-and-reinsert behavior.
- `color` and `message` are written as legacy-compatible nullable values.
- No indexes are created by the app. A future unique `text_xp_channels.guild` index still requires duplicate audit first.
- Color validation accepts common legacy color-code values and safe CSS color names. Broader `validate-color` parity can be expanded if staging finds a documented accepted value that Go rejects.

## Not Implemented

- `events/LevelSystem.js` / text XP accrual.
- `/聊天排行榜`, rank image rendering, rank buttons, and the old XP profile card lookup behind `/聊天經驗`; the current `/聊天經驗` command is implemented separately as a disabled replacement response only.
- XP-to-coin rewards.
- chat level-role config.
- voice XP and voice level-role behavior.
- Message Content or Guild Messages intent enablement.
- Usage counter writes to `all_use_count`.

## Rollout Notes

Do not sync `聊天經驗設定` or `聊天經驗刪除` unless the same staging runtime has `MHCAT_FEATURE_TEXT_XP_CONFIG_ENABLED=true`. `mhcat-staging-preflight` and the staging scripts reject unpaired command-sync/runtime flags.

Production rollout still requires a live audit of `text_xp_channels` duplicate `{guild}` rows and a rollback review with the Node.js bot because Node continues to read the same collection.
