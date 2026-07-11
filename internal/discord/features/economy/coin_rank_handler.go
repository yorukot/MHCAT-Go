package economy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreeconomy "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/economy"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	coinRankFileName            = "user-info.png"
	coinRankFileContentType     = "image/png"
	coinRankMissingUserTitle    = "<a:Discord_AnimatedNo:1015989839809757295> | 找不到資料!請於幾分鐘後重試!"
	coinRankMissingUserColor    = 0xED4245
	coinRankMissingUsername     = "找不到該名使用者"
	legacyRankYearBackEmoji     = "<:lefft:1079286176332136480>"
	legacyRankPageBackEmoji     = "<:left:1079286186570436609>"
	legacyRankPageForwardEmoji  = "<:right:1079285288645447730>"
	legacyRankYearForwardEmoji  = "<:right_r:1079285582263500920>"
	legacyRankSpacerEmoji       = "<:__:1079291288748314655>"
	legacyRankTargetViewerEmoji = "<:aim:1079305123773284422>"
)

var (
	legacyCoinRankPageRe    = regexp.MustCompile(`^\[([0-9]{17,20})\]\{([0-9]+)\}coin_rank$`)
	legacyCoinRankPageAltRe = regexp.MustCompile(`^\[([0-9]{17,20})\]coin_rank \{([0-9]+)\}$`)
)

func (m Module) CoinRankHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if m.coinRank.Repository == nil {
			return domain.ErrInvalidCoinRankQuery
		}
		if err := responder.Reply(ctx, signLoadingMessage(interaction.Actor.AvatarURL)); err != nil {
			return err
		}
		message, err := m.coinRankMessage(ctx, interaction, interaction.Actor.UserID, 0)
		if err != nil {
			return err
		}
		if err := responder.EditOriginal(ctx, message); err != nil {
			return err
		}
		return m.trackCommand(ctx, interaction, CoinRankCommandName)
	}
}

func (m Module) CoinRankPageHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if m.coinRank.Repository == nil {
			return domain.ErrInvalidCoinRankQuery
		}
		viewerID, page, err := parseLegacyCoinRankPageRequest(interaction.CustomID)
		if err != nil {
			return err
		}
		if !m.coinRankUserExists(ctx, interaction.Actor.GuildID, viewerID) {
			return responder.Reply(ctx, coinRankMissingUserMessage())
		}
		message, err := m.coinRankMessage(ctx, interaction, viewerID, page)
		if err != nil {
			return err
		}
		return responder.UpdateMessage(ctx, message)
	}
}

func parseLegacyCoinRankPageRequest(raw string) (string, int, error) {
	matches := legacyCoinRankPageRe.FindStringSubmatch(raw)
	if matches == nil {
		matches = legacyCoinRankPageAltRe.FindStringSubmatch(raw)
	}
	if matches == nil {
		return "", 0, domain.ErrInvalidCoinRankQuery
	}
	page, err := strconv.Atoi(matches[2])
	if err != nil || page < 0 {
		return "", 0, domain.ErrInvalidCoinRankQuery
	}
	return matches[1], page, nil
}

func (m Module) coinRankMessage(ctx context.Context, interaction interactions.Interaction, viewerID string, page int) (responses.Message, error) {
	result, err := m.coinRank.Query(ctx, coreeconomy.CoinRankQuery{
		GuildID:  interaction.Actor.GuildID,
		ViewerID: viewerID,
		Page:     page,
	})
	if err != nil {
		return responses.Message{}, err
	}
	guild := m.lookupCoinRankGuild(ctx, interaction.Actor.GuildID)
	canvasEntries := make([]coinRankCanvasEntry, 0, len(result.Entries))
	for _, entry := range result.Entries {
		canvasEntries = append(canvasEntries, coinRankCanvasEntry{
			Rank:        entry.Rank,
			DisplayName: m.lookupCoinRankUsername(ctx, interaction.Actor.GuildID, entry.Balance.UserID),
			Coins:       entry.Balance.Coins,
		})
	}
	viewerRankText := "沒有資料"
	if result.ViewerHasBalance {
		viewerRankText = fmt.Sprintf("#%d", result.ViewerRank)
	}
	image, err := renderCoinRankPNG(coinRankCanvasView{
		GuildName:      guild.Name,
		GuildCreatedAt: guild.CreatedAt,
		GuildIconData:  fetchCoinRankGuildIcon(ctx, guild.IconURL),
		ViewerRankText: viewerRankText,
		Entries:        canvasEntries,
	})
	if err != nil {
		return responses.Message{}, err
	}
	return responses.Message{
		Files: []responses.File{{
			Name:        coinRankFileName,
			ContentType: coinRankFileContentType,
			Data:        image,
		}},
		Components:      coinRankComponents(result),
		AllowedMentions: &responses.AllowedMentions{},
	}, nil
}

