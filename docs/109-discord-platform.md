# Shared Discord Platform Contract

Status: parity/adapter-audited against active legacy Gateway listeners and Discord REST calls, current DiscordGo session/intents, interaction adapter/responder, Gateway event mapper/dispatcher, side-effect ports, cache readers, audit-log mapping, app wiring, feature contracts, and race coverage. Live Discord cache, hierarchy, rate-limit, and intent smoke remain required.

## Gateway And Intents

Guilds is the only default intent. Guild Members, Guild Messages, Message Content, Guild Message Reactions, and Voice States are independently opt-in and config validation rejects each feature unless Gateway and its required intents are enabled. This intentionally narrows the legacy always-on privileged intent set.

Gateway handlers map Discord events into typed message, reaction, member, channel, and voice events before feature dispatch. Event feature gates are independent, so config-only commands do not silently enable message/member/voice ownership. Handler removal and dispatcher shutdown are part of the runtime lifecycle.

## Interaction Responses

The interaction adapter preserves command/subcommand/options, resolved user/role/channel metadata, actor permissions, guild/channel/message identity, custom IDs, modal fields, and requester avatars. The responder supports reply, defer, deferred update, modal, message update, original edit, follow-up create/edit/delete, attachments, components, embeds, ephemeral flags, and explicit attachment clearing.

Allowed mentions default to an empty parse set for both interactions and ordinary side effects. Features must explicitly allow reviewed user/role/reply mentions. Response-state tracking prevents generic errors after a successful ACK and preserves each feature's public/ephemeral lifecycle.

## Side-Effect Ports

The shared DiscordGo client implements feature-owned ports for channel/DM send and edit/delete, typing, message fetch/history/bulk delete, reactions, guild/member/role/channel inspection, role add/remove, nickname, kick/ban, voice move/disconnect, channel create/edit/delete and permission overwrites, member counts, bot/guild info, and audit-log queries. Legacy cache-dependent behavior and hierarchy decisions remain feature-specific and are covered in each parity contract.

DiscordGo owns REST rate-limit coordination and returns API errors to feature services. Go generally awaits required writes and marks explicitly optional DM/role/notification effects best-effort according to feature contracts instead of reproducing unobserved promise rejection. No adapter retries ambiguous application-level writes on its own.

## Cache And Compatibility Boundaries

The supported deployment is one session/shard. Cache-derived guild/member/role/channel values can differ from discord.js during cold start, partial events, or privileged-intent gaps. Config validation, fail-closed feature checks, and live staging smoke are required before mutation features. Multi-session cache aggregation is outside the current deployment contract.

Discord custom IDs and command registration are covered by [12-component-modal-grammar.md](12-component-modal-grammar.md) and [108-command-registration-dispatch.md](108-command-registration-dispatch.md). Startup/shutdown/shard scope is covered by [107-runtime-lifecycle.md](107-runtime-lifecycle.md).

## Verification

```bash
go test ./internal/adapters/discordgo ./internal/discord/events \
  ./internal/discord/interactions ./internal/discord/responses ./internal/app
go test -race ./internal/adapters/discordgo ./internal/discord/events ./internal/app
go vet ./...
```

Coverage includes minimal/explicit intents, event registration/removal and payload mapping, interaction option resolution, response payload conversion, mention policy, files/components/modals/follow-ups, side-effect request shapes, hierarchy/cache reads, audit-log actor/channel mapping, API errors, and app wiring.

## Staging And Rollback

Enable only the intents required by the feature under test. Use disposable channels, roles, members, messages, and voice rooms; verify cache-ready success plus missing/stale/hierarchy/API-failure cases and mention behavior. Observe Discord rate limits and confirm no duplicate Node/Go event owner.

Rollback by disabling feature/event gates, gracefully stopping Go, and only then restoring Node handlers. Remove staging artifacts manually using recorded IDs. Do not infer a successful side effect from a Discord interaction response when the feature contract marks that side effect best-effort.

Production remains gated on live Gateway/cache smoke, privileged-intent review, role/channel hierarchy checks, rate-limit observation, single-session capacity, and exclusive feature ownership.
