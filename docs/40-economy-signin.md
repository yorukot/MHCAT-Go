# Economy Sign-In Slice

Status: first gated write slice for legacy `/簽到`.

## Implemented

- Local slash command definition for `簽到`.
- Runtime handler behind `MHCAT_FEATURE_ECONOMY_SIGNIN_ENABLED=false` by default.
- Command-sync include gate `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SIGNIN=false` by default.
- Staging preflight and shell-script pairing checks.
- Legacy public loading embed:
  - author `正在努力為您尋找資料!`
  - loading GIF URL from legacy
  - footer `MHCAT 帶給你最好的discord體驗!`
  - color `#FF5809`
- Final edit to `sign.png` with no embed and four secondary navigation buttons.
- New navigation custom IDs use `mhcat:v1:economy:sign_page:<payload>`.
- Legacy sign navigation IDs like `/<user>_sing{2026}-[7]` remain accepted by the typed parser/router.
- Mongo repository writes only `coins` and `sign_lists`.
- Calendar updates use `$addToSet` instead of the legacy duplicate-prone full object replacement.

## Intentional Legacy Bug Fixes

- The button redraw Friday label keeps `Fri.` instead of copying the legacy `Fir.` typo.
- Sixth calendar rows are rendered when the month needs them.
- The handler routes by typed `RouteKey`; it does not use legacy `customId.includes("sing")`.
- The sign-list day write is attempted only after a successful coin award.
- New versioned custom IDs are bounded by Discord's 100-character limit.

## Production Blockers

- Production duplicate audit must be clean for `coins` logical key `{guild, member}`.
- Production duplicate audit must be clean for `sign_lists` logical key `{guild, member}`.
- A unique-index plan for both logical keys must be approved and applied through the explicit Mongo index process.
- Daily-mode sign-in depends on the Asia/Taipei midnight reset. The Go refactor now has a dry-run-first one-shot reset tool, but production sign-in still needs an operator-owned reset process or a future lease-backed scheduler.
- Coin award and calendar update are not yet wrapped in a Mongo transaction. The calendar write is idempotent and repairable, but production enablement needs either transaction support or a documented repair/audit path.

## Staging Use

Use only with an isolated staging database:

```bash
export MHCAT_FEATURE_ECONOMY_SIGNIN_ENABLED=true
export MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SIGNIN=true
go run ./cmd/mhcat-staging-preflight --format text
scripts/staging/command-sync-dry-run.sh
```

Do not run command-sync apply until the dry-run plan is reviewed.

## Duplicate Audit

Run the read-only audit before any production sign-in enablement:

```bash
go run ./cmd/mhcat-mongo-audit --format text
```

Expected production-ready result:

- no `duplicate_key_risk collection=coins key=coins_guild_member`;
- no `duplicate_key_risk collection=sign_lists key=sign_lists_guild_member`;
- no duplicate-key warnings for those collections in JSON output.

If duplicate risks are present, do not apply unique indexes and do not enable production sign-in. Create a separate dry-run repair plan first.

## Daily Reset

The one-shot reset tool is documented in `docs/41-economy-daily-reset.md`.

Preview only:

```bash
go run ./cmd/mhcat-economy-reset --dry-run
```

Apply requires explicit approval:

```bash
MHCAT_JOBS_DAILY_RESET_ENABLED=true \
go run ./cmd/mhcat-economy-reset --apply
```

Do not wire this into bot startup until scheduler ownership/lease design is implemented.
