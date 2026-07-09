# Warning History Read-only Slice

Status: implemented behind explicit runtime and command-sync gates.

## Legacy Reference

- File: `MHCAT/slashCommands/警告系統/warnings.js`
- Command: `警告紀錄`
- Option: required user option `使用者`
- Cooldown: `10`
- Permission metadata: `UserPerms: '訊息管理'`
- Mongo read: `warndb.findOne({ guild: interaction.guild.id, user: user.id })`
- Legacy no-data response: red embed title `<a:Discord_AnimatedNo:1015989839809757295> | 這位使用者沒有任何警告!`
- Legacy success title: `以下是${user.username}的警告紀錄`
- Legacy row format: `- 警告者`, `- 原因`, `- 時間`

## Go Implementation

- Runtime flag: `MHCAT_FEATURE_WARNINGS_ENABLED=false` by default.
- Command sync flag: `MHCAT_COMMAND_SYNC_INCLUDE_WARNINGS=false` by default.
- Command sync requires staging mode and guild scope.
- Repository: `internal/adapters/mongo/repositories.WarningHistoryRepository`
- Service: `internal/core/services/moderation.WarningHistoryService`
- Handler: `internal/discord/features/moderation.WarningHistoryHandler`
- Collection: `warndbs`

The repository is read-only. It performs no warning writes, no warning deletes, no escalation updates, no Mongo repairs, and no index creation.

## Intentional Fixes

- The Go handler enforces Manage Messages before reading warning history, matching the legacy `UserPerms` metadata and reducing moderation-data exposure.
- The Go handler falls back to the stored moderator ID when a moderator member tag cannot be fetched, avoiding the legacy cached-member nil crash.
- Internal errors are converted to a safe legacy-style red embed instead of exposing raw driver details.

## Separate Warning Settings Slice

`/警告設定` is implemented separately behind `MHCAT_FEATURE_WARNING_SETTINGS_ENABLED=false` and `MHCAT_COMMAND_SYNC_INCLUDE_WARNING_SETTINGS=false` by default. It writes only legacy `errors_sets` threshold/action config.

`/警告清除` and `/警告全部清除` are implemented separately behind `MHCAT_FEATURE_WARNING_REMOVAL_ENABLED=false` and `MHCAT_COMMAND_SYNC_INCLUDE_WARNING_REMOVAL=false` by default. They mutate only legacy `warndbs` rows and send best-effort legacy-style DMs.

## Not Implemented

- Warning creation/escalation.
- Bulk message clear/delete.
- Kick/ban moderation actions.
- Usage count writes.
- Unique indexes on `warndbs`.

## Tests

- Command definition shape.
- Service not-found behavior.
- BSON document conversion.
- Read-only repository not-found mapping.
- Handler permission denial.
- Legacy success embed shape.
- Moderator fallback.
- App runtime wiring gate.
- Command-sync and staging-preflight flag pairing.
