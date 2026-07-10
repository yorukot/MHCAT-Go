# XP Reward Role Config Parity Audit

Status: parity-audited behind explicit runtime and command-sync gates.

## Legacy References

- `MHCAT/slashCommands/經驗系統/text_leave_role.js`
- `MHCAT/slashCommands/經驗系統/voice_leavel_role.js`
- `MHCAT/events/rank.js`
- `MHCAT/models/chat_role.js`
- `MHCAT/models/voice_role.js`

## Scope And Gates

This slice implements:

- `/聊天經驗身分組設定`;
- `/語音經驗身分組設定`;
- subcommands `增加`, `刪除`, and `設定查詢`;
- legacy pagination components `<page>text_leave_role` and `<page>voice_leave_role`;
- rollback-compatible `chat_roles` and `voice_roles` reads and writes.

Runtime requires `MHCAT_FEATURE_XP_ROLE_CONFIG_ENABLED=true`. Staging command sync also requires `MHCAT_COMMAND_SYNC_INCLUDE_XP_ROLE_CONFIG=true`. The staging preflight and scripts reject unpaired flags.

The config commands do not activate XP accrual or reward delivery by themselves. Text reward roles run through the separately gated text-XP accrual path. Voice reward roles run through the separately gated voice-XP session path.

## Preserved Command Contract

- Both command definitions and handlers require Manage Messages (`8192`).
- Slash commands defer publicly and edit the original response.
- Add accepts required integer `等級`, required role `身分組`, and optional boolean `是否自動刪除`; omission stores `false`.
- The selected role must be below the bot's highest role. Equal or higher positions return `我沒有權限給大家這個身分組(請把我的身分組調高)!`.
- The legacy count check runs before the hierarchy check. A guild with 119 rows may add its 120th row; a guild already holding 120 rows receives `你的設定已經過多，請先刪除一些!`.
- Add success uses color `#53FF53`, the channel emoji, and `成功\`增加\`/\`修改\`該設定`.
- Delete matches guild, level, and role. A missing match returns `你沒有設定過這個選項!`; success uses color `#53FF53`, the trash emoji, and `成功刪除該設定`.
- Errors use the legacy animated-no title prefix and Discord red `#ED4245`.

## Preserved Query Contract

- Query results retain Mongo's unsorted result order and nominally paginate 12 rows at a time.
- Each field preserves the legacy level, role, trash emojis, role mention shape, inline layout, and boolean text.
- Text buttons use `<page>text_leave_role`; voice buttons use `<page>voice_leave_role`.
- Previous and next buttons preserve the legacy labels, emojis, success style, and disabled states.
- The footer remains `總共: <count> 筆資料\n第 <page> / <pages> 頁(按按鈕會自動更新喔!`, including the unmatched opening parenthesis.
- An empty query preserves page `1 / 0`, no fields, and disabled navigation.
- Every slash query and component page renders a random 24-bit Discord color.

The legacy sixth-field indexing bug is preserved exactly. Field six is gated by row `page*12+12`, so a page with six through twelve total rows omits field six. When that gate exists, its level comes from row `page*5+12`, while its role and delete flag come from row `page*12+5`.

Queries also preserve stale-role cleanup behavior. Every row is checked against the Discord guild role cache. A missing role is deleted best-effort, but remains visible in the response being built from the current query result; it disappears on the next query.

## Mongo Compatibility

- Collections remain `chat_roles` and `voice_roles`.
- Writes preserve fields `guild`, misspelled string `leavel`, `role`, and `delete_when_not`.
- Reads accept legacy `leavel` values stored as strings or BSON numeric types.
- Filters use the legacy string level representation.
- No indexes are created by application startup.

## Intentional Go Differences

- Legacy add deletes one matching document and starts an unawaited insert. Go deletes every duplicate matching `{guild, leavel, role}` row, inserts one replacement, and waits for each operation.
- Legacy delete removes one `findOne` result. Go removes every duplicate matching row.
- Persistence and list failures receive the legacy-style unknown-error payload instead of leaving a deferred interaction unresolved. Stale-role cleanup remains best-effort so a cache or cleanup failure does not replace the query response.
- Legacy component handlers wait one second before updating. Go updates immediately.
- Go explicitly suppresses allowed mentions on response payloads.
- Go's component parser accepts only the documented non-negative page grammar and bounds stale out-of-range pages. The normal button IDs and page behavior remain legacy-compatible.
- Runtime usage accounting belongs to the global slash middleware, preventing route-level double counting. Component clicks are not counted as slash commands.

## Verification Coverage

Automated tests lock:

- both legacy command definitions and options;
- exact text and voice add/delete/query/error payloads;
- public defer and runtime permission behavior;
- the 120-row limit boundary and limit-before-hierarchy precedence;
- missing-delete behavior for both collections;
- hierarchy rejection and missing-role mapping;
- pagination IDs, labels, emojis, footer, random-color range, and sixth-field indexing;
- cached stale-role display-then-delete behavior;
- mixed string/numeric `leavel` decoding and rollback-compatible writes;
- route registration and runtime/command-sync gate pairing.

## Staging Checklist

1. Use an isolated staging guild and disposable `chat_roles` / `voice_roles` data.
2. Enable both reward-role config flags and run `go run ./cmd/mhcat-staging-preflight`.
3. Run command-sync dry-run and review both command definitions before apply.
4. Add a role below the bot's highest role and confirm the green response and stored string `leavel`.
5. Attempt an equal or higher role and confirm the red hierarchy error.
6. Query zero, one, twelve, and thirteen rows; verify footer counts, navigation IDs, and the preserved sixth-field behavior.
7. Delete a configured role in Discord, query once to observe and clean the stale row, then query again to confirm it is absent.
8. Delete one config and confirm a second delete returns the missing-setting error.
9. Keep the Node.js owner stopped for the same guilds while Go owns writes to these shared collections.
