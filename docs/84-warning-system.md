# Warning System Parity Contract

Status: parity-audited against all five active legacy warning slash commands, both Mongoose models, global slash dispatch, discord.js payload behavior, Discord member actions, Mongo naming and scalar coercion, runtime wiring, command sync, and staging preflight. All runtime and command-sync gates remain disabled by default. Live staging smoke is still required before production ownership.

## Scope

This contract covers:

- `/警告紀錄` history lookup;
- `/警告設定` threshold/action configuration;
- `/警告` creation, target DM, and threshold kick/ban;
- `/警告清除` one-entry removal;
- `/警告全部清除` all-entry removal;
- permissions, usage, UI, mixed BSON compatibility, duplicate rows, migration, staging, and rollback.

Legacy sources:

- `slashCommands/警告系統/warnings.js`
- `slashCommands/警告系統/erros_set.js`
- `slashCommands/警告系統/warn.js`
- `slashCommands/警告系統/remove-warn.js`
- `slashCommands/警告系統/remove-all-warnings.js`
- `models/warndb.js`
- `models/errors_set.js`
- `events/SlashCommands.js`

Message cleanup and delete-data remain separate contracts.

## Gates And Ownership

Each command family has an independent runtime and staging command-sync pair:

| Commands | Runtime | Command sync |
| --- | --- | --- |
| `/警告紀錄` | `MHCAT_FEATURE_WARNINGS_ENABLED=true` | `MHCAT_COMMAND_SYNC_INCLUDE_WARNINGS=true` |
| `/警告設定` | `MHCAT_FEATURE_WARNING_SETTINGS_ENABLED=true` | `MHCAT_COMMAND_SYNC_INCLUDE_WARNING_SETTINGS=true` |
| `/警告清除`, `/警告全部清除` | `MHCAT_FEATURE_WARNING_REMOVAL_ENABLED=true` | `MHCAT_COMMAND_SYNC_INCLUDE_WARNING_REMOVAL=true` |
| `/警告` | `MHCAT_FEATURE_WARNING_ISSUE_ENABLED=true` | `MHCAT_COMMAND_SYNC_INCLUDE_WARNING_ISSUE=true` |

Command sync additionally requires `MHCAT_STAGING_MODE=true` and guild scope. Every flag defaults to false. Preflight rejects sync without matching runtime and warns when runtime is enabled without sync.

Issue-only startup still opens both `warndbs` and `errors_sets`, because `/警告` reads escalation settings even when `/警告設定` is not routed. Startup also supplies member lookup, hierarchy, DM, kick/ban, channel-message, and clock adapters. Stop all matching Node warning owners before enabling any Go family for the same guild. The families may migrate independently, but warning issue and Node warning settings must not race on shared escalation ownership without explicit review.

## Definitions, Permissions, And Usage

All five definitions preserve exact names, descriptions, option order, required flags, and string choices. They are chat-input commands, guild-managed for staging, and publicly discoverable with no Discord default-member permission.

Legacy `UserPerms` is help metadata. Runtime behavior is exact:

- `/警告紀錄` does not check Manage Messages, despite advertising `訊息管理`;
- `/警告設定`, `/警告`, `/警告清除`, and `/警告全部清除` require Manage Messages (`8192`);
- `/警告` additionally requires the actor's highest role position to be strictly greater than the target's. Equal, lower, and self-target role comparisons are rejected.

Every command defers publicly and edits the original public response. Legacy history passes `ephemeral: true` only to an edit after a public defer, which cannot change visibility. Errors use the animated-no title prefix and discord.js `Red` (`0xED4245`). Success uses discord.js `Green` (`0x57F287`) except history, which uses `Random`.

Legacy cooldown metadata is `10`, but the global dispatcher does not enforce it. Go adds no handler cooldown. Usage belongs only to global slash middleware: when tracking is enabled, each allowed, denied, missing-data, or failed slash attempt records exactly one best-effort event. Handlers and threshold side effects do not record another event.

## History UI And Reads

`/警告紀錄 使用者:<user>` renders:

- title `以下是<username>的警告紀錄`;
- one random color;
- entries joined with one space, each shaped as:

  ````text
  \n<n> ```- 警告者: <tag>
  - 原因: <reason>
  - 時間: <time>```
  ````

- no mention parsing.

Missing rows and empty hydrated content use exact error `這位使用者沒有任何警告!`. Legacy crashes when a stored moderator is absent from member cache. Go intentionally falls back to the stored moderator ID, and can REST-fetch target/member information when cache data is absent.

History reads preserve mixed legacy rows. `guild` and `user` use Mongoose String-like scalar coercion. Array content remains ordered; a non-array scalar is hydrated as one array element. Object fields use JavaScript-like interpolation, including `undefined`, `null`, `[object Object]`, and comma-joined arrays. Malformed individual entries do not make the whole row unreadable.

## Settings UI And Persistence

