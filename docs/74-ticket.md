# Ticket Parity Contract

Status: private-channel setup, config deletion, panel submission, ticket open, and ticket close are parity-audited behind one disabled-by-default runtime ownership gate.

## Legacy References

- Setup command: `MHCAT/slashCommands/私人頻道/ticket.js`
- Config delete command: `MHCAT/slashCommands/私人頻道/ticket_delete.js`
- Panel modal submit: `MHCAT/events/modal.js`
- Ticket open button: `MHCAT/events/Ticket System.js`
- Ticket close button: `MHCAT/events/yicket system.js`
- Model: `MHCAT/models/ticket.js`
- Collection: `tickets`

## Gates And Ownership

Runtime routes require:

```bash
MHCAT_FEATURE_TICKETS_ENABLED=true
```

Publishing the managed commands for an isolated staging guild additionally requires:

```bash
MHCAT_STAGING_MODE=true
MHCAT_COMMAND_SYNC_INCLUDE_TICKETS=true
```

`MHCAT_FEATURE_TICKETS_ENABLED` owns both slash commands, the versioned setup modal, legacy ticket-shaped `nal` modal submissions, and exact legacy `tic`/`del` buttons as one boundary. Do not split those paths between Go processes or leave the corresponding Node command/modal/button handlers active while Go owns the feature. There is no shared lease or cross-process create lock.

Ticket interactions and Discord REST calls require no ticket-specific privileged intent. Command sync remains guild-scoped and staging-only. Preflight rejects command inclusion without the runtime feature and warns when runtime is enabled without command inclusion.

## Command And Usage Contract

The managed definitions preserve:

- `/私人頻道設置`, description `設置私人頻道`;
- required category option `類別`, restricted to Discord channel type `4`;
- required role option `管理員身分組`;
- `/私人頻道刪除`, description `刪除之前設置的私人頻道`;
- default Manage Messages permission metadata (`8192`) on both commands.

Both handlers also check Manage Messages at runtime, including the Discord Administrator override. Setup denial is an immediate ephemeral red reply. Delete publicly defers first, so its denial is a public edit like legacy. The exact denial title is ``<a:Discord_AnimatedNo:1015989839809757295> | 你需要有`訊息管理`才能使用此指令``.

Both legacy command modules declare `cooldown: 10`, but the active legacy slash dispatcher never enforces that metadata. Go likewise adds no feature-local cooldown.

Slash usage belongs only to the assembled runtime's global middleware. It writes one `all_use_counts` event per setup/delete slash attempt when `MHCAT_FEATURE_USAGE_TRACKING_ENABLED=true`, including denied and failed attempts. Modal submissions and `tic`/`del` components never write command usage.

## Setup And Panel Contract

`/私人頻道設置` first rejects an existing `{guild}` config with the exact ephemeral `__**錯誤**__` embed and legacy `` `<>h 刪除私人頻道` `` text. Otherwise it opens a versioned modal carrying only the selected category and admin role IDs.

The visible modal preserves:

- title `私人頻道系統!`;
- required short field `ticketcolor`, label `請輸入嵌入顏色`;
- required short field `tickettitle`, label `請輸入標題`;
- required paragraph field `ticketcontent`, label `請輸入內文`.

Modal submission publicly defers. It validates the color and required text before creating config. Submitted title and content are not trimmed; successful panel embeds retain their raw whitespace. The panel uses primary button ID `tic` with label `🎫 點我創建客服頻道!`. Success edits the deferred response with green title `<a:green_tick:994529015652163614> | 成功創建私人頻道`.

Legacy live modal ID `nal` is routed to this feature only when its first field is `ticketcolor`. That route validates and publishes the same panel without creating config because the legacy command saved config before showing the modal. This permits an in-flight legacy modal to finish after an ownership switch without guessing category or role state.

The versioned path reserves config before sending its panel. A config created after the modal was shown wins; the stale submission receives the duplicate-config edit and sends no panel. If panel delivery fails, Go removes only the exact row inserted by that submission, using its generated Mongo `_id` and guild under a bounded cancellation-independent cleanup context. A later replacement row cannot be removed by a stale rollback receipt. If the success-response edit fails after config and panel delivery, both remain because setup itself completed.

## Color Contract

Legacy first calls `validate-color@2.2.4` and then passes the same raw value to `discord.js@14.25.1` `EmbedBuilder.setColor`. Go preserves the successful intersection rather than the validator's broader accepted set:

- exactly `#RRGGBB`, with case-insensitive hexadecimal digits;
- exact case-sensitive names shared by the HTML validator and Discord `Colors`, such as `Red`, `Aqua`, `DarkGreen`, and `LightGrey`.

No surrounding whitespace is trimmed. Discord names resolve to Discord palette values, so `Red` is `0xED4245`, not CSS red. Lowercase names, 3/8-digit hex, CSS-only names, and functional forms are rejected with title `你傳送的並不是顏色(色碼)`.

Fixed legacy named colors also use Discord values: `Red` is `0xED4245`, `Green` is `0x57F287`, and the ticket-open success literal `#00DB00` remains `0x00DB00`.

