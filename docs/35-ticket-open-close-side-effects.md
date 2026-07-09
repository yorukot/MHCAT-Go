# Ticket Open/Close Side-Effect Slice

Status: implemented behind explicit side-effect wiring. Legacy source was not modified.

## Scope

This slice implements the legacy ticket component actions:

- `tic` opens a private ticket text channel.
- `del` deletes the current ticket channel when the actor is allowed to close it.
- Duplicate ticket channels are detected by guild/channel name before creation.
- Deleted setup config returns the legacy warning text and deletes the stale panel message when the source message ID is available.
- The created channel receives the legacy welcome embed and `del` button.
- The opener receives the legacy ephemeral success embed.

## Legacy UI Preserved

Open success:

- Embed title: `__**頻道**__`
- Description: `:white_check_mark: 你成功開啟了頻道!`
- Color: green
- Ephemeral: yes

Duplicate ticket:

- Embed title: `__**客服頻道**__`
- Description: `:warning: 你已經有一個客服頻道了!`
- Color: red
- Ephemeral: yes

Ticket channel welcome:

- Content: `||@everyone||`
- Embed title: `__**私人頻道**__`
- Description: `你開啟了一個私人頻道，請等待客服人員的回復!`
- Button custom ID: `del`
- Button label: `🗑️ 刪除!`
- Button style: danger

Close denied:

- Embed title: `__**私人頻道**__`
- Description: `你開啟了一個私人頻道，請靜候客服人員的回復!`
- Color: red

## Safety Changes

The welcome message preserves the visible `||@everyone||` text but suppresses allowed mentions, so it does not ping everyone.

The legacy close logic depends on fetching recent channel messages and checking a fragile message-author condition. The Go implementation uses a safer rule:

- actor with Manage Messages can close;
- otherwise, the current channel name must match the actor user ID.

This keeps the user-owned ticket close path without depending on mutable message history.

## Runtime Wiring

Ticket open/close routes are enabled by default runtime only when:

```txt
MHCAT_FEATURE_TICKETS_ENABLED=true
```

They require:

- `TicketConfigRepository`
- `DiscordChannelPort`
- `DiscordMessagePort`
- optional bot user ID for a bot permission overwrite

Default `cmd/mhcat-bot` startup keeps this flag false. Ticket commands should not be synced to a staging guild unless the same environment has ticket runtime enabled.

## Tests

Added tests cover:

- `tic` duplicate channel warning.
- `tic` missing config stale-panel deletion.
- `tic` channel creation request and permission overwrites.
- welcome message embed/button shape.
- suppressed allowed mentions.
- open success reply shape.
- `del` owner close path.
- `del` Manage Messages close path.
- `del` denied path.
- runtime dispatcher route for legacy `tic` when side effects are explicitly provided.
- DiscordGo outbound embed/button conversion.
- DiscordGo source message ID and permission-bit extraction.

## Remaining Work

- Add a staging command-sync allowlist for ticket commands only when runtime side effects are enabled in that environment.
- Run a staging guild smoke test for setup, modal submit, `tic`, and `del`.
- Decide whether full legacy message-history close semantics must be emulated or the safer permission rule is accepted as an intentional behavior fix.
