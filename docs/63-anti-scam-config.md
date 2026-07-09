# Anti-Scam Config Slice

Status: config toggle implemented behind explicit runtime and command-sync gates.

## Legacy References

- Toggle command: `MHCAT/slashCommands/群組防護/not_a_goodweb.js`
- Runtime scanner: `MHCAT/events/safe_server.js`
- Report command: `MHCAT/slashCommands/群組防護/report_web.js`
- Toggle model: `MHCAT/models/good_web.js`
- URL catalog model: `MHCAT/models/not_a_good_web.js`

## Implemented Surface

This slice implements only:

- `/防詐騙網址`

It toggles the guild `good_webs.open` flag. It does not scan messages, delete messages, send scam warnings, or report URLs.

## UI/UX Parity

The implemented command path preserves:

- command name `防詐騙網址`
- description `設定是否開啟防詐騙網址功能(輸入這個指令就會更改)`
- Manage Messages metadata and runtime check
- public defer/edit response flow
- red legacy permission error embed
- green `<:fraudalert:1000408260777611355> 自動偵測詐騙連結` success embed
- legacy boolean text format: `true` or `false`

## Mongo Compatibility

Collection: `good_webs`

Fields:

- `guild`
- `open`

Legacy deletes the found row and then inserts a replacement. Go intentionally updates all duplicate `{guild}` rows with `$set.open` and upserts only when no row matches. This preserves legacy field names while avoiding the delete/reinsert race.

The candidate `{guild:1}` singleton index remains duplicate-audit gated.

## Gates

Runtime:

```bash
MHCAT_FEATURE_ANTI_SCAM_CONFIG_ENABLED=true
```

Command sync:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_CONFIG=true
```

Both flags must be paired in staging. `mhcat-staging-preflight` rejects anti-scam command sync when the runtime flag is not enabled.

## Not Implemented

This slice does not implement:

- `/詐騙網址回報`
- `REPORT_WEBHOOK`
- writes to `not_a_good_webs`
- `safe_server.js` message deletion
- scam URL regex matching or normalization
- Gateway enablement
- Guild Messages intent
- Message Content intent

Those paths require a separate safety review for URL normalization, regex behavior, message deletion permissions, mention behavior, rate limits, and rollout ownership.

## Staging Checklist

1. Use an isolated staging guild and staging database.
2. Run command sync dry-run with `MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_CONFIG=true`.
3. Enable `MHCAT_FEATURE_ANTI_SCAM_CONFIG_ENABLED=true` before applying command definitions.
4. Run `mhcat-staging-preflight`.
5. Apply guild-scoped command sync only after paired gate checks pass.
6. Verify `/防詐騙網址` creates `good_webs.guild` and `good_webs.open=true`.
7. Run `/防詐騙網址` again and verify `good_webs.open=false`.
8. Confirm Go does not delete messages or send URL report webhooks.
