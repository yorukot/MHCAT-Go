# Economy Sign-In Slice

Status: gated write slice for legacy `/簽到` plus read-only legacy `/簽到列表`.

## Implemented

- Local slash command definition for `簽到`.
- Local slash command definition for `簽到列表`.
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
- `/簽到列表` defers publicly, reads `coins` by guild and `gift_changes.time`, and edits to the legacy `簽到人數資訊` embed with `discord.txt`.
- `/簽到列表` preserves the legacy daily marker mode (`coin.today == 1`) and rolling cooldown mode (`now - coin.today < gift_changes.time` and `> 0`), including Mongoose-visible missing/null `time` falling back to 86400 seconds and fractional, negative, or infinite numeric values retaining JavaScript comparison behavior.
- `/簽到列表` does not write to Mongo.

## Intentional Legacy Bug Fixes

- The button redraw Friday label keeps `Fri.` instead of copying the legacy `Fir.` typo.
- Sixth calendar rows are rendered when the month needs them.
- The handler routes by typed `RouteKey`; it does not use legacy `customId.includes("sing")`.
- The sign-list day write is attempted only after a successful coin award.
- New versioned custom IDs are bounded by Discord's 100-character limit.
- `/簽到列表` preserves legacy `username#discriminator` member labels, including `#0` for migrated Discord accounts. If Discord omits the discriminator entirely, it falls back to the username; missing lookups still render the legacy `使用者已消失!` fallback.

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

Smoke `/簽到列表` in both modes if possible:

- daily mode: ensure `gift_changes.time` is missing or `0`, set at least one staging `coins.today=1`, and verify the count, `有`/`沒有` text, inline names, and `discord.txt`;
- rolling mode: set `gift_changes.time` to a positive cooldown, use staging `coins.today` Unix timestamps, and verify the exported `簽到時間:` values use Taipei time.

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

The lease-gated one-shot and recurring reset paths are documented in `docs/41-economy-daily-reset.md`.

Preview only:

```bash
go run ./cmd/mhcat-economy-reset --dry-run
```

Apply requires explicit approval:

```bash
MHCAT_JOBS_DAILY_RESET_ENABLED=true \
MHCAT_SCHEDULER_LEASE_ENABLED=true \
MHCAT_SCHEDULER_LEASE_OWNER=staging-reset-cli \
go run ./cmd/mhcat-economy-reset --apply
```

The recurring worker has a separate `MHCAT_FEATURE_DAILY_RESET_SCHEDULER_ENABLED=true` gate. Keep Node `handler/cron.js` and every Go reset writer under exclusive ownership before production sign-in rollout.
