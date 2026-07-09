# Wave 3 Notes

Status: Wave 3 complete.

## Gate C Review for Wave 2

- Legacy source status: clean. `MHCAT/` reports `## main...origin/main`.
- Bot command registration: `cmd/mhcat-bot` and `internal/app` have no command registration path.
- Command sync dry-run default: preserved from Wave 2. `mhcat-command-sync` requires explicit `--apply` for writes.
- Command deletion safety: preserved from Wave 2. Deletion requires `--apply --allow-delete`.
- Bulk overwrite safety: preserved from Wave 2. Bulk overwrite requires `--apply --allow-bulk-overwrite`.
- Existing Mongo write/index creation before Wave 3: none found in `internal` or `cmd`.
- Message Content default: disabled by default through `DefaultMessageContentIntent=false`.
- Core boundary: `internal/core/**` has no DiscordGo or MongoDB driver imports.
- Wave 2 issues requiring fix before Wave 3: none found.

## Files Created

- `cmd/mhcat-mongo-audit/main.go`
- `cmd/mhcat-mongo-index/main.go`
- `internal/config/mongo_admin.go`
- `internal/config/mongo_admin_test.go`
- `internal/core/ports/repository.go`
- `internal/core/ports/transaction.go`
- `internal/adapters/mongo/database.go`
- `internal/adapters/mongo/collection.go`
- `internal/adapters/mongo/errors.go`
- `internal/adapters/mongo/errors_test.go`
- `internal/adapters/mongo/health.go`
- `internal/adapters/mongo/audit.go`
- `internal/adapters/mongo/audit_test.go`
- `internal/adapters/mongo/indexes.go`
- `internal/adapters/mongo/indexes_test.go`
- `internal/adapters/mongo/index_diff.go`
- `internal/adapters/mongo/index_diff_test.go`
- `internal/adapters/mongo/atomic.go`
- `internal/adapters/mongo/atomic_test.go`
- `internal/adapters/mongo/transaction.go`
- `internal/adapters/mongo/transaction_test.go`
- `internal/adapters/mongo/repository.go`
- `internal/adapters/mongo/repository_test.go`
- `internal/testutil/fakemongo/audit.go`
- `internal/testutil/fakemongo/indexes.go`
- `internal/testutil/fakemongo/repository.go`
- `internal/testutil/fakemongo/transaction.go`
- `testdata/mongo/audit_sample.json`
- `testdata/mongo/index_plan_valid.json`
- `testdata/mongo/index_plan_with_duplicate_risk.json`
- `testdata/mongo/live_indexes_sample.json`

## Files Updated

- `.env.example`
- `Makefile`
- `README.md`
- `docs/02-mongo-collection-catalog.md`
- `docs/03-mongo-index-plan.md`
- `docs/04-data-compatibility-plan.md`
- `docs/05-data-audit-and-repair-plan.md`
- `docs/06-architecture-decision-records.md`
- `docs/09-risk-register.md`
- `docs/10-feature-parity-checklist.md`
- `docs/11-operational-runbook.md`
- `docs/15-gate-b-architecture-freeze.md`
- `docs/18-wave-3-notes.md`

## Audit CLI Design

- `mhcat-mongo-audit` is read-only.
- It loads Mongo admin config, connects, lists collections, counts documents, lists live indexes, samples field/type shapes, runs duplicate logical key audits from the partial catalog, reports large sampled documents, and prints text or JSON.
- It does not include raw document contents by default.
- It exits before connecting if Mongo URI/database config is missing.
- Mongo URI alias warnings use existing redaction helpers.

## Index Diff Design

- `mhcat-mongo-index` defaults to dry-run.
- Desired indexes come from the partial Wave 3 collection catalog unless `--plan` is provided.
- The pure diff planner emits `create`, `exists`, `changed`, `dangerous`, `skipped`, and `unknown_remote` operation classes.
- Unknown remote indexes are never dropped.
- Changed indexes are dangerous and not modified in Wave 3.
- Unique index creation requires `--apply --allow-unique` and clean duplicate audit.
- TTL index creation requires `--apply --allow-ttl` and a retention ADR/note in the index spec.
- Apply mode only calls `CreateOne` for safe missing indexes from `SafeIndexApplyOperations`.

## Atomic Update Helper Design

- `UpdateBuilder` builds deterministic BSON updates for `$inc`, `$set`, `$setOnInsert`, `$addToSet`, `$pull`, and `$push`.
- Empty update builds fail.
- Empty or operator-prefixed field names fail.
- Conflicting operations on the same field fail.
- This is infrastructure only; no feature repository methods use it yet.

