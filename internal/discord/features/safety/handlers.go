package safety

import (
	"context"
	"errors"
	"fmt"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
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
		config, err := m.configService.Toggle(ctx, interaction.Actor.GuildID)
		if err != nil {
			return responder.EditOriginal(ctx, antiScamUnknownError(err))
		}
		return responder.EditOriginal(ctx, antiScamSuccessMessage(config.Open))
	}
}

func (m Module) ReportHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		report, err := m.reportService.Report(ctx, firstOption(interaction, "網址"), interaction.Actor.UserID)
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrInvalidScamURLReport):
				return responder.EditOriginal(ctx, antiScamErrorMessage("你輸入的不是一個網址!"))
			case errors.Is(err, domain.ErrScamURLAlreadyReported):
				return responder.EditOriginal(ctx, antiScamErrorMessage("該網站已被回報過"))
			default:
				return responder.EditOriginal(ctx, antiScamUnknownError(err))
			}
		}
		return responder.EditOriginal(ctx, scamReportSuccessMessage(report.URL))
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

func scamReportSuccessMessage(rawURL string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "<:fraudalert:1000408260777611355> 自動偵測詐騙連結",
			Description: "成功回報" + rawURL,
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

func firstOption(interaction interactions.Interaction, names ...string) string {
	for _, name := range names {
		if value, ok := interaction.Options[name]; ok {
			return value
		}
		if option, ok := interaction.CommandOptions[name]; ok {
			return option.String
		}
	}
	return ""
}
