# Test Plan

Status: Phase 1 consolidated. Legacy has only a live Discord login smoke test; the Go refactor needs tests before feature implementation claims.

## Legacy Test Inventory

- `npm test` runs `node test-startup.js`.
- `test-startup.js` requires a real `TOKEN`, logs into Discord, waits for ready, then destroys the client.
- No unit tests, fake Discord tests, Mongo repository tests, fixture tests, or CI were found.

## Unit Tests

- Domain logic for XP, level thresholds, economy, gacha, work, polls, birthday, warnings, cron validation, anti-scam normalization.
- Permission policy, owner/admin policy, role hierarchy checks, bot permission requirements.
- Cooldown/rate-limit policy.
- Config/env alias validation and secret redaction.
- Error mapping and safe user-facing errors.
- Custom ID codec and parsers.
- Mongo document conversion and compatibility decoders.
- Discord input parsing without DiscordGo in core.

## Table-Driven Command Tests

- Command metadata snapshot for all 74 slash command modules via `go run ./tools/parity-audit --legacy-root ../MHCAT --format markdown`.
- Command names, descriptions, options, choices, localizations, default permissions, cooldown metadata.
- Handler selection by command name.
- Public vs ephemeral/deferred response metadata.
- Disabled/inactive feature behavior for birthday scheduler and lottery creation.

## Discord Fake Adapter Tests

- Slash command routing.
- Invalid/unknown command.
- Permission denied and missing bot permissions.
- Cooldown hit.
- Owner-only paths and no hardcoded ID bypass.
- Defer, follow-up, edit original, delete original, ephemeral flags.
- Duplicate response prevention.
- Component routing and modal routing.
- Legacy custom ID compatibility and collision/spoofing tests.
- Discord API failure, rate limit, DM failure, channel/role/member mutation failure.
- Panic recovery and safe error response.

## Mongo Repository Contract Tests

- Legacy collection name mapping.
- BSON decode with missing fields, unknown fields, and mixed string/number/bool fields.
- Not found, conflict, timeout, canceled, invalid input error mapping.
- Upsert and `$setOnInsert`.
- Atomic `$inc` for counters.
- Array updates for poll voters, lottery members, warnings, lock users.
- Duplicate handling for candidate unique keys.
- Context timeout and cancellation.
- Node-readable Go writes.

## Mongo Integration Strategy

- Prefer local ephemeral MongoDB when available.
- If no MongoDB is available, run repository contract tests against a fake only for adapter-independent behavior and mark integration tests skipped with reason.
- Do not require production MongoDB for automated tests.

## Behavior Parity Tests

- Given legacy fixtures and equivalent internal events, assert response payload, Mongo side effects, Discord side effects, and user-facing error behavior.
- High-priority parity fixtures:
  - `ping`, `info`, `help`
  - sign-in and coin lookup/rank
  - text/voice XP lookup/rank
  - warning add/list/remove
  - poll create/vote/result/owner/export, including initial-versus-rerender text and component migration
  - announcement config/send/modal/confirmation/relay, including exact permissions, color semantics, six-second state, scalar reads, send-before-delete, and mention suppression
  - verification
  - ticket setup/panel/open/close
  - reaction-role setup/delete/events and button-role modal/add/remove
  - voice room create/lock/delete
  - cron notification setup/list/delete
  - welcome/leave templates
  - gacha draw/shop purchase

The canonical announcement fixtures, focused package commands, race coverage, and staging checklist are recorded in the [announcement parity contract](76-announcement.md). The locked boundary test is `internal/discord/features/announcements/parity_test.go`.

## Golden Fixture Tests

- Command table JSON.
- Custom ID table.
- Sample embeds/components for high-risk commands.
- Welcome/leave/announcement/ticket messages.
- Poll exact initial UI, fixed messages, UTF-16 validation, JavaScript percentage rounding, result/export output, Mongoose scalar reads, atomic predicates, and rollback compensation; canonical commands are in [75-poll.md](75-poll.md#parity-tests).
- Rank/profile/sign images with seeded inputs where deterministic rendering is feasible.

## Context and Timeout Tests

- Mongo operations.
- Discord REST calls.
- Translation calls.
- Canvas/chart/Excel rendering.
- Member/user fetches.
- Clear/bulk delete loops.
- Work/game collectors.
- Modal waits and component collectors.
- Scheduler jobs.

## Race Tests

Run `go test -race ./...` when implementation exists. Race-prone flows:

- concurrent sign-in
- XP updates
- coin-game component/timeout claims, timer generations, and two-player transaction conflicts
- gacha/shop inventory
- poll votes
- lottery joins
- work completion
- scheduler reload
- command registration diff
- concurrent ticket setup and identity-scoped failure rollback

## Benchmark Tests

- Command routing.
- XP hot path.
- Coin balance/rank queries.
- Poll result on large guild.
- Rank/profile rendering.
- Scam URL matching.
- Channel status update planning.

## Fuzz Tests

- Cron expressions.
- Custom ID parsers.
- Poll option parser.
- Lottery duration parser.
- Color/image URL validators.
- Message template placeholders.
- Domain/URL normalization.
- Numeric bounds for coin/XP/work/birthday.

## Manual Staging Checklist

- Staging bot token and staging guild configured.
- Commands registered guild-scoped only.
- Mongo points to staging/sanitized DB.
- `SHADOW_READ_ONLY=true` tested with side effects disabled.
- Canary guild feature ownership confirmed before enabling writes.
- Rollback drill proves Node can read Go-written documents.
- Role-selection smoke follows the duplicate/type audit, one-owner, bot-removal, stale-button, and exact-once usage checks in [the parity contract](73-role-selection.md).
- Ticket smoke follows the duplicate/type audit, one-owner, stale-modal, failure-compensation, overwrite, owner-close, and usage checks in the [ticket parity contract](74-ticket.md).

## Required Checks By Wave

- Wave 1: `go fmt ./...`, `go test ./...`, `go vet ./...`, safe missing-env run.
- Wave 2: domain/ports unit tests and import-boundary tests.
- Wave 3: Mongo repository contracts, compatibility fixtures, index dry-run tests.
- Wave 4: Discord fake adapter, responder, router, permission/cooldown tests.
- Wave 5+: per-feature parity tests before checklist moves to tested/verified.
- Final: `go fmt`, `go test`, `go vet`, `go test -race` if feasible, `go build` for bot and tools, and the slash-command UI parity audit.
