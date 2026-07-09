# Lottery Disabled Command Parity

## Scope

This slice ports only the current visible behavior of legacy `/抽獎設置`.

Legacy evidence:

- `MHCAT/slashCommands/抽獎系統/lotter_create.js` defines the full slash command metadata and options.
- The handler immediately runs `deferReply({ ephemeral: true })`, edits the original reply with one green embed, and returns.
- The permission check, date parsing, `lotter` document creation, public lottery message, and buttons are unreachable in the current legacy behavior.

## Implemented Behavior

- Runtime flag: `MHCAT_FEATURE_LOTTERY_DISABLED_COMMAND_ENABLED=false` by default.
- Command-sync flag: `MHCAT_COMMAND_SYNC_INCLUDE_LOTTERY_DISABLED_COMMAND=false` by default.
- Command name: `抽獎設置`.
- Options are preserved in the legacy order.
- The command defers ephemerally and edits the original reply with:

```txt
<a:green_tick:994529015652163614> | 這個指令暫時無法使用造成困擾非常抱歉!
```

- The embed color preserves legacy `client.color.greate` (`#53FF53`).
- The command definition keeps Manage Messages metadata (`8192`), but the runtime handler intentionally does not enforce it because the legacy disabled return occurs before the permission check.

## Not Implemented

- No `lotters` reads or writes.
- No lottery creation.
- No public lottery panel send.
- No `點我參加抽獎!` or `誰參加抽獎` buttons.
- No legacy `lotter*` component routes (`enter`, `search`, `restart`, `stop`) beyond parser recognition.
- No scheduler or auto-winner path.

## Safety

- Bot startup still does not register commands.
- Command sync includes this command only in staging guild scope with the explicit include flag.
- Staging preflight and scripts reject unpaired command-sync/runtime flags.
- No Mongo feature write or index creation is introduced.

## Tests

- Definition metadata and option order.
- Handler ephemeral defer and legacy unavailable embed.
- Handler does not require runtime Manage Messages permission for the disabled response.
- Runtime route is unavailable by default and available only with explicit runtime option.
- Command-sync dry-run includes `抽獎設置` only with explicit include flag.
- Staging preflight rejects unpaired sync/runtime flags.
