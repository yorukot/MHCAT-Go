# Mongo Collection Catalog

Status: Platform Wave B. The Go typed catalog now covers all 47 legacy Mongoose model files with explicit expected collection names, model names, legacy file paths, logical keys, planned indexes, and compatibility notes. Live database audit is still required before production writes or unique index apply.

## Principles

- Preserve legacy collection names and BSON field names by default.
- Schema changes are allowed when justified, but must be documented in ADRs with compatibility, audit, repair/backfill, and rollback notes.
- Do not create SQL-style migrations, migration runners, or version tables.
- Use collection catalog, BSON compatibility, index bootstrap, data audit, fixture snapshots, and dry-run repair/backfill tooling.
- Unknown fields must remain tolerated unless a repository explicitly uses projection.
- Missing fields must decode safely according to documented legacy behavior.
- Prefer patch-style writes over full document replacement so Node.js rollback remains possible.

## Model Files Detected

47 files:

- `models/Number.js`
- `models/all_use_count.js`
- `models/ann_all_set.js`
- `models/birthday.js`
- `models/birthday_set.js`
- `models/btn.js`
- `models/chat.js`
- `models/chat_role.js`
- `models/chatgpt.js`
- `models/chatgpt_get.js`
- `models/code.js`
- `models/coin.js`
- `models/create_hours.js`
- `models/cron_set.js`
- `models/errors_set.js`
- `models/ghp.js`
- `models/gift.js`
- `models/gift_change.js`
- `models/good_web.js`
- `models/guild.js`
- `models/join_message.js`
- `models/join_role.js`
- `models/leave_message.js`
- `models/lock_channel.js`
- `models/logging.js`
- `models/lotter.js`
- `models/message_reaction.js`
- `models/not_a_good_web.js`
- `models/poll.js`
- `models/role.js`
- `models/sign_list.js`
- `models/suport.js`
- `models/system.js`
- `models/text_xp.js`
- `models/text_xp_channel.js`
- `models/ticket.js`
- `models/verification.js`
- `models/voice_channel.js`
- `models/voice_channel_id.js`
- `models/voice_role.js`
- `models/voice_xp.js`
- `models/voice_xp_channel.js`
- `models/vote.js`
- `models/warndb.js`
- `models/work_set.js`
- `models/work_something.js`
- `models/work_user.js`

## Query and Write Patterns

- Common reads: `findOne({ guild })`, `findOne({ guild, member })`, `findOne({ guild, user })`, `find({ guild })`, and lookups by `{ guild, id/messageid/number/channel_id }`.
- Common writes: `new Model(...).save()`, document `.delete()`, `findOneAndDelete`, `updateMany`, raw `doc.collection.updateOne()`, and read-modify-write followed by `save()` or `updateOne()`.
- No local evidence of Mongo `aggregate`, `populate`, query `limit`, or Mongoose query `sort`.
- Rank commands load all guild rows and sort in process.
- Daily reset uses `gift_change.distinct('guild', { time: { $ne: 0 } })`, `coin.updateMany({ guild: { $nin: excludedGuilds } })`, and work-energy updates.
- Work payout uses `work_user.find({ state: { $ne: "待業中" }, end_time: { $lte: round(now_seconds) } })` with an inner effective `end_time < round(now_seconds)` guard, increments/creates `coin`, reads `gift_change` for new-balance `today`, and resets `work_user.state`.

## Catalog

