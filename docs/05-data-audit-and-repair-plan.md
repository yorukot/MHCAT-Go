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
| `audit-logging-duplicates` | duplicate `loggings` singleton rows can make logging event emitters choose different configs than setup writes | aggregate group by `guild` count > 1 | duplicate guild groups and sample `_id`s | none | read-only | no | none | no |

## Work Payout Audit Checklist

Before running `mhcat-work-payout --apply` or enabling the recurring worker, perform audit-only checks:

| Audit | Problem | Detection | Write behavior | Automatic repair |
| --- | --- | --- | --- | --- |
| Duplicate work users | Multiple `work_users` rows for the same `{guild,user}` are each treated as independent jobs keyed by `_id`, but may still indicate legacy corruption | group by `{guild,user}` with count > 1 | none | no; Go markers prevent token collision but do not choose a repair winner |
| Duplicate coin balances | Multiple `coins` rows for `{guild,member}` make the canonical balance ambiguous | group by `{guild,member}` with count > 1 | none | no; Go payout fails with `ErrWorkPayoutCoinConflict` before crediting the affected job |
| Duplicate gift config | Multiple `gift_changes` rows for a guild can make new-balance `today` behavior ambiguous | group by `guild` with count > 1 | none | no |
| Oversized gacha prize pools | More than 25 `gifts` rows for one guild can make the legacy single embed fail Discord validation | group `gifts` by `guild` with count > 25 | none | no; Go read-only prize-list splits embeds as a runtime compatibility fix |
| Gacha prize numeric drift | `gift_chence` and `gift_count` may be strings, numbers, null, or impossible values | sample `gifts` field types and range-check negative chance/count | none | no |
| Duplicate gacha prize names | Multiple `gifts` rows for `{guild,gift_name}` make legacy edit/delete affect one arbitrary row | group `gifts` by `{guild,gift_name}` with count > 1 | none | no; gated edit/delete preserve one-row legacy behavior until duplicate cleanup is approved |
| Duplicate text XP channel config | Multiple `text_xp_channels` rows for one guild can make future XP announcement behavior ambiguous | group `text_xp_channels` by `guild` with count > 1 | none | no; Go config writes update/delete all duplicates until audit/unique-index work is approved |
| Text XP config stale background | Legacy set command deletes/reinserts and drops `background`; old duplicate rows may retain mixed `background` values | sample `text_xp_channels.background` presence/type by guild | none | no; Go set unsets `background` on matched duplicate rows |
| Duplicate voice XP channel config | Multiple `voice_xp_channels` rows for one guild can make future voice XP announcement behavior ambiguous | group `voice_xp_channels` by `guild` with count > 1 | none | no; Go config writes update/delete all duplicates until audit/unique-index work is approved |
| Voice XP config stale background | Legacy command exposes `背景` but does not save it; old duplicate rows may retain mixed `background` values | sample `voice_xp_channels.background` presence/type by guild | none | no; Go set unsets `background` on matched duplicate rows |
| Duplicate join role config | Multiple `join_roles` rows for the same `{guild,role}` can block the future unique index and make member-add role assignment repeat work | group `join_roles` by `{guild,role}` with count > 1 | none | no; Go delete removes duplicate `{guild,role}` rows only when an operator deletes that specific config |
| Duplicate reaction-role mappings | Multiple `message_reactions` rows for `{guild,message,react}` can disagree on `role` and block `message_reactions_guild_message_react` | group by `{guild,message,react}` with count > 1; sample all four field BSON types; separately inventory singular `message_reaction` | none | no automatic repair; reviewed setup aligns matched plural rows and explicit delete removes matched plural rows, but singular data is preserved for external compatibility |
| Duplicate role-button mappings | Multiple `btns` rows for `{guild,number}` can disagree on `role` and block `btns_guild_number` | group by `{guild,number}` with count > 1; sample `guild`, `number`, and `role` BSON types, including exponent-form IDs | none | no automatic repair; reviewed setup aligns matched rows, and stale panels must remain auditable |
| Due non-idle rows | Large backlog can exceed one CLI/worker lease window | count `work_users` where `state != "待業中"` and `end_time < now` | none | no; the recurring worker resumes remaining idempotent rows on the next minute tick |
| Mixed numeric types | the atomic payout pipeline fails if existing `coins.coin` is non-numeric; loose `end_time` or `get_coin` can differ from legacy Mongoose casting | sample field types for `coins.coin`, `work_users.end_time`, `work_users.get_coin` | none | no |
| Payout marker shape | malformed or externally modified `coins.mhcat_work_payouts` entries can block safe token comparison | sample object types and require marker values to contain string `token` plus numeric `end_time` | none | no; fail closed and inspect before repair |
| Suspicious rewards | Zero or negative `get_coin` is legacy-compatible but may indicate bad data | sample due rows where normalized `get_coin <= 0` | none | no |

Any repair must follow the standard flow: audit first, dry-run report, explicit `--apply`, backup requirement, rollback note, and no automatic production mutation.

Role-selection audit and ownership rules are specified in the [role-selection parity contract](73-role-selection.md). Neither candidate unique index is required for runtime and neither may be applied during feature enablement or rollback.

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

## Ticket Config Rollout Audit

Ticket runtime is parity-audited but remains disabled by default. No production repair/backfill or automatic index mutation is authorized.

