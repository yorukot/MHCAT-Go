# Platform Wave B Mongo Catalog Foundation

## Gate C Review

- Legacy `MHCAT/` source remained read-only.
- No feature repository write was added.
- No Mongo index creation was run or enabled by default.
- No SQL-style migration system was added.
- `internal/core/**` remains free of MongoDB driver and DiscordGo imports.
- Platform Wave A Discord runtime infrastructure remains unchanged in scope.

## Files Created

- `internal/adapters/mongo/catalog.go`
- `internal/adapters/mongo/catalog_test.go`
- `docs/32-platform-wave-b-mongo-catalog.md`

## Files Updated

- `internal/adapters/mongo/audit.go`
- `README.md`
- `docs/02-mongo-collection-catalog.md`
- `docs/03-mongo-index-plan.md`
- `docs/04-data-compatibility-plan.md`
- `docs/05-data-audit-and-repair-plan.md`
- `docs/09-risk-register.md`
- `docs/10-feature-parity-checklist.md`

## Catalog Changes

- `DefaultCollectionCatalog()` now covers all 47 legacy `MHCAT/models/*.js` files.
- Each catalog entry records:
  - expected Mongo collection name;
  - legacy Mongoose model name;
  - legacy model file path;
  - required fields used by audit;
  - logical key candidates;
  - planned index candidates;
  - compatibility notes.
- Corrected the prior Wave 3 singular placeholder entries to Mongoose-compatible names such as `coins`, `text_xps`, `voice_xps`, `polls`, `tickets`, `guilds`, `cron_sets`, `verifications`, and `chatgpts`.
- Known special cases are documented in code notes:
  - `create_hours` remains singular-looking;
  - `role.js` maps to model `role_number` and collection `role_numbers`;
  - `suport`/`suports` preserves the legacy misspelling;
  - dashboard-shared collections are marked in notes.

## Index and Audit Behavior

- `DefaultIndexPlan(DefaultCollectionCatalog())` now uses the full corrected catalog.
- Unique index candidates still require duplicate audit and are marked dangerous unless explicitly allowed and clean.
- TTL index guardrails remain unchanged.
- Unknown remote indexes are still reported and never dropped.
- Audit remains read-only and reports missing/unknown collections against the full catalog.

## Tests Added

- Catalog validation for non-empty collection/model/file metadata.
- Legacy model directory coverage for all 47 `MHCAT/models/*.js` files.
- High-risk Mongoose collection-name assertions.
- Unique-index duplicate-audit guard assertions.
- Catalog lookup map tests by collection, model, and legacy file.
- Default index plan regression test blocking the previous singular scaffold names.

## Known Limitations

- Catalog field coverage is still intentionally incomplete until BSON fixtures and live audit are reconciled.
- No feature-specific BSON document structs were added.
- No feature repositories were implemented.
- No data repair/backfill command was added.
- Live production write safety still depends on read-only audit review, duplicate/null/missing checks, and feature-specific contract tests.

## Commands Run

- `GOCACHE=... go fmt ./internal/adapters/mongo`
- `GOCACHE=... go test ./internal/adapters/mongo`
- `GOCACHE=... go fmt ./...`
- `GOCACHE=... go test ./...`
- `GOCACHE=... go vet ./...`
- `GOCACHE=... go build ./cmd/mhcat-bot`
- `GOCACHE=... go build ./cmd/mhcat-command-sync`
- `GOCACHE=... go build ./cmd/mhcat-mongo-audit`
- `GOCACHE=... go build ./cmd/mhcat-mongo-index`
- `GOCACHE=... go build ./cmd/mhcat-staging-preflight`
- `GOCACHE=... make check`

## Next Recommended Step

Begin the first stateful feature only after selecting a narrow feature group and adding BSON fixture compatibility tests plus repository contract tests for that group. Ticket setup/read-only config or join/leave config are safer first candidates than economy, XP, poll, or scheduler writes.
