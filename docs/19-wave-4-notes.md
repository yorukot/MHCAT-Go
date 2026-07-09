# Wave 4 Notes

Status: implemented and verified.

## Gate C Review for Wave 3

- Legacy source status: clean. `MHCAT/` reports `## main...origin/main`.
- Bot command registration: `cmd/mhcat-bot` and `internal/app` have no command registration path.
- Discord Gateway default: `DefaultDiscordEnableGateway=false`; gateway open remains config-gated.
- Mongo audit CLI: read-only by design; no write/index APIs are used by `cmd/mhcat-mongo-audit`.
- Mongo index CLI: dry-run by default; index creation exists only behind explicit `mhcat-mongo-index --apply`.
- Index apply during Wave 3 verification: not run.
- Mongo feature writes: none implemented.
- Core boundary: `internal/core/**` has no DiscordGo or MongoDB driver imports.
- Wave 3 issues requiring fix before Wave 4: none found.

## Legacy Re-check

- Read-only scope: `MHCAT/events/`, `MHCAT/functions/`, `MHCAT/slashCommands/`, and `MHCAT/handler/`.
- `MHCAT/commands/` is not present in the local checkout.
- Search terms: `customId`, `custom_id`, `setCustomId`, `ModalBuilder`, `TextInputBuilder`, `StringSelectMenuBuilder`, `SelectMenuBuilder`, `ButtonBuilder`, `interaction.customId`, `interaction.fields`, `interaction.values`, `includes(`, `startsWith(`, `split(`, `interactionCreate`, `isButton`, `isStringSelectMenu`, and `isModalSubmit`.
- Result: no contradiction found against `docs/12-component-modal-grammar.md`; high-risk matching remains in `events/btn.js`, `events/modal.js`, `events/poll.js`, and `events/rank.js`.

## Files Created

- `internal/discord/customid/id.go`
- `internal/discord/customid/encoder.go`
- `internal/discord/customid/parser.go`
- `internal/discord/customid/legacy.go`
- `internal/discord/customid/modal.go`
- `internal/discord/customid/component.go`
- `internal/discord/customid/payload.go`
- `internal/discord/customid/collision.go`
- `internal/discord/customid/errors.go`
- `internal/discord/customid/*_test.go`
- `internal/discord/interactions/route_key.go`
- `internal/discord/interactions/parser_integration.go`
- `internal/discord/interactions/route_key_test.go`
- `internal/discord/interactions/parser_integration_test.go`
- `internal/adapters/discordgo/customid_adapter.go`
- `internal/adapters/discordgo/customid_adapter_test.go`
- `internal/testutil/fakeinteractions/components.go`
- `internal/testutil/fakeinteractions/modals.go`
- `testdata/customid/legacy_components_golden.json`
- `testdata/customid/legacy_modals_golden.json`
- `testdata/customid/versioned_valid.json`
- `testdata/customid/versioned_invalid.json`
- `testdata/customid/ambiguous_legacy.json`
- `testdata/customid/collision_cases.json`

## Files Updated

- `internal/discord/interactions/interaction.go`
- `internal/discord/interactions/router.go`
- `internal/adapters/discordgo/interaction_adapter.go`
- `docs/12-component-modal-grammar.md`
- `docs/19-wave-4-notes.md`
- `README.md`

## Versioned ID Format Decision

New Go-generated component/modal IDs use:

```txt
mhcat:v1:<feature>:<action>:<payload>
```

Rules implemented:

- namespace must be exactly `mhcat`;
- version support is limited to `v1`;
- feature and action are stable lowercase tokens;
- payload supports empty, token, key-value, and state-reference forms;
- encoded output is deterministic and limited to Discord's 100-character `custom_id` maximum;
- parser returns typed errors for empty, overlong, unsupported namespace/version, invalid feature/action, invalid payload, unsafe payload, unknown legacy ID, and ambiguous legacy ID.

## Legacy Compatibility Strategy

- Existing live legacy IDs are supported through explicit decoders, not broad substring router matching.
- High-confidence component fixtures cover help, ticket, poll, verification, rank pagination, sign/profile refresh, role button, bot info refresh, voice-lock prompt, shop quantity, and game help IDs.
- High-confidence modal fixtures cover announcement, ticket panel, leave message, role-button setup, cron setup, voice-lock answer, verification answer, and work captcha modals.
- Weak raw IDs such as work names, quiz answers, raw item IDs, and reused confirmation IDs are rejected as ambiguous unless a later feature wave provides message-scoped compatibility context.
- The parser preserves raw legacy IDs in parser results for future handler compatibility, but router/user-facing errors do not expose raw suspicious values.

