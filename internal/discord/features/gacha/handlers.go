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
	gachaErrorColor          = 0xED4245
	legacyGachaFallbackGuild = "這個伺服器"
	discordEmbedFieldLimit   = 25
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
				return m.track(ctx, interaction)
			}
			return responder.EditOriginal(ctx, gachaErrorMessage("很抱歉，出現了未知的錯誤，請重試!"))
		}
		guildName := m.guildName(ctx, interaction.Actor.GuildID)
		if err := responder.EditOriginal(ctx, legacyPrizePoolMessage(result, guildName, interaction.Actor.UserTag, interaction.Actor.AvatarURL, m.color())); err != nil {
			return err
		}
		return m.track(ctx, interaction)
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

func (m Module) track(ctx context.Context, interaction interactions.Interaction) error {
	if m.usage == nil {
		return nil
	}
	return m.usage.TrackCommand(ctx, ports.UsageEvent{
		CommandName: GachaPrizeListCommandName,
		UserID:      interaction.Actor.UserID,
		GuildID:     interaction.Actor.GuildID,
		Feature:     "gacha-prize-list",
	})
}
