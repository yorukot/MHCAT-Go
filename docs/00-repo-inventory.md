# Repository Inventory

Status: Phase 1 consolidated. Local `MHCAT/` is the source of truth, and it currently matches GitHub `origin/main`.

## Phase 0 Facts

- Workspace root: `/Users/yorukot/Documents/code/mhcat-refactor`
- Legacy repository path: `MHCAT/`
- Refactor workspace path: `MHCAT-REFACTOR/`
- Legacy branch: `main`
- Legacy commit: `6c80405481e79ef53c660678a3e11450ee5596a5` (`6c80405 fix: fix february issue`)
- Legacy working tree: clean (`## main...origin/main`)
- GitHub remote: `git@github.com:yorukot/MHCAT.git`
- GitHub reference check: local `main` matches `origin/main`

## Repository Summary

- Runtime: Node.js
- Discord library: `discord.js` v14
- Persistence: MongoDB via Mongoose 6
- Primary entrypoint: `shard.js`
- Shard child entrypoint: `index.js`
- Deployment: Dockerfile and PM2 ecosystem config
- Package name/version: `discordbot` `2.4.5`
- Total non-git legacy files: 190
- Slash command files: 74
- Event files: 18
- Mongoose model files: 47
- Handler files: 5
- Helper files: 5

## Entrypoints

- `shard.js`: starts `discord.js` `ShardingManager` for `./index.js`, with `totalShards: "auto"`, process mode, respawn enabled, and 5 minute spawn/login timeouts.
- `index.js`: creates a singleton Discord client, connects Mongoose, loads handlers, logs into Discord, handles restart text commands, and registers global process error hooks.
- No in-repo dashboard, HTTP server, or standalone worker entrypoint was found. README links to an external dashboard/site.

## Runtime Scripts

- `npm start`: `node shard`
- `npm run start:prod`: `NODE_ENV=production NODE_OPTIONS='--max-old-space-size=1024' node shard`
- `npm run test`: `node test-startup.js` live Discord login smoke test
- PM2 scripts: `pm2:start`, `pm2:stop`, `pm2:restart`, `pm2:logs`
- Version scripts: `bigupdate`, `update`

## Config Files

- `config.json`
- `.env.template`
- `package.json`
- `Dockerfile`
- `ecosystem.config.js`
- `crowdin.yml`
- `README.md`
- `test-startup.js`
- `chat.json`
- `topic.json`

## Env Vars

Template:

- `TOKEN`
- `MONGOOSE_CONNECTION_STRING`
- `JOIN_WEBHOOK`
- `LEAVE_WEBHOOK`
- `READY_WEBHOOK`
- `REPORT_WEBHOOK`
- `DASHBOARD_URL`

Code uses:

- `TOKEN`
- `MONGOOSE_CONNECTION_STRING`
- `JOIN_WEBHOOK`
- `LEAVE_WEBHOOK`
- `READY_WEBHOOK`
- `REPORT_WEBHOOK`
- `NODE_ENV`
- `NODE_OPTIONS`

`DASHBOARD_URL` appears in `.env.template` but is not consistently used; several dashboard/docs URLs are hardcoded.

## Command Folders

- `slashCommands/`
- No top-level `commands/` folder exists in this checkout, though the prefix dispatcher attempts to load one.

## Slash Command Folders

74 slash command files were verified across:

- `代幣系統`
- `公告系統`
- `加入設置`
- `實用工具`
- `打工系統`
- `扭蛋系統`
- `抽獎系統`
- `生日系統`
- `私人頻道`
- `管理系統`
- `統計系統`
- `經驗系統`
- `群組防護`
- `自動通知`
- `語音包廂`
- `警告系統`

## Event Handlers

- `events/Chatbot.js`
- `events/LoggingSystem.js`
- `events/MessageCreate.js`
- `events/SlashCommands.js`
- `events/Ticket System.js`
- `events/ann_message.js`
- `events/btn.js`
- `events/message_reaction.js`
- `events/modal.js`
- `events/poll.js`
- `events/rank.js`
- `events/ready.js`
- `events/safe_server.js`
- `events/text_xp.js`
- `events/voice_create.js`
- `events/voice_xp.js`
- `events/welcome.js`
- `events/yicket system.js`

## Handler Files

- `handler/index.js`
- `handler/slash_commands.js`
- `handler/channel_status.js`
- `handler/gift.js`
- `handler/cron.js`

## Mongo Models

47 Mongoose model files were verified:

- `Number.js`
- `all_use_count.js`
- `ann_all_set.js`
- `birthday.js`
- `birthday_set.js`
- `btn.js`
- `chat.js`
- `chat_role.js`
- `chatgpt.js`
- `chatgpt_get.js`
- `code.js`
- `coin.js`
- `create_hours.js`
- `cron_set.js`
- `errors_set.js`
- `ghp.js`
- `gift.js`
- `gift_change.js`
- `good_web.js`
- `guild.js`
- `join_message.js`
- `join_role.js`
- `leave_message.js`
- `lock_channel.js`
- `logging.js`
- `lotter.js`
- `message_reaction.js`
- `not_a_good_web.js`
- `poll.js`
- `role.js`
- `sign_list.js`
- `suport.js`
- `system.js`
- `text_xp.js`
- `text_xp_channel.js`
- `ticket.js`
- `verification.js`
- `voice_channel.js`
- `voice_channel_id.js`
- `voice_role.js`
- `voice_xp.js`
- `voice_xp_channel.js`
- `vote.js`
- `warndb.js`
- `work_set.js`
- `work_something.js`
- `work_user.js`