## Config Delete Contract

`/私人頻道刪除` publicly defers, checks Manage Messages, and deletes every `tickets` row for the invoking guild. Existing config returns title `刪除私人頻道設定` with description `成功刪除私人頻道的設置\n現在你可以重新創建了!`; missing config uses `你還沒有創建私人頻道的設定\n是要怎麼刪除啦!`. Both embeds use Discord `Red`.

Deletion removes config only. It does not remove published panels or already-open ticket channels. Those panels self-report missing config when used, as described below.

## Ticket Open Contract

Exact legacy button ID `tic` performs these steps:

1. Search every cached guild channel type for a channel whose name exactly equals the invoking user ID. Any match returns the ephemeral `__**客服頻道**__` duplicate warning.
2. Read the guild's first `tickets` row.
3. If config is missing, send `:x: 這個創建私人頻道的設置已經被刪除了喔，請麻煩管理員重新創建!` and then best-effort delete the stale panel message.
4. Create a guild text channel named with the invoking user ID under `ticket_channel`.
5. Send the welcome message and `del` button.
6. Reply ephemerally with `__**頻道**__` and `:white_check_mark: 你成功開啟了頻道!`.

Channel permission overwrites preserve legacy order and values:

1. `admin_id` role: allow View Channel, Send Messages, and Read Message History; deny Create Instant Invite.
2. Current guild `@everyone` role: deny View Channel.
3. Invoking member: the same allow/deny set as the admin role.
4. Current bot application member: the same allow/deny set, using interaction `ApplicationID` and the configured bot ID only as a non-runtime fallback.

The stored legacy `everyone_id` is read and retained for rollback compatibility but is not trusted for channel creation. Discord's guild ID is always the `@everyone` role ID, matching the active legacy lookup and preventing a stale stored value from exposing the channel.

The welcome preserves visible content `||@everyone||`, green `__**私人頻道**__` embed text, and danger button ID `del` labeled `🗑️ 刪除!`. Go suppresses actual everyone mention delivery while retaining the visible text. If welcome delivery fails, Go deletes only the channel it just created under a bounded cancellation-independent cleanup context, so the exact-name duplicate check does not block retry.

## Ticket Close Contract

Exact legacy button ID `del` deletes the current channel when either:

- the actor has Manage Messages; or
- the current channel name exactly equals the actor's user ID.

When the interaction omits channel name, Go resolves the current channel before applying the owner check. An empty or missing channel, or a non-manager in a channel whose name differs from that actor's user ID, receives the public red `__**私人頻道**__` denial embed. Successful deletion sends no response, matching the legacy channel-disappears behavior.

Legacy attempted an owner branch through fetched message history but compared an author ID to the literal string `"null"`, making ordinary owner close effectively unusable. Channel-name ownership restores the intended owner behavior without granting access outside the user-ID-named ticket channel.

## Mongo Compatibility

`tickets` retains these fields:

- `guild`
- `ticket_channel`
- `admin_id`
- `everyone_id`

Reads apply Mongoose-compatible String scalar decoding to every selected field. BSON strings, numeric values, Booleans, ObjectIDs, Symbols, and JavaScript-code scalars decode like Mongoose strings. Arrays, documents, null, and unsupported shapes do not become usable Discord IDs. New writes remain typed BSON strings.

Config creation validates all four values, then uses one `$setOnInsert` upsert containing a generated ordinary Mongo `_id`. An existing `{guild}` match is never overwritten and returns the duplicate-config outcome. Duplicate-key conflicts map to that same outcome. Runtime reads retain Mongo `findOne` semantics when legacy duplicates disagree.

Failure compensation deletes by the exact generated `{_id,guild}` receipt. Explicit command deletion intentionally uses `DeleteMany({guild})` so all legacy duplicate rows are removed. The application creates no startup index. The duplicate-safe non-unique `tickets_guild_lookup` index may be explicitly applied for command/component reads. Candidate unique index `tickets_guild` must not be applied until duplicate and malformed-scalar audits pass and explicit index application is reviewed; remove the lookup fallback before promoting the same key to unique.

Without that unique index there is no cross-process uniqueness guarantee for concurrent upserts. Exclusive Node/Go ownership and a single Go runtime are deployment requirements; do not treat the in-process race contract as a distributed lock.

## Intentional Safety And Reliability Differences

- Go writes config only after a valid versioned modal submission; abandoned and invalid setup no longer leaves config without a panel.
- Stale concurrent modal submissions cannot overwrite config or publish another panel.
- Panel-send failure rolls back only its own config row; welcome-send failure rolls back only its own channel.
- Mongo and Discord side effects are awaited, and cleanup can continue after request cancellation for at most five seconds.
- Explicit config deletion removes all duplicate guild rows.
- The bot and `@everyone` overwrites use current Discord identities rather than stale stored values.
- Visible `||@everyone||` text remains, but actual mention delivery is suppressed.
- Missing config is acknowledged before the stale panel is deleted; Go does not reproduce the legacy three-second deletion of the temporary reply.
- Owner close uses the ticket channel's exact user-ID name instead of the broken legacy message-history condition.
- Legacy component and modal IDs use exact bounded parsing instead of broad substring routing.
- Missing fields, stale rows, repository failures, and Discord failures return controlled errors instead of callback null dereferences or unhandled promise failures.

