# XP System Parity Contract

Status: parity-audited against the legacy XP slash commands, text and voice event handlers, Mongoose models, and current Go definitions, services, workers, handlers, repositories, app wiring, staging guards, image renderer, and race coverage. Live Discord and operator-gated real-Mongo smoke remain required before production rollout.

## Scope

This contract covers the active behavior of `/聊天經驗`, `/語音經驗`, text and voice XP accrual, level-up announcements, reward roles, XP-derived coin grants, notification-channel configuration, reward-role configuration, `/經驗值改變`, `/經驗值重製`, `/聊天排行榜`, and `/語音排行榜`.

Detailed UI and data contracts remain in [50-text-xp-config.md](50-text-xp-config.md), [51-voice-xp-config.md](51-voice-xp-config.md), [68-xp-reward-role-config.md](68-xp-reward-role-config.md), [69-xp-admin.md](69-xp-admin.md), [70-xp-reset.md](70-xp-reset.md), and [71-xp-rank.md](71-xp-rank.md).

## Profile Commands

The active legacy `/聊天經驗` and `/語音經驗` paths publicly defer and immediately edit the original response to the red removal embed `該指令即將被移除，請使用\`/我的檔案\`進行替代`. Both legacy files contain rank-card code after that unconditional return, so the renderer is unreachable and is not part of the runtime parity contract.

Go preserves the command names, descriptions, optional `玩家` option, cooldown metadata, public lifecycle, exact embed title/color, and no-Mongo behavior. It intentionally does not implement unreachable `RankCard.png` output. The replacement `/我的檔案` behavior is covered by [97-economy-profile.md](97-economy-profile.md).

## Text And Voice Runtime

Text accrual preserves legacy profile creation and update semantics, XP thresholds, configured/default announcement behavior, missing-channel fallback, reward-role changes, and XP coin rewards derived from `gift_changes.xp_multiple`. It is independently gated by message gateway and privileged intent requirements.

Voice runtime preserves join/leave state, startup reconciliation of existing joined rows, one 30-second loop per joined user, `+5 XP` ticks, level-up reset behavior, configured/default announcements, owner-DM fallbacks, reward-role changes, and XP coin rewards. It is independently gated by gateway Voice State events.

Repository writes retain rollback-compatible collection and field names, including `text_xps`, `voice_xps`, `text_xp_channels`, `voice_xp_channels`, `chat_roles`, `voice_roles`, `coins`, and `gift_changes`. No automatic index creation or data normalization occurs at runtime.

## Administration And Rankings

Configuration, reward-role, admin-adjustment, and reset commands preserve their legacy definitions, permissions, public/ephemeral lifecycle, visible messages, component ownership, field spellings such as `leavel`, and one-row or multi-row Mongo behavior documented in their focused contracts.

The two leaderboard commands preserve the legacy loading UI, 1000x500 `user-info.png` output, rank math, source-order tie behavior, duplicate viewer behavior, amount formatting, pagination IDs and disabled states, missing-user UI, source assets, and font fallback order. Parser bounds and image-fetch limits are documented safety fixes.

## Operational Gates

All XP surfaces remain disabled by default and require their matching runtime and command-sync or gateway flags. Production rollout requires exclusive Node/Go ownership, live Discord visual and event smoke, disposable Mongo integration execution, collection scalar/duplicate audit, asset/font verification, backup and rollback rehearsal, and explicit approval for XP and coin writes.

Indexes and automatic data repair remain operator-owned migration operations. Never silently merge duplicate profiles/configuration rows or coerce malformed legacy scalars during startup.

## Verification

```bash
go test ./internal/core/services/xp ./internal/adapters/mongo/documents \
  ./internal/adapters/mongo/repositories ./internal/discord/features/xp ./internal/app
go test -race ./internal/core/services/xp ./internal/adapters/mongo/repositories \
  ./internal/discord/features/xp ./internal/app
go vet ./...
```

The guarded Mongo integration harness requires a generated disposable database and skips unless both `MHCAT_RUN_MONGO_INTEGRATION_TESTS=true` and `MHCAT_MONGODB_URI` are set. Never point it at production.
