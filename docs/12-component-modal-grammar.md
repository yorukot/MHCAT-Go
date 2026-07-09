# Component and Modal Grammar

Status: Phase 1.5 Gate B input. Source evidence is local legacy code plus a read-only component/modal archaeology subagent. Legacy source was not modified.

## 1. Inventory Table

| Type | customId pattern | Created in | Parsed in | Payload fields encoded | User-visible feature | Collision risk | Legacy behavior | Proposed Go route key |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Select | `helphelphelphelpmenu` | `functions/menu.js`, `slashCommands/實用工具/help.js` | `events/btn.js` | selected help category value | Help menu | Medium | exact custom ID, values route categories | `help.category.select` |
| Select | `help-menus` | `functions/delete.js` | not found | selected category | Legacy/dead help menu | Low | appears unused; helper has syntax issues | `legacy.help_menus.unused` |
| Modal | `nal` + first input `anntag`/`ann*` | `slashCommands/公告系統/announcement.js` | `events/modal.js` | tag, color, title, content | Announcement send | High | modal ID is generic; first text input routes via `includes("ann")` | `announcement.modal.submit` |
| Button | `announcement_yes`, `announcement_no` | `events/modal.js`, `slashCommands/打工系統/work_set.js` | local collectors | confirmation choice | Announcement and work confirmation | High | same IDs reused by different features; one collector is broad/channel-local | `announcement.confirm`, `work.override.confirm` |
| Modal/Button | `nal` + `ticketcolor`; `tic`; `del` | `slashCommands/私人頻道/ticket.js`, `events/modal.js`, ticket events | `events/modal.js`, `events/Ticket System.js`, `events/yicket system.js` | ticket embed config; open/close action | Ticket/private channel | Medium | `tic` creates ticket channel; `del` deletes current channel | `ticket.panel.submit/open/close` |
| Modal | `nal` + first field `leave_msgcolor` (compat: `leave_msgcontent`) | `slashCommands/加入設置/leave_message.js` | `events/modal.js` | color, title, content | Leave message setup | Medium | first input routes, not modal ID; actual legacy order starts with color | `welcome.leave.modal.submit` |
| Modal | `nal` or legacy modal + `join_msg*` | no live creator found | `events/modal.js` | content, color, image | Join message setup | Medium/dead | parser exists; command appears dashboard-linked | `welcome.join.modal.legacy` |
| Modal/Button | `roleaddcontent<stamp><random>`; `<id>add`; `<id>delete` | `slashCommands/管理系統/releadd.js`, `events/modal.js` | `events/modal.js`, `events/btn.js` | generated role-button record ID and action suffix | Button role pickup | High | broad `includes("add")` / `includes("delete")`; DB lookup by exact button ID | `role.button.modal`, `role.button.add/remove` |
| Modal/Select | modal `<Date.valueOf()>`; inputs `cron_set*`; selects `week_menu`, `hour_menu`, `min_menu` | `slashCommands/自動通知/cron_set.js`, `events/modal.js` | `events/modal.js` local collector | cron doc ID, cron expression, message/embed fields, week/hour/min values | Automatic notification | Medium | invalid cron opens wizard with generic select IDs | `cron.modal.submit`, `cron.week/hour/min` |
| Button/Modal | `lock_start`; `<channelId>anser`; input `anser` | `events/voice_create.js` | `events/voice_create.js`, `events/modal.js` | dynamic voice channel ID and password | Locked voice room | Medium | modal ID embeds channel ID; password compared to stored answer | `voice_lock.prompt/answer` |
| Button/Modal | `<captcha>verification`; `<captcha>ver`; `mhcat:v1:verification:prompt:state=<stateID>`; `mhcat:v1:verification:answer:state=<stateID>` | `slashCommands/加入設置/verification.js`, `events/btn.js`, Go verification flow | `events/btn.js`, `events/modal.js`, Go typed router | legacy captcha answer; new state ID only | Verification | High for legacy; Low for new IDs | legacy embeds captcha answer directly in custom ID; Go-generated IDs do not | `verification.prompt/answer` |
| Button/Select | `poll_<choice>`, `see_result`, `poll_menu`, `menu_choose` | `slashCommands/管理系統/poll.js`, `events/poll.js` | `events/poll.js` | choice text, owner menu values | Polls | High | choice text is user input; parsed with broad prefix/replace | `poll.vote/result/owner_menu/max_choices` |
| Button | `[<userId>]{<page>}text_rank`, `[<userId>]text_rank {<page>}` | rank slash commands, `events/rank.js` | `events/rank.js` | user ID and page | Text rank pagination | Medium | `get_str()` extracts delimiters without strict validation | `rank.text.page` |
| Button | `[<userId>]{<page>}voice_rank`, `[<userId>]voice_rank {<page>}` | voice rank command, `events/rank.js` | `events/rank.js` | user ID and page | Voice rank pagination | Medium | same parser; some disabled center buttons use `text_rank` IDs | `rank.voice.page` |
| Button | `[<userId>]{<page>}coin_rank`, `[<userId>]coin_rank {<page>}` | coin rank command, `events/rank.js` | `events/rank.js` | user ID and page | Coin rank pagination | Medium | same parser | `rank.coin.page` |
| Button | `text_rank`, `text_rank1`, `text_rank2`, `text_rank4`, `text_rank5`, `coin_rank*` | rank command renders disabled controls | `events/rank.js` if clicked | none | Rank disabled controls | Medium | usually disabled, but broad `includes()` would route if enabled | `rank.disabled_legacy` |
| Button | `/<userId>_sing{<year>}-[<month>]` | `slashCommands/代幣系統/sing.js`, `events/btn.js` | `events/btn.js` | user ID, year, month | Sign-in calendar pagination | Medium | custom delimiter parser extracts `/`, `_`, `{}`, `[]` | `economy.sign.page` |
| Button | `<userId>my-profile` | `slashCommands/代幣系統/user-info.js`, `events/btn.js` | `events/btn.js` | user ID | User profile refresh | Medium | broad `includes("my-profile")` | `profile.refresh` |
| Button | `botinfoupdate`, `shardinfoupdate` | `slashCommands/實用工具/info.js`, `events/btn.js` | `events/btn.js` | none | Bot/shard info refresh | Low | broad check then exact branch | `info.bot.refresh`, `info.shard.refresh` |
| Button | `<page>text_leave_role`, `<page>voice_leave_role` | XP role commands, `events/rank.js` | `events/rank.js` | page | XP reward-role pagination | Medium | `replace()` then numeric conversion | `xp.text_reward.page`, `xp.voice_reward.page` |
| Button | `<lotteryId>`, `<lotteryId>search`, `<lotteryId>restart`, `<lotteryId>stop` | lottery command path; old live buttons possible | `events/btn.js` | lottery ID and action suffix | Lottery | Medium | creation path is disabled, but old buttons still route | `lottery.enter/search/reroll/stop` |
| Button | `<commodityId>`, `<commodityId>ghp`, `<digit>ghp_number`, `backghp_number`, `confirmghp_number<commodityId>` | `slashCommands/代幣系統/ghp_shop.js` | local collectors | commodity ID, quantity digit/action | Coin shop purchase | Medium | substring parser around `ghp` and `ghp_number` | `shop.item/detail/buy/qty` |
| Modal/Button | modal `<sum>` + input `captcha`; button `<workName>` | `slashCommands/打工系統/work_set.js` | local collectors | captcha sum, work name | Work system | High | work names are admin-controlled raw custom IDs | `work.captcha`, `work.job.select/confirm` |
| Select | `delete-data` or helper-provided dynamic ID | `slashCommands/管理系統/delete_data.js`, `functions/menu.js` | local collector | delete target value | Delete config data | Low | message-scoped select | `admin.delete_data.select` |
| Select | `loggin_create` | `slashCommands/管理系統/create_logging.js` | local collector | selected log types | Logging setup | Low | message-scoped select | `logging.configure.select` |
| Select | `hour_menu`, `min_menu` | `slashCommands/生日系統/birthday.js` | local collectors | hour/min values | Birthday notification time | Medium | same IDs as cron wizard | `birthday.hour/min.select` |
| Button | `yesssss`, `nooooo`, raw quiz answer text, `main_no_card`, `main_get_card`, `user_no_card`, `user_get_card`, `lookmenumber`, `teach21point`, `thansize` | `slashCommands/代幣系統/game.js` | local collectors; tutorial buttons also in `events/btn.js` | game action or answer text | Coin games | Medium/High | quiz answers become raw IDs; tutorial IDs are global | `game.accept/reject/action/help` |
| Button | random number string | `events/Chatbot.js` | not found as sent component | random ID | Chatbot report error | Low/dead | built but no sent component found | `chatbot.report_legacy` |

