# Production Mongo Read-only Audit

Status: read-only SSH audit completed on 2026-07-04 Asia/Taipei.

## Method

- Connected to `windows-vm` over SSH.
- Read the production Mongo URI from `/root/mhcat/.env` without printing it.
- Passed the URI to `mongosh` through an environment variable inside the `mongodb` container.
- Ran read-only commands only:
  - collection listing;
  - estimated document counts;
  - current index listing;
  - sampled field/type shapes with values omitted;
  - limited duplicate-risk checks for known logical keys.

No Mongo writes, index creation, repair, backfill, command registration, or bot startup happened.

## Database Summary

- Database: `mhcat-database`.
- Live collection count: 51.
- Mongo container status: healthy.
- Most collections currently have only the default `_id_` index.
- Existing non-default indexes found:
  - `text_xps`: `{ guild: 1, member: 1 }`, non-unique, name `guild_1_member_1_autocreated`.
  - `voice_xps`: `{ guild: 1, member: 1 }`, non-unique, name `guild_1_member_1_autocreated`.

## High-volume Collections

| Collection | Estimated documents | Index notes |
| --- | ---: | --- |
| `text_xps` | 2,061,173 | non-unique `{ guild, member }` exists |
| `voice_xps` | 978,166 | non-unique `{ guild, member }` exists |
| `coins` | 223,351 | `_id_` only |
| `sign_lists` | 59,705 | `_id_` only |
| `btns` | 21,462 | `_id_` only |
| `work_users` | 16,707 | `_id_` only |
| `message_reaction` | 15,329 | `_id_` only |
| `gifts` | 14,548 | `_id_` only |
| `birthdays` | 9,683 | `_id_` only |
| `ghps` | 7,849 | `_id_` only |

## Compatibility Findings

- Live collection names are Mongoose-pluralized, for example `coins`, `text_xps`, `voice_xps`, `polls`, `tickets`, `verifications`, and `work_users`.
- Both `message_reaction` and `message_reactions` exist and have the same sampled field shape. Treat this as a compatibility risk before implementing reaction-role repositories.
- Some sampled fields have mixed or nullable types that Go document structs must tolerate:
  - `errors_sets.ban_count`: number and string;
  - `create_hours.channel`: string and null;
  - `chatgpts.resid_c` / `resid_p`: string and null;
  - `ghps.need_coin`: string despite numeric meaning;
  - `text_xps.xp`, `text_xps.leavel`, `voice_xps.xp`, `voice_xps.leavel`, `voice_xps.leavejoin`: string despite numeric meaning;
  - `gift_changes` sampled documents do not all contain the same optional fields.
- `userdatas` contains `accessToken`; do not include raw document values in future audit output.
- The limited duplicate-risk checks sampled no duplicate groups for the tested high-risk logical keys, but this is not a substitute for a full duplicate audit before creating unique indexes.

## Index Planning Impact

- Do not create unique indexes on production until a full duplicate audit runs for the exact key and confirms zero duplicates.
- Treat `coins` as a high-priority future index candidate for `{ guild: 1, member: 1 }`, but apply only after duplicate audit.
- Keep `text_xps` and `voice_xps` indexes non-unique unless a future duplicate audit and ADR justify unique constraints.
- Do not drop or rename existing indexes in the refactor.
- Do not create TTL indexes until a retention ADR exists.

## Next Step

Use the local Compose Mongo service for host-side smoke runs. Use production Mongo only for read-only audits until repair/backfill/index-apply tooling is explicitly approved with dry-run evidence.
