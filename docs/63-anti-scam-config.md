# Anti-Scam Config, Report, and Message Delete Slices

Status: historical implementation note, superseded by the canonical [anti-scam parity contract](77-anti-scam.md). All runtime and command-sync gates remain disabled by default.

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

`/詐騙網址回報` validates with the loose `is-url@1.2.4` rules used by legacy, checks the existing `not_a_good_webs` catalog for a matching stored URL, and sends the legacy webhook report content when the URL has not already been reported. It does not insert into or mutate `not_a_good_webs`.

The message-delete runtime reads `good_webs.open`; when enabled for the guild, it scans message content for existing `not_a_good_webs.web` values, deletes the matching message, and sends the legacy trashbin warning text. Like legacy, it also scans bot-authored messages and preserves the warning's trailing newline.

## UI/UX Parity

The implemented config command path preserves:

- command name `防詐騙網址`
- description `設定是否開啟防詐騙網址功能(輸入這個指令就會更改)`
- public command discoverability with no Discord default-member-permission gate
- runtime Manage Messages check, including the Administrator override
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
- raw option text in validation, duplicate lookup, webhook content, and success output without trimming or normalization

The URL classifier matches `is-url@1.2.4`, not modern URL-parser semantics. It accepts protocol-relative input such as `//example.com`, arbitrary word-character schemes such as `ftp://` and `custom_1://`, localhost forms, and Unicode domains that satisfy the legacy suffix-length check. It rejects surrounding JavaScript whitespace and uses JavaScript UTF-16 code-unit length for the suffix, so `//a.😀` is accepted while `//a.é` is rejected.

## Mongo Compatibility

Collection: `good_webs`

Fields:

- `guild`
- `open`

Reads preserve Mongoose Boolean compatibility for mixed legacy BSON scalars. Boolean `true`, numeric `1`, and strings `true`, `1`, or `yes` enable scanning; Boolean `false`, numeric `0`, strings `false`, `0`, or `no`, null/missing values, and unsupported values decode as disabled.

Legacy deletes the found row and then inserts a replacement. Go intentionally updates all duplicate `{guild}` rows with `$set.open` and upserts only when no row matches. This preserves legacy field names while avoiding the delete/reinsert race.

The candidate `{guild:1}` singleton index remains duplicate-audit gated.

Collection: `not_a_good_webs`

Fields read:

- `web`

The report command reads this catalog only. It does not trim before validation, and valid reports use the same raw text for lookup and delivery. Legacy queried with a user-controlled regex; Go escapes the reported URL with `regexp.QuoteMeta` before building the Mongo regex to avoid regex injection while preserving the legacy "stored URL contains reported URL" lookup direction. The message-delete runtime scans stored `web` values and checks whether the message content contains one. Go does not write new rows to `not_a_good_webs` in this slice.

## Intentional Safety Differences

- Report lookup escapes the user-controlled Mongo regex while preserving the legacy lookup direction.
- Webhook content neutralizes triple-backtick breakout inside the reported URL.
- Go awaits webhook delivery before showing success and returns the safe generic error embed when delivery fails.
- Toggle writes update every duplicate `good_webs` row for the guild instead of deleting and reinserting only the first row.
- Message scanning checks whether content contains each catalog URL, fixing the legacy reversed-regex query that missed ordinary messages containing a listed URL.
- Go awaits successful message deletion before sending the warning.

## Parity Contracts

Focused tests lock command definitions and public discoverability, runtime permission handling, visible response payloads, `is-url@1.2.4` classification, raw URL preservation, Mongo filters and Boolean coercion, webhook formatting/sanitization, bot-message scanning, and the exact delete warning. Run:

```bash
go test ./internal/core/domain ./internal/core/services/safety ./internal/discord/features/safety ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/adapters/external ./internal/app ./internal/parity ./internal/config
```

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
