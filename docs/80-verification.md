# Verification Parity Contract

Status: parity-audited against the active legacy setup command, captcha command, button/modal handlers, Mongoose schema, slash dispatcher, and discord.js registration/response behavior. Both command families remain disabled by default. Live staging smoke is still required before production ownership.

## Scope

This contract covers:

- `/驗證設置` configuration;
- `/驗證` captcha generation and prompt;
- the verification button and answer modal;
- configured role assignment and optional nickname replacement;
- `verifications` compatibility, usage, ownership, staging, and rollback.

Legacy sources:

- `slashCommands/加入設置/verification_set.js`
- `slashCommands/加入設置/verification.js`
- `events/btn.js`
- `events/modal.js`
- `models/verification.js`
- `handler/slash_commands.js`
- `events/SlashCommands.js`

Join-role, welcome/leave delivery, account-age policy, and delete-data are separate ownership families.

## Gates And Ownership

Setup requires paired staging-only flags:

```bash
MHCAT_FEATURE_VERIFICATION_CONFIG_ENABLED=true
MHCAT_STAGING_MODE=true
MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_CONFIG=true
```

The flow is independently paired:

```bash
MHCAT_FEATURE_VERIFICATION_FLOW_ENABLED=true
MHCAT_STAGING_MODE=true
MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_FLOW=true
```

These interaction/REST paths require neither Gateway nor Guild Members intent. Command sync is guild-scoped and staging-only. Stop the matching Node command/button/modal owners before enabling Go. Config and flow may migrate separately, but Node and Go must not concurrently own either family.

## Command And Usage Contract

Both definitions preserve exact names, descriptions, setup option order/types/text, required flags, and the legacy announcement documentation URL. Legacy `UserPerms` is help metadata only: registration sends no default member permission restriction, so `/驗證設置` remains publicly discoverable. Its handler publicly defers and requires Manage Messages at runtime. `/驗證` has no options and defers ephemerally.

Legacy cooldown metadata (`10` for setup, `5` for flow) is not enforced by the global dispatcher. Go adds no cooldown.

Setup preserves the exact success/error titles, green/red colors, emoji IDs, role mention text, raw rename whitespace, and literal `null` display when rename is absent. Role hierarchy is checked against the bot before saving. The public defer determines visibility; legacy `ephemeral:true` supplied to an edit cannot convert that original public response. Go suppresses role/everyone mentions.

Usage belongs only to global slash middleware. With usage tracking enabled, one best-effort event is recorded before route lookup and validation for every setup or flow attempt, including permission denial and missing config. Buttons, modals, and handlers do not record usage locally.

## Captcha And Completion

The prompt is an ephemeral `captcha.jpeg` with the exact green `點我進行驗證!` button and arrow emoji. The modal title and input label remain `請輸入驗證碼!` and `請輸入圖片上的驗證碼`. Slash/button errors use the animated-no prefix; modal errors retain legacy plain red titles. Modal submits defer publicly, matching the legacy handler.

Answers are exact, case-sensitive, and whitespace-sensitive. Role assignment occurs before optional rename handling. For the guild owner, the role is therefore assigned before the exact owner nickname error. Rename replaces only the first `{name}` and preserves JavaScript replacement-string behavior for `$$`, `$&`, ``$` ``, and `$'` in usernames. Role and nickname operations are awaited before success.

The captcha contract remains:

- 400x250 JPEG at quality 75;
- six letters from `ABCDEFHIJLMNOPSTUVWXYZ`;
- ten black crossing lines, 200 black circles, and 5,000 colored noise points;
- centered 90px text rotated between -0.5 and 0.5 radians;
- `Swift.ttf` when available, with documented font fallbacks.

Behavior is aligned with pinned `@haileybot/captcha-generator@1.7.0`; Go and node-canvas output is not expected to be pixel-identical.

## Challenge State And Legacy IDs

New prompts use `mhcat:v1:verification:prompt:state=<id>` and `mhcat:v1:verification:answer:state=<id>`. The answer never appears in new IDs. State is bound to guild/user, expires at the inclusive five-minute boundary, and is process-local.

Versioned completion atomically claims state so concurrent submissions cannot duplicate role/nickname side effects. Wrong answers and failed side effects release the claim for retry. Success consumes state after side effects. A failed nickname after a successful role add retains state and can repeat the idempotent role-add attempt on retry.

Strict `[A-Za-z0-9]{1,16}verification` and matching `<answer>ver` legacy IDs remain supported for live messages. They expose answers and remain reusable because no server-side state exists. Malformed broad-substring IDs are intentionally rejected. Missing button config now returns a controlled error instead of reproducing the legacy null dereference.

## Mongo Compatibility

