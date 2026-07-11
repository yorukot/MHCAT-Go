package notifications

import (
	"context"
	"math/rand/v2"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type DeliveryService struct {
	Repository  ports.AutoNotificationDeliveryRepository
	Messages    ports.DiscordMessagePort
	Channels    ports.DiscordCachedChannelReader
	RandomColor func() int
}

func (s DeliveryService) List(ctx context.Context) ([]domain.AutoNotificationSchedule, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if s.Repository == nil {
		return nil, domain.ErrInvalidAutoNotificationSchedule
	}
	return s.Repository.ListAutoNotificationDeliveries(ctx)
}

func (s DeliveryService) Deliver(ctx context.Context, guildID string, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.Repository == nil || s.Messages == nil || s.Channels == nil {
		return domain.ErrInvalidAutoNotificationSchedule
	}
	guildID = strings.TrimSpace(guildID)
	id = strings.TrimSpace(id)
	if err := domain.ValidateAutoNotificationDelete(guildID, id); err != nil {
		return err
	}
	schedule, err := s.Repository.GetAutoNotificationDelivery(ctx, guildID, id)
	if err != nil {
		return err
	}
	schedule = schedule.Normalized()
	if err := domain.ValidateAutoNotificationDelivery(schedule); err != nil {
		return err
	}
	if _, err := s.Channels.FindCachedChannelByID(ctx, schedule.GuildID, schedule.ChannelID); err != nil {
		return err
	}
	_, err = s.Messages.SendMessage(ctx, schedule.ChannelID, s.outbound(schedule.Message))
	return err
}

func (s DeliveryService) outbound(message domain.AutoNotificationMessage) ports.OutboundMessage {
	message = message.Normalized()
	out := ports.OutboundMessage{
		Content: message.Content,
		AllowedMentions: ports.AllowedMentions{
			ParseEveryone: true,
			ParseUsers:    true,
			ParseRoles:    true,
		},
	}
	if !message.HasEmbed() {
		return out
	}
	color := 0
	if message.EmbedColor == "Random" {
		color = s.randomColor()
	} else if parsed, ok := domain.ParseLegacyColorValue(message.EmbedColor); ok {
		color = parsed
	}
	out.Embeds = []ports.OutboundEmbed{{
		Title:       message.EmbedTitle,
		Description: message.EmbedDescription,
		Color:       color,
	}}
	return out
}

func (s DeliveryService) randomColor() int {
	if s.RandomColor != nil {
		return s.RandomColor() & 0xFFFFFF
	}
	return rand.IntN(1 << 24)
}
