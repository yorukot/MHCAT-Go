# Translate Utility Slice

Status: superseded by the canonical [translate parity contract](86-translate.md). Implemented behind explicit runtime and command-sync gates.

## Legacy Reference

- File: `MHCAT/slashCommands/實用工具/translate.js`
- Command: `翻譯`
- Options:
  - `要的翻譯`, required string, `你要翻譯的句子或是單詞!`
  - `目標語言`, required string, choices `zh-TW`, `en`, `ja`, `ko`, `de`, `fr`, `ru`, `es`, `zh-CN`
- Cooldown metadata: `10`
- Permission: public
- Docs URL: `https://docsmhcat.yorukot.me/docs/translate`
- Loading title: `<a:load:986319593444352071> | 我正在玩命幫你翻譯!`
- Success title: `<:translate:986870996147507231> 翻譯系統`

## Go Implementation

- Runtime flag: `MHCAT_FEATURE_TRANSLATE_ENABLED=false` by default.
- Command sync flag: `MHCAT_COMMAND_SYNC_INCLUDE_TRANSLATE=false` by default.
- Command sync requires staging mode and guild scope.
- Service: `internal/core/services/utility.TranslateService`
- Provider port: `internal/core/ports.Translator`
- HTTP adapter: `internal/adapters/external.GoogleTranslateClient`
- Handler: `internal/discord/features/utility.TranslateHandler`

The handler defers publicly, creates the legacy loading follow-up, calls the provider, and edits that same follow-up to the legacy final embed. The responder now exposes the required follow-up message handle.

## Intentional Fixes

- Provider failures now return a safe red error embed instead of leaving the loading embed stuck.
- Raw provider errors are not shown to Discord users.
- Output fields are length-bounded and backticks are sanitized to avoid broken embed formatting.
- Empty/unsupported language inputs return a safe validation error.

## Not Implemented

- Auto-chat / ChatGPT message runtime.
- Message Content intent usage.
- Provider selection/config beyond the current Google Translate-compatible adapter.
- Feature-specific Mongo reads or writes. Optional global usage middleware remains independent.

## Tests

- Command definition shape and choice count.
- Translate service validation/provider error mapping.
- HTTP adapter response parsing with `httptest`.
- Handler loading/final/error embeds.
- Runtime route gating with provider requirement.
- Command-sync and staging-preflight flag pairing.