Additional audit before enabling ticket writes:

- duplicate `tickets` documents grouped by `guild`;
- non-string/missing/empty `guild`, `ticket_channel`, `admin_id`, and `everyone_id` values;
- stale category/role IDs in staging guilds;
- dashboard/external usage confirmation and exclusive Node/Go ticket ownership;
- clean duplicate results before any reviewed `tickets_guild` unique-index apply.

Do not rewrite mixed legacy scalars or deduplicate production rows merely to enable Go. Runtime reads preserve Mongoose-compatible values, explicit config deletion removes duplicate guild rows, and new creates do not overwrite an existing match. Follow the [ticket parity contract](74-ticket.md) for staging and rollback.

## Poll Rollout Audit

Poll runtime is parity-audited but remains disabled by default. No production repair/backfill, TTL, or automatic index mutation is authorized.

Additional audit before enabling poll writes:

- duplicate `polls` documents grouped by `{guild,messageid}`;
- non-string/missing/empty `guild` and `messageid` keys;
- Mongoose scalar drift in `question`, `create_member_id`, `many_choose`, and all four Boolean fields;
- scalar, malformed, duplicate, empty, overlong, or non-string `choose_data` values;
- malformed/non-string `{id,choise,time}` entries and oversized `join_member` arrays/documents;
- Guild Members intent/member-list access, exclusive Node/Go ownership, and a reviewed plan for Go versioned component IDs before rollback;
- clean duplicate/type/shape results before any reviewed `polls_guild_message` unique-index apply.

Do not rewrite loose arrays or deduplicate production rows merely to enable Go. Runtime reads are permissive, writes preserve the legacy field shape, and no channel ID exists for automatic bulk message rewrites. Follow the [poll parity contract](75-poll.md) for staging and rollback.

## Announcement Rollout Audit

Announcement runtime is parity-audited but remains disabled by default. No production repair, deduplication, backfill, or automatic index mutation is authorized.

Audit before enabling any announcement owner:

- duplicate `guilds` documents grouped by `guild` and duplicate `ann_all_sets` documents grouped by `{guild,announcement_id}`;
- missing, null, compound, blank, and non-string keys plus Mongoose-compatible scalar drift in all announcement String fields;
- unsupported or malformed stored colors and null/compound `tag` or `title` values;
- all dashboard/external writers of shared `guilds` fields, especially `voice_detection`, and whether they replace whole documents;
- exclusive Node/Go ownership for config, send/modal/component, and relay route families;
- clean duplicate/type/shared-writer findings before any reviewed `guilds_guild` or `ann_all_sets_guild_announcement` unique-index apply.

Do not normalize raw values or repair malformed production rows merely to enable Go. Runtime reads are permissive, writes are typed patches, and unsupported relay colors leave originals unchanged. Follow the [announcement parity contract](76-announcement.md) for smoke and rollback.

## Anti-Scam Rollout Audit

Anti-scam runtime is parity-audited but disabled by default. No production normalization, repair, deduplication, backfill, or automatic index mutation is authorized.

Audit before enabling any owner:

- duplicate/missing/null/non-string `good_webs.guild` keys and Mongoose Boolean scalar drift in `open`;
- duplicate, blank, missing, null, scalar-drift, and compound `not_a_good_webs.web` values;
- raw URL variants, surrounding whitespace, regex metacharacters, overlong values, and catalog entries that match bot/warning content;
- every external/dashboard catalog/config writer and exclusive Node/Go command/event ownership;
- clean findings before any reviewed `good_webs_guild` or `not_a_good_webs_web` unique-index apply.

Do not normalize catalog values merely to enable Go; raw values are behavioral data. Follow the [anti-scam parity contract](77-anti-scam.md) for staging and rollback.

## Birthday Rollout Audit

Birthday command runtime is parity-audited but disabled by default. The inactive delivery block is not a migration target, and no production repair, deduplication, backfill, or automatic index mutation is authorized.

Audit before enabling command ownership:

- duplicate, missing, null, blank, and scalar-drift `birthday_sets.guild` and `{birthdays.guild,birthdays.user}` keys;
- Mongoose scalar drift and compound values across config String/Boolean fields and profile Number/Boolean fields;
- null, fractional, overflowing, or out-of-range birthday/time values;
- stale channel, role, and user IDs plus every dashboard/external writer;
- exclusive Node/Go command ownership and clean findings before any reviewed index apply.

Do not rewrite malformed rows merely to enable Go. Runtime reads are permissive, writes are typed and duplicate-safe, and delivery remains inactive. Follow the [birthday parity contract](78-birthday.md) for staging and rollback.

## Account-Age Rollout Audit

Account-age config and policy are parity-audited but disabled by default. No production repair, deduplication, backfill, or automatic index mutation is authorized.

Audit before enabling either owner:

- duplicate, missing, null, blank, and scalar-drift `create_hours.guild` keys;
- raw String/scalar/compound `hours` and `channel` shapes, including fractional, exponent, non-finite, zero, negative, and overlong values;
- stale/non-guild log channel IDs and cache availability;
- every external/dashboard writer and exclusive Node/Go config/event ownership;
- clean findings before any reviewed `create_hours_guild` index apply.

Do not normalize thresholds or channels merely to enable Go; raw values affect kick/log behavior. Follow the [account-age parity contract](79-account-age.md) for staging and rollback.
