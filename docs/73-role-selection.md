# Role Selection Parity Contract

Status: reaction-role setup/deletion, button-role setup/application, and reaction add/remove events are parity-audited behind one disabled-by-default runtime ownership gate.

## Legacy References

- Reaction setup command: `MHCAT/slashCommands/管理系統/role.js`
- Reaction delete command: `MHCAT/slashCommands/管理系統/role_delete.js`
- Button setup command: `MHCAT/slashCommands/管理系統/releadd.js`
- Button application runtime: `MHCAT/events/btn.js`
- Button panel modal: `MHCAT/events/modal.js`
- Reaction event runtime: `MHCAT/events/message_reaction.js`
- Models: `MHCAT/models/message_reaction.js` and `MHCAT/models/btn.js`
- Collections: `message_reactions` and `btns`

## Gates And Ownership

Runtime requires the feature, Gateway, and reaction intent gates together:

```bash
MHCAT_FEATURE_ROLE_SELECTION_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_GUILD_MESSAGE_REACTIONS_INTENT=true
```

Publishing the managed commands for staging additionally requires:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_ROLE_SELECTION=true
```

`MHCAT_FEATURE_ROLE_SELECTION_ENABLED` intentionally owns all three setup commands, the `nal` modal, generated add/delete buttons, and reaction add/remove events as one runtime boundary. There are no command-only, button-only, or event-only role-selection feature flags. Do not split these paths between Go processes or leave the corresponding Node interaction/reaction handlers active while Go owns the feature.

The feature does not require the privileged Guild Members intent. Button interactions carry the invoking member's current role IDs. Role writes use Discord REST. Reaction-add payloads carry member identity, while reaction-remove identity is resolved best-effort from Discord state as described below.

Command sync is staging-only and guild-scoped. The three commands remain publicly discoverable because legacy did not register effective default member permissions. Every setup handler enforces Manage Messages (`8192`) at runtime, including the Discord Administrator override.

## Command Definitions

The managed definitions preserve the exact legacy names, descriptions, option order, option types, and required flags:

- `/選取身分組-按鈕`
  - required role `身分組`
- `/選取身分組-表情符號`
  - required string `訊息url`
  - required role `身分組`
  - required string `表情符號`
- `/選取身分組刪除-表情符號`
  - required string `訊息url`
  - required string `表情符號`

No command has default member-permission metadata. Permission denial uses the exact red (`#ED4245`) title ``<a:Discord_AnimatedNo:1015989839809757295> | 你需要有`訊息管理`才能使用此指令``.

All three legacy modules declare `cooldown: 10`, but the active `events/SlashCommands.js` only creates an unused cooldown map and never reads command cooldown metadata. Go likewise applies no feature-local cooldown.

## Reaction Setup Contract

`/選取身分組-表情符號` publicly defers, then preserves this legacy order:

1. Check Manage Messages.
2. Check the configured role exists in cached guild state and is below the bot's highest cached role.
3. Parse the message link.
4. Resolve the target channel from the current guild cache.
5. Validate the emoji.
6. Fetch the message through Discord REST.
7. Add the reaction.
8. Save the reaction-role mapping.
9. Edit the original response with the green success title `<a:green_tick:994529015652163614> | 表情符號選取身分組成功設定`.

Role hierarchy denial therefore takes precedence over an invalid URL or emoji. A missing cached channel and a missing fetched message both render `很抱歉，找不到這個訊息`.

Setup accepts links containing either `https://discord.com/channels/` or the legacy `https://discordapp.com/channels/` prefix. After Go trims surrounding option whitespace, parsing mirrors the JavaScript substring operations rather than imposing a new canonical URL parser. The guild segment from the URL is not used for authorization: after parsing, the channel must exist in the invoking guild's cache, and the stored `guild` remains the invoking guild ID.

Custom emoji mentions are stored by emoji ID and sent to Discord as `name:id`. The ID must exist anywhere in the bot's global emoji cache, matching `client.emojis.cache`; it need not belong to the invoking guild. Unicode validation preserves the legacy unanchored and malformed regular-expression behavior, including broad surrogate-pair acceptance and the accidental `[` match. Plain non-emoji text receives the exact legacy validation error.

