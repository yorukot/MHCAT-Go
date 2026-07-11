package onboarding

import (
	"context"
	cryptorand "crypto/rand"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type LeaveMessageService struct {
	Repository ports.LeaveMessageConfigRepository
}

type LeaveMessageDeliveryService struct {
	Repository ports.LeaveMessageConfigReader
	Messages   ports.DiscordMessagePort
	Channels   ports.DiscordCachedChannelReader
}

func (s LeaveMessageService) Prepare(ctx context.Context, guildID string, channelID string) (domain.LeaveMessageConfig, error) {
	if s.Repository == nil {
		return domain.LeaveMessageConfig{}, domain.ErrInvalidLeaveMessageConfig
	}
	config := domain.LeaveMessageConfig{
		GuildID:   strings.TrimSpace(guildID),
		ChannelID: strings.TrimSpace(channelID),
	}
	if err := config.ValidateChannel(); err != nil {
		return domain.LeaveMessageConfig{}, err
	}
	return s.Repository.PrepareLeaveMessageConfig(ctx, config.GuildID, config.ChannelID)
}

func (s LeaveMessageService) Save(ctx context.Context, config domain.LeaveMessageConfig) error {
	if s.Repository == nil {
		return domain.ErrInvalidLeaveMessageConfig
	}
	config.GuildID = strings.TrimSpace(config.GuildID)
	if err := config.ValidateContent(); err != nil {
		return err
	}
	return s.Repository.SaveLeaveMessageContent(ctx, config)
}

func (s LeaveMessageDeliveryService) SendOnLeave(ctx context.Context, event LeaveMemberEvent) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.Repository == nil || s.Messages == nil || s.Channels == nil {
		return domain.ErrInvalidLeaveMessageConfig
	}
	event.GuildID = strings.TrimSpace(event.GuildID)
	event.UserID = strings.TrimSpace(event.UserID)
	if event.GuildID == "" || event.UserID == "" {
		return nil
	}
	config, err := s.Repository.GetLeaveMessageConfig(ctx, event.GuildID)
	if err != nil {
		if errors.Is(err, ports.ErrLeaveMessageConfigMissing) {
			return nil
		}
		return err
	}
	if !leaveMessageConfigDeliverable(config) {
		return nil
	}
	cached, err := legacyDeliveryChannelCached(ctx, s.Channels, event.GuildID, config.ChannelID)
	if err != nil || !cached {
		return err
	}
	color, ok := leaveMessageDeliveryColor(config.Color)
	if !ok {
		return domain.ErrInvalidLeaveMessageConfig
	}
	_, err = s.Messages.SendMessage(ctx, config.ChannelID, ports.OutboundMessage{
		Embeds:          []ports.OutboundEmbed{leaveMessageDeliveryEmbed(config, event, color)},
		AllowedMentions: ports.AllowedMentions{},
	})
	return err
}

type LeaveMemberEvent struct {
	GuildID   string
	UserID    string
	Username  string
	UserTag   string
	AvatarURL string
	Now       time.Time
}

func leaveMessageConfigDeliverable(config domain.LeaveMessageConfig) bool {
	return strings.TrimSpace(config.GuildID) != "" &&
		strings.TrimSpace(config.ChannelID) != "" &&
		config.MessageContent != "" &&
		config.Title != "" &&
		config.Color != ""
}

func leaveMessageDeliveryEmbed(config domain.LeaveMessageConfig, event LeaveMemberEvent, color int) ports.OutboundEmbed {
	now := event.Now
	if now.IsZero() {
		now = time.Now()
	}
	return ports.OutboundEmbed{
		Title:        config.Title,
		Description:  replaceLeaveMessageDescriptionPlaceholders(config.MessageContent, event),
		Color:        color,
		ThumbnailURL: strings.TrimSpace(event.AvatarURL),
		Timestamp:    now,
	}
}

func replaceLeaveMessageDescriptionPlaceholders(value string, event LeaveMemberEvent) string {
	memberName := event.Username
	if memberName == "" {
		memberName = strings.TrimSpace(event.UserTag)
	}
	if memberName == "" {
		memberName = strings.TrimSpace(event.UserID)
	}
	out := value
	out = legacyStringReplace(out, "(MEMBERNAME)", memberName)
	out = legacyStringReplace(out, "(ID)", strings.TrimSpace(event.UserID))
	out = legacyStringReplace(out, "{ID}", strings.TrimSpace(event.UserID))
	out = legacyStringReplace(out, "{MEMBERNAME}", memberName)
	return out
}

func leaveMessageDeliveryColor(value string) (int, bool) {
	if legacyRandomColor(value) {
		return randomLeaveMessageColor(), true
	}
	return domain.ParseLegacyColorValue(value)
}

func randomLeaveMessageColor() int {
	n, err := cryptorand.Int(cryptorand.Reader, big.NewInt(0x1000000))
	if err != nil {
		return 0x5865F2
	}
	return int(n.Int64())
}
