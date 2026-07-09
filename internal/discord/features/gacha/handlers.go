package gacha

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/gacha"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	gachaErrorColor                  = 0xED4245
	gachaPrizeCreateSuccessColor     = 0x53FF53
	gachaPrizeEditSuccessColor       = 0x53FF53
	gachaPrizeDeleteSuccessColor     = 0x53FF53
	gachaManageMessagesPermissionBit = int64(8192)
	legacyDoneEmoji                  = "<a:green_tick:994529015652163614>"
	legacyGachaFallbackGuild         = "這個伺服器"
	discordEmbedFieldLimit           = 25
)

func (m Module) PrizeListHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		result, err := m.service.Query(ctx, interaction.Actor.GuildID)
		if err != nil {
			if errors.Is(err, ports.ErrGachaPrizePoolEmpty) {
				if editErr := responder.EditOriginal(ctx, gachaErrorMessage("目前獎池沒有任何獎品喔!")); editErr != nil {
					return editErr
				}
				return m.track(ctx, interaction, GachaPrizeListCommandName, "gacha-prize-list")
			}
			return responder.EditOriginal(ctx, gachaErrorMessage("很抱歉，出現了未知的錯誤，請重試!"))
		}
		guildName := m.guildName(ctx, interaction.Actor.GuildID)
		if err := responder.EditOriginal(ctx, legacyPrizePoolMessage(result, guildName, interaction.Actor.UserTag, interaction.Actor.AvatarURL, m.color())); err != nil {
			return err
		}
		return m.track(ctx, interaction, GachaPrizeListCommandName, "gacha-prize-list")
	}
}

func (m Module) PrizeDeleteHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(gachaManageMessagesPermissionBit) {
			return responder.EditOriginal(ctx, gachaErrorMessage("你需要有`訊息管理`才能使用此指令"))
		}
		prizeName := gachaStringOption(interaction, gachaPrizeNameOption)
		prize, err := m.deleteService.Delete(ctx, interaction.Actor.GuildID, prizeName)
		if err != nil {
			if errors.Is(err, ports.ErrGachaPrizeMissing) {
				return responder.EditOriginal(ctx, gachaErrorMessage("找不到這個獎品!"))
			}
			return responder.EditOriginal(ctx, gachaErrorMessage("很抱歉，出現了未知的錯誤，請重試!"))
		}
		if err := responder.EditOriginal(ctx, gachaPrizeDeleteSuccessMessage(prize)); err != nil {
			return err
		}
		return m.track(ctx, interaction, GachaPrizeDeleteCommandName, "gacha-prize-delete")
	}
}

func (m Module) PrizeCreateHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{Ephemeral: true}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(gachaManageMessagesPermissionBit) {
			return responder.EditOriginal(ctx, gachaErrorMessage("你沒有權限使用這個指令"))
		}
		prizeName := gachaStringOption(interaction, gachaPrizeNameOption)
		if len([]rune(prizeName)) > 200 {
			return responder.EditOriginal(ctx, gachaErrorMessage("獎品名稱不能這麼長!(需小於200)"))
		}
		prize := domain.GachaPrizeConfig{
			GuildID:    interaction.Actor.GuildID,
			Name:       prizeName,
			Code:       gachaStringOption(interaction, gachaPrizeCodeOption),
			Chance:     gachaNumberOption(interaction, gachaPrizeChanceOption),
			AutoDelete: gachaBoolOptionDefault(interaction, gachaPrizeAutoDeleteOption, true),
			Count:      gachaIntOptionDefault(interaction, gachaPrizeCountOption, 1, true),
			GiveCoin:   gachaIntOptionDefault(interaction, gachaPrizeGiveCoinOption, 0, false),
		}
		if prize.Count <= 0 {
			return responder.EditOriginal(ctx, gachaErrorMessage("獎品必須大於1"))
		}
		if err := m.createService.Create(ctx, prize); err != nil {
			switch {
			case errors.Is(err, ports.ErrGachaPrizePoolFull):
				return responder.EditOriginal(ctx, gachaErrorMessage("你的獎品數量已經過多了!!請先刪除部分獎品!"))
			case errors.Is(err, ports.ErrGachaPrizeExists):
				return responder.EditOriginal(ctx, gachaErrorMessage("獎品名稱重複，請將之前的刪除或等待被抽中!"))
			case errors.Is(err, domain.ErrInvalidGachaPrize):
				return responder.EditOriginal(ctx, gachaErrorMessage("很抱歉，出現了未知的錯誤，請重試!"))
			default:
				return responder.EditOriginal(ctx, gachaErrorMessage("很抱歉，出現了未知的錯誤，請重試!"))
			}
		}
		if err := responder.EditOriginal(ctx, gachaPrizeCreateSuccessMessage(prize)); err != nil {
			return err
		}
		return m.track(ctx, interaction, GachaPrizeCreateCommandName, "gacha-prize-create")
	}
}

