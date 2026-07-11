# Auto-Notification Parity Contract

Status: parity-audited against the active legacy setup/list/delete commands, modal and selector handlers, `cron-validator` 1.4.0, `cron-parser` 4.9.0, Mongoose 6.4.6 hydration, `handler/cron.js`, and the current Go config, repository, delivery, app, config, preflight, real-Mongo, and race coverage. Config and delivery remain independently disabled by default. Live staging and exclusive production ownership are still required before rollout.

## Scope

This contract covers:

- `automatic-notification`, `/自動通知列表`, and `/自動通知刪除`;
- the `cron_set*` setup modal and weekday/hour/minute wizard;
- `cron_sets` BSON compatibility and recurring Discord delivery;
- scheduler lease ownership, staging, reconciliation, and rollback.

Legacy sources:

- `slashCommands/自動通知/cron_set.js`;
- `slashCommands/自動通知/cron_list.js`;
- `slashCommands/自動通知/cron_delete.js`;
- the cron section of `events/modal.js`;
- `handler/cron.js`;
- `models/cron_set.js`.

Daily economy reset and work payout share legacy scheduler history but are separate contracts and ownership boundaries. The implementation overview remains in [66-auto-notification-config.md](66-auto-notification-config.md).

## Gates And Ownership

Config runtime and command sync require the paired gates:

```bash
MHCAT_FEATURE_AUTO_NOTIFICATION_CONFIG_ENABLED=true
MHCAT_COMMAND_SYNC_INCLUDE_AUTO_NOTIFICATION_CONFIG=true
```

Recurring delivery is independent and requires:

```bash
MHCAT_FEATURE_AUTO_NOTIFICATION_DELIVERY_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_SCHEDULER_LEASE_ENABLED=true
MHCAT_SCHEDULER_LEASE_OWNER=staging-auto-notification
MHCAT_SCHEDULER_LEASE_TTL=2m
MHCAT_SCHEDULER_LEASE_TIMEOUT=10s
```

Delivery does not require command sync, Guild Messages intent, or Message Content intent. Config does not activate delivery. Every Go process needs a distinct lease owner. The lease cannot coordinate with Node, so every process loading legacy `handler/cron.js` must be stopped before Go delivery starts.

The three slash handlers do not write usage themselves. The global slash middleware owns the single `all_use_counts` event for success and failure paths when usage tracking is enabled.

## Command Definitions And Permissions

All commands are guild-scoped chat-input commands with no effective Discord default-member permission restriction. Each handler performs the legacy runtime Manage Messages check, including Discord's Administrator override.

| Command | Description and localization | Options |
| --- | --- | --- |
| `automatic-notification` | `Set where automatic notification should be send`; names: `en-US=automatic-notification`, `zh-TW=自動通知`, `zh-CN=自动通知`; descriptions: `en-US=Set where automatic notification should be send`, `zh-TW=设置自动聊天频道要在哪发送` | required channel `channel`, description `Enter channel to send!`, localized names `channel`/`頻道`/`頻道`, localized descriptions `Enter channel to send!`/`輸入要發送的頻道!`/`输入要发送的频道`, channel types text (`0`) and announcement (`5`) |
| `自動通知列表` | `查看所有的自動通知列表` | none |
| `自動通知刪除` | `刪除之前設定的自動通知` | required string `id`, description `輸入要刪除的自動通知id!` |

All three retain the legacy tutorial URL `https://youtu.be/D43zPrZU5Fw`. Permission and validation failures use the red tutorial embed; permission denial title is exactly `<a:Discord_AnimatedNo:1015989839809757295> | 你需要有\`訊息管理\`才能使用此指令`.

## Setup Identity And Pending Row

When setup handling begins, Go generates the JavaScript-shaped current Unix-millisecond decimal ID. It checks the intended ten-schedule guild limit and exact BSON string identity, then inserts:

```javascript
{
  guild: "<guild id>",
  channel: "<selected channel id>",
  id: "<millisecond id>",
  cron: null,
  message: null
}
```

