# Verification Config And Flow Slice

Superseded by the canonical [verification parity contract](80-verification.md). This file remains as historical implementation-slice context.

Status: implemented behind separate disabled-by-default config and flow gates.

## Scope

Implemented:

- `/驗證設置` command definition behind `MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_CONFIG=true`.
- `/驗證設置` runtime handler behind `MHCAT_FEATURE_VERIFICATION_CONFIG_ENABLED=true`.
- Legacy-compatible write to the Mongoose collection `verifications`.
- `/驗證` command definition behind `MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_FLOW=true`.
- `/驗證` runtime handler behind `MHCAT_FEATURE_VERIFICATION_FLOW_ENABLED=true`.
- Legacy-shaped 400x250 captcha generation as `captcha.jpeg`.
- Legacy-style green `點我進行驗證!` button.
- Legacy-style modal title `請輸入驗證碼!` and input label `請輸入圖片上的驗證碼`.
- Role assignment and optional `{name}` nickname template after successful captcha.
- Legacy `<captcha>verification` and `<captcha>ver` compatibility for old live messages.
- Staging preflight and script pairing checks for both config and flow flags.
- Account-age member-join enforcement as a separately gated family documented in the [account-age parity contract](79-account-age.md).

Outside this slice:

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
- Slash-command and button errors retain the animated-no prefix. Modal-submit errors intentionally use the legacy plain red title without that prefix.
- Captcha answers are compared exactly and case-sensitively. Leading or trailing whitespace is not removed from the submitted answer.
- Rename templates are stored and rendered verbatim, including leading, trailing, or all-space values.
- Only the first `{name}` token is replaced, matching JavaScript `String.replace`. A username containing `$$`, `$&`, ``$` ``, or `$'` receives JavaScript replacement-string treatment.

## Captcha Contract

The Go renderer follows `@haileybot/captcha-generator@1.7.0`, the version pinned by the legacy lockfile:

- 400x250 JPEG at quality 75;
- six uppercase letters selected from `ABCDEFHIJLMNOPSTUVWXYZ`;
- ten black crossing lines and 200 black circles;
- centered 90px text rotated between -0.5 and 0.5 radians;
- 5,000 colored foreground noise points.

The renderer searches first for `node_modules/@haileybot/captcha-generator/assets/Swift.ttf`, including the sibling legacy repository layout. It then falls back to the repository Comic Sans and Taipei Sans fonts, and finally a scaled built-in face. Deploy the pinned Swift asset at that package path for the closest legacy typeface. Go rasterization is behaviorally aligned but is not expected to be pixel-identical to `node-canvas`.

The generated answer remains server-side. New prompts contain only a random state ID; the answer is never included in a newly generated component or modal ID.

## Intentional Safety Fixes

- New Go-generated verification custom IDs use `mhcat:v1:verification:prompt:state=<stateID>` and `mhcat:v1:verification:answer:state=<stateID>` instead of embedding the captcha answer.
- Versioned challenges are bound to the initiating guild and user.
- Versioned challenges expire at the inclusive 5-minute boundary and are deleted after successful role/nickname completion, making successful submissions one-time.
- Legacy IDs are still decoded for old live messages. They embed the answer and remain reusable because they have no server-side challenge state.
- Legacy compatibility is intentionally strict: `[A-Za-z0-9]{1,16}verification` and a matching `[A-Za-z0-9]{1,16}ver` modal/field pair are accepted. Malformed IDs are rejected instead of using the legacy broad substring match.
- The button prompt path rechecks the verification config and role hierarchy before opening the modal, matching the legacy button behavior.
- Role assignment and nickname change are awaited before the success response.
- `/驗證設置` uses duplicate-tolerant update semantics instead of legacy delete-then-insert.
- No unique index is created. A duplicate audit is still required before any future unique index.
- Allowed mentions are suppressed in setup and verification responses.

## Mongo Compatibility

The repository uses the legacy `verifications` collection and fields `guild`, `role`, and `name`.

- Missing or BSON `null` `name` values decode as no rename.
- String `name` values are preserved verbatim.
- Saves trim only the guild and role IDs; they do not normalize the rename template.
- Existing rows are updated with `UpdateMany`, so all duplicate rows for a guild receive the same role and name. When no row matches, one row is upserted.
- No migration or unique-index creation runs during startup.

## Known Limitations

- Versioned verification challenges are process-local. A restart, process switch, or different shard/process handling the modal can make a valid answer fail even before the 5-minute expiry. Add a bounded shared challenge store before multi-process or sharded production rollout.
- Failed side effects do not consume the versioned challenge. This permits a retry after a transient Discord failure but can repeat a role-add attempt.
- Legacy IDs necessarily expose their answers and remain reusable. Keep support only for existing live-message compatibility.
- Account-age kick/log behavior has its own runtime, gateway, and Guild Members intent gates; enabling verification does not enable that policy.

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
- exact/case-sensitive answer comparison and nickname replacement-string behavior;
- prompt state ownership, inclusive expiry, one-time success, and role recheck;
- 400x250 JPEG shape, answer alphabet, and renderer noise contracts;
- strict legacy component/modal parser acceptance and malformed-ID rejection;
- Mongo update document shape and missing/null/string `name` decoding;
- fake repository behavior;
- golden setup, prompt, modal, success, and error UI;
- app route wiring only when both repository and role side-effect ports are provided;
- command-sync include gates;
- staging preflight pairing.

## Remaining Work

Before production rollout, run staging smoke for `/驗證設置` and `/驗證`, verify the Swift font asset is available where parity requires it, and implement a shared challenge store if multiple processes or shards can handle the same verification interaction. Review the account-age slice independently because it has separate side effects and intent requirements.
