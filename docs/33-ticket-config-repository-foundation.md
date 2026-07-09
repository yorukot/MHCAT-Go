# Ticket Config Repository Foundation

## Gate C Review

- Legacy `MHCAT/` source remained read-only.
- No ticket command handler or component handler was implemented.
- No Discord channel creation or role permission mutation was added.
- No production Mongo write path was wired into runtime.
- `cmd/mhcat-bot` still does not register or sync commands.
- `internal/core/**` remains free of DiscordGo and MongoDB driver imports.

## Legacy Findings

- `models/ticket.js` stores `guild`, `ticket_channel`, `admin_id`, and `everyone_id`.
- Legacy `/私人頻道設置` writes the ticket config before the modal color/title/content is submitted and validated.
- Legacy modal routing uses broad `text.includes("ticket")`.
- Legacy button `tic` creates a private text channel under `ticket_channel` and grants permissions to `admin_id`, the opener, and the bot.
- This wave intentionally does not copy the runtime side effects or UI yet.

## Files Created

- `internal/core/domain/ticket.go`
- `internal/core/ports/ticket.go`
- `internal/adapters/mongo/documents/ticket.go`
- `internal/adapters/mongo/documents/ticket_test.go`
- `internal/adapters/mongo/repositories/ticket.go`
- `internal/adapters/mongo/repositories/ticket_test.go`
- `internal/testutil/fakemongo/ticket.go`
- `internal/core/ports/ticket_test.go`
- `testdata/mongo/ticket_config_legacy.json`
- `docs/33-ticket-config-repository-foundation.md`

## Repository Design

- Core domain type: `domain.TicketConfig`.
- Core port: `ports.TicketConfigRepository`.
- Mongo BSON document preserves exact legacy field names:
  - `guild`
  - `ticket_channel`
  - `admin_id`
  - `everyone_id`
- Mongo adapter uses the corrected collection name `tickets`.
- Save uses a targeted upsert with `$set` and `$setOnInsert`; no full-document replacement.
- Delete returns `ports.ErrTicketConfigNotFound` when no config exists.
- Reads decode missing fields safely; writes validate all required fields.

## Tests Added

- Legacy BSON fixture decode test.
- Domain/document round-trip test.
- Missing-field decode safety test.
- Fake repository contract test for save, update, get, delete, not-found, validation, and context cancellation.
- Mongo adapter constructor and collection-name tests.

## Intentional Non-Implementation

- No `/私人頻道設置` command handler.
- No `/私人頻道刪除` command handler.
- No `tic` button handler.
- No modal handler.
- No Discord channel create/delete.
- No ticket panel embed/button UI.
- No runtime wiring to this repository.
- No production Mongo write occurred.

## Next Recommended Step

Implement ticket setup UI parity behind tests:

- register only local command definitions, not Discord sync;
- preserve legacy command names/options;
- use the Wave 4 typed modal/component parser rather than `includes`;
- fix the old premature-config-write bug by saving only after modal validation succeeds;
- use safe allowed mentions instead of sending a real `@everyone` ping.
