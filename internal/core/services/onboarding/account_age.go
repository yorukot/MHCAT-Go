package onboarding

import (
	"context"
	cryptorand "crypto/rand"
	"errors"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

const accountAgeKickReason = "你的創建時數低於管理員所設定的時數 Your creation time is lower than the time set by the administrator"

type AccountAgeConfigService struct {
	Repository ports.AccountAgeConfigRepository
}

type AccountAgePolicyService struct {
	Repository     ports.AccountAgeConfigReader
	DirectMessages ports.DiscordDirectMessagePort
	Members        ports.DiscordMemberPort
	Messages       ports.DiscordMessagePort
	Guilds         ports.DiscordInfoProvider
	Clock          ports.Clock
}

type AccountAgeMemberEvent struct {
	GuildID          string
	GuildName        string
	UserID           string
	UserTag          string
	AvatarURL        string
	AccountCreatedAt time.Time
	Now              time.Time
}

type AccountAgePolicyResult struct {
	Matched bool
	Kicked  bool
	Logged  bool
}

func (s AccountAgeConfigService) SetRequirement(ctx context.Context, guildID string, hours int64) (domain.AccountAgeConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.AccountAgeConfig{}, err
	}
	if s.Repository == nil {
		return domain.AccountAgeConfig{}, domain.ErrInvalidAccountAgeConfig
	}
	if hours <= 0 || hours > math.MaxInt64/3600 {
		return domain.AccountAgeConfig{}, domain.ErrInvalidAccountAgeConfig
	}
	return s.Repository.SaveAccountAgeRequirement(ctx, guildID, hours*3600)
}

func (s AccountAgeConfigService) SetLogChannel(ctx context.Context, guildID string, channelID string) (domain.AccountAgeConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.AccountAgeConfig{}, err
	}
	if s.Repository == nil {
		return domain.AccountAgeConfig{}, domain.ErrInvalidAccountAgeConfig
	}
	return s.Repository.SetAccountAgeLogChannel(ctx, guildID, channelID)
}

func (s AccountAgeConfigService) DeleteConfig(ctx context.Context, guildID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.Repository == nil {
		return domain.ErrInvalidAccountAgeConfig
	}
	return s.Repository.DeleteAccountAgeConfig(ctx, guildID)
}

func (s AccountAgeConfigService) DeleteLogChannel(ctx context.Context, guildID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.Repository == nil {
		return domain.ErrInvalidAccountAgeConfig
	}
	return s.Repository.DeleteAccountAgeLogChannel(ctx, guildID)
}

func (s AccountAgePolicyService) GateMemberAdd(ctx context.Context, event AccountAgeMemberEvent) (AccountAgePolicyResult, error) {
	if err := ctx.Err(); err != nil {
		return AccountAgePolicyResult{}, err
	}
	result := AccountAgePolicyResult{}
	if s.Repository == nil || s.Members == nil {
		return result, domain.ErrInvalidAccountAgeConfig
	}
	event.GuildID = strings.TrimSpace(event.GuildID)
	event.UserID = strings.TrimSpace(event.UserID)
	if event.GuildID == "" || event.UserID == "" {
		return result, domain.ErrInvalidAccountAgeConfig
	}
	config, err := s.Repository.GetAccountAgeConfig(ctx, event.GuildID)
	if errors.Is(err, ports.ErrAccountAgeConfigMissing) {
		return result, nil
	}
	if errors.Is(err, domain.ErrInvalidAccountAgeConfig) {
		return result, nil
	}
	if err != nil {
		return result, err
	}
	if err := config.Validate(); err != nil {
		if errors.Is(err, domain.ErrInvalidAccountAgeConfig) {
			return result, nil
		}
		return result, err
	}
	event, err = s.enrichMemberEvent(ctx, event)
	if err != nil {
		return result, err
	}
	if event.AccountCreatedAt.IsZero() {
		return result, domain.ErrInvalidAccountAgeConfig
	}
	now := event.Now
	if now.IsZero() {
		if s.Clock != nil {
			now = s.Clock.Now()
		} else {
			now = time.Now()
		}
	}
	if now.Sub(event.AccountCreatedAt).Seconds() >= float64(config.RequiredSeconds) {
		return result, nil
	}
	result.Matched = true
	if s.DirectMessages != nil {
		if _, dmErr := s.DirectMessages.SendDirectMessage(ctx, event.UserID, accountAgeDirectMessage(config, event)); dmErr != nil {
			if err := ctx.Err(); err != nil {
				return result, err
			}
		}
	}
	if err := s.Members.KickMember(ctx, event.GuildID, event.UserID, accountAgeKickReason); err != nil {
		return result, err
	}
	result.Kicked = true
	if strings.TrimSpace(config.ChannelID) == "" || s.Messages == nil {
		return result, ctx.Err()
	}
	if _, err := s.Messages.SendMessage(ctx, config.ChannelID, accountAgeLogMessage(event)); err != nil {
		return result, err
	}
	result.Logged = true
	return result, ctx.Err()
}

