# Statistics Channels Parity Contract

Status: parity-audited against the active legacy statistics slash commands, Mongoose 6.4.6 String hydration, discord.js 14.25.1 cache/type behavior, `handler/channel_status.js`, and the current Go definitions, handlers, services, Discord adapter, Mongo repository, app/config/preflight wiring, real-Mongo lifecycle tests, and race coverage. Every runtime remains independently disabled by default. Live Discord staging and exclusive rename ownership are still required before production rollout.

## Scope

This contract covers:

- `/統計系統查詢` static help;
- `/統計系統創建` base and optional counters;
- `/統計身分組人數` role counters;
- `/統計系統刪除` config deletion;
- the 20-minute `numbers`/`role_numbers` channel rename worker.

Legacy sources:

- `slashCommands/統計系統/number.js`;
- `slashCommands/統計系統/number_create.js`;
- `slashCommands/統計系統/role_create.js`;
- `slashCommands/統計系統/number_delete.js`;
- `handler/channel_status.js`;
- `models/Number.js` and `models/role.js`.

The earlier implementation overview is [57-stats-query.md](57-stats-query.md). This document is canonical for compatibility, ownership, migration, staging, and rollback.

## Gates And Usage Ownership

Each command has its own paired runtime and staging command-sync gates:

```bash
MHCAT_FEATURE_STATS_QUERY_ENABLED=true
MHCAT_COMMAND_SYNC_INCLUDE_STATS_QUERY=true

MHCAT_FEATURE_STATS_CREATE_ENABLED=true
MHCAT_COMMAND_SYNC_INCLUDE_STATS_CREATE=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true

MHCAT_FEATURE_STATS_ROLE_COUNT_ENABLED=true
MHCAT_COMMAND_SYNC_INCLUDE_STATS_ROLE_COUNT=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true

MHCAT_FEATURE_STATS_DELETE_ENABLED=true
MHCAT_COMMAND_SYNC_INCLUDE_STATS_DELETE=true
```

The event-only rename worker has no command-sync flag:

```bash
MHCAT_FEATURE_STATS_RENAME_WORKER_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true
```

Create and role count need the guild/member/channel caches populated by Gateway plus Guild Members intent. Query and delete do not. Config commands never enable the worker, and the worker never enables or registers commands.

The global slash middleware is the only production usage owner. When `MHCAT_FEATURE_USAGE_TRACKING_ENABLED=true`, it writes exactly one `all_use_counts` event before route lookup, permission checks, and handler work, including denied, invalid, missing-data, backend-failure, and success attempts. Runtime wiring nils each stats module's feature tracker to prevent a second write. The rename worker emits no usage event.

## Command Definitions

| Command | Description | Effective command permission | Options |
| --- | --- | --- | --- |
| `統計系統查詢` | `查詢統計消息` | public | none |
| `統計系統創建` | `創建統計消息` | Manage Messages (`8192`) | required string `統計頻道類型`: `文字頻道` or `語音頻道`; optional string `統計選項`: `頻道數量`, `文字頻道數量`, or `語音頻道數量` |
| `統計身分組人數` | `統計某個特定的身分組的人數` | Manage Messages (`8192`) | required channel-type string as above; required role `身分組` |
| `統計系統刪除` | `刪除統計消息` | Manage Messages (`8192`) | none |

Create, role count, and delete also repeat the Manage Messages check at runtime, including Discord's Administrator override. Legacy cooldown metadata is 10 seconds except create at 30 seconds; the legacy global dispatcher did not centrally enforce those values, and Go does not add a new cooldown.

The 74-command static audit locks exact names, descriptions, option types/order, required flags, choices, and default permissions.

## Static Query UI

Query replies publicly without deferring. It sends one random-color embed titled `統計系統查詢` with the exact legacy whitespace and text:

