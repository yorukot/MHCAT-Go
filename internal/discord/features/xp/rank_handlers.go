package xp

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/xp"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	rankFileName                = "user-info.png"
	rankFileContentType         = "image/png"
	rankMissingUserTitle        = "<a:Discord_AnimatedNo:1015989839809757295> | 找不到資料!請於幾分鐘後重試!"
	rankMissingUserColor        = 0xED4245
	rankMissingUsername         = "找不到該名使用者"
	rankLoadingAuthor           = "正在努力為您尋找資料!"
	rankLoadingIcon             = "https://media.discordapp.net/attachments/991337796960784424/1076582216127230053/6209-loading-online-circle.gif"
	rankLoadingFooter           = "MHCAT 帶給你最好的discord體驗!"
	rankDefaultAvatar           = "https://i.imgur.com/B91C90T.png"
	rankLoadingColor            = 0xFF5809
	legacyRankYearBackEmoji     = "<:lefft:1079286176332136480>"
	legacyRankPageBackEmoji     = "<:left:1079286186570436609>"
	legacyRankPageForwardEmoji  = "<:right:1079285288645447730>"
	legacyRankYearForwardEmoji  = "<:right_r:1079285582263500920>"
	legacyRankSpacerEmoji       = "<:__:1079291288748314655>"
	legacyRankTargetViewerEmoji = "<:aim:1079305123773284422>"
)

var (
	legacyXPRankPageRe    = regexp.MustCompile(`^\[([0-9]{17,20})\]\{([0-9]+)\}(text_rank|voice_rank)$`)
	legacyXPRankPageAltRe = regexp.MustCompile(`^\[([0-9]{17,20})\](text_rank|voice_rank) \{([0-9]+)\}$`)
)

func (m RankModule) TextHandler() interactions.Handler {
	return m.rankHandler(coreservice.RankKindText, TextXPRankCommandName)
}

func (m RankModule) VoiceHandler() interactions.Handler {
	return m.rankHandler(coreservice.RankKindVoice, VoiceXPRankCommandName)
}

func (m RankModule) TextPageHandler() interactions.Handler {
	return m.rankPageHandler(coreservice.RankKindText)
}

func (m RankModule) VoicePageHandler() interactions.Handler {
	return m.rankPageHandler(coreservice.RankKindVoice)
}

func (m RankModule) rankHandler(kind coreservice.RankKind, commandName string) interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if m.service.Repository == nil {
			return domain.ErrInvalidXPRankQuery
		}
		if err := responder.Reply(ctx, rankLoadingMessage(interaction.Actor.AvatarURL)); err != nil {
			return err
		}
		message, err := m.rankMessage(ctx, interaction, interaction.Actor.UserID, kind, 0)
		if err != nil {
			return err
		}
		if err := responder.EditOriginal(ctx, message); err != nil {
			return err
		}
		return m.track(ctx, interaction, commandName)
	}
}

func (m RankModule) rankPageHandler(expected coreservice.RankKind) interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if m.service.Repository == nil {
			return domain.ErrInvalidXPRankQuery
		}
		viewerID, page, kind, err := parseLegacyXPRankPageRequest(interaction.CustomID)
		if err != nil {
			return err
		}
		if kind != expected {
			return domain.ErrInvalidXPRankQuery
		}
		if !m.rankUserExists(ctx, interaction.Actor.GuildID, viewerID) {
			return responder.Reply(ctx, rankMissingUserMessage())
		}
		message, err := m.rankMessage(ctx, interaction, viewerID, kind, page)
		if err != nil {
			return err
		}
		return responder.UpdateMessage(ctx, message)
	}
}

