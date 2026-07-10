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
- Enter preserves legacy error precedence: duplicate entry (with the ended-row override), capacity, end time, required role, then forbidden role. Negative stored capacities are immediately full. A successful entry atomically appends `{id,time}`.
- Search returns the legacy participant-count embed, self-entry status, manager controls, and `discord.txt`. It shows names through 99 participants and switches to the legacy file warning at 100 while retaining every export row.
- Search preserves `username#discriminator`, including `#0` for migrated Discord usernames, the two distinct missing-user labels, and Node 20 `zh-TW` Taipei timestamps with the U+2009 separator and `24:xx:xx` midnight hour.
- Reroll draws with replacement, preserves exact stored prize whitespace, uses the bot member's guild display color, sends literal `<@>` when nobody entered, and sets `end:true` after the channel send.
- A nonpositive stored winner count preserves the legacy deferred no-op: no draw, channel send, end write, or interaction edit.
- Stop sets `end:true` and returns the legacy success embed.

Legacy reroll attempted one progressively larger channel send per winner from inside its draw loop. Go emits one aggregate winner message and caps an oversized stored winner count at 50 to avoid duplicate pings and Discord message-limit failures while preserving the final visible winner layout.

Destructive reroll/stop actions recheck authorization. Rows with `owner` require that owner or the guild owner; legacy rows without `owner` preserve the Manage Messages fallback. This intentionally closes the legacy custom-ID spoofing gap.

Other intentional safety differences are limited to generated numeric component IDs, exact participant-ID comparison instead of legacy substring matching, unconditional `end:true` entry rejection, malformed participant skipping, atomic entry guards, and explicit winner-mention allowlisting. These differences prevent forged routes, false duplicate matches, post-stop joins, malformed-row failures, concurrent over-entry, and unrelated pings without changing valid-row UI.

## Compatibility

Go reads mixed string/number/null fields, preserves prize text verbatim, and skips malformed participant entries. It writes only existing fields:

- `member`: appends `{id:<user>,time:<Unix milliseconds>}` with atomic duplicate/date/cap guards.
- `end`: sets boolean `true`.

No indexes, schema fields, migrations, or creation rows are added. Disabling the component gate returns ownership to Node without repair.

## Staging

Use disposable copied `lotters` rows and channels only. Verify enter, duplicate/cap/role errors, participant export, unauthorized stop/reroll denial, owner stop, and reroll winner sends. Do not enable Node and Go lottery button ownership for the same guild at the same time.

## Tests

- Legacy parser golden cases for all four IDs.
- Mixed BSON decoding and guarded Mongo update shape.
- Service entry precedence, negative capacity, and manager authorization rules.
- Exact handler UI/export, 0/99/100 participant boundaries, Node 20 timestamp fixtures, reroll payload/color/no-op/cap behavior, and end-state writes.
- Independent command/component runtime wiring and safe-default config.
- Gateway readiness in config and staging preflight.