func (m Module) PrizeEditHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{Ephemeral: true}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(gachaManageMessagesPermissionBit) {
			return responder.EditOriginal(ctx, gachaErrorMessage("你沒有權限使用這個指令"))
		}
		prizeName := gachaStringOption(interaction, gachaPrizeNameOption)
		if len([]rune(prizeName)) > 200 {
			return responder.EditOriginal(ctx, gachaErrorMessage("獎品名稱不能這麼長!(需小於200)"))
		}
		chance, chanceSet := gachaNumberOptionValue(interaction, gachaPrizeChanceOption)
		count := gachaIntOptionDefault(interaction, gachaPrizeCountOption, 1, true)
		if count <= 0 {
			return responder.EditOriginal(ctx, gachaErrorMessage("獎品必須大於1"))
		}
		display := domain.GachaPrizeConfig{
			GuildID:    interaction.Actor.GuildID,
			Name:       prizeName,
			Code:       gachaStringOption(interaction, gachaPrizeCodeOption),
			Chance:     chance,
			AutoDelete: gachaBoolOptionDefault(interaction, gachaPrizeAutoDeleteOption, true),
			Count:      count,
			GiveCoin:   gachaIntOptionDefault(interaction, gachaPrizeGiveCoinOption, 0, false),
		}
		edit := domain.GachaPrizeEdit{
			GuildID:    display.GuildID,
			Name:       display.Name,
			Code:       display.Code,
			Chance:     display.Chance,
			ChanceSet:  chanceSet,
			AutoDelete: display.AutoDelete,
			Count:      display.Count,
			GiveCoin:   display.GiveCoin,
		}
		if _, err := m.editService.Edit(ctx, edit); err != nil {
			switch {
			case errors.Is(err, ports.ErrGachaPrizeMissing):
				return responder.EditOriginal(ctx, gachaErrorMessage("找不到這個獎品!"))
			case errors.Is(err, domain.ErrInvalidGachaPrize):
				return responder.EditOriginal(ctx, gachaErrorMessage("很抱歉，出現了未知的錯誤，請重試!"))
			default:
				return responder.EditOriginal(ctx, gachaErrorMessage("很抱歉，出現了未知的錯誤，請重試!"))
			}
		}
		if err := responder.EditOriginal(ctx, gachaPrizeEditSuccessMessage(display, gachaChanceDisplayText(display.Chance, chanceSet))); err != nil {
			return err
		}
		return m.track(ctx, interaction, GachaPrizeEditCommandName, "gacha-prize-edit")
	}
}

func (m Module) guildName(ctx context.Context, guildID string) string {
	if m.discord == nil || strings.TrimSpace(guildID) == "" {
		return legacyGachaFallbackGuild
	}
	info, err := m.discord.GuildInfo(ctx, guildID)
	if err != nil || strings.TrimSpace(info.Name) == "" {
		return legacyGachaFallbackGuild
	}
	return info.Name
}

