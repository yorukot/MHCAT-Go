package safety

import (
	"context"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
)

const antiScamDeleteWarning = "<:trashbin:995991389043163257> | 此消息包含詐騙或釣魚連結，以自動刪除!\n"

func (m Module) RegisterEventRoutes(dispatcher *events.Dispatcher) {
	if dispatcher == nil || !m.messageEnabled {
		return
	}
	dispatcher.Register(events.TypeMessageCreate, m.MessageCreateHandler())
}

func (m Module) MessageCreateHandler() events.Handler {
	return func(ctx context.Context, event events.Event) error {
		if event.Type != events.TypeMessageCreate {
			return nil
		}
		guildID := strings.TrimSpace(event.GuildID)
		channelID := strings.TrimSpace(event.ChannelID)
		messageID := strings.TrimSpace(event.MessageID)
		if guildID == "" || channelID == "" || messageID == "" || strings.TrimSpace(event.Content) == "" {
			return nil
		}
		result, err := m.messageService.Scan(ctx, guildID, event.Content)
		if err != nil {
			return err
		}
		if !result.Delete {
			return ctx.Err()
		}
		if err := m.messages.DeleteMessage(ctx, ports.MessageRef{ChannelID: channelID, MessageID: messageID}); err != nil {
			return err
		}
		_, err = m.messages.SendMessage(ctx, channelID, ports.OutboundMessage{
			Content:         antiScamDeleteWarning,
			AllowedMentions: ports.AllowedMentions{},
		})
		if err != nil {
			return err
		}
		return ctx.Err()
	}
}
