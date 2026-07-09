# Poll Wave A: Create / Vote / Owner Menu

## Scope

Implemented the first poll slice from legacy:

- slash command definition and handler for `投票創建`;
- legacy-style public poll embed, choice buttons, result button, and owner select menu;
- Mongo-compatible `polls` document shape with legacy fields;
- repository contract for create/get/vote/toggle/max-choice;
- component handlers for vote, result text, owner toggles, and max-choice select;
- runtime feature gate `MHCAT_FEATURE_POLLS_ENABLED=false` by default;
- staging command-sync include gate `MHCAT_COMMAND_SYNC_INCLUDE_POLLS=false` by default;
- staging preflight/script pairing checks for command sync and runtime enablement.

## Legacy UI Preserved

Visible poll UI follows `slashCommands/管理系統/poll.js` and `events/poll.js`:

- title format: `<:poll:1023968837965709312> | 投票\n<question>`;
- vote line: `<:vote:1023969411369025576> **總投票人數:...**`;
- YellowSmallDot setting lines;
- choice buttons are secondary and labeled with option text;
- result button label is `查看投票結果` with `<:analysis:1023965999357243432>`;
- owner select placeholder is `🔧投票發起人操作`;
- owner select values preserve legacy values such as `poll_public_result` and `poll_end_poll`;
- success and error messages preserve legacy text where implemented.

## Intentional Internal Changes

- New Go-generated vote buttons use versioned IDs like `mhcat:v1:poll:vote:i=0` instead of embedding raw option text.
- Legacy `poll_<choice>` still parses for old live messages.
- Legacy `menu_choose` depended on an in-memory discord.js collector closure. New max-choice selects use `mhcat:v1:poll:max_choices:m=<messageID>` so they survive restarts.
- Vote add/remove is handled by repository methods rather than handler-level array replacement. The Mongo repository uses conditional `$push`/`$pull` filters to reduce lost updates.
- The custom ID parser now counts legacy custom ID length by characters and accepts the legacy 80-character poll choice limit.

## Not Implemented Yet

- Production poll command sync.
- Staging live smoke for `/投票創建`, voting, and owner menu controls.
- Duplicate-key audit before any unique index on `{guild, messageid}`.

## Tests Added

- Poll BSON decode/round-trip tests.
- Poll repository contract tests with fake repository.
- Poll handler UI/validation/vote/menu/result tests.
- DiscordGo side-effect conversion test for select menus and option emojis.
- Custom ID parser test for 80-character legacy poll choices.
- Config and command-sync gate tests.
- Runtime wiring tests showing poll routes are unavailable unless dependencies are explicitly provided.
- Staging preflight poll pairing tests.

## Next Step

Poll Wave B added legacy result chart image, `discord.txt`, and `poll_info.xlsx` export. The next step is staging smoke for the full poll flow before enabling poll commands outside staging.