## 2. Grammar

The legacy code does not have one authoritative custom ID grammar. It uses exact IDs, suffix IDs, delimiter extraction, and broad substring matching.

Known exact IDs:

```txt
helphelphelphelpmenu
help-menus
tic
del
see_result
poll_menu
menu_choose
week_menu
hour_menu
min_menu
delete-data
loggin_create
botinfoupdate
shardinfoupdate
announcement_yes
announcement_no
lock_start
yesssss
nooooo
main_no_card
main_get_card
user_no_card
user_get_card
lookmenumber
teach21point
thansize
```

New Go-generated ticket setup modal ID:

```txt
mhcat:v1:ticket:setup:c=<categoryID>,r=<adminRoleID>
```

Payload fields:

- `c`: selected category channel ID from legacy option `類別`.
- `r`: selected admin role ID from legacy option `管理員身分組`.

The ID is generated by the Wave 4 custom ID encoder and must stay within Discord's 100-character `custom_id` limit. The modal writes `tickets` only after `ticketcolor`, `tickettitle`, and `ticketcontent` are validated, then sends the legacy ticket panel as a channel message and edits the deferred modal reply with the legacy success embed.

New Go-generated verification IDs:

```txt
mhcat:v1:verification:prompt:state=<stateID>
mhcat:v1:verification:answer:state=<stateID>
```

