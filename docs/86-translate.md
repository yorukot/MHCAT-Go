# Translate Parity Contract

Status: parity-audited against the active legacy slash command, global slash dispatcher, pinned `@vitalets/google-translate-api` 8.0.0 package, discord.js 14.25.1 response/color behavior, current Google-compatible HTTP adapter, runtime wiring, command sync, and staging preflight. Runtime and command sync remain disabled by default. Live provider smoke is still required before production ownership.

## Scope

This contract covers:

- `/翻譯` command metadata, public visibility, options, choices, usage, and response lifecycle;
- source/target handling, provider requests, timeouts, response parsing, output limits, failures, staging, and rollback.

Legacy sources:

- `slashCommands/實用工具/translate.js`
- `events/SlashCommands.js`
- `config.json`
- `@vitalets/google-translate-api` 8.0.0
- discord.js 14.25.1 interaction and embed color behavior

Auto-chat, paid ChatGPT handoff, `/查看餘額`, and `/兌換` are separate features.

## Gates And Ownership

Enable only with the paired staging flags:

```bash
MHCAT_FEATURE_TRANSLATE_ENABLED=true
MHCAT_STAGING_MODE=true
MHCAT_COMMAND_SYNC_INCLUDE_TRANSLATE=true
```

Both feature flags default to false. Command sync is guild-scoped and staging-only. Preflight rejects sync without runtime and warns when runtime is enabled without sync. Runtime registration also requires a configured translate provider; the default app wiring constructs the Google-compatible HTTP adapter only when the runtime gate is enabled.

Stop the Node `/翻譯` owner before enabling Go for the same bot/guild. There is no lease or shared request identity. Running both owners can dispatch the same interaction through competing processes and create duplicate or failed responses.

The command requires normal Gateway interaction delivery but no Message Content, Guild Members, reaction, or voice intent. The deployment needs outbound HTTPS access to the translate provider. No provider API key or legacy `google-translate-token` value is used.

## Definition And Usage

The command is publicly discoverable with no Discord default-member permission:

- name `翻譯`;
- description `翻譯成各種語言`;
- docs URL `https://docsmhcat.yorukot.me/docs/translate`;
- required string `要的翻譯`, description `你要翻譯的句子或是單詞!`;
- required string `目標語言`, description `你要翻譯成的語言!`.

Target choices preserve this exact order:

| Name | Value |
| --- | --- |
| `🇹🇼中文(traditional Chinese)` | `zh-TW` |
| `🇺🇸英文(English)` | `en` |
| `🇯🇵日文(Japanese)` | `ja` |
| `🇰🇷韓語(Korean)` | `ko` |
| `🇩🇪德語(German)` | `de` |
| `🇫🇷法語(French)` | `fr` |
| `🇷🇺俄語(Russian)` | `ru` |
| `🇪🇸西班牙語(Spanish)` | `es` |
| `🇨🇳簡體中文(Simplified Chinese)` | `zh-CN` |

Legacy cooldown metadata is `10`, but the global dispatcher does not enforce cooldowns. Go adds no cooldown.

Usage belongs only to global slash middleware. With usage tracking enabled, every routed attempt records one best-effort event before provider work, including provider failures. The feature handler does not write a second event in production wiring. Provider HTTP calls and response edits never write feature-specific Mongo data.

## Exact Response Lifecycle

The handler defers publicly, creates one public loading follow-up, and edits that same follow-up with the result. The original deferred response is not edited. This preserves legacy `deferReply()` -> `followUp()` -> follow-up `edit()` behavior and the follow-up message identity.

The loading embed is exactly:

- title `<a:load:986319593444352071> | 我正在玩命幫你翻譯!`;
- discord.js named `Green`, `0x57F287`;
- no description, footer, fields, or components.

The successful embed preserves:

- title `<:translate:986870996147507231> 翻譯系統`;
- one full-range discord.js `Random` color from `0x000000` through `0xFFFFFF`;
- field `**<:edittext:986873966884962304> 原文**:` with the source in single backticks;
- field `**<:answer:986873630178832414> 目標語言:**` with the target code in single backticks;
- field `**<:translate1:986873633483939901> 譯文:**` with translated text in single backticks;
- every field is non-inline and remains in the legacy order;
- footer `<interaction.user.tag>的查詢` with the invoking user's display avatar.

Leading and trailing source whitespace is passed to the provider and preserved in the displayed source. Mentions are explicitly suppressed. Real Discord interactions always provide the user tag and avatar; internal synthetic interactions fall back to user ID for footer text.

## Provider Contract And Timing

Legacy 8.0.0 performs a Google Translate page request followed by a form-encoded `batchexecute` POST. It passes source text unchanged, uses source language `auto`, and logs rejected promises without editing the loading follow-up.

Go uses the provider port and the default `https://translate.googleapis.com/translate_a/single` adapter. It sends a form-encoded POST with:

- `client=gtx`;
- `sl=auto`;
- selected `tl` target;
- `dt=t`;
- raw source as `q`.

POST keeps Discord's maximum 6000-character string option out of the URL, including heavily encoded non-ASCII text. Successful array chunks are concatenated in response order. Empty, malformed, non-2xx, network, and decoding responses map to provider-unavailable errors without exposing provider details.

Provider work has a 10-second child-context budget. Translate-enabled runtime raises the interaction middleware lifetime to at least 15 seconds, leaving time to edit the loading follow-up after an HTTP timeout. A larger configured interaction timeout remains unchanged.