| Legacy model file | Mongoose model | Expected collection | Fields | Known query/write patterns | Indexes found | Indexes recommended | Compatibility risks | Repository methods |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `Number.js` | `Number` | `numbers` | `guild`, `parent`, many stat channel/name fields | guild singleton read/write/delete; channel rename cache | none | `{guild:1}` audit singleton before unique | model capitalization; many optional strings | stats config CRUD |
| `all_use_count.js` | `all_use_count` | `all_use_counts` | `slashcommand_name`, `count` | usage count lookup/increment | none | `{slashcommand_name:1}`; unique only after audit | can write undefined command names | command metrics increment |
| `ann_all_set.js` | `ann_all_set` | `ann_all_sets` | `guild`, `announcement_id`, `tag`, `color`, `title` | announcement config by guild/id; gated `公告頻道設置` writes by `{guild,announcement_id}` | none | `{guild:1}`, `{guild:1,announcement_id:1}` | optional color/title/tag; duplicate rows may exist before any unique index | announcement config CRUD; Go updates duplicate rows and creates no index |
| `birthday.js` | `birthday` | `birthdays` | `guild`, `user`, birthday date/time numbers, `allow` | user birthday by guild/user; list by guild; gated `是否允許管理員設定`, `刪除`, and `生日列表` now use this collection | none | `{guild:1,user:1}`, `{guild:1}` | scheduler inactive; missing date parts; duplicate `{guild,user}` rows may exist before unique index audit | birthday profile CRUD; Go updates/deletes duplicate `{guild,user}` rows and creates no index |
| `birthday_set.js` | `birthday_set` | `birthday_sets` | `guild`, `msg`, `utc`, `channel`, `everyone_can_set_birthday_date`, `role` | guild singleton config; gated `/生日系統 祝福語設定` writes by guild | none | `{guild:1}` audit singleton before unique | timezone and role optional; duplicate singleton rows may exist | birthday config CRUD; Go updates duplicate `{guild}` rows and creates no index |
| `btn.js` | `btn` | `btns` | `guild`, `number`, `role` | role button lookup by guild/number | none | `{guild:1,number:1}` | string/number ID payloads | button role config |
| `chat.js` | `chat` | `chats` | `guild`, `channel` | autochat config by guild/channel; gated `/自動聊天頻道` writes and `/自動聊天頻道刪除` deletes by guild | none | `{guild:1}` audit singleton before unique | MessageContent dependency for runtime chatbot; duplicate singleton rows may exist | chat config CRUD; Go updates/deletes duplicate `{guild}` rows and creates no index |
| `chat_role.js` | `chat_role` | `chat_roles` | `guild`, `leavel`, `role`, `delete_when_not` | XP level role by guild/level | none | `{guild:1,leavel:1,role:1}` | misspelled `leavel`; numeric-as-string risk | text level role rules |
| `chatgpt.js` | `chatgpt` | `chatgpts` | `guild`, `resid_c`, `resid_p`, `reply`, `message`, `time` | chatbot state/handoff by guild | none | `{guild:1}` | external worker unknown | chatbot state |
| `chatgpt_get.js` | `chatgpt_get` | `chatgpt_gets` | `guild`, `price` | guild price/config | none | `{guild:1}` | numeric price type | chatbot/economy config |
| `code.js` | `code` | `codes` | `code`, `price`, `time` | redeem code by code | none | `{code:1}` unique after audit | code lifetime semantics unclear | redeem code CRUD |
| `coin.js` | `coin` | `coins` | `guild`, `member`, `coin`, `today` | balance by guild/member; rank by guild; daily reset | none | `{guild:1,member:1}`, `{guild:1,coin:-1}`; unique after audit | `today` can be number/false/timestamp; races | economy balance and atomic increment |
| `create_hours.js` | `create_hours` | `create_hours` | `guild`, `hours`, `channel` | guild join-age config; gated `/帳號需創建時數` writes and member-add account-age policy reads by guild | none | `{guild:1}` candidate only after duplicate audit | `hours` is saved as a string number of seconds; `channel` may be string or null; singleton duplicates may exist | account-age config repository preserves the existing channel while changing hours, deletes duplicate guild rows on explicit config delete, and does not create indexes |
| `cron_set.js` | `cron_set` | `cron_sets` | `cron`, `guild`, `channel`, `id`, `message` | scheduled job by guild/id; list by guild | none | `{guild:1,id:1}`, `{guild:1}` | `message` is loose Discord payload | notification schedules |
| `errors_set.js` | `errors_set` | `errors_sets` | `guild`, `ban_count`, `move` | warning escalation config | none | `{guild:1}` | number/action optional | warning config |
| `ghp.js` | `ghp` | `ghps` | `guild`, `commodity_id`, `name`, `need_coin`, `commodity_description`, `code`, `auto_delete`, `role`, `commodity_count` | shop item by guild/commodity | none | `{guild:1,commodity_id:1}` | `need_coin` schema string but writes numbers | shop catalog |
| `gift.js` | `gift` | `gifts` | `guild`, `gift_name`, `gift_code`, `gift_chence`, `auto_delete`, `gift_count`, `give_coin` | gacha prize by guild/name; gated `/扭蛋獎池查詢` read-only list by guild | none | `{guild:1,gift_name:1}` | misspelled `gift_chence`; inventory races; large guild pools can exceed one Discord embed | gacha prize catalog; `ListGachaPrizes` reads only and preserves natural legacy order |
| `gift_change.js` | `gift_change` | `gift_changes` | `guild`, `coin_number`, `sign_coin`, `channel`, `xp_multiple`, `time` | economy/gacha config by guild; daily reset distinct; gated settings command patch-write by guild; gated `/扭蛋獎池查詢` reads prize-list config by guild | none | `{guild:1}`, `{time:1,guild:1}` | numeric fields may be strings/zero; `xp_multiple` is a Discord Number and must preserve fractional values | economy config; `SaveEconomyConfig` updates all duplicate guild rows before any future unique index; `GetGachaConfig` is read-only and missing config falls back to legacy defaults |
| `good_web.js` | `good_web` | `good_webs` | `guild`, `open` | anti-scam toggle by guild | none | `{guild:1}` | boolean toggle | anti-scam config |
| `guild.js` | `guild` | `guilds` | `guild`, `announcement_id`, `voice_detection` | guild singleton config; gated `公告頻道設置` writes `announcement_id` only | none | `{guild:1}` | generic model name; dashboard/shared settings must keep existing fields | guild config; Go patch-writes `announcement_id` and preserves unknown fields |
| `join_message.js` | `join_message` | `join_messages` | `guild`, `enable`, `message_content`, `color`, `channel`, `img` | guild singleton config; dashboard-shared welcome delivery read by guild | `unique` on `guild`, `enable` but autoIndex disabled | `{guild:1}`; unique after audit only | `enable` unique Boolean is unsafe; missing `enable` must decode as enabled because legacy checks `enable !== false` | join message config remains dashboard-owned; Go member-add delivery reads only and performs no config writes/index creation |
| `join_role.js` | `join_role` | `join_roles` | `guild`, `role`, `give_to_who` | join role config list by guild; gated `/加入身份組設置` insert and `/加入身份組刪除` delete by `{guild,role}` | none | `{guild:1,role:1}` unique only after duplicate audit | target semantics; duplicate rows may exist without unique index | join role config; Go uses `$setOnInsert` create and `DeleteMany` delete for duplicate cleanup, but does not enable member-add role assignment |
| `leave_message.js` | `leave_message` | `leave_messages` | `guild`, `message_content`, `title`, `color`, `channel` | guild singleton config | none | `{guild:1}` | optional embed fields | leave message config |
| `lock_channel.js` | `lock_channel` | `lock_channels` | `guild`, `channel_id`, `lock_anser`, `owner`, `text_channel`, `ok_people` | lock by guild/channel | none | `{guild:1,channel_id:1}` | password/plain text; array shape | voice lock state |
| `logging.js` | `logging` | `loggings` | `guild`, `channel_id`, `message_update`, `message_delete`, `channel_update`, `member_voice_update` | logging config by guild; gated `/set-log-channel` writes by guild | none | `{guild:1}` candidate only after duplicate audit | duplicates may exist because legacy delete+insert was not atomic; event emitters have privacy scope | logging config repository updates all duplicate guild rows with `$set` and only upserts when no rows match; no index is created by startup |
| `lotter.js` | `lotter` | `lotters` | `guild`, `date`, `gift`, `howmanywinner`, `id`, `member`, `end`, `message_channel`, `guild_new`, `yesrole`, `norole`, `maxNumber`, `owner` | lottery by guild/id; member array updates | none | `{guild:1,id:1}` | creation disabled; arrays loose | lottery repository |
| `message_reaction.js` | `message_reaction` | `message_reactions` | `guild`, `message`, `react`, `role` | reaction role by guild/message/react | none | `{guild:1,message:1,react:1}` | emoji encoding | reaction roles |
| `not_a_good_web.js` | `not_a_good_web` | `not_a_good_webs` | `web` | scam URL lookup/insert | none | `{web:1}` | regex injection; domain normalization | scam URL catalog |
| `poll.js` | `poll` | `polls` | `guild`, `messageid`, `question`, `create_member_id`, `many_choose`, `can_change_choose`, `can_see_result`, `end`, `anonymous`, `choose_data`, `join_member` | poll by guild/messageid; array updates | none | `{guild:1,messageid:1}` | loose arrays; concurrent voters | poll state |
| `role.js` | `role_number` | `role_numbers` | `guild`, `channel`, `channel_name`, `role` | role stat config by guild/role/channel | none | `{guild:1,role:1}`, `{guild:1,channel:1}` | file/model name mismatch | role statistics |
| `sign_list.js` | `sign_list` | `sign_lists` | `guild`, `member`, `date` | sign by guild/member/date; list by guild | none | `{guild:1,member:1}` | date object shape | sign-in history |
| `suport.js` | `suport` | `suports` | `support_id` | support ID lookup | none | `{support_id:1}` | misspelled model | support config |
| `system.js` | `system` | `systems` | `a`, `ram`, `cpu` | telemetry state | none | none until usage confirmed | unclear active use | telemetry if retained |
| `text_xp.js` | `text_xp` | `text_xps` | `guild`, `member`, `xp`, `leavel` | XP by guild/member; rank by guild | none | `{guild:1,member:1}`, `{guild:1}` | xp/level stored as strings | text XP repository |
| `text_xp_channel.js` | `text_xp_channel` | `text_xp_channels` | `guild`, `channel`, `background`, `color`, `message` | XP announcement config; gated `聊天經驗設定`/`聊天經驗刪除` writes by guild | none | `{guild:1}` candidate only after duplicate audit | optional image/color; singleton duplicates may exist | text XP config repository updates/deletes all duplicate guild rows, clears `background` on set, and does not create indexes |
| `ticket.js` | `ticket` | `tickets` | `guild`, `ticket_channel`, `admin_id`, `everyone_id` | ticket config by guild | none | `{guild:1}` | role/channel optional | ticket config |
| `verification.js` | `verification` | `verifications` | `guild`, `role`, `name` | verification config by guild | none | `{guild:1}` | role/name optional | verification config |
| `voice_channel.js` | `voice_channel` | `voice_channels` | `guild`, `ticket_channel`, `limit`, `name`, `parent`, `lock` | voice room config by guild/ticket_channel | none | `{guild:1,ticket_channel:1}` | trigger channel naming | voice room config |
| `voice_channel_id.js` | `voice_channel_id` | `voice_channel_ids` | `guild`, `channel_id` | dynamic voice channel state | none | `{guild:1,channel_id:1}` | cleanup on restart | dynamic voice state |
| `voice_role.js` | `voice_role` | `voice_roles` | `guild`, `leavel`, `role`, `delete_when_not` | voice level role config | none | `{guild:1,leavel:1,role:1}` | misspelled `leavel` | voice level roles |
| `voice_xp.js` | `voice_xp` | `voice_xps` | `guild`, `member`, `xp`, `leavel`, `leavejoin` | voice XP by guild/member; active session | none | `{guild:1,member:1}`, `{guild:1}` | strings; `leavejoin`; missing save bug | voice XP repository |
| `voice_xp_channel.js` | `voice_xp_channel` | `voice_xp_channels` | `guild`, `channel`, `background`, `color`, `message` | voice XP announcement config | none | `{guild:1}` | optional image/color; legacy `背景` option is visible but not saved by `voice_set.js` | voice XP config; Go config slice updates duplicate `{guild}` rows, clears `background`, and is gated by `MHCAT_FEATURE_VOICE_XP_CONFIG_ENABLED` |
| `vote.js` | `vote` | `votes` | `guild`, `Number`, `member`, `vote` | vote config/state | none | `{guild:1,Number:1}`, `{guild:1,member:1}` | capitalized `Number` field | vote repository if active |
| `warndb.js` | `warndb` | `warndbs` | `time`, `guild`, `user`, `content` | warning docs by guild/user | none | `{guild:1,user:1}` | content array shape | warnings repository |
| `work_set.js` | `work_set` | `work_sets` | `guild`, `get_energy`, `max_energy`, `captcha` | guild work config | none | `{guild:1}` | energy numeric drift | work config |
| `work_something.js` | `work_something` | `work_somethings` | `guild`, `name`, `time`, `energy`, `coin`, `role` | work task by guild/name | none | `{guild:1,name:1}` | task name as ID | work task catalog |
| `work_user.js` | `work_user` | `work_users` | `guild`, `user`, `state`, `end_time`, `energi`, `get_coin` | work user by guild/user; active jobs by state/end_time; completed-work payout by non-idle state and `end_time` | payout resets `state` to `待業中`; daily reset increments/clamps `energi` | `{guild:1,user:1}`, `{state:1,end_time:1}` | misspelled `energi`; scheduler races; duplicate guild/user rows can duplicate payouts; payout crash between coin increment and state reset can repeat payout | work user state and one-shot payout |

