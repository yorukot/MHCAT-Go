# Data Compatibility Plan

Status: Phase 1.5 Gate B review. User has approved schema changes when justified; this plan still defaults to legacy-compatible reads and rollback-compatible writes.

## Compatibility Goals

- The Go bot must read legacy Mongoose documents without requiring production data mutation.
- Go writes should remain readable by the Node.js bot during canary and rollback.
- Unknown fields must be preserved by avoiding full-document replacements.
- Missing fields must have documented zero/default behavior.
- Type mismatches must be audited before repair/backfill.
- Intentional schema changes require ADRs, fixtures, dual-read or compatibility adapter strategy, dry-run repair/backfill tooling when data must change, and rollback notes.

## Legacy Document Format

- No explicit collection names are set; expected names are Mongoose defaults.
- Most IDs are strings: guild, member/user, channel, role, message.
- XP and level fields are strings in `text_xp` and `voice_xp`.
- Many numeric fields are loose or can drift through JS writes.
- Several field names are misspelled and must be preserved in BSON tags unless changed intentionally:
  - `leavel`
  - `gift_chence`
  - `lock_anser`
  - `energi`
  - `suport`
- Loose object/array fields exist in poll, lottery, cron message payloads, warnings, lock channel users, and sign dates.

## Go BSON Struct Format

- Use exact `bson` tags for legacy fields.
- Use pointer fields where missing/null must be distinguished from zero.
- Use custom decode helpers or raw BSON audit for known mixed-type fields.
- Keep domain structs cleaner than BSON documents; convert at repository boundaries.
- Do not expose Mongo driver types outside adapters/repositories unless an ADR explicitly allows a narrow case.

## Compatibility Matrix

| Collection / field class | Legacy field | Go field | BSON tag | Type conversion | Missing field behavior | Unknown field behavior | Backward compatibility | Validation | Fixture |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Discord IDs | `guild`, `member`, `user`, `channel`, `role`, `message`, `messageid`, `channel_id` | string/domain ID wrappers | exact legacy tag | no numeric conversion by default | empty string invalid for writes; tolerated on read until audit | ignored/preserved | write strings | validate snowflake shape where possible | required |
| XP/levels | `xp`, `leavel` | compatibility numeric/string wrapper | `xp`, `leavel` | accept string and number; write legacy-compatible string until ADR changes | zero on missing, flagged by audit | preserved | Node expects string-like values | non-negative | required |
| Economy | `coin`, `today`, `coin_number`, `sign_coin`, `need_coin` | int64 wrappers | exact tags | accept number and numeric string; audit booleans in `today` | zero/false compatibility | preserved | write numbers unless legacy path proves string required | non-negative where applicable | required |
| Gacha/shop | `gift_chence`, `gift_count`, `give_coin`, `commodity_count` | int64/float wrappers | exact tags | accept number/numeric string | zero means unavailable unless legacy differs | preserved | preserve misspelled tag | validate bounds | required |
| Work | `energi`, `energy`, `time`, `end_time`, `state`, `get_coin` | int64/string wrappers | exact tags | accept number/numeric string | safe zero, audit active jobs | preserved | preserve misspelled `energi` | validate schedule and non-negative energy | required |
| Poll/lottery arrays | `choose_data`, `join_member`, `member`, `content` | raw/typed slices | exact tags | decode permissively first | empty slice on missing | preserved | avoid full replacement unless operation owns field | validate shape before mutation | required |
| Cron message | `message` | raw JSON/BSON payload DTO | `message` | decode as raw map initially | nil invalid for active job | preserved | Node-readable raw payload | validate before send | required |
| Config singletons | `guild` plus settings | document structs | exact tags | type-specific | missing means disabled/default | preserved | patch writes | validate on command input | required |

## Read Compatibility

- Decode legacy documents with missing fields.
- Tolerate unknown fields.
- Tolerate string/number drift in high-risk numeric fields.
- Detect duplicate singleton configs but do not auto-delete.
- Domain services should receive normalized values plus warnings/audit metadata where needed.

