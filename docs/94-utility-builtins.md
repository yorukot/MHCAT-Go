# Built-In Utility Parity Contract

Status: parity-audited against the active legacy `/help`, `/ping`, and `/info` command handlers, the legacy help selector and info refresh component handlers, discord.js 14.25.1 behavior, the current 74-command Go catalog, DiscordGo adapters, runtime wiring, staging guards, race coverage, and Linux host/process metric collection. Live Discord staging remains required before production rollout.

## Scope

This contract covers:

- `/help`, category queries, command-detail queries, and `helphelphelphelpmenu`;
- `/ping`;
- `/info user`, `/info bot`, `/info shard`, and `/info guild`;
- `botinfoupdate` and `shardinfoupdate`.

Legacy sources are `slashCommands/實用工具/help.js`, `ping.js`, and `info.js`, `functions/menu.js`, `events/btn.js`, and `index.js` `receiveBotInfo`. `/翻譯` is separately canonical in [86-translate.md](86-translate.md); auto-chat is canonical in [89-autochat-config.md](89-autochat-config.md), [90-autochat-fallback.md](90-autochat-fallback.md), and [91-autochat-paid.md](91-autochat-paid.md).

Historical Wave 5.8-5.11 notes describe the implementation sequence. This document supersedes their earlier intentional-difference and incomplete-scope statements.

## Registration And Ownership

`help`, `ping`, and `info` are always part of the built-in runtime registry. They have no feature runtime gate and no Mongo repository dependency. Bot startup registers routes in memory but never creates, updates, or deletes Discord application commands.

Command sync remains an explicit guarded operation. The default staging expected set is:

```text
help,ping,info
```

The 74-command static audit locks exact names, descriptions, localizations, options, required flags, choices, and default permissions. Help uses the complete 74-command catalog regardless of runtime feature gates because legacy loaded every slash-command file into its help inventory.

The global slash middleware is the only production usage owner. With `MHCAT_FEATURE_USAGE_TRACKING_ENABLED=true`, each built-in slash attempt writes exactly one `all_use_counts` event before route lookup and handler work. Help selectors and info refresh components emit no global usage event. Runtime wiring nils route-level trackers to prevent double counting.

## Help Overview

`/help` defers publicly and follows with the legacy random-color overview embed. It preserves:

- author `MHCAT`, icon `https://i.imgur.com/AQAodBA.png`, and the legacy invite URL;
- the exact Chinese description, coffee link, privacy link, and terms link;
- caller footer and current timestamp;
- the 16-category `helphelphelphelpmenu` select in legacy directory order;
- invite, support-server, and website link buttons.

Legacy constructs category fields for the overview but never calls `addFields`; therefore the overview embed has no fields. Go preserves that omission.

## Help Categories

Slash category lookup and selector category lookup are distinct legacy paths.

`/help 指令名稱:<category>` is public. Its title omits the category emoji, each command field starts with `/` followed by the command emoji and backticked name, and descriptions are plain text.

The selector category response is ephemeral. It uses the category emoji in the title, preferred-guild-locale command names/descriptions, slash-command mentions with fallback application command ID `964185876559196181`, one field per subcommand when the first option is a subcommand, localized subcommand names/descriptions, `fix` code blocks, and the four legacy per-subcommand documentation arrays. Hidden commands are omitted.

The query splits at the first literal ASCII space. Leading ASCII spaces produce an empty lookup, and tabs are not normalized into spaces. Matching remains case-insensitive after that split.

## Help Command Details

A valid command query returns the random-color, timestamped four-field `指令資料` embed. Name and description come from the complete audited command catalog. The compatibility metadata inventory contains all 74 commands, 50 commands with explicit `UserPerms`, and 35 commands with active `video` URLs.

Permission text, malformed historical URLs, and YouTube links are preserved exactly. Missing permission text renders ```` ```這個指令大家都可以用喔``` ````, and missing tutorial text renders ```` ```此指令目前沒有教學``` ````.

