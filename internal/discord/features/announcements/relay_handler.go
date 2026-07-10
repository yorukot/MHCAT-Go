package announcements

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
)

const legacyRandomColorFallback = 0x53FF53

func (m Module) RelayHandler() events.Handler {
	return func(ctx context.Context, event events.Event) error {
		if event.Type != events.TypeMessageCreate {
			return nil
		}
		if event.IsBot || strings.TrimSpace(event.GuildID) == "" || strings.TrimSpace(event.ChannelID) == "" {
			return nil
		}
		if strings.TrimSpace(event.Content) == "" {
			return nil
		}
		if m.boundReader == nil || m.messages == nil {
			return nil
		}
		config, err := m.boundReader.GetBoundAnnouncement(ctx, event.GuildID, event.ChannelID)
		if errors.Is(err, ports.ErrBoundAnnouncementConfigMissing) {
			return nil
		}
		if err != nil {
			return err
		}
		color, ok := relayColor(config.Color)
		if !ok {
			return nil
		}
		message := ports.OutboundMessage{
			Content: config.Tag,
			Embeds: []ports.OutboundEmbed{{
				Title:         config.Title,
				Description:   event.Content,
				Color:         color,
				FooterText:    "來自" + event.UserTag + "的公告",
				FooterIconURL: event.AvatarURL,
			}},
			AllowedMentions: ports.AllowedMentions{},
		}
		if _, err := m.messages.SendMessage(ctx, event.ChannelID, message); err != nil {
			return err
		}
		if event.MessageID != "" {
			if err := m.messages.DeleteMessage(ctx, ports.MessageRef{ChannelID: event.ChannelID, MessageID: event.MessageID}); err != nil {
				return err
			}
		}
		return ctx.Err()
	}
}

func relayColor(value string) (int, bool) {
	if value == "Random" || value == "RANDOM" {
		return randomLegacyColor(), true
	}
	return domain.ParseLegacyColorValue(value)
}

func randomLegacyColor() int {
	max := big.NewInt(0xFFFFFF + 1)
	value, err := rand.Int(rand.Reader, max)
	if err != nil {
		return legacyRandomColorFallback
	}
	return int(value.Int64())
}
