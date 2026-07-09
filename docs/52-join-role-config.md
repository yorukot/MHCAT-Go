# Join Role Config Slice

Status: config commands implemented, with member-add assignment implemented behind a separate explicit runtime gate.

## Scope

Implemented:

- `/加入身份組設置`
- `/加入身份組刪除`
- legacy command definitions, options, choices, and Manage Messages requirement;
- legacy-compatible `join_roles` writes;
- legacy bot-role hierarchy check before setup;
- staging command-sync/runtime pairing gates;
- `guildMemberAdd` role assignment from existing `join_roles` rows when explicitly enabled.

Not implemented:

- Guild Members intent enablement by default;
- welcome/leave message emitters;
- verification/captcha is not enabled by this join-role slice; `/驗證` is a separate verification-flow gate;
- account-age kick;
- usage-counter Mongo writes.

## Flags

Runtime:

```bash
MHCAT_FEATURE_JOIN_ROLE_CONFIG_ENABLED=true
```

Command sync:

```bash
MHCAT_COMMAND_SYNC_INCLUDE_JOIN_ROLE_CONFIG=true
```

Both flags must be paired for staging command sync. `mhcat-staging-preflight` and staging scripts reject unpaired flags.

Member-add assignment runtime:

```bash
MHCAT_FEATURE_JOIN_ROLE_ASSIGNMENT_ENABLED=true
MHCAT_DISCORD_ENABLE_GATEWAY=true
MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true
```

This event path has no command-sync flag because it registers no slash command. It is disabled by default and must be tested in staging before production.

## Legacy UI Notes

Success embeds preserve the old title `🪂 加入身分組系統` and the legacy `<@roleID>` role text shape. Allowed mentions are suppressed as a safe output guard. Error titles preserve the animated-no prefix and legacy wording for missing permission, duplicate config, missing config, and unassignable role.

## Mongo Compatibility

Collection: `join_roles`

Fields:

- `guild`
- `role`
- `give_to_who`

Create uses `$setOnInsert` for `{guild, role, give_to_who}` and returns the legacy duplicate error when `{guild, role}` already exists. Delete removes all matching `{guild, role}` rows to reduce duplicate drift when an operator explicitly deletes that config.

Assignment reads all matching `{guild}` rows and applies the legacy `give_to_who` values:

- `all_user`: assign to all joining accounts;
- `all_bot`: assign only to bots;
- `all_member`: assign only to non-bot members.

Go attempts all matching role assignments and returns a joined error if one or more role adds fail. This intentionally avoids the legacy behavior where one missing/unassignable role could stop later matching roles.

No index is created by app startup. The planned unique `{guild:1, role:1}` index still requires a duplicate audit.

## Staging Checklist

1. Use an isolated staging guild and staging database.
2. Set `MHCAT_FEATURE_JOIN_ROLE_CONFIG_ENABLED=true`.
3. Set `MHCAT_COMMAND_SYNC_INCLUDE_JOIN_ROLE_CONFIG=true`.
4. Run `go run ./cmd/mhcat-staging-preflight`.
5. Run staging command-sync dry-run and review `加入身份組設置` / `加入身份組刪除`.
6. If applying, keep guild scope, no delete, no bulk overwrite.
7. Test with a role below the bot's highest role.
8. To test member-add role assignment, also set `MHCAT_FEATURE_JOIN_ROLE_ASSIGNMENT_ENABLED=true`, `MHCAT_DISCORD_ENABLE_GATEWAY=true`, and `MHCAT_DISCORD_GUILD_MEMBERS_INTENT=true`.
9. Join with a staging member and bot account if testing `all_member`/`all_bot`.
10. Confirm welcome messages and account-age kick are not expected from the Go bot in this slice. If `/驗證` is also being tested, enable the separate verification-flow flags and follow the verification runbook.

## Next Work

Implement welcome/join-message event sending or account-age kick as separate slices after reviewing `welcome.js`, template mention policy, dashboard-shared collections, and Guild Members intent rollout. `/驗證` captcha/role/nickname is handled by the separate verification-flow gate.