Work interface/start/admin repository status:

- `internal/adapters/mongo/repositories.WorkInterfaceRepository` now reads `work_sets`, `work_somethings`, and `work_users`.
- Supported read methods: `GetWorkConfig`, `ListWorkItems`, and `GetWorkUser`.
- Supported write method: `StartWork`, exposed only through the explicit `ports.WorkStartRepository` app wiring path.
- `StartWork` can upsert a missing `work_users` row with legacy fields, then atomically deduct `energi` and set `state`, `end_time`, and `get_coin`.
- Supported admin write methods: `SaveWorkConfig`, `DeleteWorkItem`, `GrantWorkEnergy`, and `GrantWorkEnergyToAll`, exposed only through the explicit `ports.WorkAdminRepository` app wiring path.
- `SaveWorkConfig` updates/upserts `work_sets` by guild using legacy field names. It updates all duplicate legacy config rows for that guild so reads remain rollback-compatible until duplicate audit/index work is complete.
- `DeleteWorkItem` deletes matching `work_somethings` rows by `{guild,name}` and reports the legacy missing-item error when no row matches.
- `GrantWorkEnergy` upserts the target `work_users` row with `energi=max_energy` and then clamps existing energy with a Mongo aggregation update pipeline. This intentionally fixes the legacy missing-target bug that created the actor's row instead.
- `GrantWorkEnergyToAll` clamps existing `work_users` rows for the guild only and does not create missing user rows.
- No write methods are exposed for direct work item creation, payout, coin increments, scheduler state, or indexes in this slice.
- BSON compatibility preserves `energi` and mixed numeric values through document conversion tests.