`/警告設定` preserves required options in this order:

1. string `執行的動作`, description `警告他的原因`, choices `停權` and `踢出`;
2. integer `幾次警告後執行動作`, description `被警告幾次後要執行這個動作!`.

The success embed title is `警告系統`, with exact description:

```text
警告成功設為警告<threshold>次後
執行<action>
```

Discord integers have no legacy positive minimum. Zero and negative thresholds remain accepted. Existing-row writes store typed string `ban_count` and `move` values without rewriting the matched `guild` key; a new upsert stores all three as strings. If duplicate guild rows exist, Go intentionally aligns every matching duplicate with one update and upserts only when none match. This avoids legacy's two non-awaited updates splitting fields, while remaining readable by Mongoose.

Settings reads use the first matching row and accept Mongoose-coercible scalar fields. `ban_count` follows JavaScript `Number`: null and empty become zero, decimal and hexadecimal strings are accepted, and malformed/missing/compound values become `NaN`. `move` is retained as the hydrated string even when it is unknown, empty, or malformed. Command writes remain restricted to the two advertised choices.

## Warning Creation And Escalation

`/警告 使用者:<user> 原因:<reason>` preserves the raw reason, including leading/trailing or all-space text. It appends:

- `time`: `Asia/Taipei` formatted as `YYYY年MM月DD日 HH點mm分`;
- `moderator`: invoking user ID;
- `reason`: unchanged command text.

After a successful Mongo write, the public success title is:

`<a:greentick:980496858445135893> | 成功警告這位使用者!`

The target DM is best effort. It uses title `<:warning:985590881698590730> | 警告系統`, configured legacy error color `#EA0000`, guild name, raw reason, and `<username>(id:<actor ID>)`. DM failure does not replace success.

The first newly created `warndbs` row never evaluates escalation, even when the configured threshold is zero or one. Only appending to an existing row reads the first `errors_sets` row and evaluates JavaScript-equivalent `warning count >= Number(ban_count)`. `NaN` never matches.

When the threshold matches:

- exact stored `move === "停權"` bans;
- every other value, including `踢出`, unknown, empty, Boolean-coerced, or padded text, kicks;
- kick/ban uses no Discord audit-log reason, matching the omitted legacy API argument;
- the best-effort channel success embed always says the action actually taken, hardcoded as `停權` or `踢出`;
- a channel-send failure does not replace command success.

The initial success edit and DM occur before escalation. A missing settings row, unreadable threshold, or settings read failure skips escalation. Go waits for the kick/ban API result; failure replaces the original response with exact `我沒有權限ban掉他` or `我沒有權限踢出他` after the earlier success edit and DM. This preserves the visible legacy permission-failure ordering while converting asynchronous API failures into controlled UI.

## Removal Semantics And UI

Both removal commands require Manage Messages and use the same public green success title:

`<a:greentick:980496858445135893> | 這位使用者的警告成功移除!`

Their best-effort DMs use title `<:warning:985590881698590730> | 警告系統`, color `#00DB00`, guild name, actor username/ID, and either `一個__警告__` or `所有__警告__`. DM failure does not replace success.

`/警告清除` follows JavaScript `content.splice(option - 1, 1)` exactly:

- `1` removes the first entry;
- `0` removes the last entry;
- negative values count backward after subtracting one;
- a very negative index clamps to the first entry;
- an index above the array length succeeds without changing content;
- an existing empty content array also succeeds.

It mutates the exact `_id` selected by the initial first-row read and preserves mixed array values. Missing data uses `這位使用者沒有任何警告!`.

`/警告全部清除` intentionally deletes every duplicate `{guild,user}` row rather than legacy's one arbitrary `findOneAndDelete` match. Missing data preserves the distinct text without an exclamation mark: `這位使用者沒有任何警告`.

Go commits removal before returning success and DM. Legacy sends success/DM before non-awaited save or delete. Waiting is an intentional durability improvement: Go never reports a removal that Mongo rejected.

## Mongo Compatibility And Migration

Collections remain exact:

- `warndbs`: `guild`, `user`, `content`;
- `errors_sets`: `guild`, `ban_count`, `move`.

No startup read, repair, deduplication, backfill, index creation, or schema migration runs. Existing rows remain untouched until a command targets them. Appending to an existing warning uses atomic `$push` against the exact `_id`; scalar `content` is normalized to an array only for that targeted row. New warning rows use typed string fields and an array of entry objects.

Duplicate rows are supported legacy state. Reads and single-entry mutations use one arbitrary first match; remove-all deliberately cleans all matches; settings writes align all matching guild rows. No unique index is safe merely to enable Go. Candidate indexes require audits for duplicate, missing/null/blank/scalar-drift keys and all Node/dashboard/external writers. No database operation is required for migration.

## Controlled Failure Behavior

Mongo operational failures return the generic public red error:

`<a:Discord_AnimatedNo:1015989839809757295> | 很抱歉，出現了未知的錯誤，請重試!`

This intentionally replaces legacy callback throws, ignored errors, and deferred interactions that could remain unresolved. Backend details are never shown. Missing data keeps each command's legacy missing text. DMs and threshold channel announcements remain best effort.

## Intentional Differences

Intentional differences are limited to:

- history moderator/target lookup safely falls back or REST-fetches instead of crashing on a cache miss;
- warning/settings Mongo writes are awaited and use atomic/exact-row updates;
- settings writes align duplicate guild rows instead of issuing two non-awaited first-match field updates;
- remove-all deletes all duplicate target rows instead of one arbitrary row;
- Mongo and Discord action failures return controlled UI;
- mentions are explicitly suppressed.

Public definitions, runtime permission policy, exact text/colors/effective visibility, raw reason, omitted threshold audit reason, first-row threshold skip, JavaScript numeric/action/splice behavior, timestamps, DMs, and collection names are preserved.

## Migration And Staging

1. Use an isolated staging guild, disposable warning rows, and disposable target members.
2. Stop matching Node warning command owners before enabling Go.
3. Back up `warndbs` and `errors_sets`.
4. Audit duplicates, mixed `content`, scalar guild/user/settings fields, unknown actions, malformed thresholds, and current indexes.
5. Preserve rows as-is. Do not normalize, deduplicate, backfill, or create indexes before smoke.
6. Seed duplicate warning/settings rows, mixed content, decimal/null/malformed thresholds, and an unknown action.
7. Keep thresholds safely above normal test counts except in a controlled disposable kick/ban case.
8. Pair only the command families under test, run preflight and dry-run, review guild apply, and confirm one active owner.

## Parity Tests

Run focused coverage:

```bash
go test ./internal/core/domain ./internal/core/services/moderation ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/adapters/discordgo ./internal/discord/features/moderation ./internal/app ./internal/config ./internal/parity ./cmd/mhcat-command-sync ./cmd/mhcat-staging-preflight
go test -race ./internal/core/services/moderation ./internal/adapters/mongo/repositories ./internal/adapters/discordgo ./internal/discord/features/moderation ./internal/app
go vet ./internal/core/services/moderation ./internal/adapters/mongo/repositories ./internal/adapters/discordgo ./internal/discord/features/moderation ./internal/app
go run ./tools/parity-audit
```

The real-Mongo tests require a disposable database:

```bash
MHCAT_RUN_MONGO_INTEGRATION_TESTS=true \
MHCAT_MONGODB_URI='<disposable-uri>' \
go test ./internal/adapters/mongo/repositories -run '^TestWarning' -count=1
```

## Staging Smoke

1. Complete backup, audits, Node shutdown, paired flags, preflight, dry-run, reviewed guild apply, and runtime startup.
2. Confirm all five commands are publicly discoverable and every definition exactly matches legacy metadata.
3. As a non-manager, confirm history remains usable while the other four commands return exact public red Manage Messages denial.
4. Seed mixed history entries and confirm random-color rendering, exact formatting, moderator fallback, missing/empty errors, and one usage event per slash attempt when tracking is enabled.
5. Configure positive, zero, negative, and decimal-compatible settings; verify typed duplicate alignment and exact green UI without indexes.
6. Issue first and existing warnings with padded/all-space reasons; verify raw Mongo/DM text, Taipei timestamp, strict role hierarchy, first-row threshold skip, and no audit reason.
7. Exercise exact ban, normal kick, unknown-action fallback kick, decimal/null/malformed thresholds, permission failure ordering, DM failure, and action-message failure using disposable members.
8. Exercise remove indexes `1`, `0`, negative, very negative, and too large against mixed content; verify exact row identity and DMs.
9. Exercise remove-all against duplicates; verify every target row is gone while another user/guild remains.
10. Confirm no startup repair, migration, index, unrelated collection write, duplicate response, or raw internal error.
11. Disable gates, remove only managed staging commands, preserve rows/indexes, and complete rollback checks.

## Rollback

1. Disable command-sync flags and remove only the managed staging warning commands.
2. Disable all warning runtime gates and stop the Go owner before restoring Node.
3. Preserve `warndbs`, `errors_sets`, mixed values, duplicate rows, and indexes as found. Do not repair during emergency rollback.
4. Restore disposable rows from backup only if smoke mutation must be undone; review exact `_id` rows before restoring duplicates.
5. Restore Node ownership only after confirming no Go warning routes remain.
6. Recheck history, one settings write, one warning below threshold, one single removal, and one remove-all against disposable fixtures.
7. Review any overlap interval for duplicate warnings, split settings writes, DMs, kicks, or bans.

Production ownership remains blocked on live staging smoke, exclusive command ownership, disposable kick/ban verification, and reviewed duplicate/mixed-shape audits. No unique index or data rewrite is required merely to enable Go.
