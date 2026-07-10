# Data Audit and Repair Plan

Status: Phase 1.5 Gate B review. No production data mutation is authorized by default.

## Principles

- Audit first.
- Dry-run before any write.
- Explicit `--apply` required for every write operation.
- Backup or restore point required before production repair/backfill.
- No automatic destructive changes.
- Output must be reviewable and must not print secrets.
- Repairs must preserve Node.js rollback compatibility unless an ADR says otherwise.

## Audit Commands

Planned:

- `mhcat-tools mongo ping`
- `mhcat-tools mongo list-collections`
- `mhcat-tools mongo check-indexes`
- `mhcat-tools mongo audit`
- `mhcat-tools data validate --collection <name>`
- `mhcat-tools data audit-types --collection <name>`
- `mhcat-tools data audit-duplicates --collection <name> --keys <keys>`

Available now:

- `node MHCAT-REFACTOR/tools/mongo-audit-readonly.mjs`
- `go run ./cmd/mhcat-mongo-audit --format text|json`

Required env:

```txt
MONGODB_URI or MONGOOSE_CONNECTION_STRING
MONGODB_DATABASE
```

The current tool is read-only. It does not create indexes, update documents, delete documents, or write repair plans to MongoDB. If the `mongodb` npm package is missing, install it outside production first and rerun the audit in a controlled environment.

Read-only audit output includes:

- database name and redacted URI;
- expected collection count and live collection count;
- missing expected collections;
- collections not represented by Mongoose models;
- per-collection estimated document count;
- current indexes;
- sample document shapes;
- missing required field counts for hot collections;
- mixed type counts for sampled fields;
- duplicate logical key samples for planned unique candidates;
- duplicate logical key risk lines from the Go audit for catalogued unique candidates, including `coins_guild_member` and `sign_lists_guild_member`;
- impossible negative values for known counters;
- large documents over 256 KiB.

Run order:

1. Run in staging against a restored production snapshot.
2. Review missing/extra collection list, especially `message_reaction`, `userdatas`, and dashboard collections.
3. Review duplicate logical keys before any unique index plan.
   - For economy sign-in readiness, `coins_guild_member` and `sign_lists_guild_member` must both report zero duplicate groups.
   - For logging config readiness, `loggings.guild` duplicate groups should be reviewed before any future unique index or event-emitter rollout.
4. Review mixed types before choosing Go BSON strictness.
5. Run read-only against production during a low-traffic window.
6. Attach the redacted JSON summary to the rollout checklist; do not paste secrets or raw sensitive user content into docs.

## Repair / Backfill Commands

Planned:

- `mhcat-tools data repair --name <repair> --dry-run`
- `mhcat-tools data repair --name <repair> --apply`
- `mhcat-tools data backfill --name <backfill> --dry-run`
- `mhcat-tools data backfill --name <backfill> --apply`

No repair/backfill command is authorized at Gate B. Repair tooling may be added later, but it must default to dry-run and require explicit `--apply`.

## Repair Plan

