# Warning History Read-only Slice

Status: historical implementation note, superseded by the complete [warning-system parity contract](84-warning-system.md).

This file originally documented the first read-only `/警告紀錄` slice. The later full-system audit corrected and expanded that contract across `/警告紀錄`, `/警告設定`, `/警告`, `/警告清除`, and `/警告全部清除`.

The important correction is permission behavior: legacy advertises `UserPerms: '訊息管理'` for history but does not enforce it in the command body or global slash dispatcher. Go therefore keeps `/警告紀錄` publicly executable at runtime. The other four warning commands enforce Manage Messages.

The canonical contract now owns exact metadata, public defer/edit visibility, random/red/green colors, mixed `warndbs` and `errors_sets` scalar behavior, duplicate rows, JavaScript threshold/action/splice semantics, role hierarchy, DMs, kick/ban side effects, usage ownership, migration, staging, and rollback. Do not use this historical slice as current rollout guidance.
