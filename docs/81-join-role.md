# Join-Role Parity Contract

Status: parity-audited against the active legacy setup/delete commands, `guildMemberAdd` role loop, Mongoose schema, slash registration/dispatcher, config colors, and Discord cache behavior. Config and assignment remain disabled by default. Live staging smoke is still required before production ownership.

## Scope

This contract covers:

- `/加入身份組設置`;
- `/加入身份組刪除`;
- `join_roles` Mongo compatibility;
- automatic role assignment on `guildMemberAdd`;
- usage, intents, ownership, staging, and rollback.

Legacy sources:

- `slashCommands/加入設置/join_role.js`
- `slashCommands/加入設置/join_role_delete.js`
- `events/welcome.js`
- `models/join_role.js`
- `handler/slash_commands.js`
- `events/SlashCommands.js`
- `config.json`

Welcome/leave messages, account-age policy, verification, and role-selection are separate families.

## Gates And Ownership

Config commands require paired staging-only flags:

```bash
MHCAT_FEATURE_JOIN_ROLE_CONFIG_ENABLED=true
MHCAT_STAGING_MODE=true
MHCAT_COMMAND_SYNC_INCLUDE_JOIN_ROLE_CONFIG=true
```

Assignment is event-only and independently requires:

```bash
MHCAT_FEATURE_JOIN_ROLE_ASSIGNMENT_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true
```

Command sync is guild-scoped and staging-only. Stop matching Node command and member-add owners before enabling Go. Config may migrate separately from assignment, but concurrent Node/Go writes or role assignment must not overlap.

## Command And UI Contract

Both definitions preserve exact names, descriptions, role option, optional audience option, choice order/text/values, required flags, and malformed legacy documentation URLs. Legacy `UserPerms` only appears in help; registration applies no default member permissions. Both commands remain publicly discoverable, defer publicly, and enforce Manage Messages at runtime. Administrator satisfies the runtime check.

Legacy cooldown metadata is `10` but the global dispatcher does not enforce it. Go adds no cooldown.

Setup checks that the selected role exists below the bot's highest role before writing. Omitted audience defaults to `all_user`. Exact visible behavior includes:

- title `🪂 加入身分組系統`;
- success color `#53FF53` from legacy `client.color.greate`;
- red Discord error color and animated-no prefix;
- exact success/delete descriptions and emoji IDs;
- legacy `<@roleID>` text shape rather than `<@&roleID>`;
- exact permission, hierarchy, duplicate, and missing-config titles.

The public defer determines visibility; legacy `ephemeral:true` supplied to an edit cannot convert the original response. Go suppresses all mentions so the legacy text cannot ping.

Usage belongs only to global slash middleware. With usage tracking enabled, exactly one best-effort event is recorded before route/permission/validation checks for every setup/delete attempt. Handlers and member events do not write usage directly.

## Mongo Compatibility

Collection and fields remain exact:

- `join_roles.guild`: Mongoose String guild ID;
- `join_roles.role`: Mongoose String role ID;
- `join_roles.give_to_who`: optional Mongoose String audience.

Separate read/write DTOs preserve usable Mongoose String scalars. Missing, null, or empty audience remains `all_user`. Compound audience values are unusable and skipped rather than becoming universal grants. Missing/null/compound roles are invalid rows and do not abort decoding of later rows. Unknown nonempty audiences remain invalid and are skipped.

Writes are typed BSON strings. Create uses an atomic `$setOnInsert` upsert keyed by `{guild,role}` and returns the legacy duplicate error for an existing match. Without a unique index, concurrent first creates can still produce duplicates, like legacy find-then-save. Explicit delete removes all duplicate `{guild,role}` matches as an intentional cleanup fix.

Assignment reads all matching guild rows in Mongo natural order and reverses them to preserve the legacy reverse loop. No startup index, repair, deduplication, normalization, or backfill runs.

The candidate unique `{guild:1,role:1}` index remains blocked on duplicate/missing/null/blank/scalar-drift keys, malformed audiences, stale roles, external writers, and exclusive ownership review. No migration is required merely to enable Go.

## Member-Add Assignment

Audience behavior remains:

- `all_user` or falsy: every joining account;
- `all_bot`: bots only;
- `all_member`: non-bot users only;
- unknown/compound: no assignment.

For each applicable row, production Go uses cached guild roles and the cached bot member, matching `guild.roles.cache` and `guild.members.me`. Missing cached roles are skipped. A role at or above the bot's highest role sends the guild owner the exact legacy DM text through a REST-backed owner lookup. Assignable roles are added through Discord REST.

