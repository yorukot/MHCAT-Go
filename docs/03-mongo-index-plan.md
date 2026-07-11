# Mongo Index Plan

Status: Platform Wave B. The default Go index plan is now derived from the full 47-model collection catalog rather than the earlier partial singular scaffold. Duplicate/null/missing data must still be audited before any unique index apply.

## Strategy

- Index bootstrap must be idempotent.
- Do not create SQL-style migrations.
- Bot startup must not create high-risk production indexes by default.
- `mhcat-tools mongo check-indexes` and `mhcat-tools mongo ensure-indexes --dry-run` are the default operational path.
- `mhcat-tools mongo ensure-indexes --apply` is required for write operations.
- Unique indexes require duplicate/null/missing audits first.
- Large indexes require explicit maintenance window guidance and rollback notes.
- Index changes that support an intentional schema change need ADR coverage.

## Existing Indexes

- No explicit `.index()` or `schema.index()` declarations were found.
- `join_message.guild` and `join_message.enable` declare `unique: true`, but `autoIndex: false` means production creation is not guaranteed.
- `join_message.enable` unique is likely invalid because it is Boolean and would allow only one true/false document if enforced.
- Production read-only audit on 2026-07-04 found most collections have only `_id_`; `text_xps` and `voice_xps` already have non-unique `{guild:1, member:1}` indexes named `guild_1_member_1_autocreated`.
- See `docs/26-production-mongo-readonly-audit.md` for sanitized production metadata.

## Recommended Indexes

