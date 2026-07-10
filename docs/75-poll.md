# Poll Parity Contract

Status: poll creation, voting, owner controls, result rendering, and exports are parity-audited behind one disabled-by-default runtime ownership gate. Live staging smoke remains required before production ownership.

## Legacy References

- Create command: `MHCAT/slashCommands/管理系統/poll.js`
- Vote, result, and owner interactions: `MHCAT/events/poll.js`
- Model: `MHCAT/models/poll.js`
- Collection: `polls`

## Gates And Ownership

Runtime routes require:

```bash
MHCAT_FEATURE_POLLS_ENABLED=true
```

Publishing the managed command for an isolated staging guild additionally requires:

```bash
MHCAT_STAGING_MODE=true
MHCAT_COMMAND_SYNC_INCLUDE_POLLS=true
```

`MHCAT_FEATURE_POLLS_ENABLED` owns `/投票創建`, versioned poll components, and all legacy `poll_<choice>`, `see_result`, `poll_menu`, and `menu_choose` routes as one boundary. Do not leave `slashCommands/管理系統/poll.js` or `events/poll.js` active for the same bot/guilds while Go owns polls. Shared ownership can double-apply votes and toggles even though each Go repository operation is atomic.

Poll interactions require no message-content intent. Exact participation totals and bulk export member names use guild-member REST listing/fetches like legacy. Enable the privileged Guild Members intent for the application and set `MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true` for staging parity. If member counting fails, creation falls back to zero and later rerenders fall back to the current unique-voter count instead of failing the interaction.

Command sync remains guild-scoped and staging-only. Preflight rejects command inclusion without the runtime feature and warns when runtime is enabled without command inclusion.

## Command And Usage Contract

The managed definition preserves:

- name `/投票創建`;
- description `創建一個萬能的投票`;
- required string option `問題`, description `輸入你要問的問題!ex:我要買甚麼?`;
- required string option `選項`, description `輸入回答的選項，請用^將各個選項分開 ex:電腦^手機^兩個都要^!`;
- default Manage Messages permission metadata (`8192`).

The handler ephemerally defers before checking Manage Messages. Denial edits that deferred reply with title ``<a:Discord_AnimatedNo:1015989839809757295> | 你需要有`訊息管理`才能使用此指令``.

The legacy command declares `cooldown: 0`; Go adds no feature-local cooldown.

Slash usage belongs only to the assembled runtime's global middleware. With `MHCAT_FEATURE_USAGE_TRACKING_ENABLED=true`, each slash attempt writes one `投票創建` event, including denied and failed attempts. Vote, result, owner-menu, and max-choice components never write command usage.

## Creation And Validation Contract

Validation order is exact:

1. Question length must not exceed 2,500 JavaScript UTF-16 code units.
2. Splitting `選項` on `^` must produce at least 2 and at most 19 values.
3. Exact duplicate values are rejected before individual value validation.
4. Each choice must be at most 80 UTF-16 code units.
5. Empty choices are rejected.

Question and choice text is not trimmed. Leading/trailing whitespace and nonempty all-whitespace strings remain stored and rendered. Duplicate comparison is exact and case-sensitive. Domain validation and Mongo vote matching preserve the same raw choice value.

The fixed validation titles remain:

- `問題字數不可超過2500`
- `最少需要2個選項!`
- `最多只能有19個選項!`
- `選項名稱不可以重複!`
- `你輸入的選項字數不能超過80`
- `^跟^中間請填入選項，不可為空`

After validation, Go counts non-bot members, sends the public poll, then inserts its `polls` row. A failed insert deletes only the exact message just sent under a cancellation-independent five-second cleanup context. If deletion also fails, both errors are returned. A successful insert is retained if the final ephemeral success edit fails.

## Initial Poll UI

The initial public embed preserves:

- title `<:poll:1023968837965709312> | 投票\n<question>`;
- total line with `0`, current non-bot member count, and hardcoded `0.00` participation;
- max choices `1`;
- initial text `` `不能`改投其他選項 ``;
- hidden results and named voting;
- one random Discord color in the inclusive `0x000000` through `0xFFFFFF` range.

