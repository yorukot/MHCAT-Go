# Anti-Scam Parity Contract

Status: parity-audited against the legacy toggle command, report command, MessageCreate listener, Mongoose schemas, pinned `is-url` package, global slash dispatcher, and discord.js behavior. Runtime and command sync remain disabled by default. Live staging smoke is still required before production ownership.

## Scope

This contract covers:

- `/防詐騙網址` config toggling;
- `/詐騙網址回報` validation, duplicate lookup, webhook delivery, and response UI;
- `events/safe_server.js` MessageCreate scanning, deletion, and warning;
- `good_webs` and `not_a_good_webs` compatibility;
- usage accounting, intent gates, ownership, migration, and rollback.

Legacy sources:

- `slashCommands/群組防護/not_a_goodweb.js`
- `slashCommands/群組防護/report_web.js`
- `events/safe_server.js`
- `events/SlashCommands.js`
- `models/good_web.js`
- `models/not_a_good_web.js`
- pinned `is-url` 1.2.4

Account-age protection is a separate onboarding contract and is not part of this anti-scam ownership slice.

## Gates And Ownership

Toggle routes require:

```bash
MHCAT_FEATURE_ANTI_SCAM_CONFIG_ENABLED=true
MHCAT_STAGING_MODE=true
MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_CONFIG=true
```

Report routes require:

```bash
MHCAT_FEATURE_ANTI_SCAM_REPORT_ENABLED=true
MHCAT_STAGING_MODE=true
MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_REPORT=true
MHCAT_REPORT_WEBHOOK_URL=https://example.test/webhook
```

Legacy `REPORT_WEBHOOK` is accepted when the prefixed variable is absent.

Message deletion requires:

```bash
MHCAT_FEATURE_ANTI_SCAM_MESSAGE_DELETE_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true
MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true
```

The three route families are independent but share the same collections. Do not leave the equivalent Node command or `events/safe_server.js` active for the same bot/guild while Go owns that family. Config can move separately from deletion, but concurrent Node/Go config writers can race and should not overlap.

Command sync is guild-scoped and staging-only. Preflight and staging scripts reject unpaired command/runtime flags. Preflight permits Message Content for deletion only when the explicit anti-scam event gate is enabled and verifies all three gateway prerequisites.

## Command And Usage Contract

Both definitions preserve exact names, descriptions, option metadata, required flags, and public discoverability. Legacy cooldown `10` is metadata only; its dispatcher does not enforce cooldowns, so Go adds no command cooldown.

`/防詐騙網址` publicly defers, then requires Manage Messages at runtime. Its exact denial title is:

```txt
<a:Discord_AnimatedNo:1015989839809757295> | 你需要有`訊息管理`才能使用此指令
```

`/詐騙網址回報` is public and has no runtime permission check.

Usage belongs only to the global slash middleware. With `MHCAT_FEATURE_USAGE_TRACKING_ENABLED=true`, it records one best-effort attempt before route lookup and validation. Toggle/report handlers and MessageCreate events do not write usage directly.

## Toggle UI And Data

The success embed preserves:

- title `<:fraudalert:1000408260777611355> 自動偵測詐騙連結`;
- description `您的防詐騙啟用狀態已改為:\ntrue` or `false`;
- discord.js named `Green`, `0x57F287`;
- public deferred edit behavior.

A missing `good_webs` row creates `{guild:<id>,open:true}`. An existing row toggles its Mongoose-decoded Boolean. Boolean `true`, numeric `1`, and strings `true`, `1`, or `yes` decode enabled; false/zero/no forms and unusable values decode disabled.

Legacy deletes the first row and starts an unawaited replacement insert. Go uses one upserting `UpdateMany` with typed `$set.open`, updates all duplicate `{guild}` matches, preserves unrelated fields, and creates one row only when no row matches. This avoids both a delete/reinsert gap and a second Mongo round trip while retaining rollback-readable fields.

## Report UI And Validation

The command preserves required string option `網址` with description `回報網址`, raw input text, and public defer/edit behavior.

The validator matches pinned `is-url` 1.2.4, not modern URL parsing. It accepts protocol-relative URLs, arbitrary ASCII word-character schemes, localhost forms, and qualifying Unicode domains. It uses JavaScript whitespace classification and UTF-16 suffix length, so `//a.😀` is accepted while `//a.é` is rejected. Surrounding whitespace is not trimmed into validity.

Invalid input returns:

```txt
<a:Discord_AnimatedNo:1015989839809757295> | 你輸入的不是一個網址!
```

A known stored URL returns:

```txt
<a:Discord_AnimatedNo:1015989839809757295> | 該網站已被回報過
```

Success preserves the fraud-alert title, green color, and description `成功回報<raw-url>`. The webhook content remains:

````txt
```<raw-url>```
by:<@reporter-id>
````

The outer fence above is illustrative; the actual payload contains one triple-backtick URL block followed by the reporter mention.

Go escapes the user-controlled Mongo regex while preserving the legacy lookup direction: a stored `web` containing the reported raw URL counts as known. It sanitizes triple-backtick breakout and awaits a successful 2xx webhook response before showing success. It never inserts into `not_a_good_webs`.

## Message Delete Contract

The event runtime ignores DMs, missing IDs, empty/whitespace-only content, missing config, disabled config, and content with no catalog substring. Like legacy, it does not exclude bot authors.

For a match, Go deletes the original message and then sends exact content:

```txt
<:trashbin:995991389043163257> | 此消息包含詐騙或釣魚連結，以自動刪除!
```

The stored string includes a trailing newline. The warning suppresses mentions.

Legacy queries `{web:{$regex:message.content}}`, which reverses the intended search and misses ordinary messages containing a listed URL. Go intentionally scans catalog values and checks `content.includes(web)`. Raw message and stored catalog whitespace are preserved rather than normalized.