| Collection | Index keys | Unique | Sparse / partial | TTL | Reason | Legacy query evidence | Risk | Bootstrap behavior | Rollback note |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `coins` | `{guild:1, member:1}` | candidate after audit | no | no | balance lookup/update | economy commands, sign, games, work payout | duplicates may exist; work payout rejects an ambiguous logical balance | dry-run duplicate audit first | drop index only after traffic drained |
| `coins` | `{guild:1, coin:-1}` | no | no | no | coin ranking | coin rank command | large collection sort | explicit apply | safe to drop if rank slows |
| `text_xps` | `{guild:1, member:1}` | candidate after audit | no | no | text XP lookup/update | text XP event and commands | duplicates/type drift | dry-run duplicate audit first | drop after disabling Go writes |
| `text_xps` | `{guild:1}` | no | no | no | rank list | rank command loads guild rows | large scan remains possible | apply after audit | safe to drop |
| `voice_xps` | `{guild:1, member:1}` | candidate after audit | no | no | voice XP lookup/update | voice state event | duplicate active sessions | dry-run duplicate audit first | drop after disabling Go writes |
| `voice_xps` | `{guild:1}` | no | no | no | voice rank list | voice rank command | large scan remains possible | apply after audit | safe to drop |
| `work_users` | `{guild:1, user:1}` | candidate after audit | no | no | work state lookup/update | work interface/job completion | duplicate rows are independently marked/paid but may represent bad legacy state | audit first; required before considering unique | drop after feature disabled |
| `work_users` | `{state:1, end_time:1}` | no | partial possible for active jobs | no | due job scan | minute work completion loop and Go one-shot/recurring payout | type drift in `end_time`; broad scan until index exists | explicit apply only after scheduler ownership review | safe to drop |
| `cron_sets` | `{guild:1, id:1}` | candidate after audit | no | no | schedule lookup/delete | cron list/delete | duplicate schedules | audit first; preserve one-row semantics in [92-auto-notification.md](92-auto-notification.md) | drop after scheduler disabled |
| `polls` | `{guild:1, messageid:1}` | unique candidate after audit; non-unique lookup safe now | no | no | create/component vote/result/owner lookup | poll command and components | duplicate/malformed keys and loose scalar/array shapes | apply non-unique lookup now; require exclusive ownership and clean audit before unique; see [75-poll.md](75-poll.md) | drop only after poll traffic is drained |
| `message_reactions` | `{guild:1, message:1, react:1}` | candidate after audit | no | no | reaction-role lookup | reaction add/remove | emoji normalization | audit first | safe to drop |
| `lock_channels` | `{guild:1, channel_id:1}` | candidate after audit | no | no | voice lock lookup | lock modal/voice state | plaintext lock data | audit first | safe to drop |
| `voice_channels` | `{guild:1, ticket_channel:1}` and `{guild:1, parent:1}` | trigger candidate after audit; parent lookup safe now | no | no | trigger lookup and category cleanup | voice state event / config delete | name mismatch | audit unique trigger; parent lookup needs no duplicate cleanup | safe to drop |
| `voice_channel_ids` | `{guild:1, channel_id:1}` | candidate after audit | no | no | dynamic channel cleanup | voice room deletion | stale channels | audit first | safe to drop |
| `ghps` | `{guild:1, commodity_id:1}` | unique candidate after audit; non-unique lookup safe now | no | no | shop item lookup and guild list | transactional shop purchase/admin | duplicate commodity IDs | apply non-unique lookup now; audit before unique | safe to drop |
| `gifts` | `{guild:1, gift_name:1}` | unique candidate after audit; non-unique lookup safe now | no | no | gacha prize pool/lookup/edit/delete | transactional gacha draw and admin writes | duplicate names | apply non-unique lookup now; audit before unique | safe to drop |
| `gift_changes` | `{guild:1}` | unique candidate after audit; non-unique lookup safe now | no | no | economy config singleton | sign/gacha config | duplicates | apply non-unique lookup now; audit before unique | safe to drop |
| `gift_changes` | `{time:1, guild:1}` | no | partial for `time != 0` if useful | no | legacy daily reset exclusion; current Go reset scans and normalizes loose `time` values | Node `distinct('guild', {time: {$ne: 0}})`; Go full projected scan | partial semantics do not cover Go normalization | dry-run only until data known | safe to drop |
| singleton config collections | `{guild:1}` | candidate after audit | no | no | guild config lookup | many `findOne({guild})` paths | duplicates common | dry-run duplicate audit first | drop if config resolver handles duplicates |
| `chats`, `chatgpts`, `chatgpt_gets`, `create_hours`, `good_webs`, `guilds`, `join_messages`, `leave_messages`, `loggings`, `tickets`, `verifications`, `work_sets` | `{guild:1}` | candidate after audit | no | no | guild singleton lookup | command/event config reads | duplicates/unknown live names | dry-run duplicate audit first | safe to drop |
| `chat_roles`, `voice_roles` | `{guild:1, leavel:1, role:1}` | candidate after audit | no | no | level role lookup | XP role assignment | misspelled numeric string | audit first | safe to drop |
| `birthdays` | `{guild:1, user:1}` | unique candidate after audit; non-unique lookup safe now | no | no | user birthday lookup and guild list | birthday commands; scheduler remains inactive | duplicate profiles | apply non-unique lookup now; audit before unique | safe to drop |
| `birthday_sets` | `{guild:1}` | unique candidate after audit; non-unique lookup safe now | no | no | birthday config lookup | birthday commands | singleton duplicates | apply non-unique lookup now; audit before unique | safe to drop |
| `warndbs` | `{guild:1, user:1}` | candidate after audit | no | no | warnings lookup | warn/warnings commands | array shape | audit first | safe to drop |
| `all_use_counts` | `{slashcommand_name:1}` | candidate after audit | no | no | usage counter lookup | slash dispatcher | undefined names exist | repair before unique | safe to drop |
| `codes` | `{code:1}` | blocked for parity | no | no | redeem code lookup | `兌換` command | duplicates are observable first-match state | audit only; do not create for redeem | safe to drop if previously experimental and unused |
| `not_a_good_webs` | `{web:1}` | candidate after normalization decision | no | no | scam URL lookup | report/safe server | regex/domain normalization | audit and normalize first | safe to drop |

## Phase 1.5 Required Index Plan Table

A live read-only index inventory has now run, but full duplicate/null/missing audits remain required before unique indexes. The table below is the Gate B plan format. `candidate` means the index is behaviorally desirable but must not be unique until live duplicate/missing/null audit passes.

