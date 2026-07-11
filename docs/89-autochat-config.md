# Auto-Chat Configuration Parity Contract

Status: parity-audited against the active legacy set/delete slash commands, global slash dispatcher, Mongoose 6.4.6 String hydration and document deletion, Mongo first-match behavior, current Go document/repository/service/handler/runtime wiring, command sync, and staging preflight. Runtime and command sync remain disabled by default. Live staging smoke is still required before production ownership.

## Scope

This contract covers:

- `/自動聊天頻道` and `/自動聊天頻道刪除` metadata, permissions, usage, public UI, and failures;
- `chats` lookup, one-row replacement/deletion, duplicates, mixed BSON, partial progress, staging, and rollback.

Legacy sources:

- `slashCommands/實用工具/chat.js`
- `slashCommands/實用工具/chat_delete.js`
- `models/chat.js`
- `events/SlashCommands.js`
- Mongoose 6.4.6
- discord.js 14.25.1

The local corpus MessageCreate handler and paid ChatGPT handoff are separate runtime contracts. These config commands do not enable Gateway events, Message Content, balance reads/debits, `chatgpts`, or the external worker.

## Gates And Ownership

Enable only with paired staging flags:

```bash
MHCAT_FEATURE_AUTOCHAT_CONFIG_ENABLED=true
MHCAT_STAGING_MODE=true
MHCAT_COMMAND_SYNC_INCLUDE_AUTOCHAT_CONFIG=true
```

Both feature flags default to false. Command sync is guild-scoped and staging-only. Preflight rejects sync without runtime and warns when runtime is enabled without sync. Runtime startup opens the existing `chats` collection and registers both routes only when the repository is available.

Stop the Node owners for both config commands before enabling Go for the same bot/guild. There is no shared lease. Concurrent owners can compete to acknowledge the same interaction and can replace different duplicate rows.

The commands require normal Gateway interaction delivery and Mongo read/write access. They require no Message Content, Guild Members, reaction, or voice intent. They are publicly discoverable but enforce Manage Messages at runtime.

## Definitions And Usage

`/自動聊天頻道` remains:

- chat-input name `自動聊天頻道`;
- description `設定自動聊天頻道要在哪裡發送`;
- docs URL `https://docsmhcat.yorukot.me/docs/chat_xp_set`;
- one required channel option `頻道`, description `輸入頻道!`;
- channel types `0` and `5` only, Discord guild text and announcement/news channels;
- no Discord default-member permission.

`/自動聊天頻道刪除` remains:

- chat-input name `自動聊天頻道刪除`;
- description `刪除自動聊天頻道要在哪裡發送`;
- the same docs URL;
- no options and no Discord default-member permission.

Both commands require Manage Messages (`8192`) at runtime. Discord Administrator retains its normal override. Legacy cooldown metadata is `10`, but the global dispatcher does not enforce cooldowns. Go adds no cooldown.

Usage belongs only to global slash middleware. With tracking enabled, every routed success, permission denial, missing config, validation failure, or backend failure records exactly one best-effort command increment before handler work. Production wiring removes the handler-level tracker to prevent a second success-only event.

## Exact UI

Both commands defer publicly and edit the deferred original response. No ephemeral flag is set. Mentions are explicitly suppressed.

Set success contains one embed:

- title `自動聊天系統`;
- description `您的自動聊天頻道成功創建\n您目前的自動聊天頻道為 <#<channel-id>>`;
- discord.js named `Green`, exact `0x57F287`;
- no author, footer, fields, components, or files.

Delete success contains one embed:

- title `自動聊天系統`;
- description `您的自動聊天頻道成功刪除`;
- exact `0x57F287`;
- no author, footer, fields, components, or files.

Permission denial uses one `0xED4245` red embed with exact title:

```text
<a:Discord_AnimatedNo:1015989839809757295> | 你需要有`訊息管理`才能使用此指令
```

Deleting when no matching row exists uses:

