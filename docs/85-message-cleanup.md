# Message Cleanup Parity Contract

Status: parity-audited against the active legacy slash command, installed discord.js 14.25.1 behavior, global slash dispatch, Discord message REST limits, runtime wiring, command sync, and staging preflight. Runtime and command sync remain disabled by default. Live staging smoke is still required before production ownership.

## Scope

This contract covers:

- `/刪除訊息` metadata, permission gates, count validation, optional user filtering, deletion, and UI;
- Discord pagination, batch size, age cutoff, operation timing, failures, usage, staging, and rollback.

Legacy sources:

- `slashCommands/管理系統/clear.js`
- `events/SlashCommands.js`
- `config.json`
- discord.js `TextBasedChannel.bulkDelete`

The feature does not use Mongo. Delete-data is a separate destructive configuration command.

## Gates And Ownership

Enable only with the paired staging flags:

```bash
MHCAT_FEATURE_MESSAGE_CLEANUP_ENABLED=true
MHCAT_STAGING_MODE=true
MHCAT_COMMAND_SYNC_INCLUDE_MESSAGE_CLEANUP=true
```

Both feature flags default to false. Command sync is guild-scoped and staging-only. Preflight rejects sync without runtime and warns when runtime is enabled without sync. Startup requires the Discord side-effect adapter but opens no feature repository and performs no startup deletion.

Stop the Node `/刪除訊息` owner before enabling Go for the same bot/guild. There is no lease or shared job identity, so concurrent owners can delete overlapping messages independently.

The command requires Discord Gateway interaction delivery, but no Guild Messages or Message Content privileged intent. Message retrieval is REST-based. The bot still needs channel access, Read Message History, and permission to delete the selected messages.

## Definition And Usage

The definition remains publicly discoverable with no Discord default-member permission:

- name `刪除訊息`;
- description `刪除大量訊息`;
- required integer `刪除數量`, description `設定要刪除幾個訊息(最高1000超過200需要管理者權限)(只能刪除14天內的消息)`;
- optional user `使用者`, description `選擇是否要刪除某個特定的使用者的訊息(如填選這項，第一項代表的將是檢測訊息數量)`.

Legacy cooldown metadata is `30`, but the global dispatcher does not enforce it. Go adds no cooldown.

Usage belongs only to global slash middleware. With tracking enabled, every denied, invalid, failed, or successful slash attempt records exactly one best-effort event before the handler. The cleanup handler never writes a second success-only event.

## Permission And Validation

The handler defers ephemerally before validation. It requires Manage Messages (`8192`) for every request. Counts above `200` additionally require Administrator (`8`). The exact permission label is:

`訊息管理(刪除超過200則需要有權限)`

Counts above `1000` return exact error `不可刪除超過1000則消息!!!!!`. That maximum check occurs before the above-200 Administrator check, after the universal Manage Messages check.

Legacy zero and negative behavior depends on the optional user:

- filtered zero schedules ten callbacks that each resolve to the same `0/0` success without fetching or deleting; Go returns one equivalent success immediately;
- unfiltered zero schedules no callbacks and leaves the deferred interaction unresolved;
- filtered negative values eventually issue invalid fetch limits and can leave asynchronous errors/responses;
- unfiltered negative values also leave the deferred interaction unresolved.

Go intentionally converts unfiltered zero and every negative count to the generic controlled error. Missing or malformed synthetic integer options use the same error.

## Exact UI

All responses remain ephemeral because the original defer is ephemeral. Errors use discord.js `Red` (`0xED4245`) and title:

`<a:Discord_AnimatedNo:1015989839809757295> | <error>`

Success uses the legacy configured `client.color.greate`, exact `#53FF53` (`0x53FF53`), not discord.js named `Green`. The title is:

`<a:green_tick:994529015652163614> | 清理完成!`

The exact description is:

```text
**成功清除:**`<deleted>`/`<requested>`
**<:deletebutton:981971559679950848> 如果沒有成功清完全
代表可能超過14天或沒這麼多訊息給清**
```

Mentions are suppressed. Discord or adapter failures map to generic error `很抱歉，出現了未知的錯誤，請重試!` without exposing API details.

## Unfiltered Deletion

Without `使用者`, the request targets recent channel messages regardless of author. Both implementations use at most `ceil(requested/100)` fetch/delete rounds and never request more than 100 messages per Discord call.

Legacy schedules each round six seconds apart. It starts the fetch/delete promise and, on the final timer, edits success immediately before that final asynchronous delete completes. The displayed count can therefore omit the last batch, commonly showing `0/<requested>` for one-batch cleanup even when messages are deleted later. Promise failures can occur after the response path.

Go performs each fetch/delete sequentially, waits for Discord, and reports the actual number confirmed deleted. If fewer messages are available, it returns the lower actual count. A cleanup-specific 60-second middleware floor permits the maximum ten sequential rounds even when the global interaction timeout is shorter.

## User-Filtered Deletion

With `使用者`, the requested number is the desired target-user deletion count. Legacy schedules exactly ten rounds six seconds apart and fetches up to `min(requested - deleted, 100)` current recent messages each time. It filters by exact author ID. Because it has no `before` cursor, a page with no matching user can be fetched repeatedly until all ten rounds expire.

