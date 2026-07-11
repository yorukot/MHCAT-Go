package utility

import (
	"context"
	"errors"
	"fmt"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

var ErrUnsupportedInfoSubcommand = errors.New("unsupported info subcommand")

func (m Module) InfoHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		switch interaction.Subcommand {
		case "", "bot":
			return m.handleInfoBot(ctx, interaction, responder)
		case "shard":
			return m.handleInfoShard(ctx, interaction, responder)
		case "user":
			return m.handleInfoUser(ctx, interaction, responder)
		case "guild":
			return m.handleInfoGuild(ctx, interaction, responder)
		default:
			if err := responder.Reply(ctx, responses.Message{Content: "這個 info 子指令尚未在 Go 版實作。"}); err != nil {
				return err
			}
			return m.track(ctx, interaction, "info")
		}
	}
}

func (m Module) handleInfoBot(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
	if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
		return err
	}
	info, degraded := m.status.Info(ctx)
	msg := legacyInfoBotMessage(info)
	if degraded {
		msg = legacyInfoErrorMessage()
	}
	if err := responder.FollowUp(ctx, msg); err != nil {
		return err
	}
	return m.track(ctx, interaction, "info")
}

func (m Module) handleInfoShard(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
	if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
		return err
	}
	_, degraded := m.status.Info(ctx)
	msg := legacyInfoShardMessage()
	if degraded {
		msg = legacyInfoErrorMessage()
	}
	if err := responder.FollowUp(ctx, msg); err != nil {
		return err
	}
	return m.track(ctx, interaction, "info")
}

func (m Module) handleInfoUser(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
	if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
		return err
	}
	if m.discord == nil {
		if err := responder.EditOriginal(ctx, legacyInfoLookupErrorMessage()); err != nil {
			return err
		}
		return m.track(ctx, interaction, "info")
	}
	userID := interaction.Options["user"]
	if userID == "" {
		userID = interaction.Actor.UserID
	}
	if interaction.Actor.GuildID == "" || userID == "" {
		if err := responder.EditOriginal(ctx, legacyInfoLookupErrorMessage()); err != nil {
			return err
		}
		return m.track(ctx, interaction, "info")
	}
	info, err := m.discord.UserInfo(ctx, interaction.Actor.GuildID, userID)
	if err != nil {
		if err := responder.EditOriginal(ctx, legacyInfoLookupErrorMessage()); err != nil {
			return fmt.Errorf("respond to info user lookup failure: %w", err)
		}
		return m.track(ctx, interaction, "info")
	}
	if err := responder.EditOriginal(ctx, legacyInfoUserMessage(info)); err != nil {
		return err
	}
	return m.track(ctx, interaction, "info")
}

func (m Module) handleInfoGuild(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
	if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
		return err
	}
	if m.discord == nil || interaction.Actor.GuildID == "" {
		if err := responder.EditOriginal(ctx, legacyInfoLookupErrorMessage()); err != nil {
			return err
		}
		return m.track(ctx, interaction, "info")
	}
	info, err := m.discord.GuildInfo(ctx, interaction.Actor.GuildID)
	if err != nil {
		if err := responder.EditOriginal(ctx, legacyInfoLookupErrorMessage()); err != nil {
			return fmt.Errorf("respond to info guild lookup failure: %w", err)
		}
		return m.track(ctx, interaction, "info")
	}
	if err := responder.EditOriginal(ctx, legacyInfoGuildMessage(info)); err != nil {
		return err
	}
	return m.track(ctx, interaction, "info")
}

func (m Module) InfoBotRefreshHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		info, degraded := m.status.Info(ctx)
		if degraded {
			if err := responder.Reply(ctx, legacyInfoErrorMessage()); err != nil {
				return err
			}
			return m.track(ctx, interaction, "info")
		}
		if err := responder.UpdateMessage(ctx, legacyInfoBotRefreshMessage(info)); err != nil {
			return err
		}
		if err := responder.FollowUp(ctx, legacyInfoRefreshSuccessMessage()); err != nil {
			return err
		}
		return m.track(ctx, interaction, "info")
	}
}

func (m Module) InfoShardRefreshHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		info, degraded := m.status.Info(ctx)
		if degraded {
			if err := responder.Reply(ctx, legacyInfoErrorMessage()); err != nil {
				return err
			}
			return m.track(ctx, interaction, "info")
		}
		if err := responder.UpdateMessage(ctx, legacyInfoShardRefreshMessage(info)); err != nil {
			return err
		}
		return m.track(ctx, interaction, "info")
	}
}