````text

        我的統計系統是每**10分鐘更新一次**`(因為discord api每10分鐘才能更新一次)`
        輸入 /統計系統創建 [選擇要`文字頻道`或是`語音頻道`] [輸入想創建的統計名稱]

        **用戶查詢**
        ```
用戶總數 (伺服器的總人數)
使用者總數 (伺服器非機器人人數)
機器人數 (伺服器總共的機器人數量)```
        **伺服器頻道**
        ```
頻道數量 (頻道總數量)
文字頻道數量 (文字頻道總數)
語音頻道數量 (語音頻道總數)```

````

The actual description's two visually blank interior lines each contain eight spaces; they are shown empty above to avoid repository trailing-whitespace damage. The visible claim says ten minutes even though the active worker interval is twenty minutes. That inconsistency is legacy UI and remains unchanged. Query reads no stats collection and performs no Discord channel operation. Only optional global usage tracking can write Mongo for this attempt.

## Base Creation

Create defers publicly. A permission/type failure sends a public follow-up. On first creation it sends a green loading follow-up titled `<a:lodding:980493229592043581> | 正在進行設置中!`, creates the category and three counters, then edits that same follow-up to:

```text
<a:greentick:980496858445135893> | 成功創建!頻道(不要動到數字就沒問題)跟類別的名稱都能自行更改喔!
```

When a base config already exists, an optional counter skips the loading message and sends the same success embed as a new follow-up after completion.

The base category name is exactly `伺服器統計數據(這串可隨便改)`. Counter names are:

| Counter | Text channel | Voice channel |
| --- | --- | --- |
| all members | `總人數: <count>` | `總人數:<count>` |
| non-bot members | `總成員: <count>` | `總成員:<count>` |
| bots | `總BOT數: <count>` | `總BOT數:<count>` |

Counts come from the cached guild state:

- all members is `guild.memberCount`;
- bots is the number of cached members whose user is a bot;
- non-bots is `guild.memberCount - cached bot count`.

This intentionally permits member-count/cache disagreement, matching legacy.

Text-channel overwrites are ordered bot member then `@everyone` role. The bot gets View Channel, Manage Messages, and Send Messages; everyone gets View Channel and is denied Send Messages. Voice counters give the bot View Channel, Manage Messages, and Connect; everyone gets View Channel and is denied Connect. Go uses the interaction application ID for the current bot overwrite, with configured bot ID only as fallback.

## Initial `numbers` Shape

The first successful creation writes typed String-compatible values for:

```text
guild, parent,
memberNumber, memberNumber_name,
userNumber, userNumber_name,
BotNumber, BotNumber_name
```

It also writes explicit BSON null for all 38 legacy optional fields:

```text
channelnumber, channelnumber_name,
textnumber, textnumber_name,
voicenumber, voicenumber_name,
categoriesnumber, categoriesnumber_name,
rolesnumber, rolesnumber_name,
rolenumber, rolenumber_name,
norolenumber, norolenumber_name,
emojisnumber, emojisnumber_name,
staticnumber, staticnumber_name,
gifnumber, gifnumber_name,
stickersnumber, stickersnumber_name,
boostsnumber, boostsnumber_name,
tier, tier_name,
onlinenumber, onlinenumber_name,
dndnumber, dndnumber_name,
idlenumber, idlenumber_name,
offlinenumber, offlinenumber_name,
onlinebotnumber, onlinebotnumber_name,
statusnumber, statusnumber_name
```

Go uses a patch/upsert and preserves unrelated fields on an existing exact guild row. Repository construction and app startup create no collection or index.

## Optional Counters

Once any `numbers` row is selected for the guild, `統計選項` is required. The selected optional channel field is checked using JavaScript truthiness after Mongoose hydration: any non-empty String, including whitespace-only text, means already configured.

The new names are:

| Option | Name | Displayed value |
| --- | --- | --- |
| `頻道數量` | `總頻道數: <count>` | every cached channel, including categories |
| `文字頻道數量` | `總文字頻道數: <count>` | zero |
| `語音頻道數量` | `總語音頻道數: <count>` | zero |

