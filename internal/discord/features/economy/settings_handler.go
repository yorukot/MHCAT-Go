package economy

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	coreeconomy "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/economy"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	economySettingsManageMessagesPermission = int64(8192)
	economySettingsErrorColor               = 0xED4245
	economySettingsMaxError                 = "最高代幣設定數只能是999999999"
	economySettingsCooldownError            = "必須大於-1(0代表0:00重製)"
	economySettingsNonNegativeError         = "設定數必須大於或等於0"
	economySettingsUnknownError             = "很抱歉，出現了未知的錯誤，請重試!"
	economySettingsPermissionError          = "你需要有`訊息管理`才能使用此指令"
)

func (m Module) SettingsHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if m.settings.Repository == nil {
			return domain.ErrInvalidEconomySettings
		}
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		command, userErr, ok := economySettingsCommandFromInteraction(interaction)
		if !ok {
			return responder.EditOriginal(ctx, economySettingsErrorMessage(userErr))
		}
		if !interaction.Actor.HasPermission(economySettingsManageMessagesPermission) {
			return responder.EditOriginal(ctx, economySettingsErrorMessage(economySettingsPermissionError))
		}
		config, err := m.settings.Save(ctx, command)
		if err != nil {
			if errors.Is(err, domain.ErrInvalidEconomySettings) {
				return responder.EditOriginal(ctx, economySettingsErrorMessage(economySettingsUnknownError))
			}
			return err
		}
		if err := responder.EditOriginal(ctx, economySettingsSuccessMessage(config, command.SignCooldownHours, interaction.Actor.AvatarURL, m.randomColor())); err != nil {
			return err
		}
		return m.trackCommand(ctx, interaction, EconomySettingsCommandName)
	}
}

func economySettingsCommandFromInteraction(interaction interactions.Interaction) (domain.EconomySettingsCommand, string, bool) {
	gachaCost, ok := integerOption(interaction, "coin-raffle-takes")
	if !ok {
		return domain.EconomySettingsCommand{}, economySettingsUnknownError, false
	}
	cooldownHours, ok := integerOption(interaction, "check-in-cooldown-time")
	if !ok {
		return domain.EconomySettingsCommand{}, economySettingsUnknownError, false
	}
	signCoins, ok := integerOption(interaction, "check-in-give-coins")
	if !ok {
		return domain.EconomySettingsCommand{}, economySettingsUnknownError, false
	}
	notificationChannel := strings.TrimSpace(interaction.Options["notification-channel"])
	if notificationChannel == "" {
		return domain.EconomySettingsCommand{}, economySettingsUnknownError, false
	}
	xpMultiple, ok := numberOption(interaction, "level-up-multiply-amount")
	if !ok {
		return domain.EconomySettingsCommand{}, economySettingsUnknownError, false
	}
	switch {
	case gachaCost > coreeconomy.MaxLegacyCoinBalance || signCoins > coreeconomy.MaxLegacyCoinBalance:
		return domain.EconomySettingsCommand{}, economySettingsMaxError, false
	case cooldownHours < 0:
		return domain.EconomySettingsCommand{}, economySettingsCooldownError, false
	case gachaCost < 0 || signCoins < 0 || xpMultiple < 0:
		return domain.EconomySettingsCommand{}, economySettingsNonNegativeError, false
	}
	return domain.EconomySettingsCommand{
		GuildID:           interaction.Actor.GuildID,
		GachaCost:         gachaCost,
		SignCooldownHours: cooldownHours,
		SignCoins:         signCoins,
		NotificationID:    notificationChannel,
		XPMultiple:        xpMultiple,
	}, "", true
}

func economySettingsSuccessMessage(config domain.EconomyConfig, cooldownHours int64, avatarURL string, color int) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: fmt.Sprintf("<:money:997374193026994236>成功改變每次扭蛋及抽獎代幣數\n扭蛋所需代幣:`%d`\n簽到給予代幣數:`%d`\n等級提升給予倍數:`%s`\n每次簽到所需時間:`%d 小時`",
				config.GachaCost,
				config.SignCoins,
				strconv.FormatFloat(config.XPMultiple, 'f', -1, 64),
				cooldownHours,
			),
			Description: "通知頻道:" + channelMention(config.ChannelID),
			Footer: &responses.EmbedFooter{
				Text:    "MHCAT",
				IconURL: avatarURL,
			},
			Color: color,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func economySettingsErrorMessage(content string) responses.Message {
	if strings.TrimSpace(content) == "" {
		content = economySettingsUnknownError
	}
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:Discord_AnimatedNo:1015989839809757295> | " + content,
			Color: economySettingsErrorColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func integerOption(interaction interactions.Interaction, name string) (int64, bool) {
	if value, ok := interaction.CommandOptions[name]; ok && value.Type == interactions.CommandOptionInteger {
		return value.Int, true
	}
	parsed, err := strconv.ParseInt(strings.TrimSpace(interaction.Options[name]), 10, 64)
	return parsed, err == nil
}

func numberOption(interaction interactions.Interaction, name string) (float64, bool) {
	if value, ok := interaction.CommandOptions[name]; ok && value.Type == interactions.CommandOptionNumber {
		return value.Float, true
	}
	parsed, err := strconv.ParseFloat(strings.TrimSpace(interaction.Options[name]), 64)
	return parsed, err == nil
}

func channelMention(channelID string) string {
	channelID = strings.TrimSpace(channelID)
	if strings.HasPrefix(channelID, "<#") && strings.HasSuffix(channelID, ">") {
		return channelID
	}
	return "<#" + channelID + ">"
}
