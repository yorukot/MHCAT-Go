# Command Registration And Slash Dispatch Contract

Status: parity/operations-audited against `handler/slash_commands.js`, `events/SlashCommands.js`, all 74 active legacy slash definitions, current typed definitions, static parity audit, command-sync planner/client/CLI, staging guards, interaction adapter/router/middleware/dispatcher, usage repository, responder lifecycle, app wiring, and race coverage. Live Discord registration and Gateway smoke remain required.

## Definitions And Registration

The static audit reports 74 legacy files, 74 unique legacy command names, 74 Go definitions, 74 exact metadata matches, zero missing/extra names, and zero parse warnings. Names, descriptions, localization fields, option type/order/text/requirements/choices, and default permission metadata are feature-owned and covered by the parity audit plus feature contracts.

Legacy creates every global command from every shard after ready and later deletes unknown commands. Go intentionally removes that startup race. `mhcat-bot` never mutates application commands. `mhcat-command-sync` fetches remote state, validates the typed registry, computes deterministic create/update/delete/no-op plans, and defaults to dry-run. Writes require explicit `--apply`; deletion and bulk overwrite require separate acknowledgements.

Staging apply is guild-scoped, feature definitions require paired runtime/include gates, and preflight/scripts reject unpaired or unsafe configuration. Global production apply remains an explicit operator step. Rollback must use a reviewed sync plan; bot restart alone never changes command registration.

## Slash Dispatch

Discord interactions are adapted into typed commands/options/actors and routed by exact slash name. Feature handlers own legacy permission checks, defer/reply/update/modal lifecycles, UI, and docs links. Metadata cooldown values remain descriptive because legacy allocated a cooldown map but did not enforce it; Go adds no hidden throttle.

An unknown or renamed slash command receives the exact legacy green fallback embed, animated error emoji, help footer text, and requester avatar. The dispatcher still returns `ErrRouteNotFound` for observability and feature-gate tests without sending a second generic error. Unknown components/modals continue through controlled parser/router errors.

The runtime tracks whether an interaction has already replied, deferred, updated, or opened a modal and sends a generic controlled error only when a failed handler has not responded. Timeouts, panic recovery, structured route logging, and context cancellation are safety improvements. Error logging/redaction must not expose tokens, webhook URLs, Mongo credentials, or raw backend details to Discord.

## Usage Ownership

Legacy attempts an `all_use_count` write for every interaction before checking its type, effectively intending one count per slash attempt but risking undefined component rows and stale read/write loss. Go's optional global middleware records exactly one best-effort `all_use_counts` increment for non-empty slash command names before route lookup, permissions, and handler work. Components/modals record none, and feature handlers do not double count.

Usage tracking is disabled by default. Enabling it requires the explicit feature gate and Mongo repository. Tracking failure never blocks command handling, matching legacy best-effort behavior while using an atomic increment.

## Verification

```bash
go run ./tools/parity-audit --legacy-root ../MHCAT --format markdown
go test ./internal/parity ./internal/discord/commands \
  ./internal/discord/interactions ./internal/discord/runtime \
  ./internal/adapters/discordgo ./internal/adapters/usage \
  ./cmd/mhcat-command-sync ./cmd/mhcat-staging-preflight ./internal/app
go test -race ./internal/discord/interactions ./internal/discord/runtime ./internal/app
go vet ./...
```

Coverage includes registry validation, diff/no-op/create/update/delete plans, dry-run no-write, apply acknowledgements, strict/staging scope, every feature include/runtime pair, option adaptation, exact routing, usage-before-handler/permission, no component usage, panic/timeout/cancellation, duplicate-response prevention, unknown fallback UI, and app registration.

## Staging And Rollback

1. Run parity audit, preflight, and command-sync dry-run. Review every planned command and ensure only intended feature/runtime pairs are enabled.
2. Apply only to an isolated staging guild. Verify create/update/no-op behavior, exact visible command metadata, one success route, one permission error, unknown fallback, one usage increment per slash, and no component increment.
3. Start Gateway separately and confirm bot startup performs no command REST mutation.
4. Before production, capture the existing remote command inventory and a rollback sync plan.

Rollback by disabling feature runtime/sync flags and applying the reviewed command plan before or after gracefully stopping Go as ownership requires. Do not restore legacy per-shard registration or allow Node and Go registration owners to race.

Production remains gated on live guild sync/Gateway smoke, reviewed global scope, exclusive registration ownership, and per-feature rollout approval.