An unknown query returns the fixed red title `無效的指令! 使用 /help 查看所有指令!` and does not expose internal state.

## Ping

`/ping` replies directly and publicly:

```text
<latency emoji> **Pong!** `<milliseconds>`ms
```

The elapsed value is integer milliseconds from interaction creation to handler sampling. Missing creation time displays zero. A future creation time remains negative, matching JavaScript subtraction.

Thresholds are exact:

| Milliseconds | Emoji |
| --- | --- |
| `<=125`, including negative | `<:icons_goodping:1084881470075703367>` |
| `126..180` | `<:icons_idelping:1084881570013388860>` |
| `>180` | `<:icons_badping:1084881519581069482>` |

## Info User

`/info user` defers publicly, selects the optional user or caller, reads guild-member state first with context-aware REST fallback, and edits the original response. The random-color embed preserves the exact title, member display-avatar thumbnail, user ID, JavaScript-rounded account creation timestamp, and JavaScript-rounded guild join timestamp. The three fields are not inline.

Animated guild-member avatars use the Discord CDN GIF form. Missing or failed lookups return a controlled red embed without leaking adapter errors.

## Info Guild

`/info guild` defers publicly, reads cached guild state with context-aware REST fallback, and edits the original response. The random-color embed preserves the exact title, guild icon thumbnail, 1024-size banner, and eight inline fields:

1. guild ID;
2. member count;
3. boost count/tier;
4. JavaScript-rounded creation timestamp plus relative timestamp;
5. owner mention;
6. emoji count;
7. preferred locale flag and code;
8. numeric verification level and legacy Chinese explanation.

Unknown locale prefixes render literal `undefined`. Animated icons and banners use Discord CDN GIF URLs, matching current discord.js defaults. Failed lookups return a controlled red embed.

## Info Bot

`/info bot` defers publicly and posts a public follow-up. The random-color timestamped embed preserves the seven legacy fields and green `botinfoupdate` button.

Linux collection matches Node host/process intent:

- CPU model is the first `/proc/cpuinfo` model name;
- CPU usage is aggregate host usage sampled over one second and displayed with two decimals;
- RAM is rounded host `(totalmem-freemem)\totalmem` MiB plus two-decimal percentage;
- boot time uses JavaScript-equivalent rounded `now-process uptime`;
- guild and user totals come from cached guild state;
- the supported Go runtime currently owns one Discord session/shard.

Metric collection honors interaction cancellation. Linux `/proc` or `sysinfo` failure returns the controlled bot-info error. Non-Linux builds retain a portable process fallback but require live platform smoke before production use.

`botinfoupdate` preserves the legacy component lifecycle: it ephemerally defers, then posts a new public bot-status follow-up. It does not update the old message and does not send a success confirmation. The refreshed shard label is `集群數量`, and the refreshed total-server field intentionally has an empty field name because legacy omitted `name`.

## Info Shard

`/info shard` defers publicly and follows with a random-color timestamped title plus the green `shardinfoupdate` button. The initial embed has no shard fields because legacy adds fields only after refresh. It does not query bot metrics.

`shardinfoupdate` updates the existing message immediately. Its one local-shard field preserves:

- shard ID, cached guild count, and summed cached member count;
- rounded Go heap reservation as the V8 `heapTotal` analogue;
- process RSS from `/proc/self/statm`;
- uptime as unbounded `HHhMMmSSs`;
- bare integer heartbeat latency milliseconds with no `ms` suffix.

Shard refresh does not run the one-second host CPU sampler. The current app wiring uses `ShardCount: 1`; multi-process aggregation is not claimed. A future multi-shard deployment must add and stage an explicit cross-process aggregation owner before increasing that count.

## Components And Errors

Only exact legacy IDs execute behavior after typed parsing: `helphelphelphelpmenu`, `botinfoupdate`, and `shardinfoupdate`. Substring-containing IDs do not acquire inner behavior.

