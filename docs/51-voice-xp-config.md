# Voice XP Config Slice

Status: implemented behind explicit runtime and command-sync gates.

## Legacy References

- `MHCAT/slashCommands/經驗系統/voice_set.js`
- `MHCAT/slashCommands/經驗系統/voice_set_delete.js`
- `MHCAT/models/voice_xp_channel.js`

## Implemented Scope

- Slash command: `語音經驗設定`
- Slash command: `語音經驗刪除`
- Runtime flag: `MHCAT_FEATURE_VOICE_XP_CONFIG_ENABLED=true`
- Command sync flag: `MHCAT_COMMAND_SYNC_INCLUDE_VOICE_XP_CONFIG=true`
- Mongo collection: `voice_xp_channels`
- Mongo fields preserved: `guild`, `channel`, `background`, `color`, `message`
- Permission: Manage Messages (`8192`) at command definition and runtime check
- Discord behavior: public defer, legacy-style green/red embeds, optional preview message

This command slice is announcement-config only. It does not enable Voice State intent, rank cards, or voice XP runtime by itself. Voice reward-role config is implemented separately behind `MHCAT_FEATURE_XP_ROLE_CONFIG_ENABLED=true`.

Voice XP runtime is implemented separately behind `MHCAT_FEATURE_VOICE_XP_SESSIONS_ENABLED=true`, with `MHCAT_DISCORD_ENABLE_GATEWAY=true` and `MHCAT_DISCORD_VOICE_STATE_INTENT=true`. That event slice mirrors the legacy join/leave session flag by upserting missing `voice_xps` rows with `xp:"0"`, `leavel:"0"`, and `leavejoin:"join"`/`"leave"`, starts one legacy 30-second XP loop per active joined user, reconciles existing `leavejoin:"join"` rows on startup, and stops loops on leave or app shutdown. The runtime preserves the legacy `+5 XP` tick, `xp:"5"` on level-up, configured/default voice level-up announcements, owner DM fallbacks for missing/unusable level-up channels, `voice_roles` changes, and XP coin rewards after the configured announcement path succeeds.

## Legacy UI/UX Preserved

`語音經驗設定` keeps:

- required channel option `頻道`, channel types text/news (`0`, `5`);
- optional string options `訊息`, `顏色`, and `背景`;
- the legacy `背景` option description even though the Node.js command never saved it;
- red permission error title with the legacy animated-no emoji;
- invalid color error `你傳送的並不是顏色(色碼)`;
- success embed title `語音經驗系統`;
- success description `您的語音經驗升等頻道成功創建\n您目前的升等通知頻道為 <#channel>`;
- optional preview content beginning `以下為你的訊息預覽:` and the legacy `<:line:992363971803881493>` separator.

`語音經驗刪除` keeps:

- no options;
- Manage Messages runtime check;
- success embed title `語音經驗系統` and description `成功刪除!`;
- missing config error `你本來就沒有對語音經驗設定喔!`.

## Intentional Safety Fixes

- Legacy deletes one found config document and inserts a new one. Go updates every duplicate `{guild}` row and only upserts when no row exists, avoiding a temporary missing-config window and keeping duplicate legacy rows consistent until a duplicate audit and unique-index plan are approved.
- Legacy preview sends raw message content with default mentions. Go preserves the visible preview text but uses empty allowed mentions to avoid accidental `@everyone`, role, or user pings during configuration.

## Compatibility Notes

- Saving a new voice-XP config clears the legacy optional `background` field because the legacy command exposed `背景` but did not save it.
- `color` is trimmed and written as a legacy-compatible nullable value.
- `message` preserves user-provided spacing and is written as a legacy-compatible nullable value.
- No indexes are created by the app. A future unique `voice_xp_channels.guild` index still requires duplicate audit first.
- Color validation accepts common legacy color-code values and safe CSS color names. Broader `validate-color` parity can be expanded if staging finds a documented accepted value that Go rejects.
- Voice XP tick math matches the legacy interval: joined users gain `5` XP per tick, level up when `xp + 5` exceeds `level * (level / 2) * 100 + 100`, and keep `xp:"5"` after that level-up.

## Not Implemented

- `/語音排行榜`, rank image rendering, rank buttons, and the old XP profile card lookup behind `/語音經驗`; the current `/語音經驗` command is implemented separately as a disabled replacement response only.
- Voice State intent enablement by the config commands; session tracking has its own explicit event gate.
- Usage counter writes to `all_use_count`.

## Rollout Notes

Do not sync `語音經驗設定` or `語音經驗刪除` unless the same staging runtime has `MHCAT_FEATURE_VOICE_XP_CONFIG_ENABLED=true`. `mhcat-staging-preflight` and the staging scripts reject unpaired command-sync/runtime flags.

Production rollout still requires a live audit of `voice_xp_channels` duplicate `{guild}` rows and a rollback review with the Node.js bot because Node continues to read the same collection.
