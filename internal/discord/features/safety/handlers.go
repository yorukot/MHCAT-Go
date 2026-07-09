package safety

import (
	"context"
	"fmt"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	permissionManageMessages = int64(8192)
	antiScamSuccessColor     = 0x57F287
	antiScamErrorColor       = 0xED4245
)

func (m Module) ToggleHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(permissionManageMessages) {
			return responder.EditOriginal(ctx, antiScamErrorMessage("你需要有`訊息管理`才能使用此指令"))
		}
		config, err := m.service.Toggle(ctx, interaction.Actor.GuildID)
		if err != nil {
			return responder.EditOriginal(ctx, antiScamUnknownError(err))
		}
		if err := responder.EditOriginal(ctx, antiScamSuccessMessage(config.Open)); err != nil {
			return err
		}
		return m.track(ctx, interaction)
	}
}

func antiScamSuccessMessage(open bool) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "<:fraudalert:1000408260777611355> 自動偵測詐騙連結",
			Description: fmt.Sprintf("您的防詐騙啟用狀態已改為:\n%t", open),
			Color:       antiScamSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func antiScamUnknownError(err error) responses.Message {
	_ = err
	return antiScamErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
}

func antiScamErrorMessage(content string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:Discord_AnimatedNo:1015989839809757295> | " + content,
			Color: antiScamErrorColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func (m Module) track(ctx context.Context, interaction interactions.Interaction) error {
	if m.usage == nil {
		return nil
	}
	return m.usage.TrackCommand(ctx, ports.UsageEvent{
		CommandName: AntiScamCommandName,
		UserID:      interaction.Actor.UserID,
		GuildID:     interaction.Actor.GuildID,
		Feature:     "anti-scam-config",
	})
}