| Repair name | Problem | Detection query | Dry-run output | Write behavior | Safety guard | Backup requirement | Rollback | Should run automatically |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `audit-collection-names` | Code-inferred names may differ from production | list collections and compare expected names | missing/extra collection report | none | read-only | no | none | no |
| `audit-indexes` | Existing production indexes unknown | list indexes per collection | existing/planned/diff report | none | read-only | no | none | no |
| `audit-live-compatibility` | live DB may differ from code-inferred catalog | run `tools/mongo-audit-readonly.mjs` | JSON report of collections/counts/indexes/shapes/types/duplicates | none | read-only | no | none | no |
| `audit-singleton-duplicates` | guild singleton collections may have duplicates | aggregate group by `guild` count > 1 | duplicate groups and sample `_id`s | none | read-only | no | none | no |
| `audit-user-duplicates` | per-user collections may duplicate `guild/member` or `guild/user` | aggregate group by compound key count > 1 | duplicate groups and totals | none | read-only | no | none | no |
| `audit-mixed-types` | string/number/bool drift | `$type` aggregation by field | type counts and samples | none | read-only | no | none | no |
| `normalize-scam-domains` | unsafe regex URL matching | scan `not_a_good_webs.web` | normalized proposal only | optional `$set` canonical field after ADR | explicit apply | yes | restore backup or unset new field | no |
| `dedupe-singleton-config` | duplicate singleton config rows | duplicate audit output | selected winner proposal | optional archive/mark/delete after ADR | explicit apply plus sample approval | yes | restore archived docs | no |
| `backfill-canonical-numeric-fields` | future schema may need canonical numbers | mixed type audit | proposed canonical field writes | optional `$set` new fields, not replacing legacy | explicit apply | yes | unset backfilled fields | no |
| `repair-undefined-command-counts` | `all_use_count` may contain undefined command names | find missing/null/undefined names | candidate rows | optional delete/archive after ADR | explicit apply | yes | restore archived docs | no |
| `audit-logging-duplicates` | duplicate `loggings` singleton rows can make message event emitters choose different configs than setup writes | aggregate group by `guild` count > 1 | duplicate guild groups and sample `_id`s | none | read-only | no | none | no |

## Work Payout Audit Checklist

Before running `mhcat-work-payout --apply`, perform audit-only checks:

| Audit | Problem | Detection | Write behavior | Automatic repair |
| --- | --- | --- | --- | --- |
| Duplicate work users | Multiple `work_users` rows for the same `{guild,user}` can duplicate or mask payouts | group by `{guild,user}` with count > 1 | none | no |
| Duplicate coin balances | Multiple `coins` rows for `{guild,member}` can make `$inc` target an arbitrary duplicate until uniqueness is resolved | group by `{guild,member}` with count > 1 | none | no |
| Duplicate gift config | Multiple `gift_changes` rows for a guild can make new-balance `today` behavior ambiguous | group by `guild` with count > 1 | none | no |
| Oversized gacha prize pools | More than 25 `gifts` rows for one guild can make the legacy single embed fail Discord validation | group `gifts` by `guild` with count > 25 | none | no; Go read-only prize-list splits embeds as a runtime compatibility fix |
| Gacha prize numeric drift | `gift_chence` and `gift_count` may be strings, numbers, null, or impossible values | sample `gifts` field types and range-check negative chance/count | none | no |
| Duplicate gacha prize names | Multiple `gifts` rows for `{guild,gift_name}` make legacy edit/delete affect one arbitrary row | group `gifts` by `{guild,gift_name}` with count > 1 | none | no; gated edit/delete preserve one-row legacy behavior until duplicate cleanup is approved |
| Duplicate text XP channel config | Multiple `text_xp_channels` rows for one guild can make future XP announcement behavior ambiguous | group `text_xp_channels` by `guild` with count > 1 | none | no; Go config writes update/delete all duplicates until audit/unique-index work is approved |
| Text XP config stale background | Legacy set command deletes/reinserts and drops `background`; old duplicate rows may retain mixed `background` values | sample `text_xp_channels.background` presence/type by guild | none | no; Go set unsets `background` on matched duplicate rows |
| Duplicate voice XP channel config | Multiple `voice_xp_channels` rows for one guild can make future voice XP announcement behavior ambiguous | group `voice_xp_channels` by `guild` with count > 1 | none | no; Go config writes update/delete all duplicates until audit/unique-index work is approved |
| Voice XP config stale background | Legacy command exposes `背景` but does not save it; old duplicate rows may retain mixed `background` values | sample `voice_xp_channels.background` presence/type by guild | none | no; Go set unsets `background` on matched duplicate rows |
| Duplicate join role config | Multiple `join_roles` rows for the same `{guild,role}` can block the future unique index and make member-add role assignment repeat work | group `join_roles` by `{guild,role}` with count > 1 | none | no; Go delete removes duplicate `{guild,role}` rows only when an operator deletes that specific config |
| Due non-idle rows | Large backlog can exceed one lease window | count `work_users` where `state != "待業中"` and `end_time < now` | none | no |
| Mixed numeric types | `$inc` can fail if existing `coins.coin` is non-numeric; loose `end_time` or `get_coin` can differ from legacy Mongoose casting | sample field types for `coins.coin`, `work_users.end_time`, `work_users.get_coin` | none | no |
| Suspicious rewards | Zero or negative `get_coin` is legacy-compatible but may indicate bad data | sample due rows where normalized `get_coin <= 0` | none | no |

