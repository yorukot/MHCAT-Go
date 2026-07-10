package autochat

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/autochat"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
)

const (
	LegacyAutoChatPaidResponseDelay = 10 * time.Second
	legacyAutoChatUnsafeInputDelay  = 4 * time.Second
	legacyAutoChatBusyWarningDelay  = 2 * time.Second
)

const (
	legacyAutoChatUnsafeInputMessage  = "為防止伺服器招到tag攻擊，請勿在與機器人聊天時含有@"
	legacyAutoChatBusyMessage         = "一次只能傳送一個消息，請等待機器人回復完成後在繼續!"
	legacyAutoChatUnsafeOutputMessage = "由於chatGPT回傳回來的訊息含有@，為防止遭到tag攻擊，已自動迴避該則消息!"
)

type PaidRuntimeModule struct {
	service  coreservice.PaidHandoffService
	messages ports.DiscordMessagePort
	typing   ports.DiscordTypingPort
	clock    ports.Clock
	wait     func(context.Context, time.Duration) error
}

func NewPaidRuntimeModule(configs ports.AutoChatConfigReader, balances ports.BalanceRepository, handoff ports.AutoChatPaidRepository, messages ports.DiscordMessagePort, typing ports.DiscordTypingPort, clock ports.Clock) (PaidRuntimeModule, error) {
	if messages == nil {
		return PaidRuntimeModule{}, errors.New("paid autochat message port is required")
	}
	service, err := coreservice.NewPaidHandoffService(configs, balances, handoff)
	if err != nil {
		return PaidRuntimeModule{}, err
	}
	if clock == nil {
		clock = ports.SystemClock{}
	}
	return PaidRuntimeModule{
		service:  service,
		messages: messages,
		typing:   typing,
		clock:    clock,
		wait:     waitForAutoChatReply,
	}, nil
}

func (m PaidRuntimeModule) RegisterEventRoutes(dispatcher *events.Dispatcher) {
	if dispatcher == nil || m.messages == nil {
		return
	}
	dispatcher.Register(events.TypeMessageCreate, m.MessageCreateHandler())
}

func (m PaidRuntimeModule) MessageCreateHandler() events.Handler {
	return func(ctx context.Context, event events.Event) error {
		if event.Type != events.TypeMessageCreate || event.IsBot || event.Member != nil && event.Member.IsBot {
			return nil
		}
		guildID := strings.TrimSpace(event.GuildID)
		channelID := strings.TrimSpace(event.ChannelID)
		messageID := strings.TrimSpace(event.MessageID)
		if guildID == "" || channelID == "" || messageID == "" {
			return nil
		}
		submission, err := m.service.Submit(ctx, guildID, channelID, event.Content, m.clock.Now())
		if err != nil {
			return err
		}
		switch submission.State {
		case domain.AutoChatPaidIgnored:
			return ctx.Err()
		case domain.AutoChatPaidUnsafeInput:
			if err := m.sendTransientWarning(ctx, channelID, messageID, legacyAutoChatUnsafeInputMessage, legacyAutoChatUnsafeInputDelay); err != nil {
				return err
			}
			return events.ErrStopPropagation
		case domain.AutoChatPaidBusy:
			if err := m.sendTransientWarning(ctx, channelID, messageID, legacyAutoChatBusyMessage, legacyAutoChatBusyWarningDelay); err != nil {
				return err
			}
			return events.ErrStopPropagation
		case domain.AutoChatPaidQueued:
			if m.typing != nil {
				_ = m.typing.SendTyping(ctx, channelID)
			}
			if err := m.wait(ctx, LegacyAutoChatPaidResponseDelay); err != nil {
				return err
			}
			response, err := m.service.Response(ctx, guildID, submission.Dispatch.RequestTimeMilli)
			if err != nil {
				return err
			}
			content := response.Content
			if strings.Contains(content, "@") {
				content = legacyAutoChatUnsafeOutputMessage
			}
			_, err = m.messages.SendMessage(ctx, channelID, ports.OutboundMessage{
				Content:          content,
				ReplyToMessageID: messageID,
				AllowedMentions:  ports.AllowedMentions{},
			})
			if err != nil {
				return err
			}
			return events.ErrStopPropagation
		default:
			return ctx.Err()
		}
	}
}

func (m PaidRuntimeModule) sendTransientWarning(ctx context.Context, channelID string, sourceMessageID string, content string, delay time.Duration) error {
	warning, err := m.messages.SendMessage(ctx, channelID, ports.OutboundMessage{
		Content:         content,
		AllowedMentions: ports.AllowedMentions{},
	})
	if err != nil {
		return err
	}
	_ = m.messages.DeleteMessage(ctx, ports.MessageRef{ChannelID: channelID, MessageID: sourceMessageID})
	if err := m.wait(ctx, delay); err != nil {
		return err
	}
	_ = m.messages.DeleteMessage(ctx, warning)
	return ctx.Err()
}
