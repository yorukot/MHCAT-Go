# Ticket Setup/Delete UI Parity Slice

Status: historical slice note, superseded by the canonical [ticket parity contract](74-ticket.md). Legacy source was not modified.

## Scope

This slice implements the low side-effect part of the legacy private-channel ticket system:

- Slash command definitions for `私人頻道設置` and `私人頻道刪除`.
- Legacy command option names:
  - `類別`
  - `管理員身分組`
- Legacy permission intent through default member permission `ManageMessages` (`8192`).
- Setup modal with title `私人頻道系統!`.
- Modal fields:
  - `ticketcolor` / `請輸入嵌入顏色`
  - `tickettitle` / `請輸入標題`
  - `ticketcontent` / `請輸入內文`
- Ticket panel embed and legacy open button custom ID `tic`, sent as a channel message after modal submit.
- Delete success/failure embed text matching legacy messages.
- Legacy modal submit route `nal` + first field `ticketcolor`, for live setup modals created before the Go versioned custom ID rollout.
- Runtime Manage Messages check and duplicate-config error matching the legacy command.

## Intentional Bug Fix

Legacy `slashCommands/私人頻道/ticket.js` writes the `tickets` config before the modal is submitted and before the color/title/content fields are validated.

The Go implementation only persists `domain.TicketConfig` after modal submit succeeds and the color/title/content fields pass validation. Tests verify that invalid color input edits the deferred reply with the legacy error message and does not save config. This remains an intentional behavior fix; live legacy `nal` modal submits are still accepted for compatibility.

## Runtime Wiring

Ticket setup/delete routes are registered when `MHCAT_FEATURE_TICKETS_ENABLED=true` in the default app runtime, or when tests call `app.BuildRuntime` with an explicit `TicketConfigRepository`.

The default value is false, so ticket Mongo writes and Discord side effects remain opt-in. Bot startup still does not sync slash commands.

The modal submit path requires a `DiscordMessagePort` because legacy behavior sends the ticket panel to the current channel and edits the modal reply to a green success embed. If side-effect ports are absent, setup can still show the modal in tests, but modal submission returns a side-effect configuration error instead of silently faking the panel.

## Command Sync

Ticket command definitions exist in `internal/discord/features/ticket` and can be included in the staging guild registry only with `MHCAT_COMMAND_SYNC_INCLUDE_TICKETS=true` plus the paired runtime gate.

Before syncing ticket commands to a staging guild:

1. Stop all Node and extra Go ticket owners.
2. Audit duplicate/malformed `tickets` rows and stale Discord IDs.
3. Enable the paired runtime and staging command-sync flags.
4. Run command sync dry-run and apply only reviewed guild-scoped commands.

## Remaining Work

- Run the canonical staging smoke for setup modal, stale submit, panel compensation, `tic`, welcome compensation, and `del`.
- Test only exact 6-digit hash-prefixed hex and supported case-sensitive Discord color names; the broader CSS/3-digit validator set does not survive discord.js `setColor`.

## Tests

Added tests cover:

- Command definition validation and ownership metadata.
- Route registration.
- Setup modal shape.
- No config write before modal submit.
- Valid versioned modal submit saves `tickets` config, sends the panel to the channel, and edits the deferred reply with `<a:green_tick:994529015652163614> | 成功創建私人頻道`.
- Legacy `nal` modal submit sends the same panel and success edit without needing the versioned setup payload.
- Invalid color edits the legacy error and does not save.
- Successful legacy color inputs are the validator/discord.js intersection: exact `#RRGGBB` and supported case-sensitive Discord names such as `Aqua`; lowercase/CSS-only names and 3-digit hex are rejected.
- Missing Manage Messages returns the legacy permission-denied embed.
- Existing config returns the legacy duplicate setup embed and does not overwrite the config.
- Delete success and missing-config legacy messages.
- DiscordGo modal response conversion.
- Optional app runtime ticket route wiring.