Go preserves the ten-round ceiling and per-request limit, but advances a `before` cursor after a nonmatching page. It can therefore inspect older recent pages and delete the requested target-user messages without repeatedly reading the same viewport. Non-target messages remain untouched. One matching message uses Discord's single-delete endpoint; larger matching batches use bulk delete.

## Fourteen-Day Boundary

Go calculates a 14-day cutoff for every fetched message and never submits an older message to either delete endpoint. A batch with recent and old messages deletes only recent eligible IDs. User-filter scanning can continue past recent non-target pages but stops when no recent page remains.

Legacy calls `bulkDelete(messages)` with `filterOld=false`:

- two or more IDs containing an old message can make Discord reject the entire bulk request;
- exactly one fetched ID is special-cased by discord.js to the single-delete endpoint, which can delete a message older than 14 days.

Uniformly retaining old messages is an intentional safety difference consistent with the command's visible `(只能刪除14天內的消息)` contract.

## Failures And Partial Progress

Go stops on the first fetch or delete error and returns the generic ephemeral error. Earlier confirmed batches remain deleted; Discord message deletion is not transactional and cannot be rolled back. The success embed is sent only after all planned rounds complete or safely stop for no eligible messages.

Legacy timer/promise failures are outside the surrounding `try/catch` and can leave an unhandled rejection, stale defer, premature success, or partial deletion. Controlled error UI and accurate completion timing are intentional reliability improvements.

## Data And Migration

The feature reads and deletes Discord messages only. It does not read/write Mongo, create indexes, mutate configuration, write audit rows, or require a database migration. Enabling or disabling it leaves all Mongo collections unchanged.

No data backfill, schema change, or database operation is required. Migration consists only of exclusive command ownership, reviewed guild command sync, Discord permission checks, and disposable-channel smoke.

## Intentional Differences

Intentional differences are limited to:

- unfiltered zero/negative and filtered negative requests receive controlled errors instead of unresolved/asynchronous legacy behavior;
- filtered zero emits one equivalent success instead of up to ten repeated edits;
- Go waits for deletion and reports the actual count instead of editing before the final promise;
- target-user scans advance through older pages instead of repeating a nonmatching viewport;
- old messages are uniformly retained, including one-item batches;
- API failures return controlled UI after any partial progress;
- cleanup receives a 60-second interaction lifetime floor;
- mentions are explicitly suppressed.

Exact definition text, public discoverability, runtime permission thresholds, 1000 maximum, 100-message batches, ten filtered rounds, author matching, ephemeral UI text, red/error and configured success colors, and no-Mongo behavior are preserved.

## Parity Tests

Run focused coverage:

```bash
go test ./internal/adapters/discordgo ./internal/discord/features/moderation ./internal/app ./internal/config ./internal/parity ./cmd/mhcat-command-sync ./cmd/mhcat-staging-preflight
go test -race ./internal/adapters/discordgo ./internal/discord/features/moderation ./internal/app
go vet ./internal/adapters/discordgo ./internal/discord/features/moderation ./internal/app
go run ./tools/parity-audit --legacy-root ../MHCAT --format markdown
```

The DiscordGo adapter tests use an in-process Discord REST server to lock 100-message pagination, actual counts, target scanning, single deletion, and age filtering. No Mongo integration test applies.

## Staging Smoke

1. Use a staging-only text channel containing only disposable messages; stop the Node cleanup owner.
2. Enable paired flags, run preflight and command-sync dry-run, review guild apply, and start one Go owner.
3. Confirm the command is publicly discoverable; verify non-Manage Messages denial is exact red and ephemeral.
4. As a manager, test counts `1001`, `201` without Administrator, `0` without user, `0` with user, and `-1`; verify exact UI and no unintended deletion.
5. Seed more than 100 recent messages, run unfiltered cleanup as an Administrator, and verify batching plus the actual `#53FF53` completion count.
6. Seed interleaved target/non-target messages across multiple pages, run filtered cleanup, and verify only target messages are deleted and scanning advances.
7. Seed recent and older-than-14-day messages, including a page with exactly one old message; verify every old message remains.
8. Force fetch, first-delete, and later-delete failures; record partial progress and verify generic error with no raw details.
9. With usage tracking enabled separately, verify one event per denied/invalid/successful slash attempt and no duplicate handler event.
10. Confirm no Mongo collection or index changes, no duplicate initial response, and no command outside the managed guild scope changed.

## Rollback

1. Disable command-sync inclusion and remove only the managed staging `刪除訊息` command.
2. Disable the runtime gate and stop the Go owner before restoring Node.
3. Do not attempt to restore deleted Discord messages automatically; Discord deletion is irreversible.
4. Review any owner-overlap interval and channel audit log for duplicate/partial deletion.
5. Restore Node only after confirming no Go cleanup route remains.
6. Recheck one permission denial and one small disposable cleanup under the restored owner.

Production ownership remains blocked on live disposable-channel smoke, exclusive ownership, reviewed bot/channel permissions, partial-failure testing, and acceptance of the documented safer differences. No database migration is required.
