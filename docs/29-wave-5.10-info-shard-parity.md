# Wave 5.10 Info Shard UI Parity

Status: historical implementation note. Superseded by the parity-audited [built-in utility contract](94-utility-builtins.md). The earlier intentional immediate-field change was reverted: the initial shard embed is empty and fields appear only after `shardinfoupdate`.

## Trigger

After Wave 5.9 restored `/info bot`, the next low-risk read-only legacy UI gap in the same command was `/info shard`.

## Legacy References

Read-only legacy files used:

- `MHCAT/slashCommands/實用工具/info.js`
- `MHCAT/events/btn.js`
- `docs/12-component-modal-grammar.md`

## Runtime Behavior Changed

`/info shard` now uses the legacy-style response shape:

- deferred slash interaction;
- follow-up embed titled `<:vagueness:999527612634374184> 以下是每個分片的資訊!!`;
- initial embed has no shard field;
- green `更新` button with custom ID `shardinfoupdate`;
- shard fields appear only after the refresh button, matching legacy.

The `shardinfoupdate` legacy component route is now registered through the typed custom ID parser. It updates the existing message with the same shard field shape:

- `分片ID`;
- `公會數量`;
- `使用者數量`;
- `記憶體`;
- `上線時間`;
- `延遲`.

## Command Definition Impact

The local `info` command definition now includes the legacy `shard` subcommand. Bot startup still does not sync commands. Staging must run command sync dry-run and explicit staging apply before `/info shard` appears in Discord.

## Files Updated

- `internal/discord/commands/builtin_registry.go`
- `internal/discord/commands/builtin_registry_test.go`
- `internal/discord/features/utility/legacy_info.go`
- `internal/discord/features/utility/status_handler.go`
- `internal/discord/features/utility/status_handler_test.go`
- `internal/discord/features/utility/module.go`
- `internal/app/wiring_test.go`
- `README.md`
- `docs/10-feature-parity-checklist.md`
- `docs/23-staging-smoke-runbook.md`
- `docs/24-wave-5.4-staging-results.md`
- `docs/28-wave-5.9-utility-ui-parity.md`
- `docs/29-wave-5.10-info-shard-parity.md`

## What Is Not Implemented In Wave 5.10

- production/global command sync;
- command sync apply in this wave;
- Mongo writes or usage counter writes.

Follow-up: Wave 5.11 implements the remaining read-only legacy `/info user` and `/info guild` subcommands using driver-agnostic Discord snapshot providers and user-option parsing.

## Intentional Behavior Fix

Legacy `info.js` initially sends `/info shard` with only a title and update button, then adds the fields only after `shardinfoupdate`. The final parity audit reverted the earlier UX change and preserves this sequence.

## Tests Added or Updated

- Built-in registry test asserts `info shard` is present.
- `/info shard` handler asserts the legacy title and refresh button.
- `shardinfoupdate` handler asserts message update with shard field.
- `shardinfoupdate` route is tested through the legacy custom ID parser and router.
- App runtime wiring test asserts `/info shard` routes end to end.

## Commands Run

- `go test ./internal/discord/features/utility`
- `go test ./internal/app`
- `go test ./internal/discord/commands`

## Operational Note

Run staging command sync dry-run before applying this definition change. No production/global sync, delete, or bulk overwrite should be used.
