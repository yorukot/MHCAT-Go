# Wave 5.8 Help UI Parity Correction

Status: implemented and verified locally.

## Trigger

Manual staging use showed that `/help` still returned the Wave 5.1 placeholder text:

- `MHCAT help`
- `Implemented commands:`
- a plain bullet list

That did not preserve the legacy MHCAT interface. The refactor goal is behavior parity, not a redesigned help menu.

## Legacy References

Read-only legacy files used:

- `MHCAT/slashCommands/實用工具/help.js`
- `MHCAT/functions/menu.js`
- `MHCAT/events/btn.js`
- `MHCAT/config.json`

## Runtime Behavior Changed

`/help` now uses the legacy-style response shape:

- deferred interaction response;
- MHCAT author embed with the original icon/invite URL;
- original Chinese overview copy;
- legacy command category fields;
- legacy category select menu custom ID `helphelphelphelpmenu`;
- original invite/support/website link buttons;
- `/help 指令名稱:<command>` command detail embed;
- unknown command red embed;
- legacy help select-menu category response, ephemeral.

The staging slash command definitions do not need to be re-applied for this change because the command name/options did not change.

## Files Updated

- `internal/discord/responses/responder.go`
- `internal/adapters/discordgo/responder.go`
- `internal/adapters/discordgo/interaction_adapter.go`
- `internal/discord/interactions/interaction.go`
- `internal/discord/features/utility/legacy_help.go`
- `internal/discord/features/utility/help_handler.go`
- `internal/discord/features/utility/module.go`
- `internal/discord/features/utility/help_handler_test.go`
- `internal/app/wiring_test.go`
- `internal/testutil/fakeinteractions/components.go`
- `README.md`
- `docs/10-feature-parity-checklist.md`
- `docs/20-wave-5.1-notes.md`
- `docs/27-wave-5.8-help-ui-parity.md`

## What Is Not Implemented

The help UI can list legacy categories and command documentation entries before every feature handler is implemented. That matches legacy help behavior and does not mean these features are runtime-ready:

- ticket;
- poll/vote;
- verification;
- role buttons;
- coin/economy;
- XP;
- gacha/work/lottery;
- cron;
- moderation;
- ChatGPT/external API;
- full `info user/guild/shard`;
- translate;
- auto-chat.

## Tests Added or Updated

- Help handler overview now asserts embed/select/link-button output instead of placeholder text.
- Help command detail now asserts the legacy `指令資料` embed fields.
- Unknown help command now asserts the legacy red invalid-command embed.
- Legacy `helphelphelphelpmenu` select route is registered and tested through the parsed route key.
- App runtime wiring tests now assert legacy help embeds.
- DiscordGo responder tests compile with embed/component conversion.

## Commands Run

- `go test ./internal/discord/features/utility`
- `go test ./internal/adapters/discordgo`
- `go test ./internal/discord/responses`
- `go test ./...`

## Operational Note

Restart the staging bot process after deploying this change. No command sync apply is required for this help UI fix.
