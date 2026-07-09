# Ticket Runtime and Staging Sync Gates

Status: implemented. Legacy source was not modified.

## Runtime Gate

Ticket runtime side effects are disabled by default.

Enable runtime ticket handlers with:

```txt
MHCAT_FEATURE_TICKETS_ENABLED=true
```

When enabled, the default app runtime:

1. uses the already connected Mongo client;
2. constructs the `tickets` repository;
3. constructs Discord channel/message side-effect ports from the default DiscordGo session;
4. wires ticket setup/delete/open/close routes into the runtime dispatcher.

This does not register slash commands and does not create Mongo indexes.

If custom test/fake app adapters are used with the feature flag enabled, startup fails instead of silently enabling a partial ticket runtime. Tests cover this guardrail.

## Command Sync Gate

Ticket slash commands are still excluded from command sync by default.

Include ticket command definitions only with:

```txt
MHCAT_STAGING_MODE=true
MHCAT_COMMAND_SYNC_INCLUDE_TICKETS=true
MHCAT_COMMAND_SYNC_SCOPE=guild
```

The command-sync CLI rejects ticket inclusion when staging mode is off or scope is not guild.

The staging command-sync scripts also refuse to include ticket commands unless:

```txt
MHCAT_FEATURE_TICKETS_ENABLED=true
```

When ticket inclusion is enabled, the staging expected command list is extended with:

```txt
私人頻道設置
私人頻道刪除
```

Deletion and bulk overwrite remain rejected by staging safety checks.

## Remaining Work

- Run staging command-sync dry-run with ticket inclusion.
- Run staging command-sync apply only after reviewing the plan.
- Run a real staging guild smoke test:
  - `私人頻道設置`
  - modal submit
  - `tic`
  - `del`
- Decide whether the missing bot-user permission overwrite matters in guilds where the bot lacks administrator/manage-channel visibility.