## Write Compatibility

- Use `$set`, `$inc`, `$setOnInsert`, `$addToSet`, `$pull`, and targeted array updates.
- Avoid replacing entire documents.
- Preserve legacy field names while Node rollback remains required.
- If a new canonical schema is adopted, dual-write only after ADR and tests prove Node rollback is no longer needed or a compatibility bridge exists.

## Rollback Compatibility

- Stop Go process.
- Disable Go command registration and feature flags.
- Restart Node.js bot with same MongoDB.
- Node must be able to read Go-written documents during canary.
- Any index applied by Go tooling must have rollback notes, especially unique indexes.

## Schema Change Policy

Schema changes are acceptable when they solve a real issue: correctness, security, performance, Discord API constraints, maintainability, or data consistency. Required before applying:

- ADR explaining old schema, new schema, and alternatives.
- Live data audit or representative fixture evidence.
- Compatibility tests showing legacy documents still decode.
- Dry-run repair/backfill plan if data mutation is needed.
- Explicit `--apply` operational step for data mutation.
- Rollback/restore guide.
- Feature-level behavior tests.

## Phase 1.5 Compatibility Decisions

Wave 1 will not change production document schemas. It may implement config parsing, Mongo connectivity, health checks, and read-only audit support only.

Feature waves may introduce schema changes only when they are:

- additive first;
- dual-read where old and new fields may coexist;
- rollback-compatible with Node.js while rollback is required;
- covered by live data audit or representative fixtures;
- documented in an ADR and this compatibility plan;
- paired with dry-run repair/backfill tooling if existing data must change.

Required changed-schema plan format:

| Collection | Legacy fields | New fields | Read strategy | Write strategy | Backfill needed | Rollback strategy | Dashboard impact |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Existing collections in Wave 1 | unchanged | none | legacy BSON tags only | no feature writes | no | stop Go; Node continues reading legacy docs | none |
| `mhcat_component_states` or equivalent future state collection | none | `state_id`, `feature`, `action`, `payload`, `owner_user_id`, `guild`, `expires_at`, `created_at` | used only when new versioned custom ID payload is too large/sensitive | insert with TTL after ADR; never required for legacy IDs | no legacy backfill | stop Go; old messages still use legacy IDs; drop collection only after TTL expiry | backup/export behavior must be decided before user-facing reliance |
| `mhcat_scheduler_locks` | none | `_id` equal to `lock_name`, `lock_name`, `owner`, `fence`, `expires_at`, `created_at`, `updated_at` | scheduler reads lease state only; expiry uses UTC instants | single-document acquire/renew/release with owner/fence checks; notification delivery uses `auto-notification-delivery`; daily-reset CLI/worker use `daily-reset`; release preserves documents and next acquire increments fence | no legacy backfill | stop Go workers, confirm relevant lease release/expiry, then restore Node cron ownership | no dashboard dependency expected |
| `loggings` | `guild`, `channel_id`, `message_update`, `message_delete`, `channel_update`, `member_voice_update` | none | config writes and message/channel/voice event emitters use the exact legacy fields by `guild` | `$set` legacy fields on all rows matching `guild`; `$setOnInsert` `guild` only when no rows match | no | disable Go logging-config and all three logging-event flags; Node.js can continue reading the same fields | no known dashboard dependency, but log channel privacy should be reviewed |
| `message_reactions` | `guild`, `message`, `react`, `role` | none | query normalized string `{guild,message,react}`; decode every selected field with Mongoose String scalar compatibility | setup updates every duplicate key's typed string `role`, upserts only when absent, and explicit delete removes every duplicate key | no; malformed non-string keys require audit rather than automatic coercion | disable the single Go role-selection gate; Node reads the same typed strings | dashboard backup exports plural and singular reaction collections; preserve both and follow the [role-selection parity contract](73-role-selection.md) |
| `btns` | `guild`, `number`, `role` | none | query normalized string `{guild,number}`; decode every selected field with Mongoose String scalar compatibility | setup updates every duplicate key's typed string `role` and upserts only when absent; there is no button-config delete flow | no; malformed non-string keys require audit rather than automatic coercion | disable the single Go role-selection gate; existing and Go-created panels remain Node-readable | dashboard backup includes `btns`; JavaScript decimal/exponent IDs remain strings; see the [role-selection parity contract](73-role-selection.md) |
| `cron_sets` | `guild`, `id`, `cron`, `channel`, `message` | none | list reads by `guild`; delivery scans rows with string `cron` and object `message`, then reloads `{guild,id}` before each send; direct and simplified setup write the exact legacy cron/message shape | setup inserts `{guild,id,channel,cron:null,message:null}` then direct or weekday/hour/minute completion sets `cron`/`message`; delete removes one `{guild,id}` row; list cleanup removes abandoned pending drafts; delivery does not mutate rows | no | disable Go config and delivery flags; stop Go and confirm lease release before restoring Node `handler/cron.js` | dashboard backup/export includes `cron_sets`; delivery and config gates remain independent |
| `lotters` | `guild`, `id`, `date`, `gift`, `howmanywinner`, `member`, `end`, `message_channel`, `yesrole`, `norole`, `maxNumber`, `owner` | none | permissive mixed string/number/null reads; malformed participant entries are skipped | entry uses one guarded update that appends `{id,time}` to `member`; stop/reroll set only `end:true`; no creation writes | no | disable Go lottery component flag; Node.js can continue reading the same participant/end fields | dashboard backup/export already includes `lotters`; creation remains disabled |
| `text_xp_channels` | `guild`, `channel`, `background`, `color`, `message` | none | read not required for the current config-only slice; future XP announcer must read exact legacy fields | `$set` `channel`, nullable `color`, nullable `message`, `$unset` `background` on all rows matching `guild`; `$setOnInsert` `guild` only when no rows match; delete removes all duplicate guild rows | no | disable Go text-XP config flag; Node.js can continue reading the same fields | no known dashboard dependency; Message Content/XP rollout remains separate |
| `voice_xp_channels` | `guild`, `channel`, `background`, `color`, `message` | none | read not required for the current config-only slice; future voice XP announcer must read exact legacy fields | `$set` `channel`, nullable `color`, nullable raw `message`, `$unset` `background` on all rows matching `guild` because legacy exposed but did not save `背景`; `$setOnInsert` `guild` only when no rows match; delete removes all duplicate guild rows | no | disable Go voice-XP config flag; Node.js can continue reading the same fields | no known dashboard dependency; Voice State/XP rollout remains separate |
| `join_roles` | String `guild`, String `role`, optional String `give_to_who` | none | Mongoose-compatible scalar DTO; missing/falsy audience defaults all users; malformed rows skip | typed `$setOnInsert`; explicit delete removes duplicate `{guild,role}` rows | no | stop assignment/config owners separately; Node reads typed fields | backup/export includes collection; confirm every writer; see [81-join-role.md](81-join-role.md) |
| Canonical numeric fields for XP/economy/work/gacha, if adopted later | legacy string/number fields such as `xp`, `leavel`, `coin`, `today`, `energi`, `gift_chence` | optional `*_num` or feature-specific canonical fields, exact names TBD by ADR | dual-read canonical first, fallback to legacy | dual-write or legacy-compatible write until rollback no longer needed | dry-run backfill likely | unset new fields or keep ignored by Node | dashboard impact must be reviewed for work/economy views |

## Work Payout Compatibility

`mhcat-work-payout` and the separately gated recurring worker share the same repository, configured lease, and rollback-compatible marker field on paid `coins` rows. Neither path requires a backfill or a new collection:

| Collection | Legacy fields | New fields | Read strategy | Write strategy | Backfill needed | Rollback strategy | Dashboard impact |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `work_users` | `guild`, `user`, `state`, `end_time`, `get_coin` | none | read due rows with non-idle `state` and effective `end_time < round(now_seconds)` guard; `_id` identifies duplicate rows independently | reset only the exact `_id`/guild/user/state/end-time/reward snapshot to `待業中` after payout | no | Node.js bot can continue reading the same fields | none expected |
| `coins` | `guild`, `member`, `coin`, `today` | `mhcat_work_payouts.<v1_work_row_hash>.token`, `.end_time` | resolve at most one `{guild,member}` row, then target stable `_id`; marker reads are Go-only | one aggregation-pipeline update conditionally increments `coin` and stores the latest per-work-row marker; missing rows use deterministic ObjectID | no; markers are lazy on first Go payout | stop Go and leave markers in place; Node ignores unknown fields; do not unset an in-flight marker | dashboard backup/export must preserve the additive object; no known UI dependency |
| `gift_changes` | `guild`, `time` | none | read config by `guild`; missing config uses daily marker | read-only | no | Node.js bot can continue reading the same fields | none expected |
| `gifts` | `guild`, `gift_name`, `gift_chence`, `gift_count` | none | read prizes by `guild`; decode `gift_chence` and `gift_count` from numeric or numeric-string legacy values | read-only | no | Node.js bot can continue reading the same fields | none expected |

Intentional fix: when `gift_changes.time == 0`, new Go-created coin rows use `today=1` for daily-reset mode instead of copying the legacy JavaScript truthiness bug that produced `today=now_seconds`.
| Anti-scam normalized URL/domain fields, if adopted later | `not_a_good_webs.web` | optional normalized domain/canonical URL field | match normalized field first, fallback to `web` | set canonical field while preserving `web` | dry-run normalization/backfill | unset canonical field; Node keeps `web` | reporting workflow must be confirmed |
| Dashboard-shared work settings | `work_somethings.guild/name/time/energy/coin/role` | none approved yet | preserve `{guild,name}` identity | patch writes only; no full replacement | no | Node/dashboard keep reading legacy fields | high: dashboard writes same collection |
| Dashboard-shared welcome settings | `join_messages.guild/enable/message_content/color/channel/img` | none approved yet | preserve exact fields and tolerate `img` missing/null | patch writes only; do not enforce `enable` uniqueness | no | Node/dashboard keep reading legacy fields | high: dashboard writes same collection |
| Leave-message config | `leave_messages.guild/message_content/title/color/channel` | none | preserve exact fields and tolerate missing/null optional embed fields | prepare slash command upserts missing guild row with null embed fields; modal submit uses one atomic `$set` for `message_content`, `title`, `color` | no | Node.js bot can continue reading the same fields | dashboard/shared ownership not confirmed; treat as shared until inspected |
| Dashboard-shared guild settings | `guilds.guild/announcement_id/voice_detection` | none approved yet | preserve exact fields | patch writes only | no | Node/dashboard keep reading legacy fields | high: dashboard writes voice detection |
| Announcement config slice | `guilds.announcement_id`, `ann_all_sets.guild/announcement_id/tag/color/title` | none | read/write exact legacy field names | patch writes only; duplicate-tolerant updates before upsert | no | Node bot can continue reading legacy config after rollback | dashboard backup/export keeps existing collections and fields |

## Open Questions

- Actual production collection names and indexes.
- Dashboard/shared-worker access to Mongo collections.
- Which singleton configs have duplicates in production.
- Whether `chatgpt` documents are consumed outside this repo.
- Whether disabled birthday/lottery behavior should remain disabled.

## Phase 1.5 Open Items

- Live audit has not run because no Mongo URI/database was available.
- Dashboard is locally confirmed and shared; production deployment status and DB name still need manual confirmation.
- ChatGPT worker is inferred but not found locally; preserve `chatgpts`/`chatgpt_gets` handoff until confirmed retired.
- Dashboard backup expects singular `message_reaction`; live audit must confirm existence and shape.

## Wave 3 Compatibility Note

Wave 3 does not change any production schema and does not implement feature writes. It adds:

- read-only audit report structures;
- partial catalog metadata;
- index diff planning;
- atomic update builder helpers for future repositories;
- transaction runner shell for future multi-document features.

Compatibility impact:

- Missing fields are reported, not backfilled.
- Mixed types are reported from samples, not normalized.
- Unknown collections are reported, not deleted.
- Unique indexes are blocked unless duplicate audit is clean.
- TTL indexes are blocked unless a retention decision exists.
- Node.js rollback compatibility is unchanged because Wave 3 performs no feature data writes by default.

## Platform Wave B Compatibility Note

Platform Wave B corrects the in-code collection catalog to cover all 47 legacy Mongoose models and their expected Mongo collection names. It does not change production schemas and does not implement feature writes.

Compatibility impact:

- Go audit/index tooling now compares live Mongo collections against the full legacy model catalog instead of the earlier high-risk subset.
- Collection names are explicit and contract-tested before repository implementation.
- Field-level BSON strictness is still incomplete; feature waves must add document structs, legacy fixtures, and mixed-type decode tests before writes.
- Node.js rollback compatibility is unchanged because Platform Wave B performs no feature data writes, repairs, or index creation by default.

## Ticket Config Compatibility

Ticket config compatibility is parity-audited through the gated setup/delete/open/close runtime.

| Collection | Legacy fields | New fields | Read strategy | Write strategy | Backfill needed | Rollback strategy | Dashboard impact |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `tickets` | `guild`, `ticket_channel`, `admin_id`, `everyone_id` | none beyond ordinary Mongo `_id` | Mongoose-compatible String scalar decoding; malformed compound/null values stay unusable | create-if-absent `$setOnInsert`; exact `{_id,guild}` failure rollback; explicit guild delete removes duplicates | no | stop all Go ticket routes before restoring Node; typed strings and legacy `tic`/`del` IDs remain readable | none known; confirm before rollout |

Go writes only after valid versioned modal input, never overwrites an existing guild row, and creates no startup index. Candidate `tickets_guild` remains blocked on duplicate/malformed-value audit and exclusive ownership. See the [ticket parity contract](74-ticket.md) for write ordering, compensation, migration, and rollback.

## Poll Compatibility

Poll compatibility is parity-audited through the gated create/vote/result/owner runtime.

| Collection | Legacy fields | New fields | Read strategy | Write strategy | Backfill needed | Rollback strategy | Dashboard impact |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `polls` | `guild`, `messageid`, `question`, `create_member_id`, `many_choose`, `can_change_choose`, `can_see_result`, `end`, `anonymous`, `choose_data`, `join_member` | none | separate Mongoose-compatible scalar/loose-array read DTO; malformed Mixed entries are skipped | typed insert plus targeted conditional `$push`/`$pull` and atomic toggle pipelines; no full replacement | no | stop all Go poll routes before restoring Node; typed rows are readable, but Go-rendered versioned component IDs require a separate message rollback plan | dashboard backup already exports `polls`; preserve collection and fields |

Raw question/choice whitespace and misspelled `join_member[].choise` remain exact. Candidate `polls_guild_message` remains blocked on duplicate/key/scalar/array audit and exclusive ownership; no startup or TTL index is created. See the [poll parity contract](75-poll.md) for component migration, atomic predicates, staging, and rollback.

## Announcement Compatibility

Announcement compatibility is parity-audited across config, one-time send, confirmation, and bound relay routes.

| Collection | Legacy fields | New fields | Read strategy | Write strategy | Backfill needed | Rollback strategy | Dashboard impact |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `guilds` | `guild`, `announcement_id`, shared `voice_detection` and unknown dashboard fields | none | separate Mongoose-compatible String scalar DTO; compound values remain unusable | typed `$set` patch of `announcement_id` across duplicate matches, preserving unrelated fields | no | stop Go config/send ownership before restoring Node; typed strings remain readable | shared collection: confirm every dashboard writer and preserve unknown fields |
| `ann_all_sets` | `guild`, `announcement_id`, `tag`, `color`, `title` | none | separate Mongoose-compatible String scalar DTO; raw whitespace/case remains exact | typed updates/deletes across duplicate exact matches, upsert only when absent | no | stop Go config/relay ownership before restoring Node; typed rows remain readable | confirm external writers before rollout |

