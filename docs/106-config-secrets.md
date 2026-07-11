# Config And Secrets Contract

Status: parity/security-audited against legacy environment/config access, current Go config loading/aliases/validation, `.env.example`, ignored local `.env`, staging guards, redaction helpers, and a production-source AST scan. External rotation of any credential exposed by legacy history remains an operator action.

## Sources And Defaults

Runtime credentials and deployment identities come only from environment-backed config. `.env` is ignored and untracked; `.env.example` contains names and safe defaults with empty token, Mongo URI, webhook, owner, and deployment-specific ID values. Feature, command-sync, Gateway intent, write-job, and lease flags default off.

Legacy aliases remain accepted only where documented for migration, including `REPORT_WEBHOOK`; canonical names use the `MHCAT_` prefix. Validation pairs sensitive dependencies with their runtime gates, including report webhook presence, staging guild scope, Gateway/intents, write acknowledgements, lease owner/TTL, and command-sync runtime flags.

## Secret Handling

Discord tokens, webhook URLs, Mongo credentials, passwords, API keys, and secret-like long values are redacted before diagnostic output. Redaction preserves only bounded non-secret context such as a last-four suffix. Controlled Discord errors do not expose backend/webhook errors.

No production Go source may embed a Discord webhook URL or assign/compare a literal Discord snowflake through an `ownerID`, `adminID`, or `operatorID` identifier. `internal/config/security_scan_test.go` parses every non-test Go file under `internal` and `cmd` to enforce this. UI emoji/application assets are not deployment privileges and remain allowed.

Legacy MHCAT special-welcome guild/bot/channel IDs are empty-by-default environment values and must be supplied as one reviewed set. Guild ownership checks use Discord API/cache data rather than configured privileged IDs.

## Verification

```bash
go test ./internal/config ./cmd/mhcat-staging-preflight ./internal/app
go test -race ./internal/config ./internal/app
go vet ./...
git check-ignore .env
git ls-files .env .env.example
```

Expected repository evidence: `.env` is ignored and absent from tracked files; `.env.example` is tracked; the AST scan, redaction tests, alias tests, validation tests, staging no-secret tests, and full repository checks pass.

## Operations

Never print or paste live config during audit. Run staging preflight before Gateway, command-sync apply, or recurring writes. Rotate/revoke any credential previously exposed by legacy source/history, inspect provider logs, and update only the deployment secret store. Do not commit populated env files or encode secrets/privileged IDs in custom IDs, docs, fixtures, shell scripts, or error messages.

Production remains gated on operator confirmation that legacy webhook/token/database credentials were rotated where required and that the deployment secret store contains the reviewed canonical variables.