func parseLegacyXPRankPageRequest(raw string) (string, int, coreservice.RankKind, error) {
	matches := legacyXPRankPageRe.FindStringSubmatch(raw)
	pageIndex := 2
	kindIndex := 3
	if matches == nil {
		matches = legacyXPRankPageAltRe.FindStringSubmatch(raw)
		pageIndex = 3
		kindIndex = 2
	}
	if matches == nil {
		return "", 0, "", domain.ErrInvalidXPRankQuery
	}
	page, err := strconv.Atoi(matches[pageIndex])
	if err != nil || page < 0 {
		return "", 0, "", domain.ErrInvalidXPRankQuery
	}
	kind := coreservice.RankKindText
	if matches[kindIndex] == "voice_rank" {
		kind = coreservice.RankKindVoice
	}
	return matches[1], page, kind, nil
}

func (m RankModule) rankMessage(ctx context.Context, interaction interactions.Interaction, viewerID string, kind coreservice.RankKind, page int) (responses.Message, error) {
	result, err := m.service.Query(ctx, coreservice.RankQuery{
		GuildID:  interaction.Actor.GuildID,
		ViewerID: viewerID,
		Kind:     kind,
		Page:     page,
	})
	if err != nil {
		return responses.Message{}, err
	}
	guild := m.lookupRankGuild(ctx, interaction.Actor.GuildID)
	canvasEntries := make([]rankCanvasEntry, 0, len(result.Entries))
	for _, entry := range result.Entries {
		canvasEntries = append(canvasEntries, rankCanvasEntry{
			Rank:        entry.Rank,
			DisplayName: m.lookupRankUsername(ctx, interaction.Actor.GuildID, entry.Profile.UserID),
			TotalXP:     entry.TotalXP,
		})
	}
	viewerRankText := "沒有資料!"
	if result.ViewerHasProfile {
		viewerRankText = strconv.Itoa(result.ViewerRank)
	}
	image, err := renderRankPNG(rankCanvasView{
		GuildName:      guild.Name,
		GuildCreatedAt: guild.CreatedAt,
		ViewerRankText: viewerRankText,
		Title:          rankCanvasTitle(kind),
		Entries:        canvasEntries,
	})
	if err != nil {
		return responses.Message{}, err
	}
	return responses.Message{
		Files: []responses.File{{
			Name:        rankFileName,
			ContentType: rankFileContentType,
			Data:        image,
		}},
		Components:      rankComponents(result),
		AllowedMentions: &responses.AllowedMentions{},
	}, nil
}

func rankCanvasTitle(kind coreservice.RankKind) string {
	if kind == coreservice.RankKindVoice {
		return "語音經驗排行榜"
	}
	return "聊天經驗排行榜"
}

func rankLegacyKind(kind coreservice.RankKind) string {
	if kind == coreservice.RankKindVoice {
		return "voice_rank"
	}
	return "text_rank"
}

func (m RankModule) lookupRankGuild(ctx context.Context, guildID string) ports.DiscordGuildInfo {
	if m.guilds == nil || strings.TrimSpace(guildID) == "" {
		return ports.DiscordGuildInfo{ID: guildID, Name: guildID, CreatedAt: time.Unix(0, 0).UTC()}
	}
	info, err := m.guilds.GuildInfo(ctx, guildID)
	if err != nil {
		return ports.DiscordGuildInfo{ID: guildID, Name: guildID, CreatedAt: time.Unix(0, 0).UTC()}
	}
	if strings.TrimSpace(info.Name) == "" {
		info.Name = guildID
	}
	return info
}

func (m RankModule) lookupRankUsername(ctx context.Context, guildID string, userID string) string {
	if m.guilds == nil || strings.TrimSpace(guildID) == "" || strings.TrimSpace(userID) == "" {
		return rankMissingUsername
	}
	info, err := m.guilds.UserInfo(ctx, guildID, userID)
	if err != nil || strings.TrimSpace(info.Username) == "" {
		return rankMissingUsername
	}
	return info.Username
}