func legacyPrizePoolMessage(result domain.GachaPrizePool, guildName string, actorTag string, avatarURL string, color int) responses.Message {
	if strings.TrimSpace(guildName) == "" {
		guildName = legacyGachaFallbackGuild
	}
	if strings.TrimSpace(actorTag) == "" {
		actorTag = "你"
	}
	return responses.Message{
		Embeds:          legacyPrizePoolEmbeds(result, guildName, actorTag, avatarURL, color),
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func legacyPrizePoolEmbeds(result domain.GachaPrizePool, guildName string, actorTag string, avatarURL string, color int) []responses.Embed {
	fields := legacyPrizeFields(result.Prizes)
	if len(fields) == 0 {
		return []responses.Embed{{
			Title:       fmt.Sprintf("<:list:992002476360343602> 以下是%s的獎池", guildName),
			Color:       color,
			Description: legacyPrizePoolDescription(result),
			Footer: &responses.EmbedFooter{
				Text:    actorTag + "的查詢",
				IconURL: avatarURL,
			},
		}}
	}
	embeds := make([]responses.Embed, 0, (len(fields)+discordEmbedFieldLimit-1)/discordEmbedFieldLimit)
	for start := 0; start < len(fields); start += discordEmbedFieldLimit {
		end := start + discordEmbedFieldLimit
		if end > len(fields) {
			end = len(fields)
		}
		embed := responses.Embed{
			Title:  fmt.Sprintf("<:list:992002476360343602> 以下是%s的獎池", guildName),
			Color:  color,
			Fields: append([]responses.EmbedField(nil), fields[start:end]...),
		}
		if start == 0 {
			embed.Description = legacyPrizePoolDescription(result)
			embed.Footer = &responses.EmbedFooter{
				Text:    actorTag + "的查詢",
				IconURL: avatarURL,
			}
		} else {
			embed.Title = fmt.Sprintf("<:list:992002476360343602> 以下是%s的獎池 (%d)", guildName, len(embeds)+1)
		}
		embeds = append(embeds, embed)
	}
	return embeds
}

func legacyPrizePoolDescription(result domain.GachaPrizePool) string {
	gachaCost := result.Config.GachaCost
	signCoins := result.Config.SignCoins
	xpMultiple := result.Config.XPMultiple
	if !result.ConfigFound {
		gachaCost = coreservice.DefaultGachaCost
		signCoins = coreservice.DefaultSignCoins
		xpMultiple = coreservice.DefaultXPMultiple
	}
	return fmt.Sprintf("**<:money:997374193026994236> 扭蛋所需代幣:**`%d`個\n<:calendar:990254384812290048> **簽到給予代幣數:**`%d`個\n**<:levelup:990254382845157406> 等級提升給予倍數:**`%s`倍",
		gachaCost,
		signCoins,
		formatLegacyNumber(xpMultiple),
	)
}

func legacyPrizeFields(prizes []domain.GachaPrize) []responses.EmbedField {
	fields := make([]responses.EmbedField, 0, len(prizes))
	for _, prize := range prizes {
		fields = append(fields, responses.EmbedField{
			Name:   fmt.Sprintf("<:id:985950321975128094> 獎品名: %s", prize.Name),
			Value:  fmt.Sprintf("**<:dice:997374185322057799> 獎品中獎機率:** : `%s`%%\n'<:counter:994585977207140423> **獎品數量:** `%d`", formatLegacyNumber(prize.Chance), prize.Count),
			Inline: true,
		})
	}
	return fields
}

func formatLegacyNumber(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func gachaErrorMessage(content string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:Discord_AnimatedNo:1015989839809757295> | " + content,
			Color: gachaErrorColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func gachaPrizeDeleteSuccessMessage(prize domain.GachaPrize) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       legacyDoneEmoji + "成功刪除!",
			Description: "獎品名:" + strings.TrimSpace(prize.Name),
			Color:       gachaPrizeDeleteSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func gachaPrizeCreateSuccessMessage(prize domain.GachaPrizeConfig) responses.Message {
	code := strings.TrimSpace(prize.Code)
	if code == "" {
		code = "該獎品無代碼"
	}
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: legacyDoneEmoji + "設置成功",
			Color: gachaPrizeCreateSuccessColor,
			Fields: []responses.EmbedField{
				{Name: "<:id:985950321975128094> **獎品名:**", Value: prize.Name, Inline: true},
				{Name: "<:dice:997374185322057799> **獎品機率:**", Value: formatLegacyNumber(prize.Chance), Inline: true},
				{Name: "<:security:997374179257102396> **獎品代碼:**", Value: code, Inline: true},
				{Name: "<:counter:994585977207140423> **獎品數量:**", Value: fmt.Sprintf("%d個", prize.Count), Inline: true},
				{Name: "<:trashbin:995991389043163257> **自動刪除:**", Value: strconv.FormatBool(prize.AutoDelete), Inline: true},
				{Name: "<:givemoney:1019632789110399068> **給予代幣數:**", Value: fmt.Sprintf("%d個", prize.GiveCoin), Inline: true},
			},
		}},
		AllowedMentions: &responses.AllowedMentions{},
		Ephemeral:       true,
	}
}

func gachaPrizeEditSuccessMessage(prize domain.GachaPrizeConfig, chanceText string) responses.Message {
	code := strings.TrimSpace(prize.Code)
	if code == "" {
		code = "該獎品無代碼"
	}
	if strings.TrimSpace(chanceText) == "" {
		chanceText = "null"
	}
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: legacyDoneEmoji + "編輯成功成功",
			Color: gachaPrizeEditSuccessColor,
			Fields: []responses.EmbedField{
				{Name: "<:id:985950321975128094> **獎品名:**", Value: prize.Name, Inline: true},
				{Name: "<:dice:997374185322057799> **獎品機率:**", Value: chanceText, Inline: true},
				{Name: "<:security:997374179257102396> **獎品代碼:**", Value: code, Inline: true},
				{Name: "<:counter:994585977207140423> **獎品數量:**", Value: fmt.Sprintf("%d個", prize.Count), Inline: true},
				{Name: "<:trashbin:995991389043163257> **自動刪除:**", Value: strconv.FormatBool(prize.AutoDelete), Inline: true},
				{Name: "<:givemoney:1019632789110399068> **給予代幣數:**", Value: fmt.Sprintf("%d個", prize.GiveCoin), Inline: true},
			},
		}},
		AllowedMentions: &responses.AllowedMentions{},
		Ephemeral:       true,
	}
}

