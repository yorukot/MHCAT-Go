# Economy Settings Slice

Status: implemented behind explicit runtime and command-sync gates. The canonical audited behavior and rollout contract is [99-economy-settings.md](99-economy-settings.md).

## Legacy Reference

- File: `MHCAT/slashCommands/代幣系統/coin_cange.js`
- Model: `MHCAT/models/gift_change.js`
- Command: `coin-related-settings`
- Localized names: `zh-TW: 代幣相關設定`, `zh-CN: 代币相关设定`
- Permission metadata/runtime check: `訊息管理` / Manage Messages
- Cooldown metadata: `10`
- Docs URL: `https://docsmhcat.yorukot.meocs/required_coins`

Options, all required:

- `coin-raffle-takes`, integer
- `check-in-cooldown-time`, integer hours; `0` stores `time=0`
- `check-in-give-coins`, integer
- `notification-channel`, channel type text or announcement
- `level-up-multiply-amount`, number

## Go Implementation

- Runtime flag: `MHCAT_FEATURE_ECONOMY_SETTINGS_ENABLED=false` by default.
- Command sync flag: `MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SETTINGS=false` by default.
- Command sync requires staging mode and guild scope.
- Service: `internal/core/services/economy.SettingsService`
- Repository port: `internal/core/ports.EconomySettingsRepository`
- Mongo adapter: `internal/adapters/mongo/repositories.EconomyRepository.SaveEconomyConfig`
- Handler: `internal/discord/features/economy.SettingsHandler`

The handler defers, validates options, checks Manage Messages, writes `gift_changes`, and edits the original response with the legacy success/error embed text.

## Mongo Compatibility

The write preserves legacy field names:

- `guild`
- `coin_number`
- `sign_coin`
- `channel`
- `xp_multiple`
- `time`

`check-in-cooldown-time` is stored as hours multiplied by `3600`, matching legacy. `xp_multiple` is preserved as `float64` because the Discord option is a legacy `Number`.

## Intentional Fixes

- Existing config replacement sequences and checks delete-one then insert-one instead of launching unawaited writes. Extra duplicate rows remain untouched, preserving legacy final-state behavior.
- Negative gacha cost, sign-in reward, and XP multiplier remain accepted; only negative cooldown is rejected. Oversized positive cooldowns retain JavaScript double seconds in BSON.
- Error responses keep the initial public defer state instead of trying to become ephemeral after defer, which Discord does not allow.

## Not Implemented

- Gacha, XP reward, notification-channel consumer behavior.
- Unique index creation for `gift_changes.guild`.
- Production rollout approval for economy config writes.
- Production usage count writes remain controlled by the separate global usage gate.

## Tests

- Command definition shape, localizations, channel type constraints, ownership, and Manage Messages metadata.
- Service validation and hour-to-second conversion.
- BSON decode/write-field compatibility.
- Handler permission, validation, success embed, and repository save behavior.
- Runtime route gating so settings-only does not publish `/代幣查詢` or `/簽到`.
- Command-sync and staging-preflight flag pairing.