Duplicate detection uses an exact Mongo `{guild:<string>,id:<string>}` query. A numeric `id` that String-coerces to the same display value does not block the new string ID; a second exact string ID does. Completion patches only `cron` and `message`, preserving `channel` and unknown fields.

Go intentionally enforces the ten-row limit. Legacy displayed the limit error inside a nested callback but continued to insert.

## Modal And Direct Cron

The modal custom ID is the generated schedule ID and title is `自動發送通知系統!`. It contains these fields in order:

| Custom ID | Style | Required | Label |
| --- | --- | --- | --- |
| `cron_setcron` | short | yes | `請輸入corn表達式(如想用簡化版，請直接輸入取消或cancel就可以簡易設置corn)` |
| `cron_setmsg` | paragraph | no | `請輸入文字(如不輸入這項請務必輸入下面三項)` |
| `cron_setcolor` | short | no | `請輸入你的嵌入訊息顏色(如不輸入嵌入訊息相關，請務必輸入文字)` |
| `cron_settitle` | short | no | `請輸入你的嵌入標題(如不輸入嵌入訊息相關，請務必輸入文字)` |
| `cron_setcontent` | paragraph | no | `請輸入嵌入內文(如不輸入嵌入訊息相關，請務必輸入文字)` |

Input content, title, description, and cron whitespace remain raw. Direct cron first uses the legacy numeric five-field `cron-validator` grammar and then scheduler parsing. Aliases, descriptors, six-field forms, malformed expressions, exact `cancel`, and exact `取消` enter the wizard. The next two direct occurrences must be at least 15 minutes apart: exactly 15 minutes is accepted.

Weekday `7` means Sunday, including ranges, lists, and steps. Go canonicalizes Sunday to the scheduler's `0-6` representation only for parsing; it persists and lists the original direct expression unchanged.

## Wizard

The wizard preserves the public legacy weekday, hour, and minute option labels/order, selection limits, message text, and JavaScript-rounded relative deadline. It expires exactly five minutes after creation.

Go uses process-local owner-scoped state and versioned custom IDs:

```text
mhcat:v1:cron:week:state=<id>
mhcat:v1:cron:hour:state=<id>
mhcat:v1:cron:minute:state=<id>
```

Only the initiating user in the initiating guild may advance it. A different user, stale ID, expiry, or restart returns a controlled ephemeral retry error and performs no completion write.

Hours retain the legacy order `1` through `23`, then `0`; minutes are `0,5,...,55`. The final expression is `<minutes> <hours> * * <weekdays>`. Legacy repeatedly assigned the weekday field while iterating selections, so its effective output is all seven days as `*`, otherwise the selected weekday values joined together; Go reproduces that effective expression.

## Message Payload And Preview

At least plain content or an embed title/description must exist. A color alone is invalid. Plain content with a color but no title/description remains a plain payload because legacy only constructed the embed when `content || title` was truthy for the embed fields.

New writes use one of these rollback-compatible shapes:

```javascript
{ content: "<raw text>" }

{
  content: "<raw text or null>",
  embeds: [{ data: { title: "<raw title>", description: "<raw description>", color: <number> } }]
}
```

Accepted new color input is `Random`, `#RRGGBB`, or a Discord named color that also passed the legacy HTML-color validator. `Random` is resolved once and the numeric value is persisted. Existing numeric colors, string colors, and literal `Random` remain readable. Invalid CSS-only colors return the controlled legacy color error instead of reproducing the legacy `EmbedBuilder` exception.

Completion content is exactly:

```text
:white_check_mark:**以下是該自動通知id:**`<id>`
使用`/自動通知刪除 id:<id>`進行刪除
~~我只是個分隔線，下面是你的訊息預覽~~
```

The interaction completion is awaited. The separate preview send to the interaction channel is best-effort and does not roll back the completed row.

## List And Delete

List reads by exact guild. Rows with null or missing `cron` are abandoned drafts: Go omits them from the response and removes them. Legacy issued deletion for strict null rows but could still render the already-loaded draft once; the Go behavior is an intentional cleanup difference.