Any repair must follow the standard flow: audit first, dry-run report, explicit `--apply`, backup requirement, rollback note, and no automatic production mutation.

## Index Bootstrap Guardrails

- Safe mode on app startup may check Mongo connectivity and compare expected index definitions, but it must not create production indexes by default.
- `mhcat-command-sync` and `mhcat-mongo-audit` must be separate operational commands from shard startup.
- Unique indexes require duplicate audit for the exact key, missing/null review, and dashboard compatibility review.
- TTL indexes require a data-retention ADR and explicit operator approval.
- Large collection indexes require a maintenance-window note and rollback/drop guidance.

## No Automatic Production Mutation

These actions are blocked until a future ADR and explicit dry-run/apply tooling exist:

- deduplicating singleton configs;
- changing field types in place;
- deleting unknown collections;
- deleting stale dynamic voice rows;
- normalizing anti-scam URLs in place;
- creating unique indexes;
- adding TTL indexes;
- changing dashboard-shared collection schemas.

## Manual Review Checkpoints

- Confirm target database and collection.
- Confirm backup/restore point exists.
- Confirm dry-run count and sample output.
- Confirm affected shard/feature flags are paused if needed.
- Confirm Node.js rollback compatibility.
- Confirm rollback plan.

## Safety Defaults

- Tools default to dry-run.
- `--apply` refuses to run unless `BOT_ENV`, database name, and operation name are printed and confirmed.
- Destructive operations require an additional confirmation flag.
- Repair output redacts secrets and truncates user content samples.

## Wave 3 Audit and Index Commands

Available:

```bash
go run ./cmd/mhcat-mongo-audit --format json
go run ./cmd/mhcat-mongo-index --dry-run --format json
```

`mhcat-mongo-audit` is read-only and reports:

- collections;
- document counts;
- live indexes;
- sampled field/type shapes;
- required-field gaps from the partial catalog;
- mixed field types;
- duplicate logical key risks where configured;
- large sampled documents;
- unknown and missing catalog collections.

`mhcat-mongo-index` is dry-run by default and reports:

- missing planned indexes;
- existing matching indexes;
- changed/dangerous indexes;
- unknown remote indexes.

Still not implemented in Wave 3:

- data repair writes;
- backfill writes;
- index deletion;
- feature repository writes;
- SQL-style migrations.

## Platform Wave B Audit Note

Platform Wave B expands the default Go collection catalog from the earlier high-risk subset to all 47 legacy Mongoose model files. This improves read-only audit coverage:

- missing expected collection reporting now covers every model-backed collection;
- unknown live collections are easier to identify against the full catalog;
- duplicate-audit candidates are generated from corrected plural collection names;
- dashboard-shared and external-worker-risk collections remain flagged in catalog notes.

No repair/backfill command was added in Platform Wave B. No automatic production mutation is authorized.

## Ticket Config Repository Foundation Note

Ticket config repository code was added after Platform Wave B, but it is not wired to runtime handlers. No production repair/backfill is needed for this foundation step.

Additional audit before enabling ticket writes:

- duplicate `tickets` documents grouped by `guild`;
- missing/empty `ticket_channel`, `admin_id`, and `everyone_id`;
- stale category/role IDs in staging guilds;
- dashboard/external usage confirmation if a future dashboard ticket panel exists.