Go intentionally returns controlled red embeds for provider/lookup failures instead of reproducing legacy rejected promises, reply-after-defer failures, or raw exception handling. Optional missing media is omitted, and impossible missing IDs/timestamps use safe visible fallbacks. These differences prevent internal leakage and malformed Discord payloads without changing successful UI.

## Mongo And Startup

Built-in utility handlers require no Mongo collection and perform no schema migration, backfill, index creation, or startup write. Construction of command catalogs, modules, providers, and routes is in-memory only. When global usage tracking is disabled, all covered slash and component paths write no Mongo data.

When usage tracking is enabled, only `all_use_counts` is affected under its separate global contract. No utility-specific database migration is required.

## Verification

Run:

```bash
go test ./internal/discord/features/utility \
  ./internal/core/services/utility \
  ./internal/discord/commandcatalog \
  ./internal/discord/commands \
  ./internal/discord/customid \
  ./internal/discord/interactions \
  ./internal/discord/responses \
  ./internal/adapters/discordgo \
  ./internal/app ./internal/config \
  ./cmd/mhcat-staging-preflight

go test -race ./internal/discord/features/utility \
  ./internal/core/services/utility \
  ./internal/discord/commands \
  ./internal/discord/interactions \
  ./internal/adapters/discordgo ./internal/app

go vet ./internal/discord/features/utility \
  ./internal/core/services/utility \
  ./internal/discord/commands \
  ./internal/discord/interactions \
  ./internal/adapters/discordgo ./internal/app

go run ./tools/parity-audit --legacy-root ../MHCAT --format markdown
```

The slash audit must remain `74/74` with zero drift, missing definitions, extras, and parser errors. Linux tests exercise real host/process reads; deterministic adapter and handler tests lock exact rendered values and lifecycle. A Darwin adapter compile verifies the portable build-tag path.

## Staging Smoke

1. Use the staging guild and guarded guild-only command sync. Confirm the plan includes managed `help`, `ping`, and `info`, with no delete, bulk overwrite, or global operation.
2. Run `/help`; verify no overview fields, 16 selector options in order, three link buttons, random color, timestamp, and caller footer.
3. Exercise slash category, selector category, command detail, unknown command, localized command, subcommand fields, malformed tutorial URLs, and leading-space/tab queries.
4. Run `/ping` near normal thresholds and verify public direct reply, emoji, integer value, and no defer.
5. Run `/info user` for self and another cached/uncached member. Verify exact avatar, timestamps, and controlled missing-member failure.
6. Run `/info guild`; verify icon/banner, all eight fields, locale, verification text, and owner mention.
7. Run `/info bot`; allow the one-second CPU sample and verify host CPU/RAM, boot timestamp, counts, random color, and refresh button.
8. Click `botinfoupdate`; verify ephemeral defer behavior, a new public follow-up, `集群數量`, unnamed server-count field, and no success message/update of the old message.
9. Run `/info shard`; verify the initial embed has no fields. Click refresh and verify immediate local process heap/RSS, uptime, and bare latency.
10. With usage tracking enabled against disposable rows, verify one event per help/ping/info slash attempt and no event for selectors or refresh buttons. Repeat with tracking disabled and confirm no Mongo writes.
11. Force cancellation and provider/Discord lookup failures. Verify controlled errors, no secret leakage, no duplicate response, and no new collection/index.

## Rollback

1. Revert the bot binary/runtime deployment. No utility data rollback exists because these handlers own no utility collection.
2. If command definitions were staged, use guarded guild command sync to restore the reviewed prior definitions; never use startup sync, global mutation, delete, or bulk overwrite.
3. If usage tracking was enabled, disable it and reconcile only disposable `all_use_counts` rows under the global usage procedure.
4. Restore Node ownership only after the Go interaction process is stopped, then smoke help, ping, all info subcommands, and all three component IDs.

Production rollout remains gated on live Discord staging of media URLs, host metrics, interaction lifecycle, and the explicit single-shard deployment assumption.