No startup index, repair, deduplication, or backfill is authorized. Candidate unique indexes remain blocked on duplicate keys, scalar drift, malformed values, shared writers, and exclusive ownership. Process-local confirmation state has no Mongo migration; wait six seconds during rollback. See the [announcement parity contract](76-announcement.md).

## Anti-Scam Compatibility

Anti-scam compatibility is parity-audited across config, report, and message deletion.

| Collection | Legacy fields | New fields | Read strategy | Write strategy | Backfill needed | Rollback strategy | Dashboard impact |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `good_webs` | `guild`, `open` | none | exact guild lookup plus Mongoose-compatible Boolean scalar decode | typed `$set.open` across duplicate matches; upsert only when absent | no | stop Go config/deletion ownership before restoring Node; typed values remain readable | confirm external config writers |
| `not_a_good_webs` | `web` | none | Mongoose-compatible String scalar decode; raw values preserved; compounds skipped | read-only | no | stop Go report/deletion ownership before restoring Node; preserve catalog exactly | confirm catalog ingestion/curation owner |

No URL normalization, canonical field, repair, deduplication, or startup index is approved. Candidate indexes remain blocked on key/type/raw-variant audits and exclusive ownership. See the [anti-scam parity contract](77-anti-scam.md).

## Birthday Compatibility

Birthday command compatibility is parity-audited; the commented delivery block remains inactive.

| Collection | Legacy fields | New fields | Read strategy | Write strategy | Backfill needed | Rollback strategy | Dashboard impact |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `birthday_sets` | `guild`, `msg`, `utc`, `channel`, `everyone_can_set_birthday_date`, `role` | none | exact key lookup plus Mongoose-compatible scalar DTO | typed patches across duplicate matches; upsert only when absent | no | stop Go command ownership, preserve rows, restore Node command only | backup/export already includes collection; confirm writers |
| `birthdays` | `guild`, `user`, birthday/time Number fields, `allow` | none | exact profile lookup and broad list with permissive scalar DTO; null numbers remain absent | typed patches across duplicate matches; delete removes duplicate matches | no | stop Go command ownership; typed values remain Mongoose-readable | backup/export already includes collection; confirm writers |

No repair, deduplication, backfill, scheduler restoration, or startup index is authorized. Candidate indexes remain blocked on duplicate/key/type/value/ownership findings. See the [birthday parity contract](78-birthday.md).

## Account-Age Compatibility

Account-age config and member policy are parity-audited as separate ownership families.

| Collection | Legacy fields | New fields | Read strategy | Write strategy | Backfill needed | Rollback strategy | Dashboard impact |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `create_hours` | `guild`, String `hours`, nullable String `channel` | none | exact guild lookup, Mongoose-compatible scalar strings, JavaScript Number threshold parsing | typed String/null one-row patches; full config delete removes duplicate guild matches | no | stop policy/config owners separately; typed rows remain Mongoose-readable | confirm backup/export and all external writers |

No normalization, repair, deduplication, backfill, or startup index is approved. Candidate `{guild:1}` remains blocked on key/type/value/writer/ownership audits. See the [account-age parity contract](79-account-age.md).

## Verification Compatibility

Verification setup and captcha completion are parity-audited as independent ownership families.

| Collection | Legacy fields | New fields | Read strategy | Write strategy | Backfill needed | Rollback strategy | Dashboard impact |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `verifications` | String `guild`, String `role`, nullable String `name` | none | exact guild lookup plus Mongoose-compatible scalar String DTO; unusable required role fails safely | typed String/null patches across duplicate matches; upsert only when absent | no | stop flow/config owners separately; typed rows remain Mongoose-readable | backup/export includes collection; confirm every external writer |

No normalization, repair, deduplication, backfill, challenge-state migration, or startup index is approved. Candidate `{guild:1}` remains blocked on key/type/value/writer/ownership findings. Process-local challenge state expires naturally and is not migrated. See the [verification parity contract](80-verification.md).
