# Work Command Dashboard, Interface, and Admin Slice

Status: implemented behind explicit runtime and command-sync flags.

## Legacy Behavior

Legacy `/打工系統 新增打工事項` currently returns a dashboard redirect instead of writing `work_somethings`.

Visible UI:

- Embed title: `<a:announcement:1005035747197337650> | 該指令已經移往控制面板，請前往控制面板進行設定`
- Embed color: `#df1f2f`
- Link button URL: `https://mhcat.yorukot.me//guilds/<guildID>/work`
- Link button label: `點我前往儀錶板設定!`
- Link button emoji: `<a:arrow:986268851786375218>`

The Go handler preserves that UI exactly for the implemented subcommand.

Legacy `/打工系統 打工介面` reads `work_sets`, `work_somethings`, and `work_users`, optionally prompts a captcha modal, renders a work-list embed, lets the user inspect a job detail view, and starts a selected job by updating `work_users`.

Legacy admin subcommands write the same work collections:

- `/打工系統 打工系統設定` updates the guild's `work_sets` config.
- `/打工系統 打工事項刪除` removes a named `work_somethings` entry.
- `/打工系統 增加個人精力` and `/打工系統 增加全體精力` clamp `work_users.energi` to the configured max.

The Go handler preserves the visible success/error embeds and permission behavior, while intentionally fixing legacy bugs where the delete permission path could call an undefined helper and missing-user energy grants could create a row for the admin instead of the target user.

## Implemented

- `internal/discord/features/work`
- `internal/core/services/work`
- `internal/core/domain/work.go`
- `internal/core/ports/work.go`
- `internal/adapters/mongo/repositories/work_interface.go`
- Legacy-shaped `打工系統` slash command definition.
- Runtime route behind `MHCAT_FEATURE_WORK_ENABLED=true`.
- Staging command-sync include gate behind `MHCAT_COMMAND_SYNC_INCLUDE_WORK=true`.
- Staging preflight pairing check for runtime/sync flags.
- Exact legacy dashboard redirect UI for `新增打工事項`.
- Legacy-style `打工介面` list embed, role filtering, no-config/no-items/no-role errors, captcha modal, wrong-captcha content, job-detail embed, active confirm button, busy override prompt, cancel response, energy error, and success embed.
- Legacy-style `打工系統設定`, `打工事項刪除`, `增加個人精力`, and `增加全體精力` admin success/error embeds.
- Versioned `mhcat:v1:work:<action>:job=<hash>,user=<userID>` component IDs for new job/detail/start/override/cancel buttons.
- Explicit start repository wiring. A read-only work repository still renders detail with the confirm button disabled; the app wires the Mongo repository as a start repository only when `MHCAT_FEATURE_WORK_ENABLED=true`.
- Explicit admin repository wiring. Admin setup/delete/energy writes are available only when a `WorkAdminRepository` is supplied; the app supplies the Mongo repository only when `MHCAT_FEATURE_WORK_ENABLED=true`.
- Atomic start update for `work_users`: missing rows are upserted with `energi=max_energy`, then start uses a conditional `FindOneAndUpdate` with `energi >= cost`, `$inc` on `energi`, and `$set` of `state`, `end_time`, and `get_coin`.
- Atomic admin energy grants: individual grants upsert target-user rows with `energi=max_energy`; existing-user and all-user grants clamp with a Mongo aggregation update pipeline.
- Work config saves update all duplicate legacy config rows for a guild instead of deleting and recreating one row, which is rollback-compatible and avoids a temporary missing-config window.
- Tests for command definition, dashboard redirect UI, read-only interface UI, admin setup/delete/energy UI, document conversion, repository boundaries, route registration, config, command sync, staging preflight, and app runtime gating.

## Not Implemented Yet

- `新增打工事項` direct Mongo creation remains dashboard-only, matching the current legacy visible behavior.
- Recurring scheduler wiring.
- Dashboard compatibility writes beyond preserving the link.

Temporary intentional differences:

- Missing `work_users` rows are treated as read-only defaults while rendering the list. The start path explicitly creates a legacy-compatible row only when the user presses the confirm button.
- New job buttons use versioned bounded custom IDs instead of raw job-name custom IDs. The visible button label remains the legacy job name, but routing avoids legacy collision/length bugs.
- New component IDs also carry the original requester ID, preserving the legacy "only the requester can use this menu" behavior without raw custom ID routing.
- Legacy `setColor("Random")` is represented by a stable non-error color so tests are deterministic.
- Energy grants reject zero/negative amounts as invalid input. Legacy did not consistently guard this path, but allowing negative grants through an admin command is unsafe and can corrupt energy state.
- `增加個人精力` creates a missing row for the requested target user. The legacy code accidentally used the admin actor ID in that branch.

Do not sync the work command to production until the recurring payout/reset ownership and dashboard compatibility review are complete or a staging-only partial-command rollout is explicitly accepted.

## Safety

- Disabled by default.
- Mongo writes are limited to explicit work runtime paths: `work_users` start, `work_sets` setup, `work_somethings` delete, and `work_users` energy grants. There are still no writes to `coins`, payout state, scheduler state, or indexes from this command handler.
- The start path is not hidden behind a concrete type assertion; runtime wiring passes a `WorkStartRepository` explicitly.
- The admin path is also explicit; runtime wiring passes a `WorkAdminRepository` and admin subcommands require Manage Messages.
- No Discord command sync runs from bot startup.
- Command-sync inclusion requires staging mode and guild scope.
- Staging preflight requires `MHCAT_COMMAND_SYNC_INCLUDE_WORK=true` and `MHCAT_FEATURE_WORK_ENABLED=true` to be paired.

## Verification

Commands run:

- `go test ./internal/discord/features/work ./internal/config ./cmd/mhcat-command-sync ./cmd/mhcat-staging-preflight ./internal/app`
- `go test ./internal/core/services/work ./internal/adapters/mongo/repositories ./internal/testutil/fakemongo ./internal/discord/features/work ./internal/app`
- `go test ./...`
- `go vet ./...`
- `go build ./cmd/mhcat-bot ./cmd/mhcat-command-sync ./cmd/mhcat-mongo-audit ./cmd/mhcat-mongo-index ./cmd/mhcat-staging-preflight ./cmd/mhcat-work-payout ./cmd/mhcat-economy-reset ./cmd/mhcat-scheduler-lease`
- `make check`
- `go run ./cmd/mhcat-command-sync` with an empty environment
- `go run ./cmd/mhcat-bot` with an empty environment

Results:

- All tests, vet, builds, and `make check` passed.
- Empty-env CLI checks failed safely with missing-config errors.
- Legacy `MHCAT/` source remained unmodified.
- `cmd/mhcat-bot` still does not register or sync Discord commands.
- No Mongo index creation was introduced. Mongo feature writes are limited to the explicit work start/admin paths described above.
- `internal/core/**` still has no DiscordGo or MongoDB driver imports.
- `internal/discord/features/work` and `internal/core/services/work` still have no DiscordGo or MongoDB driver imports.
- Build artifacts generated by verification were removed.

## Next Work

Continue the remaining work-system rollout:

- decide whether completed-but-unsettled legacy rows should still force the old overwrite warning or use `end_time <= now` as a bug fix;
- keep payout and daily reset ownership separate from command handlers.