Choice buttons are secondary. The result button is success style, labeled `查看投票結果`, with `<:analysis:1023965999357243432>`. Rows contain at most five buttons and the owner select occupies the final row.

The initial owner menu preserves these labels in order:

1. `公開投票結果`
2. `啟用多選投票`
3. `允許變更選項`
4. `改為匿名投票`
5. `結束投票`
6. `匯出為excel檔`

The initial end description is `讓該投票變為無法再變更選項或投票(可再次開啟)`. Later rerenders intentionally preserve the legacy dynamic labels `可以變更選項`, `終止投票`, and the trailing-`讓` description typo.

Creation success is ephemeral title `<a:green_tick:994529015652163614> | 成功創建投票!`.

## Component ID Contract

New Go messages use bounded IDs:

- `mhcat:v1:poll:vote:i=<choice-index>`
- `mhcat:v1:poll:result:`
- `mhcat:v1:poll:owner_menu:`
- `mhcat:v1:poll:max_choices:m=<source-message-id>`

Go also accepts old live IDs:

- `poll_<raw-choice>` up to the legacy choice bound, including `:`;
- exact `see_result`;
- exact `poll_menu`;
- exact `menu_choose`, which receives a controlled stale-menu error because the legacy collector closure and source message ID do not survive ownership transfer or restart.

A legacy choice named `menu` also has ID `poll_menu`. Go distinguishes that button from the owner select by interaction shape: no selected values routes to choice `menu`; a string-select value routes to owner controls.

Versioned IDs deliberately replace raw user text, broad substring routing, and restart-local collectors. A successful Go rerender upgrades a legacy poll message to versioned components.

## Vote Contract

Vote interactions ephemerally defer, load the poll by `{guild,messageid}`, resolve the exact choice, and apply one conditional Mongo update.

- Ended poll: `該投票已被結束!`
- Duplicate choice while changes are disabled: `很抱歉，該投票不支援更改選項!`
- Choice limit: `你已經達到該投票最大上限`, reason `如需更改選項，請將原來所選的選項點掉!`
- Add success: ``<a:green_tick:994529015652163614> | 你成功投給`<choice>`!`` with the legacy change hint.
- Remove success: ``<a:green_tick:994529015652163614> | 成功取消投給`<choice>`!``.

Success uses Discord named `Green` (`0x57F287`); errors use named `Red` (`0xED4245`). Vote timestamps remain decimal Unix milliseconds as strings.

Add/remove is atomic across concurrent Go voters. It rejects choices absent from `choose_data`, enforces the current Mongoose-compatible `many_choose` value in the update predicate, and prevents duplicate `{id,choise}` entries. After a change, Go best-effort rerenders the source poll with current totals, dynamic owner controls, and a new random color.

Participation and result percentages match JavaScript `Number.toFixed(2)`, including binary halfway behavior such as `1/32 * 100 -> 3.13` and `23/160 * 100 -> 14.37`.

## Result And Export Contract

Results are visible when `can_see_result` is true or the actor is `create_member_id`. No votes returns `還沒有人參與投票!`. A private result returns `這個投票不是公開的!` with reason `如需公開該投票，請使用下方選擇器!`.

The result response preserves:

- random-color title `<:poll:1023968837965709312> | <question>`;
- one field per choice: ``<choice>(共<n>人 `<percentage>`%)``;
- concatenated user mentions for ordinary results;
- global `該投票為匿名，無法查看誰有進行投票` suppression for anonymous polls;
- global `由於人數過多，無法顯示所有人` suppression above 50 vote rows;
- `file.jpg` and `discord.txt` attachments.

`discord.txt` keeps one legacy row per vote with user ID, `username#discriminator` (including modern `username#0`), exact choice, and timestamp. Missing members use `使用者已退出伺服器!`; anonymous rows replace both ID and name with `該投票為匿名`.

Owner export returns content `<:sheets:1023972957330100324> | **以下是該投票的excel表格!**` and `poll_info.xlsx`. Anonymous export is rejected with `該投票為匿名，無法查看投票資訊!`.