Payload fields:

- `state`: opaque bounded challenge state ID.

The new verification IDs intentionally do not embed the captcha answer. Legacy `<captcha>verification` and `<captcha>ver` IDs are still decoded for live-message compatibility.

Modal routing patterns:

```txt
nal + firstTextInputId=anntag|anncolor|anntitle|anncontent
nal + firstTextInputId=ticketcolor|tickettitle|ticketcontent
nal + firstTextInputId=leave_msgcolor (compatibility: leave_msgcontent)
nal + firstTextInputId=roleaddcontent<YYYYMMDDHHmm><random>
<timestamp-or-generated-id> + firstTextInputId=cron_setcron|cron_setmsg|cron_setcolor|cron_settitle|cron_setcontent
<channelID>anser + textInputId=anser
<captcha>ver + textInputId=<captcha>ver
<sum> + textInputId=captcha
```

Dynamic component patterns:

```txt
<channelID>anser
<captcha>verification
<captcha>ver
roleaddcontent<YYYYMMDDHHmm><random>
<roleButtonId>add
<roleButtonId>delete
poll_<choiceText>
/<userID>_sing{<year>}-[<month>]
<userID>my-profile
<page>text_leave_role
<page>voice_leave_role
<lotteryID>
<lotteryID>search
<lotteryID>restart
<lotteryID>stop
<commodityID>
<commodityID>ghp
<digit>ghp_number
backghp_number
confirmghp_number<commodityID>
<workName>
<quizAnswerText>
```

Rank delimiter patterns:

```txt
[<userID>]{<page>}text_rank
[<userID>]text_rank {<page>}
[<userID>]{<page>}voice_rank
[<userID>]voice_rank {<page>}
[<userID>]{<page>}coin_rank
[<userID>]coin_rank {<page>}
```

Confidence:

- High for IDs created and parsed in local source.
- Medium for dead or dashboard-shifted paths where parser exists but no current creator was found.
- Low for the chatbot random button because creation was seen but no send path was confirmed.

## 3. Collision and Ambiguity Review

High-risk legacy matching:

- `events/btn.js` routes with `includes("sing")`, `includes("my-profile")`, `includes("add")`, `includes("delete")`, `includes("lotter")`, `includes("verification")`, `includes("teach21point")`, and `includes("thansize")`.
- `events/poll.js` routes `poll_` IDs from raw user-provided choice text.
- Go Poll Wave A accepts legacy `poll_<choiceText>` up to the legacy 80-character choice limit and counts custom ID length by characters for legacy compatibility. New Go-generated poll vote IDs do not embed raw choice text; they use bounded versioned IDs such as `mhcat:v1:poll:vote:i=0`.
- `events/rank.js` routes by broad rank suffixes and uses delimiter extraction without validating missing delimiters.
- `events/modal.js` routes by the first text input custom ID instead of the modal custom ID, and many modals use generic `nal`.
- Several independent `interactionCreate` listeners can observe the same interaction.