func (m RankModule) rankUserExists(ctx context.Context, guildID string, userID string) bool {
	if m.guilds == nil {
		return true
	}
	info, err := m.guilds.UserInfo(ctx, guildID, userID)
	return err == nil && strings.TrimSpace(info.Username) != ""
}

func rankLoadingMessage(avatarURL string) responses.Message {
	if strings.TrimSpace(avatarURL) == "" {
		avatarURL = rankDefaultAvatar
	}
	return responses.Message{
		Embeds: []responses.Embed{{
			Author: &responses.EmbedAuthor{
				Name:    rankLoadingAuthor,
				IconURL: rankLoadingIcon,
			},
			Footer: &responses.EmbedFooter{
				Text:    rankLoadingFooter,
				IconURL: avatarURL,
			},
			Color: rankLoadingColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func rankMissingUserMessage() responses.Message {
	return responses.Message{
		Ephemeral: true,
		Embeds: []responses.Embed{{
			Title: rankMissingUserTitle,
			Color: rankMissingUserColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func rankComponents(page coreservice.RankPage) []responses.ComponentRow {
	kind := rankLegacyKind(page.Kind)
	components := []responses.ComponentRow{{
		Components: []responses.Component{
			legacyRankButton(fmt.Sprintf("[%s]{%d}%s", page.ViewerID, page.Page-10, kind), "", legacyRankYearBackEmoji, responses.ButtonStyleSuccess, page.Page-10 < 0),
			legacyRankButton(fmt.Sprintf("[%s]{%d}%s", page.ViewerID, page.Page-1, kind), "", legacyRankPageBackEmoji, responses.ButtonStyleSuccess, page.Page-1 == -1),
			legacyRankButton("text_rank", fmt.Sprintf("%d/%d", page.Page+1, page.TotalPages), "", responses.ButtonStyleSecondary, true),
			legacyRankButton(fmt.Sprintf("[%s]{%d}%s", page.ViewerID, page.Page+1, kind), "", legacyRankPageForwardEmoji, responses.ButtonStyleSuccess, page.Page+1 >= page.TotalPages),
			legacyRankButton(fmt.Sprintf("[%s]{%d}%s", page.ViewerID, page.Page+10, kind), "", legacyRankYearForwardEmoji, responses.ButtonStyleSuccess, page.Page+10 >= page.TotalPages),
		},
	}}
	if page.ViewerHasProfile {
		targetPage := page.ViewerRank / coreservice.RankPageSize
		components = append(components, responses.ComponentRow{
			Components: []responses.Component{
				legacyRankButton("text_rank1", "", legacyRankSpacerEmoji, responses.ButtonStyleSecondary, true),
				legacyRankButton("text_rank2", "", legacyRankSpacerEmoji, responses.ButtonStyleSecondary, true),
				legacyRankButton(fmt.Sprintf("[%s]%s {%d}", page.ViewerID, kind, targetPage), "", legacyRankTargetViewerEmoji, responses.ButtonStyleSecondary, false),
				legacyRankButton("text_rank4", "", legacyRankSpacerEmoji, responses.ButtonStyleSecondary, true),
				legacyRankButton("text_rank5", "", legacyRankSpacerEmoji, responses.ButtonStyleSecondary, true),
			},
		})
	}
	return components
}

func legacyRankButton(customID string, label string, emoji string, style responses.ButtonStyle, disabled bool) responses.Component {
	return responses.Component{
		Type:     responses.ComponentTypeButton,
		CustomID: customID,
		Label:    label,
		Emoji:    emoji,
		Style:    style,
		Disabled: disabled,
	}
}

func (m RankModule) track(ctx context.Context, interaction interactions.Interaction, commandName string) error {
	if m.usage == nil {
		return nil
	}
	return m.usage.TrackCommand(ctx, ports.UsageEvent{
		CommandName: commandName,
		UserID:      interaction.Actor.UserID,
		GuildID:     interaction.Actor.GuildID,
		Feature:     "xp-rank",
	})
}