The list embed title is `<:list:992002476360343602> 以下是<guild name>的所有自動通知id`. Its description begins `輸入\`/自動通知刪除 id\`可進行刪除之前設定的自動通知` and preserves each stored cron's raw whitespace in the numbered `id`, `cron設定`, and channel rows.

Delete defers publicly and removes one arbitrary exact `{guild:<string>,id:<string>}` row. Duplicate identities are not bulk-deleted. Success is the green `自動通知系統` embed with description `<:trashbin:995991389043163257>成功刪除該自動通知`; a missing ID uses the red tutorial error. Guild scoping prevents another guild from listing or deleting the row.

## BSON Read Boundary

Legacy Mongoose selects delivery rows where:

- `guild` hydrates as a String;
- `cron` is not null or missing;
- `message` is an Object/Mixed value.

Go scans the same effective boundary. `guild` must be a BSON string. `cron`, `channel`, and `id` are read from raw BSON and use Mongoose-compatible scalar String coercion. Unsupported identity values invalidate only that row; they do not abort reconciliation.

`message` remains Mixed. Nested values are not Mongoose String-cast. Valid plain string content and valid legacy embed containers are decoded; a numeric nested `content` is not silently converted to text. A malformed `embeds` container is treated as absent so valid plain content can still send. Scalar, null, array, or otherwise invalid message roots are skipped row-locally.

No setup, list, delete, repository construction, or worker startup rewrites existing active rows. Startup creates no `cron_sets` collection and no indexes.

## Delivery And Routing

The worker interprets five-field schedules in fixed UTC+8 named `Asia/Taipei`. Persisted rows created outside setup may use any five-field expression accepted by the delivery parser. Invalid schedules and malformed rows are logged/skipped independently.

For each scheduled occurrence, the callback reloads one exact `{guild,id}` identity before sending. This makes config changes visible and prevents a deleted row from sending even before reconciliation. Duplicate active `{guild,id}` rows produce one in-memory schedule and one arbitrary reload/send, matching one-row repository ownership rather than scheduling duplicates.

Before sending, the target channel must exist in the Discord cache and its cached guild ID must equal the schedule guild. A malformed row cannot route into a channel owned by another guild. Missing, uncached, or cross-guild channels are skipped.

Delivery preserves valid plain content and legacy embed data, resolves compatible stored colors, and allows user, role, and everyone mentions like legacy `channel.send`. It does not enable replied-user mentions because delivery is not a reply.

## Lease Lifecycle

The worker uses `_id` and `lock_name` `auto-notification-delivery` in `mhcat_scheduler_locks`.

It:

1. acquires before scheduling any row;
2. reconciles immediately after acquisition;
3. renews while active;
4. reconciles at most every 30 seconds, or every one-third of a shorter TTL;
5. adds new rows, replaces changed cron expressions, and removes deleted rows;
6. cancels all in-memory entries on renewal failure, ownership loss, or expiry;
7. releases its exact owner/fence lease during graceful shutdown.

Another lease holder schedules nothing. Lease failures do not mutate `cron_sets`. The lease collection retains documents after release and relies only on MongoDB's default `_id` index.

## Intentional Differences

Go intentionally differs from legacy by:

- enforcing the intended ten-schedule guild limit;
- omitting and cleaning pending rows during list instead of rendering them once;
- scheduling duplicate `{guild,id}` rows once;
- awaiting config completion writes;
- reconciling immediately and periodically under a renewable lease;
- returning a controlled error for invalid CSS-only colors;
- including missing `cron` fields in draft cleanup, while legacy checked strict null;
- patching cron and message together instead of issuing two unawaited updates;
- using owner-scoped versioned wizard state instead of collision-prone generic IDs;
- requiring a cached guild-bound channel before delivery;
- isolating malformed BSON rows and payloads instead of failing the full scan.

The command definitions, runtime permission, direct cron grammar and spacing, 15-minute boundary, effective wizard output, payload shape, visible UI, one-row delete, UTC+8 schedule, and mention behavior remain compatible.

## Data And Migration

No schema migration, backfill, normalization, collection rename, startup write, or startup index is required. The dashboard backup/export contract already includes `cron_sets`.

Before staging, perform a read-only audit of:

- exact `cron_sets` collection and database names;
- null/missing drafts and malformed/mixed scalar fields;
- duplicate `{guild,id}` rows;
- channel ownership for every active row;
- message root/content/embed/color shapes;
- current indexes, which should remain only `_id_` unless separately approved;
- all Node and Go scheduler processes and lease owners.

Do not normalize production rows merely to enable Go. A `{guild:1,id:1}` index remains a non-unique candidate after audit because legacy permits duplicate IDs and one-row delete/reload semantics.

## Verification

Run the focused contract suites:

```bash
go test ./internal/discord/features/notifications \
  ./internal/core/services/notifications \
  ./internal/adapters/mongo/documents \
  ./internal/adapters/mongo/repositories \
  ./internal/app ./internal/config ./cmd/mhcat-staging-preflight

go test -race ./internal/discord/features/notifications \
  ./internal/core/services/notifications \
  ./internal/adapters/mongo/documents \
  ./internal/adapters/mongo/repositories ./internal/app

go vet ./internal/discord/features/notifications \
  ./internal/core/services/notifications \
  ./internal/adapters/mongo/documents \
  ./internal/adapters/mongo/repositories ./internal/app

go run ./tools/parity-audit --legacy-root ../MHCAT --format markdown
```

The parity audit must remain `74/74` with zero drift and zero missing names. Real-Mongo coverage locks startup no-mutation, draft cleanup, config lifecycle, patch preservation, one-row delete, guild isolation, exact ID type identity, malformed row isolation, and the default-only index state.

## Staging And Reconciliation

1. Use an isolated database, guild, and channel; snapshot `cron_sets` and `mhcat_scheduler_locks`.
2. Enable only config/sync first. Sync the three commands to the staging guild after dry-run and preflight review.
3. Exercise direct cron with raw whitespace, exactly/below 15 minutes, weekday `7`, invalid input, `cancel`, and `取消`.
4. Exercise all wizard stages, all-seven/sparse weekdays, cross-user access, expiry, and restart loss.
5. Verify exact success/error/list/delete UI, one global usage event per attempt, numeric persisted colors, best-effort preview, draft cleanup, exact string ID handling, and one-row delete.
6. Confirm config-only testing created no lease, recurring send, index, or unrelated row mutation.
7. Stop every Node `handler/cron.js` owner. Audit every active row before enabling delivery so no unintended channel can receive a message.
8. Seed one unique active fixture and enable Gateway, delivery, and lease gates with a unique owner. Require preflight's ownership/write warning.
9. Verify one immediate schedule registration, one UTC+8 send, allowed mentions, cached guild-bound routing, and the expected lease owner/fence.
10. Change cron/message, add/remove rows, and allow one reconciliation interval. Verify replacement/removal without duplicate sends.
11. Exercise malformed identities, cron, message roots, nested content/embeds/colors, duplicate identities, missing channels, and cross-guild channel IDs; valid rows must continue.
12. Start a second distinct Go owner and verify it schedules nothing. Force renewal loss/expiry and verify all entries stop before reacquisition.
13. Gracefully stop Go, verify release, remove only disposable fixtures, and confirm no new `cron_sets` index or startup mutation.

Reconcile any ambiguous send by schedule identity, lease owner/fence, target guild/channel, scheduled UTC+8 occurrence, application logs, and Discord message evidence before retrying or restoring another owner.

## Rollback

1. Disable `MHCAT_FEATURE_AUTO_NOTIFICATION_DELIVERY_ENABLED` and stop every Go delivery process.
2. Verify `auto-notification-delivery` is released or expired and no in-memory Go scheduler can send.
3. Disable config runtime and command-sync inclusion if config ownership is also rolling back.
4. Preserve active `cron_sets`; remove only documented disposable staging rows. No schema/index rollback is normally needed.
5. Restore Node only after Go delivery is quiescent and ownership is exclusive.
6. Smoke one known schedule under the restored owner and verify exactly one send in the correct guild/channel.

Production ownership remains blocked on isolated live smoke, a read-only row/index/channel audit, exclusive Node/Go ownership, unique Go lease owner values, and acceptance of the intentional differences above.