```text
<a:Discord_AnimatedNo:1015989839809757295> | 你沒有設定過，我不知道要刪除甚麼!
```

Controlled validation/backend/write failures use:

```text
<a:Discord_AnimatedNo:1015989839809757295> | 很抱歉，出現了未知的錯誤，請重試!
```

Raw Mongo errors, credentials, query details, and stored values are never shown.

## Mongo Read Contract

The Mongoose model name `chat` resolves to collection `chats` with schema fields `guild: String` and `channel: String`. Both owners query one row by exact `{guild: <interaction/event guild ID>}` with no sort. Duplicate rows therefore retain arbitrary first-match behavior.

Go mirrors relevant Mongoose String hydration for `channel`:

| Stored `channel` | Hydrated routing value |
| --- | --- |
| string | exact string, including whitespace |
| integer/Decimal128/numeric double | JavaScript/Mongoose number string |
| boolean | `true` / `false` |
| ObjectID | lowercase hexadecimal string |
| BSON binary/Buffer | UTF-8 string |
| NaN/infinity | `NaN` / `Infinity` forms |
| null/missing/compound array or document | unusable empty value |

The configured channel is compared to the Discord channel ID without trimming. A padded stored string such as ` channel-1 ` remains inactive, matching legacy strict `!==`. Null and malformed compound values are also inactive rather than backend errors. Numeric or binary values containing an exact channel ID remain usable after Mongoose-compatible coercion.

The config commands themselves write normal Discord guild and channel IDs as BSON strings. Mixed-value read compatibility exists for migrated or externally modified legacy rows; it does not normalize them at startup.

## Set Replacement Contract

Legacy set performs one unsorted `findOne({guild})`, calls document-instance delete when a row exists, and constructs a fresh model containing only `guild` and `channel`.

Go preserves that data shape while awaiting each operation:

1. find one arbitrary matching row;
2. when present, delete that exact fetched `_id`;
3. insert fresh `{guild, channel}` with a new `_id`;
4. leave every other duplicate row unchanged;
5. drop extra fields from the selected row because the replacement is minimal.

When no row exists, one fresh row is inserted. There is no update-many, upsert, deduplication, transaction, retry, repair, or index creation.

## Delete Contract

Delete performs one unsorted `findOne({guild})` and deletes only that fetched document identity. With duplicates, one arbitrary row is removed per successful command invocation; other duplicates remain and can continue to drive MessageCreate runtime selection. A subsequent delete reports missing only after no matching row remains.

This is intentionally different from deleting all guild rows. Operators must not use the command as a deduplication or complete cleanup tool.

## Failure And Partial Progress

Set is intentionally non-transactional:

| Failure point | Go state and UI |
| --- | --- |
| lookup fails | no confirmed write; generic red error |
| selected-row delete fails/loses race | no replacement; generic red error |
| insert fails after selected-row delete | selected row lost, duplicates unchanged; generic red error |
| Discord edit fails after insert | replacement completed; error returned to dispatcher |

Delete lookup/delete failures return generic red UI. A row that disappeared after lookup maps to missing behavior. Discord edit failure can occur after the row was already deleted.

Legacy does not await document deletion or save. It can send success before persistence, can insert despite a failed old-row delete, and commonly treats a delete lookup backend failure as missing. Go intentionally waits and fails visibly. Operations are not retried automatically because replay can replace or delete another duplicate row.

## Data And Migration

No automatic database migration is required or performed. Enabling config opens `chats` but creates no index, backfill, normalization, or deduplication.

Before staging, audit and snapshot:

- duplicate, missing, null, padded, numeric, binary, and compound `guild`/`channel` values;
- all current `chats` indexes;
- dashboard, Node, Go, and other external writers;
- which row currently wins unsorted `findOne` for each staging guild;
- whether every staging config row and target channel is disposable.

Do not deduplicate or create a unique `{guild:1}` index solely to enable config parity. A unique index can be considered only as a separate, owner-wide migration after config, local fallback, paid handoff, dashboard, and external writers agree on singleton policy and existing duplicates are explicitly reconciled. It is not required for these commands.