## Reaction Delete Contract

`/選取身分組刪除-表情符號` also publicly defers and requires Manage Messages. Unlike setup, it accepts only the `https://discord.com/channels/` host because the legacy delete command did not accept `discordapp.com`. It preserves cached-channel lookup, REST message fetch, and the unusual legacy behavior of adding the supplied reaction before deleting configuration.

The URL guild segment is ignored after the target channel is confirmed in the invoking guild cache. A missing logical mapping renders `這個表情符號沒有在這個訊息上設定`; success renders the green title `表情符號選取身分組成功刪除` without the animated tick.

Go deliberately normalizes a custom emoji mention to its stored ID before deletion. Legacy setup stored the ID but legacy delete queried the raw mention, making custom-emoji deletion unusable. This is an intentional reliability fix. Explicit delete removes every duplicate row matching the logical key.

## Button Setup And Panel Contract

`/選取身分組-按鈕` does not defer. Permission or hierarchy failures are immediate ephemeral replies. Authorized setup:

- checks the selected role against cached guild roles and the bot's cached highest role;
- generates the JavaScript-shaped base ID from the current UTC+8 `YYYYMMDDHHmm` timestamp plus `Math.random() * 10000000000`, including decimal or exponent formatting;
- writes `<base>add` and `<base>delete` mappings to `btns`;
- opens modal ID `nal` with title `領取身分系統!`;
- uses paragraph input ID `roleaddcontent<base>`, label `請輸入身分訊息內文`, and the legacy optional input behavior.

Modal submit publicly defers, sends a separate panel message, then edits the public deferred response with `<a:green_tick:994529015652163614> | 成功創建領取身分組`. The panel preserves:

- embed title `選取身分組`;
- the submitted content exactly, including empty content;
- the bot member's cached display color;
- primary button `<base>add`, label `✅點我增加身分組!`;
- danger button `<base>delete`, label `❎點我刪除身分組!`.

The typed custom-ID parser accepts only the documented legacy numeric/decimal/exponent base shape plus exact `add` or `delete` suffixes. It does not reproduce the global `includes('add') || includes('delete')` collision risk from `events/btn.js`.

## Button Application Contract

Both button actions defer ephemerally and read `btns` by `{guild,number}` using the raw custom ID. Cached role existence is checked before member-state errors. For an existing role, member-state errors take precedence over hierarchy denial, matching legacy:

- add when already owned: `<a:Discord_AnimatedNo:1015989839809757295> | 你已經擁有身分組了!`;
- remove when not owned: `<a:Discord_AnimatedNo:1015989839809757295> |  你沒有這個身分組!`, including the doubled space;
- missing add role: `<a:error:980086028113182730> | 請通知群主管裡員找不到這個身分組!`;
- missing remove role: `<a:error:980086028113182730> | 找不到這個身分組!`;
- hierarchy denial: `<a:error:980086028113182730> | 請通知群主管裡員我沒有權限給你這個身分組(請把我的身分組調高)!`.

Successful REST writes return the exact green titles `<a:green_tick:994529015652163614> | 成功增加身分組!` and `<a:green_tick:994529015652163614> | 成功刪除身分組!`.

The role state comes from the interaction member payload, and the add/remove operation uses the Discord REST endpoint directly. A current member can therefore use the button even when absent from the local guild-member cache. This is an intentional availability improvement over cache-dependent legacy code and does not require Guild Members intent.

A stale button with no `btns` row receives the ephemeral text `很抱歉，出現了錯誤!` instead of leaving the deferred interaction unresolved. Repository or Discord write failures use the legacy plain-text operational fallback and do not send a false success. Go awaits role writes; legacy started the promise and could display success before Discord rejected it.

## Reaction Event Contract

For reaction add/remove events, Go:

- ignores events known to be from bots;
- uses the custom emoji ID when present and otherwise the Unicode emoji name;
- reads `message_reactions` by `{guild,message,react}`;
- treats a missing mapping as silent;
- checks the configured role with cached guild/bot hierarchy state;
- adds or removes the role through Discord REST;
- sends no success response.

