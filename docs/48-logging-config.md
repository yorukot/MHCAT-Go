# Logging Config, Message Event, and Channel Event Slices

Status: implemented behind explicit runtime and command-sync gates.

## Legacy Reference

- File: `MHCAT/slashCommands/管理系統/create_logging.js`
- Event file: `MHCAT/events/LoggingSystem.js`
- Model: `MHCAT/models/logging.js`
- Command: `set-log-channel`
- Localized names: `設置日誌`, `设置日志`
- Component: `loggin_create`
- Collection: `loggings`

## Implemented Scope

- `/set-log-channel` command definition with legacy name/description localizations.
- Runtime Manage Messages check with the legacy red error embed.
- Legacy-style yellow `日誌系統` prompt embed.
- Select menu labels, descriptions, placeholder, min value `1`, and max value `4`.
- `loggings` Mongo write compatibility:
  - `guild`
  - `channel_id`
  - `message_update`
  - `message_delete`
  - `channel_update`
  - `member_voice_update`
- Runtime gate: `MHCAT_FEATURE_LOGGING_CONFIG_ENABLED=false` by default.
- Command-sync gate: `MHCAT_COMMAND_SYNC_INCLUDE_LOGGING_CONFIG=false` by default and staging guild only.
- Message update/delete runtime gate: `MHCAT_FEATURE_LOGGING_MESSAGE_EVENTS_ENABLED=false` by default.
- Message update/delete log emitters read `loggings` by `guild`, require the selected legacy toggles, send legacy-style embeds to `channel_id`, include attachment URLs, and use best-effort audit-log attribution for delete events.
- Channel update runtime gate: `MHCAT_FEATURE_LOGGING_CHANNEL_EVENTS_ENABLED=false` by default.
- Channel update log emitters read `loggings` by `guild`, require `channel_update`, send legacy-style topic and permission-overwrite embeds to `channel_id`, and use best-effort audit-log attribution.

## Intentional Safety Fixes

- Legacy deletes the existing `loggings` document and inserts a new one. Go updates all existing `{guild}` duplicates and only upserts when none exist, avoiding a temporary missing-config window and keeping duplicate legacy rows consistent until audit/index work is approved.
- Legacy `loggin_create` relied on a process-local Discord collector closure to remember the selected channel. Go-generated messages use an invisible versioned custom ID with the selected channel ID in the payload. Old orphaned `loggin_create` components are recognized but return a safe rerun message because the channel cannot be recovered.
- The Go select handler uses a deferred message update before saving Mongo config so slow writes do not miss Discord's interaction response deadline.
- Message update/delete event emitters are separately gated, require Gateway, Guild Messages, and Message Content, depend on cached old/deleted message payloads, and suppress allowed mentions in emitted embeds.
- Channel update event emitters are separately gated, require Gateway, depend on cached old channel payloads, compare permission overwrites by ID, and suppress allowed mentions in emitted embeds. Voice-state log sends remain disabled pending their own privacy/rate-limit review.

## Not Implemented

- Full `events/LoggingSystem.js` parity beyond message update/delete and channel topic/permission updates.
- Voice-state log sends.
- Any event queue or feature-specific rate-limit policy beyond the existing Discord client behavior.

## Tests

- Command definition/localization/channel-type tests.
- Handler permission/prompt/select-save/legacy-component tests.
- Service validation tests.
- BSON document compatibility tests.
- Runtime wiring tests.
- Command-sync and staging-preflight gate tests.
- Message update/delete and channel update event conversion and emitter tests.
