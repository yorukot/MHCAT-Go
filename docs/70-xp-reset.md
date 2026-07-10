# XP Reset Parity Audit

Status: parity-audited behind explicit runtime, command-sync, gateway, and intent gates.

## Legacy References

- `MHCAT/slashCommands/經驗系統/reset_xp.js`
- `MHCAT/models/text_xp.js`
- `MHCAT/models/voice_xp.js`

## Scope And Gates

This slice implements `/經驗值重製` with:

- `重製個人聊天經驗`;
- `重製個人語音經驗`;
- `聊天經驗重製`;
- `語音經驗重製`;
- the owner message confirmation flow used by both full-guild resets.

Runtime requires all of:

```bash
MHCAT_FEATURE_XP_RESET_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_GUILD_MESSAGES_INTENT=true
MHCAT_DISCORD_MESSAGE_CONTENT_INTENT=true
```

Staging command sync also requires `MHCAT_COMMAND_SYNC_INCLUDE_XP_RESET=true`. Config validation, staging preflight, and scripts reject incomplete flag combinations.

## Preserved Slash Contract

- The command has no Discord default-member permission because legacy authorizes by guild ownership at runtime.
- The handler defers publicly, resolves the guild owner, and rejects non-owners with ``你必須擁有`服主`才能使用`` in the legacy red error embed.
- Individual subcommands require user option `使用者`; if a malformed interaction omits it, the invoking owner remains the legacy fallback target.
- Missing individual data returns `這位使用者還沒有任何的經驗值喔!`.
- Individual success remains plain content: the green-tick emoji, ` | 成功清除<@user>的聊天經驗` or `語音經驗`.
- Full reset returns the exact warning ``:warning: | 一但刪除，___**將無法復原**___，如確定要還原請於60秒內輸入`^確認^`(只有一次機會)!!!``.

## Preserved Confirmation Contract

- Only the invoking owner in the same guild and channel can consume the pending confirmation.
- The first matching message consumes the request, whether correct or not.
- The exact content `^確認^` is required within 60 seconds. The 60-second boundary is expired, and timeout produces no follow-up.
- Wrong content returns the legacy animated-no error `你輸入了錯誤的確認!因此視為取消還原`.
- Full success uses the trash emoji, title `成功刪除伺服器內所有聊天經驗` or `語音經驗`, and legacy color `#53FF53`.
- Empty voice data returns `伺服器沒有任何語音經驗的資料!`.
- Empty text data intentionally reports success. Legacy checks `if (!data)` after Mongoose `find()`, and an empty array is truthy; the voice branch separately checks `data.length === 0`.

## Mongo Compatibility

- Collections remain `text_xps` and `voice_xps`.
- Individual filters remain `{guild, member}`; full filters remain `{guild}`.
- Go removes all duplicate matching rows for an explicit individual reset. Legacy `findOne().delete()` removes one.
- Go removes all guild rows with one awaited `DeleteMany`. Legacy starts one unawaited delete per result row.
- No indexes are created by application startup.

## Intentional Go Differences

- Confirmation results are sent to the confirmation channel rather than as Discord replies to the owner's confirmation message. Visible content and embeds remain the same.
- Go suppresses allowed mentions on slash and channel responses.
- One process-local request is stored per `{guild, channel, owner}`. Re-arming in the same channel replaces the earlier request, so one owner message cannot trigger multiple text/voice collectors as it could in legacy.
- Persistence failures receive a generic legacy-style error instead of escaping an asynchronous callback. Deletes are awaited before success is sent.
- The grouped runtime gate requires Gateway, Guild Messages, and Message Content even when an operator intends to use only individual resets; this prevents exposing full-reset subcommands without a functioning confirmation path.
- Runtime usage accounting belongs to the global slash middleware, preventing route-level double counting.

Pending confirmations remain process-local and are lost on restart, matching the practical lifetime of legacy collectors.

## Verification Coverage

Automated tests lock:

- exact command definitions, subcommand order, and user options;
- public defer, owner check, owner fallback, individual success, and missing-profile errors;
- exact warning, wrong-confirmation, full-success, empty-voice, and unknown-error payloads;
- same-guild/channel/owner scoping, one-shot consumption, 60-second expiration, and latest-request replacement;
- text-empty success versus voice-empty error behavior;
- guild-scoped deletes and preservation of other guilds;
- duplicate-tolerant repository delete filters;
- runtime, gateway, intent, and command-sync gate validation.

## Staging Checklist

1. Use an isolated staging guild and disposable `text_xps` / `voice_xps` rows.
2. Enable all runtime/intent flags plus the XP reset command-sync flag.
3. Run `go run ./cmd/mhcat-staging-preflight`, then review command-sync dry-run before apply.
4. Verify a non-owner receives the exact owner-only error.
5. Reset disposable individual text and voice profiles, then repeat to verify the missing-profile error.
6. Arm a full text reset; send another user's message and confirm it is ignored, then send a wrong owner message and confirm the request is cancelled without mutation.
7. Re-arm and send `^確認^` within 60 seconds; verify only the staging guild's rows are deleted and the success color/title match.
8. Repeat for voice data and verify an empty voice collection returns its legacy error.
9. Verify an empty text collection still returns the preserved legacy success response.
10. Keep the Node.js XP owner stopped for the same guilds while Go owns these destructive writes.
