# Operational And Migration Tools Contract

Status: implementation-audited against the current CLI sources, 47-model collection catalog, index plan, staging scripts, configuration guards, unit tests, parity report, and repository runbooks. These tools have no legacy bot equivalent and are intentionally excluded from Discord UI parity.

## Tool Inventory

- `mhcat-command-sync` compares typed definitions with Discord, defaults to dry-run, and requires explicit apply/delete acknowledgements.
- `mhcat-mongo-audit` performs read-only collection, field-shape, type, duplicate-key, index, and large-document reporting in text or JSON.
- `mhcat-mongo-index` defaults to index diff only. Apply is explicit; unique and TTL indexes require additional gates, and unique candidates require a clean duplicate audit.
- `mhcat-staging-preflight` validates staging guild scope, paired runtime/sync flags, gateway intents, scheduler gates, and write prerequisites before rollout commands run.
- `mhcat-economy-reset` and `mhcat-work-payout` default to dry-run and require explicit write plus lease gates for apply.
- `mhcat-scheduler-lease` defaults to status and requires explicit apply for acquire, renew, or release.
- `tools/parity-audit` statically compares all legacy slash definitions with the Go catalog and reports drift, missing, extra, and parse failures.

Bot startup invokes none of the command, index, repair, or audit mutations. Operational commands are separate processes with bounded contexts and redacted configuration errors.

## Migration Contract

The Mongo catalog covers all 47 legacy Mongoose model files and preserves collection names shared with Node, dashboards, and external workers. Audit output is the authoritative input to migration decisions; an empty or absent collection in one environment is not evidence that it may be removed globally.

Index application is optional for runtime parity. No feature startup creates an index. Unique indexes may be applied only after exact-key duplicate, missing/null, scalar-type, and external-writer review. TTL indexes additionally require an approved retention decision. The tool creates only reviewed safe missing indexes and does not delete unknown or changed indexes.

Automatic data repair and backfill are intentionally not implemented. Legacy duplicates and mixed scalars often affect observable first-row, multi-row, or coercion semantics, so a generic repair command could break both Go parity and Node rollback. Any production repair must be a named, reviewed migration with audit evidence, dry-run output, explicit apply, backup, rollback, and collection-specific winner rules.

## Operator Workflow

1. Restore a production snapshot into an isolated database and run the audit in JSON mode.
2. Review collections, scalar shapes, duplicate logical keys, indexes, oversized documents, and external ownership.
3. Run command-sync, index, reset, and payout dry-runs as applicable; retain redacted reports.
4. Rehearse backup/restore and Node-to-Go ownership handoff before any apply.
5. Apply only the specifically approved operation with staging scope and explicit gates.
6. Re-audit after mutation and retain the before/after evidence for rollback.

Never point guarded integration tests at production. Never use index creation or data cleanup merely to make a parity test pass.

## Verification

```bash
go test ./cmd/mhcat-command-sync ./cmd/mhcat-mongo-audit ./cmd/mhcat-mongo-index \
  ./cmd/mhcat-staging-preflight ./cmd/mhcat-economy-reset \
  ./cmd/mhcat-work-payout ./cmd/mhcat-scheduler-lease ./tools/parity-audit
go test ./internal/adapters/mongo ./internal/parity ./internal/config
go run ./tools/parity-audit --legacy-root ../MHCAT --format markdown
go vet ./...
```

Live Mongo audit/index output and Discord command-sync diff remain environment-specific operator evidence and are not fabricated by automated tests.