Deletion is awaited before warning. A delete failure sends no misleading warning. A warning failure does not restore the already-deleted message. Bot-authored matches can therefore still delete bot output, preserving the legacy recursion risk if the catalog itself matches warning/bot content.

## Mongo Compatibility

Collections and fields remain exact:

- `good_webs`: `guild`, `open`;
- `not_a_good_webs`: `web`.

`good_webs` writes use typed BSON string/Boolean fields readable by Mongoose. `not_a_good_webs` is read-only in Go. Its read DTO applies Mongoose-compatible String scalar coercion; null and compound object/array values remain unusable and are skipped safely. Raw usable values are not trimmed or normalized.

The application creates no startup index and runs no repair. Candidate unique indexes on `good_webs.guild` and `not_a_good_webs.web` remain blocked on duplicate keys, missing/null keys, scalar drift, raw URL variants, external catalog writers, and ownership review. No canonical URL field or backfill is approved.

## Intentional Safety And Reliability Differences

- User-controlled duplicate regex input is escaped.
- Ordinary messages containing a catalog URL are detected despite the legacy reversed query.
- Webhook code-block breakout is neutralized and delivery is awaited.
- Backend errors return a controlled interaction error instead of leaking raw errors and webhook/database secrets into the channel.
- Toggle writes align duplicate rows and preserve unrelated fields instead of delete/reinsert.
- Message deletion is awaited before warning; failed deletion emits no success-like warning.
- Empty/whitespace-only messages are retained instead of reproducing the legacy empty-regex edge case.
- Empty allowed mentions are explicit on interaction responses and deletion warnings.

Exact command metadata, stable embeds/messages, permission behavior, raw input, successful legacy URL classification, webhook payload shape, bot scanning, trailing newline, collection/field names, and typed rollback compatibility are preserved.

## Runtime Ownership And Migration

Before enabling any family:

1. Stop or gate the corresponding Node command/event owner for the target bot and guild.
2. Audit duplicate `good_webs.guild` and duplicate `not_a_good_webs.web` values.
3. Audit missing/null/non-string keys, Boolean drift in `open`, compound `web` values, raw URL variants, regex metacharacters, and external/dashboard catalog writers.
4. Preserve both collections and all fields. Do not normalize URLs, backfill canonical fields, deduplicate, or create indexes during ownership transfer.
5. Use an isolated staging guild/database and a non-production webhook endpoint.
6. Enable only the required config, report, or deletion gates and run preflight plus command-sync dry-run before apply.
7. For deletion, complete privacy review for Message Content and verify Node/Go never process the same message.

## Parity Tests

Focused tests lock definitions, permissions, exact UI/colors, pinned URL classification, raw values, duplicate filters, Mongoose scalar reads, typed duplicate-safe writes, webhook formatting/sanitization/status handling, global usage ownership, bot-message scanning, delete/warning ordering, app wiring, gates, command sync, and staging preflight. Run:

```bash
go test ./internal/core/domain ./internal/core/services/safety ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/adapters/external ./internal/discord/features/safety ./internal/app ./internal/config ./internal/parity ./cmd/mhcat-command-sync ./cmd/mhcat-staging-preflight
go test -race ./internal/core/services/safety ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/discord/features/safety ./internal/app ./cmd/mhcat-staging-preflight
go vet ./internal/core/domain ./internal/core/services/safety ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/adapters/external ./internal/discord/features/safety ./internal/app ./cmd/mhcat-staging-preflight
```

## Staging Smoke

1. Use an isolated guild/database with a manager, non-manager, safe webhook endpoint, disposable channel, and no active Node anti-scam owner.
2. Audit duplicate/type/raw-URL findings in both collections and confirm no repair/index apply is planned.
3. Pair config/report runtime and sync flags as needed; run preflight and command-sync dry-run before staging apply.
4. Verify both commands remain publicly discoverable. Confirm only toggle requires Manage Messages and its denial is public.
5. Run toggle create/true, existing/false, duplicate-row alignment, numeric/string Boolean seed, and non-manager denial cases; compare exact embeds and BSON fields.
6. Test valid `https`, `ftp`, custom scheme, protocol-relative, localhost, Unicode/emoji suffix, surrounding whitespace, short suffix, and malformed URL cases.
7. Verify a new report sends exact webhook content and success UI; verify known URL duplicate UI, escaped regex metacharacters, code-block input, webhook non-2xx/timeout, and no `not_a_good_webs` writes.
8. Enable deletion with all gateway flags. Test exact and embedded catalog URLs, raw whitespace, bot-authored content, clean/DM/disabled events, and the exact trailing-newline warning.
9. Deny delete permission and verify no warning; deny warning-send permission and verify deletion remains. Confirm Node and Go never process the same message.
10. With usage tracking enabled separately, verify one increment per slash attempt and none for MessageCreate events.
11. Disable gates, remove only managed staging commands, preserve data, and execute rollback checks.

## Rollback

1. Disable both anti-scam command-sync include flags and remove only those managed staging commands.
2. Disable Go config/report/deletion runtime gates. Keep Node owners stopped until no Go process can receive the same command or MessageCreate family.
3. Disable Message Content/Guild Messages only when no other enabled feature requires them.
4. Preserve `good_webs` and `not_a_good_webs`; typed Go toggle writes remain Mongoose-readable. Do not normalize values or mutate indexes during emergency rollback.
5. Restore only the intended Node command/event branches and verify one toggle, one safe report, and one exact catalog-message deletion in staging.
6. Review messages processed during any ownership overlap because duplicate deletion/warnings cannot be transactionally undone.
7. Re-enable production ownership only after confirming the alternate runtime is stopped.