After successful result, max-choice prompt, Excel export, and all owner toggle paths, Go best-effort refreshes the source poll like legacy. The anonymous-to-named lock error also refreshes it.

## Owner Controls

Only `create_member_id` may use the owner menu. Public-result, change-choice, and end controls are atomic Mongoose-aware flips. Anonymous is one-way and a second attempt returns `匿名的投票無法改為實名!`.

Owner-toggle success titles intentionally concatenate the done emoji directly with text, without ` | `, matching legacy `done_embed`. They use Discord named `Green`.

Multi-choice requires at least three choices and offers values `1` through `len(choices)-1`. Go edits the acknowledged ephemeral response with a versioned select instead of reproducing the legacy public follow-up and in-memory collector. Selection updates the source poll and replaces the prompt with random-color title ``<a:green_tick:994529015652163614> | 成功將最多選擇數量設為<n>`` while removing components.

## Mongo Compatibility

`polls` retains exact legacy fields:

- `guild`
- `messageid`
- `question`
- `create_member_id`
- `many_choose`
- `can_change_choose`
- `can_see_result`
- `end`
- `anonymous`
- `choose_data`
- `join_member`, containing `{id,choise,time}`

Writes remain typed BSON strings, integers, booleans, and arrays readable by Mongoose. Reads use a separate permissive DTO:

- schema `String` fields apply Mongoose-compatible scalar conversion;
- `many_choose` accepts exact integer Mongoose-number forms and normalizes missing, invalid, or nonpositive values to `1`;
- booleans accept Mongoose true/false scalar forms;
- a scalar loose-array field is treated as one element like a Mongoose `Array` path;
- non-string choices and malformed vote entries are skipped because the legacy button/export code cannot safely use them.

Vote predicates recognize Mongoose true forms (`true`, `"true"`, numeric/string `1`, and `"yes"`) and convert ordinary numeric scalar forms for `many_choose`. Owner flips use one atomic update pipeline and return that update's document. Unknown fields are untouched; writes never replace the full document.

Mongo `findOne`/`UpdateOne` duplicate-row semantics remain. The application creates no startup index. Candidate unique index `polls_guild_message` on `{guild:1,messageid:1}` must not be applied until duplicate keys, malformed keys, scalar drift, loose-array shapes, and large vote arrays have been audited. There is no TTL index; the preserved `超過30天` error text does not imply this module deletes old polls.

## Intentional Safety And Reliability Differences

- New components use versioned bounded IDs; legacy broad substring routing is not reproduced.
- The max-choice prompt is ephemeral and restart-stable instead of a public follow-up tied to an in-memory collector.
- Missing polls, malformed components, invalid choices, and repository/Discord failures return controlled errors instead of unresolved defers, null dereferences, or unhandled promises.
- Per-choice validation stops cleanly even though the legacy `forEach` callback did not stop its outer function.
- Mongo writes are awaited. Votes and owner flips use conditional atomic updates instead of full embedded-array replacement and read-modify-write toggles.
- Failed creation cleans up its exact Discord message under a bounded cancellation-independent context.
- A vote rerender uses normalized embed indentation instead of preserving one legacy branch's accidental four-space indentation.
- Chart output is a deterministic 500x500 standard-library JPEG using the legacy background and slice colors, not a pixel-identical Chart.js rendering.
- Excel output is a deterministic minimal OpenXML workbook, not byte-identical `write-excel-file` output.
- Numeric timestamps use deterministic `YYYY/MM/DD HH:mm:ss 台北標準時間` text instead of host-ICU punctuation and spacing.
- Member tags are fetched once/bulk through a port rather than repeatedly inline; missing lookup batches degrade to the legacy exited-member text.

Raw visible text, UTF-16 limits, validation precedence, initial/dynamic label differences, option order, button rows, result authorization, colors, random color range, percentage rounding, vote semantics, export row text, and legacy typos are preserved.

## Runtime Ownership And Migration

Before enabling Go:

1. Stop every Node process that loads the poll command or `events/poll.js` for the target bot.
2. Audit `polls` for duplicate `{guild,messageid}` keys, malformed/string-drift keys, boolean/number scalar drift, invalid `choose_data`, malformed `join_member`, and large documents.
3. Preserve `polls`; do not backfill or create `polls_guild_message` during the ownership switch.
4. Enable the runtime flag and Guild Members intent in an isolated staging guild.
5. Run staging preflight, review command-sync dry-run, and apply only the managed staging command.

Legacy live poll messages remain usable by Go, including colon choices and the `menu` collision. Once Go rerenders one, its components become versioned.

## Parity Tests

Focused tests lock exact command metadata, initial message text/components, fixed messages, UTF-16 and validation order, raw whitespace, colors, percentage rounding, legacy IDs, result suppression, artifact names/content, Mongo scalar decoding/predicates, atomic toggles, rollback cleanup, centralized usage, runtime wiring, and feature/command gates. Run:

```bash
go test ./internal/core/domain ./internal/core/ports ./internal/discord/features/poll ./internal/discord/customid ./internal/discord/interactions ./internal/adapters/discordgo ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/testutil/fakemongo ./internal/app ./internal/parity ./internal/config ./cmd/mhcat-command-sync ./cmd/mhcat-staging-preflight
go test -race ./internal/core/domain ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/discord/features/poll ./internal/discord/customid ./internal/testutil/fakemongo
go test -race ./internal/app -run Poll
go vet ./...
```

## Staging Smoke

1. Use an isolated guild/database, a manager, ordinary members, a member with a modern discriminator, and choices covering whitespace, emoji, `menu`, and `A:B`.
2. Stop Node poll command/event ownership and confirm only one Go runtime receives interactions.
3. Audit duplicate/malformed `polls` rows and confirm no poll index is being applied.
4. Enable the runtime, command-sync, staging, and Guild Members intent settings; run preflight and command-sync dry-run before apply.
5. Verify command metadata and runtime denial through an existing/direct interaction without Manage Messages.
6. Verify duplicate, empty, overlong, UTF-16 emoji, whitespace, and 2/19-choice boundaries in exact validation order.
7. Create a poll and verify the exact initial title, `不能`, initial owner labels, random color, row layout, success edit, and typed Mongo row.
8. Force Mongo insertion failure and request cancellation; verify the sent message is deleted and no usable row remains.
9. Vote through versioned buttons and old `poll_<choice>` messages, including `poll_menu` as choice `menu` and a colon choice. Verify exact add/remove/limit/end errors and timestamps.
10. Concurrently vote different choices at the max limit and verify the stored array never exceeds the limit or duplicates a user/choice pair.
11. Toggle public, change-choice, anonymous, and end states; verify dynamic labels, direct-concatenated success titles, one-way anonymous behavior, and source refreshes.
12. Open max-choice selection, restart Go before selection, then complete it and verify the source message ID survives and the prompt components are removed.
13. View private/public/owner results; verify JavaScript percentage edge cases, anonymous and over-50 suppression, source refresh, JPEG, and `discord.txt` including modern `#0` tags.
14. Export Excel and verify headers, rows, escaped text, deterministic Taipei time, anonymous denial, and source refresh.
15. Confirm each slash attempt increments usage once when enabled while every component writes no usage event.

## Rollback

1. Disable `MHCAT_COMMAND_SYNC_INCLUDE_POLLS` and remove the managed staging command through command sync.
2. Keep Node poll handlers stopped while any Go process can route poll interactions; never overlap owners.
3. Preserve `polls`; typed Go rows and embedded votes remain Mongoose-readable. Do not mutate/drop indexes during emergency rollback.
4. Go-created and Go-rerendered messages use `mhcat:v1:poll:*` IDs that unmodified legacy Node does not understand. Before restoring Node ownership, either keep Go poll runtime serving those messages, deploy reviewed Node support for versioned IDs, or explicitly retire/recreate every affected live poll. The current schema does not store channel IDs, so there is no automatic bulk message rewrite path.
5. Legacy messages that have never been rerendered by Go remain Node-compatible. The transient versioned max-choice prompt is Go-only and may be discarded/reopened.
6. Restore Node only after confirming no Go process can route `/投票創建` or any poll component for the same bot.