Legacy raw panel text, supported color intersection, fixed colors, option order, response visibility, duplicate channel lookup across all types, overwrite order, user-ID channel naming, panel/button IDs, typos, and success/error text are preserved rather than normalized.

## Runtime Ownership And Migration

Node and Go cannot safely share ticket ownership for the same bot/guilds. Concurrent owners can show duplicate modals, race singleton upserts, publish duplicate panels, create duplicate channels, or process `del` twice. Before enabling Go:

1. Stop every Node process that loads both ticket slash commands, `events/modal.js` ticket submissions, `events/Ticket System.js`, or `events/yicket system.js` for the target bot.
2. Audit `tickets` by `{guild}` for duplicates, malformed scalar values, empty fields, and stale category/admin-role IDs.
3. Preserve `tickets`; Go reads legacy scalar forms and writes rollback-compatible typed strings.
4. Enable the runtime feature, run staging preflight, review command-sync dry-run, and apply only the two managed staging guild commands.
5. Do not create or require `tickets_guild` during the ownership switch.

## Parity Tests

Focused tests lock exact definitions, permission visibility, modal/panel payloads, raw text, color resolution, stale-modal rejection, create receipts, rollback identity, duplicate-safe delete, all-type duplicate lookup, overwrite order, bot/everyone identities, welcome mention suppression, missing-config acknowledgment order, owner/manager close, exact legacy routes, centralized usage, the complete routed workflow, and concurrent setup behavior. Run:

```bash
go test ./internal/core/domain ./internal/core/ports ./internal/discord/features/ticket ./internal/discord/customid ./internal/discord/interactions ./internal/adapters/mongo/documents ./internal/adapters/mongo/repositories ./internal/testutil/fakemongo ./internal/app ./internal/parity ./internal/config ./cmd/mhcat-command-sync ./cmd/mhcat-staging-preflight
go test -race ./internal/core/ports ./internal/adapters/mongo/repositories ./internal/discord/features/ticket
go test -race ./internal/app -run Ticket
```

## Staging Smoke

1. Use an isolated guild/database, disposable category and text channel, disposable admin role, manager, ordinary member, and bot role with channel-management permission.
2. Stop all corresponding Node ticket command/modal/button ownership and confirm only one Go runtime can receive interactions.
3. Audit duplicate/malformed `tickets` rows and stale Discord IDs; do not create an index.
4. Enable both ticket flags, run `mhcat-staging-preflight`, and review command-sync dry-run before applying the two managed guild commands.
5. Confirm both commands are visible only according to Manage Messages metadata and still return the exact runtime denial when invoked without permission through an existing command or direct interaction.
6. Start setup, abandon one modal, and verify no row is written. Submit invalid colors and verify no row or panel. Test accepted hex and named colors plus rejected lowercase/short/CSS-only values.
7. Submit valid raw-whitespace title/content. Verify one typed row, exact panel text/button, public success edit, and no mention delivery.
8. Open two setup modals before either submit. Submit both and verify only one config/panel wins while the stale modal gets the duplicate error.
9. Force panel send failure and verify only that creation's row is removed. Create/delete/recreate config and verify a stale rollback cannot delete the replacement.
10. Press `tic` as an ordinary member. Verify channel parent/name/type, overwrite order and bits, current guild/bot IDs, welcome UI, suppressed everyone ping, and ephemeral success.
11. Press `tic` again and verify the all-channel-type name check blocks a duplicate. Delete config, press a stale panel, and verify acknowledgment occurs before best-effort panel deletion.
12. Force welcome send failure and verify the newly created channel is removed so a retry can proceed.
13. Close as the owner without Manage Messages, then as a manager. Verify a different ordinary user gets the exact denial.
14. Seed duplicate config rows, run `/私人頻道刪除`, and verify every guild row is removed while panels/channels remain untouched.
15. Confirm each slash attempt increments usage once when enabled while modal and button interactions do not increment it.

## Rollback

1. Disable `MHCAT_COMMAND_SYNC_INCLUDE_TICKETS` and remove the reviewed managed staging commands through command sync.
2. Disable `MHCAT_FEATURE_TICKETS_ENABLED`, then stop Go interaction ownership before restoring Node.
3. Preserve `tickets`; typed Go writes remain readable by the legacy Mongoose model.
4. Do not delete duplicate production rows, rewrite scalar values, or apply/drop `tickets_guild` during emergency rollback.
5. Existing Go-created panels use legacy `tic`/`del` IDs and remain usable after Node resumes, provided their config and category/role IDs remain valid.
6. Restore Node only after confirming no Go process can route ticket slash commands, `nal` ticket submissions, `tic`, or `del` for the same bot.
