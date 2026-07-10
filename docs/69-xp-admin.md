# XP Admin Parity Audit

Status: parity-audited behind explicit runtime and command-sync gates.

## Legacy References

- `MHCAT/slashCommands/經驗系統/xp_add.js`
- `MHCAT/models/text_xp.js`
- `MHCAT/models/voice_xp.js`

## Scope And Gates

This slice implements `/經驗值改變` with subcommands `聊天經驗改變` and `語音經驗改變`.

Runtime requires `MHCAT_FEATURE_XP_ADMIN_ENABLED=true`. Staging command sync also requires `MHCAT_COMMAND_SYNC_INCLUDE_XP_ADMIN=true`. The staging preflight and scripts reject unpaired flags.

The command mutates existing XP profile data. It does not enable text or voice XP accrual, profile cards, rank rendering, reward roles, coin rewards, or gateway intents.

## Preserved Command Contract

- The command definition and runtime handler require Kick Members (`2`).
- Both subcommands require user `使用者` and integer `經驗值`.
- The interaction defers publicly and edits the original response.
- Permission denial returns the legacy animated-no title with `你需要有\`踢出用戶\`才能使用此指令` and Discord red `#ED4245`.
- Success uses title `<:xp:990254386792005663> 經驗系統`, the green-tick emoji, the selected user mention, label `增加:` even for a negative amount, and legacy color `#53FF53`.
- Go explicitly suppresses allowed mentions in the response.

## Preserved Adjustment Math

Text thresholds use the exact JavaScript operation order:

```txt
parseInt(level * (level / 3) * 100 + 100)
```

Voice thresholds use:

```txt
parseInt(level * (level / 2) * 100 + 100)
```

The legacy loop semantics are also preserved:

- adding exactly the amount calculated as remaining at the current threshold leaves the profile at the current level and stores the legacy loop result rather than leveling up;
- one more XP crosses the threshold;
- subtracting exactly the current XP leaves the level unchanged at zero XP;
- subtracting one more XP moves to the previous level;
- zero delta increments the level and normally resets XP to zero;
- sufficiently negative adjustments may create negative levels or XP.

These outcomes are unusual but are observable legacy behavior, so the admin command does not clamp them.

## Mongo Compatibility

- Collections remain `text_xps` and `voice_xps`.
- Filters remain `{guild, member}`.
- Writes preserve string `xp` and misspelled string `leavel`.
- Reads accept integer-valued strings, BSON `int32`, `int64`, and doubles for both fields.
- A missing text profile is created with the adjusted XP and level.
- A missing voice profile additionally receives `leavejoin: leave`.
- Existing voice profiles retain their current `leavejoin` value.
- Like legacy `findOne` / `updateOne`, one matching duplicate row is read and one matching row is updated. No index is created by application startup.

## Intentional Go Differences

- Legacy creates a missing profile, waits 500 milliseconds for text or one second for voice, reads it again, then responds. Go performs the read, adjustment, and save without artificial delay.
- Legacy starts separate unawaited level and XP updates, which can expose a partially updated profile. Go writes both fields together and waits for the result.
- Go upserts a missing profile directly instead of relying on an unawaited insert becoming visible before the delayed second read.
- Repository failures receive the legacy-style generic red error instead of escaping an asynchronous callback and potentially leaving the deferred interaction unresolved.
- Runtime usage accounting belongs to the global slash middleware, preventing route-level double counting.
- The compatibility model intentionally treats XP and levels as integers. Fractional or malformed legacy values are outside the supported profile schema; integer-valued legacy strings and BSON numeric types remain supported.

## Verification Coverage

Automated tests lock:

- exact legacy command definitions and required options;
- public defer, permission error, generic persistence error, and exact success payload;
- positive, negative, exact-boundary, zero-delta, and below-zero adjustment behavior for text and voice;
- JavaScript floating-point operation order at a threshold where reassociation differs by one;
- missing-profile creation and existing-profile adjustment;
- string XP/level writes, mixed numeric reads, and `leavejoin: leave` on voice insert;
- runtime and command-sync gate pairing.

## Staging Checklist

1. Use an isolated staging guild and disposable `text_xps` / `voice_xps` rows.
2. Enable both XP admin flags and run `go run ./cmd/mhcat-staging-preflight`.
3. Run command-sync dry-run and review `/經驗值改變` before apply.
4. Test positive, exact-threshold, negative, and zero values against a disposable text profile.
5. Repeat against an existing voice profile and confirm `leavejoin` is unchanged.
6. Adjust a member with no text or voice profile and verify string `xp` / `leavel`; verify the new voice row also has `leavejoin: leave`.
7. Confirm a user without Kick Members receives the exact red permission response.
8. Keep the Node.js XP owner stopped for the same guilds while Go owns these shared writes.