## Go Document Struct Guidance

- Every struct must include exact legacy `bson` tags, including misspellings such as `leavel`, `gift_chence`, `lock_anser`, `energi`, and `suport`.
- Include `primitive.ObjectID` for `_id` where repositories need identity, but keep it out of core domain unless behavior requires it.
- Decode unknown fields safely by default.
- Use custom compatibility decoders or raw BSON audits for mixed string/number/bool fields before enforcing strict types.
- If a new canonical schema is introduced, support dual-read and rollback-compatible writes until the rollout plan declares the legacy format retired.

## Phase 1.5 Live Audit Status

No `MONGODB_URI` / `MONGOOSE_CONNECTION_STRING` plus `MONGODB_DATABASE` were available in the Phase 1.5 shell environment, so no live database query was executed.

Read-only audit tooling was added at `MHCAT-REFACTOR/tools/mongo-audit-readonly.mjs`. It is designed to inspect:

- existing collections;
- document counts;
- current indexes;
- sample document shapes;
- missing required fields;
- mixed field types;
- duplicate logical keys;
- impossible negative values for known counters;
- large documents over 256 KiB;
- collections not represented by Mongoose models;
- Mongoose model collections not represented by the live database.

The tool must be run and reviewed before:

- implementing repository writes against production data;
- applying any unique index;
- applying any TTL index;
- performing schema repair/backfill;
- enabling production canary writes.

