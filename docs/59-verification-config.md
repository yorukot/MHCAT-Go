# Verification Config And Flow Slice

Status: implemented behind separate disabled-by-default config and flow gates.

## Scope

Implemented:

- `/驗證設置` command definition behind `MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_CONFIG=true`.
- `/驗證設置` runtime handler behind `MHCAT_FEATURE_VERIFICATION_CONFIG_ENABLED=true`.
- Legacy-compatible write to the Mongoose collection `verifications`.
- `/驗證` command definition behind `MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_FLOW=true`.
- `/驗證` runtime handler behind `MHCAT_FEATURE_VERIFICATION_FLOW_ENABLED=true`.
- Captcha image generation as `captcha.jpeg`.
- Legacy-style green `點我進行驗證!` button.
- Legacy-style modal title `請輸入驗證碼!` and input label `請輸入圖片上的驗證碼`.
- Role assignment and optional `{name}` nickname template after successful captcha.
- Legacy `<captcha>verification` and `<captcha>ver` compatibility for old live messages.
- Staging preflight and script pairing checks for both config and flow flags.

Not implemented:

- Account-age kick behavior from `create_hours`.
- Distributed/persistent verification challenge storage.
- Mongo index creation or repair/backfill.
- Usage-counter Mongo writes.

## Legacy Surface Preserved

Legacy sources:

- `MHCAT/slashCommands/加入設置/verification_set.js`
- `MHCAT/slashCommands/加入設置/verification.js`
- `MHCAT/events/btn.js`
- `MHCAT/events/modal.js`

`/驗證設置` metadata:

- Name: `驗證設置`
- Description: `設置驗證完成後要給甚麼身份組`
- Required role option: `身分組`, description `輸入身份組!`
- Optional string option: `改名`, description `輸入名稱，{name}代表原本的名稱ex:平名 | {name} 就會變成 平名 | 夜貓`
- Permission intent: Manage Messages

`/驗證` metadata:

- Name: `驗證`
- Description: `確保你不是機器人`
- Public command with no options.

Preserved visible flow:

- `/驗證` defers ephemerally before generating the captcha prompt.
- Prompt response attaches `captcha.jpeg`.
- Prompt button label is `點我進行驗證!`.
- Prompt button emoji is `<a:arrow:986268851786375218>`.
- Modal title is `請輸入驗證碼!`.
- Modal input label is `請輸入圖片上的驗證碼`.
- Success title is `<a:green_tick:994529015652163614> | 驗證成功，成功給予你身分組及改名(有的話)!`.
- Missing config error is `這服的管理員沒有設置驗證系統，所以不能使用喔!`.
- Missing role error is `驗證身分組已經不存在了，請通管理員!`.
- Role hierarchy error is `請通知群主管裡員我沒有權限給你這個身分組(請把我的身分組調高)!`.
- Wrong answer error is `你的驗證碼輸入錯誤，請重試(如果看不清楚的話可以重打指令)`.
- Owner nickname error is `你是伺服器服主，我沒有權限改你的名字!`.

## Intentional Safety Fixes

- New Go-generated verification custom IDs use `mhcat:v1:verification:prompt:state=<stateID>` and `mhcat:v1:verification:answer:state=<stateID>` instead of embedding the captcha answer.
- Legacy IDs are still decoded for old live messages.
- The button prompt path rechecks the verification config and role hierarchy before opening the modal, matching the legacy button behavior.
- Role assignment and nickname change are awaited before the success response.
- `/驗證設置` uses duplicate-tolerant update semantics instead of legacy delete-then-insert.
- No unique index is created. A duplicate audit is still required before any future unique index.
- Allowed mentions are suppressed in setup and verification responses.

## Known Limitations

- Verification challenges are stored in process memory with a 5-minute TTL. A restart, process switch, or different shard/process handling the modal can make a valid answer fail. A future production hardening slice should add a bounded Mongo-backed or otherwise shared challenge store before multi-process rollout.
- The generated captcha image uses a small standard-library renderer instead of the old Node captcha package. The user-facing contract remains a `captcha.jpeg` image plus the same button/modal flow.
- Account-age kick remains a separate member-join policy slice requiring Guild Members intent and staging review.

## Enablement

Config setup command:

```bash
export MHCAT_STAGING_MODE=true
export MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_CONFIG=true
export MHCAT_FEATURE_VERIFICATION_CONFIG_ENABLED=true
```

Full verification flow:

```bash
export MHCAT_STAGING_MODE=true
export MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_FLOW=true
export MHCAT_FEATURE_VERIFICATION_FLOW_ENABLED=true
```

Run preflight before syncing:

```bash
go run ./cmd/mhcat-staging-preflight --format text
```

The command-sync include flag and runtime feature flag must be paired. Bot startup still does not register or sync commands.

## Tests

Coverage includes:

- command definition parity;
- service validation and role hierarchy rejection;
- prompt state ownership and role recheck;
- Mongo update document shape for `/驗證設置`;
- fake repository behavior;
- handler success/error UI;
- app route wiring only when both repository and role side-effect ports are provided;
- command-sync include gates;
- staging preflight pairing.

## Remaining Work

Before production rollout, run staging smoke for `/驗證設置` and `/驗證`, then implement a shared challenge store if the Go bot will run with multiple processes or shards that can handle the same verification interaction. Account-age kick remains a separate reviewed slice.