## Validation And Output Safety

Normal Discord traffic supplies one of the nine registered target values. Go also rejects missing, whitespace-only, or synthetic unsupported inputs with a controlled validation error.

Each displayed value is bounded to Discord's 1024 UTF-16-unit embed field limit, including the two backticks and a three-unit ellipsis when truncation is needed. Supplementary characters are never split. Backticks inside user/provider text become apostrophes so user input cannot terminate the code span.

These limits affect only the Discord result rendering. The complete raw source, up to Discord's option limit, is sent to the provider before the source field is bounded.

## Failure UI

Legacy provider rejection logs the raw error and can leave the public loading follow-up stuck. Go intentionally edits that same follow-up to a safe error embed:

- generic title `<a:Discord_AnimatedNo:1015989839809757295> | 很抱歉，翻譯失敗，請稍後再試!`;
- synthetic validation title `<a:Discord_AnimatedNo:1015989839809757295> | 請輸入有效的翻譯內容與目標語言!`;
- legacy configured error color `#EA0000` (`0xEA0000`);
- no raw endpoint, status body, token, request, or internal error detail.

Failure to defer, create the loading follow-up, or edit it is returned to the dispatcher and structured logging. A pre-response defer failure can receive centralized safe UI; after a successful defer, the dispatcher does not create a duplicate response. A successful provider call followed by a Discord edit failure is not retried automatically and can leave the loading follow-up unchanged.

## Data And Migration

The feature performs no Mongo read, feature write, index creation, schema migration, backfill, deduplication, or startup database operation. No database migration is required.

Optional global usage tracking remains independent and can update the shared usage collection exactly as it does for every slash command. That generic write is not translation state and must not be enabled solely for translation migration.

Migration consists of exclusive command ownership, reviewed guild command sync, outbound provider connectivity, and live staging smoke. No legacy provider token, cached request, or translation row needs copying.

## Intentional Differences

Intentional differences are limited to:

- the Go adapter uses a stable Google-compatible single endpoint instead of the legacy page scrape plus batched RPC;
- provider and interaction deadlines replace the legacy unbounded promise;
- provider/validation failures edit the loading follow-up with safe UI instead of logging and remaining stuck;
- long fields are UTF-16 bounded and embedded backticks are neutralized instead of risking Discord rejection or broken markdown;
- mentions are explicitly suppressed;
- empty/malformed provider responses are controlled failures.

Exact public definition, option/choice text and order, raw source whitespace, auto source detection, target codes, public defer/loading follow-up/edit lifecycle, loading title/color, final title/fields/footer, random color, and no-Mongo behavior are preserved.

## Parity Tests

Run focused coverage:

```bash
go test ./internal/adapters/external ./internal/core/services/utility ./internal/discord/features/utility ./internal/app ./internal/config ./internal/parity ./cmd/mhcat-command-sync ./cmd/mhcat-staging-preflight
go test -race ./internal/adapters/external ./internal/core/services/utility ./internal/discord/features/utility ./internal/app
go vet ./internal/adapters/external ./internal/core/services/utility ./internal/discord/features/utility ./internal/app
go run ./tools/parity-audit --legacy-root ../MHCAT --format markdown
```

Adapter tests lock POST form parameters, 6000-character non-ASCII transport, response parsing, and bad-status handling. Handler tests lock exact metadata/UI, follow-up identity, random-color injection, raw whitespace, UTF-16 output bounds, safe provider/timeout errors, and usage ownership. Automated tests do not call the live provider.

## Staging Smoke

1. Use an isolated staging guild, stop the Node `/翻譯` owner, and confirm outbound HTTPS policy permits the provider.
2. Enable paired flags, run preflight and command-sync dry-run, review guild apply, and start one Go owner.
3. Confirm the command is publicly discoverable, has no default permission restriction, and lists all nine choices in exact order.
4. Run `/翻譯 要的翻譯:你好 目標語言:en`; verify one public defer, the exact green loading follow-up, and the same follow-up becoming a random-color result with exact fields/footer.
5. Test `zh-TW`, `zh-CN`, `ja`, and one Latin target; compare provider output with the legacy owner for representative text before cutover.
6. Submit source with visible leading/trailing spaces and mixed emoji; verify the provider receives meaningful content, the source field preserves spaces, and no field exceeds Discord limits.
7. Test a long non-ASCII source near the Discord option limit; verify the request is accepted as POST and the result or controlled provider error replaces loading.
8. Block or redirect provider egress in staging; verify timeout/failure replaces the loading follow-up with exact `0xEA0000` generic UI and no raw details.
9. With usage tracking enabled separately, verify exactly one `翻譯`/`utility` event for success and provider failure.
10. Confirm no Mongo feature collection or index changes, no Message Content intent requirement, no duplicate follow-up, and no command outside the managed guild changed.

## Rollback

1. Disable command-sync inclusion and remove only the managed staging `翻譯` command.
2. Disable the runtime gate and stop the Go owner before restoring Node.
3. Restore Node only after confirming no Go translate route remains.
4. Review any ownership-overlap interval for duplicate interaction responses and provider traffic.
5. Recheck one short translation and the exact loading/result UI under the restored owner.
6. Do not run a Mongo rollback; the feature has no translation data or indexes.

Production ownership remains blocked on live provider connectivity, representative cross-language output comparison, forced failure/timeout smoke, exclusive ownership, and acceptance of the documented safe differences. No database migration is required.
