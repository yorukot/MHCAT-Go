# Anti-Scam Config, Report, and Message Delete Slices

Status: config toggle, URL report command, and message-delete runtime implemented behind explicit gates.

## Legacy References

- Toggle command: `MHCAT/slashCommands/群組防護/not_a_goodweb.js`
- Runtime scanner: `MHCAT/events/safe_server.js`
- Report command: `MHCAT/slashCommands/群組防護/report_web.js`
- Toggle model: `MHCAT/models/good_web.js`
- URL catalog model: `MHCAT/models/not_a_good_web.js`

## Implemented Surface

These slices implement:

- `/防詐騙網址`
- `/詐騙網址回報`
- `safe_server.js`-style MessageCreate scanning/deletion

`/防詐騙網址` toggles the guild `good_webs.open` flag.

`/詐騙網址回報` validates an HTTP or HTTPS URL, checks the existing `not_a_good_webs` catalog for a matching stored URL, and sends the legacy webhook report content when the URL has not already been reported. It does not insert into or mutate `not_a_good_webs`.

The message-delete runtime reads `good_webs.open`; when enabled for the guild, it scans message content for existing `not_a_good_webs.web` values, deletes the matching message, and sends the legacy trashbin warning text.

## UI/UX Parity

The implemented config command path preserves:

- command name `防詐騙網址`
- description `設定是否開啟防詐騙網址功能(輸入這個指令就會更改)`
- Manage Messages metadata and runtime check
- public defer/edit response flow
- red legacy permission error embed
- green `<:fraudalert:1000408260777611355> 自動偵測詐騙連結` success embed
- legacy boolean text format: `true` or `false`

The implemented report command path preserves:

- command name `詐騙網址回報`
- description `回報詐騙網站`
- required string option `網址` with description `回報網址`
- public defer/edit response flow
- red legacy invalid-URL embed title `你輸入的不是一個網址!`
- red legacy duplicate embed title `該網站已被回報過`
- green `<:fraudalert:1000408260777611355> 自動偵測詐騙連結` success embed
- legacy success text format `成功回報<url>`
- legacy webhook content shape with the URL in a code block and reporter mention on the next line

## Mongo Compatibility

Collection: `good_webs`

Fields:

- `guild`
- `open`

Legacy deletes the found row and then inserts a replacement. Go intentionally updates all duplicate `{guild}` rows with `$set.open` and upserts only when no row matches. This preserves legacy field names while avoiding the delete/reinsert race.

The candidate `{guild:1}` singleton index remains duplicate-audit gated.

Collection: `not_a_good_webs`

Fields read:

- `web`

The report command reads this catalog only. Legacy queried with a user-controlled regex; Go escapes the reported URL with `regexp.QuoteMeta` before building the Mongo regex to avoid regex injection while preserving the legacy "stored URL contains reported URL" lookup direction. The message-delete runtime scans stored `web` values and checks whether the message content contains one. Go does not write new rows to `not_a_good_webs` in this slice.

## Gates

Runtime:

```bash
MHCAT_FEATURE_ANTI_SCAM_CONFIG_ENABLED=true
MHCAT_FEATURE_ANTI_SCAM_REPORT_ENABLED=true
MHCAT_FEATURE_ANTI_SCAM_MESSAGE_DELETE_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true
MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true
MHCAT_REPORT_WEBHOOK_URL=https://example.test/webhook
```

Command sync:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_CONFIG=true
MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_REPORT=true
```

The legacy `REPORT_WEBHOOK` alias is also accepted for the report webhook URL. Runtime and command-sync flags must be paired in staging. `mhcat-staging-preflight` rejects anti-scam command sync when the matching runtime flag is not enabled, and config validation rejects report runtime when neither webhook URL variable is set. Message deletion is runtime-only and has no command-sync flag.

## Not Implemented

This slice does not implement:

- writes to `not_a_good_webs`

Gateway, Guild Messages intent, and Message Content intent remain disabled by default and must be enabled explicitly with `MHCAT_FEATURE_ANTI_SCAM_MESSAGE_DELETE_ENABLED=true`.

## Staging Checklist

1. Use an isolated staging guild and staging database.
2. Run command sync dry-run with the matching `MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_*` flags.
3. Enable the matching `MHCAT_FEATURE_ANTI_SCAM_*_ENABLED=true` runtime flags before applying command definitions.
4. Run `mhcat-staging-preflight`.
5. Apply guild-scoped command sync only after paired gate checks pass.
6. Verify `/防詐騙網址` creates `good_webs.guild` and `good_webs.open=true`.
7. Run `/防詐騙網址` again and verify `good_webs.open=false`.
8. With `MHCAT_REPORT_WEBHOOK_URL` or `REPORT_WEBHOOK` pointed at a safe staging endpoint, run `/詐騙網址回報 網址:https://bad.example`.
9. Verify the webhook receives the legacy report content and the interaction edits to the success embed.
10. Seed `not_a_good_webs.web` with a matching staging URL and verify `/詐騙網址回報` returns `該網站已被回報過`.
11. With `MHCAT_FEATURE_ANTI_SCAM_MESSAGE_DELETE_ENABLED=true` and gateway message-content flags enabled, keep `good_webs.open=true`, send a staging message containing the seeded URL, and verify Go deletes the message and sends `<:trashbin:995991389043163257> | 此消息包含詐騙或釣魚連結，以自動刪除!`.
12. Confirm Go does not write to `not_a_good_webs`.