A missing or unassignable configured role sends the legacy best-effort DM. Add uses `<a:error:980086028113182730>`; remove uses `<a:Discord_AnimatedNo:1015989839809757295>`. Closed DMs do not turn the event into a failure. Repository, malformed-row, cache availability, and Discord REST failures are returned to the shared event dispatcher for logging and do not send a misleading hierarchy DM.

Reaction-add gateway payloads include the member. Go retains that member in Discord state so a later remove in the same process can still identify and ignore a bot. Discord reaction-remove payloads do not include a member. A remove first observed after process restart is best-effort: if the member is absent from state and is not the current bot user, bot identity is unknown and the event is handled as a non-bot. No privileged member fetch or new intent is added to hide this limitation.

## Mongo Compatibility

`message_reactions` retains:

- `guild`
- `message`
- `react`
- `role`

Its logical key is `{guild,message,react}`. `btns` retains `guild`, `number`, and `role`, with logical key `{guild,number}`.

Every field in a selected row uses Mongoose-compatible String scalar decoding. BSON strings, numeric values (including JavaScript-style exponent text), Booleans, ObjectIDs, Symbols, and JavaScript-code scalars decode like Mongoose String values. Arrays, documents, null, and other unsupported shapes do not become valid IDs. Logical-key queries remain normalized BSON strings, matching Discord and legacy runtime inputs; malformed non-string key rows require migration/audit rather than relying on decoder coercion. New writes remain typed BSON strings.

Setup writes use one upserting `UpdateMany` per reaction key and one `BulkWrite` for a validated button batch. Every duplicate logical row's `role` is updated, while a key with no match creates one row. This keeps duplicate legacy rows aligned without assuming a unique index, avoids partial writes from a later invalid button config, and removes per-button Mongo round trips. Reaction delete removes every duplicate logical row. Runtime reads retain Mongo `findOne` semantics when pre-existing duplicates disagree.

The application creates no startup indexes. Duplicate-safe non-unique lookup indexes `message_reactions_guild_message_react_lookup` and `btns_guild_number_lookup` may be explicitly applied for runtime reads without first deduplicating data. The catalog also lists candidate unique indexes `message_reactions_guild_message_react` and `btns_guild_number`; do not apply either unique index until a full duplicate/malformed-scalar audit passes, explicit index application is reviewed, and Node/Go write ownership is exclusive. Remove the same-key non-unique fallback before promoting a key to unique. Index migration remains optional for runtime correctness and must not be coupled to enabling the feature.

## Intentional Safety And Reliability Differences

- Custom-emoji delete uses the stored emoji ID instead of reproducing the broken raw-mention lookup.
- Setup and button writes align duplicate rows and await Mongo results instead of racing unawaited delete/save callbacks.
- Reaction creation is awaited before configuration is saved, preventing mappings for reactions Discord rejected.
- Missing channel/message, stale button, repository, parser, and Discord failures produce controlled outcomes instead of null dereferences, unhandled callback errors, or permanently deferred interactions.
- Button role writes use REST and support members absent from local member cache.
- Operational reaction failures are logged without sending a false missing-role/hierarchy DM.
- Legacy component IDs are parsed by bounded exact rules instead of broad substring routing.
- Surrounding whitespace on slash option strings and persisted IDs is normalized before validation and writes.
- User, role, and everyone mentions are suppressed in generated responses and DMs.
- Slash usage is centralized and cannot fail a successful command handler.

The broad legacy Unicode-emoji test, URL substring parsing, ignored URL guild segment, cached hierarchy checks, public modal confirmation, missing-config reaction silence, button error typos/spacing, direction-specific DM icons, and post-restart remove limitation are preserved rather than corrected.

## Runtime Ownership And Migration

Node and Go cannot safely share this feature for the same bot/guilds. Concurrent owners can race `message_reactions` replacement, create mismatched button pairs, double-add reactions, apply roles twice, and emit duplicate failure DMs. Before enabling Go:

1. Stop every Node process that loads the three setup commands, `events/btn.js`, `events/modal.js`, or `events/message_reaction.js` for the target bot.
2. Audit `message_reactions` by `{guild,message,react}` and `btns` by `{guild,number}` for duplicates and malformed scalar values.
3. Preserve existing collections; Go reads legacy scalar forms and writes rollback-compatible strings.
4. Enable the single runtime flag, Gateway, reaction intent, and reviewed staging command sync together.
5. Do not create or require candidate unique indexes during the ownership switch.

## Parity Tests

Focused tests lock exact command definitions, public/ephemeral response visibility, modal/panel payloads, legacy ID formatting and routing, message-link parsing, Unicode/custom emoji handling, cached lookup/hierarchy order, exact button and reaction errors, event failure classification, bot handling, REST writes for uncached members, Mongoose scalar reads, duplicate-safe writes, collection/filter names, gates, command sync, staging preflight, usage ownership, and the full button router workflow. Run:

```bash
go test ./internal/core/domain ./internal/core/services/roles ./internal/discord/features/roles ./internal/discord/customid ./internal/discord/interactions ./internal/discord/events ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/adapters/discordgo ./internal/testutil/fakemongo ./internal/app ./internal/parity ./internal/config ./cmd/mhcat-command-sync ./cmd/mhcat-staging-preflight
```

## Staging Smoke

1. Use an isolated guild, database, text channel, disposable message, manager, ordinary member, bot test user, and role below the bot's highest role.
2. Stop all corresponding Node command, modal/button, and reaction-event ownership.
3. Audit duplicate and malformed `message_reactions`/`btns` rows; do not create indexes.
4. Enable all four required flags, run `mhcat-staging-preflight`, and review command-sync dry-run before apply.
5. Confirm all three commands are discoverable to an ordinary member but return the exact runtime Manage Messages denial.
6. Run reaction setup with both accepted link hosts, Unicode emoji, and a globally cached custom emoji. Verify cached-channel/REST-message lookup order, one added reaction, exact success UI, and typed fields.
7. Repeat setup for the same logical key after seeding duplicates; verify every duplicate points to the new role and no extra row is inserted.
8. Verify invalid role hierarchy precedes URL/emoji errors, invalid custom emoji is rejected, and missing channel/message text is exact.
9. React and unreact as a human; verify one add/remove REST role write and no success message. Exercise missing config (silent), missing role, and hierarchy DMs with exact direction icons.
10. React and unreact as a bot in the same process and verify no role write. Record that a first-seen remove after restart has best-effort bot detection when the member is absent from state.
11. Delete Unicode and custom-emoji mappings. Verify the command adds the reaction first, removes every matching duplicate row, and returns the exact host/missing/success UI.
12. Run button setup, inspect both typed `btns` rows and JavaScript-shaped IDs, submit empty and non-empty optional panel content, and compare the public modal confirmation, bot display color, labels, styles, and raw IDs.
13. Press add/delete as a member absent from local member cache, then exercise already-owned, not-owned, missing-role, hierarchy, stale-config, and forced REST-failure responses.
14. Confirm each slash attempt increments `all_use_counts` once when usage tracking is enabled, while modal submits, button presses, and reaction events do not increment it.

## Rollback

1. Disable `MHCAT_COMMAND_SYNC_INCLUDE_ROLE_SELECTION` and remove the reviewed managed guild commands through command sync.
2. Disable `MHCAT_FEATURE_ROLE_SELECTION_ENABLED`, then stop Go Gateway ownership before restoring Node.
3. Preserve `message_reactions` and `btns`; typed Go string writes remain readable by the legacy Mongoose models.
4. Do not drop duplicates, rewrite scalar values, or apply/remove candidate indexes as part of emergency rollback.
5. Restore Node only after confirming no Go process can handle the modal/buttons or reaction events for the same bot.

Slash-command usage belongs to the assembled runtime's global interaction middleware. It increments `all_use_counts` exactly once per role-selection slash attempt when `MHCAT_FEATURE_USAGE_TRACKING_ENABLED=true`, including permission-denied and failed attempts. Modal submissions, button presses, and reaction gateway events do not write usage counters.
