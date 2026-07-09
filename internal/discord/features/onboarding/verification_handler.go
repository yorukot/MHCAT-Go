package onboarding

import (
	"context"
	"errors"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

func (m Module) VerificationSetHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(permissionManageMessages) {
			return responder.EditOriginal(ctx, verificationErrorMessage("你需要有`訊息管理`才能使用此指令"))
		}
		config := domain.VerificationConfig{
			GuildID:        interaction.Actor.GuildID,
			RoleID:         firstOption(interaction, "身分組"),
			RenameTemplate: firstOption(interaction, "改名"),
		}
		if err := m.verificationService.Save(ctx, config); err != nil {
			return responder.EditOriginal(ctx, verificationErrorFromError(err))
		}
		if err := responder.EditOriginal(ctx, verificationSuccessMessage(config.RoleID, config.RenameTemplate)); err != nil {
			return err
		}
		return m.trackFeature(ctx, interaction, VerificationSetCommandName, "verification-config")
	}
}

func verificationSuccessMessage(roleID string, renameTemplate string) responses.Message {
	name := strings.TrimSpace(renameTemplate)
	if name == "" {
		name = "null"
	}
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "<a:green_tick:994529015652163614> 設置成功!",
			Color:       joinRoleSuccessColor,
			Description: "<:roleplaying:985945121264635964>身分組: <@&" + strings.TrimSpace(roleID) + ">!\n <:id:985950321975128094>改名為:" + name,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func verificationErrorFromError(err error) responses.Message {
	switch {
	case errors.Is(err, ports.ErrDiscordRoleNotAssignable):
		return verificationErrorMessage("我沒有權限為大家增加這個身分組，請將我的身分組位階調高")
	case errors.Is(err, domain.ErrInvalidVerificationConfig):
		return verificationErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
	default:
		return verificationErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
	}
}

func verificationErrorMessage(content string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:Discord_AnimatedNo:1015989839809757295> | " + content,
			Color: joinRoleErrorColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}