## Repository Boundary Design

- `internal/core/ports/repository.go` defines driver-agnostic paging and health contracts.
- `internal/adapters/mongo/repository.go` provides a small adapter base around a Mongo collection.
- Core ports do not expose `bson`, `mongo.Collection`, or other driver types.
- Feature repositories remain deferred to future waves.

## Transaction Runner Design

- `internal/core/ports/transaction.go` defines `TransactionRunner`.
- `internal/adapters/mongo/transaction.go` wraps Mongo driver sessions in the adapter layer.
- Fake transaction tests cover commit, rollback, and cancellation behavior.
- No feature code uses transactions in Wave 3.
- Transactions remain reserved for true multi-document or multi-collection atomicity.

## Error Mapping Design

- Mongo adapter errors map into stable categories: not found, conflict, timeout, canceled, invalid, transient, and unknown.
- Mapped errors wrap original errors for debugging.
- Safe messages avoid exposing raw Mongo connection details.
- Core packages do not depend on Mongo driver error types.

## Tests Added

- Mongo admin config defaults, dry-run defaults, unique/TTL guard defaults, invalid duration.
- Mongo error mapping for not found, duplicate key, timeout, canceled, validation, transient, unknown, and safe message behavior.
- Audit analyzer tests for deterministic empty report, unknown collection, missing collection, mixed field types, large documents, missing required field, and no raw document values.
- Index plan/diff tests for create, exists, changed, unknown remote, unique duplicate audit requirement, TTL ADR requirement, safe apply operations, and deterministic output.
- Atomic update builder tests for all supported operators and safety failures.
- Repository and transaction contract tests with fake Mongo utilities.

## Commands Run

- `GOCACHE=/private/tmp/mhcat-refactor-gocache GOMODCACHE=/private/tmp/mhcat-refactor-gomodcache go fmt ./...`
- `GOCACHE=/private/tmp/mhcat-refactor-gocache GOMODCACHE=/private/tmp/mhcat-refactor-gomodcache go test ./...`
- `GOCACHE=/private/tmp/mhcat-refactor-gocache GOMODCACHE=/private/tmp/mhcat-refactor-gomodcache go vet ./...`
- `GOCACHE=/private/tmp/mhcat-refactor-gocache GOMODCACHE=/private/tmp/mhcat-refactor-gomodcache go build ./cmd/mhcat-bot`
- `GOCACHE=/private/tmp/mhcat-refactor-gocache GOMODCACHE=/private/tmp/mhcat-refactor-gomodcache go build ./cmd/mhcat-command-sync`
- `GOCACHE=/private/tmp/mhcat-refactor-gocache GOMODCACHE=/private/tmp/mhcat-refactor-gomodcache go build ./cmd/mhcat-mongo-audit`
- `GOCACHE=/private/tmp/mhcat-refactor-gocache GOMODCACHE=/private/tmp/mhcat-refactor-gomodcache go build ./cmd/mhcat-mongo-index`
- `GOCACHE=/private/tmp/mhcat-refactor-gocache GOMODCACHE=/private/tmp/mhcat-refactor-gomodcache make check`
- `MHCAT_MONGODB_URI= MONGOOSE_CONNECTION_STRING= MHCAT_MONGODB_DATABASE= GOCACHE=/private/tmp/mhcat-refactor-gocache GOMODCACHE=/private/tmp/mhcat-refactor-gomodcache go run ./cmd/mhcat-mongo-audit`
- `MHCAT_MONGODB_URI= MONGOOSE_CONNECTION_STRING= MHCAT_MONGODB_DATABASE= GOCACHE=/private/tmp/mhcat-refactor-gocache GOMODCACHE=/private/tmp/mhcat-refactor-gomodcache go run ./cmd/mhcat-mongo-index`

## Known Limitations

- Live Mongo audit was not run because no Mongo URI/database was provided.
- The Wave 3 collection catalog is partial and intentionally marked incomplete.
- The default index plan is conservative; planned unique indexes remain dangerous until live duplicate audit is clean.
- TTL indexes are structurally supported but no TTL index is enabled by default.
- No feature repositories or production feature writes are implemented.
- No index deletion or index modification path is implemented.
- No SQL-style migration system exists.

## Next Recommended Step

Start Wave 4 only after reviewing Wave 3 output. Wave 4 should implement component/modal custom ID parsers and legacy compatibility decoders with golden tests, while keeping feature behavior out of scope unless explicitly authorized.