Go intentionally fixes the legacy loop's order-dependent `return` statements. A nonmatching audience, missing role, hierarchy failure, owner-DM failure, or role-add failure no longer suppresses valid older rows. Adds and warnings are awaited, later rows are attempted, and joined errors are reported through continue-and-report event semantics so independent welcome behavior can still run.

Account-age policy is registered first and stops propagation for a matched too-new account. Welcome delivery and join-role assignment are separately gated. Their successful ordering is deterministic in Go but was callback-race dependent in Node; neither contract depends on which independent side effect appears first.

## Intentional Differences

- Explicit delete removes duplicate matching rows.
- The member-add loop continues after nonmatching/bad/failing rows.
- Role adds and owner warnings are awaited and errors are logged.
- Event role/hierarchy checks are explicitly cache-only.
- Mentions are suppressed.
- Global middleware is the only slash usage owner.

Exact command metadata/UI, runtime permission, audience behavior, reverse row order, cached hierarchy semantics, owner warning text, collection/fields, and account-age stop behavior are preserved.

## Migration And Staging

1. Use an isolated staging guild/database, disposable roles, and disposable members/bots.
2. Stop matching Node config and member-add owners.
3. Audit duplicate/malformed `join_roles` rows, scalar types, audience values, stale role IDs, external writers, and indexes.
4. Preserve data as-is. Do not normalize audiences, deduplicate, backfill, or index merely to enable Go.
5. Pair config runtime/sync flags; run preflight and command-sync dry-run before reviewed apply.
6. Enable assignment only with Gateway/Guild Members and confirmed role-cache readiness.
7. Confirm account-age policy ownership/order when both member-add families are staged.

## Parity Tests

Focused coverage locks metadata/public visibility, runtime permissions, exact UI/colors/errors, audience defaults, scalar/compound reads, typed writes, duplicate behavior, usage ownership, reverse/event rules, cached role hierarchy, exact owner warning, failure continuation, account-age propagation, app wiring, gates, command sync, and preflight. Run:

```bash
go test ./internal/core/domain ./internal/core/services/onboarding ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/adapters/discordgo ./internal/discord/features/onboarding ./internal/discord/events ./internal/app ./internal/config ./internal/parity ./cmd/mhcat-command-sync ./cmd/mhcat-staging-preflight
go test -race ./internal/core/services/onboarding ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/adapters/discordgo ./internal/discord/features/onboarding ./internal/discord/events ./internal/app
go vet ./internal/core/services/onboarding ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/adapters/discordgo ./internal/discord/features/onboarding ./internal/discord/events ./internal/app
```

## Staging Smoke

1. Review the read-only `join_roles` audit and confirm one owner per family.
2. Run preflight, command-sync dry-run, reviewed guild apply, and config runtime startup.
3. Confirm both commands are publicly discoverable; test runtime permission denial and Administrator behavior.
4. Verify exact setup/delete UI, `#53FF53`, omitted/three audience choices, hierarchy rejection, duplicate/missing errors, mention suppression, and one usage event per attempt.
5. Seed scalar, missing/null/empty/unknown/compound audience rows, stale roles, duplicates, and multiple natural-order rows in disposable data; verify safe reads and reverse processing.
6. With assignment disabled, confirm config changes do not affect joining users.
7. Enable Gateway/Guild Members assignment and join with a human and bot; verify only matching roles.
8. Test missing and too-high roles, exact owner DM, denied owner DM, denied role add, and continuation to later valid rows.
9. Enable welcome separately and confirm independent failures do not suppress the other family.
10. Enable account-age separately and confirm matched accounts receive no welcome/join roles while allowed accounts continue.
11. Disable gates, remove only managed staging commands, preserve Mongo data, and perform rollback checks.

## Rollback

1. Disable command-sync inclusion and remove only the two managed staging commands.
2. Disable assignment and config runtime gates; stop every Go owner.
3. Preserve `join_roles`; typed writes remain Mongoose-readable. Do not repair data or indexes during emergency rollback.
4. Restore Node commands and/or `welcome.js` role loop only after confirming no Go owner remains.
5. Verify one config create/delete and one human/bot assignment matrix in staging.
6. Review any overlap interval for duplicate role attempts or config rows.

Production ownership remains blocked on live staging smoke and a reviewed duplicate/type/writer audit. The index is optional and must never be created merely to enable Go.