Ambiguous or unsafe payloads:

- `poll_<choiceText>` can contain user-controlled text and overlap with other broad routes.
- `<workName>` can contain admin-controlled text and overlap with global routes.
- `<quizAnswerText>` is raw answer text.
- `<roleButtonId>add` and `<roleButtonId>delete` use suffixes that are also common English substrings.
- `announcement_yes` / `announcement_no` are reused by announcement confirmation and work override confirmation.
- `week_menu`, `hour_menu`, and `min_menu` are reused by cron and birthday flows.
- Captcha answers are embedded in custom IDs for verification and work.
- Plain voice lock answers are stored in Mongo and compared against modal input.

Parser risks:

- Variable-length positional fields make IDs hard to validate.
- `get_str()` style parsing can return empty or unintended values when delimiters are missing or repeated.
- Several handlers use `replace()` instead of prefix/suffix-aware parsing.
- There is no central ownership check for all legacy buttons; collectors sometimes rely on channel/message locality.

## 4. Go Design Decision

Decision: C. Introduce versioned custom IDs for newly generated components while supporting legacy IDs for old live messages.

Rationale:

- Existing Discord messages may still contain legacy IDs, so the Go bot must decode legacy IDs during rollout.
- Preserving broad `includes()` routing would preserve the highest collision and spoofing risks.
- Fully replacing legacy IDs would break live tickets, polls, rank pages, verification prompts, shops, games, and setup modals already posted before rollout.

Go routing rules:

- A single component/modal router owns all interaction dispatch.
- New IDs use `mhcat:v1:<feature>:<action>:<payload>`.
- Payload is bounded, parseable, and validated before routing.
- Payload must not contain secrets, raw captcha answers, lock passwords, access tokens, webhook URLs, or Mongo IDs unless the value is safe by design.
- If payload is too large or sensitive, store state in Mongo and put only an opaque state ID in `customId`.
- Legacy IDs are decoded by explicit compatibility decoders, not by broad substring matching.
- A legacy decoder must return one route key, a typed payload, a confidence level, and a reason for rejection on parse failure.
- Collision tests must prove broad legacy patterns cannot misroute to a different Go handler.

Compatibility decoder priorities:

1. Exact IDs with known route keys.
2. Strong delimiter patterns with snowflake/numeric validation.
3. Legacy modal first-input routing only for documented setup flows.
4. Weak raw-text patterns only inside message-scoped collector compatibility, never as global catch-all routes.

Required tests:

- Golden tests for every pattern above.
- Parse rejection tests for missing delimiters, malformed snowflakes, negative pages, oversized payloads, and conflicting substrings.
- Legacy live-message compatibility tests for `tic`, `del`, poll votes, rank pages, verification, voice lock, shop, work, games, and cron/birthday selects.
- New `mhcat:v1` encode/decode round-trip tests with max-length checks.

## 5. Wave 4 Re-check and Implementation Notes

Read-only re-check scope:

- `MHCAT/events/`
- `MHCAT/functions/`
- `MHCAT/slashCommands/`
- `MHCAT/handler/`
- `MHCAT/commands/` was requested but is not present in the local checkout.

Search terms used included `customId`, `custom_id`, `setCustomId`, `ModalBuilder`, `TextInputBuilder`, `StringSelectMenuBuilder`, `ButtonBuilder`, `interaction.customId`, `interaction.fields`, `interaction.values`, `includes(`, `startsWith(`, `split(`, `interactionCreate`, `isButton`, `isStringSelectMenu`, and `isModalSubmit`.

Result:

- The implementation findings did not contradict the Phase 1.5 grammar. The high-risk broad matching remains concentrated in `events/btn.js`, `events/modal.js`, `events/poll.js`, and `events/rank.js`.
- Wave 4 implements explicit parser rules for the high-confidence IDs covered by golden fixtures.
- New Go-generated IDs use `mhcat:v1:<feature>:<action>:<payload>`.
- Ambiguous legacy IDs such as `announcement_yes`, `announcement_no`, `week_menu`, `hour_menu`, `min_menu`, and raw alphanumeric item/work/lottery IDs are rejected with `ErrAmbiguousID` unless a later feature wave has a message-scoped compatibility context.
- Legacy modal routing preserves the observed first-input routing contract for documented setup modals, but it is isolated behind the parser and not exposed as broad router matching.
