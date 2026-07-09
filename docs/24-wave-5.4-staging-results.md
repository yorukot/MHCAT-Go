# Wave 5.4 Staging Results

Status: staging guild command apply and gateway smoke completed; manual Discord interaction checks remain pending.

## Goal

Validate the Wave 5.3 staging-only workflow for the current utility slice:

- guild-scoped command sync dry-run;
- optional guild-scoped command sync apply;
- gateway ready smoke;
- manual `/ping`, `/help`, `/help 指令名稱:ping`, and `/info bot` checks;
- no bot-startup command registration;
- no global command mutation;
- no command deletion;
- no bulk overwrite;
- no Mongo feature writes or index creation;
- no Message Content intent.

## Preflight Result

- Legacy source status: clean. `MHCAT/` reports `## main...origin/main`.
- Required staging env probe: failed without printing values.
- Missing live prerequisites: staging Discord token, staging application ID, staging guild ID, staging Mongo URI, or staging Mongo database were not all present in the process environment.
- Message Content check: no real gateway run was attempted; default remains disabled in config.
- Real staging target confirmation: not possible without staging env.

## Wave 5.6 Attempt Result

Attempt date: 2026-07-04 Asia/Taipei.

Environment presence check:

- `MHCAT_DISCORD_TOKEN`: missing.
- `MHCAT_DISCORD_APPLICATION_ID`: missing.
- `MHCAT_STAGING_GUILD_ID`: missing.
- `MHCAT_MONGODB_URI`: missing.
- `MHCAT_MONGODB_DATABASE`: missing.

Command run:

```bash
go run ./cmd/mhcat-staging-preflight --format text
```

Result: exited non-zero before any Discord or Mongo action. The report confirmed command deletion, bulk overwrite, and Message Content were disabled, and reported missing required staging values. No raw secret or Mongo URI was printed.

Live staging actions were intentionally skipped:

- no command sync dry-run, because preflight did not pass;
- no command sync apply;
- no gateway smoke;
- no manual interaction smoke.

## Wave 5.6 Follow-up Result

Attempt date: 2026-07-04 Asia/Taipei.

Additional findings:

- Local `.env` now has Discord token, application ID, staging guild ID, and command sync safety variables.
- `MHCAT_MONGODB_URI` remains empty locally, so full staging preflight still fails before gateway smoke.
- A local Mongo Compose file was added for host-side smoke, but Docker is not running in this environment:
  - active Docker context: OrbStack;
  - missing socket: `/Users/yorukot/.orbstack/run/docker.sock`;
  - default Docker socket also unavailable.

Command sync dry-run:

```bash
scripts/staging/command-sync-dry-run.sh
```

Result: passed after network access was allowed. The dry-run planned only low-risk guild-scoped creates for managed commands:

- `help`
- `info`
- `ping`

No Discord command writes were performed. No command deletion, bulk overwrite, or global command mutation happened.

Command sync apply: not run. Reason: full staging preflight still fails because no local Mongo URI is available, and the dry-run result has not been followed by an explicit staging apply step.

Gateway smoke: not run. Reason: local Mongo is unavailable until Docker/OrbStack is started or a safe staging Mongo URI is configured.

Production Mongo read-only audit:

- Completed through `windows-vm` using the remote Mongo URI from `/root/mhcat/.env` without printing it.
- Output included collection names, estimated counts, indexes, and sampled field/type shapes only.
- No document values, writes, repairs, backfills, or index creation were performed.
- Sanitized findings are recorded in `docs/26-production-mongo-readonly-audit.md`.

## Wave 5.6 Compose Default Update

Attempt date: 2026-07-04 Asia/Taipei.

Local `.env` was changed to use the local Compose Mongo default instead of the production URI:

```txt
MHCAT_MONGODB_URI=mongodb://127.0.0.1:27018/mhcat-database?directConnection=true
MHCAT_MONGODB_DATABASE=mhcat-database
```

`compose.yml` now provides the local Mongo service. The service was started successfully after OrbStack/Docker access was available:

- container: `mhcat-refactor-mongodb`;
- exposed host address: `127.0.0.1:27018`;
- health: healthy.

Verification:

- staging preflight passed with `.env` sourced;
- `mhcat-mongo-audit` connected read-only to local Compose Mongo and reported missing expected catalog collections because the local database is empty;
- `mhcat-bot` initialized successfully with gateway disabled and local Mongo ping;
- `mhcat-mongo-index --dry-run` connected and created no indexes.

No production Mongo write, local Mongo feature write, index creation, command registration, command deletion, or bulk overwrite happened.

## Wave 5.7 Staging Apply and Gateway Smoke

Attempt date: 2026-07-04 Asia/Taipei.

Prerequisites:

- Local Compose Mongo was healthy.
- Staging preflight passed with `.env` sourced.
- Message Content intent remained disabled.
- Command deletion and bulk overwrite remained disabled.
- Command sync scope remained guild-only.

Initial command sync dry-run:

- planned low-risk creates for managed commands only:
  - `help`;
  - `info`;
  - `ping`.
- no Discord writes were performed during dry-run.
- no global command mutation, delete, or bulk overwrite was planned.

Staging command sync apply:

- ran with `MHCAT_STAGING_MODE=true` and `MHCAT_STAGING_ALLOW_COMMAND_APPLY=true` for the invocation only.
- applied three guild-scoped managed command creates:
  - `help`;
  - `info`;
  - `ping`.
