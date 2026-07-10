# External Compatibility

Status: Phase 1.5 Gate B input. Evidence comes from the local legacy checkout and the nearby local `../mhcat-mono` workspace. No network access, installs, writes, or legacy edits were performed.

## Summary

The MongoDB collections are not private to the Discord bot. A local `mhcat-dashboard` repository exists and actively reads/writes several bot collections. It also exposes a guild backup API that treats many bot collections as a public export contract.

Design consequence: schema changes are allowed, but existing collection names and legacy fields must remain readable until the dashboard and any workers are updated. Prefer additive changes, dual-read compatibility, and rollback-compatible writes.

## Dependency Matrix

| System | Evidence | Collections used | Fields used | Read/write | Compatibility risk | Required adapter behavior |
| --- | --- | --- | --- | --- | --- | --- |
| `mhcat-dashboard` admin/settings | Legacy README links dashboard; bot commands redirect welcome/work settings to dashboard; dashboard APIs write settings under `../mhcat-mono/mhcat-dashboard/src/pages/api/savedata/**` | `join_messages`, `guilds`, `work_somethings`; reads `warndbs`; auth cache `userdatas` | `join_messages`: `guild`, `enable`, `message_content`, `color`, `channel`, `img`; `guilds`: `guild`, `voice_detection`; `work_somethings`: `guild`, `name`, `time`, `energy`, `coin`, `role`; `warndbs`: `guild`, `user`, `time`, `content`; `userdatas`: `id`, `accessToken` | Read/write/delete for settings; read-only for warnings; read/write for dashboard auth cache | High: live admin paths share bot collections; dashboard schemas contain unsafe uniqueness such as `work_something.guild` unique while bot behavior uses `{guild,name}` | Preserve exact collection/model names and BSON fields. Keep `work_somethings` keyed by `{guild,name}`. Accept `null` for optional `img`, `role`, and `voice_detection`. Do not create unique indexes copied from dashboard schemas without duplicate audit. |
| `mhcat-dashboard` backup/export | `src/pages/api/backup/[id].js` enumerates guild collections and reads from `mhcat-database` | `ann_all_sets`, `birthdays`, `birthday_sets`, `btns`, `chats`, `chat_roles`, `chatgpts`, `chatgpt_gets`, `coins`, `create_hours`, `cron_sets`, `errors_sets`, `ghps`, `gifts`, `gift_changes`, `good_webs`, `guilds`, `join_messages`, `join_roles`, `leave_messages`, `lock_channels`, `loggings`, `lotters`, `message_reactions`, singular `message_reaction`, `numbers`, `polls`, `role_numbers`, `sign_lists`, `text_xps`, `text_xp_channels`, `tickets`, `verifications`, `voice_channels`, `voice_channel_ids`, `voice_roles`, `voice_xps`, `voice_xp_channels`, `votes`, `warndbs`, `work_sets`, `work_somethings`, `work_users` | Filters by `guild`; exports full documents | Read-only export | High: renaming/removing fields or moving collections breaks user backups and restores | Keep legacy collections readable during rollout. If new canonical collections are added, update dashboard backup list before relying on them operationally. Keep `guild` on guild-scoped documents. |
| Dashboard DB selection | Dashboard strips the DB path from `MONGO_URI` in `src/util/connect/connectMongodb.js`; backup explicitly calls `useDb('mhcat-database')` | All dashboard model collections plus backup database | DB name and collection names | N/A | High until production URI/default DB behavior is confirmed | Go config must make DB name explicit. Production rollout must verify bot DB and dashboard DB are the same or intentionally bridged. |
| ChatGPT service/worker | Bot has no OpenAI dependency. `events/Chatbot.js` writes prompt state into `chatgpts`, waits, then rereads `chatgpts.message`; no local worker code found | `chats`, `chatgpts`, `chatgpt_gets` | `chats.guild/channel`; `chatgpts.guild/resid_c/resid_p/reply/message/time`; `chatgpt_gets.guild/price` | Bot read/write; external writer inferred but unconfirmed | High if the worker is still deployed; it likely depends on the same handoff fields | Preserve the handoff contract until the worker is confirmed retired. Patch fields, preserve `resid_*`, `reply`, `message`, `time`, and numeric `price`. |
| Cron/worker | No separate cron worker repo found locally; cron runs inside legacy bot `handler/cron.js`; Go restores automatic notifications and daily reset as separate recurring workers and has a crash-idempotent payout repository | `cron_sets`, `coins`, `gift_changes`, `work_sets`, `work_users`, `mhcat_scheduler_locks` | `cron`, `guild`, `channel`, `id`, loose `message`; `coin.today`; additive `coin.mhcat_work_payouts`; `work_user.energi`; `work_set.max_energy/get_energy`; `gift_change.time`; lease owner/fence/expiry | Bot internal read/write | High: Node does not participate in Go leases or payout markers, so overlapping owners can duplicate sends, resets, or payouts | Keep payloads and payout markers rollback-compatible. Every Go process needs a unique owner; Node `handler/cron.js` must be disabled before Go ownership. Dashboard full-document backup/export should preserve the additive coin marker. |
| Webhook/reporting systems | Env webhooks exist; report command sends `REPORT_WEBHOOK`; hardcoded lifecycle webhook exists in `MHCAT/index.js:253` and is redacted in risk register | `not_a_good_webs`, `good_webs` | `not_a_good_webs.web`, `good_webs.guild/open` | Bot reads/toggles; webhook send only; possible manual external action unconfirmed | Medium: external moderation may consume reports and manually update bad URL collection | Preserve `not_a_good_webs.web` and `good_webs.open`. Confirm reporting workflow before changing anti-scam schema. Rotate/revoke hardcoded webhook. |
| Website/docs/admin panel | `mhcat-docs` is static Docusaurus content; search found no Mongo/Mongoose usage | None found | None | None | Low | No DB adapter needed. Update public docs/site links only if dashboard routes, backup format, or feature ownership changes. |
| Nearby `../mhcat-mono/mhcat` bot copy | Local sibling repo appears to be another MHCAT bot copy | Same as legacy bot if deployed | Same as legacy bot | Potential full bot read/write | High only if deployed against same token/database | Confirm production source of truth. Never run two bots against the same token/guilds without exclusive event and scheduler ownership. |

