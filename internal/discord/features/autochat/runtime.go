package autochat

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/autochat"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
)

const (
	legacyFallbackMinimumDelay = time.Second
	legacyFallbackDelayRange   = 4 * time.Second
)

type RuntimeModule struct {
	service   coreservice.FallbackService
	messages  ports.DiscordMessagePort
	typing    ports.DiscordTypingPort
	randomInt func(int) int
	wait      func(context.Context, time.Duration) error
}

func NewRuntimeModule(configs ports.AutoChatConfigReader, balances ports.BalanceRepository, messages ports.DiscordMessagePort, typing ports.DiscordTypingPort) (RuntimeModule, error) {
	if messages == nil {
		return RuntimeModule{}, errors.New("autochat fallback message port is required")
	}
	service, err := coreservice.NewFallbackService(configs, balances)
	if err != nil {
		return RuntimeModule{}, err
	}
	return RuntimeModule{
		service:   service,
		messages:  messages,
		typing:    typing,
		randomInt: secureRandomInt,
		wait:      waitForAutoChatReply,
	}, nil
}

func (m RuntimeModule) RegisterEventRoutes(dispatcher *events.Dispatcher) {
	if dispatcher == nil || m.messages == nil {
		return
	}
	dispatcher.Register(events.TypeMessageCreate, m.MessageCreateHandler())
}

func (m RuntimeModule) MessageCreateHandler() events.Handler {
	return func(ctx context.Context, event events.Event) error {
		if event.Type != events.TypeMessageCreate || event.IsBot {
			return nil
		}
		if event.Member != nil && event.Member.IsBot {
			return nil
		}
		guildID := strings.TrimSpace(event.GuildID)
		channelID := strings.TrimSpace(event.ChannelID)
		messageID := strings.TrimSpace(event.MessageID)
		if guildID == "" || channelID == "" || messageID == "" {
			return nil
		}
		reply, err := m.service.Reply(ctx, guildID, channelID, event.Content)
		if err != nil || reply.Content == "" {
			return err
		}
		if reply.UseTypingDelay {
			if m.typing != nil {
				_ = m.typing.SendTyping(ctx, channelID)
			}
			delay := legacyFallbackMinimumDelay + time.Duration(m.randomOffset(int(legacyFallbackDelayRange/time.Millisecond)))*time.Millisecond
			if err := m.wait(ctx, delay); err != nil {
				return err
			}
		}
		_, err = m.messages.SendMessage(ctx, channelID, ports.OutboundMessage{
			Content:          reply.Content,
			ReplyToMessageID: messageID,
			AllowedMentions:  ports.AllowedMentions{},
		})
		return err
	}
}

func (m RuntimeModule) randomOffset(maximum int) int {
	if maximum <= 0 || m.randomInt == nil {
		return 0
	}
	value := m.randomInt(maximum) % maximum
	if value < 0 {
		value = -value
	}
	return value
}

func secureRandomInt(maximum int) int {
	if maximum <= 0 {
		return 0
	}
	value, err := rand.Int(rand.Reader, big.NewInt(int64(maximum)))
	if err != nil {
		return 0
	}
	return int(value.Int64())
}

func waitForAutoChatReply(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