- no delete.
- no bulk overwrite.
- no global command mutation.

Post-apply command sync dry-run:

- `help`: unchanged.
- `info`: unchanged.
- `ping`: unchanged.
- no Discord writes were performed during post-apply dry-run.

Gateway smoke:

- ran with `MHCAT_STAGING_MODE=true` and `MHCAT_STAGING_ALLOW_GATEWAY_SMOKE=true` for the invocation only.
- bot initialized with `gateway_enabled=true`.
- gateway smoke reported ready.
- no command sync or command registration ran from bot startup.
- no Mongo feature write or index creation happened.

Manual interaction smoke:

- pending user action in the staging guild:
  - `/ping`;
  - `/help`;
  - `/help 指令名稱:ping`;
  - `/info bot`.
- record only sanitized pass/fail facts after manual execution.

Follow-up: manual `/help` showed the temporary Wave 5.1 text placeholder instead of the legacy help menu. Wave 5.8 fixes this in runtime code only by restoring the legacy embed/select/link-button help interface. No staging command definition re-apply is required for that fix.

Follow-up: Wave 5.9 similarly restores the legacy `/info bot` embed and `botinfoupdate` refresh button UI in runtime code only. No staging command definition re-apply is required for that fix.

Follow-up: Wave 5.10 adds the legacy `/info shard` subcommand shape. Unlike Wave 5.8/5.9 UI-only fixes, this requires staging command sync dry-run and explicit staging apply before it appears in Discord.

Follow-up: Wave 5.11 adds the legacy `/info user` and `/info guild` subcommand shapes. This also requires staging command sync dry-run and explicit staging apply before those subcommands appear in Discord.

## Command Sync Dry-Run Result

Not run.

Reason: required staging env was absent. Running `scripts/staging/command-sync-dry-run.sh` would fail before contacting Discord because it requires:

- `MHCAT_DISCORD_TOKEN`
- `MHCAT_DISCORD_APPLICATION_ID`
- `MHCAT_STAGING_GUILD_ID`

Expected future command:

```bash
scripts/staging/command-sync-dry-run.sh
```

## Command Sync Apply Result

Not run.

Reason: staging dry-run was not run, required staging env was absent, and apply requires explicit opt-in:

```bash
MHCAT_STAGING_MODE=true \
MHCAT_STAGING_ALLOW_COMMAND_APPLY=true \
scripts/staging/command-sync-apply-guild.sh
```

Expected future behavior:

- guild scope only;
- create/update managed `help`, `ping`, and `info` only;
- no delete;
- no bulk overwrite;
- no global command mutation.

## Gateway Smoke Result

Not run.

Reason: required staging Discord and Mongo env was absent. Gateway smoke requires explicit opt-in:

```bash
MHCAT_STAGING_MODE=true \
MHCAT_STAGING_ALLOW_GATEWAY_SMOKE=true \
scripts/staging/gateway-smoke.sh
```

Expected future behavior:

- Mongo connects and pings;
- Discord Gateway opens;
- ready or timeout is reported;
- `InteractionCreate` handler is registered;
- no command sync or command registration;
- no Mongo feature write;
- clean shutdown.

## Manual Interaction Smoke Result

Not run.

Manual checks remain pending in a staging guild:

- `/ping`
- `/help`
- `/help 指令名稱:ping`
- `/info bot`
- `/info bot` `更新` button after Wave 5.9
- `/info shard` after Wave 5.10 command sync apply
- `/info shard` `更新` button after Wave 5.10 command sync apply
- `/info user` after Wave 5.11 command sync apply
- `/info guild` after Wave 5.11 command sync apply

Record only sanitized pass/fail facts. Do not record tokens, interaction tokens, private guild IDs, private channel IDs, private user content, or raw Mongo URIs.

## Local Verification Result

- `go fmt ./...`: passed.
- `go test ./...`: passed.
- `go vet ./...`: passed.
- `go build ./cmd/mhcat-bot`: passed.
- `go build ./cmd/mhcat-command-sync`: passed.
- `go build ./cmd/mhcat-mongo-audit`: passed.
- `go build ./cmd/mhcat-mongo-index`: passed.
- `make check`: passed.
- `go run ./cmd/mhcat-command-sync` without env: exited non-zero with missing config error; no panic; no secret output.
- `go run ./cmd/mhcat-bot` without env: exited non-zero with missing config error; no panic; no secret output.
- Generated local binaries from `go build` were removed after verification.

## Boundary Result

- Legacy `MHCAT/` source remained unmodified.
- No production/global command sync was run.
- No staging dry-run/apply/gateway smoke was run because staging env was absent.
- No command deletion or bulk overwrite happened.
- `cmd/mhcat-bot` still has no command sync/apply or application command mutation path.
- Gateway remains disabled by default.
- Message Content remains disabled by default.
- Runtime command scope remains low-risk utility only: `help`, `ping`, `info bot`, `info shard`, `info user`, and `info guild`.
- No high-risk feature group was implemented.
- Usage tracking remains no-op.
- No Mongo feature write or index creation happened.
- `internal/core/**` and utility feature services still avoid DiscordGo and MongoDB driver imports.
- No hardcoded secrets, operator IDs, webhooks, raw Mongo URIs, or staging identifiers were added.
- No SQL-style migration directory was added.

## Next Step

Run the manual interaction checks in the staging guild and update this file with sanitized pass/fail results. Do not proceed to Mongo-writing or high-risk feature waves until the runtime interaction smoke is confirmed.