## Dashboard-Owned or Dashboard-Shared Collections

These collections are shared contracts during the refactor:

- `join_messages`: dashboard reads and writes welcome configuration. Go must preserve `guild`, `enable`, `message_content`, `color`, `channel`, and `img`.
- `guilds`: dashboard reads/writes `voice_detection`; legacy also stores announcement-related guild config.
- `work_somethings`: dashboard creates, updates, deletes, and lists work jobs by `{guild,name}`. Do not enforce `guild` uniqueness even though the dashboard schema declares it.
- `warndbs`: dashboard reads warnings by `guild`.
- `userdatas`: dashboard-owned OAuth/session cache, not a bot collection. Go bot should not depend on it unless a future ADR introduces dashboard integration.
- Backup/export list: dashboard users may expect all listed guild-scoped collections to remain exportable.

## Required Adapter Behavior

- Existing collection names remain readable by Go.
- Existing fields remain readable and are not renamed in-place.
- Go repositories use patch-style updates and avoid full document replacement for shared collections.
- Any new canonical fields are additive and optional while the dashboard is still on the legacy schema.
- Any unique index on shared collections must be preceded by live duplicate audit and dashboard compatibility review.
- Dashboard backup impact must be included in every schema-change ADR.
- If a new state collection is introduced, the dashboard backup/export behavior must be explicitly decided: include, exclude, or admin-only export.

## Manual Confirmations Required

- Confirm production bot DB name and dashboard DB name, especially dashboard `MONGO_URI` default DB vs backup `mhcat-database`.
- Confirm whether `mhcat-dashboard` is currently deployed and whether users rely on backup/export.
- Confirm whether an external ChatGPT worker still exists and obtain its schema contract.
- Confirm whether any webhook/reporting system writes `not_a_good_webs`.
- Confirm whether `MHCAT/` or `../mhcat-mono/mhcat` is the production bot source.
- Run live Mongo audit for singular `message_reaction`, `userdatas`, existing indexes, and dashboard-declared unique indexes.

## Gate B Decision

External compatibility is resolved enough to start Wave 1 skeleton work because Wave 1 will not write shared collections, register commands, or change schemas. It is not resolved enough to start feature repositories, schema cleanup, production index creation, or ChatGPT feature parity until the manual confirmations and live Mongo audit are complete.
