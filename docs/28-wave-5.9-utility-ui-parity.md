# Wave 5.9 Utility UI Parity

Status: implemented and verified locally.

## Trigger

After restoring the legacy `/help` menu in Wave 5.8, the remaining enabled utility command with a rebuilt interface was `/info bot`. Legacy `info.js` renders an embed with system fields and a green `botinfoupdate` refresh button.

## Legacy References

Read-only legacy files used:

- `MHCAT/slashCommands/實用工具/info.js`
- `MHCAT/events/btn.js`
- `MHCAT/config.json`
- `docs/12-component-modal-grammar.md`

## Runtime Behavior Changed

`/info bot` now uses the legacy-style response shape:

- deferred slash interaction;
- follow-up embed titled `<a:mhcat:996759164875440219> MHCAT目前系統使用量:`;
- legacy field labels for CPU model, CPU usage, shard count, RAM usage, uptime, total guilds, and total users;
- legacy green refresh button with custom ID `botinfoupdate`, emoji `<:update:1020532095212335235>`, and label `更新`;
- safe red error embed when bot info cannot be collected.

The `botinfoupdate` legacy component route is now registered through the typed custom ID parser. It updates the existing message through the responder state machine and sends an ephemeral success follow-up matching the legacy `done` emoji text. The refresh path also preserves the legacy `集群數量` label used in `events/btn.js`.

## Files Updated

- `internal/core/ports/botinfo.go`
- `internal/core/services/utility/status.go`
- `internal/adapters/discordgo/botinfo.go`
- `internal/adapters/discordgo/responder.go`
- `internal/adapters/discordgo/responder_test.go`
- `internal/discord/responses/responder.go`
- `internal/discord/responses/state.go`
- `internal/discord/responses/state_test.go`
- `internal/discord/runtime/dispatcher.go`
- `internal/discord/features/utility/legacy_info.go`
- `internal/discord/features/utility/status_handler.go`
- `internal/discord/features/utility/status_handler_test.go`
- `internal/discord/features/utility/module.go`
- `internal/app/wiring_test.go`
- `internal/testutil/fakediscord/responder.go`
- `README.md`
- `docs/09-risk-register.md`
- `docs/10-feature-parity-checklist.md`
- `docs/20-wave-5.1-notes.md`
- `docs/21-wave-5.2-notes.md`
- `docs/28-wave-5.9-utility-ui-parity.md`

## What Is Not Implemented

No new slash command definitions were added. These remain intentionally out of scope:

- `/info user`;
- `/info guild`;
- command usage Mongo writes;
- production/global command sync;
- any non-utility feature group.

Follow-up: Wave 5.10 implements `/info shard` and `shardinfoupdate`.

CPU model, CPU usage, and memory values are provided by the Go runtime/system provider where available. The UI labels match legacy even when a value is a safe fallback.

## Tests Added or Updated

- `/info bot` handler asserts the legacy embed fields and refresh button.
- Degraded provider path asserts the legacy red error embed without leaking internal errors.
- `botinfoupdate` handler asserts message update plus ephemeral success follow-up.
- `botinfoupdate` route is tested through the legacy custom ID parser and router.
- Responder state tests cover update-message then follow-up.
- DiscordGo responder tests cover timestamp and success button conversion.

## Commands Run

- `go test ./internal/discord/features/utility`
- `go test ./internal/discord/responses`
- `go test ./internal/adapters/discordgo`
- `go test ./internal/app`

## Operational Note

Restart the staging bot process after deploying this change. No command sync apply is required because the `/info bot` command definition did not change.