Dashboard compatibility evidence from `../mhcat-mono/mhcat-dashboard` confirms these bot collections are shared with external settings/backup flows: `join_messages`, `guilds`, `work_somethings`, `warndbs`, and the full guild backup list in `docs/13-external-compatibility.md`.

## Mongo Risks

- Mongoose `autoIndex: false` means declared indexes may not exist in production.
- Unique indexes must not be assumed until data audit checks for duplicates.
- Local code does not prove actual production collection names or indexes.
- Dashboard schema declarations must not be copied blindly; `join_message.enable` and dashboard `work_something.guild` unique declarations are unsafe without live duplicate audit and behavior review.
- Dashboard backup includes singular `message_reaction`; live audit must confirm whether that collection exists and whether it differs from `message_reactions`.
- Many collections are singleton-per-guild by intent but may contain duplicates.
- Non-atomic read-modify-write exists for coins, XP, gacha inventory, poll voters, lottery members, sign lists, work state, and warnings.
- Several fields have mixed/loose types: XP/level strings, numeric counters as strings, booleans in numeric fields, loose arrays/objects.

## Tests Required

- Collection-name compatibility tests.
- BSON fixture decode tests for every model.
- Missing-field and unknown-field decode tests.
- Mixed-type tests for XP/level, coin/today, shop price, energy, booleans, and loose arrays/objects.
- Repository contract tests for not found, upsert, delete, atomic increment, duplicate handling, and context cancellation.
- Duplicate-audit and dry-run repair tests for every unique candidate.