| Collection | Index name | Keys | Unique | Sparse/Partial | TTL | Query supported | Duplicate risk | Build strategy |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `numbers` | `numbers_guild` | `{guild:1}` | candidate | no | no | guild stats config singleton | singleton duplicates and scalar/whitespace keys | audit only, then reviewed dry-run/explicit apply; preserve [93-stats.md](93-stats.md) duplicate behavior |
| `all_use_counts` | `all_use_counts_command` | `{slashcommand_name:1}` | candidate | no | no | slash usage counter | null/undefined command names | audit invalid names before unique |
| `ann_all_sets` | `ann_all_sets_guild_announcement` | `{guild:1,announcement_id:1}` | candidate | no | no | announcement config/relay lookup | duplicate or scalar-drift keys, malformed values, shared Node writer | audit types/duplicates and exclusive ownership; no startup apply; see [contract](76-announcement.md) |
| `ann_all_sets` | `ann_all_sets_guild_announcement_lookup` | `{guild:1,announcement_id:1}` | no | no | no | cached per-message relay lookup | non-unique index is duplicate-safe; redundant after unique approval | explicit apply; remove before promoting the same-key unique index |
| `birthdays` | `birthdays_guild_user` | `{guild:1,user:1}` | candidate | no | no | birthday profile lookup | duplicate/missing/scalar-drift keys, malformed profile values, ownership | audit only; explicit reviewed apply after clean findings; see [contract](78-birthday.md) |
| `birthdays` | `birthdays_guild_user_lookup` | `{guild:1,user:1}` | no | no | no | birthday profile lookup and guild list | non-unique index is duplicate-safe | explicit apply; remove before promoting the same-key unique index |
| `birthday_sets` | `birthday_sets_guild` | `{guild:1}` | candidate | no | no | birthday guild config | duplicate/missing/scalar-drift keys, external writers | audit only; explicit reviewed apply after clean findings; see [contract](78-birthday.md) |
| `birthday_sets` | `birthday_sets_guild_lookup` | `{guild:1}` | no | no | no | birthday config lookup | non-unique index is duplicate-safe | explicit apply; remove before promoting the same-key unique index |
| `btns` | `btns_guild_number` | `{guild:1,number:1}` | candidate | no | no | role button lookup | duplicate button IDs | audit, dry-run, explicit apply |
| `chats` | `chats_guild` | `{guild:1}` | owner-wide candidate only | no | no | autochat config/runtime first match | duplicates are observable and all config/runtime/dashboard writers must agree | no apply solely for config; audit under [89-autochat-config.md](89-autochat-config.md) |
| `chat_roles` | `chat_roles_guild_level_role` | `{guild:1,leavel:1,role:1}` | candidate | no | no | text level role lookup | misspelled level/type drift | audit, dry-run, explicit apply |
| `chatgpts` | `chatgpts_guild` | `{guild:1}` | owner-wide candidate only | no | no | paid worker handoff singleton | worker compatibility, duplicate rows, and all writers remain rollout blockers | no startup apply; audit worker/duplicates and follow [91-autochat-paid.md](91-autochat-paid.md) |
| `chatgpt_gets` | `chatgpt_gets_guild` | `{guild:1}` | owner-wide candidate only | no | no | balance/redeem/autochat shared guild lookup | owner-specific duplicate behavior; redeem preserves duplicates while paid fails closed | no startup apply; reconcile [87](87-balance-query.md), [88](88-redeem.md), [90](90-autochat-fallback.md), and [91](91-autochat-paid.md) before any index |
| `codes` | `codes_code` | `{code:1}` | blocked | no | no | redeem code lookup | changes duplicate behavior and can reject existing rows | do not apply under [88-redeem.md](88-redeem.md) |
| `coins` | `coins_guild_member` | `{guild:1,member:1}` | candidate | no | no | balance lookup/update and work-payout target resolution | duplicate balances; payout fails closed until repaired | audit before unique; explicit apply |
| `coins` | `coins_guild_coin_rank` | `{guild:1,coin:-1}` | no | no | no | coin ranking | none for non-unique | explicit apply after live count review |
| `create_hours` | `create_hours_guild` | `{guild:1}` | candidate | no | no | account-age config/policy lookup | duplicate/missing/scalar-drift keys, malformed values, external writers, ownership | audit only; reviewed explicit apply after clean findings; no startup apply; see [contract](79-account-age.md) |
| `cron_sets` | `cron_sets_guild_id` | `{guild:1,id:1}` | candidate | no | no | schedule list/delete/send | duplicate schedule IDs | audit, dry-run, explicit apply; keep non-unique per [92-auto-notification.md](92-auto-notification.md) |
| `errors_sets` | `errors_sets_guild` | `{guild:1}` | candidate | no | no | warning escalation config | singleton duplicates | audit, dry-run, explicit apply |
| `ghps` | `ghps_guild_commodity_id` | `{guild:1,commodity_id:1}` | candidate | no | no | shop item lookup | duplicate commodity IDs | audit, dry-run, explicit apply |
| `ghps` | `ghps_guild_commodity_id_lookup` | `{guild:1,commodity_id:1}` | no | no | no | shop item lookup and guild list | non-unique index is duplicate-safe | explicit apply; remove before promoting the same-key unique index |
| `gifts` | `gifts_guild_gift_name` | `{guild:1,gift_name:1}` | candidate | no | no | gacha prize lookup | duplicate names | audit, dry-run, explicit apply |
| `gifts` | `gifts_guild_gift_name_lookup` | `{guild:1,gift_name:1}` | no | no | no | gacha pool prefix scan and prize inventory lookup | non-unique index is duplicate-safe | explicit apply; remove before promoting the same-key unique index |
| `gift_changes` | `gift_changes_guild` | `{guild:1}` | candidate | no | no | economy/gacha config | singleton duplicates | audit, dry-run, explicit apply |
| `gift_changes` | `gift_changes_guild_lookup` | `{guild:1}` | no | no | no | economy/gacha hot-path config lookup | non-unique index is duplicate-safe | explicit apply; remove before promoting the same-key unique index |
| `gift_changes` | `gift_changes_time_guild` | `{time:1,guild:1}` | no | optional partial `time != 0` only after ADR | no | legacy reset exclusion; not required by current Go normalized scan | partial semantics risk and no current Go query benefit | dry-run only until data known |
| `good_webs` | `good_webs_guild` | `{guild:1}` | candidate | no | no | anti-scam toggle/deletion gate | singleton duplicates, key/type drift, concurrent writers | audit and exclusive ownership; no startup apply; see [contract](77-anti-scam.md) |
| `guilds` | `guilds_guild` | `{guild:1}` | candidate | no | no | announcement channel config and dashboard voice detection | scalar-drift keys, singleton duplicates, dashboard/shared writers | audit with dashboard compatibility and announcement ownership review; no startup apply; see [contract](76-announcement.md) |
| `join_messages` | `join_messages_guild` | `{guild:1}` | candidate | no | no | welcome message config/dashboard writes | singleton duplicates; unsafe `enable` unique | do not create `enable` unique; audit first |
| `join_messages` | `join_messages_guild_lookup` | `{guild:1}` | no | no | no | member-join welcome lookup | non-unique index is duplicate-safe | explicit apply; remove before promoting the same-key unique index |
| `join_roles` | `join_roles_guild_role` | `{guild:1,role:1}` | candidate | no | no | join-role config and member-add assignment | duplicate/missing/null/blank/scalar-drift keys, malformed audiences, stale roles, external writers | require exclusive ownership and clean audit, then dry-run/explicit apply only; never startup-create; see [81-join-role.md](81-join-role.md) |
| `join_roles` | `join_roles_guild_lookup` | `{guild:1}` | no | no | no | member-join role list lookup | non-unique index is duplicate-safe | explicit apply |
| `leave_messages` | `leave_messages_guild` | `{guild:1}` | candidate | no | no | leave message config | singleton duplicates | audit, dry-run, explicit apply |
| `leave_messages` | `leave_messages_guild_lookup` | `{guild:1}` | no | no | no | member-leave message lookup | non-unique index is duplicate-safe | explicit apply; remove before promoting the same-key unique index |
| `lock_channels` | `lock_channels_guild_channel` | `{guild:1,channel_id:1}` | candidate | no | no | voice lock lookup | stale/duplicate channel rows | audit, dry-run, explicit apply |
| `loggings` | `loggings_guild` | `{guild:1}` | candidate | no | no | logging config | singleton duplicates | audit, dry-run, explicit apply |
| `loggings` | `loggings_guild_lookup` | `{guild:1}` | no | no | no | cached logging event lookup | non-unique index is duplicate-safe | explicit apply; remove before promoting the same-key unique index |
| `lotters` | `lotters_guild_id` | `{guild:1,id:1}` | candidate | no | no | lottery lookup/member updates | duplicate lottery IDs | audit, dry-run, explicit apply |
| `message_reactions` | `message_reactions_guild_message_react` | `{guild:1,message:1,react:1}` | candidate | no | no | reaction role lookup | emoji normalization/duplicates | audit, dry-run, explicit apply |
| `message_reactions` | `message_reactions_guild_message_react_lookup` | `{guild:1,message:1,react:1}` | no | no | no | reaction event lookup | non-unique index is duplicate-safe | explicit apply; remove before promoting the same-key unique index |
| `message_reaction` | `message_reaction_audit_only` | `{guild:1}` | no | no | no | dashboard backup lists singular collection | unknown live shape | audit live existence before any writes |
| `not_a_good_webs` | `not_a_good_webs_web` | `{web:1}` | candidate after explicit ADR | no | no | anti-scam report/delete lookup | duplicate/raw variants, scalar drift, external writers | audit only; no normalization/startup apply; see [contract](77-anti-scam.md) |
| `polls` | `polls_guild_messageid` | `{guild:1,messageid:1}` | candidate | no | no | poll create/vote/result/owner lookup | duplicate keys, malformed key/scalar/array shapes, ambiguous rollback | require exclusive ownership plus clean duplicate/type/shape audit, dry-run, and explicit apply; never create at startup; see [poll parity contract](75-poll.md) |
| `polls` | `polls_guild_messageid_lookup` | `{guild:1,messageid:1}` | no | no | no | poll command and component hot-path lookup | non-unique index is duplicate-safe | explicit apply; remove before promoting the same-key unique index |
| `role_numbers` | `role_numbers_guild_role` | `{guild:1,role:1}` | candidate | no | no | role stats config | duplicate/scalar/whitespace role configs | audit only, then reviewed dry-run/explicit apply; see [93-stats.md](93-stats.md) |
| `role_numbers` | `role_numbers_guild_channel` | `{guild:1,channel:1}` | candidate | no | no | stats channel lookup | duplicate/scalar/whitespace channel configs | audit only, then reviewed dry-run/explicit apply; see [93-stats.md](93-stats.md) |
| `sign_lists` | `sign_lists_guild_member` | `{guild:1,member:1}` | candidate | no | no | sign-in history lookup | duplicate user sign docs | audit, dry-run, explicit apply |
| `suports` | `suports_support_id` | `{support_id:1}` | candidate | no | no | support lookup | duplicate support IDs | audit, dry-run, explicit apply |
| `systems` | `systems_none` | none | no | no | no | unclear active use | unknown | no index until usage confirmed |
| `text_xps` | `text_xps_guild_member` | `{guild:1,member:1}` | candidate | no | no | text XP lookup/update | duplicate XP docs | audit before unique; explicit apply |
| `text_xps` | `text_xps_guild_rank` | `{guild:1,leavel:-1,xp:-1}` | no | no | no | text XP rank | mixed string/number sort risk | apply only after type audit |
| `text_xp_channels` | `text_xp_channels_guild` | `{guild:1}` | candidate | no | no | text XP announcement config and `聊天經驗設定`/`聊天經驗刪除` lookup | singleton duplicates | audit, dry-run, explicit apply; no startup index creation |
| `tickets` | `tickets_guild` | `{guild:1}` | candidate | no | no | ticket panel config | singleton duplicates and malformed scalar IDs | require exclusive ownership plus clean duplicate/type audit, dry-run, and explicit apply; never create at startup; see [ticket parity contract](74-ticket.md) |
| `verifications` | `verifications_guild` | `{guild:1}` | candidate | no | no | verification config | duplicate/missing/null/blank/scalar-drift guilds, malformed role/name values, external writers | require exclusive ownership and clean audit, then dry-run and explicit apply; never startup-create; see [80-verification.md](80-verification.md) |
| `voice_channels` | `voice_channels_guild_ticket_channel` | `{guild:1,ticket_channel:1}` | candidate | no | no | dynamic voice trigger config | duplicate trigger channels | audit, dry-run, explicit apply |
| `voice_channels` | `voice_channels_guild_parent_lookup` | `{guild:1,parent:1}` | ready | no | no | dynamic voice category cleanup | none | dry-run, explicit apply |
| `voice_channel_ids` | `voice_channel_ids_guild_channel` | `{guild:1,channel_id:1}` | candidate | no | no | dynamic voice channel state | stale/duplicate state | audit, dry-run, explicit apply |
| `voice_roles` | `voice_roles_guild_level_role` | `{guild:1,leavel:1,role:1}` | candidate | no | no | voice level role lookup | misspelled level/type drift | audit, dry-run, explicit apply |
| `voice_xps` | `voice_xps_guild_member` | `{guild:1,member:1}` | candidate | no | no | voice XP lookup/update | duplicate active state | audit before unique; explicit apply |
| `voice_xps` | `voice_xps_guild_rank` | `{guild:1,leavel:-1,xp:-1}` | no | no | no | voice XP rank | mixed string/number sort risk | apply only after type audit |
| `voice_xp_channels` | `voice_xp_channels_guild` | `{guild:1}` | candidate | no | no | voice XP announcement config lookup/write by `語音經驗設定` / `語音經驗刪除` | singleton duplicates | duplicate audit first; dry-run and explicit apply only; app does not create it on startup |
| `votes` | `votes_guild_number` | `{guild:1,Number:1}` | candidate | no | no | vote state/config lookup | capitalized field drift | audit, dry-run, explicit apply |
| `votes` | `votes_guild_member` | `{guild:1,member:1}` | candidate | no | no | member vote lookup | duplicate user votes | audit, dry-run, explicit apply |
| `warndbs` | `warndbs_guild_user` | `{guild:1,user:1}` | no by default | no | no | warnings lookup/dashboard reads | multiple warning docs may be valid | non-unique explicit apply after behavior review |
| `work_sets` | `work_sets_guild` | `{guild:1}` | candidate | no | no | work config | singleton duplicates | audit, dry-run, explicit apply |
| `work_somethings` | `work_somethings_guild_name` | `{guild:1,name:1}` | candidate | no | no | dashboard/bot work job lookup | duplicate names per guild | do not create dashboard `guild` unique; audit first |
| `work_users` | `work_users_guild_user` | `{guild:1,user:1}` | candidate | no | no | user work state lookup | duplicate jobs/payments | audit before unique; explicit apply |
| `work_users` | `work_users_guild_energi` | `{guild:1,energi:1}` | no | no | no | daily refill/clamp candidate scan | mixed energy scalar types can reduce range selectivity but do not block creation | explicit apply |
| `work_users` | `work_users_state_end_time` | `{state:1,end_time:1}` | no | optional active-job partial after audit | no | due job scan for one-shot and recurring work payout | type drift in `end_time`; changed partial semantics risk | explicit apply only after scheduler ownership review |

