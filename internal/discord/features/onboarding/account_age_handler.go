package onboarding

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/onboarding"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	permissionKickMembersBit = int64(2)
	accountAgeSuccessColor   = 0x53FF53
)

func (m Module) AccountAgeHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(permissionKickMembersBit) {
			return responder.EditOriginal(ctx, accountAgeErrorMessage("你需要有`踢出用戶`才能使用此指令"))
		}
		switch strings.TrimSpace(interaction.Subcommand) {
		case "小時數":
			hours, ok := integerOption(interaction, "小時數")
			if !ok || hours <= 0 {
				return responder.EditOriginal(ctx, accountAgeErrorMessage("不可為負數或0!!!"))
			}
			if _, err := m.accountAgeService.SetRequirement(ctx, interaction.Actor.GuildID, hours); err != nil {
				return responder.EditOriginal(ctx, accountAgeErrorFromError(err))
			}
			if err := responder.EditOriginal(ctx, accountAgeHoursSuccessMessage(hours)); err != nil {
				return err
			}
			return m.trackFeature(ctx, interaction, AccountAgeCommandName, "account-age-config")
		case "被踢出資訊頻道":
			channelID := firstOption(interaction, "頻道")
			if _, err := m.accountAgeService.SetLogChannel(ctx, interaction.Actor.GuildID, channelID); err != nil {
				return responder.EditOriginal(ctx, accountAgeMissingHoursError(err))
			}
			if err := responder.EditOriginal(ctx, accountAgeChannelSuccessMessage(channelID)); err != nil {
				return err
			}
			return m.trackFeature(ctx, interaction, AccountAgeCommandName, "account-age-config")
		case "創建時數刪除":
			if err := m.accountAgeService.DeleteConfig(ctx, interaction.Actor.GuildID); err != nil {
				return responder.EditOriginal(ctx, accountAgeErrorFromError(err))
			}
			if err := responder.EditOriginal(ctx, accountAgeDeleteSuccessMessage("已刪除帳號需創建時數所有設定")); err != nil {
				return err
			}
			return m.trackFeature(ctx, interaction, AccountAgeCommandName, "account-age-config")
		case "被踢出資訊頻道刪除":
			if err := m.accountAgeService.DeleteLogChannel(ctx, interaction.Actor.GuildID); err != nil {
				return responder.EditOriginal(ctx, accountAgeErrorFromError(err))
			}
			if err := responder.EditOriginal(ctx, accountAgeDeleteSuccessMessage("已刪除被踢出資訊頻道")); err != nil {
				return err
			}
			return m.trackFeature(ctx, interaction, AccountAgeCommandName, "account-age-config")
		default:
			return responder.EditOriginal(ctx, accountAgeErrorMessage("很抱歉，出現了未知的錯誤，請重試!"))
		}
	}
}

func accountAgeHoursSuccessMessage(hours int64) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "<a:green_tick:994529015652163614>群組防護系統",
			Description: "已為您設定必須創建帳號" + coreservice.AccountAgeRoundedDays(hours) + "天才能加入伺服器",
			Color:       accountAgeSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func accountAgeChannelSuccessMessage(channelID string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "<a:green_tick:994529015652163614>群組防護系統",
			Description: "已為您設定當未達創建時數時會在:\n<#" + strings.TrimSpace(channelID) + ">發送使用者資運",
			Color:       accountAgeSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func accountAgeDeleteSuccessMessage(description string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "<:trashbin:995991389043163257>群組防護系統",
			Description: description,
			Color:       accountAgeSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func accountAgeErrorFromError(err error) responses.Message {
	switch {
	case errors.Is(err, ports.ErrAccountAgeConfigMissing):
		return accountAgeErrorMessage("你還沒設定過喔!")
	case errors.Is(err, domain.ErrInvalidAccountAgeConfig):
		return accountAgeErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
	default:
		return accountAgeErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
	}
}

func accountAgeMissingHoursError(err error) responses.Message {
	if errors.Is(err, ports.ErrAccountAgeConfigMissing) {
		return accountAgeErrorMessage("你必須先設定`/帳號需創建時數 小時數`")
	}
	return accountAgeErrorFromError(err)
}

func accountAgeErrorMessage(content string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:Discord_AnimatedNo:1015989839809757295> | " + content,
			Color: joinRoleErrorColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func integerOption(interaction interactions.Interaction, name string) (int64, bool) {
	if option, ok := interaction.CommandOptions[name]; ok {
		if option.Int != 0 || option.String == "0" {
			return option.Int, true
		}
	}
	value := firstOption(interaction, name)
	if value == "" {
		return 0, false
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	return parsed, err == nil
}
