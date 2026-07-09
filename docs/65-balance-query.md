# Balance Query Slice

Status: gated read-only legacy `/查看餘額`.

## Implemented

- Local slash command definition for `查看餘額`.
- Runtime handler behind `MHCAT_FEATURE_BALANCE_QUERY_ENABLED=false` by default.
- Command-sync include gate `MHCAT_COMMAND_SYNC_INCLUDE_BALANCE_QUERY=false` by default.
- Staging preflight and shell-script pairing checks.
- Read-only Mongo lookup from legacy `chatgpt_gets` by `{guild}`.
- Legacy ephemeral defer/edit response with green embed author `伺服器目前剩於餘額: <price>` and the legacy success gif icon.
- Missing `chatgpt_gets` row displays `0`, matching legacy `data ? data.price : 0`.

## Scope

This slice only implements the visible `/查看餘額` query. It does not enable ChatGPT/autochat message runtime, does not use Message Content intent, does not write `chatgpts` or `chatgpt_gets`, and does not create indexes.

## Staging

Pair both flags before command sync:

```bash
MHCAT_FEATURE_BALANCE_QUERY_ENABLED=true
MHCAT_COMMAND_SYNC_INCLUDE_BALANCE_QUERY=true
```

Smoke with an isolated staging database. Verify both cases:

- no `chatgpt_gets` row for the guild returns `0`;
- a staging `chatgpt_gets` row with `price` renders that price in the ephemeral author embed.