## Bootstrap Behavior

- Startup default: no index writes.
- Development startup may support low-risk `--ensure-indexes-dev` after tests exist.
- Production index apply must go through `mhcat-tools`.
- Unique candidates are disabled until audits prove no duplicates/nulls/missing keys.
- The tool should print collection, key, uniqueness, existing state, affected count estimate if available, and risk classification.

## Tests

- Plan serialization golden test.
- Dry-run output tests.
- Duplicate audit tests for every unique candidate.
- Existing-index comparison tests.
- Context cancellation tests.
- Error mapping tests for duplicate key and command timeout.

## Wave 3 Index Diff Shell

Wave 3 added:

- `cmd/mhcat-mongo-index`
- `internal/adapters/mongo/indexes.go`
- `internal/adapters/mongo/index_diff.go`

Default behavior:

- dry-run only;
- list live collections and indexes;
- compare live indexes with the local plan;
- output deterministic text or JSON;
- never drop indexes;
- never modify existing indexes;
- never create indexes unless `--apply` is explicit.

Apply guardrails:

- safe missing non-unique/non-TTL indexes may be created only in `--apply` mode;
- unique index creation additionally requires `--allow-unique` and a clean duplicate audit;
- TTL index creation additionally requires `--allow-ttl` and a retention ADR/note in the plan;
- changed indexes are marked dangerous and are not recreated in Wave 3;
- unknown remote indexes are reported as `unknown_remote` and never dropped.

## Platform Wave B Contract Update

- `DefaultIndexPlan(DefaultCollectionCatalog())` now uses all corrected Mongoose collection names.
- Tests block regressions to old singular scaffold names such as `coin`, `text_xp`, `voice_xp`, `poll`, `ticket`, `guild`, `cron_set`, `verification`, and `chatgpt`.
- Unique candidates remain guarded by `RequiresDuplicateAudit`.
- Non-unique candidate indexes may appear as safe create operations only when `mhcat-mongo-index --apply` is explicitly used; no apply was run in Platform Wave B.

## Go System Collection Index Notes

`mhcat_scheduler_locks` uses `_id == lock_name`, so MongoDB's default `_id` index is the lease identity index. Do not add a TTL index to `expires_at`; expiry is lease query semantics, not data retention. Any future auxiliary index must be added through the same dry-run index process and ADR review.