Legacy discord.js v14 compared numeric channel types with stale strings. Therefore total includes every cached channel, while text and voice counts are both zero. Go preserves this defect.

The channel uses the requested text/voice type and corresponding base overwrite shape. It is attached to the stored parent only when that exact raw parent ID exists in the guild cache; whitespace is not trimmed. Missing parents create an unparented channel.

The selected ID/name pair is patched together. Voice creation displays the current voice count but intentionally stores the text-count value in `voicenumber_name`, reproducing the legacy variable bug.

Exact controlled error titles append these texts to the standard red animated-no prefix:

- permission: `你需要有\`訊息管理\`才能使用此指令`;
- invalid type: `你沒有進行設置要文字頻道還是語音頻道!或是你打錯了!`;
- missing optional choice: `由於你已經創建過了，所以你必須說明你要創建的統計名稱，或是刪除現有的統計資料(使用統計資料刪除)!`;
- already configured: `這個統計你已經創建過了!`;
- unknown option: `沒有這項統計可以創建欸QQ`;
- backend/Discord failure: `很抱歉，出現了未知的錯誤，請重試!`.

## Role Counter

Role count defers publicly, immediately creates the same green loading follow-up, then edits that follow-up for permission, validation, missing-config, or success results. It requires one selected `numbers` base row.

The member count is cache-only. The selected `@everyone` role counts all cached guild members. The channel name is exactly `<role name>: <count>` for both text and voice types. Stored parent lookup uses the exact raw ID and otherwise leaves the channel unparented.

Text role counters intentionally differ from base text counters: the bot overwrite grants View Channel, **Manage Channels**, and Send Messages. Voice role counters grant View Channel, Manage Messages, and Connect. The everyone overwrite matches the corresponding base channel.

The write shape is exactly:

```javascript
{ guild: "<guild>", channel: "<new channel>", channel_name: "<count>", role: "<role>" }
```

Go removes all exact `{guild,role}` duplicates before inserting one typed replacement. Like legacy delete-plus-save, replacement is non-transactional; an insert failure after deletion can leave no config, and the newly created Discord channel is not automatically removed.

Success uses title `統計特定身分組成功創建` and exact description `已成功為您創建統計特定身分組\n頻道:<#<channel>> 名字可以更改喔，不要動到數字就好awa`. Invalid type uses the create error above; a missing base uses `你還沒創建過統計頻道，請先使用\`/統計系統創建\``; other failures use the generic error.

## Delete

Delete defers publicly and edits the original response. It reads one arbitrary exact-string guild row, retains that row's raw parent value for UI, then removes all exact guild duplicates. It never deletes the category, counter channels, or any `role_numbers` row.

Success title is:

```text
<a:greentick:980496858445135893> | 成功刪除，該類別以下的頻道我已經管不了囉!(類別id:<raw parent>)
```

Missing config uses `你還沒有創建過統計數據，是要刪除甚麼啦!`; permission and generic errors use the common red titles. Legacy marked edit payloads `ephemeral:true` after a public defer, which cannot change the original audience; Go preserves the effective public response.

Removing all duplicate `numbers` rows instead of one arbitrary row is an intentional cleanup difference.

## BSON Compatibility

Every field in both schemas is Mongoose `String`. Go read DTOs therefore accept the established Mongoose scalar forms: BSON String/Symbol, Boolean, JavaScript-formatted numeric types, Decimal128, ObjectID hex, unsigned Timestamp decimal, binary UTF-8, JavaScript Date text, and regular-expression text. Missing, undefined, null, arrays, documents, JavaScript/Code, MinKey, MaxKey, and other unsupported compounds become unusable empty values without aborting a broad worker cursor.

Stored values are not trimmed. This matters behaviorally:

- spaced guild/channel/role IDs remain Discord cache misses;
- a spaced parent does not attach a new channel;
- whitespace-only optional IDs remain truthy/configured;
- rename replacement searches for the exact raw old-counter substring;
- delete success displays the raw selected parent value.