func fetchCoinRankGuildIcon(ctx context.Context, iconURL string) []byte {
	parsed, err := url.Parse(strings.TrimSpace(iconURL))
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return nil
	}
	if strings.EqualFold(path.Ext(parsed.Path), ".gif") {
		parsed.Path = strings.TrimSuffix(parsed.Path, path.Ext(parsed.Path)) + ".png"
		parsed.RawPath = ""
	}
	reqCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(reqCtx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil
	}
	const maxIconBytes = 2 << 20
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, io.LimitReader(response.Body, maxIconBytes+1)); err != nil || buf.Len() > maxIconBytes {
		return nil
	}
	return buf.Bytes()
}

func (m Module) lookupCoinRankGuild(ctx context.Context, guildID string) ports.DiscordGuildInfo {
	if m.discord == nil || strings.TrimSpace(guildID) == "" {
		return ports.DiscordGuildInfo{ID: guildID, Name: guildID, CreatedAt: time.Unix(0, 0).UTC()}
	}
	info, err := m.discord.GuildInfo(ctx, guildID)
	if err != nil {
		return ports.DiscordGuildInfo{ID: guildID, Name: guildID, CreatedAt: time.Unix(0, 0).UTC()}
	}
	if strings.TrimSpace(info.Name) == "" {
		info.Name = guildID
	}
	return info
}

func (m Module) lookupCoinRankUsername(ctx context.Context, guildID string, userID string) string {
	if m.discord == nil || strings.TrimSpace(guildID) == "" || strings.TrimSpace(userID) == "" {
		return coinRankMissingUsername
	}
	info, err := m.discord.UserInfo(ctx, guildID, userID)
	if err != nil || strings.TrimSpace(info.Username) == "" {
		return coinRankMissingUsername
	}
	username := strings.TrimSpace(info.Username)
	discriminator := strings.TrimSpace(info.Discriminator)
	if discriminator != "" && discriminator != "0" {
		return username + "#" + discriminator
	}
	return username
}

func (m Module) coinRankUserExists(ctx context.Context, guildID string, userID string) bool {
	if m.discord == nil {
		return true
	}
	info, err := m.discord.UserInfo(ctx, guildID, userID)
	return err == nil && strings.TrimSpace(info.Username) != ""
}

func coinRankMissingUserMessage() responses.Message {
	return responses.Message{
		Ephemeral: true,
		Embeds: []responses.Embed{{
			Title: coinRankMissingUserTitle,
			Color: coinRankMissingUserColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func coinRankComponents(page coreeconomy.CoinRankPage) []responses.ComponentRow {
	components := []responses.ComponentRow{{
		Components: []responses.Component{
			legacyCoinRankButton(fmt.Sprintf("[%s]{%d}coin_rank", page.ViewerID, page.Page-10), "", legacyRankYearBackEmoji, responses.ButtonStyleSuccess, page.Page-10 < 0),
			legacyCoinRankButton(fmt.Sprintf("[%s]{%d}coin_rank", page.ViewerID, page.Page-1), "", legacyRankPageBackEmoji, responses.ButtonStyleSuccess, page.Page-1 == -1),
			legacyCoinRankButton("coin_rank", fmt.Sprintf("%d/%d", page.Page+1, page.TotalPages), "", responses.ButtonStyleSecondary, true),
			legacyCoinRankButton(fmt.Sprintf("[%s]{%d}coin_rank", page.ViewerID, page.Page+1), "", legacyRankPageForwardEmoji, responses.ButtonStyleSuccess, page.Page+1 >= page.TotalPages),
			legacyCoinRankButton(fmt.Sprintf("[%s]{%d}coin_rank", page.ViewerID, page.Page+10), "", legacyRankYearForwardEmoji, responses.ButtonStyleSuccess, page.Page+10 >= page.TotalPages),
		},
	}}
	if page.ViewerHasBalance {
		targetPage := page.ViewerRank / coreeconomy.CoinRankPageSize
		components = append(components, responses.ComponentRow{
			Components: []responses.Component{
				legacyCoinRankButton("coin_rank1", "", legacyRankSpacerEmoji, responses.ButtonStyleSecondary, true),
				legacyCoinRankButton("coin_rank2", "", legacyRankSpacerEmoji, responses.ButtonStyleSecondary, true),
				legacyCoinRankButton(fmt.Sprintf("[%s]coin_rank {%d}", page.ViewerID, targetPage), "", legacyRankTargetViewerEmoji, responses.ButtonStyleSecondary, false),
				legacyCoinRankButton("coin_rank4", "", legacyRankSpacerEmoji, responses.ButtonStyleSecondary, true),
				legacyCoinRankButton("coin_rank5", "", legacyRankSpacerEmoji, responses.ButtonStyleSecondary, true),
			},
		})
	}
	return components
}

func legacyCoinRankButton(customID string, label string, emoji string, style responses.ButtonStyle, disabled bool) responses.Component {
	return responses.Component{
		Type:     responses.ComponentTypeButton,
		CustomID: customID,
		Label:    label,
		Emoji:    emoji,
		Style:    style,
		Disabled: disabled,
	}
}
