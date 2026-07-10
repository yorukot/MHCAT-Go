# Voice XP Config Slice

Status: parity-audited behind explicit runtime and command-sync gates.

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
- Usage: the global slash middleware writes one `all_use_counts` event before every routed command when `MHCAT_FEATURE_USAGE_TRACKING_ENABLED=true`

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
- Go awaits Mongo writes and surfaces a safe legacy-style error payload when persistence fails; the legacy callback launched delete/save/edit operations without awaiting their results.

## Compatibility Notes

- Saving a new voice-XP config clears the legacy optional `background` field because the legacy command exposed `背景` but did not save it.
- `color` is preserved verbatim when present and written as a legacy-compatible nullable value.
- `message` preserves user-provided spacing and is written as a legacy-compatible nullable value.
- No indexes are created by the app. A future unique `voice_xp_channels.guild` index still requires duplicate audit first.
- Color validation mirrors the pinned legacy `validate-color` 2.2.4 package, including 3/4/6/8-digit prefixed hex, HTML names and special names, and the package's `rgb`/`rgba`/`hsl`/`hsla`/`hwb`/`lab`/`lch` grammar. Bare hex and outer whitespace remain invalid.
- Voice XP tick math matches the legacy interval: joined users gain `5` XP per tick, level up when `xp + 5` exceeds `level * (level / 2) * 100 + 100`, and keep `xp:"5"` after that level-up.

## Not Implemented

- The old XP profile-card lookup and rank-card render behind `/語音經驗`; the current `/語音經驗` command is implemented separately as the legacy disabled replacement response. `/語音排行榜` is implemented under its own XP-rank gates.
- Voice State intent enablement by the config commands; session tracking has its own explicit event gate.

## Rollout Notes

Do not sync `語音經驗設定` or `語音經驗刪除` unless the same staging runtime has `MHCAT_FEATURE_VOICE_XP_CONFIG_ENABLED=true`. `mhcat-staging-preflight` and the staging scripts reject unpaired command-sync/runtime flags.

Production rollout still requires a live audit of `voice_xp_channels` duplicate `{guild}` rows and a rollback review with the Node.js bot because Node continues to read the same collection.