Command inputs and newly generated Discord IDs remain normalized typed strings. Exact Mongo query filters remain BSON strings, so a numeric stored `guild` that hydrates during an unfiltered worker scan does not match a later string `{guild}` update filter.

Malformed rows remain row-local in worker lists. No read, startup, or worker construction normalizes or rewrites them.

## Rename Worker

The worker waits twenty minutes before its first run and repeats every twenty minutes. It has no lease. A process serializes its own runs, but multiple Go processes and Node do not coordinate; exclusive deployment ownership is mandatory.

For each base row, it obtains cached guild stats and considers these six optional channel IDs:

```text
memberNumber, userNumber, BotNumber,
channelnumber, textnumber, voicenumber
```

For each valid cached channel, the new name replaces only the first exact occurrence of the stored old-counter String. If the old String is absent or empty, the entire channel name becomes the new decimal count. Stored whitespace is significant. No REST lookup fills a cache miss.

Role rows use cached role/member state. A deleted role has zero matching cached members and therefore renames to zero. A missing guild/channel, malformed identity, or Discord failure is skipped and logged without stopping later valid rows. Go intentionally continues after a missing role channel where legacy's callback `return` stopped the remaining role scan.

After a successful rename or a same-name decision, Go batches changed base counter fields into one `$set` and updates role `channel_name` separately. It avoids a redundant Discord PATCH when the computed name is already current and does not advance stored counters after a missing channel or failed rename. Legacy launched unawaited independent writes and could advance counters around failed/overlapping operations.

Counter updates retain legacy arbitrary duplicate selection: base writes filter by `{guild}`, role writes by `{guild,role}`. The worker scans all duplicate documents, but each update can target Mongo's arbitrary first matching row. Do not create a unique index or deduplicate without a production audit and migration decision.

## Intentional Differences

Go intentionally differs by:

- awaiting channel and Mongo operations and returning controlled failures;
- using patch/upsert for base creation;
- patching each optional ID/name pair together;
- removing all exact `{guild,role}` rows before role replacement;
- removing all exact guild rows on delete;
- serializing each process's worker runs;
- batching base counter writes;
- continuing after malformed/missing role rows;
- skipping stored-counter advancement after missing/failed channel renames;
- avoiding redundant same-name Discord PATCH calls;
- suppressing mentions in command responses;
- using the runtime application ID for the bot overwrite.

Go deliberately preserves command UI, count/cache quirks, channel names/types/overwrites, null field shape, voice stored-value bug, raw String whitespace, arbitrary duplicate worker updates, twenty-minute cadence, first-match rename, deleted-role zero, and the absence of channel deletion.

## Data And Indexes

No schema migration, backfill, repair, normalization, collection rename, startup write, startup index, or distributed lease is required. Dashboard backup/export already includes `numbers` and `role_numbers`.

Before staging or index work, audit:

- database and exact plural collection names;
- every field's BSON type, null/missing/compound values, and raw whitespace;
- duplicate `numbers.guild`, `role_numbers.{guild,role}`, and `role_numbers.{guild,channel}` keys;
- stale/missing/cross-guild category, channel, and role IDs;
- unknown fields and dashboard/backup writers;
- existing indexes and every active Node/Go rename process.

Startup and repository construction still create only MongoDB's `_id_` index. Duplicate-safe non-unique `numbers_guild_lookup` and `role_numbers_guild_role_lookup` indexes may be explicitly applied for config and rename traffic without changing first-match behavior. Candidate unique `{guild:1}`, `{guild:1,role:1}`, and `{guild:1,channel:1}` indexes remain blocked until live duplicate/type/ownership audits and explicit review; remove a same-key lookup fallback before promotion. Unique indexes would alter observable first-match/duplicate behavior and can break Node rollback.

## Verification

Run:

```bash
go test ./internal/discord/features/stats \
  ./internal/core/services/stats ./internal/core/domain \
  ./internal/adapters/mongo/documents \
  ./internal/adapters/mongo/repositories \
  ./internal/adapters/discordgo ./internal/app ./internal/config \
  ./cmd/mhcat-staging-preflight

go test -race ./internal/discord/features/stats \
  ./internal/core/services/stats ./internal/core/domain \
  ./internal/adapters/mongo/documents \
  ./internal/adapters/mongo/repositories \
  ./internal/adapters/discordgo ./internal/app

go vet ./internal/discord/features/stats \
  ./internal/core/services/stats ./internal/core/domain \
  ./internal/adapters/mongo/documents \
  ./internal/adapters/mongo/repositories \
  ./internal/adapters/discordgo ./internal/app

go run ./tools/parity-audit --legacy-root ../MHCAT --format markdown
```

The static audit must remain `74/74`. Opt-in real-Mongo tests prove scalar hydration, malformed-row isolation, startup no-mutation, base/optional patch preservation, duplicate delete/replacement behavior, guild isolation, null shape, and default-only indexes.

## Staging Smoke

1. Use an isolated guild/database. Snapshot `numbers`, `role_numbers`, current channel/role IDs, and indexes; stop Node `channel_status.js` and every extra Go worker.
2. Enable/sync query only. Verify the exact public static embed, one optional global usage event, and no stats collection/channel mutation.
3. Enable/sync create with Gateway and Guild Members. Exercise text and voice base creation; verify loading/edit order, names, overwrite order/bitfields, cached member/bot counts, full BSON null shape, and current application bot ID.
4. Exercise each optional counter, missing choice, duplicate including whitespace-only ID, invalid type, missing/spaced parent, Discord failure, and Mongo failure. Verify total/text/voice stale-type counts and voice stored-value bug.
5. Enable/sync role count. Exercise text, voice, `@everyone`, missing base, missing/spaced parent, duplicate `{guild,role}`, invalid type, permission denial, and write failure. Verify loading message edit, exact success/error UI, channel shape, and replacement data.
6. Enable/sync delete against copied rows. Verify raw first parent in success, all same-guild `numbers` duplicates removed, other guild and all `role_numbers` rows retained, and no Discord channel deletion.
7. Confirm exactly one global usage event for success, denial, validation failure, missing config, and backend failure when tracking is enabled.
8. Seed valid, scalar, whitespace, null, compound-malformed, duplicate, missing-channel, deleted-role, and cross-guild fixtures for both collections.
9. Enable only the rename worker and wait the full first twenty-minute interval. Verify exact first-substring replacement, whole-name fallback, stale count semantics, deleted-role zero, cache-only misses, row-local continuation, and expected arbitrary duplicate update boundary.
10. Force no-op names, Discord failures, Mongo update failures, cancellation, and shutdown. Verify no redundant PATCH, no counter advance after failed/missing channels, no overlapping local run, and prompt shutdown.
11. Confirm no new indexes, collections, schema fields, normalization, usage from the worker, or command routes outside the explicitly enabled gates.

## Rollback

1. Disable `MHCAT_FEATURE_STATS_RENAME_WORKER_ENABLED` and stop all Go workers; wait for any in-flight run to finish.
2. Disable stats command runtimes and command-sync inclusion as required. Do not delete guild commands until the reviewed sync plan says to do so.
3. Inspect channels whose rename or config write may have completed before failure. Reconcile raw Mongo counters against current names before retrying.
4. Preserve legacy `numbers`/`role_numbers` and unknown fields. Restore only backed-up disposable rows; do not normalize or create/drop indexes during rollback.
5. Restore Node `channel_status.js` only after no Go worker can rename channels.
6. Smoke one base and one role counter over a full legacy interval and verify one owner performs each rename.

Production rollout remains blocked on live Discord smoke, a read-only BSON/duplicate/index/channel audit, exclusive Node/Go worker ownership, and explicit acceptance of the intentional reliability differences above.