## Platform Wave B Go Catalog Implementation

Platform Wave B moves the typed catalog to `internal/adapters/mongo/catalog.go`.

Coverage:

- all 47 legacy `MHCAT/models/*.js` files;
- explicit `LegacyModelFile` and `LegacyMongooseModel` metadata;
- corrected Mongoose-compatible collection names such as `coins`, `text_xps`, `voice_xps`, `polls`, `tickets`, `guilds`, `cron_sets`, `verifications`, and `chatgpts`;
- special-case coverage for `create_hours`, `role_number`, `suport`, and dashboard-shared collections.

Every catalog entry remains marked incomplete for field-level BSON strictness. The catalog is complete for model/collection identity, but it still exists to drive read-only audit output and index diff planning, not to assert that production data is clean.

Current policy:

- Collection specs are complete for legacy model coverage and partial for field-level schema validation.
- Unique planned indexes require duplicate audit.
- Missing/unknown live collections are reported, not repaired.
- Raw document values are not included in audit output by default.
- No feature repository writes are implemented in Platform Wave B.
- Contract tests fail if a legacy model file is missing from the catalog, a catalog entry references a missing legacy file, or the corrected catalog regresses to singular placeholder names.

## Go System Collections

These collections are Go operational infrastructure, not legacy Mongoose model collections. They are intentionally not part of `DefaultCollectionCatalog()` because that catalog is contract-tested against the 47 legacy model files.

| Collection | Purpose | Fields | Indexes | Compatibility notes |
| --- | --- | --- | --- | --- |
| `mhcat_scheduler_locks` | single-owner lease primitive for future recurring jobs | `_id` equal to `lock_name`, `lock_name`, `owner`, `fence`, `expires_at`, `created_at`, `updated_at` | default Mongo `_id` only | no Node dependency; no backfill; disable Go schedulers and leases naturally expire for rollback |

## Warning History Repository Status

The warning-history slice adds a read-only repository for the legacy `warndbs` collection.

- `internal/adapters/mongo/repositories.WarningHistoryRepository` reads by `{guild,user}` only.
- It returns `ErrWarningsNotFound` for missing documents and empty `content` arrays to preserve the legacy user-facing "no warnings" branch.
- It does not create, update, delete, backfill, repair, or index warning documents.
- It does not implement warning creation/removal/escalation, `errors_sets`, bulk delete, kick, ban, or moderation cleanup paths.
- Moderator lookup failures are handled in the Discord feature layer by falling back to the stored moderator ID, intentionally avoiding the legacy cached-member nil crash.
- Before enabling broader moderation writes, audit `warndbs` for duplicate `{guild,user}` rows, mixed `content` element shapes, missing `moderator`/`reason`/`time` fields, and dashboard backup compatibility.
