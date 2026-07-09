# MHCAT-REFACTOR Wave 4: Component / Modal Custom ID Parser + Legacy Compatibility

Wave 3 已完成並通過：

- go fmt ./...
- go test ./...
- go vet ./...
- go build ./cmd/mhcat-bot
- go build ./cmd/mhcat-command-sync
- go build ./cmd/mhcat-mongo-audit
- go build ./cmd/mhcat-mongo-index
- make check

Current boundary status:

- Legacy `MHCAT/` source remained unmodified.
- No Discord command registration happened.
- No Discord Gateway connection happens by default.
- No Mongo feature write happened.
- No Mongo index creation happened.
- Mongo audit CLI remains read-only.
- Mongo index CLI defaults to dry-run.
- No SQL-style migration system exists.
- `internal/core/**` avoids MongoDB driver and DiscordGo imports.
- Wave 4 is safe only within parser / compatibility / router-adapter scope.

現在開始 Wave 4：component/modal custom ID parser、legacy compatibility decoders、versioned encoder、golden tests、collision tests。

不要做 feature parity。
不要搬 ticket/poll/verification/button-role 等功能邏輯。
不要寫 Mongo feature data。
不要註冊 Discord commands。
不要修改 legacy source。

---

# Wave 4 Goals

建立所有互動式 component/modal 後續 feature 會使用的 ID 基礎層：

1. Versioned custom ID encoder for new Go-generated components.
2. Versioned custom ID parser for `mhcat:v1:<feature>:<action>:<payload>`.
3. Legacy custom ID decoders based on `docs/12-component-modal-grammar.md`.
4. Modal custom ID parser.
5. Select/button component ID parser.
6. Route key normalization.
7. Collision / ambiguity detection.
8. Golden tests from documented legacy grammar.
9. Fuzz/property tests for parser robustness.
10. Router integration so Wave 2 router routes by parsed key, not raw string includes.
11. Safe parse errors that do not leak internal details.
12. Docs update.

Wave 4 的產物必須讓我們能回答：

- 每個 legacy custom ID 會被解析成哪個 feature/action？
- 哪些 legacy ID pattern 是 ambiguous？
- 新的 `mhcat:v1` ID 是否保證 <= 100 characters？
- payload 是否可逆、bounded、safe？
- router 是否完全不使用 broad `includes(...)`？
- modal submit 是否和 message component parser 分離？
- parse failure 是否能走 safe responder error？
- legacy live messages 是否仍有相容入口？

---

# Hard Limits

Do not:

1. Modify legacy `MHCAT/`.
2. Implement real feature behavior.
3. Implement slash command feature handlers.
4. Register Discord commands.
5. Connect Discord Gateway by default.
6. Write feature data to Mongo.
7. Create Mongo indexes.
8. Add Mongo repair/backfill writes.
9. Add SQL-style migrations.
10. Add hardcoded guild IDs, channel IDs, role IDs, operator IDs, webhook URLs, tokens, or secrets.
11. Use broad `strings.Contains(rawID, "...")` as router logic.
12. Let `internal/core/**` import DiscordGo.
13. Let `internal/core/**` import MongoDB driver.
14. Put DiscordGo types in parser public API outside DiscordGo adapter boundaries.
15. Log raw tokens, webhook URLs, MongoDB passwords, or full suspicious payloads.
16. Store secrets in `custom_id`.
17. Put unbounded or user-controlled raw text into `custom_id`.

Allowed:

1. Read legacy files to confirm grammar.
2. Update docs.
3. Create parser/encoder packages.
4. Create fake interaction fixtures.
5. Add router integration tests.
6. Add golden tests.
7. Add fuzz tests.
8. Add benchmark tests if useful.
9. Add safe test-only example IDs.
10. Add payload state ID interface without Mongo write implementation.

---

# Gate C Review Before Coding

Before implementing Wave 4, create/update:

```txt
docs/19-wave-4-notes.md
````

Add a Gate C review section for Wave 3:

1. Confirm legacy source is still clean.
2. Confirm no Discord command registration path was added to `cmd/mhcat-bot`.
3. Confirm no Discord Gateway connection happens by default.
4. Confirm Mongo audit CLI is read-only.
5. Confirm Mongo index CLI defaults to dry-run.
6. Confirm no index apply was run.
7. Confirm no Mongo feature writes exist.
8. Confirm `internal/core/**` has no DiscordGo or MongoDB driver imports.
9. Record any Wave 3 issue that must be fixed before Wave 4.

If a serious boundary violation is found, fix that before continuing.

---

# Required Legacy Re-check

Use `docs/12-component-modal-grammar.md` as the source of truth, but re-check legacy files read-only before coding.

Inspect at least:

```txt
MHCAT/events/
MHCAT/functions/
MHCAT/slashCommands/
MHCAT/commands/
MHCAT/handler/
```

Search for:

```txt
customId
custom_id
setCustomId
ModalBuilder
TextInputBuilder
StringSelectMenuBuilder
SelectMenuBuilder
ButtonBuilder
interaction.customId
interaction.fields
interaction.values
includes(
startsWith(
split(
interactionCreate
isButton
isStringSelectMenu
isModalSubmit
```

Update `docs/12-component-modal-grammar.md` if actual code contradicts the existing docs.

Do not modify legacy source.

---

# Required File Tree Additions

Create or update only these Wave 4 areas:

```txt
MHCAT-REFACTOR/
  internal/
    discord/
      customid/
        id.go
        encoder.go
        encoder_test.go
        parser.go
        parser_test.go
        legacy.go
        legacy_test.go
        modal.go
        modal_test.go
        component.go
        component_test.go
        payload.go
        payload_test.go
        collision.go
        collision_test.go
        errors.go
        fuzz_test.go
        benchmark_test.go

      interactions/
        route_key.go
        route_key_test.go
        parser_integration.go
        parser_integration_test.go

    adapters/
      discordgo/
        customid_adapter.go
        customid_adapter_test.go

    testutil/
      fakeinteractions/
        components.go
        modals.go

  docs/
    19-wave-4-notes.md

  testdata/
    customid/
      legacy_components_golden.json
      legacy_modals_golden.json
      versioned_valid.json
      versioned_invalid.json
      ambiguous_legacy.json
      collision_cases.json
```

If existing files already cover the same responsibility, update them instead of duplicating.

---

# Versioned Custom ID Format

Implement the new format:

```txt
mhcat:v1:<feature>:<action>:<payload>
```

Rules:

1. Prefix must be exactly `mhcat`.
2. Version must initially support only `v1`.
3. Feature must be a stable lowercase token.
4. Action must be a stable lowercase token.
5. Payload may be empty only if the route declares no payload needed.
6. Full encoded ID must be 1–100 characters.
7. Payload must be bounded.
8. Payload must not include raw secrets.
9. Payload must not include untrusted raw text unless encoded and length-checked.
10. Parser must reject unknown version unless compatibility mode explicitly accepts it.
11. Parser must reject malformed IDs with typed errors.
12. Parser must preserve enough information for router and future handlers.
13. Encoder must be deterministic.
14. Encoder output must round-trip through parser.

Suggested model:

```go
type ID struct {
    Namespace string
    Version   string
    Feature   string
    Action    string
    Payload   Payload
    Legacy    bool
    Raw       string
}

type RouteKey struct {
    Kind    InteractionKind
    Feature string
    Action  string
    Version string
    Legacy  bool
}

type InteractionKind string

const (
    InteractionKindComponent InteractionKind = "component"
    InteractionKindModal     InteractionKind = "modal"
)
```

Names can be adjusted, but the concepts must exist.

---

# Payload Strategy

Implement a small payload strategy.

Support at least:

1. Empty payload.
2. Plain token payload for safe short IDs.
3. Key-value payload where useful.
4. State ID payload for future Mongo-backed state.

Rules:

1. Encoded payload must keep full custom ID <= 100 characters.
2. Payload parser must reject unsupported characters.
3. Payload parser must reject overlong payload.
4. Payload parser must not panic on malformed input.
5. Payload parser must not log raw suspicious values by default.
6. If payload would exceed 100 characters, encoder should return a typed error recommending state ID.
7. Do not implement Mongo state storage in Wave 4.
8. Do not write any state to Mongo.

Suggested future-safe interface only:

```go
type StateReference struct {
    ID string
}

type StateStore interface {
    CreateState(ctx context.Context, data any) (StateReference, error)
    LoadState(ctx context.Context, ref StateReference) (any, error)
}
```

Do not implement real Mongo store yet.

---

# Legacy Compatibility Decoder

Implement legacy decoders based on `docs/12-component-modal-grammar.md`.

Requirements:

1. Support all documented high-confidence legacy patterns.
2. Mark low-confidence or ambiguous patterns as ambiguous typed errors.
3. Preserve raw legacy ID only inside parser result; do not expose it to logs by default.
4. Normalize legacy IDs into `RouteKey`.
5. Do not dispatch via `includes`.
6. If legacy ID has overlapping patterns, deterministic priority must be documented and tested.
7. If legacy ID embeds multiple positional fields, parser must validate field count.
8. If legacy ID embeds Snowflakes, parser should validate Snowflake shape where practical.
9. If legacy ID embeds user-controlled text, parser must reject or mark unsafe unless documented as legacy-compatible.
10. If legacy parser cannot safely identify route, return typed `ErrAmbiguousID` or `ErrUnknownLegacyID`.

Required feature groups to cover if present in docs:

```txt
ticket
poll
vote
verification
button-role / role button
help menu
gacha / gift / lottery
cron / schedule
logging
report
chatgpt
modal forms
```

If a group is not actually present in `docs/12-component-modal-grammar.md`, do not invent it. Document the absence.

---

# Collision and Ambiguity Detection

Create collision analysis under:

```txt
internal/discord/customid/collision.go
```

It should support:

1. Checking whether two parser rules can match the same ID.
2. Checking documented legacy examples for duplicate route keys.
3. Checking versioned examples for duplicate route keys.
4. Detecting overly broad rules.
5. Producing a deterministic report.

At minimum, write tests for:

1. exact duplicate patterns;
2. prefix overlap;
3. delimiterless overlap;
4. `includes`-style legacy ambiguity;
5. versioned route no collision;
6. unknown legacy IDs skipped or reported.

This is static/test analysis only. Do not build a runtime regex engine that becomes hard to maintain unless needed.

---

# Parser Error Model

Add typed errors:

```txt
ErrEmptyID
ErrTooLong
ErrInvalidNamespace
ErrUnsupportedVersion
ErrInvalidFeature
ErrInvalidAction
ErrInvalidPayload
ErrUnknownLegacyID
ErrAmbiguousID
ErrUnsafePayload
```

Rules:

1. Errors must be comparable or detectable with `errors.Is`.
2. Errors should include safe context.
3. Errors must not include full raw secrets or suspicious payloads.
4. Parser must never panic on arbitrary input.
5. Router should convert parse errors into safe user-facing errors.

---

# DiscordGo Adapter

Create/update:

```txt
internal/adapters/discordgo/customid_adapter.go
```

Responsibilities:

1. Convert DiscordGo message component interaction data into internal parser input.
2. Convert DiscordGo modal submit data into internal parser input.
3. Extract selected values and submitted modal fields into driver-agnostic internal structs.
4. Not expose DiscordGo types to `internal/core/**`.
5. Not implement feature behavior.
6. Not send messages.
7. Not register commands.

If DiscordGo supports new Discord component types not yet modeled, preserve unknown fields safely or return typed unsupported error. Document limitations.

---

# Router Integration

Update Wave 2 interaction router so component/modal routes are based on parsed `RouteKey`.

Requirements:

1. Slash commands still route by exact command name.
2. Components route by parsed `RouteKey`.
3. Modals route by parsed `RouteKey`.
4. Unknown parsed route returns typed not-found.
5. Parse failure returns typed bad-interaction error.
6. No broad `includes` routing.
7. Middleware order remains deterministic.
8. Panic recovery still works.
9. Timeout middleware still applies.
10. Permission checker still receives route metadata.

Do not add real route handlers beyond tests.

---

# Golden Fixtures

Create:

```txt
testdata/customid/legacy_components_golden.json
testdata/customid/legacy_modals_golden.json
testdata/customid/versioned_valid.json
testdata/customid/versioned_invalid.json
testdata/customid/ambiguous_legacy.json
testdata/customid/collision_cases.json
```

Golden entries should include:

```json
{
  "name": "example",
  "kind": "component",
  "raw": "legacy-id-here",
  "want": {
    "feature": "ticket",
    "action": "close",
    "version": "legacy",
    "legacy": true
  },
  "confidence": "high",
  "notes": "source file and line reference if available"
}
```

Rules:

1. Do not include real secrets.
2. Do not include real private user data.
3. Real Discord snowflakes from public legacy code may be replaced with safe placeholder-like numeric strings unless exact matching requires shape.
4. Every high-confidence legacy pattern from docs/12 should have at least one golden test.
5. Ambiguous examples must assert `ErrAmbiguousID`.
6. Unknown examples must assert `ErrUnknownLegacyID`.

---

# Fuzz and Robustness Tests

Add fuzz tests where Go version supports it.

Fuzz target examples:

```go
func FuzzParseCustomID(f *testing.F)
func FuzzParseVersionedID(f *testing.F)
func FuzzParseLegacyID(f *testing.F)
```

Fuzz requirements:

1. No panic.
2. No infinite loop.
3. No allocation explosion for long input.
4. Overlong inputs return `ErrTooLong`.
5. Parser remains deterministic.
6. Parse -> encode -> parse round trip for valid generated IDs.

Normal `go test ./...` must not require long fuzz runs.

---

# Benchmark Tests

Optional but recommended:

```go
func BenchmarkParseVersionedID(b *testing.B)
func BenchmarkParseLegacyID(b *testing.B)
```

Benchmark goal is not performance tuning yet; it is to catch obviously inefficient parser design.

---

# Docs Updates

Create/update:

```txt
docs/19-wave-4-notes.md
```

Include:

1. Gate C review for Wave 3.
2. Files created.
3. Files updated.
4. Versioned ID format decision.
5. Legacy compatibility strategy.
6. Parser rule priority.
7. Collision findings.
8. Ambiguous legacy patterns.
9. Router integration.
10. Test cases added.
11. Commands run.
12. Known limitations.
13. Exact next recommended step.

Also update if materially changed:

```txt
docs/12-component-modal-grammar.md
docs/06-architecture-decision-records.md
docs/09-risk-register.md
docs/10-feature-parity-checklist.md
docs/11-operational-runbook.md
docs/15-gate-b-architecture-freeze.md
```

---

# README Updates

Update README with:

1. Wave 4 status.
2. New custom ID format.
3. Legacy compatibility note.
4. Warning that parser exists but feature handlers do not.
5. Warning that no command registration happens from bot startup.
6. Warning that no Mongo feature writes happen.
7. Testing commands.

---

# Makefile Updates

Ensure `make check` still runs:

```bash
go fmt ./...
go test ./...
go vet ./...
go build ./cmd/mhcat-bot
go build ./cmd/mhcat-command-sync
go build ./cmd/mhcat-mongo-audit
go build ./cmd/mhcat-mongo-index
```

No new unsafe default targets.

---

# Test Requirements

No test should require live Discord or live MongoDB.

## Versioned encoder/parser tests

1. valid empty payload ID parses;
2. valid token payload ID parses;
3. encoder output <= 100 characters;
4. too-long ID fails;
5. unsupported namespace fails;
6. unsupported version fails;
7. invalid feature fails;
8. invalid action fails;
9. invalid payload fails;
10. encode -> parse round trip works;
11. parser is deterministic;
12. raw secrets are not logged in error strings.

## Legacy parser tests

1. every high-confidence legacy component fixture parses;
2. every high-confidence legacy modal fixture parses;
3. ambiguous legacy fixture returns `ErrAmbiguousID`;
4. unknown legacy fixture returns `ErrUnknownLegacyID`;
5. legacy positional field count is validated;
6. legacy Snowflake-like fields are validated where practical;
7. legacy unsafe payload is rejected or marked unsafe;
8. parser priority is deterministic.

## Collision tests

1. duplicate patterns detected;
2. overlapping prefix patterns detected;
3. delimiterless overlap detected;
4. includes-style ambiguity detected;
5. versioned IDs do not collide with legacy unless explicitly allowed;
6. report output deterministic.

## Router integration tests

1. component route uses parsed route key;
2. modal route uses parsed route key;
3. unknown component parse error returns safe error;
4. unknown modal parse error returns safe error;
5. no raw custom ID leaked in user-facing error;
6. middleware order still deterministic;
7. panic recovery still works;
8. timeout middleware still applies;
9. permission checker receives route key.

## DiscordGo adapter tests

1. component interaction extracts custom ID;
2. select values extracted;
3. modal custom ID extracted;
4. modal fields extracted into internal structs;
5. unsupported component type returns typed error;
6. DiscordGo types do not appear in core APIs.

## Fuzz tests

1. arbitrary input never panics;
2. long input returns bounded error;
3. valid generated IDs round-trip.

---

# Acceptance Commands

Run from `MHCAT-REFACTOR`:

```bash
go fmt ./...
go test ./...
go vet ./...
go build ./cmd/mhcat-bot
go build ./cmd/mhcat-command-sync
go build ./cmd/mhcat-mongo-audit
go build ./cmd/mhcat-mongo-index
make check
```

Also run targeted parser tests:

```bash
go test ./internal/discord/customid -run Test
go test ./internal/discord/interactions -run Test
go test ./internal/adapters/discordgo -run Test
```

Do not run long fuzzing unless explicitly requested. Normal fuzz targets only need to compile under `go test ./...`.

---

# Boundary Verification

After implementation, verify and report:

1. Legacy source unmodified.
2. No Discord command registration happened.
3. No Discord Gateway connection happened by default.
4. No Mongo feature write happened.
5. No Mongo index creation happened.
6. Parser supports `mhcat:v1:<feature>:<action>:<payload>`.
7. New encoded IDs are limited to <= 100 characters.
8. Legacy high-confidence IDs have golden tests.
9. Ambiguous legacy IDs return typed errors.
10. Router does not use broad `includes`.
11. Parser errors do not leak raw secrets.
12. No feature command behavior was implemented.
13. `internal/core/**` avoids DiscordGo and MongoDB driver imports.
14. No hardcoded secrets/operator IDs/webhooks were added.

---

# Final Report Format

When Wave 4 is complete, report:

```txt
Wave 4 Complete / Failed

Gate C review:
Files created:
Files updated:
Commands run:
Test results:
Build results:
Parser coverage:
Legacy compatibility coverage:
Collision findings:
Security checks:
Boundary checks:
What is intentionally not implemented:
Known limitations:
Recommended next step:
```

Also explicitly answer:

1. Did legacy source remain unmodified?
2. Did any Discord command registration happen?
3. Did any Discord Gateway connection happen by default?
4. Did any Mongo feature write or index creation happen?
5. Does parser support the new `mhcat:v1:<feature>:<action>:<payload>` format?
6. Are encoded IDs guaranteed to fit Discord's 100-character `custom_id` limit?
7. Are high-confidence legacy IDs covered by golden tests?
8. Are ambiguous legacy IDs rejected with typed errors?
9. Does router avoid broad `includes` routing?
10. Do parser errors avoid leaking raw secrets?
11. Was any real feature behavior implemented?
12. Does `internal/core/**` avoid DiscordGo and MongoDB driver imports?
13. Is Wave 5 safe to start?