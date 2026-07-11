# Work System Parity Contract

Status: parity-audited against `slashCommands/打工系統/work_set.js`, `handler/gift.js`, `models/work_set.js`, `models/work_something.js`, `models/work_user.js`, `models/coin.js`, and current Go definitions, handlers, domain/service boundaries, Mongo adapters, payout worker/CLI, app wiring, guarded Mongo contracts, and race coverage. Live Discord and operator-gated disposable-Mongo smoke remain required before production ownership.

## Scope And Ownership

This contract covers `/打工系統` setup, dashboard redirect, item deletion, public work interface, captcha, role-filtered list/detail/start/override/cancel interactions, individual/all energy grants, `work_sets`, `work_somethings`, `work_users`, and completed-job payout into `coins`. Exact command metadata, option order/text/requirements, advertised permission text, cooldown metadata `5`, emoji, and docs paths are preserved. Legacy did not centrally enforce cooldowns; Go adds no local throttle.

Command runtime is disabled by default behind `MHCAT_FEATURE_WORK_ENABLED=true`. Staging sync separately requires `MHCAT_COMMAND_SYNC_INCLUDE_WORK=true`; preflight rejects unpaired or non-staging sync. Completed-job payout has independent CLI/worker/write/lease gates documented in [43-work-payout.md](43-work-payout.md). Node command and payout ownership must stop before Go owns either write path.

## Command And UI Contract

`新增打工事項` remains the current legacy red dashboard-redirect embed with the same double-slash URL, link label, emoji, and no Mongo write. Setup, delete, and energy commands defer publicly, require Manage Messages, and preserve legacy success/error response classes, text, emojis, colors, mentions, and docs links. Interface remains public.

Interface preserves optional arithmetic captcha modal behavior, exact wrong-answer content, random list/detail colors, guild title, requester footer, typo `按扭`, inline field text, natural Mongo item order, role filtering, five buttons per row and 25-button limit. Go uses versioned requester-bound component IDs instead of unscoped process-local collectors; foreign users are denied privately and stale/missing selections fail closed.

Detail preserves exact item/reward/energy text and confirm button. Start preserves insufficient-energy, busy override, yes/no cancel, success title/body, and exact scalar completion timestamp. Opening the writable interface creates a missing legacy idle user at the configured max; read-only construction synthesizes the same view without writing.

## Admin Data Behavior

Setup accepts signed Discord integers. It deletes one naturally selected `work_sets` row and inserts the new row, matching legacy delete/recreate behavior; untouched duplicates remain. Delete removes one naturally selected `{guild,name}` work item and leaves other duplicates. No command creates, repairs, merges, or normalizes indexes/rows.

Individual and all-user grants accept positive, zero, and negative amounts and clamp only above `max_energy`. Existing `energi` and configured max values use JavaScript-number coercion, preserving decimals, numeric strings, null-as-zero, infinities, and malformed/NaN propagation. The all-user command updates existing guild rows only. Go intentionally fixes the legacy missing-target bug by creating/updating the selected target rather than the invoking admin.

## Start Arithmetic And Scalars

Item `time`, `energy`, and `coin`; config `get_energy`/`max_energy`; and user `end_time`, `energi`, and `get_coin` preserve display text and JavaScript-number arithmetic for migrated BSON scalars. This includes decimal, signed, null, infinity, and malformed values. List duration division emits JavaScript `NaN`/`Infinity` spelling. Interface state and relative timestamps use the preserved `end_time` scalar.

Start atomically checks `Number(energi) >= Number(energy)` when the cost is numeric, then writes `state`, `now + time`, `Number(energi) - Number(energy)`, and reward. A `NaN` item cost follows JavaScript comparison behavior and does not trigger insufficient energy. Existing user scalar conversion and subtraction occur in one Mongo update pipeline.

## Payout Contract

Legacy strict due behavior is `state != "待業中"`, Mongo `end_time <= rounded_now`, then JavaScript `end_time < rounded_now`. Decimal, zero/null, negative, infinity, and NaN values retain that comparison behavior. Rewards and existing balances retain JavaScript-number coercion, including decimal/signed/non-finite values.

Payout selects one natural duplicate coin row like legacy `findOne`, then targets its stable `_id`; other balance duplicates remain unchanged. Duplicate work rows remain independent jobs. State reset changes only `state`, leaving `end_time`, `energi`, and `get_coin` unchanged. No Discord message is sent.

Go intentionally replaces legacy stale read/write payout with an atomic balance increment and deterministic per-work-row marker. CLI/worker leases, exact-snapshot reset, crash retries, and stale-job rejection are safety improvements. The documented `gift_change.time == 0` normalization fix remains intentional. Node does not honor Go markers, so ownership must be exclusive.

## Verification

```bash
go test ./internal/core/domain ./internal/adapters/mongo/documents \
  ./internal/core/services/work ./internal/core/services/economy \
  ./internal/testutil/fakemongo ./internal/adapters/mongo/repositories \
  ./internal/discord/features/work ./internal/app ./cmd/mhcat-work-payout
go test -race ./internal/core/domain ./internal/core/services/work \
  ./internal/core/services/economy ./internal/adapters/mongo/repositories \
  ./internal/discord/features/work ./internal/app ./cmd/mhcat-work-payout
go vet ./...
go run ./tools/parity-audit --legacy-root ../MHCAT --format markdown
```

Guarded scalar/duplicate/order evidence requires a disposable Mongo database:

```bash
MHCAT_RUN_MONGO_INTEGRATION_TESTS=true \
MHCAT_MONGODB_URI='<disposable-uri>' \
go test ./internal/adapters/mongo/repositories \
  -run '^(TestWorkInterfaceMongoIntegration|TestWorkPayoutMongoIntegration)'
```

The harness drops generated databases. Never use production.

## Staging And Rollback

1. Stop Node work command and `handler/gift.js` payout ownership. Back up/audit `work_sets`, `work_somethings`, `work_users`, affected `coins`, `gift_changes`, and payout markers by `_id`, preserving duplicates and BSON types.
2. Enable paired command flags only in an isolated staging guild. Enable payout separately with unique lease owners and reviewed TTL/timeout settings.
3. Verify setup/redirect/delete/captcha/list/detail/start/override/cancel/grant UI, permissions, requester isolation, role visibility, scalar rows, duplicates, and one usage event per slash.
4. Exercise payout strict-boundary, decimal/null/infinity/NaN, duplicate balance/work rows, crash retry, stale marker, two-replica lease contention, and state-change conflict cases.

Rollback by disabling command sync/runtime and payout gates before restoring Node. Restore related collections from `_id` backups; retain or explicitly account for Go payout markers so Node/Go transitions cannot replay credits. Do not merge duplicates, normalize scalar types, create unique indexes, or infer balances/energy from another collection during rollback.

Production remains gated on live Discord smoke, disposable Mongo execution, backup/restore rehearsal, scalar/duplicate audit, exclusive command and payout ownership, and explicit economy-write approval.