func (s AccountAgePolicyService) enrichMemberEvent(ctx context.Context, event AccountAgeMemberEvent) (AccountAgeMemberEvent, error) {
	if s.Guilds == nil {
		return event, nil
	}
	if event.AccountCreatedAt.IsZero() {
		user, err := s.Guilds.UserInfo(ctx, event.GuildID, event.UserID)
		if err != nil {
			return event, err
		}
		event.AccountCreatedAt = user.CreatedAt
		if strings.TrimSpace(event.AvatarURL) == "" {
			event.AvatarURL = user.AvatarURL
		}
	}
	if strings.TrimSpace(event.GuildName) == "" {
		guild, err := s.Guilds.GuildInfo(ctx, event.GuildID)
		if err != nil {
			if err := ctx.Err(); err != nil {
				return event, err
			}
			return event, nil
		}
		event.GuildName = guild.Name
	}
	return event, nil
}

func AccountAgeRoundedDays(hours int64) string {
	rounded := math.Round((float64(hours)/24)*10) / 10
	return strconv.FormatFloat(rounded, 'f', -1, 64)
}

func AccountAgeHoursString(seconds int64) string {
	return strconv.FormatFloat(float64(seconds)/3600, 'f', -1, 64)
}

func accountAgeDirectMessage(config domain.AccountAgeConfig, event AccountAgeMemberEvent) ports.OutboundMessage {
	guildName := strings.TrimSpace(event.GuildName)
	if guildName == "" {
		guildName = strings.TrimSpace(event.GuildID)
	}
	if guildName == "" {
		guildName = "此伺服器"
	}
	return ports.OutboundMessage{
		Embeds: []ports.OutboundEmbed{{
			Title:         "<a:warn:1000814885506129990> | 帳號創建時數過低!",
			Description:   "由於你帳號創建時數低於該伺服器所設定的時數\n已將您踢出`" + guildName + "`，如有問題請詢問該服服主\n\nSince your account creation hours are lower than the hours set by the server\nyou have been kicked out of `" + guildName + "` .\nIf you have any questions, please ask the server owner",
			Color:         0xEA0000,
			FooterText:    "管理員所設定時間: " + AccountAgeHoursString(config.RequiredSeconds) + " 小時",
			FooterIconURL: strings.TrimSpace(event.AvatarURL),
		}},
		AllowedMentions: ports.AllowedMentions{},
	}
}

func accountAgeLogMessage(event AccountAgeMemberEvent) ports.OutboundMessage {
	return ports.OutboundMessage{
		Embeds: []ports.OutboundEmbed{{
			Title: "低於管理員所設定的時數",
			Fields: []ports.OutboundEmbedField{{
				Name:  "該使用者帳號創建時間:",
				Value: "<t:" + strconv.FormatInt(legacyRoundedUnix(event.AccountCreatedAt), 10) + ">",
			}},
			ThumbnailURL:  strings.TrimSpace(event.AvatarURL),
			FooterText:    "BAN:" + strings.TrimSpace(event.UserTag),
			FooterIconURL: strings.TrimSpace(event.AvatarURL),
			Color:         randomAccountAgeEmbedColor(),
		}},
		AllowedMentions: ports.AllowedMentions{},
	}
}

func legacyRoundedUnix(value time.Time) int64 {
	if value.IsZero() {
		return 0
	}
	return (value.UnixMilli() + 500) / 1000
}

func randomAccountAgeEmbedColor() int {
	n, err := cryptorand.Int(cryptorand.Reader, big.NewInt(0x1000000))
	if err != nil {
		return 0x5865F2
	}
	return int(n.Int64())
}