## Assets / Generated Files

- `asset/`: rank/profile backgrounds, Discord fallback images, verification icon, MHCAT image.
- `fonts/`: 8 top-level font files.
- `fonts/language/`: 8 language/emoji font files.
- `lang/zh-tw.json`
- Cartography reported roughly 115 MB of fonts.

## External Services

Confirmed or strongly evidenced:

- Discord gateway, REST, application commands, webhooks, messages, DMs, reactions, buttons, select menus, modals, roles, channels, members, audit logs, sharding.
- MongoDB through Mongoose.
- Google Translate package.
- Captcha generator.
- Canvas/canvacord/chart rendering and Discord CDN image loading.
- Excel/text attachment generation.
- PM2 and Docker runtime.
- External dashboard/docs/support links.

No direct OpenAI API call was found. `chatgpt` Mongo documents appear to be local state or a handoff to an external worker/dashboard not present in this checkout.

## Deployment Mechanism

- Docker: Node 20 image with native canvas dependencies, `CMD ["node", "shard.js"]`.
- PM2: app `mhcat`, fork mode, one instance, logs under `./logs`, max memory restart at `1G`, restart delay and kill timeout configured.
- No GitHub Actions/CI, systemd unit, or docker-compose file was found.

## Legacy Module Map

- `shard.js`: process sharding supervisor.
- `index.js`: Discord client singleton, Mongo connection, handler load, login, global error hooks.
- `handler/slash_commands.js`: global application command create/delete on `ready`.
- `handler/index.js`: event and prefix-command loader.
- `handler/channel_status.js`: periodic statistic channel renames.
- `handler/gift.js`: work completion loop and commented birthday/lottery scheduling.
- `handler/cron.js`: persisted automatic notifications and daily reset jobs.
- `events/SlashCommands.js`: slash command dispatch and command usage counter.
- `events/btn.js`, `events/modal.js`, `events/poll.js`, `events/rank.js`: interactions, buttons, selects, modals, pagination.
- `events/MessageCreate.js`, `events/Chatbot.js`, `events/ann_message.js`, `events/text_xp.js`, `events/safe_server.js`: message-content behaviors.
- `events/voice_xp.js`, `events/voice_create.js`: voice XP and dynamic voice rooms.
- `events/welcome.js`: join/leave/verification/role behavior.
- `events/LoggingSystem.js`: message/channel/voice logging.
- `events/message_reaction.js`: reaction roles.

## Known Bugs / Risks

- Hardcoded Discord webhook URL exists in `index.js` lifecycle code; rotate/revoke if still valid and do not copy.
- Hardcoded operator/admin IDs exist in code.
- `functions/delete.js` is syntactically invalid and appears unreferenced.
- `functions/lang.js` is empty.
- Prefix loader scans a missing `commands/` directory.
- Slash commands are created/deleted globally on every shard `ready`.
- Multiple `interactionCreate` listeners and broad `customId.includes(...)` matching can collide or double-handle.
- `cmd.run(...)` is not awaited in slash dispatch, so async errors can bypass local handling.
- `all_use_count` can increment for non-command interactions before type checks.
- `voice_xp.js` has a `data.save` typo without `()`.
- `safe_server.js` and report paths use user-controlled strings in Mongo regex contexts.
- Birthday scheduler and lottery creation logic appear disabled/commented while related commands/models remain.
- `join_message.enable` has a schema `unique` flag on a Boolean field and is unsafe if ever indexed.

## Performance Risks

- Guild member cache uses `maxSize: Infinity`.
- Per-message XP and logging paths query/write Mongo frequently with no confirmed indexes.
- Rank/profile/sign/poll rendering sorts large guild datasets and renders images inline.
- Poll flows can fetch all guild members/users.
- Voice XP uses per-member intervals and needs restart cleanup.
- Channel status loop renames channels periodically and can hit rate limits.
- Canvas/chart/Excel features are memory-heavy.
- Cron and shard restart behavior can double-run without ownership/lease design.

## Security Risks

- Hardcoded webhook URL and admin IDs.
- Privileged gateway intents enabled: `GuildMembers`, `GuildMessages`, `GuildMessageReactions`, `GuildVoiceStates`, `MessageContent`.
- Raw error helpers and global stack logging can leak sensitive data.
- No central permission, cooldown, owner, or `allowedMentions` policy.
- Message content logging can expose user data.
- OAuth invite evidence uses Administrator permissions.

## Unclear Areas

- Actual production Mongo collection names and existing indexes.
- Whether dashboard or external workers still share Mongo collections, especially `chatgpt`.
- Which disabled/commented features should remain disabled versus restored.
- Canonical owner/admin IDs and whether restart should remain Discord-controlled.
- Target Go deployment topology: single process, multi-shard, multiple replicas, PM2, systemd, or container orchestrator.
- Whether all privileged intents are still approved and acceptable.

## Gate A Inventory Status

- All entrypoints found: yes.
- All commands found: yes, 74 slash command files; no `commands/` directory.
- All events found: yes, 18 event files.
- All components/modals found: partially; broad custom ID patterns are documented, exact payload grammar still needs tests.
- All Mongo models found: yes, 47 files.
- All env vars found: yes for local code/template.
- All external APIs found: mostly yes; external dashboard/worker status remains unknown.