Optional global usage tracking may update `all_use_counts`. That generic collection is not auto-chat config state and requires no config migration.

## Intentional Differences

Intentional differences are limited to:

- Go awaits lookup/delete/insert and sends success only after persistence completes;
- backend/write failures receive controlled generic red UI instead of being ignored, mistaken for missing, or followed by premature success;
- a selected-row delete race aborts instead of inserting an additional duplicate;
- invalid synthetic/no-guild/channel interactions receive controlled error handling;
- mentions are explicitly suppressed.

Public definitions, channel types, docs URL, runtime permission and Administrator override, unenforced cooldown, public defer/edit lifecycle, exact embeds, arbitrary first match, one-row delete-plus-insert replacement, one-row deletion, duplicate preservation, fresh `_id`, dropped extra fields, and Mongoose channel coercion are preserved.

## Parity Tests

Run focused coverage:

```bash
go test ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/core/domain ./internal/core/ports ./internal/core/services/autochat ./internal/discord/features/autochat ./internal/app ./internal/config ./internal/parity ./cmd/mhcat-command-sync ./cmd/mhcat-staging-preflight
go test -race ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/core/services/autochat ./internal/discord/features/autochat ./internal/app
go vet ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/core/services/autochat ./internal/discord/features/autochat ./internal/app
go run ./tools/parity-audit --legacy-root ../MHCAT --format markdown
```

Run the opt-in integration tests only against disposable Mongo:

```bash
MHCAT_RUN_MONGO_INTEGRATION_TESTS=true \
MHCAT_MONGODB_URI='<disposable-uri>' \
go test ./internal/adapters/mongo/repositories \
  -run '^TestAutoChatConfigMongoIntegration' \
  -count=1
```

The integration tests create and drop only generated test databases. They lock one-row replacement/deletion, duplicate and other-guild preservation, fresh `_id`, dropped extra fields, malformed selected rows, and Mongoose-compatible channel hydration.

## Staging Smoke

1. Use an isolated staging guild/database, stop the Node command owners, and snapshot `chats`.
2. Audit types, duplicates, indexes, selected first matches, and external writers without repairing them. Use only disposable rows/channels.
3. Enable paired flags, run preflight and command-sync dry-run, review guild apply, and start one Go owner.
4. Confirm both commands are publicly discoverable with no default permission; verify exact definitions and channel type restrictions.
5. As a normal member, verify exact public red permission denial. Repeat as Manage Messages and Administrator users.
6. Set a disposable channel; verify exact green UI and a fresh minimal `{guild, channel}` row.
7. Seed duplicate rows with extra fields; set again and verify only one selected row is replaced while another remains unchanged.
8. Delete repeatedly; verify one arbitrary row per invocation, exact green UI, then exact missing red UI only after the final row is gone.
9. Exercise padded, numeric, binary, null, and compound channel fixtures against local/paid routing only when their separate gates are explicitly under test.
10. Force lookup/delete/insert failures and verify generic red UI, no raw detail, documented partial state, and exactly one usage increment for every outcome.

## Rollback

1. Disable command-sync inclusion and remove only the two managed staging commands.
2. Disable the config runtime gate and stop the Go owner before restoring Node.
3. Reconcile every staging attempt before restoring fixtures: success or error UI does not by itself prove final Mongo state.
4. Restore lost/replaced rows only from reviewed snapshots; preserve original `_id`, mixed values, duplicates, and extra fields where rollback requires them.
5. Do not create/drop indexes, normalize values, or deduplicate as part of emergency rollback.
6. Restore Node only after confirming no Go config routes remain, then smoke one disposable set/delete sequence under the restored owner.

Production ownership remains blocked on live staging smoke, exclusive command ownership, reviewed shared writers/types/duplicates, and an accepted repair procedure for non-transactional partial progress. No automatic migration or index is required.
