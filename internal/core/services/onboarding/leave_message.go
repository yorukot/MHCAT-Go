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
	config.Color = strings.TrimSpace(config.Color)
	if err := config.ValidateContent(); err != nil {
		return err
	}
	return s.Repository.SaveLeaveMessageContent(ctx, config)
}

func (s LeaveMessageDeliveryService) SendOnLeave(ctx context.Context, event LeaveMemberEvent) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.Repository == nil || s.Messages == nil {
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
	_, err = s.Messages.SendMessage(ctx, config.ChannelID, ports.OutboundMessage{
		Embeds:          []ports.OutboundEmbed{leaveMessageDeliveryEmbed(config, event)},
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
		strings.TrimSpace(config.Color) != ""
}

func leaveMessageDeliveryEmbed(config domain.LeaveMessageConfig, event LeaveMemberEvent) ports.OutboundEmbed {
	now := event.Now
	if now.IsZero() {
		now = time.Now()
	}
	return ports.OutboundEmbed{
		Title:        config.Title,
		Description:  replaceLeaveMessageDescriptionPlaceholders(config.MessageContent, event),
		Color:        leaveMessageDeliveryColor(config.Color),
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

func leaveMessageDeliveryColor(value string) int {
	if strings.EqualFold(strings.TrimSpace(value), "Random") {
		return randomLeaveMessageColor()
	}
	if parsed, ok := domain.ParseLegacyColorValue(value); ok {
		return parsed
	}
	return 0xED4245
}

func randomLeaveMessageColor() int {
	n, err := cryptorand.Int(cryptorand.Reader, big.NewInt(0x1000000))
	if err != nil {
		return 0x5865F2
	}
	return int(n.Int64())
}