Collection and fields remain exact:

- `verifications.guild`: Mongoose String guild ID;
- `verifications.role`: Mongoose String role ID;
- `verifications.name`: nullable/optional Mongoose String rename template.

Separate read/write DTOs preserve Mongoose behavior. Reads scalar-cast usable strings, symbols, JavaScript values, booleans, numbers, decimals, and ObjectIDs. Missing/null/compound optional names become no rename. Missing/null/compound roles become invalid controlled config instead of BSON decode failures. Rename whitespace is preserved exactly.

Writes trim only guild/role IDs and emit typed BSON strings/null. Existing duplicate guild rows are all updated; an absent row is upserted. Reads retain legacy first-row behavior. No startup index, repair, deduplication, or backfill runs.

The candidate unique `{guild:1}` index remains blocked on duplicate, missing/null/blank/scalar-drift guild keys, unusable roles, external/dashboard writers, and exclusive ownership review. No database migration is required merely to enable Go.

## Intentional Differences

- New IDs hide answers, bind guild/user, expire, and serialize successful completion.
- Legacy IDs remain compatible but strict and reusable.
- Missing button config and malformed state return controlled errors.
- Duplicate config rows are updated together instead of delete/reinsert.
- Discord side effects are awaited; failed side effects retain versioned state.
- Allowed mentions are explicitly suppressed.
- Global middleware is the only slash usage owner.

Exact metadata/UI, runtime permissions, answer comparison, side-effect order, nickname replacement, Mongo names/fields, and legacy live-ID support are preserved.

## Migration And Staging

1. Use an isolated staging guild/database, disposable role, and disposable member.
2. Stop matching Node setup/flow/component/modal owners.
3. Audit duplicate and malformed `verifications` rows, scalar types, stale role IDs, rename templates, external writers, and current indexes.
4. Preserve data as-is. Do not normalize, deduplicate, backfill, or index merely to enable Go.
5. Pair runtime/sync flags; run preflight and command-sync dry-run before reviewed apply.
6. Keep a single Go process for flow smoke unless a shared challenge store is implemented.
7. Confirm the target role is below the bot role and the bot can change nicknames where tested.

## Parity Tests

Focused coverage locks metadata, public/runtime permissions, exact setup/prompt/modal/error UI, visibility, answer/nickname quirks, role-before-owner-error ordering, captcha shape, strict legacy IDs, challenge ownership/TTL/atomic claims, scalar reads, typed duplicate-safe writes, usage ownership, app wiring, gates, command sync, and preflight. Run:

```bash
go test ./internal/core/domain ./internal/core/services/onboarding ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/discord/customid ./internal/discord/features/onboarding ./internal/app ./internal/config ./internal/parity ./cmd/mhcat-command-sync ./cmd/mhcat-staging-preflight
go test -race ./internal/core/services/onboarding ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/discord/customid ./internal/discord/features/onboarding ./internal/app
go vet ./internal/core/services/onboarding ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/discord/customid ./internal/discord/features/onboarding ./internal/app
```

## Staging Smoke

1. Review the read-only `verifications` audit and confirm one active runtime owner.
2. Run preflight, command-sync dry-run, reviewed guild apply, and runtime startup.
3. Confirm both commands are publicly discoverable; test setup Manage Messages denial and success.
4. Verify exact setup UI, absent/raw/all-space rename values, hierarchy rejection, typed writes, duplicate updates, and one usage event per attempt.
5. Test flow missing config, missing role, and hierarchy errors.
6. Verify `captcha.jpeg`, button/modal text, ephemeral prompt/button errors, and public modal completion response.
7. Submit wrong/case/whitespace answers, then a correct answer; verify exact error/success titles and one role add.
8. Test no rename, first-token rename, replacement-string usernames, nickname failure retry, and owner role-before-error behavior.
9. Re-submit after versioned success and verify rejection; test foreign user/guild state, five-minute expiry, and restart invalidation.
10. Exercise one existing legacy button/modal ID only in disposable staging and confirm compatibility/reusability.
11. Disable gates, remove only managed staging commands, preserve Mongo data, and perform rollback checks.

## Rollback

1. Disable both command-sync includes and remove only their managed staging commands.
2. Disable config/flow runtime gates and stop every Go process handling their interactions.
3. Preserve `verifications`; typed Go writes remain Mongoose-readable. Do not repair data or indexes during emergency rollback.
4. Restore Node setup and/or flow only after confirming no Go owner remains.
5. Verify one setup update and one complete captcha/role/nickname flow in staging.
6. Review any ownership overlap for duplicate config writes or member side effects.

Production multi-process/sharded flow remains blocked on a bounded shared challenge store or verified sticky interaction routing. Single-process staging remains supported.