## Parser Rule Priority

Component priority:

1. `mhcat:v1` versioned IDs.
2. Ambiguous exact legacy IDs are rejected.
3. Exact high-confidence legacy IDs.
4. Strong delimiter patterns with snowflake/numeric validation.
5. Bounded documented legacy prefixes/suffixes.
6. Raw alphanumeric IDs are rejected as ambiguous instead of globally routed.

Modal priority:

1. `mhcat:v1` versioned IDs.
2. Documented legacy `nal` first-field routes.
3. Documented cron, voice-lock, verification, and work-captcha modal routes.
4. Unknown or incomplete modal IDs are rejected.

## Collision Findings

- Duplicate exact patterns are detected.
- Prefix overlap is detected.
- Delimiterless overlap is detected.
- Legacy broad-rule ambiguity is detected in static collision tests.
- Versioned examples do not collide because route identity includes namespace, version, feature, and action.
- Runtime router dispatch does not use `strings.Contains` or JavaScript-style `includes`.

## Ambiguous Legacy Patterns

- `announcement_yes` and `announcement_no`: reused by announcement confirmation and work override confirmation.
- `week_menu`, `hour_menu`, and `min_menu`: reused by cron and birthday flows.
- Raw alphanumeric IDs: can represent shop items, work names, quiz answers, or lottery IDs.
- Poll choice payloads remain supported only under the documented `poll_` prefix and bounded parser rule.
- Role-button suffixes are restricted to numeric-like IDs ending in `add` or `delete`, avoiding broad English substring matching.

## Router Integration

- Wave 2 router now supports normalized `RouteKey` for components and modals.
- `DefaultCustomIDParser` translates raw custom IDs into route keys before lookup.
- Slash commands still route by exact command name.
- Components and modals route by parsed route key.
- Parse failure returns `ErrBadInteractionID` and uses the safe responder error path.
- Existing pre-parsed Wave 2 component/modal key tests still work when no parser is configured.
- Middleware order, timeout, permission shell, and panic recovery remain deterministic.

## Test Cases Added

- Versioned encoder/parser tests for empty payload, token payload, length limit, invalid namespace/version/feature/action/payload, unsafe payload redaction, deterministic parse, and encode-parse round trip.
- Payload tests for empty/token/key-value/state payloads, overlong rejection, invalid character rejection, and secret-like payload rejection.
- Legacy component and modal golden tests from `testdata/customid`.
- Ambiguous and unknown legacy ID rejection tests.
- Collision analysis tests for duplicate, prefix, delimiterless, broad, and versioned non-collision cases.
- Router parser integration tests for component/modal route keys, parse failure safe errors, route not found, middleware order, timeout, permission metadata, and panic recovery.
- DiscordGo adapter tests for component IDs, select values, modal IDs, modal fields, and unsupported component types.
- Fuzz targets for component, versioned, and legacy parsers; benchmarks for versioned and legacy parser paths.

## Commands Run

- `go fmt ./...`: passed. The first sandboxed attempt could not access the Go build cache; rerun with approved cache access passed and formatted Wave 4 files.
- `go test ./...`: passed.
- `go vet ./...`: passed.
- `go build ./cmd/mhcat-bot`: passed.
- `go build ./cmd/mhcat-command-sync`: passed.
- `go build ./cmd/mhcat-mongo-audit`: passed.
- `go build ./cmd/mhcat-mongo-index`: passed.
- `make check`: passed.
- `go test ./internal/discord/customid ./internal/discord/interactions ./internal/adapters/discordgo`: passed.
- `go test ./internal/discord/customid -run Test`: passed.
- `go test ./internal/discord/interactions -run Test`: passed.
- `go test ./internal/adapters/discordgo -run Test`: passed.

## Known Limitations

- The legacy decoder intentionally does not globally route raw work names, quiz answers, or item IDs; those need message-scoped feature context in Wave 5.
- The parser provides only a `StateStore` interface for future state references. No Mongo state storage exists in Wave 4.
- The collision analyzer is static and conservative; it is not a runtime regex engine.
- Some medium/low-confidence legacy paths from Phase 1.5 remain documented but not enabled as global routes unless safely identifiable.

## Next Recommended Step

Begin Wave 5 only after selecting the first feature group. The safest first feature group is low-risk utility/status commands or a tightly scoped component-backed flow with golden parser fixtures already present. Feature work must add domain services, repository methods only where needed, and behavior parity tests before enabling Discord command registration.
