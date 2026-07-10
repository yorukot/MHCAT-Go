# Lottery Disabled Command and Existing Components

## Scope

Legacy `/抽獎設置` immediately returns an ephemeral unavailable embed, so Go keeps lottery creation and public panel generation disabled. Legacy `events/btn.js` still handles buttons on previously created lottery messages, so Go restores those existing-message actions behind a separate runtime gate.

Legacy evidence:

- `MHCAT/slashCommands/抽獎系統/lotter_create.js` defines the option-rich command but returns before permission checks, date parsing, `lotters` creation, or panel sends.
- `MHCAT/events/btn.js` still routes numeric `<id>lotter`, `search`, `restart`, and `stop` button IDs.
- `MHCAT/models/lotter.js` stores string numeric fields, nullable role/cap fields, loose `member` entries, and `end`.

## Disabled Command

- Runtime flag: `MHCAT_FEATURE_LOTTERY_DISABLED_COMMAND_ENABLED=false`.
- Command-sync flag: `MHCAT_COMMAND_SYNC_INCLUDE_LOTTERY_DISABLED_COMMAND=false`.
- The command defers ephemerally and returns the legacy green unavailable embed.
- Runtime does not enforce Manage Messages because the legacy unavailable return occurs before that check.
- It performs no `lotters` read/write and sends no public panel.

## Existing Components

- Runtime flag: `MHCAT_FEATURE_LOTTERY_COMPONENTS_ENABLED=false`.
- Gateway is required; no command-sync flag is involved.
- Accepted legacy IDs are limited to `<13-20 digits>lotter` plus optional `search`, `restart`, or `stop` suffixes.
- Enter validates end time, duplicate entry, capacity, required role, and forbidden role, then atomically appends `{id,time}`.
- Search returns the legacy participant-count embed, `discord.txt`, self-entry status, and manager controls.
- Reroll sends one legacy-style winner message to `message_channel` and sets `end:true`.
- Stop sets `end:true` and returns the legacy success embed.

Legacy reroll attempted one progressively larger channel send per winner from inside its draw loop. Go emits one aggregate winner message and caps an oversized stored winner count at 50 to avoid duplicate pings and Discord message-limit failures while preserving the final visible winner layout.

Destructive reroll/stop actions recheck authorization. Rows with `owner` require that owner or the guild owner; legacy rows without `owner` preserve the Manage Messages fallback. This intentionally closes the legacy custom-ID spoofing gap.

## Compatibility

Go reads mixed string/number/null fields and skips malformed participant entries. It writes only existing fields:

- `member`: appends `{id:<user>,time:<Unix milliseconds>}` with atomic duplicate/date/cap guards.
- `end`: sets boolean `true`.

No indexes, schema fields, migrations, or creation rows are added. Disabling the component gate returns ownership to Node without repair.

## Staging

Use disposable copied `lotters` rows and channels only. Verify enter, duplicate/cap/role errors, participant export, unauthorized stop/reroll denial, owner stop, and reroll winner sends. Do not enable Node and Go lottery button ownership for the same guild at the same time.

## Tests

- Legacy parser golden cases for all four IDs.
- Mixed BSON decoding and guarded Mongo update shape.
- Service entry and manager authorization rules.
- Exact handler UI/export, reroll send, and end-state writes.
- Independent command/component runtime wiring and safe-default config.
- Gateway readiness in config and staging preflight.