func gachaStringOption(interaction interactions.Interaction, names ...string) string {
	for _, name := range names {
		if value := strings.TrimSpace(interaction.Options[name]); value != "" {
			return value
		}
		if option, ok := interaction.CommandOptions[name]; ok {
			return strings.TrimSpace(option.String)
		}
	}
	return ""
}

func gachaNumberOption(interaction interactions.Interaction, name string) float64 {
	value, _ := gachaNumberOptionValue(interaction, name)
	return value
}

func gachaNumberOptionValue(interaction interactions.Interaction, name string) (float64, bool) {
	if option, ok := interaction.CommandOptions[name]; ok {
		return option.Float, true
	}
	value, ok := interaction.Options[name]
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, ok
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, true
	}
	return parsed, true
}

func gachaChanceDisplayText(value float64, set bool) string {
	if !set {
		return "null"
	}
	return formatLegacyNumber(value)
}

func gachaIntOptionDefault(interaction interactions.Interaction, name string, fallback int64, zeroIsFallback bool) int64 {
	if option, ok := interaction.CommandOptions[name]; ok {
		if zeroIsFallback && option.Int == 0 {
			return fallback
		}
		return option.Int
	}
	value := strings.TrimSpace(interaction.Options[name])
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	if zeroIsFallback && parsed == 0 {
		return fallback
	}
	return parsed
}

func gachaBoolOptionDefault(interaction interactions.Interaction, name string, fallback bool) bool {
	if option, ok := interaction.CommandOptions[name]; ok {
		return option.Bool
	}
	value := strings.TrimSpace(interaction.Options[name])
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func (m Module) track(ctx context.Context, interaction interactions.Interaction, commandName string, feature string) error {
	if m.usage == nil {
		return nil
	}
	return m.usage.TrackCommand(ctx, ports.UsageEvent{
		CommandName: commandName,
		UserID:      interaction.Actor.UserID,
		GuildID:     interaction.Actor.GuildID,
		Feature:     feature,
	})
}
